package crdt

import (
	"fmt"
	"unicode"
)

// Invokes the closure f for each atom of the causal block, including the head and except for Deletes.
// Returns the number of atoms visited.
//
// The closure should return 'false' to cut the traversal short, as in a 'break' statement. Otherwise, return true.
//
// The causal block is defined as the contiguous range containing the head and all of its descendents.
//
// Time complexity: O(atoms), or, O(avg. block size)
func walkCausalBlock2(weave []Atom, headPos int, f func(pos int, atom Atom, isDeleted bool) bool) int {
	block := weave[headPos:]
	if len(block) == 0 {
		return 0
	}
	head := block[0]
	i := 0
	count := 1
	for i < len(block) {
		atom := block[i]
		if i > 0 && atom.Cause.Timestamp < head.ID.Timestamp {
			// First atom whose parent has a lower timestamp (older) than head is the
			// end of the causal block.
			break
		}
		pos := headPos + i
		// Checks if current atom is deleted.
		isDeleted := false
		for i = i + 1; i < len(block); i++ {
			if _, ok := block[i].Value.(Delete); ok {
				isDeleted = true
			} else {
				break
			}
		}
		// Invokes closure, exiting if it returns false.
		if !f(pos, atom, isDeleted) {
			break
		}
		count++
	}
	return count
}

// ----

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
	// ID ...
	ID AtomID

	t            *CausalTree
	lastKnownPos int
}

func (p *treePosition) atomIndex() int {
	size := len(p.t.Weave)
	for i := p.lastKnownPos; i < size; i++ {
		if p.t.Weave[i].ID == p.ID {
			p.lastKnownPos = i
			return i
		}
	}
	panic(fmt.Sprintf("atomID %v not found after position %d (weave size: %d)", p.ID, p.lastKnownPos, size))
}

func (p *treePosition) walk(f func(pos int, atom Atom, isDeleted bool) bool) {
	headPos := p.atomIndex()
	walkCausalBlock2(p.t.Weave, headPos, f)
}

// ---- String value

// String is a Value representing a sequence of Unicode codepoints.
type String struct{ treePosition }

func (*String) isValue() {}

// walkChars skips the string head.
func (s *String) walkChars(f func(i int, atom Atom, isDeleted bool) bool) {
	s.walk(func(pos int, atom Atom, isDeleted bool) bool {
		switch atom.Value.(type) {
		case InsertStr:
			return true
		case InsertChar:
			return f(pos, atom, isDeleted)
		default:
			panic(fmt.Sprintf("unexpected atom type in String: %T", atom.Value))
		}
	})
}

// IsDeleted returns whether the string has been deleted.
func (s *String) IsDeleted() bool {
	i := s.atomIndex()
	weave := s.t.Weave
	if i+1 >= len(weave) {
		return false
	}
	if _, ok := weave[i+1].Value.(Delete); ok {
		return true
	}
	return false
}

// Snapshot returns the string represented by the atom.
// Ignores whether the string was deleted.
func (s *String) Snapshot() string {
	var chars []rune
	s.walkChars(func(pos int, atom Atom, isDeleted bool) bool {
		if !isDeleted {
			chars = append(chars, atom.Value.(InsertChar).Char)
		}
		return true
	})
	return string(chars)
}

// Len returns the size of the string (in codepoints).
// Ignores whether the string was deleted.
func (s *String) Len() int {
	var size int
	s.walkChars(func(pos int, atom Atom, isDeleted bool) bool {
		if !isDeleted {
			size++
		}
		return true
	})
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
	return &String{
		treePosition{
			ID:           cur.t.Weave[headPos].ID,
			t:            cur.t,
			lastKnownPos: headPos,
		},
	}
}

// Index moves the cursor to the given string position.
// Returns an error if index is out of range [-1:Len()-1]. Ignores whether string is deleted.
func (cur *StringCursor) Index(i int) error {
	if i < -1 {
		return fmt.Errorf("out of bounds")
	}
	s := cur.GetString()
	if i == -1 {
		// Move cursor to string head.
		cur.ID = s.ID
		cur.lastKnownPos = s.lastKnownPos
		return nil
	}
	// Walk the string starting from head to find the char at position i.
	indexPos := -1
	var count int
	s.walkChars(func(pos int, atom Atom, isDeleted bool) bool {
		if isDeleted {
			return true
		}
		if count == i {
			// Found the i-th char, break.
			indexPos = pos
			return false
		}
		count++
		return true
	})
	if indexPos == -1 {
		return fmt.Errorf("out of bounds")
	}
	cur.ID = cur.t.Weave[indexPos].ID
	cur.lastKnownPos = indexPos
	return nil
}

// Value returns the character pointed by the cursor.
// Returns an error if cursor is pointing to the string head.
func (cur *StringCursor) Value() (rune, error) {
	pos := cur.atomIndex()
	atom := cur.t.Weave[pos]
	switch v := atom.Value.(type) {
	case InsertChar:
		return v.Char, nil
	case InsertStr:
		return unicode.ReplacementChar, fmt.Errorf("out of bounds")
	default:
		panic("Unexpected atom type")
	}
}

// Insert inserts a new character after the cursor.
// The cursor is moved to the new character.
// Returns an error if atom insertion failed.
func (cur *StringCursor) Insert(ch rune) (*Char, error) {
	pos := cur.atomIndex()
	atomPos, err := cur.t.addAtom2(pos, InsertChar{ch})
	if err != nil {
		return nil, err
	}
	// Move cursor to new atom.
	cur.lastKnownPos = atomPos
	cur.ID = cur.t.Weave[atomPos].ID
	return &Char{cur.treePosition, cur.lastKnownHeadPos}, nil
}

// Delete removes the character pointed by the cursor.
// Returns an error if cursor is pointing to the string head.
func (cur *StringCursor) Delete() error {
	pos := cur.atomIndex()
	atom := cur.t.Weave[pos]
	if _, ok := atom.Value.(InsertStr); ok {
		return fmt.Errorf("out of bounds")
	}
	if _, err := cur.t.addAtom2(pos, Delete{}); err != nil {
		return err
	}
	// Fix cursor position, moving it one to the left.
	// In this sense, "delete" is like the backspace key.
	//    v
	// abcdef
	//   v
	// abcef
	//  v
	// abef
	s := cur.GetString()
	prevPos := cur.lastKnownHeadPos
	s.walkChars(func(pos int, atom Atom, isDeleted bool) bool {
		if atom.ID == cur.ID {
			prev := cur.t.Weave[prevPos]
			cur.ID = prev.ID
			cur.lastKnownPos = prevPos
			return false
		}
		if !isDeleted {
			prevPos = pos
		}
		return true
	})
	return nil
}

// ---- String char

// Char is an immutable tree location, pointing to an InsertChar atom.
type Char struct {
	treePosition

	// Store the last known position of the string's head, e.g., InsertStr atom.
	lastKnownHeadPos int
}

func (ch *Char) Snapshot() rune {
	pos := ch.atomIndex()
	return ch.t.Weave[pos].Value.(InsertChar).Char
}

// TODO: (*Char).IsDeleted() bool
// TODO: (*Char).GetString() *String
// TODO: (*Char).GetStringCursor() *StringCursor

// ---- CausalTree methods

// StringValue returns a wrapper over InsertStr.
func (t *CausalTree) StringValue(atomID AtomID) (*String, error) {
	i := t.atomIndex(atomID)
	atom := t.Weave[i]
	if _, ok := atom.Value.(InsertStr); !ok {
		return nil, fmt.Errorf("%v is not an InsertStr atom: %T (%v)", atomID, atom, atom)
	}
	return &String{treePosition{
		ID:           atomID,
		t:            t,
		lastKnownPos: i,
	}}, nil
}

// SetString sets the tree register to a new string and returns it.
func (t *CausalTree) SetString() (*String, error) {
	// TODO: change implementation once we remove internal cursor from CausalTree.
	if err := t.InsertStr(); err != nil {
		return nil, err
	}
	return t.StringValue(t.Cursor)
}

// DeleteAtom deletes the given atom from the tree.
func (t *CausalTree) DeleteAtom(atomID AtomID) error {
	// TODO: change implementation once we remove internal cursor from CausalTree.
	t.Cursor = atomID
	return t.Delete()
}

// TODO: MOVE THIS TO ctree.go AFTER ISSUE #5.
//
// This is a copy of addAtom, but using the known position of the cause.
func (t *CausalTree) addAtom2(causePos int, value AtomValue) (int, error) {
	t.Timestamp++
	if t.Timestamp == 0 {
		// Overflow
		return -1, ErrStateLimitExceeded
	}
	var causeID AtomID
	if causePos >= 0 {
		cause := t.Weave[causePos]
		causeID = cause.ID
		if err := cause.Value.ValidateChild(value); err != nil {
			return -1, err
		}
	}
	i := siteIndex(t.Sitemap, t.SiteID)
	atomID := AtomID{
		Site:      uint16(i),
		Index:     uint32(len(t.Yarns[i])),
		Timestamp: t.Timestamp,
	}
	atom := Atom{
		ID:    atomID,
		Cause: causeID,
		Value: value,
	}
	atomPos := t.insertAtomAtCursor2(causePos, atom)
	t.Yarns[i] = append(t.Yarns[i], atom)
	return atomPos, nil
}

// TODO: MOVE THIS TO ctree.go AFTER ISSUE #5.
//
// This is a copy of insertAtomAtCursor, but using the known position of the cause.
func (t *CausalTree) insertAtomAtCursor2(causePos int, atom Atom) int {
	if causePos < 0 {
		// Cursor is at initial position.
		t.insertAtom(atom, 0)
		return 0
	}
	causeID := t.Weave[causePos].ID
	insertPos := causePos + 1
	walkCausalBlock(t.Weave[causePos:], func(a Atom) bool {
		if a.Cause == causeID && a.Compare(atom) < 0 {
			// a is the first child smaller than atom, break.
			return false
		}
		insertPos++
		return true
	})
	t.insertAtom(atom, insertPos)
	return insertPos
}
