package render

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/Djarvur/c4drill/internal/graph"
	"github.com/Djarvur/c4drill/internal/model"
	"github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"
)

// buildHTMLLabelForType returns the appropriate HTML label for a unit type.
func buildHTMLLabelForType(label *graph.Label, t model.UnitType, iconRelPath string) string {
	if label == nil {
		return ""
	}

	switch {
	case graph.IsPersonType(t):
		return buildPersonHTMLLabel(label, iconRelPath)
	case graph.IsDbType(t):
		return buildDbHTMLLabel(label, iconRelPath)
	case graph.IsQueueType(t):
		return buildQueueHTMLLabel(label, iconRelPath)
	case graph.IsSystemType(t):
		return buildSystemHTMLLabel(label, iconRelPath)
	case graph.IsContainerType(t):
		return buildContainerHTMLLabel(label, iconRelPath)
	case graph.IsComponentType(t):
		return buildComponentHTMLLabel(label, iconRelPath)
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

// extractIcon extracts an icon path for a unit, using base64 or file path.
// Returns empty string on error (graceful degradation).
func extractIcon(iconExtractor *IconExtractor, iconType, borderColor string, useBase64 bool) string {
	if iconExtractor == nil || borderColor == "" {
		return ""
	}

	var (
		iconPath string
		err      error
	)

	if useBase64 {
		iconPath, err = iconExtractor.ExtractSVGBase64(iconType, borderColor)
	} else {
		iconPath, err = iconExtractor.Extract(iconType, borderColor)
	}

	if err != nil {
		return "" // Graceful degradation
	}

	return iconPath
}

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
// outputDir is the base directory where the SVG will be written (for icon extraction).
// useBase64 indicates whether to embed icons as base64 data URIs (needed for WASM graphviz).
func buildCgraph(gv *graphviz.Graphviz, g *graph.Graph, outputDir string, useBase64 bool) (*cgraph.Graph, error) {
	cg, err := gv.Graph()
	if err != nil {
		return nil, fmt.Errorf("create cgraph: %w", err)
	}

	// Configure graph-level settings
	configureGraphSettings(cg, g)

	// Create icon extractor for this render (only if outputDir is provided)
	var iconExtractor *IconExtractor
	if outputDir != "" {
		iconExtractor = NewIconExtractor(outputDir)
	}

	// Build node lookup map
	nodeMap := make(map[string]*cgraph.Node)

	// Create top-level nodes (not in clusters)
	if err := createTopLevelNodes(cg, g.Nodes, nodeMap, iconExtractor, useBase64); err != nil {
		return nil, err
	}

	// Create clusters with their nodes
	for _, cluster := range g.Clusters {
		if err := createCluster(cg, cluster, nodeMap, iconExtractor, useBase64); err != nil {
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
	iconExtractor *IconExtractor,
	useBase64 bool,
) error {
	for _, node := range nodes {
		if node.IsInCluster {
			continue // Will be created inside cluster
		}

		cn, err := createNode(cg, node, iconExtractor, useBase64)
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
func configureGraphSettings(cg *cgraph.Graph, g *graph.Graph) {
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

	// Font settings
	cg.SetFontName("Helvetica")
	cg.SetFontSize(fontSizeGraph)

	// Build combined label with navigation and title
	var labelParts []string

	// Navigation bar (back-link + breadcrumbs) for C2/C3
	if g.Navigation != nil {
		navLabel := BuildNavigationLabel(g.Navigation)
		if navLabel != "" {
			labelParts = append(labelParts, navLabel)
		}
	}

	// Graph title
	if g.Title != "" {
		labelParts = append(labelParts, g.Title)
	}

	// Set combined label
	if len(labelParts) > 0 {
		cg.SetLabel(joinLabels(labelParts))
		cg.SetLabelLocation(cgraph.TopLocation)
	}
}

// joinLabels combines multiple label parts with newlines.
func joinLabels(parts []string) string {
	result := ""

	var resultSb120 strings.Builder

	for i, part := range parts {
		if i > 0 {
			resultSb120.WriteString("\n")
		}

		resultSb120.WriteString(part)
	}

	result += resultSb120.String()

	return result
}

// createNode creates a cgraph.Node from a graph.Node.
// useBase64 indicates whether to embed icons as base64 data URIs.
func createNode(
	cg *cgraph.Graph,
	node *graph.Node,
	iconExtractor *IconExtractor,
	useBase64 bool,
) (*cgraph.Node, error) {
	cn, err := cg.CreateNodeByName(node.ID)
	if err != nil {
		return nil, fmt.Errorf("create node by name: %w", err)
	}

	// HTML labels with shape=box and style=rounded for proper visual appearance.
	cn.SetShape(cgraph.BoxShape)

	// Extract icon and build HTML label
	iconRelPath := extractNodeIcon(node, iconExtractor, useBase64)

	if err := setNodeLabel(cg, cn, node, iconRelPath); err != nil {
		return nil, err
	}

	// Apply style
	applyNodeStyle(cn, node.Style)

	// Set URL for clickable nodes (explore links)
	if node.ExploreURL != "" {
		cn.SetURL(node.ExploreURL)
	}

	return cn, nil
}

// extractNodeIcon extracts the icon path for a node.
func extractNodeIcon(node *graph.Node, iconExtractor *IconExtractor, useBase64 bool) string {
	if node.Style == nil {
		return ""
	}

	return extractIcon(iconExtractor, iconTypeForUnit(node.Type), node.Style.BorderColor, useBase64)
}

// setNodeLabel builds and sets the HTML label for a node.
func setNodeLabel(cg *cgraph.Graph, cn *cgraph.Node, node *graph.Node, iconRelPath string) error {
	if node.Label == nil {
		return nil
	}

	htmlLabel := buildHTMLLabelForType(node.Label, node.Type, iconRelPath)
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
// useBase64 indicates whether to embed icons as base64 data URIs.
func createCluster(
	parent *cgraph.Graph,
	cluster *graph.Cluster,
	nodeMap map[string]*cgraph.Node,
	iconExtractor *IconExtractor,
	useBase64 bool,
) error {
	// Name must start with "cluster_" for GraphViz to render as cluster
	subgraph, err := parent.CreateSubGraphByName("cluster_" + cluster.ID)
	if err != nil {
		return fmt.Errorf("create subgraph: %w", err)
	}

	// Extract icon and set label
	iconRelPath := extractClusterIcon(cluster, iconExtractor, useBase64)

	if err := setClusterLabel(parent, subgraph, cluster, iconRelPath); err != nil {
		return err
	}

	// Apply style
	if err := applyClusterStyle(subgraph, cluster.Style); err != nil {
		return err
	}

	// Create nodes inside cluster
	for _, node := range cluster.Nodes {
		cn, err := createNode(subgraph, node, iconExtractor, useBase64)
		if err != nil {
			return fmt.Errorf("create node %s in cluster: %w", node.ID, err)
		}

		nodeMap[node.ID] = cn
	}

	// Create nested clusters recursively
	for _, nestedCluster := range cluster.Clusters {
		if err := createCluster(subgraph, nestedCluster, nodeMap, iconExtractor, useBase64); err != nil {
			return fmt.Errorf("create nested cluster %s: %w", nestedCluster.ID, err)
		}
	}

	return nil
}

// extractClusterIcon extracts the icon path for a cluster.
func extractClusterIcon(cluster *graph.Cluster, iconExtractor *IconExtractor, useBase64 bool) string {
	if cluster.Style == nil {
		return ""
	}

	return extractIcon(iconExtractor, iconTypeForUnit(cluster.Type), cluster.Style.BorderColor, useBase64)
}

// setClusterLabel builds and sets the HTML label for a cluster.
func setClusterLabel(parent *cgraph.Graph, subgraph *cgraph.Graph, cluster *graph.Cluster, iconRelPath string) error {
	if cluster.Label == nil {
		return nil
	}

	htmlLabel := buildHTMLLabelForType(cluster.Label, cluster.Type, iconRelPath)
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

	// Set edge label
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
