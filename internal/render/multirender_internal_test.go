package render

import (
	"fmt"
	"testing"

	"github.com/Djarvur/c4drill/internal/graph"
	"github.com/Djarvur/c4drill/internal/model"
)

// TestMultipleSequentialSVGRenders tests whether multiple SVG renders
// in sequence cause WASM memory issues.
//
//nolint:paralleltest
func TestMultipleSequentialSVGRenders(t *testing.T) {
	tmpDir := t.TempDir()

	for i := range 5 {
		g := &graph.Graph{
			Direction: "TB",
			EdgeStyle: "spline",
			Title:     fmt.Sprintf("Graph %d", i),
			Nodes: []*graph.Node{
				{
					ID:   "a",
					Type: model.TypeSystem,
					Label: &graph.Label{
						Name:        "System A",
						Description: "desc",
						Technology:  "Go",
					},
					Style: &graph.NodeStyle{
						BorderColor: "#3C7FC0",
						FillColor:   "#438DD5",
						FontColor:   "#FFF",
					},
				},
				{
					ID:   "b",
					Type: model.TypeContainer,
					Label: &graph.Label{
						Name:        "Container B",
						Description: "desc",
						Technology:  "Java",
					},
					Style: &graph.NodeStyle{
						BorderColor: "#78A8D8",
						FillColor:   "#85BBF0",
						FontColor:   "#FFF",
					},
				},
			},
			Edges: []*graph.Edge{
				{
					Source: "a",
					Target: "b",
					Label:  &graph.EdgeLabel{Description: "calls"},
				},
			},
		}

		svg, err := RenderSVGWithOutput(g, tmpDir)
		if err != nil {
			t.Fatalf("Render %d failed: %v", i, err)
		}

		t.Logf("Render %d OK: %d bytes", i, len(svg))
	}
}
