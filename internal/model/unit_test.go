package model_test

import (
	"testing"

	"github.com/Djarvur/c4drill/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestClonePreservesMirror is the HS-1 load-bearing test (Plan 31-02): the
// validator mutates Unit.LinksFrom in place (internal/validator/index.go:70-81)
// appending Link{Mirror:true} entries. A Clone that drops Mirror (e.g. via
// reflect/gob/json) corrupts multiplicity counting for every instantiation
// after the first. This test proves the hand-rolled Clone preserves Mirror on
// BOTH Links and LinksFrom, and that the clone's slices do not share backing
// arrays with the original (append to one must not affect the other).
func TestClonePreservesMirror(t *testing.T) {
	t.Parallel()

	orig := &model.Unit{
		Type: model.TypeContainer,
		Name: "Original",
		Links: []model.Link{
			{Peer: "target", Technology: "HTTP"},
		},
		LinksFrom: []model.Link{
			{Peer: "source", Mirror: true},
		},
	}

	clone := orig.Clone()
	require.NotNil(t, clone, "Clone must return a non-nil unit")

	// Mirror preserved on LinksFrom.
	require.Len(t, clone.LinksFrom, 1, "clone LinksFrom length")
	assert.True(t, clone.LinksFrom[0].IsMirror(), "clone LinksFrom[0].Mirror must be preserved (HS-1)")
	assert.Equal(t, "source", clone.LinksFrom[0].Peer, "clone LinksFrom[0].Peer")

	// Mirror=false on Links also preserved (round-trip).
	require.Len(t, clone.Links, 1, "clone Links length")
	assert.False(t, clone.Links[0].IsMirror(), "clone Links[0].Mirror (authored, not mirror)")
	assert.Equal(t, "target", clone.Links[0].Peer, "clone Links[0].Peer")

	// Backing-array disjointness: appending to the clone's LinksFrom must NOT
	// affect the original (a shallow slice copy would share the header).
	clone.LinksFrom = append(clone.LinksFrom, model.Link{Peer: "extra", Mirror: true})
	assert.Len(t, orig.LinksFrom, 1, "original LinksFrom must be unchanged after clone append")
	assert.Len(t, clone.LinksFrom, 2, "clone LinksFrom must reflect the append")
}

// TestCloneRecursesSubunits verifies Clone deep-copies the Subunits map with
// every child *Unit independently cloned (pointer-disjoint). Modifying a
// clone's child must not affect the original's child. SubunitOrder is cloned.
func TestCloneRecursesSubunits(t *testing.T) {
	t.Parallel()

	orig := &model.Unit{
		Type: model.TypeSystem,
		Name: "Parent",
		SubunitOrder: []string{"api", "db"},
		Subunits: map[string]*model.Unit{
			"api": {
				Type: model.TypeContainer,
				Name: "API",
				Links: []model.Link{
					{Peer: "db"},
				},
			},
			"db": {
				Type: model.TypeContainerDb,
				Name: "Database",
			},
		},
	}

	clone := orig.Clone()
	require.NotNil(t, clone, "Clone must return a non-nil unit")

	// SubunitOrder cloned (disjoint backing array).
	require.Equal(t, []string{"api", "db"}, clone.SubunitOrder, "clone SubunitOrder")
	clone.SubunitOrder = append(clone.SubunitOrder, "extra")
	assert.Len(t, orig.SubunitOrder, 2, "original SubunitOrder unchanged after clone append")

	// Subunits map is a fresh map with pointer-disjoint children.
	require.Contains(t, clone.Subunits, "api", "clone has 'api' subunit")
	require.Contains(t, clone.Subunits, "db", "clone has 'db' subunit")

	cloneAPI := clone.Subunits["api"]
	origAPI := orig.Subunits["api"]
	require.NotNil(t, cloneAPI, "clone api non-nil")
	require.NotNil(t, origAPI, "orig api non-nil")
	assert.NotSame(t, origAPI, cloneAPI, "clone child must be a different pointer (pointer-disjoint)")

	// Mutate the clone's child; original's child must be unaffected.
	cloneAPI.Name = "Mutated API"
	assert.Equal(t, "API", origAPI.Name, "original child Name unchanged after clone mutation")

	// Child Links are independent (backing-array disjoint).
	require.Len(t, cloneAPI.Links, 1, "clone child Links length")
	cloneAPI.Links[0].Peer = "mutated-peer"
	assert.Equal(t, "db", origAPI.Links[0].Peer, "original child Link.Peer unchanged")
}

// TestCloneNilSafe verifies Clone on a nil *Unit returns nil (no panic).
func TestCloneNilSafe(t *testing.T) {
	t.Parallel()

	var nilUnit *model.Unit
	clone := nilUnit.Clone()
	assert.Nil(t, clone, "Clone of nil *Unit must return nil")
}

// TestClonePreservesAllValueFields verifies Clone copies every value-type field
// of Unit (Type, Name, Description, Technology, Reference, Color, Style, etc.).
func TestClonePreservesAllValueFields(t *testing.T) {
	t.Parallel()

	orig := &model.Unit{
		Type:        model.TypeComponent,
		Name:        "Comp",
		Description: "desc",
		Technology:  "Go",
		Reference:   "https://example.com",
		Color:       "blue",
		Style:       "rounded",
		Border:      "solid",
		Edges:       "spline",
		Width:       3.5,
		Height:      2.0,
		Expanded:    []string{"a", "b"},
	}

	clone := orig.Clone()
	require.NotNil(t, clone)

	assert.Equal(t, orig.Type, clone.Type, "Type")
	assert.Equal(t, orig.Name, clone.Name, "Name")
	assert.Equal(t, orig.Description, clone.Description, "Description")
	assert.Equal(t, orig.Technology, clone.Technology, "Technology")
	assert.Equal(t, orig.Reference, clone.Reference, "Reference")
	assert.Equal(t, orig.Color, clone.Color, "Color")
	assert.Equal(t, orig.Style, clone.Style, "Style")
	assert.Equal(t, orig.Border, clone.Border, "Border")
	assert.Equal(t, orig.Edges, clone.Edges, "Edges")
	assert.Equal(t, orig.Width, clone.Width, "Width")
	assert.Equal(t, orig.Height, clone.Height, "Height")
	assert.Equal(t, orig.Expanded, clone.Expanded, "Expanded")

	// Expanded backing-array disjoint.
	clone.Expanded = append(clone.Expanded, "c")
	assert.Len(t, orig.Expanded, 2, "original Expanded unchanged after clone append")
}
