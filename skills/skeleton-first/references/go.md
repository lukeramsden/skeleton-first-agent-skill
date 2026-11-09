# Go skeletons

## Stubs

Every function body in the skeleton is exactly:

```go
panic("skeleton: unimplemented")
```

Use that exact string. The checker counts stubs by it, and a grep for it lists what is left to implement.

Do not return zero values from stubs. A stub that returns `nil, nil` lets tests pass by accident.

## Packages

- One package per component. The package doc comment (in `doc.go`) says what the component owns and lists allowed imports: `// Depends on: store, clock.`
- Put interfaces in the package that *uses* them, not the one that implements them. This keeps dependencies pointing one way and lets each component be tested with fakes.
- Define sentinel errors (`var ErrUnknownTitle = errors.New(...)`) in the skeleton, since contracts refer to them.

## Types

- Write structs with their real fields, unexported where the representation is private. The representation is part of the requirement.
- Add `// Invariants: INV-n` to a type's doc comment when the rule is about its fields. The checker reads type comments too, but participants are still functions: list the methods that can break it.

## Contracts

Use the grammar in [contracts.md](contracts.md). Indent bullets with three spaces after `//` so `gofmt` keeps them as a list. `gofmt` puts a blank `//` line after a list; leave it there.

```go
// Return takes back copies of title.
//
// Requires: copies > 0.
// Ensures:
//   - available[title] SHALL rise by copies and onLoan[title] SHALL drop by copies.
//
// Errors:
//   - ErrUnknownTitle IF title is not in the catalogue, before any mutation.
//   - ErrNotOnLoan IF copies > onLoan[title], before any mutation.
//
// Invariants: INV-1.
func (l *Library) Return(title string, copies int) error {
	panic("skeleton: unimplemented")
}
```

## Checks before review

All of these must pass on the skeleton:

```bash
go build ./...
go vet ./...
scripts/check-skeleton .
```

`check-skeleton` runs the first two itself.

## Tests

Write test files against the skeleton during implementation, one component at a time. Property tests (`testing/quick` or `pgregory.net/rapid`) suit `Ensures` rules and invariants well: an invariant is a property that holds after any sequence of calls to its participants.
