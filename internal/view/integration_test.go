package view_test

import (
	"testing"

	"github.com/Djarvur/c4drill/internal/model"
	"github.com/Djarvur/c4drill/internal/parser"
	"github.com/Djarvur/c4drill/internal/view"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegrationC1ViewMultiLevelModel tests that GenerateC1View with a multi-level
// model produces the correct top-level view showing only top-level units.
func TestIntegrationC1ViewMultiLevelModel(t *testing.T) {
	t.Parallel()

	// Create a model with nested units (system -> container -> component)
	m := &parser.Model{
		Properties: model.Properties{
			Name: "Multi-Level System",
		},
		Units: map[string]*model.Unit{
			"mainsystem": {
				Type:        model.TypeSystem,
				Name:        "Main System",
				Description: "The main system",
				Technology:  "Go",
				Subunits: map[string]*model.Unit{
					"api": {
						Type:        model.TypeContainer,
						Name:        "API",
						Description: "API container",
						Subunits: map[string]*model.Unit{
							"handler": {
								Type:        model.TypeComponent,
								Name:        "Handler",
								Description: "Request handler",
							},
						},
					},
					"web": {
						Type:        model.TypeContainer,
						Name:        "Web",
						Description: "Web frontend",
					},
				},
			},
			"external": {
				Type:        model.TypeSystemExternal,
				Name:        "External System",
				Description: "An external system",
			},
		},
	}

	v := view.GenerateC1View(m)

	require.NotNil(t, v)
	// Test 1: C1 view only shows top-level units (not nested containers/components)
	assert.Equal(t, view.LevelC1, v.Level)
	assert.Contains(t, v.Units, "mainsystem")
	assert.Contains(t, v.Units, "external")

	// Nested units should NOT appear at C1 level
	assert.NotContains(t, v.Units, "mainsystem.api")
	assert.NotContains(t, v.Units, "mainsystem.api.handler")

	// HasSubunits should be true for system with subunits
	assert.True(t, v.Units["mainsystem"].HasSubunits)
}

// TestIntegrationC2ViewNestedSystem tests that GenerateC2View with a nested system
// produces the correct container view showing subunits.
func TestIntegrationC2ViewNestedSystem(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"mainsystem": {
				Type:        model.TypeSystem,
				Name:        "Main System",
				Description: "The main system",
				Technology:  "Go",
				Edges:       "spline",
				Subunits: map[string]*model.Unit{
					"api": {
						Type:        model.TypeContainer,
						Name:        "API",
						Description: "API container",
						Technology:  "Go",
					},
					"web": {
						Type:        model.TypeContainer,
						Name:        "Web",
						Description: "Web frontend",
						Technology:  "React",
					},
					"db": {
						Type:        model.TypeContainerDb,
						Name:        "Database",
						Description: "Data store",
						Technology:  "PostgreSQL",
					},
				},
			},
		},
	}

	v := view.GenerateC2View(m, "mainsystem")

	require.NotNil(t, v)
	// Test 2: C2 view shows containers of expanded system
	assert.Equal(t, view.LevelC2, v.Level)
	assert.Contains(t, v.Units, "mainsystem.api")
	assert.Contains(t, v.Units, "mainsystem.web")
	assert.Contains(t, v.Units, "mainsystem.db")

	// Parent and ExpandedUnit should be set
	assert.Equal(t, "mainsystem", v.Parent)
	assert.Equal(t, "mainsystem", v.ExpandedUnit)

	// Edges should come from expanded unit
	assert.Equal(t, "spline", v.Edges)

	// FullPath should use dotted notation
	assert.Equal(t, "mainsystem.api", v.Units["mainsystem.api"].FullPath)
}

// TestIntegrationC3ViewNestedContainer tests that GenerateC3View with a nested container
// produces the correct component view.
func TestIntegrationC3ViewNestedContainer(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"mainsystem": {
				Type: model.TypeSystem,
				Name: "Main System",
				Subunits: map[string]*model.Unit{
					"api": {
						Type:        model.TypeContainer,
						Name:        "API",
						Description: "API container",
						Technology:  "Go",
						Edges:       "square",
						Subunits: map[string]*model.Unit{
							"handler": {
								Type:        model.TypeComponent,
								Name:        "Handler",
								Description: "Request handler",
								Technology:  "Go",
							},
							"service": {
								Type:        model.TypeComponent,
								Name:        "Service",
								Description: "Business logic",
								Technology:  "Go",
							},
							"repo": {
								Type:        model.TypeComponentDb,
								Name:        "Repository",
								Description: "Data access",
								Technology:  "Go",
							},
						},
					},
				},
			},
		},
	}

	v := view.GenerateC3View(m, "mainsystem.api")

	require.NotNil(t, v)
	// Test 3: C3 view shows components of expanded container
	assert.Equal(t, view.LevelC3, v.Level)
	assert.Contains(t, v.Units, "mainsystem.api.handler")
	assert.Contains(t, v.Units, "mainsystem.api.service")
	assert.Contains(t, v.Units, "mainsystem.api.repo")

	// Edges should come from expanded container
	assert.Equal(t, "square", v.Edges)

	// FullPath uses dotted notation for nested units
	assert.Equal(t, "mainsystem.api.handler", v.Units["mainsystem.api.handler"].FullPath)
}

// TestIntegrationExternalBoundaryNodes tests that external boundary nodes appear
// when links reference out-of-scope units.
func TestIntegrationExternalBoundaryNodes(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"internal": {
				Type: model.TypeSystem,
				Name: "Internal System",
				Links: map[string]model.Link{
					"externalapi": {
						Target:      "externalapi",
						Technology:  "HTTP",
						Description: "Calls external API",
					},
				},
			},
		},
	}

	v := view.GenerateC1View(m)

	require.NotNil(t, v)
	// Test 4: External boundary nodes appear when links reference out-of-scope units
	assert.Contains(t, v.Units, "externalapi")
	assert.True(t, v.Units["externalapi"].IsExternal)
	assert.Equal(t, model.TypeSystemExternal, v.Units["externalapi"].Unit.Type)
}

// TestIntegrationViewRespectsExpandedAttribute tests that the view respects
// the per-unit expanded attribute.
func TestIntegrationViewRespectsExpandedAttribute(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"expanded_system": {
				Type:        model.TypeSystem,
				Name:        "Expanded System",
				Expanded:    []string{"expanded_system"}, // This unit is expanded
				Subunits: map[string]*model.Unit{
					"container": {
						Type: model.TypeContainer,
						Name: "Container",
					},
				},
			},
			"collapsed_system": {
				Type:        model.TypeSystem,
				Name:        "Collapsed System",
				// No Expanded attribute - should be collapsed
				Subunits: map[string]*model.Unit{
					"container": {
						Type: model.TypeContainer,
						Name: "Container",
					},
				},
			},
		},
	}

	v := view.GenerateC1View(m)

	require.NotNil(t, v)
	// Test 5: View respects per-unit expanded attribute
	assert.True(t, v.Units["expanded_system"].IsExpanded)
	assert.False(t, v.Units["collapsed_system"].IsExpanded)
}

// TestIntegrationFullPathDottedNotation tests that FullPath correctly uses
// dotted notation for nested units at all levels.
func TestIntegrationFullPathDottedNotation(t *testing.T) {
	t.Parallel()

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

	// Test 6: FullPath correctly uses dotted notation for nested units

	// C1: top-level units have simple paths
	c1 := view.GenerateC1View(m)
	require.NotNil(t, c1)
	assert.Equal(t, "system", c1.Units["system"].FullPath)

	// C2: containers have "system.container" paths
	c2 := view.GenerateC2View(m, "system")
	require.NotNil(t, c2)
	assert.Equal(t, "system.container", c2.Units["system.container"].FullPath)

	// C3: components have "system.container.component" paths
	c3 := view.GenerateC3View(m, "system.container")
	require.NotNil(t, c3)
	assert.Equal(t, "system.container.component", c3.Units["system.container.component"].FullPath)
}

// TestIntegrationLinksWithExternalBoundary tests that links to external units
// create proper external boundary nodes.
func TestIntegrationLinksWithExternalBoundary(t *testing.T) {
	t.Parallel()

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
						Links: map[string]model.Link{
							"externalservice": {
								Target:      "externalservice",
								Technology:  "HTTP",
								Description: "External call",
							},
						},
					},
				},
			},
		},
	}

	// At C2 level, external boundary nodes should be created for container's links
	v := view.GenerateC2View(m, "system")
	require.NotNil(t, v)
	assert.Contains(t, v.Units, "externalservice")
	assert.True(t, v.Units["externalservice"].IsExternal)
}
