package atom

import (
	"fmt"

	"github.com/crdteam/causal-tree/crdt/index"
)

// AtomID is the unique identifier of an atom.
type AtomID struct {
	// Site is the index in the sitemap of the site that created an atom.
	Site uint16
	// Index is the order of creation of this atom in the given site.
	// Or: the atom index on its site's yarn.
	Index uint32
	// Timestamp is the site's Lamport timestamp when the atom was created.
	Timestamp uint32
}

// AtomID's methods

// +--------+
// | String |
// +--------+

// String returns a string representation of an atom ID.
// The format is "S<site>@T<timestamp>".
//
// Example: "S0@T01"
func (id AtomID) String() string {
	return fmt.Sprintf("S%d@T%02d", id.Site, id.Timestamp)
}

// +----------+
// | Ordering |
// +----------+

/*
	Compares two AtomIDs to determine their relative ordering based on their timestamps and sites.

	Ordering Rules

		1. Timestamp: An AtomID with an older timestamp is considered "older" (i.e., smaller).
		2. Site: If timestamps are equal, the AtomID with a larger site is considered "older" (i.e., smaller).

	Return Values

	Returns one of three possible values depending on the comparison:

		- -1: The current AtomID is older than the other.
		- +1: The current AtomID is younger than the other.
		- 0: Both AtomIDs are equal.

	Important

	This ordering ensures that older timestamps take precedence over sites, and within the same timestamp, larger sites are considered older.
*/
func (id AtomID) Compare(other AtomID) int {
	// Ascending according to timestamp (older first)
	if id.Timestamp < other.Timestamp {
		return -1
	}
	if id.Timestamp > other.Timestamp {
		return +1
	}
	// Descending according to site (younger first)
	if id.Site > other.Site {
		return -1
	}
	if id.Site < other.Site {
		return +1
	}
	return 0
}

// +---------------+
// | Remap indices |
// +---------------+

/*
	RemapSite remaps the site index of an AtomID using the given index map.
	It returns a new AtomID with the remapped site index.
*/
func (id AtomID) RemapSite(m index.Map) AtomID {
	return AtomID{
		Site:      uint16(m.Get(int(id.Site))),
		Index:     id.Index,
		Timestamp: id.Timestamp,
	}
}
