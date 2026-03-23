package render_test

import (
	"testing"

	"github.com/Djarvur/c4drill/internal/graph"
	"github.com/Djarvur/c4drill/internal/model"
	"github.com/Djarvur/c4drill/internal/render"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIconExtractorIntegration tests the integration of icon extraction with rendering.
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestIconExtractorIntegration(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a simple graph with one node that should have an icon
	g := &graph.Graph{
		Direction: "TB",
		EdgeStyle: "spline",
		Title:     "Test Graph",
		Nodes: []*graph.Node{
			{
				ID:          "test_node",
				Type:        model.TypeComponent,
				IsInCluster: false,
				Label: &graph.Label{
					Name:        "Test Component",
					Technology:  "Go",
					Description: "A test component",
				},
				Style: &graph.NodeStyle{
					BorderColor: "#78A8D8",
					FontColor:   "#78A8D8",
				},
			},
		},
	}

	// Render with output directory - icons should be base64 embedded
	output, err := render.RenderWithOutput(g, "dot", tmpDir)
	require.NoError(t, err)

	// Verify the DOT output contains base64 data URI for the icon
	outputStr := string(output)
	assert.Contains(t, outputStr, `<img src="`, "DOT should contain img tag")
	assert.Contains(t, outputStr, `data:image/svg+xml;base64,`, "DOT should contain SVG base64 data URI")
}

// TestRenderDOTWithoutOutputDir tests that RenderDOT (without output dir) does NOT extract icons.
func TestRenderDOTWithoutOutputDir(t *testing.T) {
	t.Parallel()

	// Create a simple graph with one node
	g := &graph.Graph{
		Direction: "TB",
		EdgeStyle: "spline",
		Title:     "Test Graph",
		Nodes: []*graph.Node{
			{
				ID:          "test_node",
				Type:        model.TypeComponent,
				IsInCluster: false,
				Label: &graph.Label{
					Name:        "Test Component",
					Technology:  "Go",
					Description: "A test component",
				},
				Style: &graph.NodeStyle{
					BorderColor: "#78A8D8",
					FontColor:   "#78A8D8",
				},
			},
		},
	}

	// Render WITHOUT output directory
	output, err := render.RenderDOT(g)
	require.NoError(t, err)

	// Verify the DOT output does NOT contain img tags (because no icon extraction)
	outputStr := string(output)
	assert.NotContains(t, outputStr, `<img src="`, "Without outputDir, DOT should NOT contain img tags")
}

// TestRenderSVGWithOutputDir tests that SVG output includes icon references.
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestRenderSVGWithOutputDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a simple graph with one node
	g := &graph.Graph{
		Direction: "TB",
		EdgeStyle: "spline",
		Title:     "Test Graph",
		Nodes: []*graph.Node{
			{
				ID:          "test_person",
				Type:        model.TypePerson,
				IsInCluster: false,
				Label: &graph.Label{
					Name:        "Test User",
					Description: "A test person",
				},
				Style: &graph.NodeStyle{
					BorderColor: "#8A8A8A",
					FontColor:   "#8A8A8A",
				},
			},
		},
	}

	// Render SVG with output directory
	svg, err := render.RenderSVGWithOutput(g, tmpDir)
	require.NoError(t, err)

	svgStr := string(svg)

	// SVG should contain base64-encoded SVG icons as image elements
	assert.Contains(t, svgStr, `<image href="data:image/svg+xml;base64,`,
		"SVG should contain base64-encoded SVG icons in image elements")
}
