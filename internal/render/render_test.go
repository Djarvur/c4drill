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

// testGraph creates a simple test graph with one node.
func testGraph(nodeID string) *graph.Graph {
	return &graph.Graph{
		Title:     "Test Diagram",
		Direction: "TB",
		Nodes: []*graph.Node{
			{
				ID:    nodeID,
				Label: &graph.Label{Name: "Test Node"},
				Shape: graph.ShapeHTML,
				Style: &graph.NodeStyle{
					FillColor:   "#438DD5",
					BorderColor: "#3C7FC0",
					FontColor:   "#FFFFFF",
				},
			},
		},
	}
}

// Note: Tests in this file do NOT use t.Parallel() because the go-graphviz
// library uses a WASM-based rendering engine that has concurrency issues
// with parallel map writes. See: https://github.com/goccy/go-graphviz/issues

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestRenderDOT(t *testing.T) {
	t.Run("returns valid DOT bytes for a simple graph with one node", func(t *testing.T) {
		g := testGraph("test_node")

		output, err := render.RenderDOT(g)
		require.NoError(t, err, "RenderDOT should not return error for valid graph")
		assert.NotEmpty(t, output, "RenderDOT should return non-empty bytes")
		assert.True(t, strings.Contains(string(output), "digraph") ||
			strings.Contains(string(output), "graph"),
			"DOT output should contain graph/digraph keyword")
	})

	t.Run("returns error for nil graph input", func(t *testing.T) {
		output, err := render.RenderDOT(nil)
		require.Error(t, err, "RenderDOT should return error for nil graph")
		assert.Nil(t, output, "RenderDOT should return nil bytes for nil graph")
	})

	t.Run("rendered DOT contains expected node ID in output", func(t *testing.T) {
		g := testGraph("my_special_node")

		output, err := render.RenderDOT(g)
		require.NoError(t, err)
		assert.Contains(t, string(output), "my_special_node",
			"DOT output should contain the node ID")
	})
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestRenderSVG(t *testing.T) {
	t.Run("returns valid SVG bytes for a simple graph with one node", func(t *testing.T) {
		g := testGraph("test_node")

		output, err := render.RenderSVG(g)
		require.NoError(t, err, "RenderSVG should not return error for valid graph")
		assert.NotEmpty(t, output, "RenderSVG should return non-empty bytes")
		assert.Contains(t, string(output), "<?xml", "SVG output should contain XML declaration")
		assert.Contains(t, string(output), "<svg", "SVG output should contain svg element")
	})

	t.Run("returns error for nil graph input", func(t *testing.T) {
		output, err := render.RenderSVG(nil)
		require.Error(t, err, "RenderSVG should return error for nil graph")
		assert.Nil(t, output, "RenderSVG should return nil bytes for nil graph")
	})
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestRenderFormatDispatch(t *testing.T) {
	t.Run("dot format returns same result as RenderDOT", func(t *testing.T) {
		g := testGraph("test_node")

		renderOutput, err := render.Render(g, "dot")
		require.NoError(t, err, "Render with dot format should not return error")

		dotOutput, err := render.RenderDOT(g)
		require.NoError(t, err, "RenderDOT should not return error")

		assert.Equal(t, dotOutput, renderOutput,
			"Render(g, 'dot') should return same result as RenderDOT(g)")
	})

	t.Run("svg format returns same result as RenderSVG", func(t *testing.T) {
		g := testGraph("test_node")

		renderOutput, err := render.Render(g, "svg")
		require.NoError(t, err, "Render with svg format should not return error")

		svgOutput, err := render.RenderSVG(g)
		require.NoError(t, err, "RenderSVG should not return error")

		assert.Equal(t, svgOutput, renderOutput,
			"Render(g, 'svg') should return same result as RenderSVG(g)")
	})
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestRenderErrors(t *testing.T) {
	t.Run("invalid format returns error with clear message", func(t *testing.T) {
		g := testGraph("test_node")

		output, err := render.Render(g, "invalid")
		require.Error(t, err, "Render with invalid format should return error")
		assert.Nil(t, output, "Render with invalid format should return nil bytes")
		require.ErrorContains(t, err, "unsupported format",
			"Error message should mention unsupported format")
	})

	t.Run("returns error for nil graph", func(t *testing.T) {
		output, err := render.Render(nil, "dot")
		require.Error(t, err, "Render should return error for nil graph")
		assert.Nil(t, output, "Render should return nil bytes for nil graph")
	})
}
