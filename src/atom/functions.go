package atom

// -----

// Invokes the closure f with each atom of the causal block. Returns the number of atoms visited.
//
// The closure should return 'false' to cut the traversal short, as in a 'break' statement. Otherwise, return true.
//
// The causal block is defined as the contiguous range containing the head and all of its descendents.
//
// Time complexity: O(atoms), or, O(avg. block size)
func WalkCausalBlock(block []Atom, f func(Atom) bool) int {
	if len(block) == 0 {
		return 0
	}
	head := block[0]
	for i, atom := range block[1:] {
		if atom.Cause.Timestamp < head.ID.Timestamp {
			// First atom whose parent has a lower timestamp (older) than head is the
			// end of the causal block.
			return i + 1
		}
		if !f(atom) {
			break
		}
	}
	return len(block)
}

// Invokes the closure f with each direct children of the block's head.
//
// The index i corresponds to the index on the causal block, not on the child's order.
func WalkChildren(block []Atom, f func(Atom) bool) {
	WalkCausalBlock(block, func(atom Atom) bool {
		if atom.Cause == block[0].ID {
			return f(atom)
		}
		return true
	})
}

// Returns the size of the causal block, including its head.
func CausalBlockSize(block []Atom) int {
	return WalkCausalBlock(block, func(atom Atom) bool { return true })
}

// Time complexity: O(atoms)
func MergeWeaves(w1, w2 []Atom) []Atom {
	var i, j int
	var weave []Atom
	for i < len(w1) && j < len(w2) {
		a1, a2 := w1[i], w2[j]
		if a1 == a2 {
			// Atoms are equal, append it to the weave.
			weave = append(weave, a1)
			i++
			j++
			continue
		}
		if a1.ID.Site == a2.ID.Site {
			// Atoms are from the same site and can be compared by timestamp.
			// Insert younger one (larger timestamp) in weave.
			if a1.ID.Timestamp < a2.ID.Timestamp {
				weave = append(weave, a2)
				j++
			} else {
				weave = append(weave, a1)
				i++
			}
		} else {
			// Atoms are concurrent; append first causal block, according to heads' order.
			if a1.Compare(a2) >= 0 {
				n1 := i + CausalBlockSize(w1[i:])
				weave = append(weave, w1[i:n1]...)
				i = n1
			} else {
				n2 := j + CausalBlockSize(w2[j:])
				weave = append(weave, w2[j:n2]...)
				j = n2
			}
		}
	}
	if i < len(w1) {
		weave = append(weave, w1[i:]...)
	}
	if j < len(w2) {
		weave = append(weave, w2[j:]...)
	}
	return weave
}

// Deletes all the descendants of atom into the weave.
// Time complexity: O(len(block))
func DeleteDescendants(block []Atom, atomIndex int) {
	causalBlockSz := CausalBlockSize(block[atomIndex:])
	for i := 0; i < causalBlockSz; i++ {
		block[atomIndex+i] = Atom{}
	}
}
