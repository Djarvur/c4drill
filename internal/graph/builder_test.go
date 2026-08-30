package graph_test

import (
	"fmt"
	"path/filepath"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/Djarvur/c4drill/internal/c4d"
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

// TestBuildGraph_ExpandedClusterRendersNestedSubClusters pins the compact-root
// C1 semantics for an author-expanded cluster (BUG-1-ROOT-COMPACT revision of
// CTX-03): a child unit that itself has subunits renders as a COLLAPSED node
// (with the 🔍 affordance) unless the view's visible paths reach inside it —
// only chain-visible children unfold as a NESTED cluster. The pre-compact
// behavior unfolded every grandchild into the root, flooding the non-expanded
// diagram into a de-facto --expanded render.
func TestBuildGraph_ExpandedClusterRendersNestedSubClusters(t *testing.T) {
	t.Parallel()

	// mainSystem is author-expanded, so its direct children render inside the
	// mainSystem cluster. Child auth is a container WITH subunits but nothing
	// links into it from outside — compact C1 keeps it collapsed. Child iam
	// has an external deep link to iam.iamApi (CTX-02 chain), so iam unfolds
	// as a nested cluster carrying its chain.
	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"webUser": {
				Type:  model.TypePersonExternal,
				Name:  "Web User",
				Links: []model.Link{{Peer: "mainSystem.iam.iamApi"}},
			},
			"mainSystem": {
				Type:         model.TypeSystem,
				Name:         "Main System",
				Expanded:     []string{"mainSystem"},
				SubunitOrder: []string{"auth", "auditLog", "iam"},
				Subunits: map[string]*model.Unit{
					"auth": {
						Type:         model.TypeContainer,
						Name:         "Auth",
						SubunitOrder: []string{"authApi", "authDb"},
						Subunits: map[string]*model.Unit{
							"authApi": {Type: model.TypeComponent, Name: "Auth API"},
							"authDb":  {Type: model.TypeComponent, Name: "Auth DB"},
						},
					},
					"auditLog": {Type: model.TypeComponent, Name: "Audit Log"},
					"iam": {
						Type:         model.TypeContainer,
						Name:         "IAM",
						SubunitOrder: []string{"iamApi"},
						Subunits: map[string]*model.Unit{
							"iamApi": {Type: model.TypeComponent, Name: "IAM API"},
						},
					},
				},
			},
		},
	}

	v := view.GenerateC1View(m)
	require.NotNil(t, v)

	g := graph.BuildGraph(v)

	// mainSystem (expanded, has subunits) renders as a cluster.
	require.Len(t, g.Clusters, 1)
	mainCluster := g.Clusters[0]
	require.Equal(t, "mainSystem", mainCluster.ID)

	// Compact C1: the auth subunit-container renders as a COLLAPSED NODE with
	// the 🔍 affordance — its unlinked grandchildren (authApi, authDb) stay
	// hidden (no whole sibling subtrees in the root).
	var authNode *graph.Node

	for _, node := range mainCluster.Nodes {
		if node.ID == "mainSystem.auth" {
			authNode = node
		}
	}

	require.NotNil(t, authNode,
		"unlinked subunit-container must render as a collapsed node inside the expanded cluster")
	require.NotNil(t, authNode.Label)
	assert.Contains(t, authNode.Label.Name, "🔍",
		"the collapsed container node keeps its drill affordance")
	for _, id := range []string{"mainSystem.auth.authApi", "mainSystem.auth.authDb"} {
		assert.NotContains(t, collectNodeIDs(g), id,
			"unlinked grandchildren must not flood the root")
	}

	// Chain-visible child: iam has a visible descendant (the CTX-02 chain to
	// iam.iamApi), so it renders as a NESTED cluster with the chain inside.
	require.Len(t, mainCluster.Clusters, 1)
	iamCluster := mainCluster.Clusters[0]
	require.Equal(t, "mainSystem.iam", iamCluster.ID)

	iamNodeIDs := make(map[string]bool, len(iamCluster.Nodes))
	for _, node := range iamCluster.Nodes {
		iamNodeIDs[node.ID] = true
	}

	assert.True(t, iamNodeIDs["mainSystem.iam.iamApi"],
		"the chain target renders inside the nested iam cluster")

	// Leaf dispatch unchanged: auditLog is a direct node of the mainSystem
	// cluster.
	auditLogFound := false

	for _, node := range mainCluster.Nodes {
		if node.ID == "mainSystem.auditLog" {
			auditLogFound = true
		}
	}

	assert.True(t, auditLogFound,
		"auditLog leaf child must render as a direct node of the mainSystem cluster")

	// The author-expanded mainSystem cluster label does not carry 🔍.
	require.NotNil(t, mainCluster.Label)
	assert.NotContains(t, mainCluster.Label.Name, "🔍",
		"the author-expanded mainSystem cluster label must NOT carry 🔍")

	// Drill affordance URL: BuildGraphWithPath's walk must REACH nested
	// clusters and assign the iam cluster its explore URL.
	gWithPath := graph.BuildGraphWithPath(v, "", "diagram", "svg")
	require.Len(t, gWithPath.Clusters, 1)

	var iamClusterWithPath *graph.Cluster

	for _, nested := range gWithPath.Clusters[0].Clusters {
		if nested.ID == "mainSystem.iam" {
			iamClusterWithPath = nested
			break
		}
	}

	require.NotNil(t, iamClusterWithPath, "nested iam cluster exists after BuildGraphWithPath")
	assert.Equal(t, graph.ComputeExploreURL("", "mainSystem.iam", "diagram", "svg"), iamClusterWithPath.ExploreURL,
		"the nested container cluster must get its explore URL assigned by the recursive walk")
}

// TestBuildGraph_DeepLinkTargetRendersInsideUnfoldedCluster pins the CTX-02
// graph contract: a collapsed ancestor whose chain was inserted for a deep
// link target renders as a RECURSIVE cluster (the UnfoldChain dispatch), the
// true target renders as a real node inside its chain, and the link edge
// terminates at that node — no dangling endpoint that graphviz would
// materialize as an implicit top-level node.
func TestBuildGraph_DeepLinkTargetRendersInsideUnfoldedCluster(t *testing.T) {
	t.Parallel()

	// Same minimal shape as the view-layer deepLinkChainModel (view package):
	// webApp links to linuxSystem.sshAuth.sshd while linuxSystem is NOT
	// author-expanded — the chain must unfold it.
	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"webApp": {
				Type: model.TypeContainer,
				Name: "Web App",
				Links: []model.Link{
					{Peer: "linuxSystem.sshAuth.sshd"},
				},
			},
			"linuxSystem": {
				Type: model.TypeSystem,
				Name: "Linux System",
				Subunits: map[string]*model.Unit{
					"sshAuth": {
						Type: model.TypeContainer,
						Name: "SSH Auth",
						Subunits: map[string]*model.Unit{
							"sshd": {Type: model.TypeContainer, Name: "SSHD"},
						},
					},
				},
			},
		},
	}

	v := view.GenerateC1View(m)
	require.NotNil(t, v)

	g := graph.BuildGraph(v)
	require.NotNil(t, g)

	// (i) linuxSystem must NOT render as a plain top-level node.
	for _, node := range g.Nodes {
		assert.NotEqual(t, "linuxSystem", node.ID,
			"the chain-bearing collapsed ancestor must not render as a plain top-level node")
	}

	// (ii) it renders as a cluster containing the nested chain cluster with
	// the true target inside.
	require.Len(t, g.Clusters, 1, "linuxSystem must unfold as a cluster")

	var sshAuthCluster *graph.Cluster

	for _, nested := range g.Clusters[0].Clusters {
		if nested.ID == "linuxSystem.sshAuth" {
			sshAuthCluster = nested
			break
		}
	}

	require.NotNil(t, sshAuthCluster,
		"linuxSystem cluster must contain the nested linuxSystem.sshAuth chain cluster")

	sshdInside := false

	for _, node := range sshAuthCluster.Nodes {
		if node.ID == "linuxSystem.sshAuth.sshd" {
			sshdInside = true
		}
	}

	assert.True(t, sshdInside, "the true link target must render inside its container chain")

	// (iii) some edge terminates at the TRUE target — no dangling endpoint
	// (graphviz would materialize an implicit top-level node for it).
	trueTargetEdge := false

	for _, edge := range g.Edges {
		if edge.Source == "linuxSystem.sshAuth.sshd" || edge.Target == "linuxSystem.sshAuth.sshd" {
			trueTargetEdge = true
		}
	}

	assert.True(t, trueTargetEdge, "an edge must terminate at the true deep-link target")
}

// TestC1RootStaysCompactOnDeepCrossFixture pins BUG-1-ROOT-COMPACT: the
// non-expanded root (C1) of the deepcross fixture — deep nesting, cross-system
// links, external actors deep-linking into nested subunits, author-expanded
// nested-heavy systems — is a COMPACT context diagram. It depicts exactly:
// top-level units (the author-expanded systems as clusters), the expanded
// units' direct visible subunits, external/boundary nodes, and CTX-02
// deep-link chain paths to true targets. Never whole sibling subtrees or the
// entire model (the reported regression: root ≈ expanded, 55 → 260 titles).
//
// Expected set derived by hand from testdata/deepcross.toml:
//   - boundary: adminCli, metricsScraper, mobileUser, webUser;
//   - actors (expanded) → identity cluster with chain leaves sessionMgr and
//     tokenApi (deep links from webUser/mobileUser from OUTSIDE actors);
//     audit renders as its collapsed node (its children ledger/archive are
//     linked only from INSIDE actors — internal detail, not visible);
//   - yic (expanded) → pipeline ⊃ ingest ⊃ api (chain from adminCli's
//     linkFrom and actors' cross-system archive link), store ⊃ warehouse
//     (chain from metricsScraper); writer is linked only from inside yic.
//
// The test parses the fixture TOML directly (precedent:
// TestViewAncestorChainInvariant) so the pin and the committed golden share
// one source of truth.
func TestC1RootStaysCompactOnDeepCrossFixture(t *testing.T) {
	t.Parallel()

	m, err := parser.ParseFile(filepath.Join("..", "..", "cmd", "c4drill", "testdata", "deepcross.toml"))
	require.NoError(t, err, "parse deepcross fixture")

	v := view.GenerateC1View(m)
	require.NotNil(t, v)

	g := graph.BuildGraph(v)
	require.NotNil(t, g)

	expected := []string{
		"adminCli", "metricsScraper", "mobileUser", "webUser",
		"actors.audit",
		"actors.identity.sessionMgr", "actors.identity.tokenApi",
		"yic.pipeline.ingest.api", "yic.pipeline.store.warehouse",
	}

	sort.Strings(expected)

	assert.Equal(t, expected, collectNodeIDs(g),
		"the non-expanded root must depict exactly the compact C1 set — no whole sibling subtrees")

	// Every edge endpoint must be depicted: a node, or a cluster ID — a
	// resolved edge may legitimately point at a depicted container; how the
	// converter emits edges whose target is a cluster is pre-existing
	// behavior outside this pin.
	depicted := make(map[string]bool, len(expected))

	for _, id := range expected {
		depicted[id] = true
	}

	for _, c := range collectClusterTree(g.Clusters) {
		depicted[c.ID] = true
	}

	for _, e := range g.Edges {
		assert.True(t, depicted[e.Source], "edge source %s must be depicted", e.Source)
		assert.True(t, depicted[e.Target], "edge target %s must be depicted", e.Target)
	}
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

// nodeByID finds one node by ID, failing the test when absent.
//
//nolint:unparam // id is "app" today; the helper stays generic for future callers
func nodeByID(t *testing.T, g *graph.Graph, id string) *graph.Node {
	t.Helper()

	for _, n := range g.Nodes {
		if n.ID == id {
			return n
		}
	}

	t.Fatalf("node %q not found", id)

	return nil
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

		app := nodeByID(t, g, "app")
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

		app := nodeByID(t, g, "app")
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

		app := nodeByID(t, g, "app")
		assert.Equal(t, "#AA0000", app.Style.BorderColor)
		assert.Equal(t, "dotted", app.Style.BorderStyle)
	})

	t.Run("unset fields keep the type palette", func(t *testing.T) {
		t.Parallel()

		m := baseModel(&model.Unit{Type: model.TypeSystem, Name: "App"})

		g := graph.BuildGraph(view.GenerateC1View(m))

		app := nodeByID(t, g, "app")
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
					Type:     model.TypeBox,
					Name:     "Group",
					Color:    "#00AA00",
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
			Type:     model.TypeSystem,
			Name:     "App",
			Color:    "#123456",
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

func TestBuildLegend(t *testing.T) {
	t.Parallel()

	t.Run("rows only for what the view shows", func(t *testing.T) {
		t.Parallel()

		m := &parser.Model{
			Properties: model.Properties{Name: "Test"},
			Units: map[string]*model.Unit{
				"app": {Type: model.TypeSystem, Name: "App"},
				"ext": {Type: model.TypeSystemExternal, Name: "Ext"},
			},
		}

		g := graph.BuildGraph(view.GenerateC1View(m))
		require.NotNil(t, g.Legend, "legend default-on")

		// One row per entity kind: system (internal, C1 blue) + external
		// system (grey). No links carry kind colours, so no kind rows — the
		// legend explains nothing the diagram does not show.
		require.Len(t, g.Legend.Entries, 2)
		assert.Equal(t, "system", g.Legend.Entries[0].Label)
		assert.Equal(t, model.PersonBorder, g.Legend.Entries[0].Color)
		assert.Equal(t, "external system", g.Legend.Entries[1].Label)
		assert.Equal(t, model.PersonExternalBorder, g.Legend.Entries[1].Color)
	})

	t.Run("element rows one per entity kind present", func(t *testing.T) {
		t.Parallel()

		v := &view.View{
			ShowLegend: true,
			Units: map[string]*view.Entry{
				"p": {Unit: &model.Unit{Type: model.TypePerson}},
				"c": {Unit: &model.Unit{Type: model.TypeContainerBox}},
				"k": {Unit: &model.Unit{Type: model.TypeComponent}},
				"d": {Unit: &model.Unit{Type: model.TypeComponentDb}},
			},
		}

		g := graph.BuildGraph(v)
		require.NotNil(t, g.Legend)
		require.Len(t, g.Legend.Entries, 4)
		assert.Equal(t, "person", g.Legend.Entries[0].Label)
		assert.Equal(t, model.PersonBorder, g.Legend.Entries[0].Color)
		assert.Equal(t, "container", g.Legend.Entries[1].Label)
		assert.Equal(t, model.ContainerBorder, g.Legend.Entries[1].Color, "containerBox folds into container")
		assert.Equal(t, "component", g.Legend.Entries[2].Label)
		assert.Equal(t, model.ComponentBorder, g.Legend.Entries[2].Color)
		assert.Equal(t, "component db", g.Legend.Entries[3].Label, "db labels carry their level")
		assert.Equal(t, model.ComponentBorder, g.Legend.Entries[3].Color)
	})

	t.Run("external rows one per entity kind", func(t *testing.T) {
		t.Parallel()

		v := &view.View{
			ShowLegend: true,
			Units: map[string]*view.Entry{
				"pe": {Unit: &model.Unit{Type: model.TypePersonExternal}, IsExternal: true},
				"de": {Unit: &model.Unit{Type: model.TypeDbExternal}, IsExternal: true},
				"se": {Unit: &model.Unit{Type: model.TypeSystemExternal}, IsExternal: true},
			},
		}

		g := graph.BuildGraph(v)
		require.NotNil(t, g.Legend)
		require.Len(t, g.Legend.Entries, 3)

		labels := make([]string, 0, 3)
		for _, entry := range g.Legend.Entries {
			labels = append(labels, entry.Label)
			assert.Equal(t, model.PersonExternalBorder, entry.Color, "all C1 externals share the grey")
		}

		assert.Equal(t, []string{"external person", "external system", "external db"}, labels)
	})

	t.Run("expanded unit contributes its boundary cluster row", func(t *testing.T) {
		t.Parallel()

		v := &view.View{
			ShowLegend:        true,
			ExpandedUnit:      "app",
			ExpandedUnitModel: &model.Unit{Type: model.TypeSystem, Name: "App"},
			Units: map[string]*view.Entry{
				"app.api": {Unit: &model.Unit{Type: model.TypeContainer}},
			},
		}

		g := graph.BuildGraph(v)
		require.NotNil(t, g.Legend)
		require.Len(t, g.Legend.Entries, 2)
		assert.Equal(t, "system", g.Legend.Entries[0].Label, "boundary cluster colour explained")
		assert.Equal(t, "container", g.Legend.Entries[1].Label)
	})

	t.Run("kind rows only for colours drawn on edges", func(t *testing.T) {
		t.Parallel()

		m := &parser.Model{
			Properties: model.Properties{Name: "Test"},
			Units: map[string]*model.Unit{
				"app": {
					Type: model.TypeSystem,
					Name: "App",
					Links: []model.Link{
						{Peer: "db", Kind: model.KindRead},
						{Peer: "cache", Kind: model.KindWrite},
					},
				},
				"db":    {Type: model.TypeDb, Name: "DB"},
				"cache": {Type: model.TypeSystem, Name: "Cache"},
			},
		}

		g := graph.BuildGraph(view.GenerateC1View(m))
		require.NotNil(t, g.Legend)

		// Elements (system, system db) + the two kinds actually used.
		// read-write is not on any edge, so it must not be advertised.
		require.Len(t, g.Legend.Entries, 4)
		assert.Equal(t, "system", g.Legend.Entries[0].Label)
		assert.Equal(t, "system db", g.Legend.Entries[1].Label)
		assert.Equal(t, "read", g.Legend.Entries[2].Label)
		assert.Equal(t, model.LinkReadColour, g.Legend.Entries[2].Color)
		assert.Equal(t, "write", g.Legend.Entries[3].Label)
		assert.Equal(t, model.LinkWriteColour, g.Legend.Entries[3].Color)
	})

	t.Run("explicit colour wins over kind, row dropped", func(t *testing.T) {
		t.Parallel()

		m := &parser.Model{
			Properties: model.Properties{Name: "Test"},
			Units: map[string]*model.Unit{
				"app": {
					Type:  model.TypeSystem,
					Name:  "App",
					Links: []model.Link{{Peer: "db", Kind: model.KindRead, Color: "#FF0000"}},
				},
				"db": {Type: model.TypeDb, Name: "DB"},
			},
		}

		g := graph.BuildGraph(view.GenerateC1View(m))
		require.NotNil(t, g.Legend)

		// The edge draws red, not the kind green — a "read" row would lie.
		require.Len(t, g.Legend.Entries, 2)
		assert.Equal(t, "system", g.Legend.Entries[0].Label)
		assert.Equal(t, "system db", g.Legend.Entries[1].Label)
	})

	t.Run("custom lines after defaults, colourless falls back to grey", func(t *testing.T) {
		t.Parallel()

		m := &parser.Model{
			Properties: model.Properties{
				Name: "Test",
				LegendLines: []model.LegendLine{
					{Label: "Batch", Color: "#C0392B", Style: "dashed"},
					{Label: "Plain"},
				},
			},
			Units: map[string]*model.Unit{
				"app": {Type: model.TypeSystem, Name: "App"},
			},
		}

		g := graph.BuildGraph(view.GenerateC1View(m))
		require.NotNil(t, g.Legend)
		require.Len(t, g.Legend.Entries, 3)

		assert.Equal(t, "system", g.Legend.Entries[0].Label)

		// Custom line after defaults, verbatim; the Style hint is a leftover
		// of the swatch legend and no longer renders.
		assert.Equal(t, "Batch", g.Legend.Entries[1].Label)
		assert.Equal(t, "#C0392B", g.Legend.Entries[1].Color)

		// Colourless custom line renders in the muted secondary grey.
		assert.Equal(t, "Plain", g.Legend.Entries[2].Label)
		assert.Equal(t, model.ArrowColor, g.Legend.Entries[2].Color)
	})

	t.Run("nothing to explain yields no legend", func(t *testing.T) {
		t.Parallel()

		m := &parser.Model{
			Properties: model.Properties{Name: "Test"},
			Units:      map[string]*model.Unit{},
		}

		g := graph.BuildGraph(view.GenerateC1View(m))
		assert.Nil(t, g.Legend, "empty view explains nothing")
	})

	t.Run("legend=false disables", func(t *testing.T) {
		t.Parallel()

		falsy := false

		m := &parser.Model{
			Properties: model.Properties{Name: "Test", Legend: &falsy},
			Units: map[string]*model.Unit{
				"app": {Type: model.TypeSystem, Name: "App"},
			},
		}

		g := graph.BuildGraph(view.GenerateC1View(m))
		assert.Nil(t, g.Legend, "legend=false yields no legend")
	})
}

// TestNoFeatureModelStability pins BC-01 at the source level: a model using
// none of the v1.13 features round-trips through the canonical TOML writer
// with NO legend/kind keys emitted, and canonical-equal after the round trip.
// Source-level byte stability for no-feature models is what keeps convert
// output stable across the upgrade.
func TestNoFeatureModelStability(t *testing.T) {
	t.Parallel()

	m, err := parser.ParseFile("../../cmd/c4drill/testdata/multilevel.toml")
	require.NoError(t, err, "fixture parses")

	out, err := c4d.EmitTOML(m)
	require.NoError(t, err, "EmitTOML")

	assert.NotContains(t, out, "legend", "no legend keys on a no-feature model")
	assert.NotContains(t, out, "\nkind = ", "no kind keys on a no-feature model")

	reparsed, err := parser.Parse([]byte(out))
	require.NoError(t, err, "round-trip parses")
	assert.True(t, c4d.CanonicalEqual(m, reparsed), "round-trip is canonically equal")
}

// Plain-guard tests (PLAIN-01/PLAIN-02): with View.Plain set, author-custom
// formatting falls back to type-palette/builder defaults at the graph layer —
// unit color/style/border overrides are skipped (including expanded-unit
// clusters), link color/style/length/rank/label-position are neutralized, and
// properties.edges is ignored — while kind-derived edge colours and the legend
// survive. Plain defaults to false, so the no-flag path is pinned by
// TestBuildGraph_DefaultPathUnchanged (BC-01 at the builder level).

func TestBuildGraph_PlainSkipsUnitOverrides(t *testing.T) {
	t.Parallel()

	t.Run("node falls back to the type palette", func(t *testing.T) {
		t.Parallel()

		m := &parser.Model{
			Properties: model.Properties{Name: "Test"},
			Units: map[string]*model.Unit{
				"app": {
					Type:   model.TypeSystem,
					Name:   "App",
					Color:  "#08427B",
					Border: "#AA0000",
					Style:  "dotted",
				},
				"db": {Type: model.TypeDb, Name: "Database"},
			},
		}

		v := view.GenerateC1View(m)
		v.Plain = true
		g := graph.BuildGraph(v)

		// Same assertions TestUnitStyleOverrides uses for the unset-fields
		// case: the node style must equal the type-palette default the unit
		// would receive with no overrides.
		app := nodeByID(t, g, "app")
		assert.Empty(t, app.Style.FillColor, "author color ignored under plain")
		assert.Equal(t, model.PersonBorder, app.Style.BorderColor, "author border ignored under plain")
		assert.Equal(t, model.PersonBorder, app.Style.FontColor)
		assert.Equal(t, "solid", app.Style.BorderStyle, "author style ignored under plain")
	})

	t.Run("expanded-unit cluster falls back to the type palette", func(t *testing.T) {
		t.Parallel()

		m := &parser.Model{
			Properties: model.Properties{Name: "Test"},
			Units: map[string]*model.Unit{
				"app": {
					Type:     model.TypeSystem,
					Name:     "App",
					Color:    "#123456",
					Border:   "#AA0000",
					Expanded: []string{"app"},
					Subunits: map[string]*model.Unit{
						"api": {Type: model.TypeContainer, Name: "API"},
					},
					SubunitOrder: []string{"api"},
				},
			},
		}

		v := view.GenerateC1View(m)
		v.Plain = true
		g := graph.BuildGraph(v)

		require.Len(t, g.Clusters, 1)
		style := g.Clusters[0].Style
		assert.Empty(t, style.FillColor, "author color ignored under plain")
		assert.Equal(t, model.PersonBorder, style.BorderColor, "author border ignored under plain")
		assert.Equal(t, "solid", style.BorderStyle, "author style ignored under plain")
	})

	t.Run("C2 boundary cluster falls back to the type palette", func(t *testing.T) {
		t.Parallel()

		m := &parser.Model{
			Properties: model.Properties{Name: "Test"},
			Units: map[string]*model.Unit{
				"app": {
					Type:     model.TypeSystem,
					Name:     "App",
					Color:    "#123456",
					Expanded: []string{"app"},
					Subunits: map[string]*model.Unit{
						"api": {Type: model.TypeContainer, Name: "API"},
					},
					SubunitOrder: []string{"api"},
				},
			},
		}

		v := view.GenerateC2View(m, "app")
		require.NotNil(t, v)
		v.Plain = true
		g := graph.BuildGraph(v)

		require.Len(t, g.Clusters, 1)
		style := g.Clusters[0].Style
		assert.Empty(t, style.FillColor, "author color ignored under plain")
		assert.Equal(t, model.PersonBorder, style.BorderColor, "author border ignored under plain")
	})
}

func TestBuildGraph_PlainNeutralizesEdgeFormatting(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test", Edges: "straight"},
		Units: map[string]*model.Unit{
			"app": {
				Type: model.TypeSystem,
				Name: "App",
				Links: []model.Link{{
					Peer:          "db",
					Color:         "#BA4A00",
					Style:         "dashed",
					Length:        3,
					Rank:          model.RankReverse,
					LabelPosition: model.LabelTail,
					Kind:          model.KindWrite,
				}},
			},
			"db": {Type: model.TypeDb, Name: "Database"},
		},
	}

	v := view.GenerateC1View(m)
	v.Plain = true
	g := graph.BuildGraph(v)

	require.Len(t, g.Edges, 1)
	edge := g.Edges[0]

	assert.Equal(t, model.LinkWriteColour, edge.Color,
		"custom colour ignored; the kind-derived colour applies (kind colours STAY)")
	assert.NotEqual(t, "#BA4A00", edge.Color, "author colour absent under plain")
	assert.Equal(t, "solid", edge.Style, "custom line style falls back to the default")
	assert.Zero(t, edge.MinLen, "length ignored — no minlen")
	assert.False(t, edge.RankReverse, "rank=reverse ignored — no endpoint swap")
	assert.False(t, edge.NoConstraint, "rank-based constraint suppression absent")
	assert.Equal(t, "middle", edge.Label.Position, "label position falls back to the builder default")
	assert.Empty(t, g.EdgeStyle, "properties.edges ignored under plain")

	// Collapsed pairs must not leak author styles either: the AGG-02
	// aggregate style override stays inert under plain while the aggregate's
	// kind/source-border colour logic still applies.
	t.Run("collapsed pair keeps the default style", func(t *testing.T) {
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
							Links: []model.Link{{Peer: "ext", Style: "dashed"}},
						},
						"etl": {
							Type:  model.TypeContainer,
							Name:  "ETL",
							Links: []model.Link{{Peer: "ext", Style: "dashed"}},
						},
					},
					SubunitOrder: []string{"api", "etl"},
				},
				"ext": {Type: model.TypeSystemExternal, Name: "Ext"},
			},
		}

		v := view.GenerateC1View(m)
		v.Plain = true
		g := graph.BuildGraph(v)

		require.Len(t, g.Edges, 1)
		assert.Equal(t, "solid", g.Edges[0].Style, "aggregate style override inert under plain")
	})
}

func TestBuildGraph_PlainKeepsKindColourAndLegend(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"app": {
				Type:  model.TypeSystem,
				Name:  "App",
				Links: []model.Link{{Peer: "db", Kind: model.KindRead}},
			},
			"db": {Type: model.TypeDb, Name: "Database"},
		},
	}

	v := view.GenerateC1View(m)
	v.Plain = true
	g := graph.BuildGraph(v)

	require.Len(t, g.Edges, 1)
	assert.Equal(t, model.LinkReadColour, g.Edges[0].Color,
		"kind-derived edge colour is semantic, not author formatting — it STAYS")

	require.NotNil(t, g.Legend, "legend survives plain mode")

	found := false

	for _, entry := range g.Legend.Entries {
		if entry.Label == "read" && entry.Color == model.LinkReadColour {
			found = true
		}
	}

	assert.True(t, found, "the drawn kind colour keeps its legend row")
}

func TestBuildGraph_PlainCopiedFromView(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"app": {
				Type:     model.TypeSystem,
				Name:     "App",
				Expanded: []string{"app"},
				Subunits: map[string]*model.Unit{
					"api": {Type: model.TypeContainer, Name: "API"},
				},
				SubunitOrder: []string{"api"},
			},
		},
	}

	v := view.GenerateC1View(m)
	v.Plain = true

	assert.True(t, graph.BuildGraph(v).Opts.Plain, "BuildGraph copies View.Plain")
	assert.True(t, graph.BuildExpandedGraph(v).Opts.Plain,
		"BuildExpandedGraph copies View.Plain too (--plain x --expanded, Pitfall 5)")
}

// TestBuildGraph_DefaultPathUnchanged pins BC-01 at the builder level: with
// Plain at its zero value the builder reproduces today's behavior exactly —
// author unit overrides apply and custom link colour/style/length/rank and
// label position all take effect.
func TestBuildGraph_DefaultPathUnchanged(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test", Edges: "straight"},
		Units: map[string]*model.Unit{
			"app": {
				Type:   model.TypeSystem,
				Name:   "App",
				Color:  "#08427B",
				Border: "#AA0000",
				Style:  "dotted",
				Links: []model.Link{{
					Peer:          "db",
					Color:         "#BA4A00",
					Style:         "dashed",
					Length:        3,
					Rank:          model.RankReverse,
					LabelPosition: model.LabelTail,
				}},
			},
			"db": {Type: model.TypeDb, Name: "Database"},
		},
	}

	v := view.GenerateC1View(m) // Plain keeps its zero value (false)
	g := graph.BuildGraph(v)

	assert.False(t, g.Opts.Plain, "default path leaves Graph.Opts.Plain false")

	app := nodeByID(t, g, "app")
	assert.Equal(t, "#08427B", app.Style.FillColor, "author color applied on the default path")
	assert.Equal(t, "#AA0000", app.Style.BorderColor, "author border applied on the default path")
	assert.Equal(t, "dotted", app.Style.BorderStyle, "author style applied on the default path")

	require.Len(t, g.Edges, 1)
	edge := g.Edges[0]
	assert.Equal(t, "#BA4A00", edge.Color, "custom link colour applied on the default path")
	assert.Equal(t, "dashed", edge.Style)
	assert.Equal(t, 3, edge.MinLen)
	assert.True(t, edge.RankReverse)
	assert.Equal(t, "tail", edge.Label.Position)
	assert.Equal(t, "straight", g.EdgeStyle, "properties.edges honored on the default path")
}

// ---- Phase 38 WRAP-01/02/03: ancestor wrapper clusters on C2/C3 views ----
//
// v1.15 scoping reversal: boundary/sibling entries and the expanded unit render
// INSIDE their complete ancestor-container chains. Fully external entries stay
// top-level. Wrapping is cluster-structure only: node IDs and edge endpoints
// are unchanged.

// wrapTestMultilevelModel parses and validates the shared multilevel fixture.
func wrapTestMultilevelModel(t *testing.T) *parser.Model {
	t.Helper()

	m, err := parser.ParseFile("../../cmd/c4drill/testdata/multilevel.toml")
	require.NoError(t, err)

	valErrors := validator.Validate(m)
	require.Empty(t, valErrors, "model should be valid")

	return m
}

// collectClusterTree walks the graph's cluster tree recursively and returns
// every cluster (including nested ones) in depth-first order.
func collectClusterTree(clusters []*graph.Cluster) []*graph.Cluster {
	out := make([]*graph.Cluster, 0)

	var walk func(cs []*graph.Cluster)

	walk = func(cs []*graph.Cluster) {
		for _, c := range cs {
			out = append(out, c)
			walk(c.Clusters)
		}
	}

	walk(clusters)

	return out
}

// collectNodeIDs returns the sorted multiset of every node ID reachable in the
// graph: top-level nodes plus nodes inside the recursive cluster tree.
func collectNodeIDs(g *graph.Graph) []string {
	ids := make([]string, 0, len(g.Nodes))
	for _, n := range g.Nodes {
		ids = append(ids, n.ID)
	}

	for _, c := range collectClusterTree(g.Clusters) {
		for _, n := range c.Nodes {
			ids = append(ids, n.ID)
		}
	}

	sort.Strings(ids)

	return ids
}

// TestBoundaryEntryWrappedInAncestorChain pins WRAP-01: on a C3 drill-down of
// mainSystem.storages.localStorage, a boundary entry resolving to
// mainSystem.sshAuth (an in-model sibling branch) appears inside a wrapper
// cluster for mainSystem — not at g.Nodes top level.
func TestBoundaryEntryWrappedInAncestorChain(t *testing.T) {
	t.Parallel()

	m := wrapTestMultilevelModel(t)

	v := view.GenerateC3View(m, "mainSystem.storages.localStorage")
	require.NotNil(t, v)
	g := graph.BuildGraph(v)

	for _, id := range g.Nodes {
		assert.NotEqual(t, "mainSystem.sshAuth", id.ID,
			"boundary entry with an in-model ancestor must not hang at top level")
	}

	// Find the wrapper cluster for mainSystem and the boundary node inside it.
	var mainWrapper *graph.Cluster

	for _, c := range collectClusterTree(g.Clusters) {
		if c.ID == "wrap_"+mainSystemPathSegment {
			mainWrapper = c
		}
	}

	require.NotNil(t, mainWrapper, "wrapper cluster for mainSystem must exist")
	found := false

	for _, n := range mainWrapper.Nodes {
		if n.ID == "mainSystem.sshAuth" {
			found = true
		}
	}

	assert.True(t, found, "mainSystem.sshAuth must render inside the mainSystem wrapper cluster")
}

// mainSystemPathSegment is the top-level unit of the multilevel fixture.
const mainSystemPathSegment = "mainSystem"

// TestFullyExternalBoundaryStaysTopLevel pins WRAP-02's exception: an entry
// with no in-model ancestor (externalSys at model root) remains a top-level
// node in g.Nodes.
func TestFullyExternalBoundaryStaysTopLevel(t *testing.T) {
	t.Parallel()

	m := wrapTestMultilevelModel(t)

	// In the C2 view of mainSystem, externalSys is linked from
	// mainSystem.storages.externalStorage.client and diverges at the model
	// root — a fully external boundary entry.
	v := view.GenerateC2View(m, "mainSystem")
	g := graph.BuildGraph(v)

	found := false

	for _, n := range g.Nodes {
		if n.ID == "externalSys" {
			found = true
		}
	}

	assert.True(t, found, "fully external boundary entry stays top-level")
}

// TestExpandedUnitAncestorSkeleton pins WRAP-02: a C3 view of
// mainSystem.storages.localStorage shows wrapper clusters mainSystem ⊃
// mainSystem.storages containing the expanded unit's boundary cluster; wrapper
// labels use AncestorNames pretty names.
func TestExpandedUnitAncestorSkeleton(t *testing.T) {
	t.Parallel()

	m := wrapTestMultilevelModel(t)

	v := view.GenerateC3View(m, "mainSystem.storages.localStorage")
	g := graph.BuildGraph(v)

	require.Len(t, g.Clusters, 1, "exactly one root cluster: the outermost wrapper")
	root := g.Clusters[0]
	assert.Equal(t, "wrap_"+mainSystemPathSegment, root.ID)
	assert.Equal(t, "Main System", root.Label.Name, "wrapper label uses the AncestorNames pretty name")

	require.Len(t, root.Clusters, 1)
	storages := root.Clusters[0]
	assert.Equal(t, "wrap_mainSystem.storages", storages.ID)
	assert.Equal(t, "Storages Registry", storages.Label.Name)

	require.Len(t, storages.Clusters, 1)
	boundary := storages.Clusters[0]
	assert.Equal(t, "mainSystem.storages.localStorage", boundary.ID,
		"the expanded unit's boundary cluster nests innermost")
	require.NotEmpty(t, boundary.Nodes, "expanded unit's subunits render inside it")
}

// TestWrapperClustersHaveNoExploreURL pins the affordance rule: wrapper
// clusters are containers, not drill targets — no ExploreURL; affordances stay
// only on the expanded unit's own cluster subtree.
func TestWrapperClustersHaveNoExploreURL(t *testing.T) {
	t.Parallel()

	m := wrapTestMultilevelModel(t)

	v := view.GenerateC3View(m, "mainSystem.storages.localStorage")
	g := graph.BuildGraphWithPath(v, "mainSystem/storages/localStorage", "multilevel", "svg")

	for _, c := range collectClusterTree(g.Clusters) {
		if strings.HasPrefix(c.ID, "wrap_") {
			assert.Empty(t, c.ExploreURL, "wrapper cluster %s must not carry a drill affordance", c.ID)
		}
	}
}

// TestBoundaryViewNodeSetInvariant pins WRAP-03: wrapping changes cluster
// structure only. Over the multilevel fixture's C3 view of
// mainSystem.storages.localStorage, the multiset of node IDs in the built
// graph is exactly the view's renderable entry set (v.UnitOrder) — no node
// lost, duplicated, or renamed by wrapping.
func TestBoundaryViewNodeSetInvariant(t *testing.T) {
	t.Parallel()

	m := wrapTestMultilevelModel(t)

	v := view.GenerateC3View(m, "mainSystem.storages.localStorage")
	g := graph.BuildGraph(v)

	expected := append([]string(nil), v.UnitOrder...)
	sort.Strings(expected)

	assert.ElementsMatch(t, expected, collectNodeIDs(g),
		"wrapping must not change the depicted node set (WRAP-03)")
}

// TestEdgeEndpointsUnchangedByWrapping locks the edge contract: wrapping adds
// no new endpoint IDs — every edge source/target is a depicted node ID, and no
// endpoint references a wrapper cluster.
func TestEdgeEndpointsUnchangedByWrapping(t *testing.T) {
	t.Parallel()

	m := wrapTestMultilevelModel(t)

	v := view.GenerateC3View(m, "mainSystem.storages.localStorage")
	g := graph.BuildGraph(v)

	nodeSet := make(map[string]bool, len(v.UnitOrder))
	for _, id := range collectNodeIDs(g) {
		nodeSet[id] = true
	}

	require.NotEmpty(t, g.Edges)

	for _, e := range g.Edges {
		assert.False(t, strings.HasPrefix(e.Source, "wrap_"), "edge source must never be a wrapper")
		assert.False(t, strings.HasPrefix(e.Target, "wrap_"), "edge target must never be a wrapper")
		assert.True(t, nodeSet[e.Source], "edge source %s must be a depicted node", e.Source)
		assert.True(t, nodeSet[e.Target], "edge target %s must be a depicted node", e.Target)
	}
}

// ---- Phase 38 KEY-01/KEY-02: granular suppression switches ----
//
// --no-colors / --no-styles / --no-length / --no-rank each suppress exactly
// one formatting aspect; --plain remains the UNION of the four (KEY-02): the
// plain path must be identical with and without the granular flags set.
// Guard mapping (plan-pinned): NoColors → applyUnitOverrides skips Color/
// Border fills AND createEdge skips the author link colour AND kind-derived
// colours; NoStyles → Style/BorderStyle + the applyCollapsedPairStyle
// aggregate style override; NoLength → MinLen 0; NoRank → RankReverse/
// NoConstraint false (which neutralizes all converter emission sites).

// TestNoColorsSuppressesAllColouring: with NoColors, author unit Color/Border
// fills are skipped, the author link colour is skipped, and kind-derived
// colouring is absent — the edge falls back to the D-01 source-border default
// and the legend drops its kind rows emergently (legendKindEntries reads the
// final edge colours).
func TestNoColorsSuppressesAllColouring(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"app": {
				Type:   model.TypeSystem,
				Name:   "App",
				Color:  "#123456",
				Border: "#AA0000",
				Links: []model.Link{{
					Peer: "db",
					Kind: model.KindRead,
				}},
			},
			"db": {Type: model.TypeDb, Name: "Database"},
		},
	}

	v := view.GenerateC1View(m)
	v.NoColors = true
	g := graph.BuildGraph(v)

	app := nodeByID(t, g, "app")
	assert.Empty(t, app.Style.FillColor, "author color fill skipped under --no-colors")
	assert.Equal(t, model.PersonBorder, app.Style.BorderColor,
		"author border skipped — type-palette default applies")

	require.Len(t, g.Edges, 1)
	edge := g.Edges[0]
	assert.NotEqual(t, model.LinkReadColour, edge.Color, "kind-derived colour suppressed")
	assert.Equal(t, model.PersonBorder, edge.Color,
		"edge falls back to the D-01 source-border default (structural colour, not decoration)")

	require.NotNil(t, g.Legend, "legend survives --no-colors")
	for _, entry := range g.Legend.Entries {
		assert.NotEqual(t, "read", entry.Label,
			"kind legend rows drop emergently when the kind colour is no longer drawn")
		assert.NotEqual(t, model.LinkReadColour, entry.Color)
	}
}

// TestNoStylesSuppressesStyleAndBorders: with NoStyles, author unit Style is
// skipped (BorderStyle falls back to solid), the author link style falls back
// to solid, and the applyCollapsedPairStyle aggregate style override is not
// applied.
func TestNoStylesSuppressesStyleAndBorders(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"app": {
				Type:  model.TypeSystem,
				Name:  "App",
				Style: "dotted",
				Links: []model.Link{{Peer: "db", Style: "dashed"}},
			},
			"db": {Type: model.TypeDb, Name: "Database"},
		},
	}

	v := view.GenerateC1View(m)
	v.NoStyles = true
	g := graph.BuildGraph(v)

	app := nodeByID(t, g, "app")
	assert.Equal(t, "solid", app.Style.BorderStyle, "author style skipped under --no-styles")

	require.Len(t, g.Edges, 1)
	assert.Equal(t, "solid", g.Edges[0].Style, "author link style falls back to solid")
	assert.Equal(t, model.PersonBorder, g.Edges[0].Color, "colour aspects are NOT touched by --no-styles (source-border default)")
	assert.Equal(t, 0, g.Edges[0].MinLen) // fixture has no length anyway

	t.Run("collapsed pair keeps the default style", func(t *testing.T) {
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
							Links: []model.Link{{Peer: "ext", Style: "dashed"}},
						},
						"etl": {
							Type:  model.TypeContainer,
							Name:  "ETL",
							Links: []model.Link{{Peer: "ext", Style: "dashed"}},
						},
					},
					SubunitOrder: []string{"api", "etl"},
				},
				"ext": {Type: model.TypeSystemExternal, Name: "Ext"},
			},
		}

		v := view.GenerateC1View(m)
		v.NoStyles = true
		g := graph.BuildGraph(v)

		require.Len(t, g.Edges, 1)
		assert.Equal(t, "solid", g.Edges[0].Style,
			"aggregate style override not applied under --no-styles")
	})
}

// TestNoLengthSuppressesMinlen: with NoLength, edge.MinLen stays 0 despite
// link.Length = 3.
func TestNoLengthSuppressesMinlen(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"app": {
				Type:  model.TypeSystem,
				Name:  "App",
				Links: []model.Link{{Peer: "db", Length: 3}},
			},
			"db": {Type: model.TypeDb, Name: "Database"},
		},
	}

	v := view.GenerateC1View(m)
	v.NoLength = true
	g := graph.BuildGraph(v)

	require.Len(t, g.Edges, 1)
	assert.Zero(t, g.Edges[0].MinLen, "length suppressed under --no-length")
}

// TestNoRankSuppressesRanking: with NoRank, RankReverse and NoConstraint stay
// false despite link.Rank reverse/equal — which automatically neutralizes all
// three converter emission sites (endpoint swap, dir inversion, constraint).
func TestNoRankSuppressesRanking(t *testing.T) {
	t.Parallel()

	t.Run("rank=reverse yields no endpoint swap", func(t *testing.T) {
		t.Parallel()

		m := &parser.Model{
			Properties: model.Properties{Name: "Test"},
			Units: map[string]*model.Unit{
				"app": {
					Type:  model.TypeSystem,
					Name:  "App",
					Links: []model.Link{{Peer: "db", Rank: model.RankReverse}},
				},
				"db": {Type: model.TypeDb, Name: "Database"},
			},
		}

		v := view.GenerateC1View(m)
		v.NoRank = true
		g := graph.BuildGraph(v)

		require.Len(t, g.Edges, 1)
		assert.False(t, g.Edges[0].RankReverse, "rank=reverse suppressed under --no-rank")
	})

	t.Run("rank=equal yields no constraint suppression", func(t *testing.T) {
		t.Parallel()

		m := &parser.Model{
			Properties: model.Properties{Name: "Test"},
			Units: map[string]*model.Unit{
				"app": {
					Type:  model.TypeSystem,
					Name:  "App",
					Links: []model.Link{{Peer: "db", Rank: model.RankEqual}},
				},
				"db": {Type: model.TypeDb, Name: "Database"},
			},
		}

		v := view.GenerateC1View(m)
		v.NoRank = true
		g := graph.BuildGraph(v)

		require.Len(t, g.Edges, 1)
		assert.False(t, g.Edges[0].NoConstraint, "rank=equal suppressed under --no-rank")
	})
}

// TestDefaultPathUnchangedByOptFlags: with all four granular flags false,
// colors/styles/minlen/rank are all present exactly as without the fields —
// the switches are strictly opt-in (BC contract).
func TestDefaultPathUnchangedByOptFlags(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"app": {
				Type:   model.TypeSystem,
				Name:   "App",
				Color:  "#08427B",
				Border: "#AA0000",
				Style:  "dotted",
				Links: []model.Link{{
					Peer:   "db",
					Color:  "#BA4A00",
					Style:  "dashed",
					Length: 3,
					Rank:   model.RankReverse,
				}},
			},
			"db": {Type: model.TypeDb, Name: "Database"},
		},
	}

	v := view.GenerateC1View(m)
	// NoColors/NoStyles/NoLength/NoRank keep their zero values (false).
	g := graph.BuildGraph(v)

	app := nodeByID(t, g, "app")
	assert.Equal(t, "#08427B", app.Style.FillColor, "author color applied on the default path")
	assert.Equal(t, "#AA0000", app.Style.BorderColor, "author border applied on the default path")
	assert.Equal(t, "dotted", app.Style.BorderStyle, "author style applied on the default path")

	require.Len(t, g.Edges, 1)
	edge := g.Edges[0]
	assert.Equal(t, "#BA4A00", edge.Color)
	assert.Equal(t, "dashed", edge.Style)
	assert.Equal(t, 3, edge.MinLen)
	assert.True(t, edge.RankReverse)
}

// TestPlainImpliesAllAspects (KEY-02, builder level): Plain=true yields the
// same suppression as Plain + all four granular flags set — union semantics.
// NOTE: kind-derived colours are semantic and survive plain (v1.14 contract,
// pinned by the plain goldens), so the union lock holds precisely because the
// granular guards defer to plain on kind colouring.
func TestPlainImpliesAllAspects(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test", Edges: "straight"},
		Units: map[string]*model.Unit{
			"app": {
				Type:   model.TypeSystem,
				Name:   "App",
				Color:  "#08427B",
				Border: "#AA0000",
				Style:  "dotted",
				Links: []model.Link{{
					Peer:   "db",
					Color:  "#BA4A00",
					Style:  "dashed",
					Length: 3,
					Rank:   model.RankReverse,
					Kind:   model.KindRead,
				}},
			},
			"db": {Type: model.TypeDb, Name: "Database"},
		},
	}

	vPlain := view.GenerateC1View(m)
	vPlain.Plain = true

	vUnion := view.GenerateC1View(m)
	vUnion.Plain = true
	vUnion.NoColors = true
	vUnion.NoStyles = true
	vUnion.NoLength = true
	vUnion.NoRank = true

	gPlain := graph.BuildGraph(vPlain)
	gUnion := graph.BuildGraph(vUnion)

	assert.Equal(t, gPlain.EdgeStyle, gUnion.EdgeStyle)

	appPlain := nodeByID(t, gPlain, "app")

	var appUnion *graph.Node

	require.Len(t, gUnion.Nodes, 2)
	for _, node := range gUnion.Nodes {
		if node.ID == "app" {
			appUnion = node
		}
	}

	require.NotNil(t, appUnion)
	assert.Equal(t, appPlain.Style, appUnion.Style, "unit styling identical under plain and plain+switches")

	require.Len(t, gPlain.Edges, 1)
	require.Len(t, gUnion.Edges, 1)
	assert.Equal(t, gPlain.Edges[0].Color, gUnion.Edges[0].Color, "edge colour identical (kind colour survives both)")
	assert.Equal(t, gPlain.Edges[0].Style, gUnion.Edges[0].Style)
	assert.Equal(t, gPlain.Edges[0].MinLen, gUnion.Edges[0].MinLen)
	assert.Equal(t, gPlain.Edges[0].RankReverse, gUnion.Edges[0].RankReverse)
	assert.Equal(t, gPlain.Edges[0].NoConstraint, gUnion.Edges[0].NoConstraint)
}

// TestSwitchCombination (Task 3 lock): pairwise and full combinations of the
// granular switches each suppress exactly the union of their aspects — and
// nothing else — over the post-WRAP multilevel boundary view, with the 38-01
// wrapper-cluster nesting intact in every combination.
func TestSwitchCombination(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"mainSystem": {
				Type: model.TypeSystem,
				Name: "Main System",
				Subunits: map[string]*model.Unit{
					"storages": {
						Type: model.TypeContainer,
						Name: "Storages",
						Subunits: map[string]*model.Unit{
							"localStorage": {
								Type: model.TypeContainer,
								Name: "Local Storage",
								Subunits: map[string]*model.Unit{
									"client": {
										Type:  model.TypeComponent,
										Name:  "Client",
										Color: "#AA0000",
										Style: "dotted",
										Links: []model.Link{{
											Peer:   "externalSys",
											Color:  "#BA4A00",
											Style:  "dashed",
											Length: 3,
											Rank:   model.RankReverse,
											Kind:   model.KindWrite,
										}},
									},
								},
							},
						},
					},
				},
			},
			"externalSys": {Type: model.TypeSystemExternal, Name: "External Sys"},
		},
	}

	type opts struct {
		noColors, noStyles, noLength, noRank bool
	}

	build := func(o opts) *graph.Graph {
		v := view.GenerateC3View(m, "mainSystem.storages.localStorage")
		v.NoColors = o.noColors
		v.NoStyles = o.noStyles
		v.NoLength = o.noLength
		v.NoRank = o.noRank

		return graph.BuildGraph(v)
	}

	findClient := func(t *testing.T, g *graph.Graph) *graph.Node {
		t.Helper()

		for _, id := range collectNodeIDs(g) {
			if id == "mainSystem.storages.localStorage.client" {
				for _, n := range g.Nodes {
					if n.ID == id {
						return n
					}
				}

				for _, c := range collectClusterTree(g.Clusters) {
					for _, n := range c.Nodes {
						if n.ID == id {
							return n
						}
					}
				}
			}
		}

		t.Fatal("client node not found")

		return nil
	}

	findEdge := func(t *testing.T, g *graph.Graph) *graph.Edge {
		t.Helper()

		require.Len(t, g.Edges, 1)
		require.Equal(t, "mainSystem.storages.localStorage.client", g.Edges[0].Source)

		return g.Edges[0]
	}

	assertWrappingIntact := func(t *testing.T, g *graph.Graph) {
		t.Helper()

		found := false

		for _, c := range collectClusterTree(g.Clusters) {
			if c.ID == "wrap_mainSystem.storages" {
				found = true
			}
		}

		assert.True(t, found, "wrapper-cluster nesting from 38-01 must stay intact under flag combinations")
	}

	// --no-colors + --no-length: colour and length suppressed; style and rank
	// still applied.
	g := build(opts{noColors: true, noLength: true})
	assertWrappingIntact(t, g)

	client := findClient(t, g)
	assert.Empty(t, client.Style.FillColor, "colour aspect suppressed")
	assert.Equal(t, "dotted", client.Style.BorderStyle, "style aspect kept")

	e := findEdge(t, g)
	assert.NotEqual(t, "#BA4A00", e.Color, "author link colour suppressed")
	assert.NotEqual(t, model.LinkWriteColour, e.Color, "kind colour suppressed")
	assert.Zero(t, e.MinLen, "length aspect suppressed")
	assert.Equal(t, "dashed", e.Style, "style aspect kept")
	assert.True(t, e.RankReverse, "rank aspect kept")

	// --no-styles + --no-rank: style and rank suppressed; colour and length
	// still applied.
	g = build(opts{noStyles: true, noRank: true})
	assertWrappingIntact(t, g)

	client = findClient(t, g)
	assert.Equal(t, "#AA0000", client.Style.FillColor, "colour aspect kept")
	assert.Equal(t, "solid", client.Style.BorderStyle, "style aspect suppressed")

	e = findEdge(t, g)
	assert.Equal(t, "#BA4A00", e.Color, "colour aspect kept")
	assert.Equal(t, 3, e.MinLen, "length aspect kept")
	assert.Equal(t, "solid", e.Style, "style aspect suppressed")
	assert.False(t, e.RankReverse, "rank aspect suppressed")
	assert.False(t, e.NoConstraint, "rank aspect suppressed")

	// All four: full suppression — the granular union of everything except
	// the plain-only label-position/properties.edges aspects.
	g = build(opts{noColors: true, noStyles: true, noLength: true, noRank: true})
	assertWrappingIntact(t, g)

	client = findClient(t, g)
	assert.Empty(t, client.Style.FillColor)
	assert.Equal(t, "solid", client.Style.BorderStyle)

	e = findEdge(t, g)
	assert.NotEqual(t, "#BA4A00", e.Color)
	assert.NotEqual(t, model.LinkWriteColour, e.Color)
	assert.Equal(t, "solid", e.Style)
	assert.Zero(t, e.MinLen)
	assert.False(t, e.RankReverse)
	assert.False(t, e.NoConstraint)
}

// ---- quick 260831-01u BUG-2: --no-labels suppresses EDGE labels only ----
//
// Re-baselined semantics (superseding the phase-38 LBL-01 all-labels pin):
// under NoLabels the builder drops ONLY edge label content. Nodes, clusters
// (wrapper, boundary, expanded, nested) and the legend keep their labels; the
// legend was already exempt (LBL-03) and URL attributes are structural.

// noLabelsModel builds a model with named/tech/description units, a collapsed
// expanded box, a referenced unit and a kind link — the edge label is the one
// thing --no-labels must silence.
func noLabelsModel() *parser.Model {
	return &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"app": {
				Type:        model.TypeSystem,
				Name:        "Application",
				Technology:  "Go",
				Description: "the main app",
				Reference:   "https://docs.example.com/app",
				Links: []model.Link{{
					Peer:        "db",
					Technology:  "SQL",
					Description: "reads rows",
					Kind:        model.KindRead,
				}},
			},
			"svc": {
				Type:     model.TypeBox,
				Name:     "Service Context",
				Expanded: []string{"svc"},
				Subunits: map[string]*model.Unit{
					"api": {Type: model.TypeContainer, Name: "API", Technology: "REST"},
				},
				SubunitOrder: []string{"api"},
			},
			"db": {Type: model.TypeDb, Name: "Database", Technology: "Postgres"},
		},
	}
}

// TestNoLabelsNodesKeepFullLabels: under --no-labels NODE labels survive with
// their full content (name, technology, description, glyphs) — the flag mutes
// edge label text only.
func TestNoLabelsNodesKeepFullLabels(t *testing.T) {
	t.Parallel()

	v := view.GenerateC1View(noLabelsModel())
	v.NoLabels = true
	g := graph.BuildGraph(v)

	app := nodeByID(t, g, "app")
	require.NotNil(t, app.Label, "node label content must SURVIVE NoLabels (edge labels only)")
	assert.Equal(t, "Application 📖", app.Label.Name, "name incl. glyph retained")
	assert.Equal(t, "Go", app.Label.Technology, "technology retained")
	assert.Equal(t, "the main app", app.Label.Description, "description retained")
	assert.Equal(t, graph.ShapeForType(model.TypeSystem), app.Shape,
		"the unit's plain default shape (ShapeForType) is retained")
	assert.Equal(t, "app", app.ID, "node ID retained")
	assert.Equal(t, "https://docs.example.com/app", app.ReferenceURL,
		"reference URL is structural and survives labels-off")

	// Cluster-resident leaf nodes keep their labels too.
	var api *graph.Node

	for _, c := range collectClusterTree(g.Clusters) {
		for _, n := range c.Nodes {
			if n.ID == "svc.api" {
				api = n
			}
		}
	}

	require.NotNil(t, api, "expanded child node present")
	require.NotNil(t, api.Label, "cluster child node label survives NoLabels")
	assert.Equal(t, "API", api.Label.Name)
	assert.Equal(t, graph.ShapeForType(model.TypeContainer), api.Shape)
}

func TestNoLabelsEdgesHaveNoLabel(t *testing.T) {
	t.Parallel()

	v := view.GenerateC1View(noLabelsModel())
	v.NoLabels = true
	g := graph.BuildGraph(v)

	require.Len(t, g.Edges, 1)
	assert.Nil(t, g.Edges[0].Label, "edge label must be nil under NoLabels")
	assert.Equal(t, model.LinkReadColour, g.Edges[0].Color,
		"colour is not a label — kind colour semantics untouched")
}

// TestNoLabelsClustersKeepLabels: under --no-labels every cluster kind —
// wrapper, boundary, expanded, nested — keeps its label; edge labels are the
// only suppression.
func TestNoLabelsClustersKeepLabels(t *testing.T) {
	t.Parallel()

	m := wrapTestMultilevelModel(t)

	v := view.GenerateC3View(m, "mainSystem.storages.localStorage")
	v.NoLabels = true
	g := graph.BuildGraph(v)

	clusters := collectClusterTree(g.Clusters)
	require.NotEmpty(t, clusters, "cluster structure retained")

	for _, c := range clusters {
		require.NotNil(t, c.Label,
			"cluster %s label content SURVIVES NoLabels (edge labels only)", c.ID)
		assert.NotEmpty(t, c.Label.Name, "cluster %s label has name text", c.ID)

		for _, n := range c.Nodes {
			require.NotNil(t, n.Label,
				"node %s label content SURVIVES NoLabels (edge labels only)", n.ID)
		}
	}

	// Structure — including the 38-01 wrapper clusters — survives.
	foundBoundary, foundWrapMain, foundWrapStorages := false, false, false

	for _, c := range clusters {
		switch c.ID {
		case "mainSystem.storages.localStorage":
			foundBoundary = true
		case "wrap_mainSystem":
			foundWrapMain = true
		case "wrap_mainSystem.storages":
			foundWrapStorages = true
		}
	}

	assert.True(t, foundBoundary, "expanded unit's boundary cluster retained")
	assert.True(t, foundWrapMain, "wrapper cluster retained (labelled)")
	assert.True(t, foundWrapStorages, "nested wrapper cluster retained")
}

func TestNoLabelsLegendStays(t *testing.T) {
	t.Parallel()

	v := view.GenerateC1View(noLabelsModel())
	v.NoLabels = true
	g := graph.BuildGraph(v)

	require.NotNil(t, g.Legend, "legend is metadata, not an element label — it STAYS (LBL-03)")
	assert.NotEmpty(t, g.Legend.Entries)
}

func TestNoLabelsCopiedFromView(t *testing.T) {
	t.Parallel()

	v := view.GenerateC1View(noLabelsModel())
	v.NoLabels = true

	assert.True(t, graph.BuildGraph(v).Opts.NoLabels, "BuildGraph copies View.NoLabels")
	assert.True(t, graph.BuildExpandedGraph(v).Opts.NoLabels,
		"BuildExpandedGraph copies View.NoLabels too (--no-labels x --expanded, LBL-02)")
}

func TestNoLabelsOptInDefaultPathUntouched(t *testing.T) {
	t.Parallel()

	v := view.GenerateC1View(noLabelsModel())
	g := graph.BuildGraph(v)

	assert.False(t, g.Opts.NoLabels, "default path leaves Opts.NoLabels false")

	app := nodeByID(t, g, "app")
	require.NotNil(t, app.Label)
	assert.Equal(t, "Application 📖", app.Label.Name, "default path keeps label content (incl. glyph)")
	require.NotNil(t, g.Edges[0].Label)
	assert.Equal(t, "reads rows", g.Edges[0].Label.Description)
}
