/*
	Package crdt provides primitives to operate on replicated data types.

	Replicated data types are structured such that they can be copied across multiple sites
	in a distributed environment, mutated independently at each site, and they still may be
	merged back without conflicts.

	This implementation is based on the Causal Tree structure proposed by Victor Grishchenko [1],
	following the excellent explanation by Archagon [2].

	[1]: GRISCHENKO, VICTOR. Causal trees: towards real-time read-write hypertext.
	[2]: http://archagon.net/blog/2018/03/24/data-laced-with-history/

	In a causal tree, each operation on the data structure is represented by an atom, which
	has a single other atom as its cause, thus creating a tree shape.
	When inserting a char, for example, its cause is the char to its left.

	# BEGIN ASCII ART

	T <- H <- I <- S <- _ <- I <- S <- _ <- N <- I <- C <- E
										^
										'-- V <- E <- R <- Y <- _

	# END ASCII ART
	# ALT TEXT: Sequence of letters with arrows between them, representing atoms and their causes.
				In the first line it reads "THIS_IS_NICE", and in the second the string "VERY_" points
				to the space after "IS". This represents an insertion, thus the whole tree should be
				read as "THIS_IS_VERY_NICE".

	Instead of using a pointer to reference the causing operation, references simply hold an atom
	ID containing the origin site and the (local) timestamp of creation.
	Atoms are then organized in an array to improve memory locality, at the expanse of having to
	search linearly for a given atom ID.

	By sorting the array such that atoms from the same site and time are mostly contiguous, this search
	operation is not terribly costly, and the array reads almost like the structure being represented.

	# BEGIN ASCII ART

	id cause                                  .--------------------------------------.
	|  |                                      .--------.                             |
	v  v                                      v        |                             |
	.-----.-----.-----.-----.-----.-----.-----.-----.-----.-----.-----.-----.-----.-----.-----.-----.-----.
	|01|__|02|01|03|02|04|03|05|04|06|05|07|06|08|07|13|08|14|13|15|14|16|15|17|16|09|08|10|09|11|10|12|11|
	+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+
	|  T  |  H  |  I  |  S  |  _  |  I  |  S  |  _  |  V  |  E  |  R  |  Y  |  _  |  N  |  I  |  C  |  E  |
	'-----'-----'-----'-----'-----'-----'-----'-----'-----'-----'-----'-----'-----'-----'-----'-----'-----'

	# END ASCII ART
	# ALT TEXT: Sequence of atoms in an array, representing the same string as before. Each atom is composed
				of three elements: ID, cause, and content. IDs are given sequentially, from 01 to 17.
				From 01 to 08, all atoms are sorted by ID, with contents that spell "THIS_IS_".
				To the right of 08, we have the sequence of IDs from 13 to 17, with contents that spell
				"VERY_". The cause of ID 13 is the atom 08.
				Finally, to the right of ID 17 is the sequence of atoms from 09 to 12, with contents that
				spell "NICE". The cause of ID 09 is also the atom 08.
				The first atom has no cause, and all others have as cause the atom to its left.
*/
package causaltree
