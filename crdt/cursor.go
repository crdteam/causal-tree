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

func (s *String) walk(onDeleteChar, onInsertChar func(atom Atom)) {
	i := s.atomIndex()

	isDeleting := false
	walkCausalBlock(s.t.Weave[i:], func(atom Atom) bool {
		switch atom.Value.(type) {
		case Delete:
			if atom.Cause == s.atomID {
				// String itself is deleted, interrupt iteration.
				return false
			}
			// A char atom may have multiple Delete children, so we remove it only
			// once when isDeleting=false.
			if !isDeleting {
				isDeleting = true
				onDeleteChar(atom)
			}
		case InsertChar:
			isDeleting = false
			onInsertChar(atom)
		}
		return true
	})
}

// Snapshot returns the string represented by the atom.
func (s *String) Snapshot() string {
	var chars []rune
	onDeleteChar := func(atom Atom) {
		chars = chars[:len(chars)-1]
	}
	onInsertChar := func(atom Atom) {
		chars = append(chars, atom.Value.(InsertChar).Char)
	}
	s.walk(onDeleteChar, onInsertChar)
	return string(chars)
}

// Len returns the size of the string (in codepoints).
func (s *String) Len() int {
	var size int
	onDeleteChar := func(atom Atom) {
		size--
	}
	onInsertChar := func(atom Atom) {
		size++
	}
	s.walk(onDeleteChar, onInsertChar)
	return size
}

// ----

// StringValue returns a wrapper over InsertStr.
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
