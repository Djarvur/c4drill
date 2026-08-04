package graph

import (
	"slices"
	"strings"

	"github.com/Djarvur/c4drill/internal/model"
	"github.com/Djarvur/c4drill/internal/view"
)

// BuildGraph constructs a graph structure from a view.
// The graph contains nodes, edges, and clusters ready for DOT rendering.
// For C2/C3 views, internal nodes are wrapped in a boundary cluster representing
// the expanded system/container. External boundary nodes remain at the top level.
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

	// For C2/C3 views, wrap internal nodes in a boundary cluster
	if v.Level != view.LevelC1 && v.ExpandedUnit != "" {
		boundaryCluster := buildBoundaryCluster(v)

		// Build nodes and clusters in definition order
		for _, key := range v.UnitOrder {
			entry := v.Units[key]

			// External boundary nodes go at top level
			if entry.IsExternal {
				node := buildNode(entry)
				g.Nodes = append(g.Nodes, node)
				continue
			}

			// Internal nodes go inside the boundary cluster
			if entry.IsExpanded {
				cluster := buildCluster(entry)
				boundaryCluster.Clusters = append(boundaryCluster.Clusters, cluster)
			} else {
				node := buildNode(entry)
				node.IsInCluster = true
				boundaryCluster.Nodes = append(boundaryCluster.Nodes, node)
			}
		}

		g.Clusters = append(g.Clusters, boundaryCluster)
	} else {
		// C1 view: build nodes and clusters in definition order (from view.UnitOrder)
		for _, key := range v.UnitOrder {
			entry := v.Units[key]
			if entry.IsExpanded {
				cluster := buildCluster(entry)
				g.Clusters = append(g.Clusters, cluster)
			} else {
				node := buildNode(entry)
				g.Nodes = append(g.Nodes, node)
			}
		}
	}

	// Build edges
	g.Edges = buildEdges(v)

	return g
}

// buildBoundaryCluster creates a cluster representing the expanded system/container
// boundary for C2/C3 views. Internal nodes and sub-clusters are placed inside this
// cluster, while external boundary nodes remain at the graph's top level.
func buildBoundaryCluster(v *view.View) *Cluster {
	unit := v.ExpandedUnitModel
	var label *Label
	var style *NodeStyle

	if unit != nil {
		label = &Label{
			Name:        unit.Name,
			Technology:  unit.Technology,
			Description: unit.Description,
			Icon:        IconForType(unit.Type),
		}
		style = GetStyleForType(unit.Type, false)
	} else {
		// Fallback: derive name from ExpandedUnit path
		name := v.ExpandedUnit
		if idx := strings.LastIndex(name, "."); idx >= 0 {
			name = name[idx+1:]
		}
		label = &Label{Name: name}
		style = &NodeStyle{
			BorderColor: model.PersonBorder,
			FontColor:   model.PersonBorder,
		}
	}

	return &Cluster{
		ID:        v.ExpandedUnit,
		Label:     label,
		Nodes:     make([]*Node, 0),
		Clusters:  make([]*Cluster, 0),
		Style:     style,
		Type:      unitTypeOrDefault(unit),
	}
}

// unitTypeOrDefault returns the unit type or TypeSystem as default.
func unitTypeOrDefault(unit *model.Unit) model.UnitType {
	if unit == nil {
		return model.TypeSystem
	}
	return unit.Type
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

	// Find top-level units (those without a dot in their path) in definition order
	topLevelUnits := make(map[string]*view.Entry)

	for path, entry := range v.Units {
		if !strings.Contains(path, ".") {
			topLevelUnits[path] = entry
		}
	}

	// Build nodes and nested clusters for top-level units in definition order
	for _, path := range v.UnitOrder {
		entry := topLevelUnits[path]
		if entry == nil {
			continue // Skip non-top-level entries
		}

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
	// For C1 boxes, use content-based styling
	var style *NodeStyle
	if entry.Unit.Type == model.TypeBox {
		style = GetBoxStyleByContents(entry.Unit)
	} else {
		style = GetStyleForType(entry.Unit.Type, entry.IsExternal)
	}

	cluster := &Cluster{
		ID:         "cluster_" + path,
		Label:      buildClusterLabel(entry),
		Nodes:      make([]*Node, 0),
		Clusters:   make([]*Cluster, 0),
		Style:      style,
		Type:       entry.Unit.Type,
		IsExternal: entry.IsExternal,
	}

	// Process subunits in definition order (use SubunitOrder if available)
	var childOrder []string
	if len(entry.Unit.SubunitOrder) > 0 {
		childOrder = entry.Unit.SubunitOrder
	} else {
		// Fallback to map keys for test models without explicit order
		for name := range entry.Unit.Subunits {
			childOrder = append(childOrder, name)
		}
	}

	for _, childName := range childOrder {
		childUnit := entry.Unit.Subunits[childName]
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

	// For C1 boxes, use content-based styling
	var style *NodeStyle
	if entry.Unit.Type == model.TypeBox {
		style = GetBoxStyleByContents(entry.Unit)
	} else {
		style = GetStyleForType(entry.Unit.Type, entry.IsExternal)
	}

	return &Node{
		ID:         entry.FullPath,
		Label:      label,
		Shape:      ShapeForType(entry.Unit.Type),
		Type:       entry.Unit.Type,
		Style:      style,
		IsExternal: entry.IsExternal,
	}
}

// buildCluster creates a cluster for an expanded unit.
func buildCluster(entry *view.Entry) *Cluster {
	// For C1 boxes, use content-based styling
	var style *NodeStyle
	if entry.Unit.Type == model.TypeBox {
		style = GetBoxStyleByContents(entry.Unit)
	} else {
		style = GetStyleForType(entry.Unit.Type, entry.IsExternal)
	}

	cluster := &Cluster{
		ID:         "cluster_" + entry.FullPath,
		Label:      buildClusterLabel(entry),
		Nodes:      make([]*Node, 0),
		Style:      style,
		Type:       entry.Unit.Type,
		IsExternal: entry.IsExternal,
	}

	// Add child units as nodes inside the cluster in definition order
	var childOrder []string
	if len(entry.Unit.SubunitOrder) > 0 {
		childOrder = entry.Unit.SubunitOrder
	} else {
		// Fallback to map keys for test models without explicit order
		for name := range entry.Unit.Subunits {
			childOrder = append(childOrder, name)
		}
	}

	for _, childName := range childOrder {
		childUnit := entry.Unit.Subunits[childName]
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

	// Process edges in definition order
	for _, path := range v.UnitOrder {
		entry := v.Units[path]

		// Use resolved links when available (for C1 views with edge resolution),
		// otherwise fall back to the unit's direct links.
		outLinks := entry.Unit.Links
		if entry.ResolvedLinks != nil {
			outLinks = entry.ResolvedLinks
		}

		inLinks := entry.Unit.LinksFrom
		if entry.ResolvedLinksFrom != nil {
			inLinks = entry.ResolvedLinksFrom
		}

		// Process outgoing links
		outEdges := processOutgoingLinks(path, outLinks, v.Units, seen)
		edges = append(edges, outEdges...)

		// Process incoming links (linkFrom)
		inEdges := processIncomingLinks(path, inLinks, v.Units, seen)
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
	sourceEntry := viewUnits[path] // Get source entry for color lookup

	for _, link := range links {
		if !isTargetInView(viewUnits, link.Peer) {
			continue
		}

		edge := createEdge(path, link.Peer, link, sourceEntry)
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

		sourceEntry := viewUnits[link.Peer] // Source is link.Peer for incoming links
		edge := createEdge(link.Peer, path, link, sourceEntry)
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
// Per D-01: Edge color comes from source unit's border color.
// Per D-03: If link.Color is set, it overrides the source border color.
func createEdge(source, target string, link model.Link, sourceEntry *view.Entry) *Edge {
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

	// Determine color: explicit override -> source border color (D-01, D-03)
	if link.Color != "" {
		edge.Color = link.Color
	} else if sourceEntry != nil {
		style := GetStyleForType(sourceEntry.Unit.Type, sourceEntry.IsExternal)
		edge.Color = style.BorderColor
	}

	// Copy length to MinLen (D-01: length > 0 sets minlen attribute)
	edge.MinLen = link.Length

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

	// Only types that can contain subunits can be expanded
	switch entry.Unit.Type {
	case model.TypeSystem, model.TypeBox,
		model.TypeContainer, model.TypeContainerBox,
		model.TypeComponentBox:
		// These types can have subunits
	default:
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
		// Single part means C2, parent is the root diagram
		if v.RootTitle != "" {
			return v.RootTitle
		}

		return v.ExpandedUnit
	}

	return v.Title
}

// buildBreadcrumbs creates breadcrumb items from the view.
// For C2 views (single-segment ExpandedUnit), it prepends a root breadcrumb
// pointing to the C1 diagram.
func buildBreadcrumbs(v *view.View, basename, format string) []BreadcrumbItem {
	crumbs := BuildBreadcrumbPath(v.ExpandedUnit, basename, format)

	// For C2 views, the single-segment path produces only one breadcrumb
	// (the current level). Prepend the root C1 diagram as parent breadcrumb.
	if v.Level == view.LevelC2 && len(crumbs) > 0 {
		rootName := v.RootTitle
		if rootName == "" {
			rootName = "Context"
		}

		rootCrumb := BreadcrumbItem{
			Name: rootName,
			URL:  ComputeBackLinkURL(v.ExpandedUnit, basename, format),
		}

		crumbs = append([]BreadcrumbItem{rootCrumb}, crumbs...)
	}

	return crumbs
}
