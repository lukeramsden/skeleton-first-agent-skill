package catalog

import "errors"

var (
	ErrUnknownTitle = errors.New("catalog: unknown title")
	ErrNotOnLoan    = errors.New("catalog: more copies returned than are on loan")
)

// Library tracks copies per title.
//
// Invariants: INV-1.
type Library struct {
	available map[string]int
	onLoan    map[string]int
}

// New returns an empty library.
func New() *Library {
	panic("skeleton: unimplemented")
}

// AddCopies adds newly acquired copies of title to the shelf, adding the
// title to the catalogue if it is new.
//
// Requires: copies > 0.
// Ensures: available[title] SHALL rise by copies.
// Invariants: INV-1.
func (l *Library) AddCopies(title string, copies int) {
	panic("skeleton: unimplemented")
}

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

// Return takes back copies of title.
//
// Requires: copies > 0.
// Ensures: available[title] SHALL rise by copies and onLoan[title] SHALL drop by copies.
// Errors:
//   - ErrUnknownTitle IF title is not in the catalogue, before any mutation.
//   - ErrNotOnLoan IF copies > onLoan[title], before any mutation.
//
// Invariants: INV-1.
func (l *Library) Return(title string, copies int) error {
	panic("skeleton: unimplemented")
}

// Available reports how many copies of title are on the shelf.
func (l *Library) Available(title string) int {
	panic("skeleton: unimplemented")
}
