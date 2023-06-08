package indexmap

// +---------------+
// | Remap indices |
// +---------------+
// TODO separate this type IndexMap to another package: its own package: indexmap

// Map storing conversion between indices.
// Conversion from an index to itself are not stored.
// An empty map represents an identity mapping, where every index maps to itself.
// Queries for an index that was not inserted or stored return the same index.
type IndexMap map[int]int

func (m IndexMap) Set(i, j int) {
	if i != j {
		m[i] = j
	}
}

func (m IndexMap) Get(i int) int {
	j, ok := m[i]
	if !ok {
		return i
	}
	return j
}

// -----
