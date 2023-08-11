package causal_tree

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"

	"github.com/crdteam/causal-tree/crdt/atom"
	"github.com/crdteam/causal-tree/crdt/conversion"
	"github.com/crdteam/causal-tree/crdt/generate_uuid_elements"
	"github.com/crdteam/causal-tree/crdt/index"
	"github.com/crdteam/causal-tree/crdt/weft"
	"github.com/google/uuid"
)

var (
	uuidv1 = generate_uuid_elements.RandomUUIDv1 // Stubbed for mocking in mocks_test.go
)

// +-----------------------+
// | Basic data structures |
// +-----------------------+

// CausalTree is a replicated tree data structure.
//
// This data structure allows for 64K sites and 4G atoms in total.
type CausalTree struct {
	// Weave is the flat representation of a causal tree.
	Weave []atom.Atom
	// Cursor is the ID of the causing atom for the next operation.
	Cursor atom.ID
	// Yarns is the list of atoms, grouped by the site that created them.
	Yarns [][]atom.Atom
	// Sitemap is the ordered list of site IDs. The index in this sitemap is used to represent a site in atoms
	// and yarns.
	Sitemap []uuid.UUID
	// SiteID is this tree's site UUIDv1.
	SiteID uuid.UUID
	// Timestamp is this tree's Lamport timestamp.
	Timestamp uint32
}

// Returns the index where a site is (or should be) in the sitemap.
//
// Time complexity: O(log(sites))
func siteIndex(sitemap []uuid.UUID, siteID uuid.UUID) int {
	return sort.Search(len(sitemap), func(i int) bool {
		return bytes.Compare(sitemap[i][:], siteID[:]) >= 0
	})
}

// Returns the index of an atom within the weave.
//
// Time complexity: O(atoms)
func (t *CausalTree) atomIndex(atomID atom.ID) int {
	if atomID.Timestamp == 0 {
		return -1
	}
	for i, a := range t.Weave {
		if a.ID == atomID {
			return i
		}
	}
	return len(t.Weave)
}

// Gets an atom from yarns.
//
// Time complexity: O(1)
func (t *CausalTree) getAtom(atomID atom.ID) atom.Atom {
	return t.Yarns[atomID.Site][atomID.Index]
}

// Inserts an atom in the given weave index.
//
// Time complexity: O(atoms)
func (t *CausalTree) insertAtom(a atom.Atom, i int) {
	t.Weave = append(t.Weave, atom.Atom{})
	copy(t.Weave[i+1:], t.Weave[i:])
	t.Weave[i] = a
}

// +------+
// | Fork |
// +------+

// Fork a replicated tree into an independent object.
//
// Time complexity: O(atoms)
func (t *CausalTree) Fork() (*CausalTree, error) {
	if len(t.Sitemap)-1 >= math.MaxUint16 {
		return nil, ErrSiteLimitExceeded
	}
	newSiteID := uuidv1()
	i := siteIndex(t.Sitemap, newSiteID)
	if i == len(t.Sitemap) {
		t.Yarns = append(t.Yarns, nil)
		t.Sitemap = append(t.Sitemap, newSiteID)
	} else {
		// Remap atoms in yarns and weave.
		localRemap := make(index.Map)
		for j := i; j < len(t.Sitemap); j++ {
			localRemap.Set(j, j+1)
		}
		for i, yarn := range t.Yarns {
			for j, a := range yarn {
				t.Yarns[i][j] = a.RemapSite(localRemap)
			}
		}
		for i, a := range t.Weave {
			t.Weave[i] = a.RemapSite(localRemap)
		}
		t.Cursor = t.Cursor.RemapSite(localRemap)
		// Insert empty yarn in local position.
		t.Yarns = append(t.Yarns, nil)
		copy(t.Yarns[i+1:], t.Yarns[i:])
		t.Yarns[i] = nil
		// Insert site ID into local sitemap.
		t.Sitemap = append(t.Sitemap, uuid.Nil)
		copy(t.Sitemap[i+1:], t.Sitemap[i:])
		t.Sitemap[i] = newSiteID
	}
	// Copy data to remote tree.
	n := len(t.Sitemap)
	t.Timestamp++
	remote := &CausalTree{
		Weave:     make([]atom.Atom, len(t.Weave)),
		Cursor:    t.Cursor,
		Yarns:     make([][]atom.Atom, n),
		Sitemap:   make([]uuid.UUID, n),
		SiteID:    newSiteID,
		Timestamp: t.Timestamp,
	}
	copy(remote.Weave, t.Weave)
	for i, yarn := range t.Yarns {
		remote.Yarns[i] = make([]atom.Atom, len(yarn))
		copy(remote.Yarns[i], yarn)
	}
	copy(remote.Sitemap, t.Sitemap)
	return remote, nil
}

// +-------+
// | Merge |
// +-------+

// Time complexity: O(sites)
func mergeSitemaps(s1, s2 []uuid.UUID) []uuid.UUID {
	var i, j int
	s := make([]uuid.UUID, 0, len(s1)+len(s2))
	for i < len(s1) && j < len(s2) {
		id1, id2 := s1[i], s2[j]
		order := bytes.Compare(id1[:], id2[:])
		if order < 0 {
			s = append(s, id1)
			i++
		} else if order > 0 {
			s = append(s, id2)
			j++
		} else {
			s = append(s, id1)
			i++
			j++
		}
	}
	if i < len(s1) {
		s = append(s, s1[i:]...)
	}
	if j < len(s2) {
		s = append(s, s2[j:]...)
	}
	return s
}

// Merge updates the current state with that of another remote tree.
// Note that merge does not move the cursor.
//
// Time complexity: O(atoms^2 + sites*log(sites))
func (t *CausalTree) Merge(remote *CausalTree) {
	// 1. Merge sitemaps.
	// Time complexity: O(sites)
	sitemap := mergeSitemaps(t.Sitemap, remote.Sitemap)

	// 2. Compute site index remapping.
	// Time complexity: O(sites*log(sites))
	localRemap := make(index.Map)
	remoteRemap := make(index.Map)
	for i, site := range t.Sitemap {
		localRemap.Set(i, siteIndex(sitemap, site))
	}
	for i, site := range remote.Sitemap {
		remoteRemap.Set(i, siteIndex(sitemap, site))
	}

	// 3. Remap atoms from local.
	// Time complexity: O(atoms)
	yarns := make([][]atom.Atom, len(sitemap))
	if len(localRemap) > 0 {
		for i, yarn := range t.Yarns {
			i := localRemap.Get(i)
			yarns[i] = make([]atom.Atom, len(yarn))
			for j, a := range yarn {
				yarns[i][j] = a.RemapSite(localRemap)
			}
		}
		for i, a := range t.Weave {
			t.Weave[i] = a.RemapSite(localRemap)
		}
	} else {
		for i, yarn := range t.Yarns {
			yarns[i] = make([]atom.Atom, len(yarn))
			copy(yarns[i], yarn)
		}
	}

	// 4. Merge yarns.
	// Time complexity: O(atoms)
	for i, yarn := range remote.Yarns {
		i := remoteRemap.Get(i)
		start := len(yarns[i])
		end := len(yarn)
		if end > start {
			// Grow yarn to accomodate remote atoms.
			yarns[i] = append(yarns[i], make([]atom.Atom, end-start)...)
		}
		for j := start; j < end; j++ {
			a := yarn[j].RemapSite(remoteRemap)
			yarns[i][j] = a
		}
	}

	// 5. Merge weaves.
	// Time complexity: O(atoms)
	remoteWeave := make([]atom.Atom, len(remote.Weave))
	for i, a := range remote.Weave {
		remoteWeave[i] = a.RemapSite(remoteRemap)
	}
	t.Weave = atom.MergeWeaves(t.Weave, remoteWeave)

	// Move created stuff to this tree.
	t.Yarns = yarns
	t.Sitemap = sitemap
	if t.Timestamp < remote.Timestamp {
		t.Timestamp = remote.Timestamp
	}
	t.Timestamp++

	// 6. Fix cursor if necessary.
	// Time complexity: O(atoms^2)
	t.Cursor = t.Cursor.RemapSite(localRemap)
	t.fixDeletedCursor()
}

// Returns whether the atom is deleted.
//
// Time complexity: O(atoms), or, O(avg. block size)
func (t *CausalTree) isDeleted(atomID atom.ID) bool {
	i := t.atomIndex(atomID)
	if i < 0 {
		return false
	}
	var isDeleted bool
	atom.WalkChildren(t.Weave[i:], func(child atom.Atom) bool {
		if _, ok := child.Value.(Delete); ok {
			isDeleted = true
			return false
		}
		// There's a child with lower priority than delete, so there can't be
		// any more delete atom ahead.
		if child.Value.AtomPriority() < (Delete{}).AtomPriority() {
			isDeleted = false
			return false
		}
		return true
	})
	return isDeleted
}

// Ensure tree's cursor isn't deleted, finding the first non-deleted ancestor.
//
// Time complexity: O(atoms^2), or, O((avg. tree height) * (avg. block size))
func (t *CausalTree) fixDeletedCursor() {
	for {
		if !t.isDeleted(t.Cursor) {
			break
		}
		t.Cursor = t.getAtom(t.Cursor).Cause
	}
}

// +-------------+
// + Time travel |
// +-------------+

// The same as weft, but using yarn's indices instead of timestamps.
type indexWeft []int

// Returns whether the provided atom is present in the yarn's view.
// The nil atom is always in view.
func (ixs indexWeft) isInView(id atom.ID) bool {
	return int(id.Index) < ixs[id.Site] || id.Timestamp == 0
}

// Checks that the weft is well-formed, not disconnecting atoms from their causes
// in other sites.
//
// Time complexity: O(atoms)
func (t *CausalTree) checkWeft(w weft.Weft) (indexWeft, error) {
	if len(t.Yarns) != len(w) {
		return nil, ErrWeftInvalidLength
	}
	// Initialize limits at each yarn.
	limits := make(indexWeft, len(w))
	for i, yarn := range t.Yarns {
		limits[i] = len(yarn)
	}
	// Look for max timestamp at each yarn.
	for i, yarn := range t.Yarns {
		tmax := w[i]
		for j, a := range yarn {
			if a.ID.Timestamp > tmax {
				limits[i] = j
				break
			}
		}
	}
	// Verify that all causes are present at the weft cut.
	for i, yarn := range t.Yarns {
		limit := limits[i]
		for _, a := range yarn[:limit] {
			if !limits.isInView(a.Cause) {
				return nil, ErrWeftDisconnected
			}
		}
	}
	return limits, nil
}

// Now returns the last known time at every site as a weft.
func (t *CausalTree) Now() weft.Weft {
	w := make(weft.Weft, len(t.Yarns))
	for i, yarn := range t.Yarns {
		n := len(yarn)
		if n == 0 {
			continue
		}
		w[i] = yarn[n-1].ID.Timestamp
	}
	return w
}

// ViewAt returns a view of the tree in the provided time in the past, represented with a weft.
//
// Time complexity: O(atoms+sites)
func (t *CausalTree) ViewAt(w weft.Weft) (*CausalTree, error) {
	limits, err := t.checkWeft(w)
	if err != nil {
		return nil, err
	}
	n := len(limits)
	yarns := make([][]atom.Atom, n)
	for i, yarn := range t.Yarns {
		yarns[i] = make([]atom.Atom, limits[i])
		copy(yarns[i], yarn)
	}
	weave := make([]atom.Atom, 0, len(t.Weave))
	for _, a := range t.Weave {
		if limits.isInView(a.ID) {
			weave = append(weave, a)
		}
	}
	sitemap := make([]uuid.UUID, n)
	copy(sitemap, t.Sitemap)
	// Set cursor, if it still exists in this view.
	cursor := t.Cursor
	if !limits.isInView(cursor) {
		cursor = atom.ID{}
	}
	//
	i := siteIndex(t.Sitemap, t.SiteID)
	tmax := w[i]
	view := &CausalTree{
		Weave:     weave,
		Cursor:    cursor,
		Yarns:     yarns,
		Sitemap:   sitemap,
		SiteID:    t.SiteID,
		Timestamp: tmax,
	}
	return view, nil
}

// +---------------------+
// | Operations - Errors |
// +---------------------+

// Errors returned by CausalTree operations
var (
	ErrSiteLimitExceeded  = errors.New("reached limit of sites: 2¹⁶ (65.536)")
	ErrStateLimitExceeded = errors.New("reached limit of states: 2³² (4.294.967.296)")
	ErrNoAtomToDelete     = errors.New("can't delete empty atom")
	ErrCursorOutOfRange   = errors.New("cursor index out of range")
	ErrWeftInvalidLength  = errors.New("weft length doesn't match with number of sites")
	ErrWeftDisconnected   = errors.New("weft disconnects some atom from its cause")
)

// +------------+
// | Operations |
// +------------+

// Time complexity: O(atoms), or, O(atoms + (avg. block size))
func (t *CausalTree) insertAtomAtCursor(a atom.Atom) {
	if t.Cursor.Timestamp == 0 {
		// Cursor is at initial position.
		t.insertAtom(a, 0)
		return
	}
	// Search for position in weave that atom should be inserted, in a way that it's sorted relative to
	// other children in descending order.
	//
	//                                  causal block of cursor
	//                      ------------------------------------------------
	// Weave:           ... [cursor] [child1] ... [child2] ... [child3] ... [not child]
	// Block indices:           0         1          c2'          c3'           end'
	// Weave indices:          c0        c1          c2           c3            end
	c0 := t.atomIndex(t.Cursor)
	var pos, i int
	atom.WalkCausalBlock(t.Weave[c0:], func(a0 atom.Atom) bool {
		i++
		if a0.Cause == t.Cursor && a0.Compare(a) < 0 && pos == 0 {
			// a is the first child smaller than atom.
			pos = i
		}
		return true
	})
	index := c0 + i + 1
	if pos > 0 {
		index = c0 + pos
	}
	t.insertAtom(a, index)
}

// Inserts the atom as a child of the cursor, and returns its ID.
//
// Time complexity: O(atoms + log(sites))
func (t *CausalTree) addAtom(value atom.Value) (atom.ID, error) {
	t.Timestamp++
	if t.Timestamp == 0 {
		// Overflow
		return atom.ID{}, ErrStateLimitExceeded
	}
	if t.Cursor.Timestamp > 0 {
		cursorAtom := t.getAtom(t.Cursor)
		if err := cursorAtom.Value.ValidateChild(value); err != nil {
			return atom.ID{}, err
		}
	}
	i := siteIndex(t.Sitemap, t.SiteID)
	atomID := atom.ID{
		Site:      uint16(i),
		Index:     uint32(len(t.Yarns[i])),
		Timestamp: t.Timestamp,
	}
	a := atom.Atom{
		ID:    atomID,
		Cause: t.Cursor,
		Value: value,
	}
	t.insertAtomAtCursor(a)
	t.Yarns[i] = append(t.Yarns[i], a)
	return atomID, nil
}

// +-------------------------+
// | Operations - Set cursor |
// +-------------------------+

// Auxiliary function that checks if 'atom' is a container.
func isContainer(a atom.Atom) bool {
	switch a.Value.(type) {
	case InsertStr, InsertCounter:
		return true
	default:
		return false
	}

}

// Time complexity: O(atoms)
func (t *CausalTree) filterDeleted() []atom.Atom {
	atoms := make([]atom.Atom, len(t.Weave))
	copy(atoms, t.Weave)
	indices := make(map[atom.ID]int)
	var hasDelete bool
	for i, a := range t.Weave {
		indices[a.ID] = i
	}
	for i, a := range t.Weave {
		if _, ok := a.Value.(Delete); ok {
			hasDelete = true
			// Deletion must always come after deleted atom, so
			// indices map must have the cause location.
			deletedAtomIdx := indices[a.Cause]
			if isContainer(atoms[deletedAtomIdx]) {
				atom.DeleteDescendants(atoms, deletedAtomIdx)
			} else {
				atoms[i] = atom.Atom{}              //Delete the "Delete" atom
				atoms[deletedAtomIdx] = atom.Atom{} //Delete the target atom
			}
		}
	}
	if !hasDelete {
		// Cheap optimization for case where there are no deletions.
		return atoms
	}
	// Move chars to fill in holes of empty atoms.
	deleted := 0
	for i, a := range atoms {
		if a.ID.Timestamp == 0 {
			deleted++
		} else {
			atoms[i-deleted] = atoms[i]
		}
	}
	atoms = atoms[:len(atoms)-deleted]
	return atoms
}

// Sets cursor to the given (tree) position.
//
// To insert an atom at the beginning, use i = -1.
func (t *CausalTree) SetCursor(i int) error {
	if i < 0 {
		if i == -1 {
			t.Cursor = atom.ID{}
			return nil
		}
		return ErrCursorOutOfRange
	}
	atoms := t.filterDeleted()
	if i >= len(atoms) {
		return ErrCursorOutOfRange
	}
	t.Cursor = atoms[i].ID
	return nil
}

// +------------+
// | Conversion |
// +------------+

// ToString returns a string representation of the CausalTree t.
func (t *CausalTree) ToString() string {
	var data []interface{}
	jsonData, err := t.ToJSON()
	if err != nil {
		panic(fmt.Sprintf("ToString: %v", err))
	}
	err = json.Unmarshal(jsonData, &data)
	if err != nil {
		panic(fmt.Sprintf("ToString: %v", err))
	}

	return conversion.ToString(data)

}

// this interface represents a generic type.
type generic interface{}

// ToJSON interprets tree as a JSON.
func (t *CausalTree) ToJSON() ([]byte, error) {
	tab := "    "
	atoms := t.filterDeleted()
	var elements []generic
	for i := 0; i < len(atoms); {
		currentAtomValue := atoms[i].Value
		switch value := currentAtomValue.(type) {
		case InsertChar:
			elements = append(elements, string(value.Char))
			i++
		case InsertStr:
			strSize := atom.CausalBlockSize(atoms[i:]) - 1
			strChars := make([]rune, strSize)

			for j, a := range atoms[i+1 : i+strSize+1] {
				strChars[j] = a.Value.(InsertChar).Char
			}
			elements = append(elements, string(strChars))
			i = i + strSize + 1
		case InsertCounter:
			counterSize := atom.CausalBlockSize(atoms[i:]) - 1
			var counterValue int32 = 0

			for _, a := range atoms[i+1 : i+counterSize+1] {
				counterValue += a.Value.(InsertAdd).Value
			}
			elements = append(elements, counterValue)
			i = i + counterSize + 1
		default:
			return nil, fmt.Errorf("ToJSON: type not specified")
		}
	}

	finalJSON, err := json.MarshalIndent(elements, "", tab)
	if err != nil {
		panic(fmt.Sprintf("ToJSON: %v", err))
	}
	return finalJSON, nil
}
