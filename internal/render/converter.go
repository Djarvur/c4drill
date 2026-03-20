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

// buildCgraph converts a graph.Graph to a cgraph.Graph for rendering.
func buildCgraph(gv *graphviz.Graphviz, g *graph.Graph) (*cgraph.Graph, error) {
	cg, err := gv.Graph()
	if err != nil {
		return nil, fmt.Errorf("create cgraph: %w", err)
	}

	// Configure graph-level settings
	configureGraphSettings(cg, g)

	// Build node lookup map
	nodeMap := make(map[string]*cgraph.Node)

	// Create top-level nodes (not in clusters)
	for _, node := range g.Nodes {
		if node.IsInCluster {
			continue // Will be created inside cluster
		}

		cn, err := createNode(cg, node)
		if err != nil {
			return nil, fmt.Errorf("create node %s: %w", node.ID, err)
		}

		nodeMap[node.ID] = cn
	}

	// Create clusters with their nodes
	for _, cluster := range g.Clusters {
		if err := createCluster(cg, cluster, nodeMap); err != nil {
			return nil, fmt.Errorf("create cluster %s: %w", cluster.ID, err)
		}
	}

	// Create edges
	for _, edge := range g.Edges {
		source := nodeMap[edge.Source]

		target := nodeMap[edge.Target]
		if source == nil || target == nil {
			continue // Skip edges with missing endpoints
		}

		if err := createEdge(cg, source, target, edge); err != nil {
			return nil, fmt.Errorf("create edge %s->%s: %w", edge.Source, edge.Target, err)
		}
	}

	return cg, nil
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
func createNode(cg *cgraph.Graph, node *graph.Node) (*cgraph.Node, error) {
	cn, err := cg.CreateNodeByName(node.ID)
	if err != nil {
		return nil, fmt.Errorf("create node by name: %w", err)
	}

	// HTML labels with shape=box and style=rounded for proper visual appearance.
	// This combination provides a clean container look that works well with HTML tables.
	cn.SetShape(cgraph.BoxShape)

	// Build and set the label using HTML tables
	// IMPORTANT: Use StrdupHTML to create HTML strings that GraphViz will
	// recognize as HTML (not quoted strings). Without this, SetLabel wraps
	// values in quotes which breaks HTML label parsing.
	if node.Label != nil {
		htmlLabel := buildHTMLLabelForType(node.Label, node.Type)
		if htmlLabel != "" {
			htmlStr, err := cg.StrdupHTML(htmlLabel)
			if err != nil {
				return nil, fmt.Errorf("create HTML label: %w", err)
			}
			cn.SetLabel(htmlStr)
		}
	}

	// Build combined style string - start with rounded for record shapes
	styles := []string{"rounded"}

	if node.Style != nil {
		// Add filled style if FillColor is specified
		if node.Style.FillColor != "" {
			styles = append(styles, "filled")
			cn.SetFillColor(node.Style.FillColor)
		}

		if node.Style.FontColor != "" {
			cn.SetFontColor(node.Style.FontColor)
		}

		if node.Style.BorderColor != "" {
			cn.SetColor(node.Style.BorderColor)
		}

		if node.Style.BorderStyle == borderStyleDashed {
			styles = append(styles, "dashed")
		}
	}

	// Set combined style using SafeSet (must be called on Node, not Base())
	cn.SafeSet("style", strings.Join(styles, ","), "")

	// Set URL for clickable nodes (explore links)
	if node.ExploreURL != "" {
		cn.SetURL(node.ExploreURL)
	}

	return cn, nil
}

// createCluster creates a subgraph cluster from a graph.Cluster.
func createCluster(parent *cgraph.Graph, cluster *graph.Cluster, nodeMap map[string]*cgraph.Node) error {
	// Name must start with "cluster_" for GraphViz to render as cluster
	subgraph, err := parent.CreateSubGraphByName("cluster_" + cluster.ID)
	if err != nil {
		return fmt.Errorf("create subgraph: %w", err)
	}

	// Cluster label - use HTML format for consistent styling with nodes
	if cluster.Label != nil {
		htmlLabel := buildHTMLLabelForType(cluster.Label, cluster.Type)
		if htmlLabel != "" {
			htmlStr, err := parent.StrdupHTML(htmlLabel)
			if err != nil {
				return fmt.Errorf("create HTML cluster label: %w", err)
			}
			subgraph.SetLabel(htmlStr)
		}
	}

	// Build combined style string - start with rounded for clusters
	styles := []string{"rounded"}

	if cluster.Style != nil {
		// Add filled style if FillColor is specified
		if cluster.Style.FillColor != "" {
			styles = append(styles, "filled")
			subgraph.SetBackgroundColor(cluster.Style.FillColor)
		}

		// Set font color for cluster label
		if cluster.Style.FontColor != "" {
			subgraph.SafeSet("fontcolor", cluster.Style.FontColor, "")
		}

		// Set border color (cluster uses 'color' for border)
		if cluster.Style.BorderColor != "" {
			subgraph.SafeSet("color", cluster.Style.BorderColor, "")
		}

		if cluster.Style.BorderStyle == borderStyleDashed {
			styles = append(styles, "dashed")
		}
	}

	// Set combined style using SafeSet (called on Graph which embeds Object)
	subgraph.SafeSet("style", strings.Join(styles, ","), "")

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
