package render

import (
	"fmt"
	"html"
	"strings"
	"unicode"

	"github.com/Djarvur/c4drill/internal/graph"
	"github.com/Djarvur/c4drill/internal/model"
	"github.com/onokonem/go-graphviz"
	"github.com/onokonem/go-graphviz/cgraph"
)

// buildHTMLLabelForType returns the appropriate HTML label for a unit type.
func buildHTMLLabelForType(label *graph.Label, t model.UnitType) string {
	if label == nil {
		return ""
	}

	switch {
	case graph.IsPersonType(t):
		return buildPersonHTMLLabel(label)
	case graph.IsDbType(t):
		return buildDbHTMLLabel(label)
	case graph.IsQueueType(t):
		return buildQueueHTMLLabel(label)
	case graph.IsSystemType(t):
		return buildSystemHTMLLabel(label)
	case graph.IsContainerType(t):
		return buildContainerHTMLLabel(label)
	case graph.IsComponentType(t):
		return buildComponentHTMLLabel(label)
	case graph.IsBoxType(t):
		return buildBoxHTMLLabel(label)
	default:
		// Fallback to generic record label for unknown types
		return buildRecordLabel(label)
	}
}

// buildEdgePlainTextLabel renders an edge label as plain text for plain mode
// (PLAIN-03): the [Technology] bracket convention plus the description — the
// same content the HTML rectangle carries, without the HTML. Empty when the
// label carries neither field, mirroring buildEdgeLabel's empty contract.
func buildEdgePlainTextLabel(label *graph.EdgeLabel) string {
	if label == nil {
		return ""
	}

	var parts []string

	if label.Technology != "" {
		parts = append(parts, "["+label.Technology+"]")
	}

	if label.Description != "" {
		parts = append(parts, label.Description)
	}

	return strings.Join(parts, " ")
}

const (
	// Border style constants.
	borderStyleDashed = "dashed"
	borderStyleDotted = "dotted"

	// Font size constants.
	fontSizeGraph = 14.0
	fontSizeEdge  = 10.0
)

// styleInfo contains style attributes for nodes and clusters.
type styleInfo struct {
	fillColor   string
	fontColor   string
	borderColor string
	borderStyle string
}

// buildStyleString builds a comma-separated style string. The "rounded"
// prefix is conditional: queue-type nodes omit it so GraphViz emits a plain
// rect <polygon> the SVG pipe post-processor (pipe.go) can parse and replace.
func buildStyleString(style *styleInfo, rounded bool) []string {
	var styles []string

	if rounded {
		styles = append(styles, "rounded")
	}

	if style != nil && style.fillColor != "" {
		styles = append(styles, "filled")
	}

	switch {
	case style == nil:
	case style.borderStyle == borderStyleDashed:
		styles = append(styles, "dashed")
	case style.borderStyle == borderStyleDotted:
		styles = append(styles, "dotted")
	}

	return styles
}

// contentCluster wraps all diagram content in the invisible root cluster the
// legend floats outside of. Graphs without content skip the wrapper — a
// legend-only diagram needs nothing to separate from.
func contentCluster(cg *cgraph.Graph, g *graph.Graph) (*cgraph.Graph, error) {
	if len(g.Nodes) == 0 && len(g.Clusters) == 0 {
		return cg, nil
	}

	wrapper, err := cg.CreateSubGraphByName(contentClusterName)
	if err != nil {
		return nil, fmt.Errorf("create content cluster: %w", err)
	}

	if err := wrapper.SafeSet("style", "invis", ""); err != nil {
		return nil, fmt.Errorf("hide content cluster: %w", err)
	}

	// Clusters inherit the root graph's label (nav + title) and would render
	// it a second time above the cluster; an explicit empty label stops the
	// inheritance.
	if err := wrapper.SafeSet("label", "", ""); err != nil {
		return nil, fmt.Errorf("clear content cluster label: %w", err)
	}

	return wrapper, nil
}

// buildCgraph converts a graph.Graph to a cgraph.Graph for rendering.
func buildCgraph(
	gv *graphviz.Graphviz,
	g *graph.Graph,
) (*cgraph.Graph, error) {
	cg, err := gv.Graph()
	if err != nil {
		return nil, fmt.Errorf("create cgraph: %w", err)
	}

	// Configure graph-level settings
	if err := configureGraphSettings(cg, g); err != nil {
		return nil, fmt.Errorf("configure graph settings: %w", err)
	}

	// Build node lookup map
	nodeMap := make(map[string]*cgraph.Node)

	// All diagram content goes inside an invisible root cluster; the legend
	// node stays outside it. Dot packs cluster-external nodes beside the
	// cluster's bounding box, never inside it, so the legend is geometrically
	// separated from the content and can never be mistaken for a diagram
	// element (user-reported tangle). No content → no wrapper, a legend-only
	// graph needs nothing to separate from.
	content, err := contentCluster(cg, g)
	if err != nil {
		return nil, err
	}

	// Create top-level nodes (not in clusters)
	if err := createTopLevelNodes(content, g.Nodes, nodeMap, g.Opts); err != nil {
		return nil, err
	}

	// Create clusters with their nodes
	for _, cluster := range g.Clusters {
		if err := createCluster(content, cluster, nodeMap, g.Opts); err != nil {
			return nil, fmt.Errorf("create cluster %s: %w", cluster.ID, err)
		}
	}

	// Create edges
	if err := createEdges(cg, g.Edges, nodeMap, clusterIndex(g.Clusters), g.Opts); err != nil {
		return nil, err
	}

	// Floating legend node (LEG-01)
	if g.Legend != nil && len(g.Legend.Entries) > 0 {
		if err := createLegendNode(cg, g.Legend); err != nil {
			return nil, err
		}
	}

	return cg, nil
}

// createLegendNode emits the legend as an isolated floating node rather than
// rows in the graph label. All content lives inside the invisible content
// cluster, so dot packs the legend beside the cluster's bounding box — the
// upper-right of the diagram (LEG-01) — never inside the content, and
// without pushing the nav/title label around. shape=plaintext keeps the node
// borderless so only the legend table itself is drawn.
func createLegendNode(cg *cgraph.Graph, legend *graph.Legend) error {
	cn, err := cg.CreateNodeByName(legendNodeName)
	if err != nil {
		return fmt.Errorf("create legend node: %w", err)
	}

	cn.SetShape(cgraph.PlainTextShape)
	cn.SetLabelHTML(BuildLegendLabel(legend))

	return nil
}

// createTopLevelNodes creates nodes that are not inside clusters.
func createTopLevelNodes(
	cg *cgraph.Graph,
	nodes []*graph.Node,
	nodeMap map[string]*cgraph.Node,
	opts graph.RenderOpts,
) error {
	for _, node := range nodes {
		if node.IsInCluster {
			continue // Will be created inside cluster
		}

		cn, err := createNode(cg, node, opts)
		if err != nil {
			return fmt.Errorf("create node %s: %w", node.ID, err)
		}

		nodeMap[node.ID] = cn
	}

	return nil
}

// clusterIndex maps every cluster (nested included) by its dotted path so
// edge endpoints that render as clusters can be recognised.
func clusterIndex(clusters []*graph.Cluster) map[string]*graph.Cluster {
	index := make(map[string]*graph.Cluster)

	var add func(cs []*graph.Cluster)

	add = func(cs []*graph.Cluster) {
		for _, c := range cs {
			index[c.ID] = c
			add(c.Clusters)
		}
	}

	add(clusters)

	return index
}

// firstNodeIn returns the first node (depth-first, declaration order) inside
// the cluster subtree — the structural anchor a compound edge attaches to
// when its logical endpoint is the cluster itself. Nil for a cluster with no
// nodes anywhere beneath it.
func firstNodeIn(c *graph.Cluster) *graph.Node {
	if c == nil {
		return nil
	}

	if len(c.Nodes) > 0 {
		return c.Nodes[0]
	}

	for _, nested := range c.Clusters {
		if node := firstNodeIn(nested); node != nil {
			return node
		}
	}

	return nil
}

// resolveEndpoint returns the cgraph node an edge endpoint attaches to and,
// when the endpoint is a unit that renders as a CLUSTER (issue #23 — an
// expanded C1 unit or a CTX-02 deep-link chain container, with no cgraph node
// of its own), the cluster_ name to clip the arrow at. The edge anchors on
// the first node inside the cluster subtree; (nil, "") when the endpoint
// resolves to neither a node nor a node-bearing cluster.
func resolveEndpoint(
	nodeMap map[string]*cgraph.Node,
	clusters map[string]*graph.Cluster,
	path string,
) (*cgraph.Node, string) {
	if node := nodeMap[path]; node != nil {
		return node, ""
	}

	c := clusters[path]
	if c == nil {
		return nil, ""
	}

	anchor := firstNodeIn(c)
	if anchor == nil {
		return nil, ""
	}

	return nodeMap[anchor.ID], "cluster_" + c.ID
}

// createEdges creates all edges from the edge list.
func createEdges(
	cg *cgraph.Graph,
	edges []*graph.Edge,
	nodeMap map[string]*cgraph.Node,
	clusters map[string]*graph.Cluster,
	opts graph.RenderOpts,
) error {
	// Fallback name uniquifier for hand-built graphs whose edges carry no
	// builder-assigned Name (BUG-3): parallel fallback edges stay distinct
	// without ever deriving names from flag-suppressible label content.
	nameCounts := make(map[string]int)

	// Set on the graph only when a compound edge was actually emitted, so
	// existing flag-off DOT/SVG output stays byte-identical.
	compound := false

	for _, edge := range edges {
		// Issue #23: cluster-rendered endpoints anchor inside their cluster
		// and clip at its boundary (compound ltail/lhead) instead of being
		// silently dropped.
		source, sourceCluster := resolveEndpoint(nodeMap, clusters, edge.Source)
		target, targetCluster := resolveEndpoint(nodeMap, clusters, edge.Target)

		if source == nil || target == nil {
			continue // Skip edges with missing endpoints
		}

		compound = compound || sourceCluster != "" || targetCluster != ""

		err := createEdge(cg, source, target, edge, opts, nameCounts, sourceCluster, targetCluster)
		if err != nil {
			return fmt.Errorf("create edge %s->%s: %w", edge.Source, edge.Target, err)
		}
	}

	if compound {
		if err := cg.SafeSet("compound", "true", ""); err != nil {
			return fmt.Errorf("set compound: %w", err)
		}
	}

	return nil
}

// configureGraphSettings applies graph-level settings from graph.Graph.
func configureGraphSettings(cg *cgraph.Graph, g *graph.Graph) error {
	// Layout direction
	if g.Direction == "LR" {
		cg.SetRankDir(cgraph.LRRank)
	} else {
		cg.SetRankDir(cgraph.TBRank)
	}

	// Edge routing style
	switch g.EdgeStyle {
	case "spline":
		cg.SetSplines("true")
	case "straight":
		cg.SetSplines("false")
	case "ortho", "square":
		// "square" is the documented alias for ortho routing (GEDGE-02)
		cg.SetSplines("ortho")
	}

	// Font settings - set default fontname for all element types
	cg.SetFontName("Helvetica")
	cg.SetFontSize(fontSizeGraph)
	// Set default fontname for nodes (kind=1) and edges (kind=2)
	if _, err := cg.Attr(1, "fontname", "Helvetica"); err != nil { // nodes
		return fmt.Errorf("set node fontname: %w", err)
	}

	if _, err := cg.Attr(2, "fontname", "Helvetica"); err != nil { // edges
		return fmt.Errorf("set edge fontname: %w", err)
	}

	// Build combined HTML graph label with navigation and title (Gap 2).
	//
	// GraphViz HTML-like labels do NOT support <a href> tags — a label
	// containing them is silently dropped at render time, which is why the
	// pre-fix navigation bar appeared as escaped literal text
	// (&lt;a href=&quot;...&quot;&gt;) when the label was plain text, and
	// disappeared entirely when wrapped as an HTML label with raw <a href>.
	// Clickable links inside an HTML label are expressed via the HREF
	// attribute on a <TD> element (rendered as <a xlink:href> in SVG). The
	// navigation TDs are produced by navigationTDs; the plain-text title is
	// HTML-escaped before embedding (threat T-03-04-02). The breadcrumb and
	// title are placed in a single borderless TABLE.
	//
	// Two GraphViz quirks drive this layout:
	//  1. Clickable breadcrumb items must be RIGHT-aligned within their cells.
	//     GraphViz sizes each table column to the widest cell in that column;
	//     left-anchored text in a wide cell leaves a gap before the next
	//     separator. RIGHT-aligning the clickable cells puts the text flush
	//     against the following ">" separator, eliminating the gap.
	//  2. The title MUST be wrapped in an explicit <FONT POINT-SIZE="14"> tag.
	//     When row 1 carries <FONT POINT-SIZE="10"> content and row 2 carries
	//     plain (default-size) content, GraphViz silently drops the title row
	//     from the rendered SVG. Wrapping both rows in explicit FONT tags (10
	//     for nav, 14 for title) makes both render.
	// The nav and title share one multi-row HTML graph label. The legend is
	// NOT part of it — it renders as a separate floating node (LEG-01), see
	// createLegendNode.
	navTDs := navigationTDs(g.Navigation)
	hasTitle := g.Title != ""

	if len(navTDs) == 0 && !hasTitle {
		return nil
	}

	colspan := len(navTDs)
	if colspan < 1 {
		colspan = 1
	}

	var rows []string
	if len(navTDs) > 0 {
		rows = append(rows, "<TR>"+strings.Join(navTDs, "")+"</TR>")
	}

	if hasTitle {
		// COLSPAN merges the title cell across all nav columns. The title is
		// wrapped in <FONT POINT-SIZE="14"> so GraphViz renders it (quirk 2).
		titleTD := fmt.Sprintf(`<TD COLSPAN="%d" ALIGN="CENTER"><FONT POINT-SIZE="14">%s</FONT></TD>`,
			colspan, html.EscapeString(g.Title))
		rows = append(rows, "<TR>"+titleTD+"</TR>")
	}

	combinedHTML := "<TABLE BORDER=\"0\" CELLBORDER=\"0\" CELLSPACING=\"0\" CELLPADDING=\"0\" ALIGN=\"CENTER\">" +
		strings.Join(rows, "") + "</TABLE>"

	// SetLabelHTML stores the label as a true HTML-like string (graphviz
	// agsafeset_html). Since graphviz 13, the StrdupHTML+SetLabel round-trip
	// loses HTML-ness (the string dict keys on is_html) and the breadcrumb
	// renders as escaped text with dead TD HREF links.
	cg.SetLabelHTML(combinedHTML)
	cg.SetLabelLocation(cgraph.TopLocation)

	return nil
}

// createNode creates a cgraph.Node from a graph.Node.
func createNode(
	cg *cgraph.Graph,
	node *graph.Node,
	opts graph.RenderOpts,
) (*cgraph.Node, error) {
	cn, err := cg.CreateNodeByName(node.ID)
	if err != nil {
		return nil, fmt.Errorf("create node by name: %w", err)
	}

	// Set shape based on unit type
	// DB uses cylinder shape, all others use box. Queues keep shape=box: edge
	// anchors are computed on the box bbox and the pipe drawn by the SVG
	// post-processor (pipe.go) is inscribed in that same bbox.
	if graph.IsDbType(node.Type) {
		cn.SetShape(cgraph.CylinderShape)
	} else {
		cn.SetShape(cgraph.BoxShape)
	}

	// Build and set label (HTML by default, plain text under --plain)
	setNodeLabel(cn, node, opts)

	// Apply style (queue nodes drop "rounded", see applyNodeStyle)
	applyNodeStyle(cn, node)

	// Queue nodes read as pipes only when wider than the default minimum
	// (graphviz width = minimum width in inches; it may still grow the node
	// for long labels). The x-margin (inches!) insets the HTML label ~40pt
	// from each border — a guaranteed clearance zone the pipe's end-cap
	// ellipses (rx <= pipeMaxCap = 16pt, 2rx <= margin) never cross, for any
	// label length. buildQueueHTMLLabel keeps ratio-based wrapping, so long
	// descriptions grow the node WIDER (a pipe), not taller (a tower).
	if graph.IsQueueType(node.Type) {
		if err := cn.SafeSet("width", "2.6", ""); err != nil {
			return nil, fmt.Errorf("set queue node width %s: %w", node.ID, err)
		}

		if err := cn.SafeSet("margin", "0.55,0.06", ""); err != nil {
			return nil, fmt.Errorf("set queue node margin %s: %w", node.ID, err)
		}
	}

	// Set URL for clickable nodes. A GraphViz node has a SINGLE URL attribute
	// (converter.go:175-200), so when a unit has BOTH a drill-down and an
	// external reference (📖), the DRILL-DOWN wins the slot: navigation is
	// the primary action, and the unit's docs render on its own child
	// diagram (the boundary frame there links to the reference). A leaf
	// without a drill-down carries the reference on the node itself.
	if node.ExploreURL != "" {
		cn.SetURL(node.ExploreURL)
	} else if node.ReferenceURL != "" {
		cn.SetURL(node.ReferenceURL)
	}

	return cn, nil
}

// setNodeLabel builds and sets the label for a node: the HTML label by
// default, the plain-text record label under plain (--plain, PLAIN-03).
// fix 260831-01u (edge-labels-only --no-labels): node labels are NOT suppressed by
// --no-labels — only edge label text is.
func setNodeLabel(cn *cgraph.Node, node *graph.Node, opts graph.RenderOpts) {
	if node.Label == nil {
		return
	}

	if opts.Plain {
		// Plain mode routes labels to the record path (labels.go): no HTML
		// tables, name/technology/description content preserved. SetLabel
		// stores a true plain string — graphviz escapes it — so author text
		// can never inject DOT/HTML markup (threat T-37-07).
		if label := buildRecordLabel(node.Label); label != "" {
			cn.SetLabel(label)
		}

		return
	}

	htmlLabel := buildHTMLLabelForType(node.Label, node.Type)
	if htmlLabel == "" {
		return
	}

	cn.SetLabelHTML(htmlLabel)
}

// applyNodeStyle applies visual styles to a node. Queue-type nodes omit the
// "rounded" style so GraphViz emits a parseable plain rect <polygon> that the
// SVG pipe post-processor (pipe.go) replaces with a horizontal-pipe <path>;
// fill/border colours and dashed/dotted borders are unchanged.
func applyNodeStyle(cn *cgraph.Node, node *graph.Node) {
	info := nodeStyleToInfo(node.Style)
	styles := buildStyleString(info, !graph.IsQueueType(node.Type))

	if node.Style != nil {
		if node.Style.FillColor != "" {
			cn.SetFillColor(node.Style.FillColor)
		}

		if node.Style.FontColor != "" {
			cn.SetFontColor(node.Style.FontColor)
		}

		if node.Style.BorderColor != "" {
			cn.SetColor(node.Style.BorderColor)
		}
	}

	_ = cn.SafeSet("style", strings.Join(styles, ","), "")
}

// nodeStyleToInfo converts graph.NodeStyle to styleInfo.
func nodeStyleToInfo(style *graph.NodeStyle) *styleInfo {
	if style == nil {
		return nil
	}

	return &styleInfo{
		fillColor:   style.FillColor,
		fontColor:   style.FontColor,
		borderColor: style.BorderColor,
		borderStyle: style.BorderStyle,
	}
}

// createCluster creates a subgraph cluster from a graph.Cluster.
func createCluster(
	parent *cgraph.Graph,
	cluster *graph.Cluster,
	nodeMap map[string]*cgraph.Node,
	opts graph.RenderOpts,
) error {
	// Name must start with "cluster_" for GraphViz to render as cluster
	subgraph, err := parent.CreateSubGraphByName("cluster_" + cluster.ID)
	if err != nil {
		return fmt.Errorf("create subgraph: %w", err)
	}

	// Set cluster label
	setClusterLabel(subgraph, cluster, opts)

	// Apply style
	if err := applyClusterStyle(subgraph, cluster.Style); err != nil {
		return err
	}

	// Create nodes inside cluster
	for _, node := range cluster.Nodes {
		cn, err := createNode(subgraph, node, opts)
		if err != nil {
			return fmt.Errorf("create node %s in cluster: %w", node.ID, err)
		}

		nodeMap[node.ID] = cn
	}

	// Create nested clusters recursively
	for _, nestedCluster := range cluster.Clusters {
		if err := createCluster(subgraph, nestedCluster, nodeMap, opts); err != nil {
			return fmt.Errorf("create nested cluster %s: %w", nestedCluster.ID, err)
		}
	}

	return nil
}

// setClusterLabel builds and sets the label for a cluster: the HTML label by
// default, the plain-text record label under plain (--plain, PLAIN-03).
// fix 260831-01u (edge-labels-only --no-labels): cluster labels are NOT suppressed by
// --no-labels — only edge label text is. The drill-down URL emission is
// structural and survives both.
func setClusterLabel(subgraph *cgraph.Graph, cluster *graph.Cluster, opts graph.RenderOpts) {
	switch {
	case cluster.Label == nil:
		return
	case opts.Plain:
		// Plain mode routes cluster labels to the record path (labels.go) —
		// no HTML tables, content preserved, escaping via SetLabel (T-37-07).
		if label := buildRecordLabel(cluster.Label); label != "" {
			subgraph.SetLabel(label)
		}
	default:
		htmlLabel := buildHTMLLabelForType(cluster.Label, cluster.Type)
		if htmlLabel == "" {
			return
		}

		subgraph.SetLabelHTML(htmlLabel)
	}

	// CTX-03: a container cluster carries its drill-down URL as the subgraph
	// URL attribute — the cluster-side analog of the node's SetURL (GraphViz
	// supports URL on clusters, so clicking the cluster frame drills into the
	// unit's dedicated diagram). SafeSet errors are ignored, matching the
	// best-effort treatment in applyNodeStyle. The URL is a structural
	// affordance: emitted under plain too. Same single-slot precedence as
	// nodes: drill-down wins when both apply — the cluster unit's 📖 docs
	// render on its child diagram's boundary frame, not here. The one
	// exception is the boundary cluster itself (a unit ON its own child
	// diagram): it has no drill-down of its own, so its reference takes the
	// slot.
	url := cluster.ExploreURL
	if url == "" {
		url = cluster.ReferenceURL
	}

	if url != "" {
		_ = subgraph.SafeSet("URL", url, "")
	}
}

// applyClusterStyle applies visual styles to a cluster.
func applyClusterStyle(subgraph *cgraph.Graph, style *graph.NodeStyle) error {
	if style == nil {
		if err := subgraph.SafeSet("style", "rounded", ""); err != nil {
			return fmt.Errorf("set cluster style: %w", err)
		}

		return nil
	}

	// Set fill color
	if style.FillColor != "" {
		subgraph.SetBackgroundColor(style.FillColor)
	}

	// Set font color
	if err := setClusterAttribute(subgraph, "fontcolor", style.FontColor); err != nil {
		return err
	}

	// Set font name for cluster labels
	if err := setClusterAttribute(subgraph, "fontname", "Helvetica"); err != nil {
		return err
	}

	// Set border color
	if err := setClusterAttribute(subgraph, "color", style.BorderColor); err != nil {
		return err
	}

	// Set combined style string (clusters always stay rounded)
	info := nodeStyleToInfo(style)
	styles := buildStyleString(info, true)

	if err := subgraph.SafeSet("style", strings.Join(styles, ","), ""); err != nil {
		return fmt.Errorf("set cluster style: %w", err)
	}

	return nil
}

// setClusterAttribute sets a cluster attribute if the value is non-empty.
func setClusterAttribute(subgraph *cgraph.Graph, attr, value string) error {
	if value == "" {
		return nil
	}

	if err := subgraph.SafeSet(attr, value, ""); err != nil {
		return fmt.Errorf("set cluster %s: %w", attr, err)
	}

	return nil
}

// edgeNameFor resolves the cgraph identity for an edge: the builder-assigned
// unique name verbatim when present, otherwise a uniquified
// "{source}_to_{target}" fallback for hand-built graphs.
//
// fix 260831-01u (flag-invariant edge identity): the builder-assigned name is
// model-derived, sanitized, and flag-independent — so CreateEdgeByName's
// find-or-create can never silently merge two builder edges whose label
// content was suppressed by a formatting flag (threats T-Q1-01/T-Q1-02).
// Label content never contributes to edge identity.
func edgeNameFor(edge *graph.Edge, nameCounts map[string]int) string {
	name := edge.Name
	if name != "" {
		return name
	}

	name = sanitizeForName(edge.Source + "_to_" + edge.Target)

	nameCounts[name]++

	if n := nameCounts[name]; n > 1 {
		return fmt.Sprintf("%s_%d", name, n)
	}

	return name
}

// createEdge creates a cgraph.Edge from a graph.Edge. sourceCluster /
// targetCluster carry the cluster_ name a compound endpoint was substituted
// for (issue #23) — empty when the endpoint is a plain node.
func createEdge(
	cg *cgraph.Graph, source, target *cgraph.Node, edge *graph.Edge,
	opts graph.RenderOpts, nameCounts map[string]int, sourceCluster, targetCluster string,
) error {
	edgeName := edgeNameFor(edge, nameCounts)

	// rank="reverse" (RANK-01): flip layout ranking by swapping the endpoints
	// at emission while keeping the visual arrow pointed at the logical target.
	// The edge name stays logical (names are layout-irrelevant and this keeps
	// tests readable).
	tail, head := source, target
	if edge.RankReverse {
		tail, head = target, source
	}

	e, err := cg.CreateEdgeByName(edgeName, tail, head)
	if err != nil {
		return fmt.Errorf("create edge by name: %w", err)
	}

	// Compound clip attributes (issue #23) attach to the DRAWN endpoints, so
	// a rank=reverse swap flips which cluster is tail and which is head.
	tailCluster, headCluster := sourceCluster, targetCluster
	if edge.RankReverse {
		tailCluster, headCluster = targetCluster, sourceCluster
	}

	if tailCluster != "" {
		if err := e.SafeSet("ltail", tailCluster, ""); err != nil {
			return fmt.Errorf("set ltail: %w", err)
		}
	}

	if headCluster != "" {
		if err := e.SafeSet("lhead", headCluster, ""); err != nil {
			return fmt.Errorf("set lhead: %w", err)
		}
	}

	// Set edge label and font
	e.SetFontName("Helvetica")

	setEdgeLabel(e, edge, opts)

	// Edge style
	switch edge.Style {
	case "dashed":
		e.SetStyle(cgraph.DashedEdgeStyle)
	case "dotted":
		e.SetStyle(cgraph.DottedEdgeStyle)
	default:
		e.SetStyle(cgraph.SolidEdgeStyle)
	}

	setEdgeDir(e, edge)

	applyEdgeAttributes(e, edge)

	return nil
}

// setEdgeLabel emits the edge label: an explicit empty label under --no-labels,
// plain text under plain (PLAIN-03), the HTML rectangle otherwise. Labels with
// no technology and no description keep the plain SetLabel("") path — see the
// COMPAT-01 note on the HTML branch.
func setEdgeLabel(e *cgraph.Edge, edge *graph.Edge, opts graph.RenderOpts) {
	if opts.NoLabels {
		// fix 260831-01u (edge-labels-only --no-labels): no edge label under
		// --no-labels — an explicit empty label keeps the emitted DOT clean
		// even for hand-built graphs carrying a non-nil Label
		// (defense-in-depth; the builder already drops it). Edge labels are
		// the ONLY thing the flag suppresses.
		e.SetLabel("")

		return
	}

	if edge.Label == nil {
		return
	}

	// Plain mode (PLAIN-03) emits the label as plain text via SetLabel —
	// never the HTML rectangle. SetLabel routes through the same escaping
	// as every existing plain label (threat T-37-07).
	if opts.Plain {
		if plainLabel := buildEdgePlainTextLabel(edge.Label); plainLabel != "" {
			e.SetLabel(plainLabel)
		} else {
			e.SetLabel("")
		}
	} else if htmlLabel := buildEdgeLabel(edge.Label); htmlLabel != "" {
		// SetLabelHTML preserves HTML-ness under graphviz 13 (same
		// agsafeset_html path nodes use — see setNodeLabel). Empty labels
		// (no technology, no description) keep the plain SetLabel("") path:
		// SafeSetHTML with "" omits the attribute, which would change the
		// emitted DOT for label-less edges and break the COMPAT-01 goldens.
		e.SetLabelHTML(htmlLabel)
	} else {
		e.SetLabel("")
	}

	e.SetFontSize(fontSizeEdge)
}

// setEdgeDir emits the per-edge dir attribute. Under rank="reverse" the
// endpoints are swapped (see createEdge), so the dir attribute inverts to
// keep the arrowhead pointing at the logical target. SetDir(ForwardDir)
// emits nothing (value equals the declared default) — do not introduce a
// per-edge dir=forward emission.
func setEdgeDir(e *cgraph.Edge, edge *graph.Edge) {
	if edge.RankReverse {
		switch edge.ArrowHead {
		case graph.ArrowReverse:
			// arrowhead at the head end — which after the swap is the original
			// source; the default (no dir attr) already draws it there.
		case graph.ArrowBoth:
			e.SetDir(cgraph.BothDir)
		case graph.ArrowNone:
			e.SetDir(cgraph.NoneDir)
		case graph.ArrowForward:
			e.SetDir(cgraph.BackDir)
		default:
			e.SetDir(cgraph.BackDir)
		}

		return
	}

	switch edge.ArrowHead {
	case graph.ArrowForward:
		e.SetDir(cgraph.ForwardDir)
	case graph.ArrowReverse:
		e.SetDir(cgraph.BackDir)
	case graph.ArrowBoth:
		e.SetDir(cgraph.BothDir)
	case graph.ArrowNone:
		e.SetDir(cgraph.NoneDir)
	}
}

// applyEdgeAttributes sets the scalar cgraph attributes (color, minlen,
// constraint, penwidth) derived from the builder edge. Mirrors the
// setNodeLabel/applyNodeStyle and setClusterLabel/applyClusterStyle split
// for nodes and clusters.
func applyEdgeAttributes(e *cgraph.Edge, edge *graph.Edge) {
	// Apply edge color (line and label) per D-01 and D-02
	if edge.Color != "" {
		e.SetColor(edge.Color)
		e.SetFontColor(edge.Color)
	}

	// Apply minlen if specified (D-01: length > 0 sets minlen attribute)
	if edge.MinLen > 0 {
		e.SetMinLen(edge.MinLen)
	}

	// rank="equal": exclude from rank computation so endpoints can share a rank
	if edge.NoConstraint {
		e.SetConstraint(false)
	}

	// Set edge penwidth per D-04: collapsed pairs (2+ links) and --expanded
	// edges carry PenWidth 2.0 from the builder; single edges (PenWidth 0)
	// render at the default 1.0.
	if edge.PenWidth > 0 {
		e.SetPenWidth(edge.PenWidth)
	} else {
		e.SetPenWidth(1.0)
	}
}

// sanitizeForName converts a string to a safe identifier for use in edge/node names.
// Replaces spaces and special characters with underscores.
func sanitizeForName(s string) string {
	var result strings.Builder

	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			result.WriteRune(r)
		} else {
			result.WriteRune('_')
		}
	}

	return result.String()
}
