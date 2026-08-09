package render

import (
	"fmt"
	"html"
	"strings"
	"unicode"

	"github.com/Djarvur/c4drill/internal/graph"
	"github.com/Djarvur/c4drill/internal/model"
	"github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"
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

const (
	// Border style constants.
	borderStyleDashed = "dashed"

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

// buildStyleString builds a comma-separated style string.
func buildStyleString(style *styleInfo) []string {
	styles := []string{"rounded"}

	if style != nil && style.fillColor != "" {
		styles = append(styles, "filled")
	}

	if style != nil && style.borderStyle == borderStyleDashed {
		styles = append(styles, "dashed")
	}

	return styles
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

	// Create top-level nodes (not in clusters)
	if err := createTopLevelNodes(cg, g.Nodes, nodeMap); err != nil {
		return nil, err
	}

	// Create clusters with their nodes
	for _, cluster := range g.Clusters {
		if err := createCluster(cg, cluster, nodeMap); err != nil {
			return nil, fmt.Errorf("create cluster %s: %w", cluster.ID, err)
		}
	}

	// Create edges
	if err := createEdges(cg, g.Edges, nodeMap); err != nil {
		return nil, err
	}

	return cg, nil
}

// createTopLevelNodes creates nodes that are not inside clusters.
func createTopLevelNodes(
	cg *cgraph.Graph,
	nodes []*graph.Node,
	nodeMap map[string]*cgraph.Node,
) error {
	for _, node := range nodes {
		if node.IsInCluster {
			continue // Will be created inside cluster
		}

		cn, err := createNode(cg, node)
		if err != nil {
			return fmt.Errorf("create node %s: %w", node.ID, err)
		}

		nodeMap[node.ID] = cn
	}

	return nil
}

// createEdges creates all edges from the edge list.
func createEdges(cg *cgraph.Graph, edges []*graph.Edge, nodeMap map[string]*cgraph.Node) error {
	for _, edge := range edges {
		source := nodeMap[edge.Source]
		target := nodeMap[edge.Target]

		if source == nil || target == nil {
			continue // Skip edges with missing endpoints
		}

		if err := createEdge(cg, source, target, edge); err != nil {
			return fmt.Errorf("create edge %s->%s: %w", edge.Source, edge.Target, err)
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
	case "ortho":
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
	navTDs := navigationTDs(g.Navigation)
	hasTitle := g.Title != ""

	if len(navTDs) == 0 && !hasTitle {
		return nil
	}

	var rows []string
	if len(navTDs) > 0 {
		rows = append(rows, "<TR>"+strings.Join(navTDs, "")+"</TR>")
	}

	if hasTitle {
		// COLSPAN merges the title cell across all nav columns. The title is
		// wrapped in <FONT POINT-SIZE="14"> so GraphViz renders it (quirk 2).
		colspan := len(navTDs)
		if colspan < 1 {
			colspan = 1
		}

		titleTD := fmt.Sprintf(`<TD COLSPAN="%d" ALIGN="CENTER"><FONT POINT-SIZE="14">%s</FONT></TD>`,
			colspan, html.EscapeString(g.Title))
		rows = append(rows, "<TR>"+titleTD+"</TR>")
	}

	combinedHTML := "<TABLE BORDER=\"0\" CELLBORDER=\"0\" CELLSPACING=\"0\" CELLPADDING=\"0\" ALIGN=\"CENTER\">" +
		strings.Join(rows, "") + "</TABLE>"

	htmlStr, err := cg.StrdupHTML(combinedHTML)
	if err != nil {
		return fmt.Errorf("create HTML graph label: %w", err)
	}

	cg.SetLabel(htmlStr)
	cg.SetLabelLocation(cgraph.TopLocation)

	return nil
}

// createNode creates a cgraph.Node from a graph.Node.
func createNode(
	cg *cgraph.Graph,
	node *graph.Node,
) (*cgraph.Node, error) {
	cn, err := cg.CreateNodeByName(node.ID)
	if err != nil {
		return nil, fmt.Errorf("create node by name: %w", err)
	}

	// Set shape based on unit type
	// DB uses cylinder shape, all others use box
	if graph.IsDbType(node.Type) {
		cn.SetShape(cgraph.CylinderShape)
	} else {
		cn.SetShape(cgraph.BoxShape)
	}

	// Build and set HTML label
	if err := setNodeLabel(cg, cn, node); err != nil {
		return nil, err
	}

	// Apply style
	applyNodeStyle(cn, node.Style)

	// Set URL for clickable nodes. A GraphViz node has a SINGLE URL attribute,
	// so when an external reference (📖) and an internal drill-down explore URL
	// both apply, the EXTERNAL reference wins the slot (ARCHITECTURE-v1.10
	// §6 (6) Option A). The glyph(s) remain the visible affordance regardless
	// of which URL the slot carries. See converter.go:175-200 for the
	// single-URL-per-node GraphViz limitation that drives this precedence.
	if node.ReferenceURL != "" {
		cn.SetURL(node.ReferenceURL)
	} else if node.ExploreURL != "" {
		cn.SetURL(node.ExploreURL)
	}

	return cn, nil
}

// setNodeLabel builds and sets the HTML label for a node.
func setNodeLabel(cg *cgraph.Graph, cn *cgraph.Node, node *graph.Node) error {
	if node.Label == nil {
		return nil
	}

	htmlLabel := buildHTMLLabelForType(node.Label, node.Type)
	if htmlLabel == "" {
		return nil
	}

	htmlStr, err := cg.StrdupHTML(htmlLabel)
	if err != nil {
		return fmt.Errorf("create HTML label: %w", err)
	}

	cn.SetLabel(htmlStr)

	return nil
}

// applyNodeStyle applies visual styles to a node.
func applyNodeStyle(cn *cgraph.Node, style *graph.NodeStyle) {
	info := nodeStyleToInfo(style)
	styles := buildStyleString(info)

	if style != nil {
		if style.FillColor != "" {
			cn.SetFillColor(style.FillColor)
		}

		if style.FontColor != "" {
			cn.SetFontColor(style.FontColor)
		}

		if style.BorderColor != "" {
			cn.SetColor(style.BorderColor)
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
) error {
	// Name must start with "cluster_" for GraphViz to render as cluster
	subgraph, err := parent.CreateSubGraphByName("cluster_" + cluster.ID)
	if err != nil {
		return fmt.Errorf("create subgraph: %w", err)
	}

	// Set cluster label
	if err := setClusterLabel(parent, subgraph, cluster); err != nil {
		return err
	}

	// Apply style
	if err := applyClusterStyle(subgraph, cluster.Style); err != nil {
		return err
	}

	// Create nodes inside cluster
	for _, node := range cluster.Nodes {
		cn, err := createNode(subgraph, node)
		if err != nil {
			return fmt.Errorf("create node %s in cluster: %w", node.ID, err)
		}

		nodeMap[node.ID] = cn
	}

	// Create nested clusters recursively
	for _, nestedCluster := range cluster.Clusters {
		if err := createCluster(subgraph, nestedCluster, nodeMap); err != nil {
			return fmt.Errorf("create nested cluster %s: %w", nestedCluster.ID, err)
		}
	}

	return nil
}

// setClusterLabel builds and sets the HTML label for a cluster.
func setClusterLabel(parent *cgraph.Graph, subgraph *cgraph.Graph, cluster *graph.Cluster) error {
	if cluster.Label == nil {
		return nil
	}

	htmlLabel := buildHTMLLabelForType(cluster.Label, cluster.Type)
	if htmlLabel == "" {
		return nil
	}

	htmlStr, err := parent.StrdupHTML(htmlLabel)
	if err != nil {
		return fmt.Errorf("create HTML cluster label: %w", err)
	}

	subgraph.SetLabel(htmlStr)

	return nil
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

	// Set combined style string
	info := nodeStyleToInfo(style)
	styles := buildStyleString(info)

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

// createEdge creates a cgraph.Edge from a graph.Edge.
func createEdge(cg *cgraph.Graph, source, target *cgraph.Node, edge *graph.Edge) error {
	// Include description in edge name to allow multiple edges between same nodes
	edgeName := edge.Source + "_to_" + edge.Target
	if edge.Label != nil && edge.Label.Description != "" {
		edgeName += "_" + sanitizeForName(edge.Label.Description)
	}

	e, err := cg.CreateEdgeByName(edgeName, source, target)
	if err != nil {
		return fmt.Errorf("create edge by name: %w", err)
	}

	// Set edge label and font
	e.SetFontName("Helvetica")

	if edge.Label != nil {
		e.SetLabel(buildEdgeLabel(edge.Label))
		e.SetFontSize(fontSizeEdge)
	}

	// Edge style
	switch edge.Style {
	case "dashed":
		e.SetStyle(cgraph.DashedEdgeStyle)
	case "dotted":
		e.SetStyle(cgraph.DottedEdgeStyle)
	default:
		e.SetStyle(cgraph.SolidEdgeStyle)
	}

	// Arrow direction
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

	// Apply edge color (line and label) per D-01 and D-02
	if edge.Color != "" {
		e.SetColor(edge.Color)
		e.SetFontColor(edge.Color)
	}

	// Apply minlen if specified (D-01: length > 0 sets minlen attribute)
	if edge.MinLen > 0 {
		e.SetMinLen(edge.MinLen)
	}

	// Set edge penwidth per D-04: collapsed pairs (2+ links) and --expanded
	// edges carry PenWidth 2.0 from the builder; single edges (PenWidth 0)
	// render at the default 1.0.
	if edge.PenWidth > 0 {
		e.SetPenWidth(edge.PenWidth)
	} else {
		e.SetPenWidth(1.0)
	}

	return nil
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
