package graph

import (
	"slices"

	"github.com/Djarvur/c4drill/internal/model"
	"github.com/Djarvur/c4drill/internal/view"
)

// BuildGraph constructs a graph structure from a view.
// The graph contains nodes, edges, and clusters ready for DOT rendering.
func BuildGraph(v *view.View) *Graph {
	if v == nil {
		return nil
	}

	g := &Graph{
		Title:     v.Title,
		Direction: "TB",
		EdgeStyle: v.Edges,
		Nodes:     make([]*Node, 0),
		Edges:     make([]*Edge, 0),
		Clusters:  make([]*Cluster, 0),
	}

	// Build nodes and clusters
	for _, entry := range v.Units {
		if entry.IsExpanded {
			cluster := buildCluster(entry)
			g.Clusters = append(g.Clusters, cluster)
		} else {
			node := buildNode(entry)
			g.Nodes = append(g.Nodes, node)
		}
	}

	// Build edges
	g.Edges = buildEdges(v)

	return g
}

// buildNode creates a node from a view entry.
func buildNode(entry *view.Entry) *Node {
	label := &Label{
		Name:        entry.Unit.Name,
		Technology:  entry.Unit.Technology,
		Description: entry.Unit.Description,
		Icon:        IconForType(entry.Unit.Type),
	}

	// Add [+] indicator for collapsed units with subunits
	if entry.HasSubunits && !entry.IsExpanded {
		label.Name += " [+]"
	}

	return &Node{
		ID:         entry.FullPath,
		Label:      label,
		Shape:      ShapeForType(entry.Unit.Type),
		Style:      GetStyleForType(entry.Unit.Type, entry.IsExternal),
		IsExternal: entry.IsExternal,
	}
}

// buildCluster creates a cluster for an expanded unit.
func buildCluster(entry *view.Entry) *Cluster {
	cluster := &Cluster{
		ID:    "cluster_" + entry.FullPath,
		Label: buildClusterLabel(entry),
		Nodes: make([]*Node, 0),
		Style: GetStyleForType(entry.Unit.Type, entry.IsExternal),
	}

	// Add child units as nodes inside the cluster
	for childName, childUnit := range entry.Unit.Subunits {
		childPath := entry.FullPath + "." + childName
		childEntry := &view.Entry{
			Unit:        childUnit,
			FullPath:    childPath,
			IsExpanded:  isUnitExpanded(entry.Unit, childName),
			HasSubunits: len(childUnit.Subunits) > 0,
			IsExternal:  view.IsExternalType(childUnit.Type),
		}

		node := buildNode(childEntry)
		node.IsInCluster = true
		cluster.Nodes = append(cluster.Nodes, node)
	}

	return cluster
}

// buildClusterLabel creates a label for the cluster (parent unit info).
func buildClusterLabel(entry *view.Entry) *Label {
	return &Label{
		Name:        entry.Unit.Name,
		Technology:  entry.Unit.Technology,
		Description: entry.Unit.Description,
		Icon:        IconForType(entry.Unit.Type),
	}
}

// isUnitExpanded checks if a child unit should be expanded.
func isUnitExpanded(parent *model.Unit, childName string) bool {
	return slices.Contains(parent.Expanded, childName)
}

// buildEdges creates edges from view links.
func buildEdges(v *view.View) []*Edge {
	edges := make([]*Edge, 0)
	seen := make(map[string]bool) // Track processed links

	for path, entry := range v.Units {
		// Process outgoing links
		outEdges := processOutgoingLinks(path, entry.Unit.Links, v.Units, seen)
		edges = append(edges, outEdges...)

		// Process incoming links (linkFrom)
		inEdges := processIncomingLinks(path, entry.Unit.LinksFrom, v.Units, seen)
		edges = append(edges, inEdges...)
	}

	return edges
}

// processOutgoingLinks processes outgoing links from a unit.
func processOutgoingLinks(
	path string,
	links map[string]model.Link,
	viewUnits map[string]*view.Entry,
	seen map[string]bool,
) []*Edge {
	edges := make([]*Edge, 0)

	for target, link := range links {
		if !isTargetInView(viewUnits, target) {
			continue
		}

		edge := createEdge(path, target, link)
		edgeKey := path + "->" + target + ":" + link.Technology + ":" + link.Description

		if markSeen(seen, edgeKey) {
			edges = append(edges, edge)
		}
	}

	return edges
}

// processIncomingLinks processes incoming links (linkFrom) to a unit.
func processIncomingLinks(
	path string,
	linksFrom map[string]model.Link,
	viewUnits map[string]*view.Entry,
	seen map[string]bool,
) []*Edge {
	edges := make([]*Edge, 0)

	for source, link := range linksFrom {
		if !isTargetInView(viewUnits, source) {
			continue
		}

		edge := createEdge(source, path, link)
		edgeKey := source + "->" + path + ":" + link.Technology + ":" + link.Description

		if markSeen(seen, edgeKey) {
			edges = append(edges, edge)
		}
	}

	return edges
}

// isTargetInView checks if a unit exists in the view.
func isTargetInView(units map[string]*view.Entry, target string) bool {
	_, exists := units[target]

	return exists
}

// createEdge creates an edge from a link with defaults applied.
func createEdge(source, target string, link model.Link) *Edge {
	edge := &Edge{
		Source: source,
		Target: target,
		Label: &EdgeLabel{
			Technology:  link.Technology,
			Description: link.Description,
			Position:    string(link.LabelPosition),
		},
		Style:     link.Style,
		ArrowHead: ArrowDirection(link.Arrow),
	}

	// Apply defaults
	if edge.Style == "" {
		edge.Style = "solid"
	}

	if edge.Label.Position == "" {
		edge.Label.Position = "middle"
	}

	return edge
}

// markSeen marks an edge key as seen and returns true if it was not already seen.
func markSeen(seen map[string]bool, edgeKey string) bool {
	if seen[edgeKey] {
		return false
	}

	seen[edgeKey] = true

	return true
}
