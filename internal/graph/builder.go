package graph

import (
	"slices"
	"strings"

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

// BuildExpandedGraph constructs a graph with nested clusters for all-expanded mode.
// Unlike BuildGraph, this creates recursive cluster structures that show all nesting levels.
func BuildExpandedGraph(v *view.View) *Graph {
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

	// Find top-level units (those without a dot in their path)
	topLevelUnits := make(map[string]*view.Entry)
	for path, entry := range v.Units {
		if !strings.Contains(path, ".") {
			topLevelUnits[path] = entry
		}
	}

	// Build nodes and nested clusters for top-level units
	for path, entry := range topLevelUnits {
		if entry.HasSubunits {
			cluster := buildNestedCluster(entry, path, v)
			g.Clusters = append(g.Clusters, cluster)
		} else {
			node := buildNode(entry)
			g.Nodes = append(g.Nodes, node)
		}
	}

	// Build edges (handles cross-level connections)
	g.Edges = buildEdges(v)

	return g
}

// buildNestedCluster recursively creates a cluster with nested clusters for subunits.
// This is used by BuildExpandedGraph to show the complete hierarchy in a single diagram.
func buildNestedCluster(entry *view.Entry, path string, v *view.View) *Cluster {
	cluster := &Cluster{
		ID:         "cluster_" + path,
		Label:      buildClusterLabel(entry),
		Nodes:      make([]*Node, 0),
		Clusters:   make([]*Cluster, 0),
		Style:      GetStyleForType(entry.Unit.Type, entry.IsExternal),
		Type:       entry.Unit.Type,
		IsExternal: entry.IsExternal,
	}

	// Process subunits
	for childName, childUnit := range entry.Unit.Subunits {
		childPath := path + "." + childName
		childEntry, exists := v.Units[childPath]
		if !exists {
			// Create entry if not in view (shouldn't happen, but be defensive)
			childEntry = &view.Entry{
				Unit:        childUnit,
				FullPath:    childPath,
				HasSubunits: len(childUnit.Subunits) > 0,
				IsExternal:  view.IsExternalType(childUnit.Type),
			}
		}

		if childEntry.HasSubunits {
			// Recursively build nested cluster
			nestedCluster := buildNestedCluster(childEntry, childPath, v)
			cluster.Clusters = append(cluster.Clusters, nestedCluster)
		} else {
			// Build node for leaf subunit
			node := buildNode(childEntry)
			node.IsInCluster = true
			cluster.Nodes = append(cluster.Nodes, node)
		}
	}

	return cluster
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
		Type:       entry.Unit.Type,
		Style:      GetStyleForType(entry.Unit.Type, entry.IsExternal),
		IsExternal: entry.IsExternal,
	}
}

// buildCluster creates a cluster for an expanded unit.
func buildCluster(entry *view.Entry) *Cluster {
	cluster := &Cluster{
		ID:         "cluster_" + entry.FullPath,
		Label:      buildClusterLabel(entry),
		Nodes:      make([]*Node, 0),
		Style:      GetStyleForType(entry.Unit.Type, entry.IsExternal),
		Type:       entry.Unit.Type,
		IsExternal: entry.IsExternal,
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
	links []model.Link,
	viewUnits map[string]*view.Entry,
	seen map[string]bool,
) []*Edge {
	edges := make([]*Edge, 0)

	for _, link := range links {
		if !isTargetInView(viewUnits, link.Peer) {
			continue
		}

		edge := createEdge(path, link.Peer, link)
		edgeKey := path + "->" + link.Peer + ":" + link.Technology + ":" + link.Description

		if markSeen(seen, edgeKey) {
			edges = append(edges, edge)
		}
	}

	return edges
}

// processIncomingLinks processes incoming links (linkFrom) to a unit.
func processIncomingLinks(
	path string,
	linksFrom []model.Link,
	viewUnits map[string]*view.Entry,
	seen map[string]bool,
) []*Edge {
	edges := make([]*Edge, 0)

	for _, link := range linksFrom {
		if !isTargetInView(viewUnits, link.Peer) {
			continue
		}

		edge := createEdge(link.Peer, path, link)
		edgeKey := link.Peer + "->" + path + ":" + link.Technology + ":" + link.Description

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

// BuildGraphWithPath constructs a graph with navigation paths.
// currentPath is the dotted path of the current diagram (empty for C1).
// basename is the output base filename.
// format is the output format (svg or dot).
func BuildGraphWithPath(v *view.View, currentPath, basename, format string) *Graph {
	g := BuildGraph(v)
	if g == nil {
		return nil
	}

	// Add explore URLs to collapsed nodes with subunits
	for _, node := range g.Nodes {
		if shouldHaveExploreLink(node, v) {
			node.ExploreURL = ComputeExploreURL(currentPath, node.ID, basename, format)
		}
	}

	// Add explore URLs to cluster nodes (if collapsed representation needed)
	for _, cluster := range g.Clusters {
		// Clusters are expanded units, their children might need explore links
		for _, node := range cluster.Nodes {
			if shouldHaveExploreLink(node, v) {
				node.ExploreURL = ComputeExploreURL(currentPath, node.ID, basename, format)
			}
		}
	}

	// Add navigation for C2/C3 views
	if v.Level != view.LevelC1 && v.ExpandedUnit != "" {
		g.Navigation = buildNavigation(v, currentPath, basename, format)
	}

	return g
}

// shouldHaveExploreLink determines if a node should have an explore link.
// Only systems and boxes that have subunits and are collapsed get explore links.
func shouldHaveExploreLink(node *Node, v *view.View) bool {
	entry, exists := v.Units[node.ID]
	if !exists {
		return false
	}

	// Only system and box types can be expanded
	if entry.Unit.Type != model.TypeSystem && entry.Unit.Type != model.TypeBox {
		return false
	}

	// Must have subunits to explore
	return entry.HasSubunits && !entry.IsExpanded
}

// buildNavigation creates the Navigation struct for C2/C3 views.
func buildNavigation(v *view.View, currentPath, basename, format string) *Navigation {
	nav := &Navigation{}

	// Back-link to parent
	nav.BackLink = &BackLink{
		Name: extractParentName(v, currentPath),
		URL:  ComputeBackLinkURL(currentPath, basename, format),
	}

	// Build breadcrumbs from the expanded unit path
	nav.Breadcrumbs = buildBreadcrumbs(v, basename, format)

	return nav
}

// extractParentName gets the display name of the parent unit.
func extractParentName(v *view.View, _ string) string {
	// Use the view's expanded unit path to determine parent name
	if v.ExpandedUnit != "" {
		parts := strings.Split(v.ExpandedUnit, ".")
		if len(parts) > 1 {
			// Return the parent part (second to last)
			return parts[len(parts)-2]
		}
		// Single part means C2, parent is the root
		return v.ExpandedUnit
	}

	return v.Title
}

// buildBreadcrumbs creates breadcrumb items from the view.
func buildBreadcrumbs(v *view.View, basename, format string) []BreadcrumbItem {
	return BuildBreadcrumbPath(v.ExpandedUnit, basename, format)
}
