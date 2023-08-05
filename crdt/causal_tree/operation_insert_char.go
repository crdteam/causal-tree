package causal_tree

import (
	"encoding/json"
	"fmt"

	atm "github.com/crdteam/causal-tree/crdt/atom"
)

// +--------------------------+
// | Operations - Insert char |
// +--------------------------+

// InsertChar represents insertion of a char to the right of another atom.
type InsertChar struct {
	// Char inserted in tree.
	Char rune
}

func (v InsertChar) AtomPriority() int { return insertCharPriority }
func (v InsertChar) MarshalJSON() ([]byte, error) {
	return json.Marshal(fmt.Sprintf("insert %c", v.Char))
}
func (v InsertChar) String() string { return string([]rune{v.Char}) }

// This function checks if the given atom value (child) is a possible child of
// the current type (v). InsertChar can only have InsertChar or Delete as a
// child.
func (v InsertChar) ValidateChild(child atm.AtomValue) error {
	switch child.(type) {
	case InsertChar, Delete:
		return nil
	default:
		return fmt.Errorf("invalid atom value after InsertChar: %T (%v)", child, child)
	}
}

// InsertChar inserts a char after the cursor position and advances the cursor.
func (t *CausalTree) InsertChar(ch rune) error {
	atomID, err := t.addAtom(InsertChar{ch})
	if err != nil {
		return err
	}
	t.Cursor = atomID
	return nil
}

// InsertCharAt inserts a char after the given (tree) position.
func (t *CausalTree) InsertCharAt(ch rune, i int) error {
	if err := t.SetCursor(i); err != nil {
		return err
	}
	return t.InsertChar(ch)
}
