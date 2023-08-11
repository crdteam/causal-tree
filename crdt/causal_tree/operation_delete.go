package causal_tree

import (
	"fmt"

	"github.com/crdteam/causal-tree/crdt/atom"
)

// +---------------------+
// | Operations - Delete |
// +---------------------+

// Delete represents deleting an element from the tree.
type Delete struct{}

func (v Delete) AtomPriority() int { return deletePriority }
func (v Delete) MarshalJSON() ([]byte, error) {
	return []byte(`"delete"`), nil
}
func (v Delete) String() string { return "⌫ " }

func (v Delete) ValidateChild(child atom.Value) error {
	return fmt.Errorf("invalid atom value after Delete: %T (%v)", child, child)
}

// Delete deletes the char at the cursor position, and relocates the cursor to its cause.
func (t *CausalTree) Delete() error {
	if t.Cursor.Timestamp == 0 {
		return ErrNoAtomToDelete
	}
	if _, err := t.addAtom(Delete{}); err != nil {
		return err
	}
	t.fixDeletedCursor()
	return nil
}

// DeleteAt deletes the char at the given (tree) position.
func (t *CausalTree) DeleteAt(i int) error {
	if err := t.SetCursor(i); err != nil {
		return err
	}
	return t.Delete()
}
