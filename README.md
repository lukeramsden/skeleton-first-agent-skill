# Skeleton First Agent Skill

An agent skill for specifying programs to coding agents as **compiling code skeletons**: types with real fields, every signature with a stub body, contracts in doc comments, and cross-function invariants. A human reviews and freezes the skeleton; the agent then implements and verifies one component at a time against it.

It is based on *Skeleton Coding: Program Structure as a Specification for Agents* (Golatkar, Achille, Soatto et al., [arXiv 2610.38214](https://arxiv.org/abs/2610.38214)). The paper finds that code-shaped skeletons with per-function contracts and global invariants beat requirements, tests, and prose design docs as specs, and that over-detailed contracts hurt.

## Install

```bash
npx skills add lukeramsden/skeleton-first-agent-skill@skeleton-first
```

Install globally for all supported agents:

```bash
npx skills add lukeramsden/skeleton-first-agent-skill@skeleton-first -g -a '*' -y
```

## What the skill provides

- A workflow: draft skeleton → check → human review → freeze → implement and verify per component → keep the skeleton current
- Three specification levels (signatures, contracts, invariants) and rules for choosing between them
- A contract grammar (`Requires`, `Ensures`, `Errors`, `Effects`, `Invariants`) with EARS-style `SHALL` rules
- An `INVARIANTS.md` template with two-way citation between invariants and the functions that must preserve them
- Go conventions, and a checker for Go skeletons

## The checker

`skills/skeleton-first/scripts/check-skeleton [project-root]` needs Go installed. It:

- runs `go build ./...` and `go vet ./...`
- checks that every invariant in `INVARIANTS.md` is cited by each listed participant, and every citation refers to a defined invariant that lists the citing function
- reports contract coverage, and lists implemented functions with no contract, most complex first
- warns about doc comments over the length budget (`-max-doc-lines`, default 15)
- counts remaining `panic("skeleton: unimplemented")` stubs

It exits non-zero if the build, vet, or invariant cross-check fails. Try it on the example:

```bash
skills/skeleton-first/scripts/check-skeleton examples/library
```

## Repository layout

```text
skills/skeleton-first/
├── SKILL.md
├── references/
│   ├── contracts.md     # contract grammar
│   └── go.md            # Go conventions
├── templates/
│   └── INVARIANTS.md
└── scripts/
    ├── check-skeleton   # wrapper
    └── checkskeleton/   # Go source
examples/library/        # a small compiling skeleton
```

## Attribution

This is an independent distillation of the paper above. It is not affiliated with or endorsed by its authors or by AWS.

## License

MIT. See [LICENSE](LICENSE).
