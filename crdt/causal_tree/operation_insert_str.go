package causal_tree

import (
	"encoding/json"
	"fmt"

	"github.com/crdteam/causal-tree/crdt/atom"
)

// +-----------------------------------+
// | Operations - Insert str container |
// +-----------------------------------+

//Inserts a string container as a child of the root atom.
type InsertStr struct{}

func (v InsertStr) AtomPriority() int { return insertStrPriority }
func (v InsertStr) MarshalJSON() ([]byte, error) {
	return json.Marshal("insert str container")
}

func (v InsertStr) String() string { return "STR: " }

func (v InsertStr) ValidateChild(child atom.Value) error {
	switch child.(type) {
	case InsertChar, Delete:
		return nil
	default:
		return fmt.Errorf("invalid atom value after InsertStr: %T (%v)", child, child)
	}
}

// InsertStr inserts a Str container after the root and advances the cursor.
func (t *CausalTree) InsertStr() error {
	t.Cursor = atom.ID{}
	atomID, err := t.addAtom(InsertStr{})
	t.Cursor = atomID
	return err
}
