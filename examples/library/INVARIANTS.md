# Invariants

Rules that span more than one function. Each participant cites the ID in its `Invariants:` doc section. IDs are never reused.

## INV-1: Copies are conserved

For every title, available + onLoan SHALL equal the total copies added with AddCopies. Neither count SHALL be negative.

Participants: Library.AddCopies, Library.Lend, Library.Return
