package weft

import "fmt"

// Weft is a clock that stores the timestamp of each site of a CausalTree.
//
// In a distributed system it's not possible to observe the whole state at an absolute time,
// but we can view the site's state at each site time.
type Weft []uint32

// Compare returns -1, +1 and 0 if this is weft is less than, greater than, or concurrent
// to the other, respectively.
//
// It panics if wefts have different sizes.
func (w Weft) Compare(other Weft) int {
	if len(w) != len(other) {
		panic(fmt.Sprintf("wefts have different sizes: %d (%v) != %d (%v)", len(w), w, len(other), other))
	}
	var hasLess, hasGreater bool
	for i, t1 := range w {
		t2 := other[i]
		if t1 < t2 {
			hasLess = true
		} else if t1 > t2 {
			hasGreater = true
		}
	}
	if hasLess && hasGreater {
		return 0
	}
	if hasLess {
		return -1
	}
	if hasGreater {
		return +1
	}
	return 0
}
