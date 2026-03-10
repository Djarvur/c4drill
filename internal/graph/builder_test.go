package graph_test

import (
	"testing"

	"github.com/Djarvur/c4drill/internal/graph"
	"github.com/Djarvur/c4drill/internal/model"
	"github.com/Djarvur/c4drill/internal/parser"
	"github.com/Djarvur/c4drill/internal/view"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildGraphBasicProperties(t *testing.T) {
	t.Parallel()

	// Create a simple model
	m := &parser.Model{
		Properties: model.Properties{
			Name:  "Test System",
			Edges: "spline",
		},
		Units: map[string]*model.Unit{
			"app": {
				Type:        model.TypeSystem,
				Name:        "App",
				Description: "Main application",
				Technology:  "Go",
			},
		},
	}

	v := view.GenerateC1View(m)
	require.NotNil(t, v)

	g := graph.BuildGraph(v)

	// Test 1: BuildGraph creates Graph with Title from View.Title
	assert.Equal(t, "Test System", g.Title)

	// Test 2: BuildGraph sets Direction to "TB" (top-to-bottom)
	assert.Equal(t, "TB", g.Direction)

	// Test 3: BuildGraph sets EdgeStyle from View.Edges
	assert.Equal(t, "spline", g.EdgeStyle)
}

func TestBuildGraphNodes(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"app": {
				Type:        model.TypeSystem,
				Name:        "App",
				Description: "Main app",
				Technology:  "Go",
			},
			"db": {
				Type:        model.TypeDb,
				Name:        "Database",
				Description: "Data store",
			},
		},
	}

	v := view.GenerateC1View(m)
	g := graph.BuildGraph(v)

	// Test 4: BuildGraph creates Node for each non-expanded ViewUnit
	assert.Len(t, g.Nodes, 2)

	// Test 6: Node.ID equals ViewUnit.FullPath
	nodeIDs := make(map[string]bool)
	for _, node := range g.Nodes {
		nodeIDs[node.ID] = true
	}

	assert.True(t, nodeIDs["app"])
	assert.True(t, nodeIDs["db"])
}

func TestBuildGraphNodeLabels(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"app": {
				Type:        model.TypeSystem,
				Name:        "App",
				Description: "Main application",
				Technology:  "Go",
			},
		},
	}

	v := view.GenerateC1View(m)
	g := graph.BuildGraph(v)
	require.Len(t, g.Nodes, 1)

	node := g.Nodes[0]

	// Test 8: Node.Label.Technology is set from unit.Technology
	assert.Equal(t, "Go", node.Label.Technology)

	// Test 9: Node.Label.Description is set from unit.Description
	assert.Equal(t, "Main application", node.Label.Description)
}

func TestBuildGraphCollapsedIndicator(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"app": {
				Type: model.TypeSystem,
				Name: "App",
				Subunits: map[string]*model.Unit{
					"api": {
						Type: model.TypeContainer,
						Name: "API",
					},
				},
			},
		},
	}

	v := view.GenerateC1View(m)
	g := graph.BuildGraph(v)

	require.Len(t, g.Nodes, 1)

	// Test 7: Node.Label.Name includes [+] for collapsed units with subunits
	assert.Contains(t, g.Nodes[0].Label.Name, "[+]")
}

func TestBuildGraphEdges(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"app": {
				Type: model.TypeSystem,
				Name: "App",
				Links: map[string]model.Link{
					"db": {
						Technology:  "SQL",
						Description: "Queries data",
					},
				},
			},
			"db": {
				Type: model.TypeDb,
				Name: "Database",
			},
		},
	}

	v := view.GenerateC1View(m)
	g := graph.BuildGraph(v)

	// Test 10: Edge created for each link where both endpoints are in view
	require.Len(t, g.Edges, 1)

	edge := g.Edges[0]
	assert.Equal(t, "app", edge.Source)
	assert.Equal(t, "db", edge.Target)

	// Test 11: Edge.Label.Technology wrapped in brackets notation
	assert.Equal(t, "SQL", edge.Label.Technology)
	assert.Equal(t, "Queries data", edge.Label.Description)
}

func TestBuildGraphEdgeDefaults(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"app": {
				Type: model.TypeSystem,
				Name: "App",
				Links: map[string]model.Link{
					"db": {}, // No style or position specified
				},
			},
			"db": {
				Type: model.TypeDb,
				Name: "Database",
			},
		},
	}

	v := view.GenerateC1View(m)
	g := graph.BuildGraph(v)

	require.Len(t, g.Edges, 1)
	edge := g.Edges[0]

	// Test 12: Edge.Style defaults to "solid" if not specified
	assert.Equal(t, "solid", edge.Style)

	// Test 13: Edge.Label.Position defaults to "middle" if not specified
	assert.Equal(t, "middle", edge.Label.Position)
}

func TestBuildGraphClusters(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"app": {
				Type:     model.TypeSystem,
				Name:     "App",
				Expanded: []string{"app"}, // Expanded
				Subunits: map[string]*model.Unit{
					"api": {
						Type: model.TypeContainer,
						Name: "API",
					},
					"web": {
						Type: model.TypeContainer,
						Name: "Web",
					},
				},
			},
		},
	}

	v := view.GenerateC1View(m)
	g := graph.BuildGraph(v)

	// Test 5: BuildGraph creates Cluster for each expanded ViewUnit
	require.Len(t, g.Clusters, 1)

	cluster := g.Clusters[0]
	assert.Equal(t, "cluster_app", cluster.ID)

	// Test 14: Cluster contains nodes for expanded unit's children
	assert.Len(t, cluster.Nodes, 2)
}

func TestBuildGraphMultipleLinks(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"app": {
				Type: model.TypeSystem,
				Name: "App",
				Links: map[string]model.Link{
					"db": {
						Technology: "SQL",
					},
				},
				LinksFrom: map[string]model.Link{
					"db": {
						Technology: "Callback",
					},
				},
			},
			"db": {
				Type: model.TypeDb,
				Name: "Database",
			},
		},
	}

	v := view.GenerateC1View(m)
	g := graph.BuildGraph(v)

	// Test 15: Multiple links between same units shown separately
	// bidirectional: app->db and db->app
	require.Len(t, g.Edges, 2)
}

//nolint:funlen // Test functions with model setup are naturally longer
func TestBuildGraphWithPathSetsExploreURL(t *testing.T) {
	t.Parallel()

	// Test 1: BuildGraphWithPath sets ExploreURL on collapsed nodes with subunits
	t.Run("sets explore URL on collapsed system with subunits", func(t *testing.T) {
		t.Parallel()

		m := &parser.Model{
			Properties: model.Properties{Name: "Test"},
			Units: map[string]*model.Unit{
				"mainsystem": {
					Type: model.TypeSystem,
					Name: "Main System",
					Subunits: map[string]*model.Unit{
						"api": {
							Type: model.TypeContainer,
							Name: "API",
						},
					},
				},
			},
		}

		v := view.GenerateC1View(m)
		g := graph.BuildGraphWithPath(v, "", "diagram", "svg")

		require.Len(t, g.Nodes, 1)
		// Node should have explore URL since it has subunits and is not expanded
		assert.Equal(t, "./mainsystem.svg", g.Nodes[0].ExploreURL)
	})

	// Test 5: Only system/box types get explore links (not person/db/queue)
	t.Run("only system and box types get explore links", func(t *testing.T) {
		t.Parallel()

		m := &parser.Model{
			Properties: model.Properties{Name: "Test"},
			Units: map[string]*model.Unit{
				"system": {
					Type: model.TypeSystem,
					Name: "System",
					Subunits: map[string]*model.Unit{
						"sub": {Type: model.TypeContainer, Name: "Sub"},
					},
				},
				"box": {
					Type: model.TypeBox,
					Name: "Box",
					Subunits: map[string]*model.Unit{
						"sub": {Type: model.TypeContainer, Name: "Sub"},
					},
				},
				"db": {
					Type: model.TypeDb,
					Name: "Database",
					Subunits: map[string]*model.Unit{
						"table": {Type: model.TypeComponent, Name: "Table"},
					},
				},
				"person": {
					Type: model.TypePerson,
					Name: "User",
					Subunits: map[string]*model.Unit{
						"role": {Type: model.TypeComponent, Name: "Role"},
					},
				},
			},
		}

		v := view.GenerateC1View(m)
		g := graph.BuildGraphWithPath(v, "", "diagram", "svg")

		require.Len(t, g.Nodes, 4)

		// Find nodes by ID and check explore URLs
		nodeMap := make(map[string]*graph.Node)
		for _, node := range g.Nodes {
			nodeMap[node.ID] = node
		}

		// System and box should have explore URLs
		assert.NotEmpty(t, nodeMap["system"].ExploreURL, "system should have explore URL")
		assert.NotEmpty(t, nodeMap["box"].ExploreURL, "box should have explore URL")

		// Db and person should NOT have explore URLs
		assert.Empty(t, nodeMap["db"].ExploreURL, "db should NOT have explore URL")
		assert.Empty(t, nodeMap["person"].ExploreURL, "person should NOT have explore URL")
	})
}

//nolint:funlen // Test functions with model setup are naturally longer
func TestBuildGraphWithPathSetsNavigation(t *testing.T) {
	t.Parallel()

	// Test 2: BuildGraphWithPath sets Navigation on C2/C3 graphs (nil for C1)
	t.Run("C1 view has no navigation", func(t *testing.T) {
		t.Parallel()

		m := &parser.Model{
			Properties: model.Properties{Name: "Test"},
			Units: map[string]*model.Unit{
				"app": {Type: model.TypeSystem, Name: "App"},
			},
		}

		v := view.GenerateC1View(m)
		g := graph.BuildGraphWithPath(v, "", "diagram", "svg")

		assert.Nil(t, g.Navigation)
	})

	// Test 3: BuildGraphWithPath computes correct back-link URL
	t.Run("C2 view has navigation with back-link", func(t *testing.T) {
		t.Parallel()

		m := &parser.Model{
			Properties: model.Properties{Name: "Test"},
			Units: map[string]*model.Unit{
				"mainsystem": {
					Type: model.TypeSystem,
					Name: "Main System",
					Subunits: map[string]*model.Unit{
						"api": {Type: model.TypeContainer, Name: "API"},
						"web": {Type: model.TypeContainer, Name: "Web"},
					},
				},
			},
		}

		v := view.GenerateC2View(m, "mainsystem")
		require.NotNil(t, v)

		g := graph.BuildGraphWithPath(v, "mainsystem", "diagram", "svg")

		require.NotNil(t, g.Navigation)
		require.NotNil(t, g.Navigation.BackLink)
		assert.Equal(t, "../diagram.svg", g.Navigation.BackLink.URL)
	})

	// Test 4: BuildGraphWithPath builds breadcrumb trail with correct URLs
	t.Run("C3 view has navigation with breadcrumbs", func(t *testing.T) {
		t.Parallel()

		m := &parser.Model{
			Properties: model.Properties{Name: "Test"},
			Units: map[string]*model.Unit{
				"mainsystem": {
					Type: model.TypeSystem,
					Name: "Main System",
					Subunits: map[string]*model.Unit{
						"api": {
							Type: model.TypeContainer,
							Name: "API",
							Subunits: map[string]*model.Unit{
								"auth": {Type: model.TypeComponent, Name: "Auth"},
							},
						},
					},
				},
			},
		}

		v := view.GenerateC3View(m, "mainsystem.api")
		require.NotNil(t, v)

		g := graph.BuildGraphWithPath(v, "mainsystem.api", "diagram", "svg")

		require.NotNil(t, g.Navigation)
		// Should have breadcrumbs: mainsystem > api
		require.Len(t, g.Navigation.Breadcrumbs, 2)

		// First breadcrumb (mainsystem) should have URL
		assert.Equal(t, "mainsystem", g.Navigation.Breadcrumbs[0].Name)
		assert.NotEmpty(t, g.Navigation.Breadcrumbs[0].URL)

		// Second breadcrumb (api - current) should NOT have URL
		assert.Equal(t, "api", g.Navigation.Breadcrumbs[1].Name)
		assert.Empty(t, g.Navigation.Breadcrumbs[1].URL)

		// Back-link should go to parent (mainsystem)
		require.NotNil(t, g.Navigation.BackLink)
		assert.Equal(t, "../mainsystem.svg", g.Navigation.BackLink.URL)
	})
}
