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
