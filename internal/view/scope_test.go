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
				Links: []model.Link{
					{Peer: "externaldb"},
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
	assert.Equal(t, model.TypeSystemExternal, v.Units["externaldb"].Unit.Type)
}

func TestGenerateC1View_ExternalBoundaryFromLinksFrom(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"api": {
				Type: model.TypeSystem,
				Name: "API",
				LinksFrom: []model.Link{
					{Peer: "externaluser"},
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
				Links: []model.Link{
					{Peer: "externaldb"},
				},
			},
			"web": {
				Type: model.TypeSystem,
				Name: "Web",
				Links: []model.Link{
					{Peer: "externaldb"},
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

// Tests for GenerateC2View

func TestGenerateC2View_ReturnsViewWithLevelC2(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"system": {
				Type: model.TypeSystem,
				Name: "System",
				Subunits: map[string]*model.Unit{
					"api": {Type: model.TypeContainer, Name: "API"},
				},
			},
		},
	}

	v := view.GenerateC2View(m, "system")

	require.NotNil(t, v)
	assert.Equal(t, view.LevelC2, v.Level)
}

func TestGenerateC2View_ContainsSubunitsOfExpandedSystem(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"system": {
				Type: model.TypeSystem,
				Name: "System",
				Subunits: map[string]*model.Unit{
					"api": {Type: model.TypeContainer, Name: "API"},
					"web": {Type: model.TypeContainer, Name: "Web"},
					"db":  {Type: model.TypeContainerDb, Name: "Database"},
				},
			},
		},
	}

	v := view.GenerateC2View(m, "system")

	require.NotNil(t, v)
	require.Len(t, v.Units, 3)
	assert.Contains(t, v.Units, "system.api")
	assert.Contains(t, v.Units, "system.web")
	assert.Contains(t, v.Units, "system.db")
}

func TestGenerateC2View_ParentSetToSystemPath(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"system": {
				Type: model.TypeSystem,
				Name: "System",
				Subunits: map[string]*model.Unit{
					"api": {Type: model.TypeContainer, Name: "API"},
				},
			},
		},
	}

	v := view.GenerateC2View(m, "system")

	require.NotNil(t, v)
	assert.Equal(t, "system", v.Parent)
}

func TestGenerateC2View_ExpandedUnitSetToSystemPath(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"system": {
				Type: model.TypeSystem,
				Name: "System",
				Subunits: map[string]*model.Unit{
					"api": {Type: model.TypeContainer, Name: "API"},
				},
			},
		},
	}

	v := view.GenerateC2View(m, "system")

	require.NotNil(t, v)
	assert.Equal(t, "system", v.ExpandedUnit)
}

func TestGenerateC2View_TitleIncludesSystemName(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"mainsystem": {
				Type: model.TypeSystem,
				Name: "Main System",
				Subunits: map[string]*model.Unit{
					"api": {Type: model.TypeContainer, Name: "API"},
				},
			},
		},
	}

	v := view.GenerateC2View(m, "mainsystem")

	require.NotNil(t, v)
	assert.Contains(t, v.Title, "Main System")
}

func TestGenerateC2View_EdgesFromExpandedUnit(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test", Edges: "spline"},
		Units: map[string]*model.Unit{
			"system": {
				Type:  model.TypeSystem,
				Name:  "System",
				Edges: "straight",
				Subunits: map[string]*model.Unit{
					"api": {Type: model.TypeContainer, Name: "API"},
				},
			},
		},
	}

	v := view.GenerateC2View(m, "system")

	require.NotNil(t, v)
	assert.Equal(t, "straight", v.Edges) // From expanded unit, not properties
}

func TestGenerateC2View_NilModelReturnsNil(t *testing.T) {
	t.Parallel()

	v := view.GenerateC2View(nil, "system")
	assert.Nil(t, v)
}

func TestGenerateC2View_NonExistentPathReturnsNil(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units:      map[string]*model.Unit{},
	}

	v := view.GenerateC2View(m, "nonexistent")
	assert.Nil(t, v)
}

// Tests for GenerateC3View

func TestGenerateC3View_ReturnsViewWithLevelC3(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"system": {
				Type: model.TypeSystem,
				Name: "System",
				Subunits: map[string]*model.Unit{
					"api": {
						Type: model.TypeContainer,
						Name: "API",
						Subunits: map[string]*model.Unit{
							"handler": {Type: model.TypeComponent, Name: "Handler"},
						},
					},
				},
			},
		},
	}

	v := view.GenerateC3View(m, "system.api")

	require.NotNil(t, v)
	assert.Equal(t, view.LevelC3, v.Level)
}

func TestGenerateC3View_ContainsSubunitsOfExpandedContainer(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"system": {
				Type: model.TypeSystem,
				Name: "System",
				Subunits: map[string]*model.Unit{
					"api": {
						Type: model.TypeContainer,
						Name: "API",
						Subunits: map[string]*model.Unit{
							"handler": {Type: model.TypeComponent, Name: "Handler"},
							"service": {Type: model.TypeComponent, Name: "Service"},
							"repo":    {Type: model.TypeComponentDb, Name: "Repository"},
						},
					},
				},
			},
		},
	}

	v := view.GenerateC3View(m, "system.api")

	require.NotNil(t, v)
	require.Len(t, v.Units, 3)
	assert.Contains(t, v.Units, "system.api.handler")
	assert.Contains(t, v.Units, "system.api.service")
	assert.Contains(t, v.Units, "system.api.repo")
}

func TestGenerateC3View_FullPathIsParentChildFormat(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"system": {
				Type: model.TypeSystem,
				Name: "System",
				Subunits: map[string]*model.Unit{
					"api": {
						Type: model.TypeContainer,
						Name: "API",
						Subunits: map[string]*model.Unit{
							"handler": {Type: model.TypeComponent, Name: "Handler"},
						},
					},
				},
			},
		},
	}

	v := view.GenerateC3View(m, "system.api")

	require.NotNil(t, v)
	require.Contains(t, v.Units, "system.api.handler")
	assert.Equal(t, "system.api.handler", v.Units["system.api.handler"].FullPath)
}

func TestGenerateC3View_ExternalBoundaryFromSubunitLinks(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"system": {
				Type: model.TypeSystem,
				Name: "System",
				Subunits: map[string]*model.Unit{
					"api": {
						Type: model.TypeContainer,
						Name: "API",
						Subunits: map[string]*model.Unit{
							"handler": {
								Type: model.TypeComponent,
								Name: "Handler",
								Links: []model.Link{
									{Peer: "externalservice"},
								},
							},
						},
					},
				},
			},
		},
	}

	v := view.GenerateC3View(m, "system.api")

	require.NotNil(t, v)
	assert.Contains(t, v.Units, "externalservice")
	assert.True(t, v.Units["externalservice"].IsExternal)
}

func TestGenerateC3View_EdgesFromExpandedUnit(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"system": {
				Type:  model.TypeSystem,
				Name:  "System",
				Edges: "spline",
				Subunits: map[string]*model.Unit{
					"api": {
						Type:  model.TypeContainer,
						Name:  "API",
						Edges: "square",
						Subunits: map[string]*model.Unit{
							"handler": {Type: model.TypeComponent, Name: "Handler"},
						},
					},
				},
			},
		},
	}

	v := view.GenerateC3View(m, "system.api")

	require.NotNil(t, v)
	assert.Equal(t, "square", v.Edges) // From expanded container, not parent system
}

func TestGenerateC3View_NilModelReturnsNil(t *testing.T) {
	t.Parallel()

	v := view.GenerateC3View(nil, "system.api")
	assert.Nil(t, v)
}

func TestGenerateC3View_NonExistentPathReturnsNil(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units:      map[string]*model.Unit{},
	}

	v := view.GenerateC3View(m, "nonexistent.path")
	assert.Nil(t, v)
}

func TestGenerateC2View_ExternalBoundaryFromSubunitLinks(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"system": {
				Type: model.TypeSystem,
				Name: "System",
				Subunits: map[string]*model.Unit{
					"api": {
						Type: model.TypeContainer,
						Name: "API",
						Links: []model.Link{
							{Peer: "externaldb"},
						},
					},
				},
			},
		},
	}

	v := view.GenerateC2View(m, "system")

	require.NotNil(t, v)
	assert.Contains(t, v.Units, "externaldb")
	assert.True(t, v.Units["externaldb"].IsExternal)
}

func TestGenerateC2View_IsExpandedForChildUnits(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"system": {
				Type:     model.TypeSystem,
				Name:     "System",
				Expanded: []string{"api"}, // Expand the api subunit
				Subunits: map[string]*model.Unit{
					"api": {
						Type: model.TypeContainer,
						Name: "API",
						Subunits: map[string]*model.Unit{
							"handler": {Type: model.TypeComponent},
						},
					},
					"web": {Type: model.TypeContainer, Name: "Web"},
				},
			},
		},
	}

	v := view.GenerateC2View(m, "system")

	require.NotNil(t, v)
	assert.True(t, v.Units["system.api"].IsExpanded)
	assert.False(t, v.Units["system.web"].IsExpanded)
}

// Tests for GenerateExpandedView

func TestGenerateExpandedView_NilModelReturnsNil(t *testing.T) {
	t.Parallel()

	// Test 1: GenerateExpandedView returns nil for nil model
	v := view.GenerateExpandedView(nil)
	assert.Nil(t, v)
}

func TestGenerateExpandedView_IncludesAllTopLevelUnits(t *testing.T) {
	t.Parallel()

	// Test 2: GenerateExpandedView includes all top-level units
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

	v := view.GenerateExpandedView(m)

	require.NotNil(t, v)
	require.Len(t, v.Units, 3)
	assert.Contains(t, v.Units, "api")
	assert.Contains(t, v.Units, "db")
	assert.Contains(t, v.Units, "user")
}

func TestGenerateExpandedView_RecursivelyIncludesNestedSubunits(t *testing.T) {
	t.Parallel()

	// Test 3: GenerateExpandedView recursively includes nested subunits
	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"mainapp": {
				Type: model.TypeSystem,
				Name: "Main App",
				Subunits: map[string]*model.Unit{
					"api": {
						Type: model.TypeContainer,
						Name: "API Container",
						Subunits: map[string]*model.Unit{
							"handler": {
								Type: model.TypeComponent,
								Name: "Handler",
							},
							"service": {
								Type: model.TypeComponent,
								Name: "Service",
							},
						},
					},
					"web": {
						Type: model.TypeContainer,
						Name: "Web Container",
					},
				},
			},
			"externaldb": {
				Type: model.TypeDbExternal,
				Name: "External DB",
			},
		},
	}

	v := view.GenerateExpandedView(m)

	require.NotNil(t, v)

	// Should include top-level units
	assert.Contains(t, v.Units, "mainapp")
	assert.Contains(t, v.Units, "externaldb")

	// Should include second-level subunits
	assert.Contains(t, v.Units, "mainapp.api")
	assert.Contains(t, v.Units, "mainapp.web")

	// Should include third-level subunits
	assert.Contains(t, v.Units, "mainapp.api.handler")
	assert.Contains(t, v.Units, "mainapp.api.service")

	// Verify full paths are correct
	assert.Equal(t, "mainapp", v.Units["mainapp"].FullPath)
	assert.Equal(t, "mainapp.api", v.Units["mainapp.api"].FullPath)
	assert.Equal(t, "mainapp.api.handler", v.Units["mainapp.api.handler"].FullPath)
}

func TestGenerateExpandedView_SkipsExternalBoundaryNodesForLinkedUnits(t *testing.T) {
	t.Parallel()

	// D-12: GenerateExpandedView no longer synthesizes boundary nodes for
	// linked units missing from the model — the validator is the single
	// gatekeeper for undefined peers (internal/validator/rules.go).
	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"api": {
				Type: model.TypeSystem,
				Name: "API",
				Subunits: map[string]*model.Unit{
					"handler": {
						Type: model.TypeComponent,
						Name: "Handler",
						Links: []model.Link{
							{Peer: "cloudstorage"},
						},
					},
				},
			},
		},
	}

	v := view.GenerateExpandedView(m)

	require.NotNil(t, v)

	// Should include all nested units
	assert.Contains(t, v.Units, "api")
	assert.Contains(t, v.Units, "api.handler")

	// Should NOT include an external boundary node for the linked unit
	assert.NotContains(t, v.Units, "cloudstorage")
}

func TestGenerateExpandedView_HasSubunitsReflectsActualState(t *testing.T) {
	t.Parallel()

	// Additional test: HasSubunits reflects actual state at each level
	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"system": {
				Type: model.TypeSystem,
				Name: "System",
				Subunits: map[string]*model.Unit{
					"container": {
						Type: model.TypeContainer,
						Name: "Container",
						Subunits: map[string]*model.Unit{
							"component": {
								Type: model.TypeComponent,
								Name: "Component",
							},
						},
					},
				},
			},
		},
	}

	v := view.GenerateExpandedView(m)

	require.NotNil(t, v)

	// system has subunits
	assert.True(t, v.Units["system"].HasSubunits)

	// container has subunits
	assert.True(t, v.Units["system.container"].HasSubunits)

	// component has no subunits
	assert.False(t, v.Units["system.container.component"].HasSubunits)
}

func TestGenerateExpandedView_IsExpandedTrueWhenHasSubunits(t *testing.T) {
	t.Parallel()

	// In expanded view, units with subunits are always shown as expanded
	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"system": {
				Type: model.TypeSystem,
				Name: "System",
				Subunits: map[string]*model.Unit{
					"container": {
						Type: model.TypeContainer,
						Name: "Container",
					},
				},
			},
			"standalone": {
				Type: model.TypeSystem,
				Name: "Standalone",
			},
		},
	}

	v := view.GenerateExpandedView(m)

	require.NotNil(t, v)

	// Units with subunits should be marked as expanded
	assert.True(t, v.Units["system"].IsExpanded)

	// Units without subunits should not be marked as expanded
	assert.False(t, v.Units["standalone"].IsExpanded)
	assert.False(t, v.Units["system.container"].IsExpanded)
}

func TestGenerateExpandedView_TitleFromProperties(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "My Architecture"},
		Units:      map[string]*model.Unit{},
	}

	v := view.GenerateExpandedView(m)

	require.NotNil(t, v)
	assert.Equal(t, "My Architecture", v.Title)
}

func TestGenerateExpandedView_LevelIsC1(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units:      map[string]*model.Unit{},
	}

	v := view.GenerateExpandedView(m)

	require.NotNil(t, v)
	assert.Equal(t, view.LevelC1, v.Level)
}

func TestGenerateC1ViewDefinitionOrder(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		UnitOrder:  []string{"zulu", "alpha", "gamma"},
		Units: map[string]*model.Unit{
			"zulu":  {Type: model.TypeSystem, Name: "Zulu"},
			"alpha": {Type: model.TypeSystem, Name: "Alpha"},
			"gamma": {Type: model.TypeDb, Name: "Gamma"},
		},
	}

	v := view.GenerateC1View(m)
	require.NotNil(t, v)

	// View should propagate UnitOrder from Model
	require.Len(t, v.UnitOrder, 3, "UnitOrder should have 3 entries")
	assert.Equal(t, "zulu", v.UnitOrder[0], "first should be zulu")
	assert.Equal(t, "alpha", v.UnitOrder[1], "second should be alpha")
	assert.Equal(t, "gamma", v.UnitOrder[2], "third should be gamma")
}
