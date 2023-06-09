package atom

import (
	"testing"

	"github.com/crdteam/causal-tree/src/utils/indexmap"
)

// Constants for max uint16 and uint32 values.
const (
	maxUint16 = ^uint16(0)
	maxUint32 = ^uint32(0)
)

func TestAtomID_String(t *testing.T) {
	testCases := []struct {
		name   string
		atomID AtomID
		want   string
	}{
		{
			name: "Case 1: Site is 1 and Timestamp is 2",
			atomID: AtomID{
				Site:      1,
				Index:     0,
				Timestamp: 2,
			},
			want: "S1@T02",
		},
		{
			name: "Case 2: Site is 3 and Timestamp is 4",
			atomID: AtomID{
				Site:      3,
				Index:     1,
				Timestamp: 4,
			},
			want: "S3@T04",
		},
		{
			name: "StandardCase",
			atomID: AtomID{
				Site:      1,
				Index:     2,
				Timestamp: 3,
			},
			want: "S1@T03",
		},
		{
			name: "ZeroCase",
			atomID: AtomID{
				Site:      0,
				Index:     0,
				Timestamp: 0,
			},
			want: "S0@T00",
		},
		{
			name: "MaxUint16Case",
			atomID: AtomID{
				Site:      maxUint16,
				Index:     0,
				Timestamp: 0,
			},
			want: "S65535@T00",
		},
		{
			name: "MaxUint32Case",
			atomID: AtomID{
				Site:      0,
				Index:     maxUint32,
				Timestamp: maxUint32,
			},
			want: "S0@T4294967295",
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
			name: "older timestamp",
			id: AtomID{
				Site:      1,
				Timestamp: 1,
			},
			other: AtomID{
				Site:      1,
				Timestamp: 2,
			},
			expect: -1,
		},
		{
			name: "newer timestamp",
			id: AtomID{
				Site:      1,
				Timestamp: 3,
			},
			other: AtomID{
				Site:      1,
				Timestamp: 2,
			},
			expect: 1,
		},
		{
			name: "equal timestamp, older site",
			id: AtomID{
				Site:      2,
				Timestamp: 2,
			},
			other: AtomID{
				Site:      1,
				Timestamp: 2,
			},
			expect: -1,
		},
		{
			name: "equal timestamp, newer site",
			id: AtomID{
				Site:      1,
				Timestamp: 2,
			},
			other: AtomID{
				Site:      2,
				Timestamp: 2,
			},
			expect: 1,
		},
		{
			name: "equal timestamp and site",
			id: AtomID{
				Site:      1,
				Timestamp: 2,
			},
			other: AtomID{
				Site:      1,
				Timestamp: 2,
			},
			expect: 0,
		},
		// Edge Cases
		{
			name: "maximum timestamp, same site",
			id: AtomID{
				Site:      1,
				Timestamp: maxUint32,
			},
			other: AtomID{
				Site:      1,
				Timestamp: 1,
			},
			expect: 1,
		},
		{
			name: "same timestamp, maximum site",
			id: AtomID{
				Site:      maxUint16,
				Timestamp: 1,
			},
			other: AtomID{
				Site:      1,
				Timestamp: 1,
			},
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
			name: "no remap",
			id: AtomID{
				Site:      1,
				Index:     1,
				Timestamp: 1,
			},
			indexMap:   indexmap.IndexMap{},
			expectSite: 1,
		},
		{
			name: "remap site 1 to 2",
			id: AtomID{
				Site:      1,
				Index:     1,
				Timestamp: 1,
			},
			indexMap: indexmap.IndexMap{
				1: 2,
			},
			expectSite: 2,
		},
		{
			name: "remap site 2 to 1, original site is 1",
			id: AtomID{
				Site:      1,
				Index:     1,
				Timestamp: 1,
			},
			indexMap: indexmap.IndexMap{
				2: 1,
			},
			expectSite: 1,
		},
		{
			name: "remap site 2 to 1, original site is 2",
			id: AtomID{
				Site:      2,
				Index:     1,
				Timestamp: 1,
			},
			indexMap: indexmap.IndexMap{
				2: 1,
			},
			expectSite: 1,
		},
		{
			name: "remap site 0 to 65535",
			id: AtomID{
				Site:      0,
				Index:     1,
				Timestamp: 1,
			},
			indexMap: indexmap.IndexMap{
				0: 65535,
			},
			expectSite: 65535,
		},
		{
			name: "remap site 65535 to 0",
			id: AtomID{
				Site:      maxUint16,
				Index:     1,
				Timestamp: 1,
			},
			indexMap: indexmap.IndexMap{
				int(maxUint16): 0,
			},
			expectSite: 0,
		},
		{
			name: "remap site 32768 to 32767",
			id: AtomID{
				Site:      32768,
				Index:     1,
				Timestamp: 1,
			},
			indexMap: indexmap.IndexMap{
				32768: 32767,
			},
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
