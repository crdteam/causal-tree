package crdt

import (
	"fmt"
	"unicode"
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

func (p *treePosition) walk(onDeletePos, onDelete, onAtom func(j int, atom Atom) bool) {
	i := p.atomIndex()

	isDeleting := false
	j := i
	walkCausalBlock(p.t.Weave[i:], func(atom Atom) bool {
		j++
		switch atom.Value.(type) {
		case Delete:
			if atom.Cause == p.atomID {
				// Position itself is deleted, maybe interrupt iteration.
				return onDeletePos(j, atom)
			}
			// Any atom may have multiple consecutive Delete children, so we call onDelete only
			// once, when isDeleting=false.
			if !isDeleting {
				isDeleting = true
				return onDelete(j, atom)
			}
			return true
		default:
			isDeleting = false
			return onAtom(j, atom)
		}
	})
}

// ---- String value

// String is a Value representing a sequence of Unicode codepoints.
type String struct{ treePosition }

func (*String) isValue() {}

func (s *String) walk(onDeleteChar, onInsertChar func(atom Atom)) {
	onDeletePos := func(j int, atom Atom) bool { return true }
	onDelete := func(j int, atom Atom) bool {
		onDeleteChar(atom)
		return true
	}
	onAtom := func(j int, atom Atom) bool {
		switch atom.Value.(type) {
		case InsertChar:
			onInsertChar(atom)
			return true
		default:
			panic(fmt.Sprintf("Invalid atom value for string: %T", atom.Value))
		}
	}
	s.treePosition.walk(onDeletePos, onDelete, onAtom)
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

// Cursor returns a cursor pointing to the string head.
func (s *String) Cursor() *StringCursor {
	return &StringCursor{s.treePosition, s.treePosition.lastKnownPos}
}

// ---- String cursor

// StringCursor is a mutable tree location, initialized to before the first char.
//
// It always points either to the head InsertStr, or to one of InsertChars within the head's
// causal block.
type StringCursor struct {
	treePosition

	// Store the last known position of the string's head, e.g., InsertStr atom.
	lastKnownHeadPos int
}

// GetString returns a pointer to the cursor's owner string.
func (cur *StringCursor) GetString() *String {
	// Find head.
	//
	// Given c0 as the last known position of the pointed-to atom, and c1 its
	// current position, we know that (c1-c0) atoms were inserted since the last
	// cursor usage. The head position may be anywhere between s0 (its last known
	// position) and s0+(c1-c0).
	//
	// We search backwards from the end of the range. The first InsertStr atom found
	// must be the head of the char's string.
	c0 := cur.lastKnownPos
	c1 := cur.atomIndex()
	s0 := cur.lastKnownHeadPos
	headPos := -1
	for j := s0 + (c1 - c0); j >= s0; j-- {
		atom := cur.t.Weave[j]
		if _, ok := atom.Value.(InsertStr); ok {
			headPos = j
			cur.lastKnownHeadPos = j
			break
		}
	}
	if headPos < 0 {
		panic(fmt.Sprintf("Invalid weave"))
	}
	return &String{
		treePosition{
			t:            cur.t,
			atomID:       cur.t.Weave[headPos].ID,
			lastKnownPos: headPos,
		},
	}
}

// Index moves the cursor to the given string position.
func (cur *StringCursor) Index(i int) error {
	if i < -1 {
		return fmt.Errorf("Out of bounds")
	}
	// Walk the string forward to find the char at position i.
	pos := -1
	var count int
	onDeletePos := func(j int, atom Atom) bool {
		return true
	}
	onDelete := func(j int, atom Atom) bool {
		count--
		if count == i {
			// Clear pos if atom is deleted.
			pos = -1
		}
		return true
	}
	onAtom := func(j int, atom Atom) bool {
		if count == i {
			pos = j
		} else if count > i {
			// Break when we're sure previous atom wasn't deleted.
			return false
		}
		count++
		return true
	}
	s := cur.GetString()
	s.treePosition.walk(onDeletePos, onDelete, onAtom)
	if pos == -1 {
		return fmt.Errorf("Out of bounds")
	}
	cur.atomID = cur.t.Weave[pos].ID
	cur.lastKnownPos = pos
	return nil
}

// Value returns the character pointed by the cursor.
func (cur *StringCursor) Value() (rune, error) {
	pos := cur.atomIndex()
	atom := cur.t.Weave[pos]
	switch v := atom.Value.(type) {
	case InsertChar:
		return v.Char, nil
	case InsertStr:
		return unicode.ReplacementChar, fmt.Errorf("Out of bounds")
	default:
		panic("Unexpected atom type")
	}
}

// ---- CausalTree methods

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
