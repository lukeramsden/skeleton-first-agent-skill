# Contract grammar

One grammar for every contract, in every language. The checker and the implementer both rely on it, so do not invent new section names.

## Sections

Contracts live in the doc comment directly above the function. After the one-line summary and a blank line, use these sections in this order. Omit any that would be empty.

| Section | Meaning |
|---|---|
| `Requires:` | What the caller must guarantee. Violating it is a caller bug; the function may panic or misbehave. |
| `Ensures:` | What the function guarantees on return, given `Requires`. |
| `Errors:` | Each error the function can return, and the state it leaves behind. |
| `Effects:` | Side effects allowed: `none`, or a list such as `reads clock`, `writes store`, `network`. Omit if the project does not track effects. |
| `Invariants:` | IDs from `INVARIANTS.md` this function must preserve, e.g. `INV-1, INV-3`. |

A function with none of `Requires`, `Ensures` or `Errors` is at level L1. Adding them makes it L2. Citing an invariant makes it L3.

## Writing rules

Write each `Ensures` and `Errors` rule as an EARS-style statement:

- Unconditional: `<subject> SHALL <response>.`
- Event-driven: `WHEN <trigger> THEN <subject> SHALL <response>.`
- Unwanted behaviour: `IF <condition> THEN <subject> SHALL <response>.`

Use `SHALL` for requirements. Do not use "should", "may", or "tries to"; the implementer cannot test them.

One rule per bullet. If a section has one rule, it can sit on the same line as the heading.

Refer to state by names that exist in the skeleton (fields, parameters, return values). If you have to invent a name to state the rule, the skeleton is missing a field or a function.

## Errors

For each error, say *when* it happens and *what state is left*:

```
// Errors:
//   - ErrUnknownTitle IF title is not in the catalogue, before any mutation.
//   - ErrStoreUnavailable IF the write fails; the in-memory counts SHALL be rolled back.
```

"Before any mutation" and "rolled back" are the two answers that matter most. Leave neither implicit.

## What not to write

- Restatements of the signature: "Takes a string and returns an int."
- Implementation steps: "Loops over the slice and…". Contracts say what, not how.
- Rules any sensible implementation already follows: "SHALL not crash."
- Rules that span several functions. Put those in `INVARIANTS.md` and cite the ID.

## Global invariants

Each entry in `INVARIANTS.md` has:

```markdown
## INV-1: Copies are conserved

For every title, available + onLoan SHALL equal the number of copies owned.

Participants: Library.Lend, Library.Return, Library.AddCopies
```

- IDs are `INV-` followed by a number, never reused, even after deletion.
- `Participants` lists every function that can break the rule, as `Type.Method` or `Function` (optionally prefixed with the package: `catalog.Library.Lend`).
- Every participant cites the ID in its `Invariants:` section, and every function citing an ID is listed as a participant. The checker enforces both directions.
