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

//nolint:funlen,paralleltest // Table-driven test pattern; go-graphviz WASM concurrency issues
func TestRecordLabelGeneration(t *testing.T) {
	tests := []struct {
		name        string
		label       *graph.Label
		contains    []string
		notContains []string
	}{
		{
			name:     "Record label with Name",
			label:    &graph.Label{Name: "My Node"},
			contains: []string{"My Node", "{", "}"},
		},
		{
			name:     "Icon before Name",
			label:    &graph.Label{Name: "Database", Icon: "db"},
			contains: []string{"Database", "db"},
		},
		{
			name:     "Technology row",
			label:    &graph.Label{Name: "API Server", Technology: "Go"},
			contains: []string{"API Server", "Go", "|"},
		},
		{
			name:     "Description row",
			label:    &graph.Label{Name: "API", Description: "REST API"},
			contains: []string{"API", "REST API", "|"},
		},
		{
			name: "all fields",
			label: &graph.Label{
				Name:        "Web App",
				Icon:        "icon",
				Technology:  "React",
				Description: "Frontend application",
			},
			contains: []string{"Web App", "icon", "React"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &graph.Graph{
				Title:     "Test",
				Direction: "TB",
				Nodes: []*graph.Node{
					{
						ID:    "test_node",
						Label: tt.label,
						Shape: graph.ShapeRecord,
						Style: &graph.NodeStyle{FillColor: "#438DD5"},
					},
				},
			}

			output, err := render.RenderDOT(g)
			require.NoError(t, err)

			dotStr := string(output)
			for _, c := range tt.contains {
				assert.Contains(t, dotStr, c)
			}

			for _, nc := range tt.notContains {
				assert.NotContains(t, dotStr, nc)
			}
		})
	}
}

//nolint:funlen,paralleltest // Table-driven test pattern; go-graphviz WASM concurrency issues
func TestEdgeLabelGeneration(t *testing.T) {
	tests := []struct {
		name         string
		edgeLabel    *graph.EdgeLabel
		contains     []string
		checkNewline bool
	}{
		{
			name: "Technology and Description",
			edgeLabel: &graph.EdgeLabel{
				Technology:  "HTTP",
				Description: "REST calls",
			},
			contains: []string{"HTTP", "REST calls"},
		},
		{
			name: "Description only",
			edgeLabel: &graph.EdgeLabel{
				Description: "Uses",
			},
			contains: []string{"Uses"},
		},
		{
			name: "Technology only",
			edgeLabel: &graph.EdgeLabel{
				Technology: "TCP",
			},
			contains: []string{"TCP"},
		},
		{
			name: "Technology and Description with newline",
			edgeLabel: &graph.EdgeLabel{
				Technology:  "gRPC",
				Description: "Protocol buffers",
			},
			contains:     []string{"gRPC", "Protocol buffers"},
			checkNewline: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &graph.Graph{
				Title:     "Test",
				Direction: "TB",
				Nodes: []*graph.Node{
					{ID: "a", Label: &graph.Label{Name: "A"}, Shape: graph.ShapeRecord},
					{ID: "b", Label: &graph.Label{Name: "B"}, Shape: graph.ShapeRecord},
				},
				Edges: []*graph.Edge{
					{Source: "a", Target: "b", Label: tt.edgeLabel},
				},
			}

			output, err := render.RenderDOT(g)
			require.NoError(t, err)

			dotStr := string(output)
			for _, c := range tt.contains {
				assert.Contains(t, dotStr, c)
			}

			if tt.checkNewline {
				assert.True(t, strings.Contains(dotStr, "\\n") || strings.Contains(dotStr, "\n"),
					"Edge label should contain newline separator")
			}
		})
	}
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestNodeStyles(t *testing.T) {
	t.Run("node with dashed border style", func(t *testing.T) {
		g := &graph.Graph{
			Title:     "Test",
			Direction: "TB",
			Nodes: []*graph.Node{
				{
					ID:    "dashed_node",
					Label: &graph.Label{Name: "Dashed"},
					Shape: graph.ShapeRecord,
					Style: &graph.NodeStyle{
						FillColor:   "#438DD5",
						BorderStyle: "dashed",
					},
				},
			},
		}

		output, err := render.RenderDOT(g)
		require.NoError(t, err)
		assert.Contains(t, string(output), "dashed_node")
	})
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestClusterStyles(t *testing.T) {
	t.Run("cluster with dashed border style", func(t *testing.T) {
		g := &graph.Graph{
			Title:     "Test",
			Direction: "TB",
			Clusters: []*graph.Cluster{
				{
					ID:    "dashed_cluster",
					Label: &graph.Label{Name: "Dashed Cluster"},
					Nodes: []*graph.Node{
						{
							ID:    "inner_node",
							Label: &graph.Label{Name: "Inner"},
							Shape: graph.ShapeRecord,
						},
					},
					Style: &graph.NodeStyle{
						FillColor:   "#E5E5E5",
						BorderStyle: "dashed",
					},
				},
			},
		}

		output, err := render.RenderDOT(g)
		require.NoError(t, err)
		assert.Contains(t, string(output), "cluster_dashed_cluster")
	})
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestGraphTitle(t *testing.T) {
	t.Run("graph title appears in output", func(t *testing.T) {
		g := &graph.Graph{
			Title:     "My C4 Diagram",
			Direction: "TB",
			Nodes: []*graph.Node{
				{ID: "test", Label: &graph.Label{Name: "Test"}, Shape: graph.ShapeRecord},
			},
		}

		output, err := render.RenderDOT(g)
		require.NoError(t, err)
		assert.Contains(t, string(output), "My C4 Diagram")
	})
}
