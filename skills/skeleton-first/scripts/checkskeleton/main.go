// Command checkskeleton checks a Go skeleton against the skeleton-first rules:
// it builds and vets the project, cross-checks INVARIANTS.md against the
// Invariants: citations in doc comments, and reports contract coverage,
// over-long contracts, and remaining stubs.
//
// Exit status is 1 if the build, vet, or invariant cross-check fails.
// Coverage and length findings are warnings.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const stubMessage = "skeleton: unimplemented"

var (
	invIDRe       = regexp.MustCompile(`INV-\d+`)
	invHeadingRe  = regexp.MustCompile(`^#{1,6}\s+(INV-\d+)\b`)
	participantRe = regexp.MustCompile(`^\s*Participants:\s*(.*)$`)
	sectionRe     = regexp.MustCompile(`^(Requires|Ensures|Errors|Effects|Invariants):`)
)

type function struct {
	key        string // pkg.Recv.Name or pkg.Name
	pos        token.Position
	exported   bool
	contracted bool     // has Requires, Ensures, or Errors
	cites      []string // INV IDs in the Invariants: section
	docLines   int
	stub       bool
	complexity int
}

type invariant struct {
	id           string
	line         int
	participants []string
}

func main() {
	invPath := flag.String("invariants", "", "path to INVARIANTS.md (default <root>/INVARIANTS.md)")
	maxDoc := flag.Int("max-doc-lines", 15, "warn when a doc comment exceeds this many lines")
	minComplexity := flag.Int("min-complexity", 4, "report uncontracted implemented functions at or above this cyclomatic complexity")
	skipBuild := flag.Bool("skip-build", false, "do not run go build and go vet")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: check-skeleton [flags] [project-root]\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	root := "."
	if flag.NArg() > 0 {
		root = flag.Arg(0)
	}
	root = resolve(root)
	if *invPath == "" {
		*invPath = filepath.Join(root, "INVARIANTS.md")
	} else {
		*invPath = resolve(*invPath)
	}

	failed := false

	if !*skipBuild {
		for _, args := range [][]string{{"build", "./..."}, {"vet", "./..."}} {
			section("go " + strings.Join(args, " "))
			cmd := exec.Command("go", args...)
			cmd.Dir = root
			out, err := cmd.CombinedOutput()
			if err != nil {
				fmt.Print(indent(string(out)))
				fmt.Println("  FAIL")
				failed = true
			} else {
				fmt.Println("  ok")
			}
		}
	}

	funcs, typeCites, err := parseProject(root)
	if err != nil {
		fatal(err)
	}

	section("invariants")
	invs, err := parseInvariants(*invPath)
	switch {
	case os.IsNotExist(err):
		fmt.Printf("  no %s\n", rel(root, *invPath))
		for _, f := range funcs {
			for _, id := range f.cites {
				fmt.Printf("  ERROR %s: %s cites %s but there is no INVARIANTS.md\n", f.pos, f.key, id)
				failed = true
			}
		}
	case err != nil:
		fatal(err)
	default:
		if checkInvariants(root, *invPath, invs, funcs, typeCites) {
			failed = true
		}
	}

	section("contract coverage")
	reportCoverage(funcs, *minComplexity)

	section("contract length")
	long := 0
	for _, f := range funcs {
		if f.docLines > *maxDoc {
			fmt.Printf("  WARN %s: %s has a %d-line doc comment (budget %d)\n", f.pos, f.key, f.docLines, *maxDoc)
			long++
		}
	}
	if long == 0 {
		fmt.Println("  ok")
	}

	section("stubs")
	stubs := 0
	for _, f := range funcs {
		if f.stub {
			stubs++
		}
	}
	fmt.Printf("  %d of %d functions are unimplemented stubs\n", stubs, len(funcs))

	if failed {
		fmt.Println("\nFAIL")
		os.Exit(1)
	}
	fmt.Println("\nPASS")
}

func parseProject(root string) ([]function, map[string][]string, error) {
	var funcs []function
	typeCites := map[string][]string{} // INV ID -> type positions citing it
	fset := token.NewFileSet()

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name := d.Name()
		if d.IsDir() {
			if path != root && (strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") || name == "vendor" || name == "testdata") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			return nil
		}
		file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return err
		}
		pkg := file.Name.Name
		for _, decl := range file.Decls {
			switch decl := decl.(type) {
			case *ast.FuncDecl:
				f := describeFunc(fset, pkg, decl)
				f.pos.Filename = rel(root, f.pos.Filename)
				funcs = append(funcs, f)
			case *ast.GenDecl:
				if decl.Tok != token.TYPE {
					continue
				}
				for _, spec := range decl.Specs {
					ts := spec.(*ast.TypeSpec)
					doc := ts.Doc
					if doc == nil && len(decl.Specs) == 1 {
						doc = decl.Doc
					}
					for _, id := range citations(doc.Text()) {
						pos := fset.Position(ts.Pos())
						pos.Filename = rel(root, pos.Filename)
						typeCites[id] = append(typeCites[id], pos.String())
					}
				}
			}
		}
		return nil
	})
	sort.Slice(funcs, func(i, j int) bool { return funcs[i].key < funcs[j].key })
	return funcs, typeCites, err
}

func describeFunc(fset *token.FileSet, pkg string, fd *ast.FuncDecl) function {
	key := pkg + "." + fd.Name.Name
	if fd.Recv != nil && len(fd.Recv.List) > 0 {
		key = pkg + "." + receiverName(fd.Recv.List[0].Type) + "." + fd.Name.Name
	}
	text := fd.Doc.Text()
	f := function{
		key:        key,
		pos:        fset.Position(fd.Pos()),
		exported:   fd.Name.IsExported(),
		cites:      citations(text),
		docLines:   strings.Count(text, "\n"),
		stub:       isStub(fd.Body),
		complexity: cyclomatic(fd.Body),
	}
	for _, line := range strings.Split(text, "\n") {
		if m := sectionRe.FindStringSubmatch(strings.TrimSpace(line)); m != nil && m[1] != "Effects" && m[1] != "Invariants" {
			f.contracted = true
		}
	}
	return f
}

func receiverName(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.StarExpr:
		return receiverName(e.X)
	case *ast.IndexExpr:
		return receiverName(e.X)
	case *ast.IndexListExpr:
		return receiverName(e.X)
	case *ast.Ident:
		return e.Name
	}
	return "?"
}

// citations returns the INV IDs listed in the Invariants: section of a doc
// comment. The section runs from its heading to the next section heading or
// blank line.
func citations(doc string) []string {
	var ids []string
	in := false
	for _, line := range strings.Split(doc, "\n") {
		t := strings.TrimSpace(line)
		if m := sectionRe.FindStringSubmatch(t); m != nil {
			in = m[1] == "Invariants"
		} else if t == "" {
			in = false
		}
		if in {
			ids = append(ids, invIDRe.FindAllString(t, -1)...)
		}
	}
	return ids
}

func isStub(body *ast.BlockStmt) bool {
	if body == nil || len(body.List) != 1 {
		return false
	}
	es, ok := body.List[0].(*ast.ExprStmt)
	if !ok {
		return false
	}
	call, ok := es.X.(*ast.CallExpr)
	if !ok || len(call.Args) != 1 {
		return false
	}
	if id, ok := call.Fun.(*ast.Ident); !ok || id.Name != "panic" {
		return false
	}
	lit, ok := call.Args[0].(*ast.BasicLit)
	return ok && strings.Trim(lit.Value, "\"`") == stubMessage
}

func cyclomatic(body *ast.BlockStmt) int {
	if body == nil {
		return 1
	}
	n := 1
	ast.Inspect(body, func(node ast.Node) bool {
		switch node := node.(type) {
		case *ast.FuncLit:
			return false // counted as its own unit, not here
		case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt:
			n++
		case *ast.CaseClause:
			if node.List != nil {
				n++
			}
		case *ast.CommClause:
			if node.Comm != nil {
				n++
			}
		case *ast.BinaryExpr:
			if node.Op == token.LAND || node.Op == token.LOR {
				n++
			}
		}
		return true
	})
	return n
}

func parseInvariants(path string) ([]invariant, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var invs []invariant
	scanner := bufio.NewScanner(f)
	line := 0
	for scanner.Scan() {
		line++
		text := scanner.Text()
		if m := invHeadingRe.FindStringSubmatch(text); m != nil {
			invs = append(invs, invariant{id: m[1], line: line})
			continue
		}
		if m := participantRe.FindStringSubmatch(text); m != nil && len(invs) > 0 {
			cur := &invs[len(invs)-1]
			for _, p := range strings.Split(m[1], ",") {
				p = strings.Trim(strings.TrimSpace(p), "`")
				if p != "" {
					cur.participants = append(cur.participants, p)
				}
			}
		}
	}
	return invs, scanner.Err()
}

// matches reports whether a participant name such as "Library.Lend" or
// "catalog.Library.Lend" names the function with the given full key.
func matches(key, participant string) bool {
	return key == participant || strings.HasSuffix(key, "."+participant)
}

func checkInvariants(root, path string, invs []invariant, funcs []function, typeCites map[string][]string) (failed bool) {
	file := rel(root, path)
	errorf := func(format string, args ...any) {
		fmt.Printf("  ERROR "+format+"\n", args...)
		failed = true
	}

	defined := map[string]invariant{}
	for _, inv := range invs {
		if prev, dup := defined[inv.id]; dup {
			errorf("%s:%d: %s is defined twice (first at line %d)", file, inv.line, inv.id, prev.line)
		}
		defined[inv.id] = inv
		if len(inv.participants) == 0 {
			errorf("%s:%d: %s has no Participants line", file, inv.line, inv.id)
		}
	}

	for _, inv := range invs {
		for _, p := range inv.participants {
			var found []function
			for _, f := range funcs {
				if matches(f.key, p) {
					found = append(found, f)
				}
			}
			if len(found) == 0 {
				errorf("%s:%d: %s lists participant %s, which does not exist", file, inv.line, inv.id, p)
				continue
			}
			if len(found) > 1 {
				errorf("%s:%d: %s participant %s is ambiguous; qualify it with the package", file, inv.line, inv.id, p)
				continue
			}
			if !contains(found[0].cites, inv.id) {
				errorf("%s: %s is a participant of %s but does not cite it in Invariants:", found[0].pos, found[0].key, inv.id)
			}
		}
	}

	for _, f := range funcs {
		for _, id := range f.cites {
			inv, ok := defined[id]
			if !ok {
				errorf("%s: %s cites %s, which is not defined in %s", f.pos, f.key, id, file)
				continue
			}
			listed := false
			for _, p := range inv.participants {
				if matches(f.key, p) {
					listed = true
				}
			}
			if !listed {
				errorf("%s: %s cites %s but is not in its Participants line", f.pos, f.key, id)
			}
		}
	}

	for id, positions := range typeCites {
		if _, ok := defined[id]; !ok {
			for _, pos := range positions {
				errorf("%s: type cites %s, which is not defined in %s", pos, id, file)
			}
		}
	}

	if !failed {
		fmt.Printf("  ok (%d invariants)\n", len(invs))
	}
	return failed
}

func reportCoverage(funcs []function, minComplexity int) {
	var exported, contracted int
	var uncontracted, risky []function
	for _, f := range funcs {
		if !f.exported {
			continue
		}
		exported++
		if f.contracted {
			contracted++
			continue
		}
		uncontracted = append(uncontracted, f)
		if !f.stub && f.complexity >= minComplexity {
			risky = append(risky, f)
		}
	}
	if exported == 0 {
		fmt.Println("  no exported functions")
		return
	}
	fmt.Printf("  %d of %d exported functions have contracts (L2+)\n", contracted, exported)

	if len(risky) > 0 {
		sort.Slice(risky, func(i, j int) bool { return risky[i].complexity > risky[j].complexity })
		fmt.Println("  WARN implemented functions without contracts, most complex first:")
		for _, f := range risky {
			fmt.Printf("    complexity %2d  %s  (%s)\n", f.complexity, f.key, f.pos)
		}
	}
	if len(uncontracted) > 0 {
		fmt.Println("  L1 only (fine for simple functions):")
		for _, f := range uncontracted {
			fmt.Printf("    %s\n", f.key)
		}
	}
}

func contains(xs []string, x string) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}

// resolve makes path absolute relative to the directory the wrapper script
// was invoked from, since the wrapper changes directory to run the checker.
func resolve(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	base := os.Getenv("CHECK_SKELETON_CWD")
	if base == "" {
		base, _ = os.Getwd()
	}
	return filepath.Join(base, path)
}

func rel(root, path string) string {
	if r, err := filepath.Rel(root, path); err == nil {
		return r
	}
	return path
}

func section(name string) { fmt.Printf("\n== %s\n", name) }

func indent(s string) string {
	s = strings.TrimRight(s, "\n")
	if s == "" {
		return ""
	}
	return "  " + strings.ReplaceAll(s, "\n", "\n  ") + "\n"
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "check-skeleton:", err)
	os.Exit(2)
}
