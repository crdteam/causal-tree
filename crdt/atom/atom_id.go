package atom

import (
	"fmt"

	"github.com/crdteam/causal-tree/crdt/utils/indexmap"
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

func (id AtomID) String() string {
	return fmt.Sprintf("S%d@T%02d", id.Site, id.Timestamp)
}

// +----------+
// | Ordering |
// +----------+

// Compare returns the relative order between atom IDs.
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

func (id AtomID) RemapSite(m indexmap.IndexMap) AtomID {
	return AtomID{
		Site:      uint16(m.Get(int(id.Site))),
		Index:     id.Index,
		Timestamp: id.Timestamp,
	}
}
