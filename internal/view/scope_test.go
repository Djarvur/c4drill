package view_test

import (
	"testing"

	"github.com/Djarvur/c4drill/internal/model"
	"github.com/Djarvur/c4drill/internal/parser"
	"github.com/Djarvur/c4drill/internal/view"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateC1View_ReturnsViewWithLevelC1(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test System"},
		Units:      map[string]*model.Unit{},
	}

	v := view.GenerateC1View(m)

	require.NotNil(t, v)
	assert.Equal(t, view.LevelC1, v.Level)
}

func TestGenerateC1View_IncludesAllTopLevelUnits(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test System"},
		Units: map[string]*model.Unit{
			"api": {
				Type: model.TypeSystem,
				Name: "API System",
			},
			"db": {
				Type: model.TypeDb,
				Name: "Database",
			},
			"user": {
				Type: model.TypePerson,
				Name: "User",
			},
		},
	}

	v := view.GenerateC1View(m)

	require.NotNil(t, v)
	require.Len(t, v.Units, 3)
	assert.Contains(t, v.Units, "api")
	assert.Contains(t, v.Units, "db")
	assert.Contains(t, v.Units, "user")
}

func TestGenerateC1View_TitleFromProperties(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "My Architecture"},
		Units:      map[string]*model.Unit{},
	}

	v := view.GenerateC1View(m)

	require.NotNil(t, v)
	assert.Equal(t, "My Architecture", v.Title)
}

func TestGenerateC1View_EdgesFromProperties(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Edges: "spline"},
		Units:      map[string]*model.Unit{},
	}

	v := view.GenerateC1View(m)

	require.NotNil(t, v)
	assert.Equal(t, "spline", v.Edges)
}

func TestGenerateC1View_FullPathEqualsUnitName(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"mainsystem": {
				Type: model.TypeSystem,
				Name: "Main System",
			},
		},
	}

	v := view.GenerateC1View(m)

	require.NotNil(t, v)
	require.Contains(t, v.Units, "mainsystem")
	assert.Equal(t, "mainsystem", v.Units["mainsystem"].FullPath)
}

func TestGenerateC1View_HasSubunitsTrueWhenUnitHasSubunits(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"system": {
				Type: model.TypeSystem,
				Name: "System",
				Subunits: map[string]*model.Unit{
					"api": {Type: model.TypeContainer},
				},
			},
			"db": {
				Type: model.TypeDb,
				Name: "Database",
			},
		},
	}

	v := view.GenerateC1View(m)

	require.NotNil(t, v)
	require.Contains(t, v.Units, "system")
	require.Contains(t, v.Units, "db")
	assert.True(t, v.Units["system"].HasSubunits)
	assert.False(t, v.Units["db"].HasSubunits)
}

func TestGenerateC1View_IsExternalBasedOnType(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"internal": {
				Type: model.TypeSystem,
				Name: "Internal",
			},
			"external": {
				Type: model.TypeSystemExternal,
				Name: "External",
			},
		},
	}

	v := view.GenerateC1View(m)

	require.NotNil(t, v)
	require.Contains(t, v.Units, "internal")
	require.Contains(t, v.Units, "external")
	assert.False(t, v.Units["internal"].IsExternal)
	assert.True(t, v.Units["external"].IsExternal)
}

func TestGenerateC1View_ExternalBoundaryNodesForReferencedUnits(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"api": {
				Type: model.TypeSystem,
				Name: "API",
				Links: map[string]model.Link{
					"externaldb": {Target: "externaldb"},
				},
			},
		},
	}

	v := view.GenerateC1View(m)

	require.NotNil(t, v)
	// Should have both the defined unit and the external boundary node
	assert.Contains(t, v.Units, "api")
	assert.Contains(t, v.Units, "externaldb")

	// External boundary node should be marked external
	assert.True(t, v.Units["externaldb"].IsExternal)
	assert.True(t, v.Units["externaldb"].Unit.Type == model.TypeSystemExternal)
}

func TestGenerateC1View_ExternalBoundaryFromLinksFrom(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"api": {
				Type: model.TypeSystem,
				Name: "API",
				LinksFrom: map[string]model.Link{
					"externaluser": {Target: "externaluser"},
				},
			},
		},
	}

	v := view.GenerateC1View(m)

	require.NotNil(t, v)
	assert.Contains(t, v.Units, "externaluser")
	assert.True(t, v.Units["externaluser"].IsExternal)
}

func TestGenerateC1View_NoDuplicateExternalBoundaryNodes(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"api": {
				Type: model.TypeSystem,
				Name: "API",
				Links: map[string]model.Link{
					"externaldb": {Target: "externaldb"},
				},
			},
			"web": {
				Type: model.TypeSystem,
				Name: "Web",
				Links: map[string]model.Link{
					"externaldb": {Target: "externaldb"},
				},
			},
		},
	}

	v := view.GenerateC1View(m)

	require.NotNil(t, v)
	// Should only have one external boundary node for externaldb
	assert.Contains(t, v.Units, "externaldb")
	assert.Len(t, v.Units, 3) // api, web, externaldb
}

func TestGenerateC1View_IsExpandedWhenUnitExpandsSelf(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"system": {
				Type:     model.TypeSystem,
				Name:     "System",
				Expanded: []string{"system"}, // Unit expands itself
			},
			"db": {
				Type: model.TypeDb,
				Name: "Database",
			},
		},
	}

	v := view.GenerateC1View(m)

	require.NotNil(t, v)
	assert.True(t, v.Units["system"].IsExpanded)
	assert.False(t, v.Units["db"].IsExpanded)
}

func TestGenerateC1View_NilModelReturnsNil(t *testing.T) {
	t.Parallel()

	v := view.GenerateC1View(nil)
	assert.Nil(t, v)
}

func TestGenerateC1View_EmptyModelReturnsEmptyView(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Empty"},
		Units:      map[string]*model.Unit{},
	}

	v := view.GenerateC1View(m)

	require.NotNil(t, v)
	assert.Empty(t, v.Units)
}
