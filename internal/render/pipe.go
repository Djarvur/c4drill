package render

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/Djarvur/c4drill/internal/graph"
)

// pipeEndRatio fixes the end-cap ellipse aspect: the cap's horizontal radius
// is this fraction of the pipe's (vertical) radius. 0.35 gives a visibly
// rounded cap without wasting the node's width budget.
const pipeEndRatio = 0.35

// polygonElement matches a GraphViz-emitted <polygon ...> element. GraphViz
// attribute values (points, colours) never contain '>', so matching to the
// first '>' yields the whole element.
var polygonElement = regexp.MustCompile(`<polygon[^>]*>`)

// polygonPoints extracts the points attribute from a polygon element.
var polygonPoints = regexp.MustCompile(`points="([^"]*)"`)

// svgAttribute extracts name="value" attribute pairs from an SVG element.
var svgAttribute = regexp.MustCompile(`([a-zA-Z-]+)="([^"]*)"`)

// pipeAttrs are the paint attributes copied from the replaced polygon onto
// the pipe path, in emission order. GraphViz 15.1.1 emits stroke-dasharray on
// the polygon for dashed node styles — copied verbatim so dashed pipes stay
// dashed.
//
//nolint:gochecknoglobals // immutable spec list; package-level avoids realloc per render (wasmMutex precedent)
var pipeAttrs = []string{"fill", "stroke", "stroke-width", "stroke-dasharray"}

// pipePathFmt is the pipe outline inscribed in a polygon bbox:
// moveto the body's top-left, half-ellipse arc bulging LEFT (left end),
// straight bottom edge, half-ellipse arc bulging RIGHT (right silhouette),
// full-ellipse arc (right cap face, drawn in place), then close along the
// straight top edge. Large-arc+sweep flags assume the y-down SVG coordinate
// system GraphViz emits (the translate transform never mirrors).
const pipePathFmt = "M%.2f,%.2f" +
	" A%.2f,%.2f 0 0,0 %.2f,%.2f" + // left end: half-ellipse bulging left
	" L%.2f,%.2f" + // straight bottom edge
	" A%.2f,%.2f 0 0,0 %.2f,%.2f" + // right silhouette side
	" A%.2f,%.2f 0 1,1 %.2f,%.2f" + // right end: full ellipse cap
	" Z"

// applyPipeRendering post-processes GraphViz SVG bytes so queue-type nodes
// render as horizontal pipes: each queue node's plain rect <polygon> is
// replaced with a <path> drawing a pipe inscribed in the same bounding box.
// Graphs without queues are returned unchanged (zero-alloc fast path).
func applyPipeRendering(g *graph.Graph, svg []byte) []byte {
	queueIDs := collectQueueNodeIDs(g)
	if len(queueIDs) == 0 {
		return svg
	}

	return replaceQueuePolygons(svg, queueIDs)
}

// collectQueueNodeIDs walks the graph — top-level nodes, then clusters
// recursively (nested clusters included) — and returns the IDs of all
// queue-type nodes in stable walk order, deduplicated. Returns nil when the
// graph has no queues.
func collectQueueNodeIDs(g *graph.Graph) []string {
	if g == nil {
		return nil
	}

	seen := make(map[string]bool)
	ids := make([]string, 0, 4)

	ids = appendQueueIDs(ids, seen, g.Nodes)
	ids = appendClusterQueueIDs(ids, seen, g.Clusters)

	if len(ids) == 0 {
		return nil
	}

	return ids
}

// appendQueueIDs appends queue-type node IDs not yet seen.
func appendQueueIDs(ids []string, seen map[string]bool, nodes []*graph.Node) []string {
	for _, node := range nodes {
		if node == nil || seen[node.ID] || !graph.IsQueueType(node.Type) {
			continue
		}

		seen[node.ID] = true
		ids = append(ids, node.ID)
	}

	return ids
}

// appendClusterQueueIDs walks clusters (and nested clusters) appending their
// queue-type node IDs.
func appendClusterQueueIDs(ids []string, seen map[string]bool, clusters []*graph.Cluster) []string {
	for _, cluster := range clusters {
		if cluster == nil {
			continue
		}

		ids = appendQueueIDs(ids, seen, cluster.Nodes)
		ids = appendClusterQueueIDs(ids, seen, cluster.Clusters)
	}

	return ids
}

// replaceQueuePolygons replaces each queue node's <polygon> with a
// horizontal-pipe <path>. Every ID is matched against its GraphViz-emitted
// `<title>ID</title>` group marker; IDs with no group in the SVG (or whose
// group has no polygon) are skipped.
func replaceQueuePolygons(svg []byte, queueIDs []string) []byte {
	if len(queueIDs) == 0 {
		return svg
	}

	s := string(svg)
	for _, id := range queueIDs {
		s = replaceQueuePolygon(s, id)
	}

	return []byte(s)
}

// replaceQueuePolygon rewrites at most one node group: the group whose title
// is exactly id. The ID is used in a plain string search — NEVER interpolated
// into a regexp — because node IDs come from user-authored TOML (T-QBX-01).
// The shape candidate is bounded by the group's own closing </g> so a missing
// polygon can never pull a later group's shape into the rewrite.
func replaceQueuePolygon(s, id string) string {
	needle := "<title>" + id + "</title>"

	titleStart := strings.Index(s, needle)
	if titleStart < 0 {
		return s
	}

	searchFrom := titleStart + len(needle)

	groupEnd := strings.Index(s[searchFrom:], "</g>")
	if groupEnd < 0 {
		return s
	}

	limit := searchFrom + groupEnd

	loc := polygonElement.FindStringIndex(s[searchFrom:])
	if loc == nil || searchFrom+loc[0] >= limit {
		return s
	}

	elStart, elEnd := searchFrom+loc[0], searchFrom+loc[1]
	element := s[elStart:elEnd]

	points := polygonPoints.FindStringSubmatch(element)
	if points == nil {
		return s
	}

	d, ok := pipePathFromPoints(points[1])
	if !ok {
		return s
	}

	return s[:elStart] + `<path d="` + d + `"` + copiedPipeAttrs(element) + `/>` + s[elEnd:]
}

// copiedPipeAttrs extracts the pipe paint attributes (fill, stroke,
// stroke-width, stroke-dasharray) from the replaced polygon element so the
// pipe path inherits the node's exact colours and line style.
func copiedPipeAttrs(element string) string {
	values := make(map[string]string, len(pipeAttrs))
	for _, match := range svgAttribute.FindAllStringSubmatch(element, -1) {
		values[match[1]] = match[2]
	}

	var b strings.Builder

	for _, attr := range pipeAttrs {
		if v, ok := values[attr]; ok {
			b.WriteString(` `)
			b.WriteString(attr)
			b.WriteString(`="`)
			b.WriteString(v)
			b.WriteString(`"`)
		}
	}

	return b.String()
}

// pipePathFromPoints parses a GraphViz polygon points list ("x,y x,y ...")
// into its bounding box and returns the inscribed pipe path.
func pipePathFromPoints(points string) (string, bool) {
	box, ok := parsePointsBBox(points)
	if !ok {
		return "", false
	}

	return pipePathFromBBox(box.x0, box.y0, box.x1, box.y1), true
}

// bbox is the axis-aligned bounding box of a polygon's coordinates.
type bbox struct {
	x0, y0, x1, y1 float64
}

// parsePointsBBox computes the min/max bounding box of an SVG points list.
// Rejects empty, unparseable, or zero-area lists (parsing is bounded by the
// GraphViz-generated input itself — a finite node polygon — so no unbounded
// work is possible here, T-QBX-02).
func parsePointsBBox(points string) (bbox, bool) {
	fields := strings.Fields(points)
	if len(fields) < 2 {
		return bbox{}, false
	}

	box := bbox{}

	first := true

	for _, field := range fields {
		comma := strings.Index(field, ",")
		if comma < 0 {
			return bbox{}, false
		}

		x, errX := strconv.ParseFloat(field[:comma], 64)

		y, errY := strconv.ParseFloat(field[comma+1:], 64)

		if errX != nil || errY != nil {
			return bbox{}, false
		}

		if first {
			box = bbox{x0: x, y0: y, x1: x, y1: y}
			first = false

			continue
		}

		box.x0 = min(box.x0, x)
		box.y0 = min(box.y0, y)
		box.x1 = max(box.x1, x)
		box.y1 = max(box.y1, y)
	}

	if box.x1-box.x0 <= 0 || box.y1-box.y0 <= 0 {
		return bbox{}, false
	}

	return box, true
}

// pipePathFromBBox draws the horizontal pipe inscribed in the bbox
// (x0,y0)-(x1,y1): the mid line cy=(y0+y1)/2 is implicit in the arc endpoints,
// ry the pipe radius, rx = 0.35*ry the end-cap extent. The left cap bulges to
// exactly x0 and the right cap face reaches exactly x1, so the pipe fills the
// former polygon's footprint and edge anchors (computed on the box bbox) stay
// valid.
func pipePathFromBBox(x0, y0, x1, y1 float64) string {
	ry := (y1 - y0) / 2
	rx := ry * pipeEndRatio

	bodyL := x0 + rx
	bodyR := x1 - rx

	return fmt.Sprintf(pipePathFmt,
		bodyL, y0, // moveto: body top-left
		rx, ry, bodyL, y1, // left end: bulge left through (x0, cy)
		bodyR, y1, // straight bottom edge
		rx, ry, bodyR, y0, // right silhouette side through (x1, cy)
		rx, ry, bodyR, y0, // right end: full ellipse cap (same start/end point)
	)
}
