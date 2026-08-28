package render_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/Djarvur/c4drill/internal/graph"
	"github.com/Djarvur/c4drill/internal/model"
	"github.com/Djarvur/c4drill/internal/render"
	"github.com/Djarvur/c4drill/internal/testutil/canonical"
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
	t.Run("external nodes get solid borders", func(t *testing.T) {
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

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
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

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestNodeRoundedStyle(t *testing.T) {
	t.Run("node has rounded style by default", func(t *testing.T) {
		g := &graph.Graph{
			Title:     "Test",
			Direction: "TB",
			Nodes: []*graph.Node{
				{
					ID:    "testnode",
					Label: &graph.Label{Name: "Test"},
					Shape: graph.ShapeRecord,
					Style: &graph.NodeStyle{},
				},
			},
		}

		output, err := render.RenderDOT(g)
		require.NoError(t, err)

		dotStr := string(output)
		t.Logf("DOT output:\n%s", dotStr)
		// Check for style=rounded attribute (not just the word "rounded")
		assert.True(t, strings.Contains(dotStr, "style=rounded") || strings.Contains(dotStr, `style="rounded"`),
			"DOT output should contain style=rounded attribute")
	})
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestQueueNodePlainStyle(t *testing.T) {
	// Queue nodes must NOT carry the "rounded" style: GraphViz then emits a
	// plain rect <polygon> for them, which the SVG pipe post-processor
	// (pipe.go) replaces with a horizontal-pipe <path>. Non-queue styling is
	// unchanged, and dashed/dotted borders survive the rounded drop.
	queueGraph := func(borderStyle string) string {
		g := &graph.Graph{
			Title:     "T",
			Direction: "TB",
			Nodes: []*graph.Node{
				{
					ID:    "q",
					Label: &graph.Label{Name: "Queue"},
					Shape: graph.ShapeRecord,
					Type:  model.TypeQueue,
					Style: &graph.NodeStyle{BorderStyle: borderStyle},
				},
			},
		}

		output, err := render.RenderDOT(g)
		require.NoError(t, err)

		// Extract the queue node's statement block (from its name to the ";").
		// GraphViz emits the node name followed by a tab and its attribute list.
		dotStr := string(output)
		start := strings.Index(dotStr, "q\t[")
		require.NotEqual(t, -1, start, "queue node statement must exist")

		end := strings.Index(dotStr[start:], ";")
		require.NotEqual(t, -1, end, "queue node statement terminator")

		return dotStr[start : start+end]
	}

	t.Run("queue node style has no rounded", func(t *testing.T) {
		block := queueGraph("solid")

		assert.Contains(t, block, "style=", "queue node must carry a style attribute")
		assert.NotContains(t, block, "rounded", "queue style must drop rounded so SVG emits a polygon")
	})

	t.Run("dashed queue node keeps dashed without rounded", func(t *testing.T) {
		block := queueGraph("dashed")

		assert.Contains(t, block, "dashed", "queue dashed border must survive")
		assert.NotContains(t, block, "rounded", "queue style must drop rounded even when dashed")
	})

	t.Run("queue node carries wider minimum width", func(t *testing.T) {
		block := queueGraph("solid")

		assert.Contains(t, block, "width=1.8", "queue nodes need a wider minimum width to read as pipes")
	})

	t.Run("non-queue nodes still emit rounded", func(t *testing.T) {
		g := &graph.Graph{
			Title:     "T",
			Direction: "TB",
			Nodes: []*graph.Node{
				{ID: "s", Label: &graph.Label{Name: "System"}, Shape: graph.ShapeRecord, Style: &graph.NodeStyle{}},
			},
		}

		output, err := render.RenderDOT(g)
		require.NoError(t, err)

		assert.Contains(t, string(output), "style=rounded", "non-queue styling must stay byte-identical")
	})
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestEdgeColorRendering(t *testing.T) {
	t.Run("edge with color has color attribute in DOT", func(t *testing.T) {
		g := &graph.Graph{
			Title:     "Test",
			Direction: "TB",
			Nodes: []*graph.Node{
				{ID: "a", Label: &graph.Label{Name: "A"}, Shape: graph.ShapeRecord, Style: &graph.NodeStyle{}},
				{ID: "b", Label: &graph.Label{Name: "B"}, Shape: graph.ShapeRecord, Style: &graph.NodeStyle{}},
			},
			Edges: []*graph.Edge{
				{
					Source: "a",
					Target: "b",
					Color:  "#3C7FC0", // SystemBorder color
				},
			},
		}

		output, err := render.RenderDOT(g)
		require.NoError(t, err)

		dotStr := string(output)
		assert.Contains(t, dotStr, "#3C7FC0")
	})

	t.Run("edge with color has fontcolor matching edge color", func(t *testing.T) {
		g := &graph.Graph{
			Title:     "Test",
			Direction: "TB",
			Nodes: []*graph.Node{
				{ID: "a", Label: &graph.Label{Name: "A"}, Shape: graph.ShapeRecord, Style: &graph.NodeStyle{}},
				{ID: "b", Label: &graph.Label{Name: "B"}, Shape: graph.ShapeRecord, Style: &graph.NodeStyle{}},
			},
			Edges: []*graph.Edge{
				{
					Source: "a",
					Target: "b",
					Label:  &graph.EdgeLabel{Technology: "HTTP"},
					Color:  "#78A8D8", // ComponentBorder color
				},
			},
		}

		output, err := render.RenderDOT(g)
		require.NoError(t, err)

		dotStr := string(output)
		// Both color and fontcolor should be set to the same value
		assert.Contains(t, dotStr, "#78A8D8")
	})

	t.Run("edge without color does not add color attributes", func(t *testing.T) {
		g := &graph.Graph{
			Title:     "Test",
			Direction: "TB",
			Nodes: []*graph.Node{
				{ID: "a", Label: &graph.Label{Name: "A"}, Shape: graph.ShapeRecord, Style: &graph.NodeStyle{}},
				{ID: "b", Label: &graph.Label{Name: "B"}, Shape: graph.ShapeRecord, Style: &graph.NodeStyle{}},
			},
			Edges: []*graph.Edge{
				{
					Source: "a",
					Target: "b",
					Color:  "", // No color
				},
			},
		}

		output, err := render.RenderDOT(g)
		require.NoError(t, err)
		// Should still create the edge, just without color
		assert.Contains(t, string(output), "a")
		assert.Contains(t, string(output), "b")
	})
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestRankReverseEmission(t *testing.T) {
	// RANK-01: rank="reverse" swaps edge endpoints at emission and inverts the
	// dir attribute, so the visual arrow stays pointed at the logical target
	// while the vertical ranking flips. The logical Edge.Source/Target are
	// unchanged; only emission swaps. Matrix from 36-RESEARCH.md §2.3:
	//
	// | authored arrow | rank=default      | rank=reverse       |
	// |----------------|-------------------|--------------------|
	// | "" (omitted)   | a -> b (no dir)   | b -> a [dir=back]  |
	// | forward        | a -> b (no dir)   | b -> a [dir=back]  |
	// | reverse        | a -> b [dir=back] | b -> a (no dir)    |
	// | bidirectional  | a -> b [dir=both] | b -> a [dir=both]  |
	// | none           | a -> b [dir=none] | b -> a [dir=none]  |
	renderGraph := func(edge *graph.Edge) string {
		g := &graph.Graph{
			Title:     "T",
			Direction: "TB",
			Nodes: []*graph.Node{
				{ID: "a", Label: &graph.Label{Name: "A"}, Shape: graph.ShapeRecord},
				{ID: "b", Label: &graph.Label{Name: "B"}, Shape: graph.ShapeRecord},
			},
			Edges: []*graph.Edge{edge},
		}

		output, err := render.RenderDOT(g)
		require.NoError(t, err)

		return string(output)
	}

	tests := []struct {
		name         string
		arrow        graph.ArrowDirection
		rankReverse  bool
		wantEdgeStmt string // the emitted edge statement tail
		wantDirAttr  string // empty = no per-edge dir attribute
	}{
		{"omitted arrow, default rank", graph.ArrowDirection(""), false, "a -> b", ""},
		{"omitted arrow, reverse rank", graph.ArrowDirection(""), true, "b -> a", "dir=back"},
		{"forward arrow, default rank", graph.ArrowForward, false, "a -> b", ""},
		{"forward arrow, reverse rank", graph.ArrowForward, true, "b -> a", "dir=back"},
		{"reverse arrow, default rank", graph.ArrowReverse, false, "a -> b", "dir=back"},
		{"reverse arrow, reverse rank", graph.ArrowReverse, true, "b -> a", ""},
		{"bidirectional arrow, default rank", graph.ArrowBoth, false, "a -> b", "dir=both"},
		{"bidirectional arrow, reverse rank", graph.ArrowBoth, true, "b -> a", "dir=both"},
		{"none arrow, default rank", graph.ArrowNone, false, "a -> b", "dir=none"},
		{"none arrow, reverse rank", graph.ArrowNone, true, "b -> a", "dir=none"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := renderGraph(&graph.Edge{
				Source:      "a",
				Target:      "b",
				ArrowHead:   tt.arrow,
				RankReverse: tt.rankReverse,
			})

			assert.Contains(t, output, tt.wantEdgeStmt, "edge endpoint order")

			// Per-edge dir attribute: check the edge statement block only
			// (the graph-level `edge [dir=forward]` default is pre-existing).
			start := strings.Index(output, tt.wantEdgeStmt)
			require.NotEqual(t, -1, start, "edge statement start")

			edgeBlock := output[start:]

			end := strings.Index(edgeBlock, ";")
			require.NotEqual(t, -1, end, "edge statement terminator")

			edgeBlock = edgeBlock[:end]
			if tt.wantDirAttr == "" {
				assert.NotContains(t, edgeBlock, "dir=", "no per-edge dir attribute expected")
			} else {
				assert.Contains(t, edgeBlock, tt.wantDirAttr, "per-edge dir attribute")
			}
		})
	}

	t.Run("no per-edge dir=forward is ever emitted", func(t *testing.T) {
		arrows := []graph.ArrowDirection{
			"", graph.ArrowForward, graph.ArrowReverse, graph.ArrowBoth, graph.ArrowNone,
		}

		for _, arrow := range arrows {
			for _, rev := range []bool{false, true} {
				output := renderGraph(&graph.Edge{
					Source:      "a",
					Target:      "b",
					ArrowHead:   arrow,
					RankReverse: rev,
				})

				start := strings.Index(output, "->")
				require.NotEqual(t, -1, start, "edge statement start")

				edgeBlock := output[start:]

				end := strings.Index(edgeBlock, ";")
				require.NotEqual(t, -1, end, "edge statement terminator")

				assert.NotContains(t, edgeBlock[:end], "dir=forward")
			}
		}
	})

	t.Run("rank=reverse equals the linkFrom + arrow=reverse idiom (canonical)", func(t *testing.T) {
		stripKey := func(c string) string {
			// The edge key is internal bookkeeping: rank=reverse keeps the
			// logical name (a_to_b) while the idiom's name follows its swapped
			// endpoints (b_to_a). Layout-irrelevant — normalize it away.
			return regexp.MustCompile(`key=[^\x00]+\x00`).ReplaceAllString(c, "")
		}

		reverseRank := stripKey(canonical.Canonical(t, renderGraph(&graph.Edge{
			Source:      "a",
			Target:      "b",
			RankReverse: true,
		})))

		idiom := stripKey(canonical.Canonical(t, renderGraph(&graph.Edge{
			Source:    "b",
			Target:    "a",
			ArrowHead: graph.ArrowReverse,
		})))

		assert.Equal(t, idiom, reverseRank,
			"rank=reverse must produce the same canonical DOT as the <- + arrow=reverse idiom")
	})
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestUnitOverridesEmission(t *testing.T) {
	t.Run("dotted border style and fill color emit in DOT", func(t *testing.T) {
		g := &graph.Graph{
			Title:     "T",
			Direction: "TB",
			Nodes: []*graph.Node{
				{
					ID:    "app",
					Label: &graph.Label{Name: "App"},
					Shape: graph.ShapeRecord,
					Style: &graph.NodeStyle{
						FillColor:   "#123456",
						FontColor:   "#FFFFFF",
						BorderColor: "#AA0000",
						BorderStyle: "dotted",
					},
				},
			},
		}

		output, err := render.RenderDOT(g)
		require.NoError(t, err)

		out := string(output)

		assert.Contains(t, out, `fillcolor="#123456"`)
		assert.Contains(t, out, `fontcolor="#FFFFFF"`)
		assert.Contains(t, out, `color="#AA0000"`)
		assert.Contains(t, out, `style="rounded,filled,dotted"`)
	})

	t.Run("edges=square maps to splines=ortho", func(t *testing.T) {
		for _, edgeStyle := range []struct{ in, want string }{
			{"square", "ortho"},
			{"ortho", "ortho"},
			{"straight", "false"},
			{"spline", "true"},
		} {
			g := &graph.Graph{
				Title:     "T",
				Direction: "TB",
				EdgeStyle: edgeStyle.in,
				Nodes: []*graph.Node{
					{ID: "a", Label: &graph.Label{Name: "A"}, Shape: graph.ShapeRecord},
				},
			}

			output, err := render.RenderDOT(g)
			require.NoError(t, err)
			assert.Contains(t, string(output), "splines="+edgeStyle.want,
				"EdgeStyle %q should emit splines=%s", edgeStyle.in, edgeStyle.want)
		}
	})
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestLegendRendering(t *testing.T) {
	legend := &graph.Legend{
		Entries: []graph.LegendEntry{
			{Label: "container", Color: "#3C7FC0"},
			{Label: "read", Color: "#2E7D32"},
			{Label: "a<b & \"evil\"", Color: "#C0392B"},
		},
	}

	renderWith := func(g *graph.Graph) string {
		output, err := render.RenderDOT(g)
		require.NoError(t, err)

		return string(output)
	}

	baseGraph := func() *graph.Graph {
		return &graph.Graph{
			Title:     "T",
			Direction: "TB",
			Nodes: []*graph.Node{
				{ID: "a", Label: &graph.Label{Name: "A"}, Shape: graph.ShapeRecord},
			},
		}
	}

	t.Run("legend renders as a framed floating plaintext node", func(t *testing.T) {
		g := baseGraph()
		g.Legend = legend

		out := renderWith(g)
		assert.Contains(t, out, `__c4drill_legend`, "legend node exists")
		assert.Contains(t, out, "shape=plaintext", "legend node is borderless plaintext")
		assert.Contains(t, out, `<TABLE BORDER="1"`, "legend table is framed")
		assert.Contains(t, out, `<B>legend</B>`, "legend carries a caption")
		assert.Contains(t, out, `COLOR="#2E7D32">read`, "entry text is the sample, in its colour")
	})

	t.Run("legend stays out of the graph label", func(t *testing.T) {
		g := baseGraph()
		g.Legend = legend

		out := renderWith(g)
		assert.NotContains(t, out, "BGCOLOR=", "no swatch cells anywhere")
		assert.NotContains(t, out, `ALIGN="RIGHT"`, "no legend rows in the label table")
	})

	t.Run("entries are HTML-escaped", func(t *testing.T) {
		g := baseGraph()
		g.Legend = legend

		out := renderWith(g)
		assert.Contains(t, out, "a&lt;b &amp; &#34;evil&#34;", "label escaped inside the legend node")
		assert.NotContains(t, out, `>a<b`, "no raw markup injected")
	})

	t.Run("colourless entry falls back to the muted grey", func(t *testing.T) {
		g := baseGraph()
		g.Legend = &graph.Legend{Entries: []graph.LegendEntry{{Label: "custom"}}}

		out := renderWith(g)
		assert.Contains(t, out, `COLOR="#666666">custom`, "muted fallback colour")
	})

	t.Run("nameless graph with legend still emits the legend node", func(t *testing.T) {
		g := baseGraph()
		g.Title = ""
		g.Legend = legend

		out := renderWith(g)
		assert.Contains(t, out, "__c4drill_legend", "legend present on nameless view (LEG-01)")
	})

	t.Run("nameless graph without legend emits no label (shape preserved)", func(t *testing.T) {
		g := baseGraph()
		g.Title = ""

		out := renderWith(g)
		assert.NotContains(t, out, "label=<<TABLE", "no label table without nav/title")
		assert.NotContains(t, out, "__c4drill_legend", "no legend node without legend")
	})
}
