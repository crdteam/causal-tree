package crdt_test

import (
	"testing"

	"github.com/brunokim/causal-tree/crdt"
)

func TestString(t *testing.T) {
	t.Run("Snapshot", func(t *testing.T) {
		trees := testOperations(t, []operation{
			{op: insertStr},
			{op: insertChar, char: 'c'},
			{op: insertChar, char: 'r'},
			{op: insertChar, char: 'd'},
			{op: insertChar, char: 't'},
		})
		str, err := trees[0].StringValue(crdt.AtomID{0, 0, 2})
		if err != nil {
			t.Fatalf("err: %v", err)
		}

		wantStr := "crdt"
		if got := str.Snapshot(); got != wantStr {
			t.Errorf("str.Snapshot() = %s (!= %s)", got, wantStr)
		}
		wantLen := 4
		if got := str.Len(); got != wantLen {
			t.Errorf("str.Len() = %d (!= %d)", got, wantLen)
		}
	})
	t.Run("DeletedChar", func(t *testing.T) {
		trees := testOperations(t, []operation{
			{op: insertStr},
			{op: insertChar, char: 'c'},
			{op: insertChar, char: 'r'},
			{op: insertChar, char: 'd'},
			{op: insertChar, char: 't'},

			// Delete char 'r' in position #2.
			{op: deleteCharAt, pos: 2},
		})
		str, err := trees[0].StringValue(crdt.AtomID{0, 0, 2})
		if err != nil {
			t.Fatalf("err: %v", err)
		}

		wantStr := "cdt"
		if got := str.Snapshot(); got != wantStr {
			t.Errorf("str.Snapshot() = %s (!= %s)", got, wantStr)
		}
		wantLen := 3
		if got := str.Len(); got != wantLen {
			t.Errorf("str.Len() = %d (!= %d)", got, wantLen)
		}
	})
	t.Run("DoubleDeleteChar", func(t *testing.T) {
		trees := testOperations(t, []operation{
			{local: 0, op: insertStr},
			{local: 0, op: insertChar, char: 'c'},
			{local: 0, op: insertChar, char: 'r'},
			{local: 0, op: insertChar, char: 'd'},
			{local: 0, op: insertChar, char: 't'},

			// Doubly delete char 'r' in position #2 via merge.
			{local: 0, op: fork, remote: 1},
			{local: 0, op: deleteCharAt, pos: 2},
			{local: 1, op: deleteCharAt, pos: 2},
			{local: 0, op: merge, remote: 1},
		})
		str, err := trees[0].StringValue(crdt.AtomID{0, 0, 2})
		if err != nil {
			t.Fatalf("err: %v", err)
		}

		wantStr := "cdt"
		if got := str.Snapshot(); got != wantStr {
			t.Errorf("str.Snapshot() = %s (!= %s)", got, wantStr)
		}
		wantLen := 3
		if got := str.Len(); got != wantLen {
			t.Errorf("str.Len() = %d (!= %d)", got, wantLen)
		}
	})
}
