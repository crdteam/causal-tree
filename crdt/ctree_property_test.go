package crdt_test

import (
	"testing"

	"github.com/brunokim/causal-tree/crdt"
	"pgregory.net/rapid"
)

// Model a CausalTree as a slice of runes, subject to insertions and deletions
// at random positions with InsertCharAt and DeleteCharAt.
//
// We don't model the more primitive operations InsertChar, DeleteChar and SetCursor
// because it's complicated to model where the cursor ends up after a deletion. The
// cursor moves to the deleted atom's cause, or its first non-deleted ancestor,
// which may not be the char next to it.
//
// TODO: perhaps this is a sign that the cursor should be more predictable...?
type runesModel struct {
	t     *crdt.CausalTree
	chars []rune
}

func newRunesModel() *runesModel {
	m := new(runesModel)
	m.t = crdt.NewCausalTree()
	return m
}

func (m *runesModel) InsertCharAt(t *rapid.T) {
	ch := rapid.Rune().Draw(t, "ch")
	i := rapid.IntRange(-1, len(m.chars)-1).Draw(t, "i")

	err := m.t.InsertCharAt(ch, i)
	if err != nil {
		t.Fatal("(*runesModel).InsertCharAt:", err)
	}

	m.chars = append(m.chars[:i+1], append([]rune{ch}, m.chars[i+1:]...)...)
}

func (m *runesModel) DeleteCharAt(t *rapid.T) {
	if len(m.chars) == 0 {
		t.Skip("empty string")
	}
	i := rapid.IntRange(0, len(m.chars)-1).Draw(t, "i")

	err := m.t.DeleteAt(i)
	if err != nil {
		t.Fatal("(*runesModel).DeleteCharAt:", err)
	}

	copy(m.chars[i:], m.chars[i+1:])
	m.chars = m.chars[:len(m.chars)-1]
}

func (m *runesModel) Check(t *rapid.T) {
	got := m.t.ToString()
	want := string(m.chars)
	if got != want {
		t.Fatalf("content mismatch: want %q but got %q", want, got)
	}
	t.Log("content:", got)
}

func TestRunesProperty(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		t.Repeat(rapid.StateMachineActions(newRunesModel()))
	})
}

// Models a causal tree as multiple rune slices, using the object API.
//
// This allows us to model multiple strings being inserted, deleted, and having
// their chars inserted and deleted.

type cursorModel struct {
	cursor *crdt.StringCursor
	chars  []rune
	index  int
}

type multipleRunesModel struct {
	t     *crdt.CausalTree
	model []*cursorModel
}

func newMultipleRunesModel() *multipleRunesModel {
	m := new(multipleRunesModel)
	m.t = crdt.NewCausalTree()
	return m
}

func (m *multipleRunesModel) SetString(t *rapid.T) {
	s, err := m.t.SetString()
	if err != nil {
		t.Fatal("(*multipleRunesModel).InsertStr:", err)
	}
	m.model = append(m.model, &cursorModel{
		cursor: s.Cursor(),
		chars:  nil,
		index:  -1,
	})
}

func (m *multipleRunesModel) DeleteString(t *rapid.T) {
	n := len(m.model)
	if n == 0 {
		t.Skip("no strings")
	}
	i := rapid.IntRange(0, n-1).Draw(t, "i")
	if err := m.t.DeleteAtom(m.model[i].cursor.GetString().ID); err != nil {
		t.Fatal("(*multipleRunesModel).DeleteString:", err)
	}
	if i < n-1 {
		copy(m.model[i:], m.model[i+1:])
	}
	m.model[n-1] = nil
	m.model = m.model[:n-1]
}

func (m *multipleRunesModel) MoveCursor(t *rapid.T) {
	n := len(m.model)
	if n == 0 {
		t.Skip("no strings")
	}
	i := rapid.IntRange(0, n-1).Draw(t, "i")
	model := m.model[i]
	size := model.cursor.GetString().Len()
	j := rapid.IntRange(-1, size-1).Draw(t, "j")
	if err := model.cursor.Index(j); err != nil {
		t.Fatal("(*multipleRunesModel).MoveCursor:", err)
	}
	model.index = j
}

func (m *multipleRunesModel) InsertChar(t *rapid.T) {
	n := len(m.model)
	if n == 0 {
		t.Skip("no strings")
	}
	ch := rapid.Rune().Draw(t, "ch")
	i := rapid.IntRange(0, n-1).Draw(t, "i")
	model := m.model[i]
	char, err := model.cursor.Insert(ch)
	if err != nil {
		t.Fatal("(*multipleRunesModel).InsertChar: Insert:", err)
	}
	if char.Snapshot() != ch {
		t.Error("(*multipleRunesModel).InsertChar: Snapshot:", err)
	}
	j := model.index + 1
	tail := append([]rune{ch}, model.chars[j:]...)
	model.chars = append(model.chars[:j], tail...)
	model.index++
}

func (m *multipleRunesModel) DeleteChar(t *rapid.T) {
	n := len(m.model)
	if n == 0 {
		t.Skip("no strings")
	}
	i := rapid.IntRange(0, n-1).Draw(t, "i")
	model := m.model[i]
	if model.index < 0 {
		t.Skip("cursor pointing to string head")
	}
	if err := model.cursor.Delete(); err != nil {
		t.Fatal("(*multipleRunesModel).DeleteChar:", err)
	}
	j := model.index
	model.chars = append(model.chars[:j], model.chars[j+1:]...)
	model.index--
}

func (m *multipleRunesModel) Check(t *rapid.T) {
	for i, model := range m.model {
		got := model.cursor.GetString().Snapshot()
		want := string(model.chars)
		if got != want {
			t.Errorf("content mismatch at #%d: want %q but got %q", i, want, got)
		} else {
			t.Logf("content at #%d: %s", i, want)
		}
	}
}

func TestMultipleRunesProperty(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		t.Repeat(rapid.StateMachineActions(newMultipleRunesModel()))
	})
}
