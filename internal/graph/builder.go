package graph

import (
	"maps"
	"slices"
	"strconv"
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

	// Build edges, then the legend — kind-colour rows need the resolved edge
	// colours so the legend only explains what the diagram actually draws.
	g.Edges = buildEdges(v)
	g.Legend = buildLegend(v, g.Edges)

	return g
}

// buildBoundaryViewGraph renders C2/C3 views: boundary/external nodes go at
// top level — outside the expanded unit's cluster — while internal nodes are
// wrapped in the boundary cluster. VisiblePaths entries (CTX-02 deep-link
// chain entries) are skipped as standalone nodes — the ancestor-cluster
// recursion renders them at their proper depth (mirrors buildC1ViewGraph).
func buildBoundaryViewGraph(v *view.View, g *Graph) {
	boundaryCluster := buildBoundaryCluster(v)

	// Build nodes and clusters in definition order
	for _, key := range v.UnitOrder {
		// CTX-02: chain entries are in Units for edge building but render
		// only inside their ancestor cluster. Nil-map reads are safe for
		// views without VisiblePaths.
		if v.VisiblePaths[key] {
			continue
		}

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
		// CTX-02: UnfoldChain entries (collapsed ancestors with an inserted
		// deep-link chain) unfold the same way.
		if (entry.IsExpanded || entry.UnfoldChain) && len(entry.Unit.Subunits) > 0 {
			cluster := buildCluster(entry, v)
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
		// CTX-02: UnfoldChain entries (collapsed ancestors with an inserted
		// deep-link chain) unfold as recursive clusters too, so the true
		// link target exists as a real node inside its chain.
		if (entry.IsExpanded || entry.UnfoldChain) && len(entry.Unit.Subunits) > 0 {
			cluster := buildCluster(entry, v)
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
		applyUnitOverrides(style, unit)
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

	// Build edges (handles cross-level connections), then the legend —
	// kind-colour rows need the resolved edge colours.
	g.Edges = buildEdges(v)
	g.Legend = buildLegend(v, g.Edges)

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

	applyUnitOverrides(style, entry.Unit)

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

	applyUnitOverrides(style, entry.Unit)

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

// applyUnitOverrides applies author-specified color/style/border fields on top
// of the palette-derived style (COLOR-01/COLOR-02). Explicit author fields win
// over both the level palette and the box-content heuristic — author intent
// beats heuristics. A dark explicit fill forces a white font (luminance rule)
// so labels stay legible; unset fields leave the style untouched.
func applyUnitOverrides(style *NodeStyle, unit *model.Unit) {
	if style == nil || unit == nil {
		return
	}

	if unit.Color != "" {
		style.FillColor = unit.Color
		if luminance(unit.Color) < 0.5 {
			style.FontColor = "#FFFFFF"
		}
	}

	if unit.Border != "" {
		style.BorderColor = unit.Border
	}

	if unit.Style != "" {
		style.BorderStyle = unit.Style
	}
}

// buildLegend assembles the diagram legend (LEG-01..03). Every row documents
// a convention the view actually uses — nothing is explained that the diagram
// does not show:
//
//  1. element rows, one per entity kind present (person, system, db, queue,
//     container, component) plus one per external entity kind ("external
//     system", …) — in the exact colours the renderer puts on those units;
//  2. link-kind colours (read/write/read-write) — only for kinds whose
//     colour survived into a drawn edge (explicit author colours win over
//     kind colours, so unresolved colours would be a lie);
//  3. author-defined custom lines (LEG-03).
//
// Line styles (solid/dashed/dotted) are deliberately not explained: the
// samples would be text approximations of a visual pattern, and a reader who
// needs them can see the pattern on the diagram itself. Colors come from the
// model constants so the legend can never drift from the renderer.
func buildLegend(v *view.View, edges []*Edge) *Legend {
	if v == nil || !v.ShowLegend {
		return nil
	}

	entries := legendElementEntries(v)
	entries = append(entries, legendKindEntries(edges)...)
	entries = append(entries, legendCustomEntries(v)...)

	if len(entries) == 0 {
		return nil
	}

	return &Legend{Entries: entries}
}

// legendElementEntries emits one row per entity kind present in the view —
// person, system, container, component and level-qualified db/queue variants
// ("system db", "container db", "component db", …) — in that entity's colour,
// then one row per external entity kind ("external system", "external db",
// …) in the external grey. Row colour always mirrors the renderer: internal
// entities take their level border colour, external entries the level's
// external grey. Grouping boxes fold into their level's entity row
// (containerBox → container) — they are layout, not a colour of their own.
// The expanded unit is scanned too: its boundary cluster carries the same
// level colour as the entity it represents.
func legendElementEntries(v *view.View) []LegendEntry {
	// Canonical row order: internal entities first (by level), then external
	// ones. db/queue carry their level in the label ("system db", "container
	// db", …) because the same entity kind exists at several levels — the
	// label is what disambiguates two same-kind rows rendered in different
	// level colours.
	order := make(map[string]int, len(legendRowOrder))
	for i, label := range legendRowOrder {
		order[label] = i
	}

	present := make(map[legendKey]bool)

	add := func(t model.UnitType, unit *model.Unit, external bool) {
		label, ok := legendEntityLabel(t)
		if !ok {
			return
		}

		present[legendKey{label: label, colour: legendRowColour(t, unit, external)}] = true
	}

	for _, entry := range v.Units {
		if entry == nil || entry.Unit == nil {
			continue
		}

		add(entry.Unit.Type, entry.Unit, entry.IsExternal)
	}

	// The boundary cluster of a C2/C3 view is styled as its unit — without
	// this the diagram's largest coloured frame goes unexplained.
	if v.ExpandedUnitModel != nil {
		add(v.ExpandedUnitModel.Type, v.ExpandedUnitModel, false)
	}

	keys := slices.Collect(maps.Keys(present))
	slices.SortFunc(keys, func(a, b legendKey) int {
		if order[a.label] != order[b.label] {
			return order[a.label] - order[b.label]
		}

		return strings.Compare(a.colour, b.colour)
	})

	entries := make([]LegendEntry, 0, len(keys))
	for _, key := range keys {
		entries = append(entries, LegendEntry{Label: key.label, Color: key.colour})
	}

	return entries
}

// legendKey identifies one legend row: the entity name plus the exact colour
// it is shown in.
type legendKey struct {
	label  string
	colour string
}

// legendRowOrder is the canonical display order of element rows: internal
// entities by level, then external ones.
//
//nolint:gochecknoglobals // fixed vocabulary, not mutable state
var legendRowOrder = []string{
	"person", "system", "system db", "system queue", "box",
	"container", "container db", "container queue",
	"component", "component db", "component queue",
	"external person", "external system", "external db", "external queue",
}

// legendEntityLabel names the entity a unit type contributes to the legend;
// ok=false for types the legend does not name. db/queue exist at every level,
// so their labels carry the level ("system db", "container db", "component
// db") — matching the model's own type vocabulary; the label is what
// disambiguates two same-kind rows rendered in different level colours.
// Grouping boxes fold into their level's entity row — they are layout, not a
// colour of their own.
func legendEntityLabel(t model.UnitType) (string, bool) {
	switch t {
	case model.TypePerson:
		return "person", true
	case model.TypeSystem:
		return "system", true
	case model.TypeDb:
		return "system db", true
	case model.TypeQueue:
		return "system queue", true
	case model.TypeBox:
		return "box", true
	case model.TypeContainer, model.TypeContainerBox:
		return "container", true
	case model.TypeContainerDb:
		return "container db", true
	case model.TypeContainerQueue:
		return "container queue", true
	case model.TypeComponent, model.TypeComponentBox:
		return "component", true
	case model.TypeComponentDb:
		return "component db", true
	case model.TypeComponentQueue:
		return "component queue", true
	case model.TypePersonExternal:
		return "external person", true
	case model.TypeSystemExternal:
		return "external system", true
	case model.TypeDbExternal:
		return "external db", true
	case model.TypeQueueExternal:
		return "external queue", true
	default:
		return "", false
	}
}

// legendRowColour returns the colour a unit's legend row is shown in: the
// level border colour, the level's external grey for external entries, or —
// for grouping boxes — the content-derived border colour GetBoxStyleByContents
// gives the box (a box with external subunits renders grey regardless of its
// own level).
func legendRowColour(t model.UnitType, unit *model.Unit, external bool) string {
	if IsBoxType(t) && HasExternalSubunits(unit) {
		return model.PersonExternalBorder
	}

	if external {
		return levelExternalColour(LevelForType(t))
	}

	return levelBorderColour(LevelForType(t))
}

// levelBorderColour returns the border colour the renderer uses for internal
// units of a C4 level (GetStyleForType).
func levelBorderColour(level int) string {
	switch level {
	case levelC1:
		return model.PersonBorder
	case levelC2:
		return model.ContainerBorder
	default:
		return model.ComponentBorder
	}
}

// levelExternalColour returns the border colour the renderer uses for
// external units of a C4 level (getExternalStyle).
func levelExternalColour(level int) string {
	switch level {
	case levelC1:
		return model.PersonExternalBorder
	case levelC2:
		return model.ContainerExternalBorder
	default:
		return model.ComponentExternalBorder
	}
}

// legendKindEntries emits a row for each link-kind colour that actually
// appears on a drawn edge. Edge colours are already resolved at this point
// (explicit colour > kind colour > source border), so the legend can never
// advertise a colour the diagram does not use.
func legendKindEntries(edges []*Edge) []LegendEntry {
	used := make(map[string]bool, len(edges))
	for _, edge := range edges {
		if edge != nil {
			used[edge.Color] = true
		}
	}

	kinds := []struct {
		colour string
		label  string
	}{
		{model.LinkReadColour, "read"},
		{model.LinkWriteColour, "write"},
		{model.LinkReadWriteColour, "read-write"},
	}

	entries := make([]LegendEntry, 0, len(kinds))
	for _, kind := range kinds {
		if used[kind.colour] {
			entries = append(entries, LegendEntry{Label: kind.label, Color: kind.colour})
		}
	}

	return entries
}

// legendCustomEntries appends the author-defined legend lines (LEG-03) after
// the defaults. Lines without a colour render in the muted secondary grey.
func legendCustomEntries(v *view.View) []LegendEntry {
	entries := make([]LegendEntry, 0, len(v.LegendLines))
	for _, line := range v.LegendLines {
		colour := line.Color
		if colour == "" {
			colour = model.ArrowColor
		}

		entries = append(entries, LegendEntry{Label: line.Label, Color: colour})
	}

	return entries
}

// luminance estimates the relative luminance of a #RRGGBB colour (0 = dark,
// 1 = light). Unparseable values (named colours, "") are treated as dark so
// the font-colour fallback keeps the level default.
func luminance(hex string) float64 {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return 0
	}

	val, err := strconv.ParseUint(hex, 16, 32)
	if err != nil {
		return 0
	}

	r := float64((val>>16)&0xFF) / 255
	g := float64((val>>8)&0xFF) / 255
	b := float64(val&0xFF) / 255

	return 0.299*r + 0.587*g + 0.114*b
}

// buildCluster creates a cluster for an expanded unit. CTX-03: children that
// themselves have subunits recurse into nested clusters (buildNestedCluster) so
// the nesting picture inside an expanded cluster matches the drill-down views;
// leaf children render as plain nodes. The caller keeps the D-07
// expanded-but-empty guard — an expanded unit with zero subunits never reaches
// this function.
func buildCluster(entry *view.Entry, v *view.View) *Cluster {
	// Boxes use content-based styling (border colour derived from subunits)
	var style *NodeStyle
	if IsBoxType(entry.Unit.Type) {
		style = GetBoxStyleByContents(entry.Unit)
	} else {
		style = GetStyleForType(entry.Unit.Type, entry.IsExternal)
	}

	applyUnitOverrides(style, entry.Unit)

	cluster := &Cluster{
		ID:         entry.FullPath,
		Label:      buildClusterLabel(entry),
		Nodes:      make([]*Node, 0),
		Clusters:   make([]*Cluster, 0),
		Style:      style,
		Type:       entry.Unit.Type,
		IsExternal: entry.IsExternal,
	}

	// Process children in definition order (use SubunitOrder if available),
	// dispatching subunit-containers to nested clusters (CTX-03) and leaves to
	// nodes.
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

		childEntry, exists := v.Units[childPath]
		if !exists {
			// Create entry if not in view (shouldn't happen, but be defensive),
			// preserving the expansion hint from the parent's Expanded list.
			childEntry = &view.Entry{
				Unit:        childUnit,
				FullPath:    childPath,
				IsExpanded:  isUnitExpanded(entry.Unit, childName),
				HasSubunits: len(childUnit.Subunits) > 0,
				IsExternal:  view.IsExternalType(childUnit.Type),
			}
		}

		if childEntry.HasSubunits {
			// Recursively build nested cluster for subunit-containers (CTX-03)
			nestedCluster := buildNestedCluster(childEntry, childPath, v)
			cluster.Clusters = append(cluster.Clusters, nestedCluster)
		} else {
			// Build node for leaf child
			node := buildNode(childEntry)
			node.IsInCluster = true
			cluster.Nodes = append(cluster.Nodes, node)
		}
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

	// CTX-03: collapsed containers rendered as (nested) clusters keep the same
	// 🔍 affordance buildNode gives collapsed units with subunits. Author-
	// expanded entries — including every entry of an all-expanded view, where
	// IsExpanded mirrors HasSubunits — stay 🔍-free (D-04).
	if entry.HasSubunits && !entry.IsExpanded {
		label.Name += " 🔍"
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

	// Collect per-pair kind/style aggregates (AGG-01..03) beside the
	// multiplicity scan so collapsed pairs keep their kind identity.
	pairAggs := collectPairAggregates(v)

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
		outEdges := processOutgoingLinks(v, path, outLinks, seen, pairCounts, pairAggs)
		edges = append(edges, outEdges...)

		// Process incoming links (linkFrom)
		inEdges := processIncomingLinks(v, path, inLinks, seen, pairCounts, pairAggs)
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

// pairAggregate collects the per-link kind/style/colour inputs of one
// (source, target) pair so collapsed edges can derive a shared identity
// (AGG-01..03).
type pairAggregate struct {
	kinds          []model.LinkKind
	styles         []string // effective styles (unstyled links default to solid)
	anyExplicitClr bool     // any constituent sets an explicit link.Color
	anyKindUnset   bool     // any constituent carries no kind (BC-safe fallback)
}

// add records one constituent link.
func (a *pairAggregate) add(link model.Link) {
	a.kinds = append(a.kinds, link.Kind)

	if link.Kind == "" {
		a.anyKindUnset = true
	}

	style := link.Style
	if style == "" {
		style = "solid"
	}

	a.styles = append(a.styles, style)

	if link.Color != "" {
		a.anyExplicitClr = true
	}
}

// kindColourFor returns the kind-derived colour for a collapsed pair and
// whether kind colouring applies at all. It does not apply when: the pair has
// a single constituent (per-edge semantics), any constituent sets an explicit
// colour (AGG-03 — the caller falls back to the D-01 source-border default),
// or any constituent lacks a kind (BC-safe, AGG-01). All-same kinds colour
// with their kind; mixed kinds colour read-write.
func (a *pairAggregate) kindColourFor() (string, bool) {
	if a == nil || a.anyExplicitClr || len(a.kinds) < 2 {
		return "", false
	}

	for _, kind := range a.kinds {
		if kind == "" {
			return "", false
		}
	}

	allSame := true

	for _, kind := range a.kinds {
		if kind != a.kinds[0] {
			allSame = false
		}
	}

	if allSame {
		return kindColour(a.kinds[0]), true
	}

	return model.LinkReadWriteColour, true
}

// sourceBorderColour returns the D-01 default edge colour for a drawn source.
func sourceBorderColour(sourceEntry *view.Entry) string {
	if sourceEntry == nil {
		return ""
	}

	return GetStyleForType(sourceEntry.Unit.Type, sourceEntry.IsExternal).BorderColor
}

// styleFor returns the collapsed pair's line style: all constituents agree →
// that style; otherwise any solid → solid; else any dashed → dashed; else
// dotted (AGG-02).
func (a *pairAggregate) styleFor(survivorStyle string) string {
	if a == nil || len(a.styles) < 2 {
		return survivorStyle
	}

	allSame := true

	for _, style := range a.styles {
		if style != a.styles[0] {
			allSame = false
		}
	}

	if allSame {
		return a.styles[0]
	}

	if slices.Contains(a.styles, "solid") {
		return "solid"
	}

	if slices.Contains(a.styles, "dashed") {
		return "dashed"
	}

	return "dotted"
}

// collectPairAggregates walks the same pre-dedup link lists as
// countPairMultiplicity and records kind/style/colour per pair. Skipped in
// expanded mode (every parallel link renders separately, COMPAT-02).
func collectPairAggregates(v *view.View) map[string]*pairAggregate {
	aggs := make(map[string]*pairAggregate)

	if v.AllExpanded {
		return aggs
	}

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
			recordPairAggregate(aggs, path+"->"+link.Peer, link)
		}

		collectIncomingPairAggregates(v, path, aggs)
	}

	return aggs
}

// collectIncomingPairAggregates records the authored incoming links of one
// unit (mirrors excluded — synthetic duplicates of outgoing links).
func collectIncomingPairAggregates(v *view.View, path string, aggs map[string]*pairAggregate) {
	entry := v.Units[path]

	inLinks := entry.Unit.LinksFrom
	if entry.ResolvedLinksFrom != nil {
		inLinks = entry.ResolvedLinksFrom
	}

	for _, link := range inLinks {
		if link.Mirror {
			continue
		}

		recordPairAggregate(aggs, link.Peer+"->"+path, link)
	}
}

// recordPairAggregate appends one constituent link to the pair's aggregate.
func recordPairAggregate(aggs map[string]*pairAggregate, key string, link model.Link) {
	agg := aggs[key]
	if agg == nil {
		agg = &pairAggregate{}
		aggs[key] = agg
	}

	agg.add(link)
}

// processOutgoingLinks processes outgoing links from a unit.
func processOutgoingLinks(
	v *view.View,
	path string,
	links []model.Link,
	seen map[string]bool,
	pairCounts map[string]int,
	pairAggs map[string]*pairAggregate,
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
			// AGG-01..03: collapsed pairs override the surviving edge's
			// colour/style from the pair aggregate.
			if !v.AllExpanded && pairCounts[edgeKey] >= 2 {
				applyCollapsedPairStyle(edge, pairAggs[edgeKey], sourceEntry)
			}

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
	pairAggs map[string]*pairAggregate,
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
			if !v.AllExpanded && pairCounts[edgeKey] >= 2 {
				applyCollapsedPairStyle(edge, pairAggs[edgeKey], sourceEntry)
			}

			edges = append(edges, edge)
		}
	}

	return edges
}

// applyCollapsedPairStyle overrides a surviving edge's colour/style from its
// pair aggregate (AGG-01..03) — called on collapsed pairs (2+ constituents)
// in resolved views only. Kind-derived colour wins (all-same kinds colour
// with the kind, mixed colour read-write); an explicit colour on any
// constituent or an unset kind falls back to the D-01 source-border default
// of the drawn edge.
func applyCollapsedPairStyle(
	edge *Edge,
	agg *pairAggregate,
	sourceEntry *view.Entry,
) {
	if colour, ok := agg.kindColourFor(); ok {
		edge.Color = colour
	} else if agg.anyExplicitClr || agg.anyKindUnset {
		edge.Color = sourceBorderColour(sourceEntry)
	}

	edge.Style = agg.styleFor(edge.Style)
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

	// rank="reverse" flips the layout ranking (RANK-01): endpoints swap at
	// emission, Source/Target stay logical
	edge.RankReverse = link.Rank == model.RankReverse

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

	// Add explore URLs to collapsed nodes with subunits — at the top level AND
	// inside clusters (CTX-03: the walk recurses into cluster.Clusters so
	// nested container clusters keep their drill affordance too).
	assignExploreURLs(g.Nodes, v, currentPath, basename, format)
	assignExploreURLsToClusters(g.Clusters, v, currentPath, basename, format)

	// Add navigation for C2/C3 views
	if v.Level != view.LevelC1 && v.ExpandedUnit != "" {
		g.Navigation = buildNavigation(v, basename, format)
	}

	return g
}

// assignExploreURLs assigns an ExploreURL to every collapsed container node in
// the list. Helper wrapping the shouldHaveExploreLink + ComputeExploreURL body
// that previously lived in BuildGraphWithPath's one-level loops, so cluster
// walks can reuse it at every nesting depth.
func assignExploreURLs(nodes []*Node, v *view.View, currentPath, basename, format string) {
	for _, node := range nodes {
		if shouldHaveExploreLink(node, v) {
			node.ExploreURL = ComputeExploreURL(currentPath, node.ID, basename, format)
		}
	}
}

// assignExploreURLsToClusters recursively walks the cluster tree: each
// collapsed container cluster gets its own ExploreURL (same predicate nodes
// use, keyed on the cluster's unit entry), the cluster's nodes are processed
// via assignExploreURLs, and nested clusters are visited (CTX-03).
func assignExploreURLsToClusters(clusters []*Cluster, v *view.View, currentPath, basename, format string) {
	for _, cluster := range clusters {
		// Clusters are expanded units by construction; a collapsed container
		// rendered as a nested cluster still drills down via its URL attribute.
		if entry, exists := v.Units[cluster.ID]; exists && entryShouldHaveExploreLink(entry) {
			cluster.ExploreURL = ComputeExploreURL(currentPath, cluster.ID, basename, format)
		}

		assignExploreURLs(cluster.Nodes, v, currentPath, basename, format)
		assignExploreURLsToClusters(cluster.Clusters, v, currentPath, basename, format)
	}
}

// shouldHaveExploreLink determines if a node should have an explore link.
// Only systems and boxes that have subunits and are collapsed get explore links.
func shouldHaveExploreLink(node *Node, v *view.View) bool {
	entry, exists := v.Units[node.ID]
	if !exists {
		return false
	}

	return entryShouldHaveExploreLink(entry)
}

// entryShouldHaveExploreLink reports whether a view entry is a collapsed
// container-capable unit that deserves a drill-down (explore) link — the
// shared predicate for nodes and nested container clusters (CTX-03).
func entryShouldHaveExploreLink(entry *view.Entry) bool {
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
