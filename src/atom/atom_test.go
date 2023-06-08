package atom

import "testing"

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
