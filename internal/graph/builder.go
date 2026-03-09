package graph

import (
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
			cluster := buildCluster(entry, v)
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
		label.Name = label.Name + " [+]"
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
func buildCluster(entry *view.Entry, v *view.View) *Cluster {
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
	for _, exp := range parent.Expanded {
		if exp == childName {
			return true
		}
	}
	return false
}

// buildEdges creates edges from view links.
func buildEdges(v *view.View) []*Edge {
	edges := make([]*Edge, 0)
	seen := make(map[string]bool) // Track processed links

	for path, entry := range v.Units {
		// Process outgoing links
		for target, link := range entry.Unit.Links {
			edgeKey := path + "->" + target + ":" + link.Technology + ":" + link.Description

			// Only include if target is in view (as regular or boundary node)
			if _, exists := v.Units[target]; !exists {
				continue
			}

			edge := &Edge{
				Source: path,
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

			// Track this edge to avoid duplicates
			if !seen[edgeKey] {
				seen[edgeKey] = true
				edges = append(edges, edge)
			}
		}

		// Process incoming links (linkFrom)
		for source, link := range entry.Unit.LinksFrom {
			edgeKey := source + "->" + path + ":" + link.Technology + ":" + link.Description

			// Only include if source is in view
			if _, exists := v.Units[source]; !exists {
				continue
			}

			edge := &Edge{
				Source: source,
				Target: path,
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

			// Track this edge to avoid duplicates
			if !seen[edgeKey] {
				seen[edgeKey] = true
				edges = append(edges, edge)
			}
		}
	}

	return edges
}
