---
name: skeleton-first
description: Use before an agent implements a new program, package, or large feature. Write the program's structure as compiling code (types, signatures, stub bodies, contracts, invariants), get it reviewed and frozen, then implement and verify one component at a time against it. Also use when writing specs or issues for agents, or when an agent's output keeps drifting from the intended design.
---

# Skeleton first

Give the implementing agent the program's structure as code, next to the behavioural spec. Do not hand it only prose and tests.

## Why

Tests and examples choose a program *within* a structure. The structure itself (types, modules, function boundaries) decides how many tests are needed to pin the behaviour down. If the agent picks the structure, you cannot tell whether your tests are enough until it has written thousands of lines.

The evidence comes from *Skeleton Coding: Program Structure as a Specification for Agents* (Golatkar, Achille, Soatto et al., arXiv 2610.38214):

- Some structures make tests far more informative than others. Two structures can admit the same number of programs, yet one is pinned down by about d tests and the other needs about 2^d.
- Interfaces make internal state testable. Testing each of n stages takes about 4n checks. End-to-end testing can take 2^n, because later stages hide earlier errors.
- Some requirements cannot be tested at all, such as state representation and maintainability. The skeleton is the only place to require them.
- Across four models and three benchmarks, the best skeletons cut mean test error by 48% compared with the best end-to-end specs. They still helped when the agent could query the reference binary.
- Formats ranked: task description < behavioural spec (requirements or tests) < prose design doc < function stubs < stubs + per-function contracts < + global invariants.
- Piling more architecture into a markdown doc made inputs 4–7× longer and never matched code-shaped skeletons.
- Correct but over-detailed contracts *lowered* pass rates, especially for small models.
- Contract coverage helped most when added to the most complex functions first.
- Keeping the skeleton current during repair cut reference queries by 32%.

Limits of that evidence: the skeletons were derived from reference implementations, so the gains are an upper bound. The cost of writing the skeleton was not counted. Contracts were informal docstrings checked by the agent or an LLM judge. All tasks were rebuilds of existing libraries. Expect smaller gains on real greenfield work, and expect the quality of the skeleton to matter more than its presence.

## Workflow

1. **Draft the skeleton.** Follow "What a skeleton contains" below. Stop when it compiles; do not write implementations.
2. **Run the checker.** For Go: `scripts/check-skeleton <project-root>`. Fix every error before showing it to anyone.
3. **Hand it to a human for review.** This is where most of the value is: the human is checking the design, not the code. Point out the decisions you were least sure of. Do not proceed until they approve.
4. **Freeze it.** Commit the skeleton on its own, with a message starting `skeleton:`. From here it is the requirement.
5. **Implement one component at a time.** For each component: write its local tests from its contract, implement it, run its checks, then accept it or redo *that component*. Never regenerate the whole program to fix one part.
6. **Keep the skeleton current during repair.** If a fix needs a different signature, type, or contract, that is a skeleton change (see below), not a quiet edit.

If you are the implementer and were given a skeleton, start at step 5.

## What a skeleton contains

The skeleton is real code in the project's language, and it must compile. No markdown architecture doc stands in for it, because prose leaves the representation open and code does not.

It contains:

- **The module tree and allowed dependencies.** One package per component. State which packages may import which, in the package doc comment of each (`// Depends on: store, clock.`).
- **Types with their real fields.** This fixes the internal representation, which tests cannot do.
- **Every function and method signature,** with a stub body. In Go: `panic("skeleton: unimplemented")`.
- **Contracts in doc comments,** written in the grammar in [references/contracts.md](references/contracts.md).
- **`INVARIANTS.md`** at the project root for rules that span more than one function. Copy [templates/INVARIANTS.md](templates/INVARIANTS.md).

Language-specific rules: [references/go.md](references/go.md). For other languages, apply the same rules: the skeleton compiles, the type checker passes, and stubs fail loudly when called.

## How much to specify

Use three levels and pick per function, not per project:

| Level | Contents | Use for |
|---|---|---|
| L1 | Signature + one-line doc | Everything, by default |
| L2 | L1 + `Requires`/`Ensures`/`Errors` | Complex, stateful, or high-risk functions |
| L3 | L2 + `Invariants: INV-n` | Functions that share state rules with others |

Rules:

- **Add contracts to the most complex functions first.** Coverage gains come from the hard functions. A contract on a getter adds tokens and nothing else.
- **Only write down decisions that steer the implementer.** If a clause restates the type signature ("returns a string"), delete it. If any reasonable implementation would satisfy it without being told, delete it.
- **Respect the budget.** A doc comment over 15 lines is a warning; cut it or split the function. Halve that if a small or cheap model will implement it. Over-specification measurably hurts.
- **Write contracts a few tests can pin down.** If you cannot think of three or four tests that would catch a violation, the contract is too vague or the component is too big.

## Design for composition

Shape components so that if each one meets its contract, the whole program is correct.

- Give each component a narrow interface. Wide interfaces need exponentially more tests.
- Put every intermediate result that matters behind a function boundary, so it can be tested directly instead of only through the final output.
- When this holds, components can go to separate agents in parallel, and the easy ones to cheaper models.

## Verification

- **Test at the interfaces.** Each component gets unit or property tests derived from its contract. End-to-end tests only check that the pieces fit.
- **Prefer mechanical checks.** In order of trust: compiler > static analysis and linters > property tests > example tests > an LLM acting as judge. Only rely on an LLM judge for components the implementer usually gets right anyway.
- **Accept or redo per component.** Verification then adds up across components instead of multiplying.
- **If the project has an effect or capability checker,** declare effects in the contract (`Effects:` in the grammar) so "does no I/O" becomes a machine-checked rule rather than a docstring promise.

## Skeleton changes

The skeleton is versioned like a schema.

- Implementing agents must not edit signatures, types, contracts, or `INVARIANTS.md` silently.
- When a change is needed, stop and propose it: what changes, why the current skeleton cannot be met, and which components are affected. Make it a separate commit starting `skeleton:` so it shows up in review and in `git log`.
- After a skeleton change, re-run the checker and re-verify every affected component.
- On later feature work, rewrite against the skeleton rather than patching around behaviour nobody specified.

## Example

```go
// Lend lends copies of title if enough are available.
//
// Requires: copies > 0.
// Ensures:
//   - WHEN it returns true THEN available[title] SHALL drop by copies
//     and onLoan[title] SHALL rise by copies.
//   - WHEN it returns false THEN no counts SHALL change.
//
// Errors: ErrUnknownTitle IF title is not in the catalogue, before any mutation.
// Invariants: INV-1.
func (l *Library) Lend(title string, copies int) (bool, error) {
	panic("skeleton: unimplemented")
}
```

A complete compiling example, with `INVARIANTS.md`, is in the repository's `examples/library` directory.
