package render

import (
	"regexp"
	"strconv"
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
<polygon fill="none" stroke="#073b6f" stroke-width="1.5" stroke-dasharray="5,2"
 points="129.6,-57.6 0,-57.6 0,-16 129.6,-16 129.6,-57.6"/>
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

// pipePoint is an (x, y) coordinate pair from a pipe d string.
type pipePoint struct {
	x, y float64
}

// pipeSegmentRe splits a machine-generated pipe d string into its command
// segments. The emitter only produces absolute M/L/A/Z commands with numeric
// data between them, so scanning for the command letters is a faithful parse.
//
//nolint:gochecknoglobals // immutable regex; package-level avoids realloc per test (pipeAttrs precedent)
var pipeSegmentRe = regexp.MustCompile(`[MLAZ][^MLAZ]*`)

// arcRecord captures one arc command: the point the pen was at when the arc
// started, the arc's endpoint, and its sweep flag.
type arcRecord struct {
	start, end pipePoint
	sweep      int
}

// pipePointAt parses the first "x,y" pair of a pipe command segment body.
func pipePointAt(t *testing.T, body string) pipePoint {
	t.Helper()

	field := strings.Fields(body)[0]

	comma := strings.Index(field, ",")
	require.Positive(t, comma, "coordinate pair must be comma-separated: %q", field)

	x, errX := strconv.ParseFloat(field[:comma], 64)
	require.NoError(t, errX, "x coordinate must parse in %q", field)

	y, errY := strconv.ParseFloat(field[comma+1:], 64)
	require.NoError(t, errY, "y coordinate must parse in %q", field)

	return pipePoint{x: x, y: y}
}

// pipeArcs parses a machine-generated pipe d string (absolute M/L/A/Z
// commands only) and returns its arcs in emission order, tracking the pen
// point through M/L/A commands. Arc segment format: "Arx,ry rotation
// large,sweep x,y".
func pipeArcs(t *testing.T, d string) []arcRecord {
	t.Helper()

	var (
		arcs []arcRecord
		pen  pipePoint
	)

	for _, seg := range pipeSegmentRe.FindAllString(d, -1) {
		switch seg[0] {
		case 'M', 'L':
			pen = pipePointAt(t, seg[1:])
		case 'A':
			fields := strings.Fields(seg[1:])
			require.Len(t, fields, 4, "arc must carry rx,ry, rotation, flags and endpoint: %q", seg)

			sweep, err := strconv.Atoi(strings.Split(fields[2], ",")[1])
			require.NoError(t, err, "arc sweep flag must parse in %q", seg)

			end := pipePointAt(t, fields[3])
			arcs = append(arcs, arcRecord{start: pen, end: end, sweep: sweep})
			pen = end
		case 'Z':
			// closepath returns the pen to the subpath start; the next
			// subpath's M repositions the pen explicitly.
		}
	}

	return arcs
}

// pipeSubpaths splits a pipe d string into its trimmed subpath segments.
func pipeSubpaths(d string) []string {
	segs := pipeSegmentRe.FindAllString(d, -1)
	for i, seg := range segs {
		segs[i] = strings.TrimSpace(seg)
	}

	return segs
}

// TestPipePathFromPoints verifies the bbox-to-path geometry: the pipe is
// inscribed in the polygon bbox with end-cap ellipses of rx = 0.35*ry. The
// full d is pinned byte-for-byte — deterministic geometry makes full-string
// equality the strongest guard.
func TestPipePathFromPoints(t *testing.T) {
	t.Parallel()

	d, ok := pipePathFromPoints("10,20 90,20 90,60 10,60")
	require.True(t, ok, "a rect polygon must convert")

	// Bbox (10,20)-(90,60): cy=40, ry=20, rx=7. Body spans x=17..83; the
	// left cap reaches x0=10 and the right outer arc reaches x1=90.
	const want = "M17.00,20.00 L83.00,20.00 A7.00,20.00 0 0,1 83.00,60.00" +
		" L17.00,60.00 A7.00,20.00 0 0,1 17.00,20.00 Z" +
		" M83.00,20.00 A7.00,20.00 0 0,0 83.00,60.00"
	assert.Equal(t, want, d, "closed body outline + cap face subpath, in one d")

	assert.True(t, strings.HasPrefix(d, "M17.00,20.00 L83.00,20.00"),
		"path starts with moveto + straight top edge, not an arc, got %q", d)
	assert.Contains(t, d, "L17.00,60.00", "bottom edge runs right → left to x0+rx")
	assert.Equal(t, 3, strings.Count(d, "A7.00,20.00"),
		"three arcs: right outer, left cap, cap face")
}

// TestPipePathFromPoints_NoCoincidentArc is the regression test for the
// capsule bug: an SVG arc whose start and end points coincide is OMITTED by
// renderers, so the old trailing "full ellipse" arc drew nothing and the
// right end rendered as a plain capsule side. No arc command may ever have
// coincident endpoints.
func TestPipePathFromPoints_NoCoincidentArc(t *testing.T) {
	t.Parallel()

	d, ok := pipePathFromPoints("10,20 90,20 90,60 10,60")
	require.True(t, ok, "a rect polygon must convert")

	arcs := pipeArcs(t, d)
	require.Len(t, arcs, 3, "right outer, left cap, cap face")

	for i, arc := range arcs {
		assert.NotEqual(t, arc.start, arc.end,
			"arc %d runs from %v back to %v — SVG omits coincident-endpoint arcs, so it draws nothing",
			i+1, arc.start, arc.end)
	}
}

// TestPipePathFromPoints_CapFaceSubpath verifies the right cap face exists as
// a SECOND subpath in the same d attribute — an M command after the body
// outline's Z — whose arc runs sweep 0 (inner, bulging left into the body)
// from (bodyR,y0) down to (bodyR,y1). The outer silhouette arc and the inner
// face arc together form the visible full ellipse of the right end.
func TestPipePathFromPoints_CapFaceSubpath(t *testing.T) {
	t.Parallel()

	d, ok := pipePathFromPoints("10,20 90,20 90,60 10,60")
	require.True(t, ok, "a rect polygon must convert")

	subpaths := pipeSubpaths(d)
	require.Len(t, subpaths, 2, "body outline + cap face share one d attribute")

	assert.True(t, strings.HasPrefix(subpaths[0], "M17.00,20.00"),
		"first subpath is the body outline, got %q", subpaths[0])
	assert.True(t, strings.HasSuffix(subpaths[0], "Z"), "body outline subpath is closed")

	assert.Equal(t, "M83.00,20.00 A7.00,20.00 0 0,0 83.00,60.00", subpaths[1],
		"cap face: new subpath starting at (bodyR,y0), sweep-0 arc to (bodyR,y1)")
}

// TestPipePathFromPoints_ArcsBulgeOutward verifies the arc directions in the
// y-down SVG coordinates GraphViz emits: the right outer arc from (bodyR,y0)
// to (bodyR,y1) must use sweep 1 (bulging right through (x1,cy)), the left
// cap arc from (bodyL,y1) back to (bodyL,y0) must use sweep 1 (bulging left
// through (x0,cy)), and the cap face arc must use sweep 0 (bulging left
// through (bodyR-rx,cy)) — so the silhouette fills the exact polygon bbox.
func TestPipePathFromPoints_ArcsBulgeOutward(t *testing.T) {
	t.Parallel()

	d, ok := pipePathFromPoints("10,20 90,20 90,60 10,60")
	require.True(t, ok, "a rect polygon must convert")

	arcs := pipeArcs(t, d)
	require.Len(t, arcs, 3, "right outer, left cap, cap face")

	assert.Equal(t, arcRecord{start: pipePoint{x: 83, y: 20}, end: pipePoint{x: 83, y: 60}, sweep: 1}, arcs[0],
		"right outer arc: sweep 1 through (x1,cy)=(90,40)")
	assert.Equal(t, arcRecord{start: pipePoint{x: 17, y: 60}, end: pipePoint{x: 17, y: 20}, sweep: 1}, arcs[1],
		"left cap arc: sweep 1 through (x0,cy)=(10,40)")
	assert.Equal(t, arcRecord{start: pipePoint{x: 83, y: 20}, end: pipePoint{x: 83, y: 60}, sweep: 0}, arcs[2],
		"cap face arc: sweep 0 through (bodyR-rx,cy)=(76,40)")
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
	// rx=7.28; body spans 7.28..122.32. The closed body outline comes first,
	// then the right cap-face subpath in the same d attribute.
	assert.Contains(t, out, `d="M7.28,-57.60 L122.32,-57.60 A7.28,20.80 0 0,1`,
		"pipe body outline starts at the polygon bbox top-left")
	assert.Contains(t, out, "M122.32,-57.60 A7.28,20.80 0 0,0 122.32,-16.00",
		"right cap face subpath present (sweep 0, inner half-ellipse)")

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
	assert.Equal(t, queuePipeSVGFixture, out,
		"a metacharacter ID matches no literal title, so nothing is rewritten")
}

// TestApplyPipeRendering_Wiring verifies the wiring contract: SVG bytes get
// the pipe pass, DOT/XDOT bytes are returned untouched (queues stay plain
// boxes in DOT per the locked design).
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestApplyPipeRendering_Wiring(t *testing.T) {
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
