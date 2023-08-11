package causal_tree

import (
	"github.com/crdteam/causal-tree/crdt/atom"
	"github.com/google/uuid"
)

// New creates an initialized empty replicated tree.
func New() *CausalTree {
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
