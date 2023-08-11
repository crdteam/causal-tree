package causal_tree

import (
	"github.com/crdteam/causal-tree/crdt/atom"
	"github.com/google/uuid"
)

// NewCausalTree creates an initialized empty replicated tree.
func NewCausalTree() *CausalTree {
	siteID := uuidv1()
	return &CausalTree{
		Weave:     nil,
		Cursor:    atom.ID{},
		Yarns:     [][]atom.Atom{nil},
		Sitemap:   []uuid.UUID{siteID},
		SiteID:    siteID,
		Timestamp: 1, // Timestamp 0 is considered invalid.
	}
}
