package atom

import (
	"fmt"

	"github.com/crdteam/causal-tree/crdt/index"
)

// ID is the unique identifier of an atom.
type ID struct {
	// Site is the index in the sitemap of the site that created an atom.
	Site uint16
	// Index is the order of creation of this atom in the given site.
	// Or: the atom index on its site's yarn.
	Index uint32
	// Timestamp is the site's Lamport timestamp when the atom was created.
	Timestamp uint32
}

// ID's methods

// +--------+
// | String |
// +--------+

// String returns a string representation of an atom ID.
// The format is "S<site>@T<timestamp>".
//
// Example: "S0@T01"
func (id ID) String() string {
	return fmt.Sprintf("S%d@T%02d", id.Site, id.Timestamp)
}

// +----------+
// | Ordering |
// +----------+

/*
	Compares two IDs to determine their relative ordering based on their timestamps and sites.

	Ordering Rules

		1. Timestamp: An ID with an older timestamp is considered "older" (i.e., smaller).
		2. Site: If timestamps are equal, the ID with a larger site is considered "older" (i.e., smaller).

	Return Values

	Returns one of three possible values depending on the comparison:

		- -1: The current ID is older than the other.
		- +1: The current ID is younger than the other.
		- 0: Both IDs are equal.

	Important

	This ordering ensures that older timestamps take precedence over sites, and within the same timestamp, larger sites are considered older.
*/
func (id ID) Compare(other ID) int {
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
	RemapSite remaps the site index of an ID using the given index map.
	It returns a new ID with the remapped site index.
*/
func (id ID) RemapSite(m index.Map) ID {
	return ID{
		Site:      uint16(m.Get(int(id.Site))),
		Index:     id.Index,
		Timestamp: id.Timestamp,
	}
}
