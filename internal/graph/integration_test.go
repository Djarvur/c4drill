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

// TestIntegrationBuildGraphFromC1View tests that BuildGraph from a C1 view
// produces correct nodes and edges.
func TestIntegrationBuildGraphFromC1View(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{
			Name:  "Test System",
			Edges: "spline",
		},
		Units: map[string]*model.Unit{
			"app": {
				Type:        model.TypeSystem,
				Name:        "App",
				Description: "Main app",
				Technology:  "Go",
				Links: []model.Link{
					{
						Peer:        "db",
						Technology:  "SQL",
						Description: "Queries",
					},
				},
			},
			"db": {
				Type:        model.TypeDb,
				Name:        "Database",
				Description: "Data store",
				Technology:  "PostgreSQL",
			},
			"user": {
				Type:        model.TypePerson,
				Name:        "User",
				Description: "End user",
			},
		},
	}

	// Test 1: BuildGraph from C1 view produces correct nodes and edges
	v := view.GenerateC1View(m)
	require.NotNil(t, v)

	g := graph.BuildGraph(v)
	require.NotNil(t, g)

	// Verify basic graph properties
	assert.Equal(t, "Test System", g.Title)
	assert.Equal(t, "TB", g.Direction)
	assert.Equal(t, "spline", g.EdgeStyle)

	// Should have 3 nodes (app, db, user)
	assert.Len(t, g.Nodes, 3)

	// Verify nodes have correct shapes based on type
	nodeMap := make(map[string]*graph.Node)
	for _, node := range g.Nodes {
		nodeMap[node.ID] = node
	}

	assert.Equal(t, graph.ShapeRecord, nodeMap["app"].Shape)
	assert.Equal(t, graph.ShapeRecord, nodeMap["db"].Shape)
	assert.Equal(t, graph.ShapeRecord, nodeMap["user"].Shape)

	// Verify icons
	assert.Empty(t, nodeMap["app"].Label.Icon)                // System has no icon
	assert.Equal(t, "\u26C1", nodeMap["db"].Label.Icon)       // DB has cylinder icon
	assert.Equal(t, "\U0001F464", nodeMap["user"].Label.Icon) // Person has person icon

	// Should have 1 edge (app -> db)
	assert.Len(t, g.Edges, 1)
	assert.Equal(t, "app", g.Edges[0].Source)
	assert.Equal(t, "db", g.Edges[0].Target)
}

// TestIntegrationBuildGraphFromC2ViewWithClusters tests that BuildGraph from C2 view
// produces correct clusters for expanded systems.
func TestIntegrationBuildGraphFromC2ViewWithClusters(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"mainsystem": {
				Type:     model.TypeSystem,
				Name:     "Main System",
				Expanded: []string{"mainsystem"}, // Expanded
				Edges:    "square",
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
				},
			},
		},
	}

	// Test 2: BuildGraph from C2 view produces correct clusters
	v := view.GenerateC1View(m) // C1 view with expanded system
	require.NotNil(t, v)

	g := graph.BuildGraph(v)
	require.NotNil(t, g)

	// Should have 1 cluster for the expanded system
	require.Len(t, g.Clusters, 1)

	cluster := g.Clusters[0]
	assert.Equal(t, "mainsystem", cluster.ID)
	assert.Equal(t, "Main System", cluster.Label.Name)

	// Cluster should contain 2 child nodes
	assert.Len(t, cluster.Nodes, 2)

	// Verify child nodes are marked as in cluster
	for _, node := range cluster.Nodes {
		assert.True(t, node.IsInCluster)
	}
}

// TestIntegrationBuildGraphExternalBoundaryNodes tests that BuildGraph with external
// boundary nodes produces correct styling.
func TestIntegrationBuildGraphExternalBoundaryNodes(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"internal": {
				Type:        model.TypeSystem,
				Name:        "Internal System",
				Description: "Our system",
				Links: []model.Link{
					{
						Peer:        "externalapi",
						Technology:  "HTTP",
						Description: "External call",
					},
				},
			},
		},
	}

	// Test 3: BuildGraph with external boundary nodes produces correct styling
	v := view.GenerateC1View(m)
	require.NotNil(t, v)

	g := graph.BuildGraph(v)
	require.NotNil(t, g)

	// Find the external boundary node
	var externalNode *graph.Node

	for _, node := range g.Nodes {
		if node.ID == "externalapi" {
			externalNode = node

			break
		}
	}

	require.NotNil(t, externalNode, "External boundary node should exist")
	assert.True(t, externalNode.IsExternal)

	// External nodes should have solid border
	require.NotNil(t, externalNode.Style)
	assert.Equal(t, "solid", externalNode.Style.BorderStyle)

	// External nodes should have transparent fill (no background color)
	assert.Empty(t, externalNode.Style.FillColor)
}

// TestIntegrationFullPipelineModelToGraph tests the full pipeline:
// model -> view -> graph produces valid structure.
func TestIntegrationFullPipelineModelToGraph(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{
			Name:  "Full Pipeline Test",
			Edges: "straight",
		},
		Units: map[string]*model.Unit{
			"system": {
				Type:        model.TypeSystem,
				Name:        "System",
				Description: "Main system",
				Technology:  "Go",
				Links: []model.Link{
					{
						Peer:        "db",
						Technology:  "SQL",
						Description: "Queries",
					},
				},
			},
			"db": {
				Type:        model.TypeDb,
				Name:        "Database",
				Description: "Data store",
				Technology:  "PostgreSQL",
			},
		},
	}

	// Test 4: Full pipeline: model -> view -> graph produces valid structure
	v := view.GenerateC1View(m)
	require.NotNil(t, v)
	assert.Equal(t, view.LevelC1, v.Level)

	g := graph.BuildGraph(v)
	require.NotNil(t, g)

	// Verify the graph structure
	assert.Equal(t, "Full Pipeline Test", g.Title)
	assert.Equal(t, "straight", g.EdgeStyle)
	assert.Len(t, g.Nodes, 2)
	require.Len(t, g.Edges, 1)

	// Verify edge has correct properties
	edge := g.Edges[0]
	assert.Equal(t, "system", edge.Source)
	assert.Equal(t, "db", edge.Target)
	assert.Equal(t, "SQL", edge.Label.Technology)
	assert.Equal(t, "Queries", edge.Label.Description)
	assert.Equal(t, "solid", edge.Style) // Default style
}

// TestIntegrationMultipleLinksBetweenSameUnits tests that multiple links
// between the same units produce separate edges.
func TestIntegrationMultipleLinksBetweenSameUnits(t *testing.T) {
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
						Description: "Reads",
					},
				},
				LinksFrom: []model.Link{
					{
						Peer:        "db",
						Technology:  "Callback",
						Description: "Notifications",
					},
				},
			},
			"db": {
				Type: model.TypeDb,
				Name: "Database",
			},
		},
	}

	// Test 5: Multiple links between same units produce separate edges
	v := view.GenerateC1View(m)
	require.NotNil(t, v)

	g := graph.BuildGraph(v)
	require.NotNil(t, g)

	// Should have 2 edges: app->db and db->app
	require.Len(t, g.Edges, 2)

	// Verify edges have different directions
	var appToDB, dbToApp bool

	for _, edge := range g.Edges {
		if edge.Source == "app" && edge.Target == "db" {
			appToDB = true

			assert.Equal(t, "SQL", edge.Label.Technology)
		}

		if edge.Source == "db" && edge.Target == "app" {
			dbToApp = true

			assert.Equal(t, "Callback", edge.Label.Technology)
		}
	}

	assert.True(t, appToDB, "Should have app->db edge")
	assert.True(t, dbToApp, "Should have db->app edge")
}

// TestIntegrationCollapsedIndicatorOnNodesWithSubunits tests that the 🔍 indicator
// appears on nodes with subunits that are not expanded.
func TestIntegrationCollapsedIndicatorOnNodesWithSubunits(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"collapsed": {
				Type: model.TypeSystem,
				Name: "Collapsed System",
				// Not expanded, has subunits
				Subunits: map[string]*model.Unit{
					"container": {
						Type: model.TypeContainer,
						Name: "Container",
					},
				},
			},
			"expanded": {
				Type:     model.TypeSystem,
				Name:     "Expanded System",
				Expanded: []string{"expanded"}, // Expanded
				Subunits: map[string]*model.Unit{
					"container": {
						Type: model.TypeContainer,
						Name: "Container",
					},
				},
			},
			"simple": {
				Type: model.TypeSystem,
				Name: "Simple System",
				// No subunits
			},
		},
	}

	// Test 6: 🔍 indicator appears on nodes with subunits
	v := view.GenerateC1View(m)
	require.NotNil(t, v)

	g := graph.BuildGraph(v)
	require.NotNil(t, g)

	nodeMap := make(map[string]*graph.Node)
	for _, node := range g.Nodes {
		nodeMap[node.ID] = node
	}

	// Collapsed system with subunits should have 🔍
	assert.Contains(t, nodeMap["collapsed"].Label.Name, "🔍", "Collapsed system should have 🔍 indicator")

	// Expanded system becomes a cluster, not a node (so check cluster label)
	// The expanded system itself becomes a cluster
	var expandedCluster *graph.Cluster

	for _, cluster := range g.Clusters {
		if cluster.ID == "expanded" {
			expandedCluster = cluster

			break
		}
	}

	require.NotNil(t, expandedCluster, "Expanded system should be a cluster")
	assert.NotContains(t, expandedCluster.Label.Name, "🔍", "Expanded cluster label should not have 🔍")

	// Simple system without subunits should not have 🔍
	assert.NotContains(t, nodeMap["simple"].Label.Name, "🔍", "Simple system should not have 🔍 indicator")
}

// TestIntegrationGraphWithAllUnitTypes tests that the graph correctly handles
// all unit types with appropriate shapes and icons.
func TestIntegrationGraphWithAllUnitTypes(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "All Types Test"},
		Units: map[string]*model.Unit{
			"person": {
				Type: model.TypePerson,
				Name: "Person",
			},
			"personExt": {
				Type: model.TypePersonExternal,
				Name: "External Person",
			},
			"system": {
				Type: model.TypeSystem,
				Name: "System",
			},
			"systemExt": {
				Type: model.TypeSystemExternal,
				Name: "External System",
			},
			"db": {
				Type: model.TypeDb,
				Name: "Database",
			},
			"dbExt": {
				Type: model.TypeDbExternal,
				Name: "External Database",
			},
			"queue": {
				Type: model.TypeQueue,
				Name: "Queue",
			},
			"queueExt": {
				Type: model.TypeQueueExternal,
				Name: "External Queue",
			},
		},
	}

	v := view.GenerateC1View(m)
	require.NotNil(t, v)

	g := graph.BuildGraph(v)
	require.NotNil(t, g)

	nodeMap := make(map[string]*graph.Node)
	for _, node := range g.Nodes {
		nodeMap[node.ID] = node
	}

	// Verify icons for each type
	assert.Equal(t, "\U0001F464", nodeMap["person"].Label.Icon)       // Person icon
	assert.Equal(t, "\U0001F464", nodeMap["personExt"].Label.Icon)    // Person icon
	assert.Empty(t, nodeMap["system"].Label.Icon)                     // No icon
	assert.Empty(t, nodeMap["systemExt"].Label.Icon)                  // No icon
	assert.Equal(t, "\u26C1", nodeMap["db"].Label.Icon)               // DB icon
	assert.Equal(t, "\u26C1", nodeMap["dbExt"].Label.Icon)            // DB icon
	assert.Equal(t, "\u255F\n\u2562", nodeMap["queue"].Label.Icon)    // Queue bars
	assert.Equal(t, "\u255F\n\u2562", nodeMap["queueExt"].Label.Icon) // Queue bars

	// Verify external nodes have solid border and external colors
	for _, id := range []string{"personExt", "systemExt", "dbExt", "queueExt"} {
		assert.True(t, nodeMap[id].IsExternal, "%s should be external", id)
		assert.Equal(t, "solid", nodeMap[id].Style.BorderStyle, "%s should have solid border", id)
	}
}

// ============================================================================
// Navigation Integration Tests (05-03)
// ============================================================================

// TestIntegration_Navigation_C1NoBackLink tests that C1 view has no navigation (root level).
func TestIntegration_Navigation_C1NoBackLink(t *testing.T) {
	t.Parallel()

	// C1 view should have no navigation (root level)
	m := buildTestModelWithExpandableSystem()
	v := view.GenerateC1View(m)
	g := graph.BuildGraphWithPath(v, "", "test", "svg")

	require.NotNil(t, g)
	assert.Nil(t, g.Navigation, "C1 should have no navigation")
}

// TestIntegration_Navigation_C2BackLink tests that C2 view has back-link to C1.
func TestIntegration_Navigation_C2BackLink(t *testing.T) {
	t.Parallel()

	// C2 view's breadcrumb root ancestor carries the up-navigation to C1.
	// (The standalone BackLink field was dropped — breadcrumb-only nav.)
	m := buildTestModelWithExpandableSystem()
	v := view.GenerateC2View(m, "mainsystem")
	g := graph.BuildGraphWithPath(v, "mainsystem", "test", "svg")

	require.NotNil(t, g)
	require.NotNil(t, g.Navigation, "C2 should have navigation")
	require.NotEmpty(t, g.Navigation.Breadcrumbs, "C2 should have breadcrumbs")
	// The root ancestor (first breadcrumb) points back to the C1 diagram.
	assert.NotEmpty(t, g.Navigation.Breadcrumbs[0].URL,
		"C2 breadcrumb root ancestor must link to C1")
	// BackLink is intentionally nil now.
	assert.Nil(t, g.Navigation.BackLink)
}

// TestIntegration_Navigation_C3Breadcrumbs tests that C3 view has breadcrumbs showing path.
func TestIntegration_Navigation_C3Breadcrumbs(t *testing.T) {
	t.Parallel()

	// C3 view should have breadcrumbs showing path
	m := buildTestModelWithNestedStructure()
	v := view.GenerateC3View(m, "mainsystem.api")
	g := graph.BuildGraphWithPath(v, "mainsystem.api", "test", "svg")

	require.NotNil(t, g)
	require.NotNil(t, g.Navigation, "C3 should have navigation")
	assert.Greater(t, len(g.Navigation.Breadcrumbs), 1, "C3 should have multiple breadcrumb items")

	// Last item should have no URL (current level)
	lastIdx := len(g.Navigation.Breadcrumbs) - 1
	assert.Empty(t, g.Navigation.Breadcrumbs[lastIdx].URL, "Current level should have no URL")
}

// TestIntegration_ExploreURL_CollapsedSystem tests collapsed system with subunits has explore URL.
func TestIntegration_ExploreURL_CollapsedSystem(t *testing.T) {
	t.Parallel()

	// Collapsed system with subunits should have explore URL
	m := buildTestModelWithExpandableSystem()
	v := view.GenerateC1View(m)
	g := graph.BuildGraphWithPath(v, "", "test", "svg")

	// Find the collapsed system node
	var collapsedNode *graph.Node

	for _, node := range g.Nodes {
		if node.ID == "mainsystem" {
			collapsedNode = node

			break
		}
	}

	require.NotNil(t, collapsedNode, "Should find mainsystem node")
	assert.NotEmpty(t, collapsedNode.ExploreURL, "Collapsed system should have explore URL")
	assert.Contains(t, collapsedNode.ExploreURL, ".svg")
}

// TestIntegration_ExploreURL_ExpandedSystem tests that fully expanded system becomes a cluster.
func TestIntegration_ExploreURL_ExpandedSystem(t *testing.T) {
	t.Parallel()

	// When a system's children are expanded, it still appears as a node in C1
	// but with its children shown inside (as a cluster)
	m := buildTestModelPreExpanded()
	v := view.GenerateC1View(m)
	g := graph.BuildGraphWithPath(v, "", "test", "svg")

	// The system should be a cluster (expanded), not a standalone node
	var mainsystemInNodes bool

	for _, node := range g.Nodes {
		if node.ID == "mainsystem" {
			mainsystemInNodes = true

			break
		}
	}

	// mainsystem should be in clusters (expanded), not in top-level nodes
	assert.False(t, mainsystemInNodes, "Fully expanded system should be in clusters, not nodes")

	// Verify it's in clusters
	var mainsystemCluster *graph.Cluster

	for _, cluster := range g.Clusters {
		if cluster.ID == "mainsystem" {
			mainsystemCluster = cluster

			break
		}
	}

	require.NotNil(t, mainsystemCluster, "Expanded system should be a cluster")
}

// TestIntegration_ExploreURL_NonExpandableTypes tests person/db/queue never have explore URLs.
func TestIntegration_ExploreURL_NonExpandableTypes(t *testing.T) {
	t.Parallel()

	// Person, db, queue should never have explore URLs
	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"user":  {Type: model.TypePerson, Name: "User"},
			"db":    {Type: model.TypeDb, Name: "Database"},
			"queue": {Type: model.TypeQueue, Name: "Queue"},
		},
	}
	v := view.GenerateC1View(m)
	g := graph.BuildGraphWithPath(v, "", "test", "svg")

	for _, node := range g.Nodes {
		assert.Empty(t, node.ExploreURL, "Non-expandable types should not have explore URL: %s", node.ID)
	}
}

// TestIntegration_Navigation_BackLinkName tests the breadcrumb root ancestor
// uses the parent (root) name. (Previously asserted on the BackLink field,
// which was dropped in favor of breadcrumb-only nav.)
func TestIntegration_Navigation_BackLinkName(t *testing.T) {
	t.Parallel()

	m := buildTestModelWithExpandableSystem()
	v := view.GenerateC2View(m, "mainsystem")
	g := graph.BuildGraphWithPath(v, "mainsystem", "test", "svg")

	require.NotNil(t, g)
	require.NotNil(t, g.Navigation)
	require.NotEmpty(t, g.Navigation.Breadcrumbs, "C2 should have breadcrumbs")
	// The root ancestor name should be derived from the root context title.
	assert.NotEmpty(t, g.Navigation.Breadcrumbs[0].Name,
		"C2 breadcrumb root ancestor must carry a name")
}

// TestIntegration_Navigation_BreadcrumbAncestorsClickable tests breadcrumb ancestors are clickable.
func TestIntegration_Navigation_BreadcrumbAncestorsClickable(t *testing.T) {
	t.Parallel()

	m := buildTestModelWithNestedStructure()
	v := view.GenerateC3View(m, "mainsystem.api")
	g := graph.BuildGraphWithPath(v, "mainsystem.api", "test", "svg")

	require.NotNil(t, g)
	require.NotNil(t, g.Navigation)
	require.Greater(t, len(g.Navigation.Breadcrumbs), 1, "C3 should have multiple breadcrumbs")

	// All items except the last should have URLs (be clickable)
	for i, item := range g.Navigation.Breadcrumbs {
		if i < len(g.Navigation.Breadcrumbs)-1 {
			assert.NotEmpty(t, item.URL, "Breadcrumb ancestor %d should have URL", i)
		}
	}
}

// ============================================================================
// Helper functions for navigation tests
// ============================================================================

// buildTestModelWithExpandableSystem creates a model with an expandable system.
func buildTestModelWithExpandableSystem() *parser.Model {
	return &parser.Model{
		Properties: model.Properties{Name: "Test System"},
		Units: map[string]*model.Unit{
			"mainsystem": {
				Type: model.TypeSystem,
				Name: "Main System",
				Subunits: map[string]*model.Unit{
					"api": {Type: model.TypeSystem, Name: "API"},
					"web": {Type: model.TypeSystem, Name: "Web App"},
				},
			},
		},
	}
}

// buildTestModelWithNestedStructure creates a model with nested structure for C3 tests.
func buildTestModelWithNestedStructure() *parser.Model {
	return &parser.Model{
		Properties: model.Properties{Name: "Test System"},
		Units: map[string]*model.Unit{
			"mainsystem": {
				Type:     model.TypeSystem,
				Name:     "Main System",
				Expanded: []string{"api"},
				Subunits: map[string]*model.Unit{
					"api": {
						Type: model.TypeSystem,
						Name: "API Container",
						Subunits: map[string]*model.Unit{
							"auth":  {Type: model.TypeSystem, Name: "Auth Service"},
							"users": {Type: model.TypeSystem, Name: "User Service"},
						},
					},
					"web": {Type: model.TypeSystem, Name: "Web App"},
				},
			},
		},
	}
}

// buildTestModelPreExpanded creates a model with pre-expanded system.
func buildTestModelPreExpanded() *parser.Model {
	m := buildTestModelWithExpandableSystem()
	// To expand the mainsystem itself, add "mainsystem" to its Expanded list
	m.Units["mainsystem"].Expanded = []string{"mainsystem"}

	return m
}
