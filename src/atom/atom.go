package atom

import (
	"encoding/json"
	"fmt"

	"github.com/crdteam/causal-tree/src/utils/indexmap"
)

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

/*
	This method returns the relative order between atoms.
	Atoms are ordered by priority, then by ID. The priority is defined by the
	AtomValue interface.

	Atoms with a lower priority are considered to be "older" than atoms with a
	higher priority.

	Return

	Values returned:
		- +1 if a > other
		- -1 if a < other
		- 0 if a == other
*/
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
