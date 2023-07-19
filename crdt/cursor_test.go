package crdt_test

import (
	"fmt"
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

func TestStringCursor(t *testing.T) {
	tests := []struct {
		desc   string
		ops    []operation
		value  string
		atomID crdt.AtomID
	}{
		{
			"only inserts",
			[]operation{
				{op: insertStr},
				{op: insertChar, char: 'c'},
				{op: insertChar, char: 'r'},
				{op: insertChar, char: 'd'},
				{op: insertChar, char: 't'},
			},
			"crdt",
			crdt.AtomID{0, 0, 2},
		},
		{
			"delete str[1]",
			[]operation{
				{op: insertStr},
				{op: insertChar, char: 'c'},
				{op: insertChar, char: 'r'},
				{op: insertChar, char: 'd'},
				{op: insertChar, char: 't'},

				// Delete char 'r'.
				{op: deleteCharAt, pos: 2},
			},
			"cdt",
			crdt.AtomID{0, 0, 2},
		},
		{
			"delete str[1] twice",
			[]operation{
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
			},
			"cdt",
			crdt.AtomID{0, 0, 2},
		},
		{
			"delete string",
			[]operation{
				{op: insertStr},
				{op: insertChar, char: 'c'},
				{op: insertChar, char: 'r'},
				{op: insertChar, char: 'd'},
				{op: insertChar, char: 't'},

				// Delete string.
				{op: deleteCharAt, pos: 0},
			},
			"",
			crdt.AtomID{0, 0, 2},
		},
	}
	for _, test := range tests {
		trees := testOperations(t, test.ops)
		str, err := trees[0].StringValue(test.atomID)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		t.Run(fmt.Sprintf("OneCursorPerChar/desc=%s", test.desc), func(t *testing.T) {
			for i, want := range test.value {
				cur := str.Cursor()
				cur.Index(i)
				got, err := cur.Value()
				if err != nil {
					t.Fatalf("iteration #%d: err: %v", i, err)
				}
				if got != want {
					t.Errorf("str[%d] = %c (!= %c)", i, got, want)
				}
			}
		})
		t.Run(fmt.Sprintf("MutableCursorForward/desc=%s", test.desc), func(t *testing.T) {
			cur := str.Cursor()
			for i, want := range test.value {
				cur.Index(i)
				got, err := cur.Value()
				if err != nil {
					t.Fatalf("iteration #%d: err: %v", i, err)
				}
				if got != want {
					t.Errorf("str[%d] = %c (!= %c)", i, got, want)
				}
			}
		})
		t.Run(fmt.Sprintf("MutableCursorBackward/desc=%s", test.desc), func(t *testing.T) {
			cur := str.Cursor()
			chars := []rune(test.value)
			n := len(chars)
			for i := n - 1; i >= 0; i-- {
				cur.Index(i)
				got, err := cur.Value()
				if err != nil {
					t.Fatalf("err: %v", err)
				}
				want := chars[i]
				if got != want {
					t.Errorf("str[%d] = %c (!= %c)", i, got, want)
				}
			}
		})
		t.Run(fmt.Sprintf("Errors/desc=%s", test.desc), func(t *testing.T) {
			cur := str.Cursor()
			chars := []rune(test.value)
			n := len(chars)
			if err := cur.Index(n); err == nil {
				t.Errorf("cur.Index(%d): want err, got nil", n)
			}
			if err := cur.Index(-2); err == nil {
				t.Errorf("cur.Index(%d): want err, got nil", -2)
			}
			if got, err := cur.Value(); err == nil {
				t.Errorf("want err, got nil (char: %c)", got)
			}
		})
	}
}
