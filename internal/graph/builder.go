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
		buildBoundaryViewGraph(v, g)
	} else {
		// C1 view: build nodes and clusters in definition order
		buildC1ViewGraph(v, g)
	}

	// Build edges
	g.Edges = buildEdges(v)

	return g
}

// buildBoundaryViewGraph renders C2/C3 views: boundary/external nodes go at
// top level — outside the expanded unit's cluster — while internal nodes are
// wrapped in the boundary cluster.
func buildBoundaryViewGraph(v *view.View, g *Graph) {
	boundaryCluster := buildBoundaryCluster(v)

	// Build nodes and clusters in definition order
	for _, key := range v.UnitOrder {
		entry := v.Units[key]

		// IsBoundary covers both genuinely external units (actors, external
		// systems) and sibling containers resolved as boundary nodes.
		// IsExternal alone is not sufficient because regular
		// containers/systems are IsExternal=false but still need top-level
		// placement when they're boundary nodes.
		if entry.IsBoundary || entry.IsExternal {
			node := buildNode(entry)
			g.Nodes = append(g.Nodes, node)

			continue
		}

		// Internal nodes go inside the boundary cluster. D-07 (same guard as
		// the C1 branch): expansion only takes effect when there are subunits
		// to show — an expanded-but-empty unit renders as a plain node.
		if entry.IsExpanded && len(entry.Unit.Subunits) > 0 {
			cluster := buildCluster(entry)
			boundaryCluster.Clusters = append(boundaryCluster.Clusters, cluster)
		} else {
			node := buildNode(entry)
			node.IsInCluster = true
			boundaryCluster.Nodes = append(boundaryCluster.Nodes, node)
		}
	}

	g.Clusters = append(g.Clusters, boundaryCluster)
}

// buildC1ViewGraph renders the C1 view: nodes and clusters in definition
// order (from view.UnitOrder). Visible subunits are rendered inside their
// parent cluster by buildCluster — skipping prevents duplicate node IDs in
// DOT. Nil-map reads are safe, so views without VisiblePaths (C2/C3,
// expanded, hand-built) are unaffected.
func buildC1ViewGraph(v *view.View, g *Graph) {
	for _, key := range v.UnitOrder {
		if v.VisiblePaths[key] {
			continue
		}

		entry := v.Units[key]
		// D-07: expansion only takes effect when there are subunits to
		// show — an expanded-but-empty unit renders as a plain node.
		if entry.IsExpanded && len(entry.Unit.Subunits) > 0 {
			cluster := buildCluster(entry)
			g.Clusters = append(g.Clusters, cluster)
		} else {
			node := buildNode(entry)
			g.Nodes = append(g.Nodes, node)
		}
	}
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
		ID:       v.ExpandedUnit,
		Label:    label,
		Nodes:    make([]*Node, 0),
		Clusters: make([]*Cluster, 0),
		Style:    style,
		Type:     unitTypeOrDefault(unit),
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
	// Boxes use content-based styling (border colour derived from subunits)
	var style *NodeStyle
	if IsBoxType(entry.Unit.Type) {
		style = GetBoxStyleByContents(entry.Unit)
	} else {
		style = GetStyleForType(entry.Unit.Type, entry.IsExternal)
	}

	cluster := &Cluster{
		ID:         path,
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

	// Add 🔍 indicator for collapsed units with subunits
	if entry.HasSubunits && !entry.IsExpanded {
		label.Name += " 🔍"
	}

	// Add 📖 indicator for units with an external reference URL (REF-02),
	// mirroring the 🔍 collapsed-cluster affordance style above.
	if entry.Unit.Reference != "" {
		label.Name += " 📖"
	}

	// Boxes use content-based styling (border colour derived from subunits)
	var style *NodeStyle
	if IsBoxType(entry.Unit.Type) {
		style = GetBoxStyleByContents(entry.Unit)
	} else {
		style = GetStyleForType(entry.Unit.Type, entry.IsExternal)
	}

	return &Node{
		ID:           entry.FullPath,
		Label:        label,
		Shape:        ShapeForType(entry.Unit.Type),
		Type:         entry.Unit.Type,
		Style:        style,
		IsExternal:   entry.IsExternal,
		ReferenceURL: entry.Unit.Reference,
	}
}

// buildCluster creates a cluster for an expanded unit.
func buildCluster(entry *view.Entry) *Cluster {
	// Boxes use content-based styling (border colour derived from subunits)
	var style *NodeStyle
	if IsBoxType(entry.Unit.Type) {
		style = GetBoxStyleByContents(entry.Unit)
	} else {
		style = GetStyleForType(entry.Unit.Type, entry.IsExternal)
	}

	cluster := &Cluster{
		ID:         entry.FullPath,
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
	label := &Label{
		Name:        entry.Unit.Name,
		Technology:  entry.Unit.Technology,
		Description: entry.Unit.Description,
		Icon:        IconForType(entry.Unit.Type),
	}

	// Add 📖 indicator for referenced expanded parents (REF-02), mirroring the
	// glyph treatment in buildNode. buildClusterLabel returns a Label (not a
	// Node), so there is no ReferenceURL to populate here; the cluster's
	// drill-down/explore URL handling is unchanged.
	if entry.Unit.Reference != "" {
		label.Name += " 📖"
	}

	return label
}

// isUnitExpanded checks if a child unit should be expanded.
func isUnitExpanded(parent *model.Unit, childName string) bool {
	return slices.Contains(parent.Expanded, childName)
}

// buildEdges creates edges from view links.
func buildEdges(v *view.View) []*Edge {
	edges := make([]*Edge, 0)
	seen := make(map[string]bool) // Track processed links

	// Count contributing links per pair (D-05) before the edge loop so
	// collapsed pairs (2+) can be thickened (D-04).
	pairCounts := countPairMultiplicity(v)

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
		outEdges := processOutgoingLinks(v, path, outLinks, seen, pairCounts)
		edges = append(edges, outEdges...)

		// Process incoming links (linkFrom)
		inEdges := processIncomingLinks(v, path, inLinks, seen, pairCounts)
		edges = append(edges, inEdges...)
	}

	return edges
}

// countPairMultiplicity counts how many contributing links land on each
// (source, target) pair (D-05). The count drives the binary penwidth of
// collapsed edges (D-04). The validator synthesizes a mirror of every outgoing
// link into the target's LinksFrom (internal/validator/index.go) with identical
// attributes; mirrors are marked on the link so they can be excluded, while
// authored linkFrom relationships are counted (WR-02).
func countPairMultiplicity(v *view.View) map[string]int {
	pairCounts := make(map[string]int)

	// Expanded mode keeps the v1.7 dedup/penwidth behavior (D-02, COMPAT-02).
	if v.AllExpanded {
		return pairCounts
	}

	countOutgoingPairs(v, pairCounts)
	countIncomingPairs(v, pairCounts)

	return pairCounts
}

// countOutgoingPairs counts every outgoing link per (source, target) pair.
func countOutgoingPairs(v *view.View, pairCounts map[string]int) {
	for _, path := range v.UnitOrder {
		entry := v.Units[path]
		if entry == nil {
			continue
		}

		outLinks := entry.Unit.Links
		if entry.ResolvedLinks != nil {
			outLinks = entry.ResolvedLinks
		}

		for _, link := range outLinks {
			key := path + "->" + link.Peer
			pairCounts[key]++
		}
	}
}

// countIncomingPairs counts every authored incoming link per pair (D-05).
// Validator-synthesized mirrors carry the Mirror flag and are excluded — they
// are synthetic duplicates of outgoing links, not additional relationships.
func countIncomingPairs(v *view.View, pairCounts map[string]int) {
	for _, path := range v.UnitOrder {
		entry := v.Units[path]
		if entry == nil {
			continue
		}

		inLinks := entry.Unit.LinksFrom
		if entry.ResolvedLinksFrom != nil {
			inLinks = entry.ResolvedLinksFrom
		}

		for _, link := range inLinks {
			if link.Mirror {
				continue
			}

			key := link.Peer + "->" + path
			pairCounts[key]++
		}
	}
}

// processOutgoingLinks processes outgoing links from a unit.
func processOutgoingLinks(
	v *view.View,
	path string,
	links []model.Link,
	seen map[string]bool,
	pairCounts map[string]int,
) []*Edge {
	edges := make([]*Edge, 0)
	sourceEntry := v.Units[path] // Get source entry for color lookup

	for _, link := range links {
		if !isTargetInView(v.Units, link.Peer) {
			continue
		}

		// D-01/D-02: pair-only dedup key in resolved views; expanded mode
		// keeps the v1.7 technology+description key (COMPAT-02).
		edgeKey := path + "->" + link.Peer
		if v.AllExpanded {
			edgeKey += ":" + link.Technology + ":" + link.Description
		}

		// D-04/D-02: binary penwidth — collapsed pairs (2+ links) thicken to
		// 2.0 in resolved views; expanded mode keeps the v1.7 2.0 prominence.
		penWidth := 0.0
		if v.AllExpanded || pairCounts[edgeKey] >= 2 {
			penWidth = 2.0
		}

		edge := createEdge(path, link.Peer, link, sourceEntry, penWidth)

		if markSeen(seen, edgeKey) {
			edges = append(edges, edge)
		}
	}

	return edges
}

// processIncomingLinks processes incoming links (linkFrom) to a unit.
func processIncomingLinks(
	v *view.View,
	path string,
	linksFrom []model.Link,
	seen map[string]bool,
	pairCounts map[string]int,
) []*Edge {
	edges := make([]*Edge, 0)

	for _, link := range linksFrom {
		if !isTargetInView(v.Units, link.Peer) {
			continue
		}

		// D-01/D-02: pair-only dedup key in resolved views; expanded mode
		// keeps the v1.7 technology+description key (COMPAT-02).
		edgeKey := link.Peer + "->" + path
		if v.AllExpanded {
			edgeKey += ":" + link.Technology + ":" + link.Description
		}

		// D-04/D-02: binary penwidth — collapsed pairs (2+ links) thicken to
		// 2.0 in resolved views; expanded mode keeps the v1.7 2.0 prominence.
		penWidth := 0.0
		if v.AllExpanded || pairCounts[edgeKey] >= 2 {
			penWidth = 2.0
		}

		sourceEntry := v.Units[link.Peer] // Source is link.Peer for incoming links
		edge := createEdge(link.Peer, path, link, sourceEntry, penWidth)

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

// kindColour maps a link kind to its edge colour (KIND-01). Unknown kinds
// return "" and fall through to the source-border default. The constants live
// in model/colors.go — shared with the legend so the two can never drift.
func kindColour(kind model.LinkKind) string {
	switch kind {
	case model.KindRead:
		return model.LinkReadColour
	case model.KindWrite:
		return model.LinkWriteColour
	case model.KindReadWrite:
		return model.LinkReadWriteColour
	default:
		return ""
	}
}

// createEdge creates an edge from a link with defaults applied.
// Per D-01: Edge color comes from source unit's border color.
// Per D-03: If link.Color is set, it overrides the source border color.
// Per KIND-01/02: kind colour applies when no explicit color is set;
// explicit Color always wins.
// Per D-04: penWidth carries the binary multiplicity thickness (0 = default,
// 2.0 = collapsed pair).
func createEdge(source, target string, link model.Link, sourceEntry *view.Entry, penWidth float64) *Edge {
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
		PenWidth:  penWidth,
	}

	// Apply defaults
	if edge.Style == "" {
		edge.Style = "solid"
	}

	if edge.Label.Position == "" {
		edge.Label.Position = "middle"
	}

	// Determine color: explicit override -> kind colour -> source border color
	// (D-01, D-03, KIND-01, KIND-02)
	if link.Color != "" {
		edge.Color = link.Color
	} else if colour := kindColour(link.Kind); colour != "" {
		edge.Color = colour
	} else if sourceEntry != nil {
		style := GetStyleForType(sourceEntry.Unit.Type, sourceEntry.IsExternal)
		edge.Color = style.BorderColor
	}

	// Copy length to MinLen (D-01: length > 0 sets minlen attribute)
	edge.MinLen = link.Length

	// rank="equal" excludes the edge from rank computation (constraint=false)
	edge.NoConstraint = link.Rank == model.RankEqual

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
		g.Navigation = buildNavigation(v, basename, format)
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
	case model.TypePerson, model.TypePersonExternal, model.TypeSystemExternal,
		model.TypeDb, model.TypeDbExternal, model.TypeQueue, model.TypeQueueExternal,
		model.TypeContainerDb, model.TypeContainerQueue,
		model.TypeComponent, model.TypeComponentDb, model.TypeComponentQueue:
		// Leaf types — no subunits
		return false
	}

	// Must have subunits to explore
	return entry.HasSubunits && !entry.IsExpanded
}

// buildNavigation creates the Navigation struct for C2/C3 views.
//
// The navigation is breadcrumb-only: the earlier BackLink field is no longer
// populated because it always duplicated the breadcrumb's nearest-ancestor
// entry (same destination, same label). The breadcrumb trail alone covers
// up-navigation to any ancestor. ComputeBackLinkURL is still used internally
// by buildBreadcrumbs to compute the root ancestor's URL.
func buildNavigation(v *view.View, basename, format string) *Navigation {
	nav := &Navigation{}

	// Build breadcrumbs from the expanded unit path
	nav.Breadcrumbs = buildBreadcrumbs(v, basename, format)

	return nav
}

// buildBreadcrumbs creates breadcrumb items from the view.
//
// The breadcrumb trail always includes the C1 root context as the first item,
// followed by every ancestor of the ExpandedUnit, ending with the current
// level (no URL). So a C3 view "mainSystem.sshAuth" produces:
//
//	[Root Context] > [Main System] > [SSH Auth (current)]
//
// Pretty display names are resolved from v.AncestorNames (populated by the
// view generators); if a name is missing the raw dotted-path segment is used
// as a fallback.
func buildBreadcrumbs(v *view.View, basename, format string) []BreadcrumbItem {
	crumbs := BuildBreadcrumbPath(v.ExpandedUnit, basename, format)

	// Resolve each crumb's display name from the view's ancestor-name map.
	// BuildBreadcrumbPath sets Name to the raw path segment (e.g. "mainSystem");
	// we replace it with the unit's pretty Name (e.g. "Main System") when known.
	parts := strings.Split(v.ExpandedUnit, ".")
	for i := range crumbs {
		if i < len(parts) {
			prefix := strings.Join(parts[:i+1], ".")
			if pretty, ok := v.AncestorNames[prefix]; ok && pretty != "" {
				crumbs[i].Name = pretty
			}
		}
	}

	// Prepend the root C1 context as the first breadcrumb so users can always
	// navigate all the way up to C1 from any C2/C3 diagram. Compute the URL
	// relative to the current path depth (C2: ../basename.svg, C3:
	// ../../basename.svg, etc.).
	rootName := v.RootTitle
	if rootName == "" {
		rootName = "Context"
	}

	// The root URL: go up one directory per path segment to reach the output
	// root (where the C1 basename.svg lives). "svg" is hardcoded because
	// clickable navigation URLs always use the browser-navigable .svg format
	// (the Gap-3 contract; see ComputeBackLinkURL and ComputeExploreURL).
	depth := len(parts)
	up := strings.Repeat("../", depth)
	rootURL := up + basename + ".svg"

	rootCrumb := BreadcrumbItem{
		Name: rootName,
		URL:  rootURL,
	}

	crumbs = append([]BreadcrumbItem{rootCrumb}, crumbs...)

	return crumbs
}
