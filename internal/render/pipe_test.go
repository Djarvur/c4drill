package render

import (
	"strings"
	"testing"

	"github.com/Djarvur/c4drill/internal/graph"
	"github.com/Djarvur/c4drill/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// queuePipeSVGFixture mirrors a GraphViz 15.1.1 SVG fragment: node groups with
// a <title>, a shape element, then text. The queue node carries a plain rect
// <polygon> (queue style dropped "rounded"); the dashed variant shows the
// stroke-dasharray attribute GraphViz emits on the polygon.
const queuePipeSVGFixture = `<?xml version="1.0" encoding="UTF-8" standalone="no"?>
<g id="graph0" class="graph">
<g id="node1" class="node">
<title>platform.queue</title>
<polygon fill="none" stroke="#073b6f" stroke-width="1.5" stroke-dasharray="5,2" points="129.6,-57.6 0,-57.6 0,-16 129.6,-16 129.6,-57.6"/>
<text xml:space="preserve" text-anchor="start" x="8" y="-25.2" font-size="14.00">Queue</text>
</g>
<g id="node2" class="node">
<title>sys</title>
<polygon fill="none" stroke="black" points="60,-130 0,-130 0,-74 60,-74 60,-130"/>
<text xml:space="preserve" text-anchor="start" x="8" y="-98" font-size="14.00">Sys</text>
</g>
</g>`

// TestCollectQueueNodeIDs verifies queue IDs are collected from top-level
// nodes and from clusters nested to any depth, excluding non-queue nodes.
func TestCollectQueueNodeIDs(t *testing.T) {
	t.Parallel()

	g := &graph.Graph{
		Nodes: []*graph.Node{
			{ID: "q1", Type: model.TypeQueue},
			{ID: "s1", Type: model.TypeSystem},
		},
		Clusters: []*graph.Cluster{
			{
				ID: "c1",
				Nodes: []*graph.Node{
					{ID: "q2", Type: model.TypeContainerQueue, IsInCluster: true},
				},
				Clusters: []*graph.Cluster{
					{
						ID: "c2",
						Nodes: []*graph.Node{
							{ID: "q3", Type: model.TypeComponentQueue, IsInCluster: true},
							{ID: "db1", Type: model.TypeDb, IsInCluster: true},
						},
					},
				},
			},
		},
	}

	assert.Equal(t, []string{"q1", "q2", "q3"}, collectQueueNodeIDs(g),
		"queue IDs must be collected from top level and nested clusters")
}

// TestCollectQueueNodeIDs_None verifies the empty result for a queue-less
// graph (drives the zero-alloc fast path in the render wiring).
func TestCollectQueueNodeIDs_None(t *testing.T) {
	t.Parallel()

	g := &graph.Graph{
		Nodes: []*graph.Node{
			{ID: "s1", Type: model.TypeSystem},
			{ID: "p1", Type: model.TypePerson},
		},
	}

	assert.Empty(t, collectQueueNodeIDs(g), "non-queue graphs must collect no IDs")
}

// TestPipePathFromPoints verifies the bbox-to-path geometry: the pipe is
// inscribed in the polygon bbox with end-cap ellipses of rx = 0.35*ry, the
// path STARTS with the left bulge arc and ENDS with the right cap arc.
func TestPipePathFromPoints(t *testing.T) {
	t.Parallel()

	d, ok := pipePathFromPoints("10,20 90,20 90,60 10,60")
	require.True(t, ok, "a rect polygon must convert")

	// Bbox (10,20)-(90,60): cy=40, ry=20, rx=7. Body spans x=17..83; the
	// left bulge reaches x0=10 and the right cap reaches x1=90.
	assert.True(t, strings.HasPrefix(d, "M17.00,20.00"), "path anchored at top-left of inscribed body, got %q", d)
	assert.Contains(t, d, "A7.00,20.00", "end-cap ellipses use rx=7 (0.35*ry), ry=20")
	assert.Contains(t, d, "L83.00,60.00", "straight bottom edge spans to x1-rx")
	assert.True(t, strings.HasSuffix(d, "Z"), "path is closed")

	// Starts with an arc command (left bulge) after the initial moveto.
	afterM := strings.TrimPrefix(d, "M")
	afterM = strings.TrimLeft(afterM, "0123456789.,-")
	assert.True(t, strings.HasPrefix(afterM, "A"), "path must start with the left-bulge arc, got %q", afterM)

	// Ends with an arc command (right cap) before the closing Z.
	beforeZ := strings.TrimSuffix(d, "Z")
	assert.True(t, strings.HasSuffix(beforeZ, "83.00,20.00"), "path must end on the right cap arc endpoint, got %q", beforeZ)
	assert.Equal(t, 3, strings.Count(d, "A7.00,20.00"),
		"three arcs: left bulge, right silhouette, right full-ellipse cap")
}

// TestPipePathFromPoints_Degenerate verifies malformed or degenerate polygons
// are rejected rather than producing nonsense geometry.
func TestPipePathFromPoints_Degenerate(t *testing.T) {
	t.Parallel()

	for _, points := range []string{
		"",               // empty
		"10,20",          // single point
		"10,20 10,20",    // zero-area bbox
		"10,x 90,60",     // unparseable coordinate
		"10 20 90 20 90", // odd coordinate count
	} {
		_, ok := pipePathFromPoints(points)
		assert.False(t, ok, "points %q must not convert", points)
	}
}

// TestReplaceQueuePolygons verifies the queue node's polygon becomes a pipe
// path with the polygon's paint attributes copied over, while non-queue
// groups pass through untouched.
func TestReplaceQueuePolygons(t *testing.T) {
	t.Parallel()

	out := string(replaceQueuePolygons([]byte(queuePipeSVGFixture), []string{"platform.queue"}))

	// Queue group: polygon replaced by a path carrying the pipe geometry and
	// the copied paint attributes.
	require.Contains(t, out, `<path d="`, "queue polygon must be replaced by a path")
	assert.NotContains(t, out, `<polygon fill="none" stroke="#073b6f"`,
		"the queue polygon must be gone")
	assert.Contains(t, out, `fill="none" stroke="#073b6f" stroke-width="1.5" stroke-dasharray="5,2"`,
		"fill, stroke, stroke-width and stroke-dasharray must be copied verbatim")

	// Inscribed geometry: bbox (0,-57.6)-(129.6,-16) → cy=-36.8, ry=20.8,
	// rx=7.28; body spans 7.28..122.32.
	assert.Contains(t, out, `d="M7.28,-57.60 A7.28,20.80 0 0,0 7.28,-16.00`,
		"pipe anchored at the polygon bbox")

	// Title and text siblings preserved.
	assert.Contains(t, out, "<title>platform.queue</title>", "group title preserved")
	assert.Contains(t, out, `>Queue</text>`, "queue label text preserved")

	// Non-queue group untouched — same polygon, byte-for-byte.
	assert.Contains(t, out,
		`<polygon fill="none" stroke="black" points="60,-130 0,-130 0,-74 60,-74 60,-130"/>`,
		"non-queue node group must pass through unchanged")
}

// TestReplaceQueuePolygons_NoQueueIDs verifies the zero-queue fast path:
// bytes are returned unchanged (same content, no rewrite pass).
func TestReplaceQueuePolygons_NoQueueIDs(t *testing.T) {
	t.Parallel()

	for _, ids := range [][]string{nil, {}} {
		out := replaceQueuePolygons([]byte(queuePipeSVGFixture), ids)
		assert.Equal(t, queuePipeSVGFixture, string(out), "no queue IDs: bytes must be returned unchanged")
	}
}

// TestReplaceQueuePolygons_RegexMetacharID verifies user-controlled node IDs
// containing regexp metacharacters are matched by PLAIN string search only —
// they must never be interpolated into a regexp (T-QBX-01) and must never
// match a foreign group.
func TestReplaceQueuePolygons_RegexMetacharID(t *testing.T) {
	t.Parallel()

	out := string(replaceQueuePolygons([]byte(queuePipeSVGFixture), []string{`.*`}))
	assert.Equal(t, queuePipeSVGFixture, string(out),
		"a metacharacter ID matches no literal title, so nothing is rewritten")
}

// TestApplyPipeRendering_SVGOnly verifies the wiring contract: SVG bytes get
// the pipe pass, DOT/XDOT bytes are returned untouched (queues stay plain
// boxes in DOT per the locked design).
func TestApplyPipeRendering_Wiring(t *testing.T) {
	// go-graphviz WASM engine has concurrency issues — no parallel.
	g := &graph.Graph{
		Title:     "T",
		Direction: "TB",
		Nodes: []*graph.Node{
			{ID: "q", Label: &graph.Label{Name: "Queue"}, Shape: graph.ShapeRecord, Type: model.TypeQueue},
		},
	}

	svg, err := RenderSVG(g)
	require.NoError(t, err)

	s := string(svg)
	require.Contains(t, s, "<title>q</title>", "precondition: queue node present in SVG")

	start := strings.Index(s, "<title>q</title>")
	groupEnd := strings.Index(s[start:], "</g>")
	require.NotEqual(t, -1, groupEnd)

	queueGroup := s[start : start+groupEnd]
	assert.Contains(t, queueGroup, `<path d="M`, "queue node group must contain the pipe path")
	assert.Contains(t, queueGroup, " A", "pipe path carries arc commands")
	assert.NotContains(t, queueGroup, "<polygon", "queue polygon must be replaced")

	dot, err := RenderDOT(g)
	require.NoError(t, err)

	assert.NotContains(t, string(dot), `<path d="M`, "DOT output must never be pipe-post-processed")
	assert.Contains(t, string(dot), "shape=box", "DOT keeps the plain box anchoring for queues")
}
