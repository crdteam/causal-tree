package atom

import (
	"encoding/json"
	"fmt"
	"math"
	"testing"

	"github.com/crdteam/causal-tree/src/utils/indexmap"
)

/*--------------------------------Aux Functions-------------------------------*/
type DummyAtomValue struct {
	Priority int
}

func (d DummyAtomValue) AtomPriority() int {
	return d.Priority
}

func (d DummyAtomValue) ValidateChild(child AtomValue) error {
	return nil
}

func (d DummyAtomValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Priority)
}

/*----------------------------------------------------------------------------*/

// +---------------+
// | Atom ID tests |
// +---------------+

func TestAtomID_String(t *testing.T) {
	testCases := []struct {
		name   string
		atomID AtomID
		want   string
	}{
		{
			name:   "Case 1: Site is 1 and Timestamp is 2",
			atomID: AtomID{Site: 1, Index: 0, Timestamp: 2},
			want:   "S1@T02",
		},
		{
			name:   "Case 2: Site is 3 and Timestamp is 4",
			atomID: AtomID{Site: 3, Index: 1, Timestamp: 4},
			want:   "S3@T04",
		},
		{
			name:   "StandardCase",
			atomID: AtomID{Site: 1, Index: 2, Timestamp: 3},
			want:   "S1@T03",
		},
		{
			name:   "ZeroCase",
			atomID: AtomID{Site: 0, Index: 0, Timestamp: 0},
			want:   "S0@T00",
		},
		{
			name:   "MaxUint16Case",
			atomID: AtomID{Site: math.MaxUint16, Index: 0, Timestamp: 0},
			want:   "S65535@T00",
		},
		{
			name:   "MaxUint32Case",
			atomID: AtomID{Site: 0, Index: math.MaxUint32, Timestamp: math.MaxUint32},
			want:   "S0@T4294967295",
		},
		// Add more test cases as needed.
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.atomID.String()
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestAtomID_Compare(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name   string
		id     AtomID
		other  AtomID
		expect int
	}{
		{
			name:   "older timestamp",
			id:     AtomID{Site: 1, Timestamp: 1},
			other:  AtomID{Site: 1, Timestamp: 2},
			expect: -1,
		},
		{
			name:   "newer timestamp",
			id:     AtomID{Site: 1, Timestamp: 3},
			other:  AtomID{Site: 1, Timestamp: 2},
			expect: 1,
		},
		{
			name:   "equal timestamp, older site",
			id:     AtomID{Site: 2, Timestamp: 2},
			other:  AtomID{Site: 1, Timestamp: 2},
			expect: -1,
		},
		{
			name:   "equal timestamp, newer site",
			id:     AtomID{Site: 1, Timestamp: 2},
			other:  AtomID{Site: 2, Timestamp: 2},
			expect: 1,
		},
		{
			name:   "equal timestamp and site",
			id:     AtomID{Site: 1, Timestamp: 2},
			other:  AtomID{Site: 1, Timestamp: 2},
			expect: 0,
		},
		// Edge Cases
		{
			name:   "maximum timestamp, same site",
			id:     AtomID{Site: 1, Timestamp: math.MaxUint32},
			other:  AtomID{Site: 1, Timestamp: 1},
			expect: 1,
		},
		{
			name:   "same timestamp, maximum site",
			id:     AtomID{Site: math.MaxUint16, Timestamp: 1},
			other:  AtomID{Site: 1, Timestamp: 1},
			expect: -1,
		},
	}

	// Run the test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.id.Compare(tc.other)
			if got != tc.expect {
				t.Errorf("expected %d, got %d", tc.expect, got)
			}
		})
	}
}

func TestAtomID_RemapSite(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name       string
		id         AtomID
		indexMap   indexmap.IndexMap
		expectSite uint16
	}{
		{
			name:       "no remap",
			id:         AtomID{Site: 1, Index: 1, Timestamp: 1},
			indexMap:   indexmap.IndexMap{},
			expectSite: 1,
		},
		{
			name:       "remap site 1 to 2",
			id:         AtomID{Site: 1, Index: 1, Timestamp: 1},
			indexMap:   indexmap.IndexMap{1: 2},
			expectSite: 2,
		},
		{
			name:       "remap site 2 to 1, original site is 1",
			id:         AtomID{Site: 1, Index: 1, Timestamp: 1},
			indexMap:   indexmap.IndexMap{2: 1},
			expectSite: 1,
		},
		{
			name:       "remap site 2 to 1, original site is 2",
			id:         AtomID{Site: 2, Index: 1, Timestamp: 1},
			indexMap:   indexmap.IndexMap{2: 1},
			expectSite: 1,
		},
		{
			name:       "remap site 0 to 65535",
			id:         AtomID{Site: 0, Index: 1, Timestamp: 1},
			indexMap:   indexmap.IndexMap{0: 65535},
			expectSite: 65535,
		},
		{
			name:       "remap site 65535 to 0",
			id:         AtomID{Site: math.MaxUint16, Index: 1, Timestamp: 1},
			indexMap:   indexmap.IndexMap{int(math.MaxUint16): 0},
			expectSite: 0,
		},
		{
			name:       "remap site 32768 to 32767",
			id:         AtomID{Site: 32768, Index: 1, Timestamp: 1},
			indexMap:   indexmap.IndexMap{32768: 32767},
			expectSite: 32767,
		},
	}

	// Run the test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.id.RemapSite(tc.indexMap)
			if got.Site != tc.expectSite {
				t.Errorf("expected site %d, got %d", tc.expectSite, got.Site)
			}
		})
	}
}

// +----------------+
// | Atom tests     |
// +----------------+

func TestAtom_String(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name string
		atom Atom
		want string
	}{
		{
			name: "atom with default IDs and dummy value",
			atom: Atom{
				ID:    AtomID{},
				Cause: AtomID{},
				Value: DummyAtomValue{Priority: 1},
			},
			want: "Atom(S0@T00,S0@T00,{1})",
		},
		{
			name: "atom with specific IDs and dummy value",
			atom: Atom{
				ID:    AtomID{Site: 1, Index: 2, Timestamp: 3},
				Cause: AtomID{Site: 4, Index: 5, Timestamp: 6},
				Value: DummyAtomValue{Priority: 2},
			},
			want: "Atom(S1@T03,S4@T06,{2})",
		},
		{
			name: "atom with max site, index, and timestamp values",
			atom: Atom{
				ID:    AtomID{Site: math.MaxUint16, Index: math.MaxUint32, Timestamp: math.MaxUint32},
				Cause: AtomID{Site: math.MaxUint16, Index: math.MaxUint32, Timestamp: math.MaxUint32},
				Value: DummyAtomValue{Priority: 3},
			},
			want: fmt.Sprintf("Atom(S%d@T%d,S%d@T%d,{3})", math.MaxUint16, math.MaxUint32, math.MaxUint16, math.MaxUint32),
		},
		{
			name: "atom with nil value",
			atom: Atom{
				ID:    AtomID{Site: 4, Index: 5, Timestamp: 6},
				Cause: AtomID{Site: 7, Index: 8, Timestamp: 9},
				Value: nil,
			},
			want: "Atom(S4@T06,S7@T09,<nil>)",
		},
		{
			name: "atom with equal ID and Cause",
			atom: Atom{
				ID:    AtomID{Site: 10, Index: 11, Timestamp: 12},
				Cause: AtomID{Site: 10, Index: 11, Timestamp: 12},
				Value: DummyAtomValue{Priority: 4},
			},
			want: "Atom(S10@T12,S10@T12,{4})",
		},
	}

	// Run the test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.atom.String()
			if got != tc.want {
				t.Errorf("expected %q, got %q", tc.want, got)
			}
		})
	}
}

func TestAtom_Compare(t *testing.T) {
	testCases := []struct {
		name string
		a    Atom
		b    Atom
		want int
	}{
		{
			name: "atom a has higher priority than atom b",
			a: Atom{
				ID:    AtomID{Site: 1, Index: 2, Timestamp: 3},
				Cause: AtomID{Site: 4, Index: 5, Timestamp: 6},
				Value: DummyAtomValue{Priority: 2},
			},
			b: Atom{
				ID:    AtomID{Site: 7, Index: 8, Timestamp: 9},
				Cause: AtomID{Site: 10, Index: 11, Timestamp: 12},
				Value: DummyAtomValue{Priority: 1},
			},
			want: 1,
		},
		{
			name: "atom a has lower priority than atom b",
			a: Atom{
				ID:    AtomID{Site: 1, Index: 2, Timestamp: 3},
				Cause: AtomID{Site: 4, Index: 5, Timestamp: 6},
				Value: DummyAtomValue{Priority: 1},
			},
			b: Atom{
				ID:    AtomID{Site: 7, Index: 8, Timestamp: 9},
				Cause: AtomID{Site: 10, Index: 11, Timestamp: 12},
				Value: DummyAtomValue{Priority: 2},
			},
			want: -1,
		},
		{
			name: "atom a and atom b have equal priorities, atom a has smaller timestamp",
			a: Atom{
				ID:    AtomID{Site: 1, Index: 2, Timestamp: 3},
				Cause: AtomID{Site: 4, Index: 5, Timestamp: 6},
				Value: DummyAtomValue{Priority: 1},
			},
			b: Atom{
				ID:    AtomID{Site: 7, Index: 8, Timestamp: 4},
				Cause: AtomID{Site: 10, Index: 11, Timestamp: 12},
				Value: DummyAtomValue{Priority: 1},
			},
			want: -1,
		},
		{
			name: "atom a and atom b have equal priorities, atom a has later timestamp",
			a: Atom{
				ID:    AtomID{Site: 1, Index: 2, Timestamp: 5},
				Cause: AtomID{Site: 4, Index: 5, Timestamp: 6},
				Value: DummyAtomValue{Priority: 1},
			},
			b: Atom{
				ID:    AtomID{Site: 7, Index: 8, Timestamp: 4},
				Cause: AtomID{Site: 10, Index: 11, Timestamp: 12},
				Value: DummyAtomValue{Priority: 1},
			},
			want: 1,
		},
		{
			name: "atom a and atom b have equal priorities and timestamps, atom a has smaller site",
			a: Atom{
				ID:    AtomID{Site: 1, Index: 2, Timestamp: 4},
				Cause: AtomID{Site: 4, Index: 5, Timestamp: 6},
				Value: DummyAtomValue{Priority: 1},
			},
			b: Atom{
				ID:    AtomID{Site: 7, Index: 8, Timestamp: 4},
				Cause: AtomID{Site: 10, Index: 11, Timestamp: 12},
				Value: DummyAtomValue{Priority: 1},
			},
			want: 1,
		},
		{
			name: "atom a and atom b have equal priorities and timestamps, atom a has greater site",
			a: Atom{
				ID:    AtomID{Site: 3, Index: 2, Timestamp: 4},
				Cause: AtomID{Site: 4, Index: 5, Timestamp: 6},
				Value: DummyAtomValue{Priority: 1},
			},
			b: Atom{
				ID:    AtomID{Site: 2, Index: 8, Timestamp: 4},
				Cause: AtomID{Site: 10, Index: 11, Timestamp: 12},
				Value: DummyAtomValue{Priority: 1},
			},
			want: -1,
		},
		{
			name: "atom a and atom b have equal priorities and timestamps, and site",
			a: Atom{
				ID:    AtomID{Site: 3, Index: 2, Timestamp: 4},
				Cause: AtomID{Site: 4, Index: 5, Timestamp: 6},
				Value: DummyAtomValue{Priority: 1},
			},
			b: Atom{
				ID:    AtomID{Site: 3, Index: 8, Timestamp: 4},
				Cause: AtomID{Site: 10, Index: 11, Timestamp: 12},
				Value: DummyAtomValue{Priority: 1},
			},
			want: 0,
		},
	}
	// Run the test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.a.Compare(tc.b)
			if got != tc.want {
				t.Errorf("expected %q, got %q", tc.want, got)
			}
		})
	}
}

func TestAtom_RemapSite(t *testing.T) {
	testCases := []struct {
		name     string
		atom     Atom
		indexMap indexmap.IndexMap
		want     Atom
	}{
		{
			name: "remap site 1 to 2",
			atom: Atom{
				ID:    AtomID{Site: 1, Index: 2, Timestamp: 1},
				Cause: AtomID{Site: 1, Index: 1, Timestamp: 1},
				Value: DummyAtomValue{Priority: 1},
			},
			indexMap: indexmap.IndexMap{
				1: 2,
			},
			want: Atom{
				ID:    AtomID{Site: 2, Index: 2, Timestamp: 1},
				Cause: AtomID{Site: 2, Index: 1, Timestamp: 1},
				Value: DummyAtomValue{Priority: 1},
			},
		},
		{
			name: "empty index map",
			atom: Atom{
				ID:    AtomID{Site: 1, Index: 2, Timestamp: 1},
				Cause: AtomID{Site: 1, Index: 1, Timestamp: 1},
				Value: DummyAtomValue{Priority: 1},
			},
			indexMap: indexmap.IndexMap{},
			want: Atom{
				ID:    AtomID{Site: 1, Index: 2, Timestamp: 1},
				Cause: AtomID{Site: 1, Index: 1, Timestamp: 1},
				Value: DummyAtomValue{Priority: 1},
			},
		},
		{
			name: "different site and cause site, both in index map",
			atom: Atom{
				ID:    AtomID{Site: 1, Index: 2, Timestamp: 1},
				Cause: AtomID{Site: 2, Index: 1, Timestamp: 1},
				Value: DummyAtomValue{Priority: 1},
			},
			indexMap: indexmap.IndexMap{1: 3, 2: 4},
			want: Atom{
				ID:    AtomID{Site: 3, Index: 2, Timestamp: 1},
				Cause: AtomID{Site: 4, Index: 1, Timestamp: 1},
				Value: DummyAtomValue{Priority: 1},
			},
		},
		{
			name: "only site in index map",
			atom: Atom{
				ID:    AtomID{Site: 1, Index: 2, Timestamp: 1},
				Cause: AtomID{Site: 2, Index: 1, Timestamp: 1},
				Value: DummyAtomValue{Priority: 1},
			},
			indexMap: indexmap.IndexMap{1: 3},
			want: Atom{
				ID:    AtomID{Site: 3, Index: 2, Timestamp: 1},
				Cause: AtomID{Site: 2, Index: 1, Timestamp: 1},
				Value: DummyAtomValue{Priority: 1},
			},
		},
		{
			name: "only cause site in index map",
			atom: Atom{
				ID:    AtomID{Site: 1, Index: 2, Timestamp: 1},
				Cause: AtomID{Site: 2, Index: 1, Timestamp: 1},
				Value: DummyAtomValue{Priority: 1},
			},
			indexMap: indexmap.IndexMap{2: 3},
			want: Atom{
				ID:    AtomID{Site: 1, Index: 2, Timestamp: 1},
				Cause: AtomID{Site: 3, Index: 1, Timestamp: 1},
				Value: DummyAtomValue{Priority: 1},
			},
		},
		{
			name: "site and cause site not in index map",
			atom: Atom{
				ID:    AtomID{Site: 1, Index: 2, Timestamp: 1},
				Cause: AtomID{Site: 2, Index: 1, Timestamp: 1},
				Value: DummyAtomValue{Priority: 1},
			},
			indexMap: indexmap.IndexMap{3: 4},
			want: Atom{
				ID:    AtomID{Site: 1, Index: 2, Timestamp: 1},
				Cause: AtomID{Site: 2, Index: 1, Timestamp: 1},
				Value: DummyAtomValue{Priority: 1},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.atom.RemapSite(tc.indexMap)
			if got.ID.Site != tc.want.ID.Site {
				t.Errorf("expected site %d, got %d", tc.want.ID.Site, got.ID.Site)
			}
			if got.Cause.Site != tc.want.Cause.Site {
				t.Errorf("expected site %d, got %d", tc.want.Cause.Site, got.Cause.Site)
			}
		})
	}
}

// +-----------------+
// | functions tests |
// +-----------------+

// TODO make the test creation process more compact: example -> creating atoms in a more compact way
func TestFunctions_WalkCausalBlock(t *testing.T) {
	testCases := []struct {
		name  string
		block []Atom
		f     func(Atom) bool
		want  int
	}{
		{
			name:  "empty block",
			block: []Atom{},
			f: func(Atom) bool {
				return true
			},
			want: 0,
		},
		{
			name: "one atom block",
			block: []Atom{
				{
					ID: AtomID{Timestamp: 1},
				},
			},
			f: func(Atom) bool {
				return true
			},
			want: 1,
		},
		{
			name: "multiple atoms, none causing the walk to stop",
			block: []Atom{
				{
					ID: AtomID{Timestamp: 1},
				},
				{
					ID:    AtomID{Timestamp: 2},
					Cause: AtomID{Timestamp: 1},
				},
				{
					ID:    AtomID{Timestamp: 3},
					Cause: AtomID{Timestamp: 2},
				},
			},
			f: func(Atom) bool {
				return true
			},
			want: 3,
		},
		{
			name: "multiple atoms, one causing the walk to stop",
			block: []Atom{
				{
					ID: AtomID{Timestamp: 1},
				},
				{
					ID:    AtomID{Timestamp: 2},
					Cause: AtomID{Timestamp: 1},
				},
				{
					ID: AtomID{Timestamp: 3},
					Cause: AtomID{
						Timestamp: 0, // this atom will cause the walk to stop
					},
				},
			},
			f: func(Atom) bool {
				return true
			},
			want: 2,
		},
		{
			name: "function f returns false",
			block: []Atom{
				{
					ID: AtomID{Timestamp: 1},
				},
				{
					ID:    AtomID{Timestamp: 2},
					Cause: AtomID{Timestamp: 1},
				},
				{
					ID:    AtomID{Timestamp: 3},
					Cause: AtomID{Timestamp: 2},
				},
			},
			f: func(Atom) bool {
				return false // this function will cause the walk to stop
			},
			want: 1,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := WalkCausalBlock(tc.block, tc.f)
			if got != tc.want {
				t.Errorf("expected %d, got %d", tc.want, got)
			}
		})
	}
}
