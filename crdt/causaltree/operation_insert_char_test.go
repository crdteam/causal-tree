package causaltree

import (
	"testing"

	"github.com/crdteam/causal-tree/crdt/atom"
)

func Test_AtomPriority(t *testing.T) {
	t.Run("Check if the atom priority returned is correct", func(t *testing.T) {
		var v InsertChar
		got := v.AtomPriority()
		expected := insertCharPriority
		if got != expected {
			t.Errorf("got %q, want %q", got, expected)
		}
	})

}

func Test_String(t *testing.T) {
	testCases := []struct {
		name     string
		input    InsertChar
		expected string
	}{
		{
			name:     "Empty rune",
			input:    InsertChar{Char: rune(0)},
			expected: "\x00",
		},
		{
			name:     "Single character",
			input:    InsertChar{Char: 'a'},
			expected: "a",
		},
		{
			name:     "Special character",
			input:    InsertChar{Char: '!'},
			expected: "!",
		},
		{
			name:     "Number character",
			input:    InsertChar{Char: '1'},
			expected: "1",
		},
		{
			name:     "Unicode character",
			input:    InsertChar{Char: '愛'},
			expected: "愛",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.input.String()
			if got != tc.expected {
				t.Errorf("got %q, want %q", got, tc.expected)
			}
		})
	}
}

func Test_ValidateChild(t *testing.T) {
	testCases := []struct {
		name   string
		input  atom.Value
		gotNil bool
	}{
		{
			name:   "InsertChar",
			input:  InsertChar{Char: 'a'},
			gotNil: true,
		},
		{
			name:   "Delete",
			input:  Delete{},
			gotNil: true,
		},
		{
			name:   "Other",
			input:  InsertStr{},
			gotNil: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var v InsertChar
			got := v.ValidateChild(tc.input)
			if tc.gotNil {
				if got != nil {
					t.Errorf("got %q, want nil", got)
				}
			} else {
				if got == nil {
					t.Errorf("got nil, want %q", tc.input)
				}
			}
		})
	}
}

func Test_InsertChar(t *testing.T) {
	t.Run("Insert a char", func(t *testing.T) {
		// Create a causal tree
		tree := New()

		// Insert a char
		err := tree.InsertChar('a')
		if err != nil {
			t.Errorf("got %q, want nil", err)
		}

		// Check if the value was inserted
		got := tree.ToString()
		expected := "a"
		if got != expected {
			t.Errorf("got %q, want %q", got, expected)
		}
	})
}
