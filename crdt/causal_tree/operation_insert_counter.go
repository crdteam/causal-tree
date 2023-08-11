package causal_tree

import (
	"encoding/json"
	"fmt"
	"strconv"

	atm "github.com/crdteam/causal-tree/crdt/atom"
)

// +------------------------------+
// | Operations - Insert Add atom |
// +------------------------------+

//Inserts an add atom as a child of the atom pointed by cursor.
type InsertAdd struct {
	//Value inserted into the counter container
	Value int32
}

func (v InsertAdd) AtomPriority() int { return insertAddPriority }
func (v InsertAdd) MarshalJSON() ([]byte, error) {
	return json.Marshal(fmt.Sprintf("insert %d", v.Value))
}

func (v InsertAdd) String() string { return strconv.FormatInt(int64(v.Value), 10) }

//InsertAdd atoms only accept child of type InsertAdd.
func (v InsertAdd) ValidateChild(child atm.Value) error {
	switch child.(type) {
	case InsertAdd:
		return nil
	default:
		return fmt.Errorf("invalid atom value after InsertAdd: %T (%v)", child, child)
	}
}

// InsertAdd inserts an InsertAdd atom after the cursor position and advances the cursor.
func (t *CausalTree) InsertAdd(val int32) error {
	atomID, err := t.addAtom(InsertAdd{val})
	if err != nil {
		return err
	}
	t.Cursor = atomID
	return nil
}

// InsertAddAt inserts an InsertAdd atom after the given (tree) position.
func (t *CausalTree) InsertAddAt(val int32, i int) error {
	if err := t.SetCursor(i); err != nil {
		return err
	}
	return t.InsertAdd(val)
}

// +---------------------------------------+
// | Operations - Insert counter container |
// +---------------------------------------+

//Inserts a counter container as a child of the root atom.
type InsertCounter struct{}

func (v InsertCounter) AtomPriority() int { return insertCounterPriority }
func (v InsertCounter) MarshalJSON() ([]byte, error) {
	return json.Marshal("insert counter container")
}

func (v InsertCounter) String() string { return "Counter: " }

func (v InsertCounter) ValidateChild(child atm.Value) error {
	switch child.(type) {
	case InsertAdd, Delete:
		return nil
	default:
		return fmt.Errorf("invalid atom value after InsertCounter: %T (%v)", child, child)
	}
}

// InsertCounter inserts a Counter container after the root and advances the cursor.
func (t *CausalTree) InsertCounter() error {
	t.Cursor = atm.ID{}
	atomID, err := t.addAtom(InsertCounter{})
	t.Cursor = atomID
	return err
}
