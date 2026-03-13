package render_test

import (
	"strings"
	"testing"

	"github.com/Djarvur/c4drill/internal/graph"
	"github.com/Djarvur/c4drill/internal/render"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Note: Tests in this file do NOT use t.Parallel() because the go-graphviz
// library uses a WASM-based rendering engine that has concurrency issues.

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestConverterCreatesNodes(t *testing.T) {
	t.Run("converter creates nodes with correct IDs from graph.Nodes", func(t *testing.T) {
		g := &graph.Graph{
			Title:     "Test Diagram",
			Direction: "TB",
			Nodes: []*graph.Node{
				{
					ID:    "node_one",
					Label: &graph.Label{Name: "Node One"},
					Shape: graph.ShapeRecord,
					Style: &graph.NodeStyle{},
				},
				{
					ID:    "node_two",
					Label: &graph.Label{Name: "Node Two"},
					Shape: graph.ShapeRecord,
					Style: &graph.NodeStyle{},
				},
			},
		}

		output, err := render.RenderDOT(g)
		require.NoError(t, err)
		assert.Contains(t, string(output), "node_one")
		assert.Contains(t, string(output), "node_two")
	})
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestConverterCreatesEdges(t *testing.T) {
	t.Run("converter creates edges connecting source to target", func(t *testing.T) {
		g := &graph.Graph{
			Title:     "Test Diagram",
			Direction: "TB",
			Nodes: []*graph.Node{
				{ID: "source_node", Label: &graph.Label{Name: "Source"}, Shape: graph.ShapeRecord},
				{ID: "target_node", Label: &graph.Label{Name: "Target"}, Shape: graph.ShapeRecord},
			},
			Edges: []*graph.Edge{
				{
					Source: "source_node",
					Target: "target_node",
					Label:  &graph.EdgeLabel{Technology: "HTTP"},
				},
			},
		}

		output, err := render.RenderDOT(g)
		require.NoError(t, err)
		// Verify edge exists in output
		dotStr := string(output)
		assert.Contains(t, dotStr, "source_node")
		assert.Contains(t, dotStr, "target_node")
	})
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestConverterCreatesClusters(t *testing.T) {
	t.Run("converter creates clusters with cluster_ prefix", func(t *testing.T) {
		g := &graph.Graph{
			Title:     "Test Diagram",
			Direction: "TB",
			Clusters: []*graph.Cluster{
				{
					ID:    "mycluster",
					Label: &graph.Label{Name: "My Cluster"},
					Nodes: []*graph.Node{
						{ID: "cluster_node", Label: &graph.Label{Name: "Cluster Node"}, Shape: graph.ShapeRecord},
					},
					Style: &graph.NodeStyle{},
				},
			},
		}

		output, err := render.RenderDOT(g)
		require.NoError(t, err)
		// Clusters must have "cluster_" prefix in GraphViz
		assert.Contains(t, string(output), "cluster_mycluster")
		assert.Contains(t, string(output), "cluster_node")
	})
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestConverterSetsGraphDirection(t *testing.T) {
	t.Run("converter sets graph direction TB by default", func(t *testing.T) {
		g := &graph.Graph{
			Title:     "Test Diagram",
			Direction: "TB",
			Nodes:     []*graph.Node{{ID: "test", Label: &graph.Label{Name: "Test"}, Shape: graph.ShapeRecord}},
		}

		output, err := render.RenderDOT(g)
		require.NoError(t, err)
		assert.Contains(t, string(output), "rankdir")
		assert.Contains(t, strings.ToLower(string(output)), "tb")
	})

	t.Run("converter sets graph direction LR when specified", func(t *testing.T) {
		g := &graph.Graph{
			Title:     "Test Diagram",
			Direction: "LR",
			Nodes:     []*graph.Node{{ID: "test", Label: &graph.Label{Name: "Test"}, Shape: graph.ShapeRecord}},
		}

		output, err := render.RenderDOT(g)
		require.NoError(t, err)
		assert.Contains(t, string(output), "rankdir")
		assert.Contains(t, strings.ToLower(string(output)), "lr")
	})
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestConverterSetsEdgeStyle(t *testing.T) {
	t.Run("converter sets edge routing style ortho", func(t *testing.T) {
		g := &graph.Graph{
			Title:     "Test Diagram",
			Direction: "TB",
			EdgeStyle: "ortho",
			Nodes: []*graph.Node{
				{ID: "a", Label: &graph.Label{Name: "A"}, Shape: graph.ShapeRecord},
				{ID: "b", Label: &graph.Label{Name: "B"}, Shape: graph.ShapeRecord},
			},
			Edges: []*graph.Edge{
				{Source: "a", Target: "b"},
			},
		}

		output, err := render.RenderDOT(g)
		require.NoError(t, err)
		assert.Contains(t, string(output), "splines")
	})
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestNodesInClusters(t *testing.T) {
	t.Run("nodes inside clusters have correct parent association", func(t *testing.T) {
		g := &graph.Graph{
			Title:     "Test Diagram",
			Direction: "TB",
			Clusters: []*graph.Cluster{
				{
					ID:    "system",
					Label: &graph.Label{Name: "System"},
					Nodes: []*graph.Node{
						{
							ID:          "internal_node",
							Label:       &graph.Label{Name: "Internal Node"},
							Shape:       graph.ShapeRecord,
							IsInCluster: true,
							Style:       &graph.NodeStyle{},
						},
					},
					Style: &graph.NodeStyle{},
				},
			},
		}

		output, err := render.RenderDOT(g)
		require.NoError(t, err)
		// Node should be inside cluster subgraph
		assert.Contains(t, string(output), "internal_node")
		assert.Contains(t, string(output), "cluster_system")
	})
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestEdgeStyles(t *testing.T) {
	t.Run("edge with dashed style", func(t *testing.T) {
		g := &graph.Graph{
			Title:     "Test Diagram",
			Direction: "TB",
			Nodes: []*graph.Node{
				{ID: "a", Label: &graph.Label{Name: "A"}, Shape: graph.ShapeRecord},
				{ID: "b", Label: &graph.Label{Name: "B"}, Shape: graph.ShapeRecord},
			},
			Edges: []*graph.Edge{
				{Source: "a", Target: "b", Style: "dashed"},
			},
		}

		output, err := render.RenderDOT(g)
		require.NoError(t, err)
		assert.Contains(t, string(output), "dashed")
	})

	t.Run("edge with dotted style", func(t *testing.T) {
		g := &graph.Graph{
			Title:     "Test Diagram",
			Direction: "TB",
			Nodes: []*graph.Node{
				{ID: "a", Label: &graph.Label{Name: "A"}, Shape: graph.ShapeRecord},
				{ID: "b", Label: &graph.Label{Name: "B"}, Shape: graph.ShapeRecord},
			},
			Edges: []*graph.Edge{
				{Source: "a", Target: "b", Style: "dotted"},
			},
		}

		output, err := render.RenderDOT(g)
		require.NoError(t, err)
		assert.Contains(t, string(output), "dotted")
	})
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestArrowDirections(t *testing.T) {
	t.Run("edge with reverse arrow", func(t *testing.T) {
		g := &graph.Graph{
			Title:     "Test Diagram",
			Direction: "TB",
			Nodes: []*graph.Node{
				{ID: "a", Label: &graph.Label{Name: "A"}, Shape: graph.ShapeRecord},
				{ID: "b", Label: &graph.Label{Name: "B"}, Shape: graph.ShapeRecord},
			},
			Edges: []*graph.Edge{
				{Source: "a", Target: "b", ArrowHead: graph.ArrowReverse},
			},
		}

		output, err := render.RenderDOT(g)
		require.NoError(t, err)
		assert.Contains(t, string(output), "dir")
	})

	t.Run("edge with both arrows", func(t *testing.T) {
		g := &graph.Graph{
			Title:     "Test Diagram",
			Direction: "TB",
			Nodes: []*graph.Node{
				{ID: "a", Label: &graph.Label{Name: "A"}, Shape: graph.ShapeRecord},
				{ID: "b", Label: &graph.Label{Name: "B"}, Shape: graph.ShapeRecord},
			},
			Edges: []*graph.Edge{
				{Source: "a", Target: "b", ArrowHead: graph.ArrowBoth},
			},
		}

		output, err := render.RenderDOT(g)
		require.NoError(t, err)
		assert.Contains(t, string(output), "dir")
	})

	t.Run("edge with no arrow", func(t *testing.T) {
		g := &graph.Graph{
			Title:     "Test Diagram",
			Direction: "TB",
			Nodes: []*graph.Node{
				{ID: "a", Label: &graph.Label{Name: "A"}, Shape: graph.ShapeRecord},
				{ID: "b", Label: &graph.Label{Name: "B"}, Shape: graph.ShapeRecord},
			},
			Edges: []*graph.Edge{
				{Source: "a", Target: "b", ArrowHead: graph.ArrowNone},
			},
		}

		output, err := render.RenderDOT(g)
		require.NoError(t, err)
		assert.Contains(t, string(output), "dir")
	})
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestExternalNodes(t *testing.T) {
	t.Run("external nodes get dashed style", func(t *testing.T) {
		g := &graph.Graph{
			Title:     "Test Diagram",
			Direction: "TB",
			Nodes: []*graph.Node{
				{
					ID:         "external_system",
					Label:      &graph.Label{Name: "External System"},
					Shape:      graph.ShapeRecord,
					IsExternal: true,
					Style:      &graph.NodeStyle{},
				},
			},
		}

		output, err := render.RenderDOT(g)
		require.NoError(t, err)
		assert.Contains(t, string(output), "external_system")
	})
}

//nolint:paralleltest,funlen // go-graphviz WASM engine has concurrency issues
func TestCreateNode_WithExploreURL(t *testing.T) {
	t.Run("node with ExploreURL gets URL attribute in DOT output", func(t *testing.T) {
		g := &graph.Graph{
			Title:     "Test Diagram",
			Direction: "TB",
			Nodes: []*graph.Node{
				{
					ID:         "expandable_system",
					Label:      &graph.Label{Name: "Expandable System"},
					Shape:      graph.ShapeRecord,
					ExploreURL: "./expandable_system.svg",
					Style:      &graph.NodeStyle{FillColor: "#438DD5"},
				},
			},
		}

		output, err := render.RenderDOT(g)
		require.NoError(t, err)
		// URL attribute should be present in DOT output
		assert.Contains(t, string(output), "URL")
		assert.Contains(t, string(output), "./expandable_system.svg")
	})

	t.Run("node without ExploreURL does not get URL attribute", func(t *testing.T) {
		g := &graph.Graph{
			Title:     "Test Diagram",
			Direction: "TB",
			Nodes: []*graph.Node{
				{
					ID:    "simple_node",
					Label: &graph.Label{Name: "Simple Node"},
					Shape: graph.ShapeRecord,
					Style: &graph.NodeStyle{},
				},
			},
		}

		output, err := render.RenderDOT(g)
		require.NoError(t, err)
		// Simple node should exist but not have URL pointing to svg
		assert.Contains(t, string(output), "simple_node")
		assert.NotContains(t, string(output), ".svg")
	})

	t.Run("node with empty ExploreURL does not get URL attribute", func(t *testing.T) {
		g := &graph.Graph{
			Title:     "Test Diagram",
			Direction: "TB",
			Nodes: []*graph.Node{
				{
					ID:         "empty_url_node",
					Label:      &graph.Label{Name: "Empty URL Node"},
					Shape:      graph.ShapeRecord,
					ExploreURL: "",
					Style:      &graph.NodeStyle{FillColor: "#438DD5"},
				},
			},
		}

		output, err := render.RenderDOT(g)
		require.NoError(t, err)
		assert.Contains(t, string(output), "empty_url_node")
		// Empty URL should not produce a URL attribute
		assert.NotContains(t, string(output), `URL=""`)
	})
}
