package render_test

import (
	"strings"
	"testing"

	"github.com/Djarvur/c4drill/internal/graph"
	"github.com/Djarvur/c4drill/internal/render"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// pngMagic is the first eight bytes of every PNG file (PNG signature).
const pngMagic = "\x89PNG\r\n\x1a\n"

// navGraph builds a graph carrying one drill-capable node, one external
// reference node, breadcrumb navigation, and a nested drill-capable cluster —
// the full navigation surface the PNG HTML wrapper must re-emit as anchors.
func navGraph() *graph.Graph {
	return &graph.Graph{
		Title:     "Nav Graph",
		Direction: "TB",
		Navigation: &graph.Navigation{
			Breadcrumbs: []graph.BreadcrumbItem{
				{Name: "Root", URL: "../model.svg"},
				{Name: "Current"},
			},
		},
		Nodes: []*graph.Node{
			{
				ID:         "child",
				Label:      &graph.Label{Name: "Child System 🔍"},
				Shape:      graph.ShapeHTML,
				ExploreURL: "sub/child.svg",
			},
			{
				ID:           "leaf",
				Label:        &graph.Label{Name: "Leaf 📖"},
				Shape:        graph.ShapeHTML,
				ReferenceURL: "https://example.com/docs",
			},
		},
		Clusters: []*graph.Cluster{
			{
				ID:         "wrapper",
				Label:      &graph.Label{Name: "Wrapper"},
				ExploreURL: "sub/wrapper.svg",
				Clusters: []*graph.Cluster{
					{
						ID:         "nested",
						Label:      &graph.Label{Name: "Nested 🔍"},
						ExploreURL: "sub/wrapper/nested.svg",
					},
				},
			},
		},
	}
}

// TestRenderPNG proves the PNG raster contract (issue #26): non-empty bytes
// carrying the PNG magic number — real raster, not a misrouted SVG/DOT
// payload — plus nil-graph and dispatch parity with RenderSVG's pattern.
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestRenderPNG(t *testing.T) {
	t.Run("returns bytes starting with the PNG magic number", func(t *testing.T) {
		g := testGraph("test_node")

		output, err := render.RenderPNG(g)
		require.NoError(t, err, "RenderPNG should not return error for valid graph")
		require.NotEmpty(t, output, "RenderPNG should return non-empty bytes")
		require.GreaterOrEqual(t, len(output), len(pngMagic), "PNG output shorter than the signature")
		assert.Equal(t, pngMagic, string(output[:len(pngMagic)]),
			"RenderPNG output must start with the 8-byte PNG signature")
		assert.NotContains(t, string(output), "<svg",
			"PNG output must not be an SVG payload misrouted to .png")
	})

	t.Run("returns error for nil graph input", func(t *testing.T) {
		output, err := render.RenderPNG(nil)
		require.Error(t, err, "RenderPNG should return error for nil graph")
		assert.Nil(t, output, "RenderPNG should return nil bytes for nil graph")
	})

	t.Run("png format dispatch returns same result as RenderPNG", func(t *testing.T) {
		g := testGraph("test_node")

		renderOutput, err := render.Render(g, "png")
		require.NoError(t, err, "Render with png format should not return error")

		pngOutput, err := render.RenderPNG(g)
		require.NoError(t, err, "RenderPNG should not return error")

		assert.Equal(t, pngOutput, renderOutput,
			"Render(g, 'png') should return same result as RenderPNG(g)")
	})
}

// TestRenderHTMLForPNG proves the HTML navigation layer over the PNG raster:
// the <img> embed, breadcrumb anchors with .svg→.html rewritten hrefs,
// drill-down links (one per drill-capable unit, clusters included), and
// external references as new-tab anchors.
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestRenderHTMLForPNG(t *testing.T) {
	t.Run("produces an HTML doc embedding the sibling PNG", func(t *testing.T) {
		g := testGraph("test_node")

		output, err := render.RenderHTMLForPNG(g, "model.png")
		require.NoError(t, err)

		s := string(output)
		assert.True(t, strings.HasPrefix(s, "<!DOCTYPE html>"),
			"wrapper must start with <!DOCTYPE html>")
		assert.Contains(t, s, `<img src="model.png"`,
			"wrapper must embed its sibling PNG by reference")
		assert.Contains(t, s, `alt="Test Diagram"`, "img alt must carry the diagram title")
		assert.Contains(t, s, `</html>`, "wrapper must close the html element")
		assert.NotContains(t, s, "<?xml", "wrapper is HTML — no XML declaration")
		assert.NotContains(t, s, "<svg", "wrapper embeds a raster, not an inline SVG")
	})

	t.Run("renders breadcrumb anchors with .svg rewritten to .html", func(t *testing.T) {
		g := navGraph()

		output, err := render.RenderHTMLForPNG(g, "model.png")
		require.NoError(t, err)

		s := string(output)
		assert.Contains(t, s, `<a href="../model.html">Root</a>`,
			"breadcrumb ancestor must anchor to its .html page (rewritten from .svg)")
		assert.Contains(t, s, `<span class="current">Current</span>`,
			"the current breadcrumb level must be plain text")
	})

	t.Run("falls back to a title-only breadcrumb without navigation", func(t *testing.T) {
		g := testGraph("test_node") // no Navigation (C1 shape)

		output, err := render.RenderHTMLForPNG(g, "model.png")
		require.NoError(t, err)

		assert.Contains(t, string(output), `<span class="current">Test Diagram</span>`,
			"C1 docs must show the title as the current breadcrumb level")
	})

	t.Run("lists one drill-down anchor per drill-capable unit including clusters", func(t *testing.T) {
		output, err := render.RenderHTMLForPNG(navGraph(), "model.png")
		require.NoError(t, err)

		s := string(output)
		assert.Contains(t, s, `<a href="sub/child.html">Child System 🔍</a>`,
			"drill-capable node must link to its child doc page (.html rewritten)")
		assert.Contains(t, s, `<a href="sub/wrapper.html">Wrapper</a>`,
			"drill-capable cluster must appear in the drill list")
		assert.Contains(t, s, `<a href="sub/wrapper/nested.html">Nested 🔍</a>`,
			"nested clusters must be walked (CTX-03 parity)")
	})

	t.Run("opens external references in a new tab", func(t *testing.T) {
		output, err := render.RenderHTMLForPNG(navGraph(), "model.png")
		require.NoError(t, err)

		s := string(output)
		assert.Contains(t, s, `<a href="https://example.com/docs" target="_blank" rel="noopener">Leaf 📖</a>`,
			"external reference must be a new-tab anchor")
	})

	t.Run("renders non-http reference schemes as inert text", func(t *testing.T) {
		g := testGraph("evil")
		g.Nodes[0].ReferenceURL = "javascript:alert(1)"

		output, err := render.RenderHTMLForPNG(g, "model.png")
		require.NoError(t, err)

		s := string(output)
		assert.NotContains(t, s, "javascript:",
			"untrusted scheme must never reach an href (T-28-02 parity: the static doc has no click shim)")
		assert.Contains(t, s, "Test Node", "the entry stays listed as inert text")
	})

	t.Run("omits the link section when the graph has no targets", func(t *testing.T) {
		g := testGraph("lonely")

		output, err := render.RenderHTMLForPNG(g, "model.png")
		require.NoError(t, err)

		assert.NotContains(t, string(output), `class="links"`,
			"image-only docs must not emit an empty link section")
	})

	t.Run("returns error for nil graph", func(t *testing.T) {
		output, err := render.RenderHTMLForPNG(nil, "model.png")
		require.Error(t, err)
		assert.Nil(t, output)
	})

	t.Run("returns error for empty png name", func(t *testing.T) {
		output, err := render.RenderHTMLForPNG(testGraph("test_node"), "")
		require.Error(t, err, "an empty img src would produce a broken doc")
		assert.Nil(t, output)
	})
}
