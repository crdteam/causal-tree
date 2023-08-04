package causal_tree

import "testing"

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
