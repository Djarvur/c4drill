package graph_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/Djarvur/c4drill/internal/graph"
	"github.com/Djarvur/c4drill/internal/model"
	"github.com/Djarvur/c4drill/internal/parser"
	"github.com/Djarvur/c4drill/internal/render"
	"github.com/Djarvur/c4drill/internal/testutil/canonical"
	"github.com/Djarvur/c4drill/internal/validator"
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

	// Test 7: Node.Label.Name includes 🔍 for collapsed units with subunits
	assert.Contains(t, g.Nodes[0].Label.Name, "🔍")
}

func TestBuildGraphEdges(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"app": {
				Type: model.TypeSystem,
				Name: "App",
				Links: []model.Link{
					{
						Peer:        "db",
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
				Links: []model.Link{
					{Peer: "db"}, // No style or position specified
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
	assert.Equal(t, "app", cluster.ID)

	// Test 14: Cluster contains nodes for expanded unit's children
	assert.Len(t, cluster.Nodes, 2)
}

// D-07: a top-level unit listed in properties.expanded with NO subunits renders
// as a plain collapsed node — expansion only takes effect when there are
// subunits to show, so no empty cluster box is produced.
func TestBuildGraphExpandedEmptyUnitRendersPlainNode(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test", Expanded: []string{"app"}},
		Units: map[string]*model.Unit{
			"app": {
				Type: model.TypeSystem,
				Name: "App",
			},
		},
	}

	v := view.GenerateC1View(m)
	g := graph.BuildGraph(v)

	// D-07: no subunits -> plain collapsed node, not an empty cluster box
	require.Empty(t, g.Clusters)
	require.Len(t, g.Nodes, 1)
	assert.Equal(t, "app", g.Nodes[0].ID)
}

// Fixture subunit-name constant for the D-04 cluster test below — repeated
// fixture literals stay below the goconst threshold.
const authComponentName = "auth"

// D-04: per-unit expanded containers render their components as a cluster
// inside their system's C2 diagram (builder.go:45 C2/C3 branch). This test
// also guards Task 1's edit scope: it fails if the C2/C3 branch is altered.
func TestBuildGraphC2ExpandedContainerRendersCluster(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"app": {
				Type:     model.TypeSystem,
				Name:     "App",
				Expanded: []string{"api"}, // Expand the api subunit
				Subunits: map[string]*model.Unit{
					"api": {
						Type: model.TypeContainer,
						Name: "API",
						Subunits: map[string]*model.Unit{
							authComponentName: {Type: model.TypeComponent},
							"store":           {Type: model.TypeComponent},
						},
					},
				},
			},
		},
	}

	v := view.GenerateC2View(m, "app")
	g := graph.BuildGraph(v)

	// D-04: boundary cluster only, with the expanded container's cluster inside
	require.Len(t, g.Clusters, 1)
	require.Len(t, g.Clusters[0].Clusters, 1)
	assert.Equal(t, "app.api", g.Clusters[0].Clusters[0].ID)
	assert.Len(t, g.Clusters[0].Clusters[0].Nodes, 2)
}

// WR-01: the D-07 empty-cluster guard also applies to the C2/C3 branch — a
// per-unit expanded entry naming a subunit WITHOUT subunits renders as a plain
// node inside the boundary cluster, not an empty cluster box.
func TestBuildGraphC2ExpandedEmptySubunitRendersPlainNode(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"app": {
				Type:     model.TypeSystem,
				Name:     "App",
				Expanded: []string{"api"}, // Expanded entry, but api has no subunits
				Subunits: map[string]*model.Unit{
					"api": {
						Type: model.TypeContainer,
						Name: "API",
					},
				},
			},
		},
	}

	v := view.GenerateC2View(m, "app")
	g := graph.BuildGraph(v)

	// WR-01: no empty cluster box — the expanded-but-empty container renders
	// as a plain node inside the boundary cluster.
	require.Len(t, g.Clusters, 1)
	require.Empty(t, g.Clusters[0].Clusters)
	require.Len(t, g.Clusters[0].Nodes, 1)
	assert.Equal(t, "app.api", g.Clusters[0].Nodes[0].ID)
}

func TestBuildGraphMultipleLinks(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"app": {
				Type: model.TypeSystem,
				Name: "App",
				Links: []model.Link{
					{
						Peer:       "db",
						Technology: "SQL",
					},
				},
				LinksFrom: []model.Link{
					{
						Peer:       "db",
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
		assert.Equal(t, "diagram/mainsystem.svg", g.Nodes[0].ExploreURL)
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

// TestBuildExpandedGraphCreatesClusterWithCorrectID verifies cluster ID creation.
func TestBuildExpandedGraphCreatesClusterWithCorrectID(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"mainapp": {
				Type: model.TypeSystem,
				Name: "Main App",
				Subunits: map[string]*model.Unit{
					"api": {
						Type: model.TypeContainer,
						Name: "API",
					},
				},
			},
		},
	}

	v := view.GenerateExpandedView(m)
	require.NotNil(t, v)

	g := graph.BuildExpandedGraph(v)
	require.NotNil(t, g)

	// Top-level unit with subunits should become a cluster
	require.Len(t, g.Clusters, 1)
	assert.Equal(t, "mainapp", g.Clusters[0].ID)
}

// TestBuildExpandedGraphRecursivelyBuildsNestedClusters verifies recursive cluster nesting.
func TestBuildExpandedGraphRecursivelyBuildsNestedClusters(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"mainapp": {
				Type: model.TypeSystem,
				Name: "Main App",
				Subunits: map[string]*model.Unit{
					"api": {
						Type: model.TypeContainer,
						Name: "API",
						Subunits: map[string]*model.Unit{
							"auth": {
								Type: model.TypeComponent,
								Name: "Auth",
							},
						},
					},
				},
			},
		},
	}

	v := view.GenerateExpandedView(m)
	g := graph.BuildExpandedGraph(v)

	require.NotNil(t, g)
	require.Len(t, g.Clusters, 1)

	// Top-level cluster
	topCluster := g.Clusters[0]
	assert.Equal(t, "mainapp", topCluster.ID)

	// Nested cluster for api (has subunits)
	require.Len(t, topCluster.Clusters, 1)
	nestedCluster := topCluster.Clusters[0]
	assert.Equal(t, "mainapp.api", nestedCluster.ID)
}

// TestBuildExpandedGraphAddsLeafSubunitsAsNodes verifies leaf subunits become nodes, not clusters.
func TestBuildExpandedGraphAddsLeafSubunitsAsNodes(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"mainapp": {
				Type: model.TypeSystem,
				Name: "Main App",
				Subunits: map[string]*model.Unit{
					"api": {
						Type: model.TypeContainer,
						Name: "API",
						// No subunits - leaf
					},
					"web": {
						Type: model.TypeContainer,
						Name: "Web",
						// No subunits - leaf
					},
				},
			},
		},
	}

	v := view.GenerateExpandedView(m)
	g := graph.BuildExpandedGraph(v)

	require.NotNil(t, g)
	require.Len(t, g.Clusters, 1)

	cluster := g.Clusters[0]
	// Both leaf subunits should be nodes, not clusters
	assert.Len(t, cluster.Nodes, 2)
	assert.Empty(t, cluster.Clusters)

	// Verify node IDs
	nodeIDs := make(map[string]bool)
	for _, node := range cluster.Nodes {
		nodeIDs[node.ID] = true
	}

	assert.True(t, nodeIDs["mainapp.api"])
	assert.True(t, nodeIDs["mainapp.web"])
}

// TestBuildExpandedGraphProducesDeeplyNestedClusters verifies deep cluster nesting.
func TestBuildExpandedGraphProducesDeeplyNestedClusters(t *testing.T) {
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
								Subunits: map[string]*model.Unit{
									"subcomponent": {
										Type: model.TypeComponent,
										Name: "SubComponent",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	v := view.GenerateExpandedView(m)
	g := graph.BuildExpandedGraph(v)

	require.NotNil(t, g)
	require.Len(t, g.Clusters, 1)

	// Level 1: system
	l1 := g.Clusters[0]
	assert.Equal(t, "system", l1.ID)
	require.Len(t, l1.Clusters, 1)

	// Level 2: container
	l2 := l1.Clusters[0]
	assert.Equal(t, "system.container", l2.ID)
	require.Len(t, l2.Clusters, 1)

	// Level 3: component
	l3 := l2.Clusters[0]
	assert.Equal(t, "system.container.component", l3.ID)
	// subcomponent is a leaf, so it's a node
	require.Len(t, l3.Nodes, 1)
	assert.Equal(t, "system.container.component.subcomponent", l3.Nodes[0].ID)
}

// TestBuildExpandedGraphHandlesMixedTopLevelUnits verifies clusters and nodes at the top level.
func TestBuildExpandedGraphHandlesMixedTopLevelUnits(t *testing.T) {
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
			"db": {
				Type: model.TypeDb,
				Name: "Database",
				// No subunits - should be a node
			},
		},
	}

	v := view.GenerateExpandedView(m)
	g := graph.BuildExpandedGraph(v)

	require.NotNil(t, g)

	// System with subunits -> cluster
	require.Len(t, g.Clusters, 1)
	assert.Equal(t, "system", g.Clusters[0].ID)

	// DB without subunits -> node
	require.Len(t, g.Nodes, 1)
	assert.Equal(t, "db", g.Nodes[0].ID)
}

// TestBuildExpandedGraphBuildsEdgesForCrossLevelConnections verifies edges across levels.
func TestBuildExpandedGraphBuildsEdgesForCrossLevelConnections(t *testing.T) {
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
							{Peer: "db", Technology: "SQL"},
						},
					},
				},
			},
			"db": {
				Type: model.TypeDb,
				Name: "Database",
			},
		},
	}

	v := view.GenerateExpandedView(m)
	g := graph.BuildExpandedGraph(v)

	require.NotNil(t, g)

	// Edge should exist between nested api and top-level db
	require.Len(t, g.Edges, 1)
	assert.Equal(t, "system.api", g.Edges[0].Source)
	assert.Equal(t, "db", g.Edges[0].Target)
}

// TestBuildExpandedGraphPreservesLengthAttributeForNestedSubunitToExternalUnit
// verifies minlen survives cross-level resolution.
func TestBuildExpandedGraphPreservesLengthAttributeForNestedSubunitToExternalUnit(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"linuxSystem": {
				Type: model.TypeSystem,
				Name: "Linux System",
				Subunits: map[string]*model.Unit{
					"storages": {
						Type: model.TypeContainer,
						Name: "Storages",
						Subunits: map[string]*model.Unit{
							"keycloakStorage": {
								Type: model.TypeContainer,
								Name: "Keycloak Storage",
								Subunits: map[string]*model.Unit{
									"client": {
										Type: model.TypeComponent,
										Name: "Keycloak Client",
										Links: []model.Link{
											{Peer: "keycloak", Length: 2},
										},
									},
								},
							},
						},
					},
				},
			},
			"keycloak": {
				Type: model.TypeSystemExternal,
				Name: "Keycloak",
			},
		},
	}

	v := view.GenerateExpandedView(m)
	g := graph.BuildExpandedGraph(v)

	require.NotNil(t, g)

	// Find the edge from client to keycloak
	var foundEdge *graph.Edge

	for _, edge := range g.Edges {
		if edge.Source == "linuxSystem.storages.keycloakStorage.client" && edge.Target == "keycloak" {
			foundEdge = edge
			break
		}
	}

	require.NotNil(t, foundEdge, "Edge from client to keycloak should exist")
	assert.Equal(t, 2, foundEdge.MinLen, "Edge should have MinLen=2 from TOML length attribute")
}

// TestBuildExpandedGraphRealToml tests that length attribute is preserved for the real TOML file.
// This is a regression test for a bug where LinksFrom entries created by the validator
// did not preserve link attributes like Length, causing edges to lose minlen values
// when the target unit was processed before the source unit.
// The fixture is the sanitized public copy of the private auth-infrastructure
// structure (D-01), committed at cmd/c4drill/testdata/multilevel.toml.
func TestBuildExpandedGraphRealToml(t *testing.T) {
	t.Parallel()

	// Use ParseFile like the command does
	m, err := parser.ParseFile("../../cmd/c4drill/testdata/multilevel.toml")
	require.NoError(t, err)

	// Run validation like the command does (this adds LinksFrom entries)
	valErrors := validator.Validate(m)
	require.Empty(t, valErrors, "model should be valid")

	// Check that the client unit has the link with length attribute
	client := m.Units["mainSystem"].Subunits["storages"].Subunits["externalStorage"].Subunits["client"]
	require.NotNil(t, client, "client unit should exist")
	require.Len(t, client.Links, 1, "client should have 1 link")
	assert.Equal(t, "externalSys", client.Links[0].Peer)
	expectedLength := client.Links[0].Length
	require.Positive(t, expectedLength, "link should have Length > 0")

	// Generate expanded view and check edges
	v := view.GenerateExpandedView(m)
	g := graph.BuildExpandedGraph(v)

	// Find the edge from client to externalSys
	var foundEdge *graph.Edge

	for _, edge := range g.Edges {
		if edge.Source == "mainSystem.storages.externalStorage.client" && edge.Target == "externalSys" {
			foundEdge = edge
			break
		}
	}

	require.NotNil(t, foundEdge, "Edge from client to externalSys should exist")
	assert.Equal(t, expectedLength, foundEdge.MinLen, "Edge MinLen should match TOML length attribute")

	// Also verify that the rendered DOT contains minlen
	dotData, err := render.RenderDOT(g)
	require.NoError(t, err)

	expectedMinlen := fmt.Sprintf("minlen=%d", expectedLength)
	assert.Contains(t, string(dotData), expectedMinlen, "DOT output should contain minlen for the edge")
}

func TestBuildGraphEdgeColor(t *testing.T) {
	t.Parallel()

	// Test: Edge color matches source unit border color (C1 internal -> PersonBorder = dark blue)
	t.Run("edge from C1 system has PersonBorder color", func(t *testing.T) {
		t.Parallel()

		m := &parser.Model{
			Properties: model.Properties{Name: "Test"},
			Units: map[string]*model.Unit{
				"app": {
					Type:  model.TypeSystem, // C1 -> PersonBorder = "#073B6F" (dark blue)
					Name:  "App",
					Links: []model.Link{{Peer: "db"}},
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
		assert.Equal(t, "#073B6F", g.Edges[0].Color) // PersonBorder (dark blue)
	})

	// Test: Edge from external system uses external border color
	t.Run("edge from external system has PersonExternalBorder color", func(t *testing.T) {
		t.Parallel()

		m := &parser.Model{
			Properties: model.Properties{Name: "Test"},
			Units: map[string]*model.Unit{
				"ext": {
					Type:  model.TypeSystemExternal, // -> SystemExternalBorder = "#8A8A8A"
					Name:  "External",
					Links: []model.Link{{Peer: "app"}},
				},
				"app": {
					Type: model.TypeSystem,
					Name: "App",
				},
			},
		}

		v := view.GenerateC1View(m)
		g := graph.BuildGraph(v)

		require.Len(t, g.Edges, 1)
		assert.Equal(t, "#8A8A8A", g.Edges[0].Color)
	})

	// Test: Explicit link.Color overrides source border color
	t.Run("edge with explicit color override uses link color", func(t *testing.T) {
		t.Parallel()

		m := &parser.Model{
			Properties: model.Properties{Name: "Test"},
			Units: map[string]*model.Unit{
				"app": {
					Type: model.TypeSystem,
					Name: "App",
					Links: []model.Link{{
						Peer:  "db",
						Color: "#FF0000", // Explicit override
					}},
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
		assert.Equal(t, "#FF0000", g.Edges[0].Color)
	})

	// Test: Edge from C2 container has ContainerBorder color
	t.Run("edge from C2 container has ContainerBorder color", func(t *testing.T) {
		t.Parallel()

		m := &parser.Model{
			Properties: model.Properties{Name: "Test"},
			Units: map[string]*model.Unit{
				"app": {
					Type:     model.TypeSystem,
					Name:     "App",
					Expanded: []string{"app"},
					Subunits: map[string]*model.Unit{
						"api": {
							Type:  model.TypeContainer, // C2 -> ContainerBorder = "#3C7FC0"
							Name:  "API",
							Links: []model.Link{{Peer: "ext"}},
						},
						"db": {
							Type: model.TypeContainerDb,
							Name: "Database",
						},
					},
				},
			},
		}

		v := view.GenerateC2View(m, "app")
		require.NotNil(t, v)
		g := graph.BuildGraph(v)

		require.Len(t, g.Edges, 1)
		assert.Equal(t, "#3C7FC0", g.Edges[0].Color)
	})

	// Test: Edge from C3 component has ComponentBorder color
	t.Run("edge from C3 component has ComponentBorder color", func(t *testing.T) {
		t.Parallel()

		m := &parser.Model{
			Properties: model.Properties{Name: "Test"},
			Units: map[string]*model.Unit{
				"app": {
					Type: model.TypeSystem,
					Name: "App",
					Subunits: map[string]*model.Unit{
						"api": {
							Type:     model.TypeContainer,
							Name:     "API",
							Expanded: []string{"api"},
							Subunits: map[string]*model.Unit{
								"auth": {
									Type:  model.TypeComponent, // C3 -> ComponentBorder = "#78A8D8"
									Name:  "Auth",
									Links: []model.Link{{Peer: "app.api.store"}},
								},
								"store": {
									Type: model.TypeComponent,
									Name: "Store",
								},
							},
						},
					},
				},
			},
		}

		v := view.GenerateC3View(m, "app.api")
		require.NotNil(t, v)
		g := graph.BuildGraph(v)

		require.Len(t, g.Edges, 1)
		assert.Equal(t, "#78A8D8", g.Edges[0].Color)
	})

	// Test: linkFrom (incoming links) get color from their source
	t.Run("linkFrom edge gets color from link source", func(t *testing.T) {
		t.Parallel()

		m := &parser.Model{
			Properties: model.Properties{Name: "Test"},
			Units: map[string]*model.Unit{
				"app": {
					Type:      model.TypeSystem,
					Name:      "App",
					LinksFrom: []model.Link{{Peer: "ext"}}, // ext -> app, source is ext
				},
				"ext": {
					Type: model.TypeSystemExternal, // -> SystemExternalBorder = "#8A8A8A"
					Name: "External",
				},
			},
		}

		v := view.GenerateC1View(m)
		g := graph.BuildGraph(v)

		require.Len(t, g.Edges, 1)
		assert.Equal(t, "#8A8A8A", g.Edges[0].Color)
	})
}

func TestBuildGraphEdgeLength(t *testing.T) {
	t.Parallel()

	// Test 1: Edge with length > 0 has MinLen set
	t.Run("edge with length > 0 has MinLen set", func(t *testing.T) {
		t.Parallel()

		m := &parser.Model{
			Properties: model.Properties{Name: "Test"},
			Units: map[string]*model.Unit{
				"app": {
					Type: model.TypeSystem,
					Name: "App",
					Links: []model.Link{
						{
							Peer:   "db",
							Length: 2,
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

		require.Len(t, g.Edges, 1)
		assert.Equal(t, 2, g.Edges[0].MinLen)
	})

	// Test 2: Edge with length 0 has MinLen 0
	t.Run("edge with length 0 has MinLen 0", func(t *testing.T) {
		t.Parallel()

		m := &parser.Model{
			Properties: model.Properties{Name: "Test"},
			Units: map[string]*model.Unit{
				"app": {
					Type: model.TypeSystem,
					Name: "App",
					Links: []model.Link{
						{
							Peer:   "db",
							Length: 0,
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

		require.Len(t, g.Edges, 1)
		assert.Equal(t, 0, g.Edges[0].MinLen)
	})

	// Test 3: Edge without length has MinLen 0 (default)
	t.Run("edge without length has MinLen 0", func(t *testing.T) {
		t.Parallel()

		m := &parser.Model{
			Properties: model.Properties{Name: "Test"},
			Units: map[string]*model.Unit{
				"app": {
					Type: model.TypeSystem,
					Name: "App",
					Links: []model.Link{
						{
							Peer: "db",
							// No Length field
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

		require.Len(t, g.Edges, 1)
		assert.Equal(t, 0, g.Edges[0].MinLen)
	})
}

// TestBuildGraphResolvedEdgeMinLen tests D-13: a link's length (minlen) takes
// effect only when BOTH drawn endpoints are the link's original units. When
// either endpoint is resolved to an ancestor, the synthesized edge carries no
// minlen.
func TestBuildGraphResolvedEdgeMinLen(t *testing.T) {
	t.Parallel()

	// Test 1: Resolved edge drops minlen — the peer resolves to the top-level
	// ancestor, so the authored length must NOT apply.
	t.Run("resolved edge drops minlen", func(t *testing.T) {
		t.Parallel()

		m := &parser.Model{
			Properties: model.Properties{Name: "Test"},
			Units: map[string]*model.Unit{
				"webUser": {
					Type: model.TypePersonExternal,
					Name: "Web User",
					Links: []model.Link{
						{Peer: "system.api.handler", Length: 2},
					},
				},
				"system": {
					Type: model.TypeSystem,
					Name: "System",
					Subunits: map[string]*model.Unit{
						"api": {
							Type: model.TypeSystem,
							Name: "API",
							Subunits: map[string]*model.Unit{
								"handler": {
									Type: model.TypeComponent,
									Name: "Handler",
								},
							},
						},
					},
				},
			},
		}

		v := view.GenerateC1View(m)
		g := graph.BuildGraph(v)

		require.Len(t, g.Edges, 1)
		assert.Zero(t, g.Edges[0].MinLen)
	})

	// Test 2: Direct pair keeps minlen — both drawn endpoints are the link's
	// original units, so the authored length applies.
	t.Run("direct pair keeps minlen", func(t *testing.T) {
		t.Parallel()

		m := &parser.Model{
			Properties: model.Properties{Name: "Test"},
			Units: map[string]*model.Unit{
				"app": {
					Type: model.TypeSystem,
					Name: "App",
					Links: []model.Link{
						{Peer: "db", Length: 2},
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
		assert.Equal(t, 2, g.Edges[0].MinLen)
	})
}

// TestBuildExpandedGraphBaselineDOT guards COMPAT-02: expanded-mode DOT output
// must stay semantically identical to the committed golden baseline
// (cmd/c4drill/testdata/multilevel.expanded.dot, D-01/D-02 contract: node/edge
// sets, attributes, cluster structure). The baseline is regenerated via
// `go run ./cmd/c4drill cmd/c4drill/testdata/multilevel.toml --format dot --expanded`.
// The comparison is order-insensitive (DI-1): the pinned go-graphviz fork emits
// map-order-dependent sibling ordering and layout geometry, so the normalized
// semantic form is compared, not the raw bytes.
func TestBuildExpandedGraphBaselineDOT(t *testing.T) {
	t.Parallel()

	m, err := parser.ParseFile("../../cmd/c4drill/testdata/multilevel.toml")
	require.NoError(t, err)

	valErrors := validator.Validate(m)
	require.Empty(t, valErrors, "model should be valid")

	v := view.GenerateExpandedView(m)
	g := graph.BuildExpandedGraph(v)

	// Expanded mode keeps the v1.7 2.0 prominence on every edge (D-02/D-04).
	for _, edge := range g.Edges {
		assert.InDelta(t, 2.0, edge.PenWidth, 0.001,
			"expanded-mode edge %s -> %s should keep penwidth 2.0", edge.Source, edge.Target)
	}

	dotData, err := render.RenderDOT(g)
	require.NoError(t, err)

	expected, err := os.ReadFile("../../cmd/c4drill/testdata/multilevel.expanded.dot")
	require.NoError(t, err)

	// DI-1: the pinned go-graphviz fork emits map-order-dependent sibling
	// ordering and layout geometry, so the golden comparison is
	// order-insensitive — canonical.Canonical normalizes both sides to a sorted,
	// geometry-stripped semantic form (D-02 contract) before comparison.
	require.Equal(t, canonical.Canonical(t, string(expected)), canonical.Canonical(t, string(dotData)),
		"expanded DOT must match the committed golden baseline semantically (COMPAT-02)")
}

func TestBuildGraphDeterministicOrder(t *testing.T) {
	t.Parallel()

	// Create units with explicit definition order (zeta, alpha, gamma)
	// Phase 26: Order is definition order, not alphabetical
	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		UnitOrder:  []string{"zeta", "alpha", "gamma"}, // Explicit definition order
		Units: map[string]*model.Unit{
			"zeta": {
				Type: model.TypeSystem,
				Name: "Zeta System",
				Links: []model.Link{
					{Peer: "alpha", Technology: "HTTP"},
				},
			},
			"alpha": {
				Type: model.TypeSystem,
				Name: "Alpha System",
				Links: []model.Link{
					{Peer: "gamma", Technology: "TCP"},
				},
			},
			"gamma": {
				Type: model.TypeDb,
				Name: "Gamma Database",
			},
		},
	}

	// Test 1: BuildGraph produces nodes in definition order
	t.Run("BuildGraph produces nodes in definition order", func(t *testing.T) {
		t.Parallel()

		v := view.GenerateC1View(m)
		g := graph.BuildGraph(v)

		require.Len(t, g.Nodes, 3)

		// Nodes should be in definition order (zeta, alpha, gamma)
		require.Equal(t, "zeta", g.Nodes[0].ID, "first node should be zeta (definition order)")
		require.Equal(t, "alpha", g.Nodes[1].ID, "second node should be alpha (definition order)")
		require.Equal(t, "gamma", g.Nodes[2].ID, "third node should be gamma (definition order)")
	})

	// Test 2: BuildGraph produces edges in definition order by source
	t.Run("BuildGraph produces edges in definition order by source", func(t *testing.T) {
		t.Parallel()

		v := view.GenerateC1View(m)
		g := graph.BuildGraph(v)

		require.Len(t, g.Edges, 2)

		// Edges should be in definition order by source (zeta first, then alpha)
		// zeta->alpha comes first, then alpha->gamma
		require.Equal(t, "zeta", g.Edges[0].Source, "first edge source should be zeta (definition order)")
		require.Equal(t, "alpha", g.Edges[0].Target, "first edge target should be alpha")
		require.Equal(t, "alpha", g.Edges[1].Source, "second edge source should be alpha (definition order)")
		require.Equal(t, "gamma", g.Edges[1].Target, "second edge target should be gamma")
	})

	// Test 3: Multiple calls produce identical output order
	t.Run("multiple calls produce identical order", func(t *testing.T) {
		t.Parallel()

		v := view.GenerateC1View(m)

		// Call BuildGraph multiple times and collect results
		orders := make([][]string, 0, 5)

		for range 5 {
			g := graph.BuildGraph(v)

			ids := make([]string, len(g.Nodes))
			for j, node := range g.Nodes {
				ids[j] = node.ID
			}

			orders = append(orders, ids)
		}

		// All orders should be identical
		for i := 1; i < len(orders); i++ {
			require.Equal(t, orders[0], orders[i], "all calls should produce same order")
		}
	})

	// Test 4: BuildExpandedGraph produces top-level items in definition order
	t.Run("BuildExpandedGraph produces top-level items in definition order", func(t *testing.T) {
		t.Parallel()

		m2 := &parser.Model{
			Properties: model.Properties{Name: "Test"},
			UnitOrder:  []string{"zeta", "alpha", "gamma"}, // Explicit definition order
			Units: map[string]*model.Unit{
				"zeta": {
					Type:         model.TypeSystem,
					Name:         "Zeta System",
					SubunitOrder: []string{"sub"},
					Subunits: map[string]*model.Unit{
						"sub": {Type: model.TypeContainer, Name: "Sub"},
					},
				},
				"alpha": {
					Type:         model.TypeSystem,
					Name:         "Alpha System",
					SubunitOrder: []string{"sub"},
					Subunits: map[string]*model.Unit{
						"sub": {Type: model.TypeContainer, Name: "Sub"},
					},
				},
				"gamma": {
					Type: model.TypeDb,
					Name: "Gamma Database",
					// No subunits - should be a node
				},
			},
		}

		v := view.GenerateExpandedView(m2)
		g := graph.BuildExpandedGraph(v)

		// Top-level clusters should be in definition order (zeta, alpha)
		require.Len(t, g.Clusters, 2)
		require.Equal(t, "zeta", g.Clusters[0].ID, "first cluster should be zeta (definition order)")
		require.Equal(t, "alpha", g.Clusters[1].ID, "second cluster should be alpha (definition order)")

		// Top-level nodes should be in alphabetical order
		require.Len(t, g.Nodes, 1)
		require.Equal(t, "gamma", g.Nodes[0].ID, "first node should be gamma")
	})

	// Test 5: buildCluster processes subunits in definition order
	t.Run("buildCluster processes subunits in definition order", func(t *testing.T) {
		t.Parallel()

		m2 := &parser.Model{
			Properties: model.Properties{Name: "Test"},
			Units: map[string]*model.Unit{
				"system": {
					Type:         model.TypeSystem,
					Name:         "System",
					Expanded:     []string{"system"},                 // Mark the system itself as expanded
					SubunitOrder: []string{"zeta", "alpha", "gamma"}, // Explicit definition order
					Subunits: map[string]*model.Unit{
						"zeta":  {Type: model.TypeContainer, Name: "Zeta"},
						"alpha": {Type: model.TypeContainer, Name: "Alpha"},
						"gamma": {Type: model.TypeContainer, Name: "Gamma"},
					},
				},
			},
		}

		v := view.GenerateC1View(m2)
		g := graph.BuildGraph(v)

		require.Len(t, g.Clusters, 1)
		cluster := g.Clusters[0]

		// Subunits should be in definition order (zeta, alpha, gamma)
		require.Len(t, cluster.Nodes, 3)
		require.Equal(t, "system.zeta", cluster.Nodes[0].ID, "first subunit should be zeta (definition order)")
		require.Equal(t, "system.alpha", cluster.Nodes[1].ID, "second subunit should be alpha (definition order)")
		require.Equal(t, "system.gamma", cluster.Nodes[2].ID, "third subunit should be gamma (definition order)")
	})

	// Test 6: buildNestedCluster processes subunits in definition order
	t.Run("buildNestedCluster processes subunits in definition order", func(t *testing.T) {
		t.Parallel()

		m2 := &parser.Model{
			Properties: model.Properties{Name: "Test"},
			Units: map[string]*model.Unit{
				"system": {
					Type:         model.TypeSystem,
					Name:         "System",
					SubunitOrder: []string{"zeta", "alpha", "gamma"}, // Explicit definition order
					Subunits: map[string]*model.Unit{
						"zeta": {
							Type:         model.TypeContainer,
							Name:         "Zeta",
							SubunitOrder: []string{"sub3", "sub1", "sub2"}, // Explicit definition order
							Subunits: map[string]*model.Unit{
								"sub3": {Type: model.TypeComponent, Name: "Sub3"},
								"sub1": {Type: model.TypeComponent, Name: "Sub1"},
								"sub2": {Type: model.TypeComponent, Name: "Sub2"},
							},
						},
						"alpha": {Type: model.TypeContainer, Name: "Alpha"},
						"gamma": {Type: model.TypeContainer, Name: "Gamma"},
					},
				},
			},
		}

		v := view.GenerateExpandedView(m2)
		g := graph.BuildExpandedGraph(v)

		require.Len(t, g.Clusters, 1)
		topCluster := g.Clusters[0]
		require.Equal(t, "system", topCluster.ID)

		// Top-level subunits should be in definition order (zeta as cluster, then alpha, gamma as nodes)
		// Nodes in cluster: alpha, gamma (in definition order after zeta cluster)
		require.Len(t, topCluster.Nodes, 2)
		require.Equal(t, "system.alpha", topCluster.Nodes[0].ID, "first node should be alpha (definition order)")
		require.Equal(t, "system.gamma", topCluster.Nodes[1].ID, "second node should be gamma")

		// Nested clusters: zeta
		require.Len(t, topCluster.Clusters, 1)
		zetaCluster := topCluster.Clusters[0]
		require.Equal(t, "system.zeta", zetaCluster.ID)

		// Zeta's subunits should be in definition order (sub3, sub1, sub2)
		require.Len(t, zetaCluster.Nodes, 3)
		require.Equal(t, "system.zeta.sub3", zetaCluster.Nodes[0].ID, "zeta sub3 should be first (definition order)")
		require.Equal(t, "system.zeta.sub1", zetaCluster.Nodes[1].ID, "zeta sub1 should be second (definition order)")
		require.Equal(t, "system.zeta.sub2", zetaCluster.Nodes[2].ID, "zeta sub2 should be third (definition order)")
	})
}

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

	// Test 3: BuildGraphWithPath computes correct breadcrumb navigation.
	// The back-link was dropped (breadcrumb-only nav); the root ancestor in the
	// breadcrumb trail now carries the up-navigation URL.
	t.Run("C2 view has breadcrumb navigation to root", func(t *testing.T) {
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
		// C2 breadcrumb: [root (Test), current (mainsystem)]
		require.Len(t, g.Navigation.Breadcrumbs, 2)
		// Root ancestor points back to the C1 diagram.
		assert.NotEmpty(t, g.Navigation.Breadcrumbs[0].URL,
			"root breadcrumb ancestor must carry a URL to the C1 diagram")
		// BackLink is intentionally nil (breadcrumb-only nav).
		assert.Nil(t, g.Navigation.BackLink)
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
		// Should have breadcrumbs: [Root Test] > Main System > API (current).
		// The root C1 context is always prepended so users can navigate all
		// the way up from any C2/C3 diagram.
		require.Len(t, g.Navigation.Breadcrumbs, 3)

		// Root context (Test) with URL back to C1
		assert.Equal(t, "Test", g.Navigation.Breadcrumbs[0].Name)
		assert.NotEmpty(t, g.Navigation.Breadcrumbs[0].URL)

		// Middle ancestor (Main System) with URL, pretty name
		assert.Equal(t, "Main System", g.Navigation.Breadcrumbs[1].Name)
		assert.NotEmpty(t, g.Navigation.Breadcrumbs[1].URL)

		// Current (API) — no URL, pretty name
		assert.Equal(t, "API", g.Navigation.Breadcrumbs[2].Name)
		assert.Empty(t, g.Navigation.Breadcrumbs[2].URL)

		// BackLink is intentionally nil (breadcrumb-only nav).
		assert.Nil(t, g.Navigation.BackLink)
	})
}

// TestBuildGraphDefinitionOrder tests that BuildGraph produces nodes in definition order.
func TestBuildGraphDefinitionOrder(t *testing.T) {
	t.Parallel()

	// Create model with explicit definition order (not alphabetical)
	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		UnitOrder:  []string{"zulu", "alpha", "gamma"},
		Units: map[string]*model.Unit{
			"zulu": {
				Type:  model.TypeSystem,
				Name:  "Zulu System",
				Links: []model.Link{{Peer: "alpha", Technology: "HTTP"}},
			},
			"alpha": {
				Type:  model.TypeSystem,
				Name:  "Alpha System",
				Links: []model.Link{{Peer: "gamma", Technology: "TCP"}},
			},
			"gamma": {
				Type: model.TypeDb,
				Name: "Gamma Database",
			},
		},
	}

	v := view.GenerateC1View(m)
	g := graph.BuildGraph(v)

	require.Len(t, g.Nodes, 3)

	// Nodes should be in DEFINITION order, not alphabetical
	require.Equal(t, "zulu", g.Nodes[0].ID, "first node should be zulu (definition order)")
	require.Equal(t, "alpha", g.Nodes[1].ID, "second node should be alpha (definition order)")
	require.Equal(t, "gamma", g.Nodes[2].ID, "third node should be gamma (definition order)")

	// Edges should also be in definition order by source
	require.Len(t, g.Edges, 2)
	require.Equal(t, "zulu", g.Edges[0].Source, "first edge source should be zulu")
	require.Equal(t, "alpha", g.Edges[1].Source, "second edge source should be alpha")
}

// TestBuildEdgesPairCollapse verifies D-01/D-03/D-06: multiple links landing on the
// same (source, target) pair collapse to a single edge in resolved views, with the
// first contributing link's attributes winning and no count suffix on the label.
func TestBuildEdgesPairCollapse(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"app": {
				Type: model.TypeSystem,
				Name: "App",
				Links: []model.Link{
					{Peer: "db", Technology: "SQL", Description: "first"},
					{Peer: "db", Technology: "HTTP", Description: "second"},
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

	g := graph.BuildGraph(v)

	// D-01: multiple links on the same pair collapse to a single edge.
	require.Len(t, g.Edges, 1)

	// D-03: the first link in definition order wins.
	assert.Equal(t, "SQL", g.Edges[0].Label.Technology)
	assert.Equal(t, "first", g.Edges[0].Label.Description)

	// D-06: labels stay plain — no count suffix, endpoints unchanged.
	assert.Equal(t, "app", g.Edges[0].Source)
	assert.Equal(t, "db", g.Edges[0].Target)
}

// TestBuildEdgesPenwidth verifies D-04: collapsed pairs (2+ contributing links)
// get penwidth 2.0, single-relationship edges stay at the renderer default (0).
func TestBuildEdgesPenwidth(t *testing.T) {
	t.Parallel()

	t.Run("single link stays at default penwidth", func(t *testing.T) {
		t.Parallel()

		m := &parser.Model{
			Properties: model.Properties{Name: "Test"},
			Units: map[string]*model.Unit{
				"app": {
					Type: model.TypeSystem,
					Name: "App",
					Links: []model.Link{
						{Peer: "db", Technology: "SQL"},
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
		assert.Zero(t, g.Edges[0].PenWidth, "single-relationship edges stay at default width (D-04)")
	})

	t.Run("collapsed pair thickens to 2.0", func(t *testing.T) {
		t.Parallel()

		m := &parser.Model{
			Properties: model.Properties{Name: "Test"},
			Units: map[string]*model.Unit{
				"app": {
					Type: model.TypeSystem,
					Name: "App",
					Links: []model.Link{
						{Peer: "db", Technology: "SQL"},
						{Peer: "db", Technology: "HTTP"},
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
		assert.InDelta(t, 2.0, g.Edges[0].PenWidth, 0.001, "collapsed pairs (2+ links) thicken to 2.0 (D-04)")
	})
}

// TestBuildEdgesPenwidthC2C3CollapsedPairs verifies WR-01: D-04/D-05 penwidth
// thickening fires in C2/C3 views. resolveSubunitCrossLinks must preserve ALL
// contributing links on ResolvedLinks (the builder's pair-only markSeen does
// the edge dedup), so countPairMultiplicity sees collapsed-pair multiplicity
// exactly as it does in C1.
func TestBuildEdgesPenwidthC2C3CollapsedPairs(t *testing.T) {
	t.Parallel()

	t.Run("C2 direct duplicate links thicken to 2.0", func(t *testing.T) {
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
							Links: []model.Link{
								{Peer: "mainsystem.db", Technology: "SQL"},
								{Peer: "mainsystem.db", Technology: "HTTP"},
							},
						},
						"db": {Type: model.TypeContainerDb, Name: "Database"},
					},
					SubunitOrder: []string{"api", "db"},
				},
			},
		}

		v := view.GenerateC2View(m, "mainsystem")
		require.NotNil(t, v)

		g := graph.BuildGraph(v)

		require.Len(t, g.Edges, 1, "two links on the same pair collapse to one edge")
		assert.Equal(t, "mainsystem.api", g.Edges[0].Source)
		assert.Equal(t, "mainsystem.db", g.Edges[0].Target)
		assert.InDelta(t, 2.0, g.Edges[0].PenWidth, 0.001, "collapsed C2 pairs thicken to 2.0 (D-04)")
	})

	t.Run("C2 descendant-contributed links thicken to 2.0", func(t *testing.T) {
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
								"a1": {
									Type: model.TypeComponent,
									Name: "A1",
									Links: []model.Link{
										{Peer: "mainsystem.db", Technology: "SQL"},
									},
								},
								"a2": {
									Type: model.TypeComponent,
									Name: "A2",
									Links: []model.Link{
										{Peer: "mainsystem.db", Technology: "HTTP"},
									},
								},
							},
							SubunitOrder: []string{"a1", "a2"},
						},
						"db": {Type: model.TypeContainerDb, Name: "Database"},
					},
					SubunitOrder: []string{"api", "db"},
				},
			},
		}

		v := view.GenerateC2View(m, "mainsystem")
		require.NotNil(t, v)

		g := graph.BuildGraph(v)

		require.Len(t, g.Edges, 1, "two descendant links resolve to one api -> db pair")
		assert.Equal(t, "mainsystem.api", g.Edges[0].Source)
		assert.Equal(t, "mainsystem.db", g.Edges[0].Target)
		assert.InDelta(t, 2.0, g.Edges[0].PenWidth, 0.001, "descendant contributions thicken the collapsed pair (D-05)")
	})

	t.Run("C3 direct duplicate links thicken to 2.0", func(t *testing.T) {
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
								"handler": {
									Type: model.TypeComponent,
									Name: "Handler",
									Links: []model.Link{
										{Peer: "mainsystem.api.repo", Technology: "SQL"},
										{Peer: "mainsystem.api.repo", Technology: "HTTP"},
									},
								},
								"repo": {Type: model.TypeComponentDb, Name: "Repository"},
							},
							SubunitOrder: []string{"handler", "repo"},
						},
					},
				},
			},
		}

		v := view.GenerateC3View(m, "mainsystem.api")
		require.NotNil(t, v)

		g := graph.BuildGraph(v)

		require.Len(t, g.Edges, 1, "two links on the same pair collapse to one edge")
		assert.Equal(t, "mainsystem.api.handler", g.Edges[0].Source)
		assert.Equal(t, "mainsystem.api.repo", g.Edges[0].Target)
		assert.InDelta(t, 2.0, g.Edges[0].PenWidth, 0.001, "collapsed C3 pairs thicken to 2.0 (D-04)")
	})
}

// TestBuildEdgesPenwidthLinkFromContributions verifies WR-02: D-05 counts
// authored linkFrom relationships toward pair multiplicity, while validator-
// synthesized mirrors (internal/validator/index.go) are excluded so they never
// double-count an outgoing link. Without the discrimination, an authored
// linkFrom on an already-outgoing pair is indistinguishable from a mirror and
// the multiplicity signal becomes authoring-dependent.
func TestBuildEdgesPenwidthLinkFromContributions(t *testing.T) {
	t.Parallel()

	t.Run("authored linkFrom contributes to multiplicity", func(t *testing.T) {
		t.Parallel()

		m := &parser.Model{
			Properties: model.Properties{Name: "Test"},
			Units: map[string]*model.Unit{
				"app": {
					Type: model.TypeSystem,
					Name: "App",
					Links: []model.Link{
						{Peer: "db", Technology: "SQL"},
					},
				},
				"db": {
					Type: model.TypeDb,
					Name: "Database",
					LinksFrom: []model.Link{
						{Peer: "app", Technology: "auth"},
					},
				},
			},
		}

		// Run validation like the command does. The validator synthesizes
		// mirrors for outgoing links, but skips the mirror here because db
		// already has an authored linkFrom to app (FindLinkByPeer).
		require.Empty(t, validator.Validate(m))

		v := view.GenerateC1View(m)
		g := graph.BuildGraph(v)

		require.Len(t, g.Edges, 1)
		assert.Equal(t, "app", g.Edges[0].Source)
		assert.Equal(t, "db", g.Edges[0].Target)
		assert.InDelta(t, 2.0, g.Edges[0].PenWidth, 0.001, "authored linkFrom counts toward multiplicity (D-05)")
	})

	t.Run("validator mirror does not double-count", func(t *testing.T) {
		t.Parallel()

		m := &parser.Model{
			Properties: model.Properties{Name: "Test"},
			Units: map[string]*model.Unit{
				"app": {
					Type: model.TypeSystem,
					Name: "App",
					Links: []model.Link{
						{Peer: "db", Technology: "SQL"},
					},
				},
				"db": {
					Type: model.TypeDb,
					Name: "Database",
					Links: []model.Link{
						{Peer: "app", Technology: "HTTP"},
					},
				},
			},
		}

		// Run validation like the command does — mirrors are added to LinksFrom
		// in both directions.
		require.Empty(t, validator.Validate(m))

		v := view.GenerateC1View(m)
		g := graph.BuildGraph(v)

		require.Len(t, g.Edges, 2)

		for _, edge := range g.Edges {
			assert.Zero(t, edge.PenWidth, "mirrors are synthetic duplicates, not contributing links (D-05)")
		}
	})
}

// TestBuildEdgesExpandedExemption verifies D-02/COMPAT-02: --expanded mode keeps
// the v1.7 technology+description dedup key and 2.0 penwidth prominence.
// GenerateExpandedView does not set AllExpanded yet (plan 01-02), so the view
// is constructed literally.
func TestBuildEdgesExpandedExemption(t *testing.T) {
	t.Parallel()

	v := &view.View{
		AllExpanded: true,
		UnitOrder:   []string{"app", "db"},
		Units: map[string]*view.Entry{
			"app": {
				Unit: &model.Unit{
					Type: model.TypeSystem,
					Name: "App",
					Links: []model.Link{
						{Peer: "db", Technology: "SQL"},
						{Peer: "db", Technology: "HTTP"},
					},
				},
				FullPath: "app",
			},
			"db": {
				Unit:     &model.Unit{Type: model.TypeDb, Name: "Database"},
				FullPath: "db",
			},
		},
	}

	g := graph.BuildGraph(v)

	// D-02: expanded mode keeps the v1.7 tech+desc dedup key (2 edges for 2 links).
	require.Len(t, g.Edges, 2)

	// D-02: expanded-mode edges keep the v1.7 2.0 prominence.
	assert.InDelta(t, 2.0, g.Edges[0].PenWidth, 0.001)
	assert.InDelta(t, 2.0, g.Edges[1].PenWidth, 0.001)
}

// TestBuildExpandedGraphIgnoresPropertiesExpanded locks D-04: --expanded mode
// ignores properties.expanded entirely — the flag expands EVERYTHING in one
// file (v1.7 contract).
func TestBuildExpandedGraphIgnoresPropertiesExpanded(t *testing.T) {
	t.Parallel()

	m, err := parser.ParseFile("../../cmd/c4drill/testdata/multilevel.toml")
	require.NoError(t, err)

	require.Empty(t, validator.Validate(m), "model should be valid")

	// Poison the global expansion hint: if a future change makes
	// GenerateExpandedView consult this, the view shrinks (D-04).
	m.Properties.Expanded = []string{"mainSystem"}

	v := view.GenerateExpandedView(m)
	total := countModelUnits(m)
	require.Len(t, v.Units, total, "expanded view must contain ALL units even with properties.expanded set (D-04)")

	g := graph.BuildExpandedGraph(v)
	require.NotNil(t, g)

	nodeIDs, labels := collectExpandedGraphNodes(g)
	assert.Contains(t, nodeIDs, "mainSystem.sshAuth.systemd.logind", "deep 4-level leaf must render as a node")
	assert.Contains(t, nodeIDs, "mainSystem.storages.externalStorage.client", "client unit must render as a node")

	for _, name := range labels {
		assert.NotContains(t, name, "🔍", "expanded mode must not render collapsed indicators (D-04)")
	}
}

// countModelUnits recursively counts all units in the model: each top-level
// unit plus every subunit at all nesting levels.
func countModelUnits(m *parser.Model) int {
	count := 0

	for _, unit := range m.Units {
		count += 1 + countUnitSubunits(unit)
	}

	return count
}

// countUnitSubunits recursively counts a unit's subunits at all nesting levels.
func countUnitSubunits(unit *model.Unit) int {
	count := 0

	for _, sub := range unit.Subunits {
		count += 1 + countUnitSubunits(sub)
	}

	return count
}

// collectExpandedGraphNodes returns all node IDs and label names in an expanded
// graph, walking top-level nodes plus a recursive walk of clusters
// (Cluster.Nodes + Cluster.Clusters per graph.go:105-120).
func collectExpandedGraphNodes(g *graph.Graph) ([]string, []string) {
	nodeIDs := make([]string, 0)
	labels := make([]string, 0)

	collectNodeIDsAndLabels(g.Nodes, &nodeIDs, &labels)
	collectClusterNodeIDsAndLabels(g.Clusters, &nodeIDs, &labels)

	return nodeIDs, labels
}

// collectClusterNodeIDsAndLabels recursively collects node IDs and label names
// from a cluster list (Cluster.Nodes + nested Cluster.Clusters).
func collectClusterNodeIDsAndLabels(clusters []*graph.Cluster, nodeIDs *[]string, labels *[]string) {
	for _, cluster := range clusters {
		if cluster.Label != nil {
			*labels = append(*labels, cluster.Label.Name)
		}

		collectNodeIDsAndLabels(cluster.Nodes, nodeIDs, labels)
		collectClusterNodeIDsAndLabels(cluster.Clusters, nodeIDs, labels)
	}
}

// collectNodeIDsAndLabels appends node IDs and non-nil label names to the
// given slices.
func collectNodeIDsAndLabels(nodes []*graph.Node, nodeIDs *[]string, labels *[]string) {
	for _, node := range nodes {
		*nodeIDs = append(*nodeIDs, node.ID)

		if node.Label != nil {
			*labels = append(*labels, node.Label.Name)
		}
	}
}

// edgeBlockFromDOT extracts the full DOT attribute block for the edge between
// source and target. The go-graphviz writer wraps edge attributes across
// multiple lines, so the block runs from the edge's first line (e.g.,
// "app -> db	[key=app_to_db,") to the line ending with "];". The DOT default
// edge block always carries penwidth=1.0, so assertions must target the
// per-edge block, not the whole document.
func edgeBlockFromDOT(t *testing.T, dot []byte, source, target string) string {
	t.Helper()

	prefix := source + " -> " + target
	lines := strings.Split(string(dot), "\n")

	for i, line := range lines {
		if !strings.HasPrefix(strings.TrimSpace(line), prefix) {
			continue
		}

		var block strings.Builder
		block.WriteString(line)

		for j := i + 1; j < len(lines); j++ {
			block.WriteString("\n")
			block.WriteString(lines[j])

			if strings.HasSuffix(strings.TrimSpace(lines[j]), "];") {
				return block.String()
			}
		}

		return block.String()
	}

	require.FailNowf(t, "edge %s -> %s not found in DOT output", source, target)

	return ""
}

// TestBuildEdgesPenwidthRendered verifies the penwidth values survive DOT
// serialization: 1.0 for single edges in resolved views, 2.0 for collapsed
// pairs in resolved views, 2.0 in expanded mode. DOT rendering is
// parallel-safe (precedent: TestBuildExpandedGraphRealToml).
func TestBuildEdgesPenwidthRendered(t *testing.T) {
	t.Parallel()

	t.Run("single edge renders penwidth=1 in resolved views", func(t *testing.T) {
		t.Parallel()

		m := &parser.Model{
			Properties: model.Properties{Name: "Test"},
			Units: map[string]*model.Unit{
				"app": {
					Type: model.TypeSystem,
					Name: "App",
					Links: []model.Link{
						{Peer: "db", Technology: "SQL"},
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

		dotData, err := render.RenderDOT(g)
		require.NoError(t, err)
		edgeBlock := edgeBlockFromDOT(t, dotData, "app", "db")
		assert.Contains(t, edgeBlock, "penwidth=1", "single edges render at 1.0 (D-04)")
	})

	t.Run("collapsed pair renders penwidth=2 in resolved views", func(t *testing.T) {
		t.Parallel()

		m := &parser.Model{
			Properties: model.Properties{Name: "Test"},
			Units: map[string]*model.Unit{
				"app": {
					Type: model.TypeSystem,
					Name: "App",
					Links: []model.Link{
						{Peer: "db", Technology: "SQL"},
						{Peer: "db", Technology: "HTTP"},
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

		dotData, err := render.RenderDOT(g)
		require.NoError(t, err)
		edgeBlock := edgeBlockFromDOT(t, dotData, "app", "db")
		assert.Contains(t, edgeBlock, "penwidth=2", "collapsed pairs render at 2.0 (D-04)")
	})

	t.Run("expanded mode renders penwidth=2", func(t *testing.T) {
		t.Parallel()

		v := &view.View{
			AllExpanded: true,
			UnitOrder:   []string{"app", "db"},
			Units: map[string]*view.Entry{
				"app": {
					Unit: &model.Unit{
						Type: model.TypeSystem,
						Name: "App",
						Links: []model.Link{
							{Peer: "db", Technology: "SQL", Description: "first"},
							{Peer: "db", Technology: "HTTP", Description: "second"},
						},
					},
					FullPath: "app",
				},
				"db": {
					Unit:     &model.Unit{Type: model.TypeDb, Name: "Database"},
					FullPath: "db",
				},
			},
		}

		g := graph.BuildGraph(v)

		dotData, err := render.RenderDOT(g)
		require.NoError(t, err)
		edgeBlock := edgeBlockFromDOT(t, dotData, "app", "db")
		assert.Contains(t, edgeBlock, "penwidth=2", "expanded mode keeps 2.0 (COMPAT-02)")
	})
}

// TestReferenceGlyph exercises REF-02: a unit with a non-empty Reference renders
// a visible 📖 marker appended to the node label name (mirroring the 🔍
// collapsed-cluster affordance at builder.go:258-261), and populates
// Node.ReferenceURL (REF-03 plumbing). A non-referenced unit renders no 📖 and
// an empty ReferenceURL.
func TestReferenceGlyph(t *testing.T) {
	t.Parallel()

	t.Run("referenced unit label has 📖 and ReferenceURL is set", func(t *testing.T) {
		t.Parallel()

		m := &parser.Model{
			Properties: model.Properties{Name: "Reference Test"},
			Units: map[string]*model.Unit{
				"app": {
					Type:      model.TypeSystem,
					Name:      "App",
					Reference: "https://example.com/docs/app",
				},
			},
		}

		v := view.GenerateC1View(m)
		g := graph.BuildGraph(v)

		require.Len(t, g.Nodes, 1)
		node := g.Nodes[0]

		// REF-02: label name contains the 📖 marker.
		assert.Contains(t, node.Label.Name, "📖", "referenced unit label must include the 📖 marker")
		// REF-03 plumbing: ReferenceURL carries the exact reference string.
		assert.Equal(t, "https://example.com/docs/app", node.ReferenceURL,
			"Node.ReferenceURL must carry the unit's reference URL exactly")
	})

	t.Run("non-referenced unit label has no 📖 and ReferenceURL is empty", func(t *testing.T) {
		t.Parallel()

		m := &parser.Model{
			Properties: model.Properties{Name: "Reference Test"},
			Units: map[string]*model.Unit{
				"app": {
					Type: model.TypeSystem,
					Name: "App",
				},
			},
		}

		v := view.GenerateC1View(m)
		g := graph.BuildGraph(v)

		require.Len(t, g.Nodes, 1)
		node := g.Nodes[0]

		assert.NotContains(t, node.Label.Name, "📖", "non-referenced unit label must NOT include 📖")
		assert.Empty(t, node.ReferenceURL, "non-referenced unit must have an empty ReferenceURL")
	})

	t.Run("expanded parent cluster label has 📖 when referenced", func(t *testing.T) {
		t.Parallel()

		// An expanded parent unit (rendered via buildCluster → buildClusterLabel)
		// carrying a reference must gain the 📖 glyph on its cluster label name.
		m := &parser.Model{
			Properties: model.Properties{Name: "Reference Cluster Test"},
			Units: map[string]*model.Unit{
				"mainsystem": {
					Type:      model.TypeSystem,
					Name:      "Main System",
					Reference: "https://example.com/docs/main",
					Subunits: map[string]*model.Unit{
						"api": {
							Type: model.TypeContainer,
							Name: "API",
						},
					},
				},
			},
		}

		// Expanded view → mainsystem renders as a cluster whose label is built
		// by buildClusterLabel.
		v := view.GenerateExpandedView(m)
		g := graph.BuildExpandedGraph(v)

		require.Len(t, g.Clusters, 1, "expanded mainsystem must render as a cluster")
		cluster := g.Clusters[0]
		require.NotNil(t, cluster.Label, "cluster must have a label")
		assert.Contains(t, cluster.Label.Name, "📖",
			"expanded referenced parent cluster label must include 📖")
	})
}

// TestReferenceURL_RenderedDOT exercises REF-03 (converter): createNode wires
// the external reference URL into the GraphViz node's single URL attribute.
// Also covers the external-wins precedence (REF-03/J) when BOTH ReferenceURL
// and ExploreURL are present on the same node.
func TestReferenceURL_RenderedDOT(t *testing.T) {
	t.Parallel()

	t.Run("referenced node DOT carries the external URL", func(t *testing.T) {
		t.Parallel()

		m := &parser.Model{
			Properties: model.Properties{Name: "Reference URL Test"},
			Units: map[string]*model.Unit{
				"app": {
					Type:      model.TypeSystem,
					Name:      "App",
					Reference: "https://example.com/docs/app",
				},
			},
		}

		v := view.GenerateC1View(m)
		g := graph.BuildGraph(v)

		dotData, err := render.RenderDOT(g)
		require.NoError(t, err)

		s := string(dotData)
		// The external reference URL must appear in the rendered DOT (URL=...).
		assert.Contains(t, s, "https://example.com/docs/app",
			"rendered DOT must carry the external reference URL on the node")
	})

	t.Run("external reference wins the single URL slot over explore URL", func(t *testing.T) {
		t.Parallel()

		// A collapsed system with subunits AND a reference: BuildGraphWithPath
		// would normally set ExploreURL for drill-down, but the external
		// ReferenceURL must win GraphViz's single URL slot (ARCHITECTURE-v1.10
		// §6 (6) Option A).
		m := &parser.Model{
			Properties: model.Properties{Name: "Reference Precedence Test"},
			Units: map[string]*model.Unit{
				"mainsystem": {
					Type:      model.TypeSystem,
					Name:      "Main System",
					Reference: "https://example.com/docs/main",
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
		node := g.Nodes[0]
		// Precondition: both URLs are populated on the Node struct.
		assert.Equal(t, "https://example.com/docs/main", node.ReferenceURL, "ReferenceURL must be set")
		assert.NotEmpty(t, node.ExploreURL, "ExploreURL must also be set (collapsed with subunits)")

		dotData, err := render.RenderDOT(g)
		require.NoError(t, err)

		s := string(dotData)
		// The external reference URL must win the single slot — the explore
		// URL (.svg drill-down) must NOT be the rendered URL for this node.
		assert.Contains(t, s, "https://example.com/docs/main",
			"rendered DOT must carry the external reference URL (external-wins precedence)")
	})
}

// TestReference_BackwardCompat exercises REF-05: a unit authored WITHOUT a
// reference field renders semantically identical to the committed v1.9 golden
// baseline. This is the existing COMPAT-02 golden (TestBuildExpandedGraphBaselineDOT)
// which renders cmd/c4drill/testdata/multilevel.toml — a fixture that has no
// reference fields — against the committed multilevel.expanded.dot baseline.
// The Phase 28 change MUST leave that golden comparison byte-for-byte identical
// (order-insensitive canonicalDOT comparison per STATE.md DI-1).
//
// This test is intentionally a thin alias that pins the REF-05 contract name
// alongside the existing COMPAT-02 regression guard so a failure here directly
// reports the backward-compat regression.
func TestReference_BackwardCompat(t *testing.T) {
	t.Parallel()

	m, err := parser.ParseFile("../../cmd/c4drill/testdata/multilevel.toml")
	require.NoError(t, err)

	valErrors := validator.Validate(m)
	require.Empty(t, valErrors, "model should be valid")

	v := view.GenerateExpandedView(m)
	g := graph.BuildExpandedGraph(v)

	dotData, err := render.RenderDOT(g)
	require.NoError(t, err)

	expected, err := os.ReadFile("../../cmd/c4drill/testdata/multilevel.expanded.dot")
	require.NoError(t, err)

	// REF-05 / DI-1: order-insensitive canonical comparison. A model with no
	// reference fields must render byte-identical (semantically) to v1.9.
	require.Equal(t, canonical.Canonical(t, string(expected)), canonical.Canonical(t, string(dotData)),
		"REF-05: a no-reference model must render identical to the v1.9 golden baseline")
}

func TestBuildGraphEdgeKindColour(t *testing.T) {
	t.Parallel()

	// KIND-01/KIND-02 precedence: explicit link.Color > kind colour > source border (D-01).
	tests := []struct {
		name string
		link model.Link
		want string
	}{
		{
			name: "kind read colours the edge green",
			link: model.Link{Peer: "db", Kind: model.KindRead},
			want: model.LinkReadColour,
		},
		{
			name: "kind write colours the edge red",
			link: model.Link{Peer: "db", Kind: model.KindWrite},
			want: model.LinkWriteColour,
		},
		{
			name: "kind read-write colours the edge purple",
			link: model.Link{Peer: "db", Kind: model.KindReadWrite},
			want: model.LinkReadWriteColour,
		},
		{
			name: "explicit color wins over kind",
			link: model.Link{Peer: "db", Kind: model.KindRead, Color: "#FF0000"},
			want: "#FF0000",
		},
		{
			name: "unknown kind falls back to source border",
			link: model.Link{Peer: "db", Kind: model.LinkKind("query")},
			want: "#073B6F", // PersonBorder for C1 system source
		},
		{
			name: "no kind keeps source border default",
			link: model.Link{Peer: "db"},
			want: "#073B6F", // PersonBorder for C1 system source
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := &parser.Model{
				Properties: model.Properties{Name: "Test"},
				Units: map[string]*model.Unit{
					"app": {
						Type:  model.TypeSystem,
						Name:  "App",
						Links: []model.Link{tt.link},
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
			assert.Equal(t, tt.want, g.Edges[0].Color)
		})
	}
}

func TestUnitStyleOverrides(t *testing.T) {
	t.Parallel()

	// COLOR-01/COLOR-02: author color/style/border fields override the type
	// palette on plain nodes AND clusters; unset fields keep the palette.

	baseModel := func(unit *model.Unit) *parser.Model {
		return &parser.Model{
			Properties: model.Properties{Name: "Test"},
			Units: map[string]*model.Unit{
				"app": unit,
				"db":  {Type: model.TypeDb, Name: "Database"},
			},
		}
	}

	t.Run("explicit dark color fills node and forces white font", func(t *testing.T) {
		t.Parallel()

		m := baseModel(&model.Unit{
			Type:  model.TypeSystem,
			Name:  "App",
			Color: "#08427B",
			Links: []model.Link{{Peer: "db"}},
		})

		g := graph.BuildGraph(view.GenerateC1View(m))
		require.Len(t, g.Nodes, 2)

		var app *graph.Node
		for _, n := range g.Nodes {
			if n.ID == "app" {
				app = n
			}
		}
		require.NotNil(t, app)
		assert.Equal(t, "#08427B", app.Style.FillColor)
		assert.Equal(t, "#FFFFFF", app.Style.FontColor, "dark fill forces white font (luminance rule)")
	})

	t.Run("light explicit color keeps the level font color", func(t *testing.T) {
		t.Parallel()

		m := baseModel(&model.Unit{
			Type:  model.TypeSystem,
			Name:  "App",
			Color: "#4A90D9",
		})

		g := graph.BuildGraph(view.GenerateC1View(m))

		var app *graph.Node
		for _, n := range g.Nodes {
			if n.ID == "app" {
				app = n
			}
		}
		require.NotNil(t, app)
		assert.Equal(t, "#4A90D9", app.Style.FillColor)
		assert.Equal(t, model.PersonBorder, app.Style.FontColor, "light fill keeps level font default")
	})

	t.Run("explicit border and dotted style override", func(t *testing.T) {
		t.Parallel()

		m := baseModel(&model.Unit{
			Type:   model.TypeSystem,
			Name:   "App",
			Border: "#AA0000",
			Style:  "dotted",
		})

		g := graph.BuildGraph(view.GenerateC1View(m))

		var app *graph.Node
		for _, n := range g.Nodes {
			if n.ID == "app" {
				app = n
			}
		}
		require.NotNil(t, app)
		assert.Equal(t, "#AA0000", app.Style.BorderColor)
		assert.Equal(t, "dotted", app.Style.BorderStyle)
	})

	t.Run("unset fields keep the type palette", func(t *testing.T) {
		t.Parallel()

		m := baseModel(&model.Unit{Type: model.TypeSystem, Name: "App"})

		g := graph.BuildGraph(view.GenerateC1View(m))

		var app *graph.Node
		for _, n := range g.Nodes {
			if n.ID == "app" {
				app = n
			}
		}
		require.NotNil(t, app)
		assert.Empty(t, app.Style.FillColor)
		assert.Equal(t, model.PersonBorder, app.Style.BorderColor)
		assert.Equal(t, model.PersonBorder, app.Style.FontColor)
		assert.Equal(t, "solid", app.Style.BorderStyle)
	})

	t.Run("box unit explicit color beats content-derived style", func(t *testing.T) {
		t.Parallel()

		m := &parser.Model{
			Properties: model.Properties{Name: "Test"},
			Units: map[string]*model.Unit{
				"grp": {
					Type:  model.TypeBox,
					Name:  "Group",
					Color: "#00AA00",
					Expanded: []string{"grp"},
					Subunits: map[string]*model.Unit{
						"inner": {Type: model.TypeSystem, Name: "Inner"},
					},
					SubunitOrder: []string{"inner"},
				},
			},
		}

		g := graph.BuildGraph(view.GenerateC1View(m))
		require.Len(t, g.Clusters, 1)
		assert.Equal(t, "#00AA00", g.Clusters[0].Style.FillColor,
			"author color wins over box-content styling")
	})

	t.Run("expanded unit cluster renders explicit color", func(t *testing.T) {
		t.Parallel()

		m := baseModel(&model.Unit{
			Type:  model.TypeSystem,
			Name:  "App",
			Color: "#123456",
			Expanded: []string{"app"},
			Subunits: map[string]*model.Unit{
				"api": {Type: model.TypeContainer, Name: "API"},
			},
			SubunitOrder: []string{"api"},
		})

		g := graph.BuildGraph(view.GenerateC1View(m))
		require.Len(t, g.Clusters, 1)
		assert.Equal(t, "#123456", g.Clusters[0].Style.FillColor)
	})
}

// TestPairAggregation pins AGG-01..03: collapsed pairs derive colour from the
// constituents' kinds (all-same → kind colour, mixed → read-write), line style
// follows precedence (all-same → it; any solid → solid; else any dashed →
// dashed; else dotted), and any explicit colour suppresses kind colouring
// (source-border default). Single-link pairs and AllExpanded mode are
// unaffected.
func TestPairAggregation(t *testing.T) {
	t.Parallel()

	// Two subunits of one system both link to the same external unit — in the
	// collapsed C1 view both links resolve to the app→ext ancestor pair, which
	// collapses to ONE edge. The pair's logical source is app (system), so the
	// D-01 default colour is PersonBorder #073B6F.
	buildModel := func(links ...model.Link) *parser.Model {
		return &parser.Model{
			Properties: model.Properties{Name: "Test"},
			Units: map[string]*model.Unit{
				"app": {
					Type: model.TypeSystem,
					Name: "App",
					Subunits: map[string]*model.Unit{
						"api": {Type: model.TypeContainer, Name: "API", Links: []model.Link{links[0]}},
						"etl": {Type: model.TypeContainer, Name: "ETL", Links: []model.Link{links[1]}},
					},
					SubunitOrder: []string{"api", "etl"},
				},
				"ext": {Type: model.TypeSystemExternal, Name: "Ext"},
			},
		}
	}

	edgeOf := func(t *testing.T, m *parser.Model) *graph.Edge {
		t.Helper()

		g := graph.BuildGraph(view.GenerateC1View(m))
		require.Len(t, g.Edges, 1, "pair collapses to one edge")

		return g.Edges[0]
	}

	tests := []struct {
		name       string
		a, b       model.Link
		wantColour string
		wantStyle  string
	}{
		{
			name:       "all read",
			a:          model.Link{Peer: "ext", Kind: model.KindRead},
			b:          model.Link{Peer: "ext", Kind: model.KindRead},
			wantColour: model.LinkReadColour,
			wantStyle:  "solid",
		},
		{
			name:       "all write",
			a:          model.Link{Peer: "ext", Kind: model.KindWrite},
			b:          model.Link{Peer: "ext", Kind: model.KindWrite},
			wantColour: model.LinkWriteColour,
			wantStyle:  "solid",
		},
		{
			name:       "mixed read+write",
			a:          model.Link{Peer: "ext", Kind: model.KindRead},
			b:          model.Link{Peer: "ext", Kind: model.KindWrite},
			wantColour: model.LinkReadWriteColour,
			wantStyle:  "solid",
		},
		{
			name:       "one kind unset suppresses kind colour",
			a:          model.Link{Peer: "ext", Kind: model.KindRead},
			b:          model.Link{Peer: "ext"},
			wantColour: "#073B6F", // ContainerBorder default (BC-safe)
			wantStyle:  "solid",
		},
		{
			name:       "explicit colour on one constituent suppresses kind colour",
			a:          model.Link{Peer: "ext", Kind: model.KindRead},
			b:          model.Link{Peer: "ext", Kind: model.KindRead, Color: "#FF00FF"},
			wantColour: "#073B6F", // AGG-03: default edge colour
			wantStyle:  "solid",
		},
		{
			name:       "all dashed",
			a:          model.Link{Peer: "ext", Style: "dashed"},
			b:          model.Link{Peer: "ext", Style: "dashed"},
			wantColour: "#073B6F",
			wantStyle:  "dashed",
		},
		{
			name:       "solid + dashed",
			a:          model.Link{Peer: "ext", Style: "solid"},
			b:          model.Link{Peer: "ext", Style: "dashed"},
			wantColour: "#073B6F",
			wantStyle:  "solid",
		},
		{
			name:       "unset (solid default) + dashed",
			a:          model.Link{Peer: "ext"},
			b:          model.Link{Peer: "ext", Style: "dashed"},
			wantColour: "#073B6F",
			wantStyle:  "solid",
		},
		{
			name:       "dashed + dotted",
			a:          model.Link{Peer: "ext", Style: "dashed"},
			b:          model.Link{Peer: "ext", Style: "dotted"},
			wantColour: "#073B6F",
			wantStyle:  "dashed",
		},
		{
			name:       "all dotted",
			a:          model.Link{Peer: "ext", Style: "dotted"},
			b:          model.Link{Peer: "ext", Style: "dotted"},
			wantColour: "#073B6F",
			wantStyle:  "dotted",
		},
		{
			name:       "kind colour with mixed styles keeps style precedence",
			a:          model.Link{Peer: "ext", Kind: model.KindWrite, Style: "dotted"},
			b:          model.Link{Peer: "ext", Kind: model.KindWrite, Style: "dotted"},
			wantColour: model.LinkWriteColour,
			wantStyle:  "dotted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			edge := edgeOf(t, buildModel(tt.a, tt.b))
			assert.Equal(t, tt.wantColour, edge.Color)
			assert.Equal(t, tt.wantStyle, edge.Style)
		})
	}

	t.Run("single link pair keeps per-edge semantics", func(t *testing.T) {
		t.Parallel()

		m := &parser.Model{
			Properties: model.Properties{Name: "Test"},
			Units: map[string]*model.Unit{
				"app": {
					Type: model.TypeSystem,
					Name: "App",
					Subunits: map[string]*model.Unit{
						"api": {
							Type:  model.TypeContainer,
							Name:  "API",
							Links: []model.Link{{Peer: "ext", Kind: model.KindRead, Style: "dashed"}},
						},
					},
					SubunitOrder: []string{"api"},
				},
				"ext": {Type: model.TypeSystemExternal, Name: "Ext"},
			},
		}

		g := graph.BuildGraph(view.GenerateC1View(m))
		require.Len(t, g.Edges, 1)
		assert.Equal(t, model.LinkReadColour, g.Edges[0].Color, "single-link pair uses kind colour directly")
		assert.Equal(t, "dashed", g.Edges[0].Style)
	})
}
