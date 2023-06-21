package atom

// -----

/*
	WalkCausalBlock traverses a block of atoms and applies a given function to
	each atom in the causal block. It returns the number of atoms visited, including
	the block head, during
	its execution (in other words, the number of atoms in which the closure
	function returned true). The closure function should return false to
	cut the traversal short, as in a break statement. Otherwise, return true.


	Stop Condition

	The function will stop under two conditions:

	1. If the function (f) applied to an atom returns 'false'. This is similar
	to a 'break' statement in a loop.

	2. If it encounters an atom whose cause's timestamp is less than the head's
	timestamp. This is because such an atom is not considered part of the causal
	block.

	Parameters

	- block ([]Atom): a slice of atoms. The causal block is defined as the contiguous range
	containing the head and all of its descendants. The head of the block is the first atom
	in the slice, and descendants are those atoms whose cause's timestamp is equal or greater
	than the head's timestamp.

	- f (func(Atom) bool): a function that takes an atom as input and returns a boolean

	Returns

	- (int): the number of atoms visited during the execution of the function. If the block is empty,
	the function returns 0.

	Time Complexity

	O(atoms), or equivalently O(avg. block size)

*/
func WalkCausalBlock(block []Atom, f func(Atom) bool) int {
	if len(block) == 0 {
		return 0
	}
	head := block[0]
	i := 1
	for ; i < len(block); i++ {
		atom := block[i]
		if atom.Cause.Timestamp < head.ID.Timestamp {
			// First atom whose parent has a lower timestamp (older) than head is the
			// end of the causal block.
			return i
		}
		if !f(atom) {
			break
		}
	}
	return i
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
