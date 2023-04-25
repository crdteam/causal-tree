package crdt

import (
	"fmt"
)

// Value represents a structure that may be converted to concrete data.
//
// Each concrete type has a Snapshot() method with appropriate return type.
type Value interface {
	isValue()
}

// treePosition stores an atom's position for a cursor.
// The atom is always defined by its ID, but we also store its last known position to
// speed up searching for it. Since trees are insert-only, the atom can only be at this
// position or to its right.
type treePosition struct {
	t            *CausalTree
	atomID       AtomID
	lastKnownPos int
}

func (p *treePosition) atomIndex() int {
	size := len(p.t.Weave)
	for i := p.lastKnownPos; i < size; i++ {
		if p.t.Weave[i].ID == p.atomID {
			p.lastKnownPos = i
			return i
		}
	}
	panic(fmt.Sprintf("atomID %v not found after position %d (weave size: %d)", p.atomID, p.lastKnownPos, size))
}

// String is a Value representing a sequence of Unicode codepoints.
type String struct{ treePosition }

func (*String) isValue() {}

// Snapshot returns the string represented by the atom.
func (s *String) Snapshot() string {
	i := s.atomIndex()

	var chars []rune
	var isDeleting bool
	walkCausalBlock(s.t.Weave[i:], func(atom Atom) bool {
		switch v := atom.Value.(type) {
		case Delete:
			if atom.Cause == s.atomID {
				// String itself is deleted, interrupt iteration.
				return false
			}
			// A char atom may have multiple Delete children, so we remove it only
			// once when isDeleting=false.
			if !isDeleting {
				isDeleting = true
				chars = chars[:len(chars)-1]
			}
		case InsertChar:
			isDeleting = false
			chars = append(chars, v.Char)
		}
		return true
	})

	return string(chars)
}

func (t *CausalTree) StringValue(atomID AtomID) (*String, error) {
	i := t.atomIndex(atomID)
	atom := t.Weave[i]
	if _, ok := atom.Value.(InsertStr); !ok {
		return nil, fmt.Errorf("%v is not an InsertStr atom: %T (%v)", atomID, atom, atom)
	}
	return &String{treePosition{
		t:            t,
		atomID:       atomID,
		lastKnownPos: t.atomIndex(atomID),
	}}, nil
}
