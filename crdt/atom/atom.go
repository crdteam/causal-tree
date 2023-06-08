package atom

import (
	"encoding/json"
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

// AtomValue is a tree operation.
type AtomValue interface {
	json.Marshaler
	// AtomPriority returns where this atom should be placed compared with its siblings.
	AtomPriority() int
	// ValidateChild checks whether the given value can be appended as a child.
	ValidateChild(child AtomValue) error
}

// Atom represents an atomic operation within a replicated tree.
type Atom struct {
	// ID is the identifier of this atom.
	ID AtomID
	// Cause is the identifier of the preceding atom.
	Cause AtomID
	// Value is the data operation represented by this atom.
	Value AtomValue
}

// Atom's methods:

// +--------+
// | String |
// +--------+

func (a Atom) String() string {
	return fmt.Sprintf("Atom(%v,%v,%v)", a.ID, a.Cause, a.Value)
}

// +----------+
// | Ordering |
// +----------+

// Compare returns the relative order between atoms.
func (a Atom) Compare(other Atom) int {
	// Ascending according to priority.
	if a.Value.AtomPriority() < other.Value.AtomPriority() {
		return -1
	}
	if a.Value.AtomPriority() > other.Value.AtomPriority() {
		return +1
	}
	return a.ID.Compare(other.ID)
}

// +---------------+
// | Remap indices |
// +---------------+

/*


 */
func (a Atom) RemapSite(m indexmap.IndexMap) Atom {
	return Atom{
		ID:    a.ID.RemapSite(m),
		Cause: a.Cause.RemapSite(m),
		Value: a.Value,
	}
}
