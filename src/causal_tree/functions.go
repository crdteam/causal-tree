package causal_tree

import (
	atm "github.com/crdteam/causal-tree/src/atom"
	"github.com/google/uuid"
)

// NewCausalTree creates an initialized empty replicated tree.
func NewCausalTree() *CausalTree {
	siteID := uuidv1()
	return &CausalTree{
		Weave:     nil,
		Cursor:    atm.AtomID{},
		Yarns:     [][]atm.Atom{nil},
		Sitemap:   []uuid.UUID{siteID},
		SiteID:    siteID,
		Timestamp: 1, // Timestamp 0 is considered invalid.
	}
}
