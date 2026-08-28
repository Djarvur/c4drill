package render_test

import (
	"strings"
	"testing"

	"github.com/Djarvur/c4drill/internal/graph"
	"github.com/Djarvur/c4drill/internal/model"
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

	t.Run("html format returns same result as RenderHTML", func(t *testing.T) {
		g := testGraph("test_node")

		renderOutput, err := render.Render(g, "html")
		require.NoError(t, err, "Render with html format should not return error")

		htmlOutput, err := render.RenderHTML(g)
		require.NoError(t, err, "RenderHTML should not return error")

		assert.Equal(t, htmlOutput, renderOutput,
			"Render(g, 'html') should return same result as RenderHTML(g)")
	})
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestRenderHTML(t *testing.T) {
	t.Run("produces a self-contained HTML document", func(t *testing.T) {
		g := testGraph("test_node")

		output, err := render.RenderHTML(g)
		require.NoError(t, err, "RenderHTML should not error for valid graph")
		require.NotEmpty(t, output, "RenderHTML should return non-empty bytes")

		s := string(output)
		assert.True(t, strings.HasPrefix(s, "<!DOCTYPE html>"),
			"HTML output should start with <!DOCTYPE html>")
		assert.Contains(t, s, "<svg", "HTML output should contain the inlined SVG")
		assert.Contains(t, s, "</html>", "HTML output should close the <html> tag")
		assert.NotContains(t, s, "<?xml",
			"XML declaration must be stripped (invalid inside HTML body)")
	})

	t.Run("injects the Safari nav shim", func(t *testing.T) {
		g := testGraph("test_node")

		output, err := render.RenderHTML(g)
		require.NoError(t, err)

		s := string(output)
		// The shim is what makes SVG <a> links clickable in Safari/WebKit.
		assert.Contains(t, s, "window.location.href",
			"HTML output must contain the nav shim that navigates via window.location")
		assert.Contains(t, s, "querySelectorAll",
			"HTML output must attach listeners to svg a elements")
	})

	t.Run("rewrites .svg hrefs to .html", func(t *testing.T) {
		// A graph with an ExploreURL that the SVG renderer turns into an
		// xlink:href="X.svg" anchor. The HTML wrapper must rewrite to .html.
		g := &graph.Graph{
			Title:     "With Link",
			Direction: "TB",
			Nodes: []*graph.Node{
				{
					ID:    "expandable_system",
					Label: &graph.Label{Name: "Expandable"},
					Shape: graph.ShapeHTML,
					Style: &graph.NodeStyle{
						FillColor:   "#438DD5",
						BorderColor: "#3C7FC0",
						FontColor:   "#FFFFFF",
					},
					ExploreURL: "expandable_system.svg",
				},
			},
		}

		// Sanity: the raw SVG output contains the .svg href
		svgBytes, err := render.RenderSVG(g)
		require.NoError(t, err)
		require.Contains(t, string(svgBytes), `expandable_system.svg`,
			"precondition: raw SVG must contain the .svg href")

		htmlBytes, err := render.RenderHTML(g)
		require.NoError(t, err)

		s := string(htmlBytes)
		assert.Contains(t, s, `expandable_system.html`,
			"HTML output must rewrite .svg href to .html")
		assert.NotContains(t, s, `expandable_system.svg"`,
			"HTML output must not retain the .svg href suffix")
	})

	t.Run("returns error for nil graph", func(t *testing.T) {
		output, err := render.RenderHTML(nil)
		require.Error(t, err, "RenderHTML with nil graph should error")
		assert.Nil(t, output, "RenderHTML with nil graph should return nil bytes")
	})
}

// TestReferenceNavShim exercises REF-04: the htmlNavShim's click handler must
// route EXTERNAL http(s)/protocol-relative URLs distinctly from internal
// drill-down navigation so Safari/WebKit follows external references. It also
// hardens against non-http(s) schemes (T-28-02): javascript:/data: URIs in a
// reference URL must NOT navigate.
func TestReferenceNavShim(t *testing.T) {
	t.Parallel()

	g := testGraph("ref_node")

	output, err := render.RenderHTML(g)
	require.NoError(t, err)

	s := string(output)
	// REF-04: external references open in a new tab via window.open.
	assert.Contains(t, s, "window.open",
		"nav shim must route external reference URLs via window.open(_, _blank)")
	// REF-04: the scheme check must be present so external http(s)// URLs are
	// detected.
	assert.Contains(t, s, "http",
		"nav shim must check the http(s)// scheme to branch external vs internal")
	// Internal drill-down still navigates the same tab.
	assert.Contains(t, s, "window.location.href",
		"nav shim must keep window.location.href for internal drill-down navigation")
	// T-28-02: a generic scheme detector MUST appear so javascript:/data:/
	// vbscript: URIs are no-ops, not fall-through to window.location.href
	// (which would execute javascript: URIs). The regex anchors on a leading
	// scheme prefix and is the gate that prevents XSS via the reference URL.
	assert.Contains(t, s, "[a-z][a-z0-9",
		"nav shim must carry a generic scheme detector so javascript:/data: are no-ops (T-28-02)")
}

// TestQueuePipeIntegration proves the end-to-end queue rendering contract:
// a queue node's SVG group carries the horizontal-pipe <path> (arc commands,
// no polygon), non-queue nodes keep their original shape elements, and the
// HTML wrapper inherits the same pipe because it funnels through
// render(g, graphviz.SVG).
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestQueuePipeIntegration(t *testing.T) {
	queueAndFriends := func() *graph.Graph {
		return &graph.Graph{
			Title:     "Queue Pipe Integration",
			Direction: "TB",
			Nodes: []*graph.Node{
				{ID: "sys", Label: &graph.Label{Name: "System"}, Shape: graph.ShapeRecord, Type: model.TypeSystem},
				{ID: "store", Label: &graph.Label{Name: "Store"}, Shape: graph.ShapeRecord, Type: model.TypeDb},
				{
					ID:    "events",
					Label: &graph.Label{Name: "Events"},
					Shape: graph.ShapeRecord,
					Type:  model.TypeQueue,
					Style: &graph.NodeStyle{BorderColor: "#073B6F", FontColor: "#073B6F", BorderStyle: "solid"},
				},
			},
		}
	}

	// svgGroupAfterTitle returns the node group body from its <title> up to
	// (excluding) the group's closing tag.
	svgGroupAfterTitle := func(t *testing.T, svg, id string) string {
		t.Helper()

		needle := "<title>" + id + "</title>"
		start := strings.Index(svg, needle)
		require.NotEqual(t, -1, start, "node %s must be present in the SVG", id)

		end := strings.Index(svg[start:], "</g>")
		require.NotEqual(t, -1, end, "node %s group must close", id)

		return svg[start : start+end]
	}

	// firstPathD extracts the first d="..." attribute value from a group.
	firstPathD := func(group string) (string, bool) {
		marker := `d="`
		start := strings.Index(group, marker)

		if start < 0 {
			return "", false
		}

		rest := group[start+len(marker):]
		end := strings.Index(rest, `"`)

		if end < 0 {
			return "", false
		}

		return rest[:end], true
	}

	t.Run("queue node renders a pipe path in SVG", func(t *testing.T) {
		svg, err := render.RenderSVG(queueAndFriends())
		require.NoError(t, err)

		queueGroup := svgGroupAfterTitle(t, string(svg), "events")
		assert.Contains(t, queueGroup, `<path d="M`, "queue node must render the pipe path")
		assert.Contains(t, queueGroup, " A", "pipe path must carry arc commands")
		assert.NotContains(t, queueGroup, "<polygon", "queue node must not keep its polygon")
	})

	t.Run("non-queue nodes keep their native shapes", func(t *testing.T) {
		svg, err := render.RenderSVG(queueAndFriends())
		require.NoError(t, err)

		// Non-queue nodes keep the pre-existing rounded-box path (bezier
		// corners, no arc commands) or polygon outlines — anything but the
		// post-processor's arc-based pipe.
		for _, id := range []string{"sys", "store"} {
			group := svgGroupAfterTitle(t, string(svg), id)

			assert.True(t, strings.Contains(group, "<polygon") || strings.Contains(group, "<path"),
				"%s keeps a native shape element", id)

			if d, ok := firstPathD(group); ok {
				assert.NotContains(t, d, " A",
					"%s path must not carry the pipe's arc commands", id)
			}
		}
	})

	t.Run("HTML output inherits the pipe path", func(t *testing.T) {
		html, err := render.RenderHTML(queueAndFriends())
		require.NoError(t, err)

		queueGroup := svgGroupAfterTitle(t, string(html), "events")
		assert.Contains(t, queueGroup, `<path d="M`, "HTML-wrapped SVG must carry the pipe path")
		assert.Contains(t, queueGroup, " A", "HTML pipe path must carry arc commands")
		assert.NotContains(t, queueGroup, "<polygon", "HTML queue node must not keep its polygon")
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
