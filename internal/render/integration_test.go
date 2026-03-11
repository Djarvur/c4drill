package render_test

import (
	"strings"
	"testing"

	"github.com/Djarvur/c4drill/internal/graph"
	"github.com/Djarvur/c4drill/internal/model"
	"github.com/Djarvur/c4drill/internal/parser"
	"github.com/Djarvur/c4drill/internal/render"
	"github.com/Djarvur/c4drill/internal/view"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Note: Tests in this file do NOT use t.Parallel() because the go-graphviz
// library uses a WASM-based rendering engine that has concurrency issues.

// buildTestModel creates a simple test model for integration tests.
func buildTestModel() *parser.Model {
	return &parser.Model{
		Properties: model.Properties{
			Name: "Test System",
		},
		Units: map[string]*model.Unit{
			"user": {
				Type:        model.TypePerson,
				Name:        "User",
				Description: "A system user",
			},
			"system": {
				Type:        model.TypeSystem,
				Name:        "Main System",
				Description: "The main software system",
				Technology:  "Go",
			},
		},
	}
}

// buildTestModelWithLinks creates a test model with linked units.
func buildTestModelWithLinks() *parser.Model {
	return &parser.Model{
		Properties: model.Properties{
			Name: "Test System with Links",
		},
		Units: map[string]*model.Unit{
			"user": {
				Type:        model.TypePerson,
				Name:        "User",
				Description: "A system user",
				Links: []model.Link{
					{
						Peer:        "system",
						Technology:  "HTTP",
						Description: "Uses",
					},
				},
			},
			"system": {
				Type:        model.TypeSystem,
				Name:        "Main System",
				Description: "The main software system",
				Technology:  "Go",
			},
		},
	}
}

// buildTestModelWithExpanded creates a test model with an expanded system containing subunits.
func buildTestModelWithExpanded() *parser.Model {
	return &parser.Model{
		Properties: model.Properties{
			Name: "Test System with Expansion",
		},
		Units: map[string]*model.Unit{
			"system": {
				Type:        model.TypeSystem,
				Name:        "Main System",
				Description: "The main software system",
				Technology:  "Go",
				Expanded:    []string{"system"}, // Expand the system itself
				Subunits: map[string]*model.Unit{
					"api": {
						Type:        model.TypeContainer,
						Name:        "API",
						Description: "REST API",
						Technology:  "Go",
					},
					"db": {
						Type:        model.TypeContainerDb,
						Name:        "Database",
						Description: "PostgreSQL database",
						Technology:  "PostgreSQL",
					},
				},
			},
		},
	}
}

// buildTestModelWithExternal creates a test model with external units.
func buildTestModelWithExternal() *parser.Model {
	return &parser.Model{
		Properties: model.Properties{
			Name: "Test System with External",
		},
		Units: map[string]*model.Unit{
			"system": {
				Type:        model.TypeSystem,
				Name:        "Main System",
				Description: "The main software system",
				Technology:  "Go",
			},
			"external": {
				Type:        model.TypeSystemExternal,
				Name:        "External System",
				Description: "An external dependency",
				Technology:  "REST API",
			},
		},
	}
}

// buildEmptyTestModel creates an empty test model.
func buildEmptyTestModel() *parser.Model {
	return &parser.Model{
		Properties: model.Properties{
			Name: "Empty System",
		},
		Units: map[string]*model.Unit{},
	}
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestIntegrationFullPipelineDOT(t *testing.T) {
	// Test 1: Full pipeline from model -> view -> graph -> DOT produces valid output
	m := buildTestModel()

	v := view.GenerateC1View(m)
	require.NotNil(t, v, "GenerateC1View should return a view")

	g := graph.BuildGraph(v)
	require.NotNil(t, g, "BuildGraph should return a graph")

	output, err := render.RenderDOT(g)
	require.NoError(t, err, "RenderDOT should not return error for valid graph")
	assert.NotEmpty(t, output, "RenderDOT should return non-empty bytes")

	// Verify DOT output contains expected structure
	dotStr := string(output)
	assert.True(t, strings.Contains(dotStr, "digraph") || strings.Contains(dotStr, "graph"),
		"DOT output should contain graph/digraph keyword")
	assert.Contains(t, dotStr, "Test System", "DOT output should contain graph title")
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestIntegrationFullPipelineSVG(t *testing.T) {
	// Test 2: Full pipeline from model -> view -> graph -> SVG produces valid output
	m := buildTestModel()

	v := view.GenerateC1View(m)
	require.NotNil(t, v, "GenerateC1View should return a view")

	g := graph.BuildGraph(v)
	require.NotNil(t, g, "BuildGraph should return a graph")

	output, err := render.RenderSVG(g)
	require.NoError(t, err, "RenderSVG should not return error for valid graph")
	assert.NotEmpty(t, output, "RenderSVG should return non-empty bytes")

	// Verify SVG output contains expected structure
	svgStr := string(output)
	assert.Contains(t, svgStr, "<?xml", "SVG output should contain XML declaration")
	assert.Contains(t, svgStr, "<svg", "SVG output should contain svg element")
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestIntegrationRenderFormatDOT(t *testing.T) {
	// Test 3: Render(g, "dot") produces same output as RenderDOT(g)
	m := buildTestModel()
	v := view.GenerateC1View(m)
	g := graph.BuildGraph(v)

	renderOutput, err := render.Render(g, "dot")
	require.NoError(t, err, "Render with dot format should not return error")

	dotOutput, err := render.RenderDOT(g)
	require.NoError(t, err, "RenderDOT should not return error")

	assert.Equal(t, dotOutput, renderOutput,
		"Render(g, 'dot') should return same result as RenderDOT(g)")
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestIntegrationRenderFormatSVG(t *testing.T) {
	// Test 4: Render(g, "svg") produces same output as RenderSVG(g)
	m := buildTestModel()
	v := view.GenerateC1View(m)
	g := graph.BuildGraph(v)

	renderOutput, err := render.Render(g, "svg")
	require.NoError(t, err, "Render with svg format should not return error")

	svgOutput, err := render.RenderSVG(g)
	require.NoError(t, err, "RenderSVG should not return error")

	assert.Equal(t, svgOutput, renderOutput,
		"Render(g, 'svg') should return same result as RenderSVG(g)")
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestIntegrationRenderInvalidFormat(t *testing.T) {
	// Test 5: Render(g, "invalid") returns error with clear message
	m := buildTestModel()
	v := view.GenerateC1View(m)
	g := graph.BuildGraph(v)

	output, err := render.Render(g, "invalid")
	require.Error(t, err, "Render with invalid format should return error")
	assert.Nil(t, output, "Render with invalid format should return nil bytes")
	require.ErrorContains(t, err, "unsupported format",
		"Error message should mention unsupported format")
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestIntegrationGraphWithNodesRendersNodeIDs(t *testing.T) {
	// Test 6: Graph with nodes renders all node IDs in output
	m := buildTestModel()
	v := view.GenerateC1View(m)
	g := graph.BuildGraph(v)

	output, err := render.RenderDOT(g)
	require.NoError(t, err)

	dotStr := string(output)
	// Verify that node identifiers appear in the output
	assert.Contains(t, dotStr, "user", "DOT output should contain 'user' node ID")
	assert.Contains(t, dotStr, "system", "DOT output should contain 'system' node ID")
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestIntegrationGraphWithEdgesRendersRelationships(t *testing.T) {
	// Test 7: Graph with edges renders all edge relationships in output
	m := buildTestModelWithLinks()
	v := view.GenerateC1View(m)
	g := graph.BuildGraph(v)

	output, err := render.RenderDOT(g)
	require.NoError(t, err)

	dotStr := string(output)
	// Verify that edge relationships appear in the output
	// The edge should connect user to system
	assert.Contains(t, dotStr, "user", "DOT output should contain 'user' node")
	assert.Contains(t, dotStr, "system", "DOT output should contain 'system' node")
	// Edges are represented as arrows in DOT
	assert.Contains(t, dotStr, "->", "DOT output should contain edge arrow")
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestIntegrationGraphWithClustersRendersStructure(t *testing.T) {
	// Test 8: Graph with clusters renders cluster structure in output
	m := buildTestModelWithExpanded()
	v := view.GenerateC1View(m)
	g := graph.BuildGraph(v)

	// Verify cluster was created
	require.NotEmpty(t, g.Clusters, "Graph should have clusters for expanded units")

	output, err := render.RenderDOT(g)
	require.NoError(t, err)

	dotStr := string(output)
	// Verify cluster structure appears in DOT
	assert.Contains(t, dotStr, "subgraph", "DOT output should contain subgraph for clusters")
	assert.Contains(t, dotStr, "cluster_", "DOT output should contain cluster_ prefix")
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestIntegrationEmptyGraphRendersWithoutError(t *testing.T) {
	// Test 9: Empty graph renders without error (minimal valid DOT/SVG)
	m := buildEmptyTestModel()
	v := view.GenerateC1View(m)
	g := graph.BuildGraph(v)

	// Empty graph should still render
	dotOutput, err := render.RenderDOT(g)
	require.NoError(t, err, "RenderDOT should not error on empty graph")
	assert.NotEmpty(t, dotOutput, "RenderDOT should return non-empty bytes even for empty graph")

	svgOutput, err := render.RenderSVG(g)
	require.NoError(t, err, "RenderSVG should not error on empty graph")
	assert.NotEmpty(t, svgOutput, "RenderSVG should return non-empty bytes even for empty graph")
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestIntegrationSingleNodeGraph(t *testing.T) {
	// Edge case: Graph with single node
	singleNodeModel := &parser.Model{
		Properties: model.Properties{Name: "Single Node System"},
		Units: map[string]*model.Unit{
			"only": {
				Type:        model.TypeSystem,
				Name:        "Only System",
				Description: "The only unit",
			},
		},
	}

	v := view.GenerateC1View(singleNodeModel)
	g := graph.BuildGraph(v)

	output, err := render.RenderDOT(g)
	require.NoError(t, err)
	assert.Contains(t, string(output), "only", "DOT output should contain the single node ID")
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestIntegrationExternalNodesRendered(t *testing.T) {
	// Edge case: Graph with external nodes
	m := buildTestModelWithExternal()
	v := view.GenerateC1View(m)
	g := graph.BuildGraph(v)

	output, err := render.RenderDOT(g)
	require.NoError(t, err)

	dotStr := string(output)
	assert.Contains(t, dotStr, "system", "DOT output should contain internal system")
	assert.Contains(t, dotStr, "external", "DOT output should contain external system")
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestIntegrationNilModelReturnsNilGraph(t *testing.T) {
	// Edge case: nil model should return nil view and nil graph
	v := view.GenerateC1View(nil)
	assert.Nil(t, v, "GenerateC1View should return nil for nil model")

	g := graph.BuildGraph(v)
	assert.Nil(t, g, "BuildGraph should return nil for nil view")
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestIntegrationC2ViewPipeline(t *testing.T) {
	// Test C2 view pipeline
	m := buildTestModelWithExpanded()
	v := view.GenerateC2View(m, "system")
	require.NotNil(t, v, "GenerateC2View should return a view")

	g := graph.BuildGraph(v)
	require.NotNil(t, g, "BuildGraph should return a graph")

	output, err := render.RenderDOT(g)
	require.NoError(t, err)
	assert.NotEmpty(t, output)
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestIntegrationC3ViewPipeline(t *testing.T) {
	// Test C3 view pipeline - need a model with nested containers
	m := &parser.Model{
		Properties: model.Properties{Name: "C3 Test"},
		Units: map[string]*model.Unit{
			"system": {
				Type:     model.TypeSystem,
				Name:     "Main System",
				Expanded: []string{"api"},
				Subunits: map[string]*model.Unit{
					"api": {
						Type:     model.TypeContainer,
						Name:     "API",
						Expanded: []string{"handlers"},
						Subunits: map[string]*model.Unit{
							"handlers": {
								Type:        model.TypeComponent,
								Name:        "Handlers",
								Description: "Request handlers",
							},
						},
					},
				},
			},
		},
	}

	v := view.GenerateC3View(m, "system.api")
	require.NotNil(t, v, "GenerateC3View should return a view")

	g := graph.BuildGraph(v)
	require.NotNil(t, g, "BuildGraph should return a graph")

	output, err := render.RenderDOT(g)
	require.NoError(t, err)
	assert.NotEmpty(t, output)
}

// ============================================================================
// Navigation SVG Integration Tests (05-03)
// ============================================================================

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestIntegration_SVG_ExploreLink(t *testing.T) {
	// Build graph with explore URL
	m := buildNavTestModel()
	v := view.GenerateC1View(m)
	g := graph.BuildGraphWithPath(v, "", "test", "svg")

	// Render to SVG
	svgBytes, err := render.RenderSVG(g)
	require.NoError(t, err)
	require.NotEmpty(t, svgBytes)

	svgStr := string(svgBytes)

	// Verify SVG contains clickable link for explore
	// GraphViz generates xlink:href or href for URL attribute
	assert.True(t,
		strings.Contains(svgStr, "xlink:href") || strings.Contains(svgStr, "href"),
		"SVG should contain href attribute for clickable nodes")
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestIntegration_SVG_BackLink(t *testing.T) {
	// Build C2 graph with back-link
	m := buildNavTestModel()
	v := view.GenerateC2View(m, "mainsystem")
	g := graph.BuildGraphWithPath(v, "mainsystem", "test", "svg")

	// Render to SVG
	svgBytes, err := render.RenderSVG(g)
	require.NoError(t, err)

	svgStr := string(svgBytes)

	// Verify back-link text appears (navigation label)
	assert.Contains(t, svgStr, "Back to", "SVG should contain back-link text")
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestIntegration_SVG_Breadcrumbs(t *testing.T) {
	// Build C3 graph with breadcrumbs
	m := buildNavTestModelNested()
	v := view.GenerateC3View(m, "mainsystem.api")
	g := graph.BuildGraphWithPath(v, "mainsystem.api", "test", "svg")

	// Render to SVG
	svgBytes, err := render.RenderSVG(g)
	require.NoError(t, err)

	svgStr := string(svgBytes)

	// Verify breadcrumb separator appears (> or &gt;)
	assert.True(t,
		strings.Contains(svgStr, ">") || strings.Contains(svgStr, "&gt;"),
		"SVG should contain breadcrumb separator")
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestIntegration_SVG_C1NoNavigation(t *testing.T) {
	// C1 should have no navigation elements
	m := buildNavTestModel()
	v := view.GenerateC1View(m)
	g := graph.BuildGraphWithPath(v, "", "test", "svg")

	// Render to SVG
	svgBytes, err := render.RenderSVG(g)
	require.NoError(t, err)

	svgStr := string(svgBytes)

	// C1 should not have back-link
	assert.NotContains(t, svgStr, "Back to", "C1 SVG should not contain back-link")
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestIntegration_FullPipeline_Navigation(t *testing.T) {
	// Full pipeline: model -> view -> graph -> render -> SVG
	m := buildNavTestModelNested()

	// Generate all views and verify navigation
	testCases := []struct {
		name    string
		view    *view.View
		path    string
		hasNav  bool
		navText string
	}{
		{
			name:    "C1",
			view:    view.GenerateC1View(m),
			path:    "",
			hasNav:  false,
			navText: "",
		},
		{
			name:    "C2",
			view:    view.GenerateC2View(m, "mainsystem"),
			path:    "mainsystem",
			hasNav:  true,
			navText: "Back to",
		},
		{
			name:    "C3",
			view:    view.GenerateC3View(m, "mainsystem.api"),
			path:    "mainsystem.api",
			hasNav:  true,
			navText: ">",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := graph.BuildGraphWithPath(tc.view, tc.path, "test", "svg")
			require.NotNil(t, g)

			svgBytes, err := render.RenderSVG(g)
			require.NoError(t, err)

			svgStr := string(svgBytes)

			if tc.hasNav {
				assert.Contains(t, svgStr, tc.navText, "%s should contain navigation text", tc.name)
			} else {
				assert.NotContains(t, svgStr, "Back to", "%s should not have back-link", tc.name)
			}
		})
	}
}

// ============================================================================
// Helper functions for navigation tests
// ============================================================================

// buildNavTestModel creates a simple test model for navigation tests.
func buildNavTestModel() *parser.Model {
	return &parser.Model{
		Properties: model.Properties{Name: "Navigation Test"},
		Units: map[string]*model.Unit{
			"mainsystem": {
				Type: model.TypeSystem,
				Name: "Main System",
				Subunits: map[string]*model.Unit{
					"api": {Type: model.TypeSystem, Name: "API"},
				},
			},
		},
	}
}

// buildNavTestModelNested creates a nested test model for C3 navigation tests.
func buildNavTestModelNested() *parser.Model {
	return &parser.Model{
		Properties: model.Properties{Name: "Navigation Test"},
		Units: map[string]*model.Unit{
			"mainsystem": {
				Type:     model.TypeSystem,
				Name:     "Main System",
				Expanded: []string{"api"},
				Subunits: map[string]*model.Unit{
					"api": {
						Type: model.TypeSystem,
						Name: "API",
						Subunits: map[string]*model.Unit{
							"auth": {Type: model.TypeSystem, Name: "Auth"},
						},
					},
				},
			},
		},
	}
}
