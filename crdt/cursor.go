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

func (p *treePosition) walk(f func(pos int, atom Atom, isDeleted bool) bool) {
	headPos := p.atomIndex()
	walkCausalBlock2(p.t.Weave, headPos, f)
}

// ---- String value

// String is a Value representing a sequence of Unicode codepoints.
type String struct{ treePosition }

func (*String) isValue() {}

func (s *String) walkChars(f func(i int, atom Atom, isDeleted bool) bool) {
	s.walk(func(pos int, atom Atom, isDeleted bool) bool {
		switch atom.Value.(type) {
		case InsertStr:
			return !isDeleted
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
// Returns the empty string if string was deleted.
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
// Returns 0 if string was deleted.
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
	if headPos < 0 {
		panic("invalid weave")
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
// Returns an error if index is out of range [-1:Len()-1], or string is deleted.
func (cur *StringCursor) Index(i int) error {
	if i < -1 {
		return fmt.Errorf("out of bounds")
	}
	s := cur.GetString()
	if s.IsDeleted() {
		return fmt.Errorf("string is deleted")
	}
	// Walk the string starting from head to find the char at position i.
	indexPos := -1
	var count int
	s.walkChars(func(pos int, atom Atom, isDeleted bool) bool {
		if isDeleted {
			return true
		}
		if count == i {
			indexPos = pos
			return false
		}
		count++
		return true
	})
	if indexPos == -1 {
		return fmt.Errorf("out of bounds")
	}
	cur.atomID = cur.t.Weave[indexPos].ID
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
