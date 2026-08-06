package view

import (
	"slices"
	"strings"

	"github.com/Djarvur/c4drill/internal/model"
	"github.com/Djarvur/c4drill/internal/parser"
)

// GenerateExpandedView creates a view containing ALL units in the model at all nesting levels.
// This is used for the --expanded mode that shows the complete hierarchy in a single diagram.
func GenerateExpandedView(m *parser.Model) *View {
	if m == nil {
		return nil
	}

	v := &View{
		Level:       LevelC1,
		Title:       m.Properties.Name,
		Edges:       m.Properties.Edges,
		AllExpanded: true,
		UnitOrder:   make([]string, 0),
		Units:       make(map[string]*Entry),
	}

	// Determine iteration order: use UnitOrder if available, otherwise fallback to map keys
	var unitOrder []string
	if len(m.UnitOrder) > 0 {
		unitOrder = m.UnitOrder
	} else {
		// Fallback for test models or models without explicit order
		for name := range m.Units {
			unitOrder = append(unitOrder, name)
		}
	}

	// Recursively add all units depth-first in definition order
	for _, name := range unitOrder {
		unit := m.Units[name]
		if unit == nil {
			continue
		}

		addUnitRecursive(v, name, unit)
	}

	return v
}

// addUnitRecursive adds a unit and all its subunits to the view recursively.
func addUnitRecursive(v *View, path string, unit *model.Unit) {
	v.Units[path] = &Entry{
		Unit:        unit,
		FullPath:    path,
		IsExpanded:  len(unit.Subunits) > 0, // Always show as expanded if has subunits
		HasSubunits: len(unit.Subunits) > 0,
		IsExternal:  IsExternalType(unit.Type),
	}
	v.UnitOrder = append(v.UnitOrder, path)

	// Determine iteration order: use SubunitOrder if available, otherwise fallback to map keys
	var subunitOrder []string
	if len(unit.SubunitOrder) > 0 {
		subunitOrder = unit.SubunitOrder
	} else {
		for name := range unit.Subunits {
			subunitOrder = append(subunitOrder, name)
		}
	}

	// Recursively add subunits in definition order
	for _, subName := range subunitOrder {
		subUnit := unit.Subunits[subName]
		if subUnit == nil {
			continue
		}

		addUnitRecursive(v, path+"."+subName, subUnit)
	}
}

// GenerateC1View creates a C1 (Context) level view showing all top-level units.
// It includes external boundary nodes for units referenced by links but not defined.
func GenerateC1View(m *parser.Model) *View {
	if m == nil {
		return nil
	}

	v := &View{
		Level:        LevelC1,
		Title:        m.Properties.Name,
		Edges:        m.Properties.Edges,
		UnitOrder:    make([]string, 0),
		Units:        make(map[string]*Entry),
		VisiblePaths: make(map[string]bool),
	}

	// Determine iteration order: use UnitOrder if available, otherwise fallback to map keys
	var unitOrder []string
	if len(m.UnitOrder) > 0 {
		unitOrder = m.UnitOrder
	} else {
		// Fallback for test models or models without explicit order
		for name := range m.Units {
			unitOrder = append(unitOrder, name)
		}
	}

	// Add all top-level units in definition order
	for _, name := range unitOrder {
		unit := m.Units[name]
		if unit == nil {
			continue
		}

		v.Units[name] = &Entry{
			Unit:        unit,
			FullPath:    name,
			IsExpanded:  isExpandedInC1(m, unit, name),
			HasSubunits: len(unit.Subunits) > 0,
			IsExternal:  IsExternalType(unit.Type),
		}
		v.UnitOrder = append(v.UnitOrder, name)
	}

	// Populate visible subunits: direct subunits of expanded top-level units
	// are rendered as nodes INSIDE the parent cluster (buildCluster renders
	// one level). They are added to v.Units + VisiblePaths so resolution
	// (D-07/D-09/D-10) can reach them, while BuildGraph skips them as
	// top-level nodes (Pitfall 5).
	for _, name := range unitOrder {
		entry := v.Units[name]
		if entry == nil || !entry.IsExpanded {
			continue
		}

		// Determine iteration order: use SubunitOrder if available, otherwise fallback to map keys
		var subunitOrder []string
		if len(entry.Unit.SubunitOrder) > 0 {
			subunitOrder = entry.Unit.SubunitOrder
		} else {
			for subName := range entry.Unit.Subunits {
				subunitOrder = append(subunitOrder, subName)
			}
		}

		for _, subName := range subunitOrder {
			subUnit := entry.Unit.Subunits[subName]
			if subUnit == nil {
				continue
			}

			fullPath := name + "." + subName
			v.Units[fullPath] = &Entry{
				Unit:        subUnit,
				FullPath:    fullPath,
				IsExpanded:  isUnitExpanded(entry.Unit, subName),
				HasSubunits: len(subUnit.Subunits) > 0,
				IsExternal:  IsExternalType(subUnit.Type),
			}
			v.UnitOrder = append(v.UnitOrder, fullPath)
			v.VisiblePaths[fullPath] = true
		}
	}

	// Add boundary nodes: scan ALL links (including nested subunits) but resolve
	// peer targets to their nearest top-level ancestor visible in the view.
	// This prevents C1 from being polluted with deeply nested subunit boundary nodes.
	addC1BoundaryNodes(v, m)

	return v
}

// isUnitExpanded checks if a unit should be expanded based on its Expanded list.
func isUnitExpanded(unit *model.Unit, unitPath string) bool {
	return slices.Contains(unit.Expanded, unitPath)
}

// isExpandedInC1 checks if a top-level unit should be shown as expanded in C1 view.
// It checks both properties.expanded (global) and per-unit expanded (self-referencing).
func isExpandedInC1(m *parser.Model, unit *model.Unit, unitPath string) bool {
	if slices.Contains(m.Properties.Expanded, unitPath) {
		return true
	}

	return slices.Contains(unit.Expanded, unitPath)
}

// createExternalBoundaryNode creates an Entry representing an external boundary node.
// If the unit exists in the model (e.g., a nested subunit), it uses the original unit data
// to preserve attributes like links with length. Otherwise, it creates a minimal placeholder.
func createExternalBoundaryNode(m *parser.Model, name string, _ string) *Entry {
	// Try to find the actual unit in the model
	actualUnit := findUnitByPath(m, name)
	if actualUnit != nil {
		// Use the actual unit data to preserve links and other attributes
		return &Entry{
			Unit:        actualUnit,
			FullPath:    name,
			IsExpanded:  false,
			HasSubunits: len(actualUnit.Subunits) > 0,
			IsExternal:  IsExternalType(actualUnit.Type),
		}
	}

	// Default to external system type for boundary nodes that don't exist in model
	return &Entry{
		Unit: &model.Unit{
			Type: model.TypeSystemExternal,
			Name: name,
		},
		FullPath:    name,
		IsExpanded:  false,
		HasSubunits: false,
		IsExternal:  true,
	}
}

// addC1BoundaryNodes scans ALL links in the model (including deeply nested subunits)
// but resolves peer targets to their nearest visible top-level ancestor.
// This prevents C1 from being polluted with ~100 boundary nodes for nested subunits.
// For example, a link from linuxSystem.sshAuth.sshd → linuxSystem.sshAuth.nss is
// internal to linuxSystem and creates no boundary node. A link from
// webUser → linuxSystem.localIDP.grpcAPIs.authAPI resolves to webUser → linuxSystem.
func addC1BoundaryNodes(v *View, m *parser.Model) {
	// Collect all link peers from ALL units (recursive) and resolve them
	var unitOrder []string
	if len(m.UnitOrder) > 0 {
		unitOrder = m.UnitOrder
	} else {
		for name := range m.Units {
			unitOrder = append(unitOrder, name)
		}
	}

	for _, name := range unitOrder {
		unit := m.Units[name]
		if unit == nil {
			continue
		}

		resolveAndAddBoundary(v, m, name, unit)
	}
}

// resolveAndAddBoundary recursively scans a unit and all its subunits for links.
// For each link, it resolves the peer to the nearest visible top-level ancestor in the view.
// If the resolved peer isn't in the view, it creates a boundary node.
// Resolved links are collected on the top-level source entry's ResolvedLinks/ResolvedLinksFrom.
func resolveAndAddBoundary(v *View, m *parser.Model, path string, unit *model.Unit) {
	// Find the deepest VISIBLE ancestor of this path (source for resolved
	// links) — a visible subunit entry inside an expanded cluster when the
	// source is one of its descendants, the top-level unit otherwise (D-09).
	resolvedSource := resolveToViewAncestor(v, path)

	sourceEntry := v.Units[resolvedSource]

	// Process outgoing links
	for _, link := range unit.Links {
		resolved := resolveToTopLevel(v, link.Peer)
		if resolved == "" {
			continue // Internal link (both sides under same ancestor or self-referencing)
		}

		if _, exists := v.Units[resolved]; !exists {
			v.Units[resolved] = createExternalBoundaryNode(m, resolved, path)
			v.UnitOrder = append(v.UnitOrder, resolved)
		}

		// Add resolved outgoing link to the source entry (D-10: within-cluster
		// edges pass because resolved != resolvedSource; D-08: no parent edge
		// is synthesized when the link resolved to a child).
		if sourceEntry != nil && resolved != resolvedSource {
			// D-13: minlen applies only when both drawn endpoints are the
			// link's original units — resolved links carry no length.
			length := 0
			if path == resolvedSource && link.Peer == resolved {
				length = link.Length
			}

			resolvedLink := model.Link{
				Peer:          resolved,
				Technology:    link.Technology,
				Description:   link.Description,
				Style:         link.Style,
				Arrow:         link.Arrow,
				Rank:          link.Rank,
				LabelPosition: link.LabelPosition,
				Color:         link.Color,
				Length:        length,
			}
			sourceEntry.ResolvedLinks = append(sourceEntry.ResolvedLinks, resolvedLink)
		}
	}

	// Process incoming links
	for _, link := range unit.LinksFrom {
		resolved := resolveToTopLevel(v, link.Peer)
		if resolved == "" {
			continue
		}

		if _, exists := v.Units[resolved]; !exists {
			v.Units[resolved] = createExternalBoundaryNode(m, resolved, path)
			v.UnitOrder = append(v.UnitOrder, resolved)
		}

		// Add resolved incoming link to the source entry (D-10/D-08 semantics
		// mirror the outgoing pass above).
		if sourceEntry != nil && resolved != resolvedSource {
			// D-13: minlen applies only when both drawn endpoints are the
			// link's original units — resolved links carry no length.
			length := 0
			if path == resolvedSource && link.Peer == resolved {
				length = link.Length
			}

			resolvedLink := model.Link{
				Peer:          resolved,
				Technology:    link.Technology,
				Description:   link.Description,
				Style:         link.Style,
				Arrow:         link.Arrow,
				Rank:          link.Rank,
				LabelPosition: link.LabelPosition,
				Color:         link.Color,
				Length:        length,
				Mirror:        link.Mirror,
			}
			sourceEntry.ResolvedLinksFrom = append(sourceEntry.ResolvedLinksFrom, resolvedLink)
		}
	}

	// Recurse into subunits
	var subunitOrder []string
	if len(unit.SubunitOrder) > 0 {
		subunitOrder = unit.SubunitOrder
	} else {
		for name := range unit.Subunits {
			subunitOrder = append(subunitOrder, name)
		}
	}

	for _, subName := range subunitOrder {
		subUnit := unit.Subunits[subName]
		if subUnit == nil {
			continue
		}

		resolveAndAddBoundary(v, m, path+"."+subName, subUnit)
	}
}

// resolveToTopLevel resolves a peer path to its nearest VISIBLE ancestor in
// the view — a top-level unit, or a visible subunit inside an expanded cluster
// (D-07). Truly-external peers with no ancestor in the view are returned
// as-is so they become boundary nodes (unchanged fallback behavior).
func resolveToTopLevel(v *View, peer string) string {
	resolved := resolveToViewAncestor(v, peer)
	if resolved != "" {
		return resolved
	}

	return peer
}

// GenerateC2View creates a C2 (Container) level view for an expanded system.
// It shows the subunits (containers) of the specified system.
func GenerateC2View(m *parser.Model, systemPath string) *View {
	if m == nil {
		return nil
	}

	// Find the expanded system unit
	systemUnit := findUnitByPath(m, systemPath)
	if systemUnit == nil {
		return nil
	}

	v := &View{
		Level:             LevelC2,
		Title:             systemUnit.Name + " - Containers",
		RootTitle:         m.Properties.Name,
		Edges:             systemUnit.Edges,
		Parent:            systemPath,
		ExpandedUnit:      systemPath,
		ExpandedUnitModel: systemUnit,
		UnitOrder:         make([]string, 0),
		Units:             make(map[string]*Entry),
		AncestorNames:     buildAncestorNames(m, systemPath),
	}

	// Determine iteration order: use SubunitOrder if available, otherwise fallback to map keys
	var subunitOrder []string
	if len(systemUnit.SubunitOrder) > 0 {
		subunitOrder = systemUnit.SubunitOrder
	} else {
		for name := range systemUnit.Subunits {
			subunitOrder = append(subunitOrder, name)
		}
	}

	// Add subunits (containers) of the expanded system in definition order
	for _, name := range subunitOrder {
		unit := systemUnit.Subunits[name]
		if unit == nil {
			continue
		}

		fullPath := systemPath + "." + name
		v.Units[fullPath] = &Entry{
			Unit:        unit,
			FullPath:    fullPath,
			IsExpanded:  isUnitExpanded(systemUnit, name),
			HasSubunits: len(unit.Subunits) > 0,
			IsExternal:  IsExternalType(unit.Type),
		}
		v.UnitOrder = append(v.UnitOrder, fullPath)
	}

	// Add external boundary nodes for links from subunits (recursively)
	addExternalBoundaryNodesForSubunits(v, m, systemUnit.Subunits, subunitOrder, systemPath)

	// Resolve links for boundary nodes (external entities have deep links that need resolving)
	resolveBoundaryNodeLinks(v)

	// Resolve cross-subunit links: when a descendant of subunit A links to a
	// descendant of subunit B (or to an external boundary node), create a
	// resolved link A -> B on subunit A's entry. Without this, edges between
	// sibling subunits are lost because the peer path is too deep for isTargetInView.
	resolveSubunitCrossLinks(v, m, systemUnit.Subunits, subunitOrder, systemPath)

	return v
}

// GenerateC3View creates a C3 (Component) level view for an expanded container.
// It shows the subunits (components) of the specified container.
func GenerateC3View(m *parser.Model, containerPath string) *View {
	if m == nil {
		return nil
	}

	// Find the expanded container unit
	containerUnit := findUnitByPath(m, containerPath)
	if containerUnit == nil {
		return nil
	}

	// Extract parent path for title
	parentPath := ""
	title := containerUnit.Name + " - Components"

	if idx := strings.LastIndex(containerPath, "."); idx > 0 {
		parentPath = containerPath[:idx]

		parentUnit := findUnitByPath(m, parentPath)
		if parentUnit != nil {
			title = parentUnit.Name + " - " + containerUnit.Name + " - Components"
		}
	}

	v := &View{
		Level:             LevelC3,
		Title:             title,
		RootTitle:         m.Properties.Name,
		Edges:             containerUnit.Edges,
		Parent:            parentPath,
		ExpandedUnit:      containerPath,
		ExpandedUnitModel: containerUnit,
		UnitOrder:         make([]string, 0),
		Units:             make(map[string]*Entry),
		AncestorNames:     buildAncestorNames(m, containerPath),
	}

	// Determine iteration order: use SubunitOrder if available, otherwise fallback to map keys
	var subunitOrder []string
	if len(containerUnit.SubunitOrder) > 0 {
		subunitOrder = containerUnit.SubunitOrder
	} else {
		for name := range containerUnit.Subunits {
			subunitOrder = append(subunitOrder, name)
		}
	}

	// Add subunits (components) of the expanded container in definition order
	for _, name := range subunitOrder {
		unit := containerUnit.Subunits[name]
		if unit == nil {
			continue
		}

		fullPath := containerPath + "." + name
		v.Units[fullPath] = &Entry{
			Unit:        unit,
			FullPath:    fullPath,
			IsExpanded:  isUnitExpanded(containerUnit, name),
			HasSubunits: len(unit.Subunits) > 0,
			IsExternal:  IsExternalType(unit.Type),
		}
		v.UnitOrder = append(v.UnitOrder, fullPath)
	}

	// Add external boundary nodes for links from subunits
	addExternalBoundaryNodesForSubunits(v, m, containerUnit.Subunits, subunitOrder, containerPath)

	// Resolve links for boundary nodes (external entities have deep links that need resolving)
	resolveBoundaryNodeLinks(v)

	// Resolve cross-subunit links: when a descendant of one component links to
	// a descendant of another component (or to an external boundary node), create
	// a resolved link between the components.
	resolveSubunitCrossLinks(v, m, containerUnit.Subunits, subunitOrder, containerPath)

	return v
}

// findUnitByPath traverses the model to find a unit by its dotted path.
// For example, "mainapp.api" returns m.Units["mainapp"].Subunits["api"].
func findUnitByPath(m *parser.Model, path string) *model.Unit {
	if path == "" {
		return nil
	}

	parts := strings.Split(path, ".")
	if len(parts) == 0 {
		return nil
	}

	// Start with top-level unit
	unit, exists := m.Units[parts[0]]
	if !exists {
		return nil
	}

	// Traverse subunits
	for i := 1; i < len(parts); i++ {
		if unit.Subunits == nil {
			return nil
		}

		subunit, exists := unit.Subunits[parts[i]]
		if !exists {
			return nil
		}

		unit = subunit
	}

	return unit
}

// buildAncestorNames returns a map from each prefix of dottedPath (including
// the full path itself) to the unit's display Name. For example, path
// "mainapp.api.v2" yields {"mainapp": <name>, "mainapp.api": <name>,
// "mainapp.api.v2": <name>}. If a unit is not found at some prefix, that
// prefix is omitted (the caller falls back to the raw segment elsewhere).
// The root context title is NOT included here; the graph builder prepends it.
func buildAncestorNames(m *parser.Model, dottedPath string) map[string]string {
	if dottedPath == "" {
		return nil
	}

	parts := strings.Split(dottedPath, ".")
	names := make(map[string]string, len(parts))

	for i := 1; i <= len(parts); i++ {
		prefix := strings.Join(parts[:i], ".")
		if u := findUnitByPath(m, prefix); u != nil && u.Name != "" {
			names[prefix] = u.Name
		}
	}

	return names
}

// addExternalBoundaryNodesForSubunits scans links from subunits (recursively) and adds
// boundary nodes for referenced units that are outside the current view's scope.
// It resolves link peers to their nearest visible ancestor in the view, so deeply
// nested links (e.g., linuxSystem.storages.localStorage.lookupAPI) are resolved
// to the nearest visible parent (e.g., linuxSystem.storages) rather than appearing
// as individual boundary nodes.
func addExternalBoundaryNodesForSubunits(v *View, m *parser.Model, subunits map[string]*model.Unit, subunitOrder []string, parentPath string) {
	// Iterate in definition order, recursing into nested subunits
	for _, name := range subunitOrder {
		unit := subunits[name]
		if unit == nil {
			continue
		}

		fullPath := parentPath + "." + name

		// Check outgoing links
		for _, link := range unit.Links {
			addResolvedBoundaryNode(v, m, link.Peer, parentPath)
		}

		// Check incoming links (LinksFrom)
		for _, link := range unit.LinksFrom {
			addResolvedBoundaryNode(v, m, link.Peer, parentPath)
		}

		// Recurse into nested subunits
		if len(unit.Subunits) > 0 {
			var childOrder []string
			if len(unit.SubunitOrder) > 0 {
				childOrder = unit.SubunitOrder
			} else {
				for childName := range unit.Subunits {
					childOrder = append(childOrder, childName)
				}
			}

			addExternalBoundaryNodesForSubunits(v, m, unit.Subunits, childOrder, fullPath)
		}
	}
}

// addResolvedBoundaryNode resolves a link peer to its nearest visible ancestor
// in the view and adds it as a boundary node if it's outside the view's scope.
// If the peer is already in the view, nothing happens.
// If the peer is a nested path whose ancestor is in the view, nothing happens
// (the ancestor is already visible).
// If the peer is completely outside the scope, it's added as a boundary node.
func addResolvedBoundaryNode(v *View, m *parser.Model, peer, scopePath string) {
	// If already in view, nothing to do
	if _, exists := v.Units[peer]; exists {
		return
	}

	// Walk up the peer's path to find the nearest visible ancestor
	for {
		idx := strings.LastIndex(peer, ".")
		if idx <= 0 {
			break
		}

		peer = peer[:idx]

		if _, exists := v.Units[peer]; exists {
			// An ancestor is in the view — peer is an internal nested reference
			return
		}
	}

	// Peer has no ancestor in the view — it's external, add as boundary node
	if _, exists := v.Units[peer]; !exists {
		v.Units[peer] = createExternalBoundaryNode(m, peer, scopePath)
		v.UnitOrder = append(v.UnitOrder, peer)
	}
}

// resolveBoundaryNodeLinks resolves links for external boundary nodes so that
// edges connect to the nearest visible ancestor in the view.
// For example, if webUser links to linuxSystem.localIDP.grpcAPIs.authAPI but the
// view only shows linuxSystem.localIDP, the link is resolved to
// webUser -> linuxSystem.localIDP.
func resolveBoundaryNodeLinks(v *View) {
	for _, path := range v.UnitOrder {
		entry := v.Units[path]
		if entry == nil || !entry.IsExternal {
			continue
		}

		if len(entry.Unit.Links) == 0 && len(entry.Unit.LinksFrom) == 0 {
			continue
		}

			// Resolve outgoing links
			if len(entry.Unit.Links) > 0 {
				resolved := make([]model.Link, 0, len(entry.Unit.Links))
				for _, link := range entry.Unit.Links {
					resolvedPeer := resolveToViewAncestor(v, link.Peer)
					if resolvedPeer != "" && resolvedPeer != path {
						// D-13: the boundary node itself is the original
						// source — minlen survives only when the peer did not
						// resolve to an ancestor.
						length := 0
						if link.Peer == resolvedPeer {
							length = link.Length
						}

						resolved = append(resolved, model.Link{
							Peer:          resolvedPeer,
							Arrow:         link.Arrow,
							Rank:          link.Rank,
							Color:         link.Color,
							Style:         link.Style,
							Technology:    link.Technology,
							Description:   link.Description,
							LabelPosition: link.LabelPosition,
							Length:        length,
						})
					}
				}

				if len(resolved) > 0 {
					entry.ResolvedLinks = resolved
				}
			}

			// Resolve incoming links (LinksFrom)
			if len(entry.Unit.LinksFrom) > 0 {
				resolved := make([]model.Link, 0, len(entry.Unit.LinksFrom))
				for _, link := range entry.Unit.LinksFrom {
					resolvedPeer := resolveToViewAncestor(v, link.Peer)
					if resolvedPeer != "" && resolvedPeer != path {
						// D-13: minlen survives only when the peer did not
						// resolve to an ancestor.
						length := 0
						if link.Peer == resolvedPeer {
							length = link.Length
						}

						resolved = append(resolved, model.Link{
							Peer:          resolvedPeer,
							Arrow:         link.Arrow,
							Rank:          link.Rank,
							Color:         link.Color,
							Style:         link.Style,
							Technology:    link.Technology,
							Description:   link.Description,
							LabelPosition: link.LabelPosition,
							Length:        length,
							Mirror:        link.Mirror,
						})
					}
				}

				if len(resolved) > 0 {
					entry.ResolvedLinksFrom = resolved
				}
			}
	}
}

// resolveToViewAncestor resolves a peer path to its nearest ancestor
// (or itself) that is present in the view.
// Returns "" if no ancestor is found in the view.
func resolveToViewAncestor(v *View, peer string) string {
	// Check if peer itself is in the view
	if _, exists := v.Units[peer]; exists {
		return peer
	}

	// Walk up the peer's path to find nearest visible ancestor
	for {
		idx := strings.LastIndex(peer, ".")
		if idx <= 0 {
			break
		}

		peer = peer[:idx]

		if _, exists := v.Units[peer]; exists {
			return peer
		}
	}

	// Check if the top-level peer (no dots) is in the view
	if _, exists := v.Units[peer]; exists {
		return peer
	}

	return ""
}

// resolveSubunitCrossLinks scans all descendants of each subunit and creates
// ResolvedLinks on the subunit entry when a descendant's link points to a
// different subunit or external boundary node in the view.
// This ensures that edges between sibling subunits (e.g., dacProxy -> authModules)
// appear in C2/C3 diagrams even though the actual links are between deeply nested
// descendants (e.g., dacProxy.Unit.Links -> authModules.otp.settingsAPI).
func resolveSubunitCrossLinks(v *View, m *parser.Model, subunits map[string]*model.Unit, subunitOrder []string, parentPath string) {
	for _, name := range subunitOrder {
		unit := subunits[name]
		if unit == nil {
			continue
		}

		fullPath := parentPath + "." + name
		subunitEntry := v.Units[fullPath]
		if subunitEntry == nil {
			continue
		}

		// Process direct links on the subunit itself
		for _, link := range unit.Links {
			addResolvedCrossLink(v, subunitEntry, fullPath, fullPath, link)
		}

		// Process direct incoming links on the subunit itself
		for _, link := range unit.LinksFrom {
			addResolvedCrossLinkFrom(v, subunitEntry, fullPath, fullPath, link)
		}

		// Recursively process descendant links
		if len(unit.Subunits) > 0 {
			var childOrder []string
			if len(unit.SubunitOrder) > 0 {
				childOrder = unit.SubunitOrder
			} else {
				for childName := range unit.Subunits {
					childOrder = append(childOrder, childName)
				}
			}

			resolveDescendantCrossLinks(v, subunitEntry, fullPath, unit.Subunits, childOrder, fullPath)
		}
	}
}

// resolveDescendantCrossLinks recursively scans a subunit's descendants for links
// and adds resolved links to the subunit entry when the link target resolves to
// a different subunit or external boundary node in the view.
// entryPath is the fixed path of the subunit entry (used to detect self-links).
// parentPath is the current recursion depth (changes as we recurse deeper).
func resolveDescendantCrossLinks(v *View, subunitEntry *Entry, entryPath string, subunits map[string]*model.Unit, subunitOrder []string, parentPath string) {
	for _, name := range subunitOrder {
		unit := subunits[name]
		if unit == nil {
			continue
		}

		fullPath := parentPath + "." + name

		// Process outgoing links
		for _, link := range unit.Links {
			addResolvedCrossLink(v, subunitEntry, entryPath, fullPath, link)
		}

		// Process incoming links (LinksFrom)
		for _, link := range unit.LinksFrom {
			addResolvedCrossLinkFrom(v, subunitEntry, entryPath, fullPath, link)
		}

		// Recurse into nested subunits
		if len(unit.Subunits) > 0 {
			var childOrder []string
			if len(unit.SubunitOrder) > 0 {
				childOrder = unit.SubunitOrder
			} else {
				for childName := range unit.Subunits {
					childOrder = append(childOrder, childName)
				}
			}

			resolveDescendantCrossLinks(v, subunitEntry, entryPath, unit.Subunits, childOrder, fullPath)
		}
	}
}

// addResolvedCrossLink resolves a link peer to the nearest visible ancestor in the view
// and adds it as a ResolvedLink on the subunit entry if it points to a different entity.
// originalSource is the path of the unit that actually authored the link — minlen
// (D-13) survives only when it is also the drawn source (sourcePath).
// Every contributing link is appended without dedup (WR-01): D-05 multiplicity
// counting in buildEdges needs the full pre-dedup set, and the builder's
// pair-only markSeen performs the edge dedup (D-01 first-wins).
func addResolvedCrossLink(v *View, subunitEntry *Entry, sourcePath string, originalSource string, link model.Link) {
	resolvedPeer := resolveToViewAncestor(v, link.Peer)
	if resolvedPeer == "" || resolvedPeer == sourcePath {
		return
	}

	// D-13: minlen applies only when both drawn endpoints are the link's
	// original units — a resolved source or resolved peer drops it.
	length := 0
	if originalSource == sourcePath && link.Peer == resolvedPeer {
		length = link.Length
	}

	subunitEntry.ResolvedLinks = append(subunitEntry.ResolvedLinks, model.Link{
		Peer:          resolvedPeer,
		Technology:    link.Technology,
		Description:   link.Description,
		Style:         link.Style,
		Arrow:         link.Arrow,
		Rank:          link.Rank,
		LabelPosition: link.LabelPosition,
		Color:         link.Color,
		Length:        length,
	})
}

// addResolvedCrossLinkFrom resolves an incoming link peer and adds it as a
// ResolvedLinksFrom on the subunit entry if it points to a different entity.
// originalSource is the path of the unit that actually authored the link — minlen
// (D-13) survives only when it is also the drawn source (sourcePath).
// Every contributing link is appended without dedup (WR-01): D-05 multiplicity
// counting in buildEdges needs the full pre-dedup set, and the builder's
// pair-only markSeen performs the edge dedup (D-01 first-wins).
func addResolvedCrossLinkFrom(v *View, subunitEntry *Entry, sourcePath string, originalSource string, link model.Link) {
	resolvedPeer := resolveToViewAncestor(v, link.Peer)
	if resolvedPeer == "" || resolvedPeer == sourcePath {
		return
	}

	// D-13: minlen applies only when both drawn endpoints are the link's
	// original units — a resolved source or resolved peer drops it.
	length := 0
	if originalSource == sourcePath && link.Peer == resolvedPeer {
		length = link.Length
	}

	subunitEntry.ResolvedLinksFrom = append(subunitEntry.ResolvedLinksFrom, model.Link{
		Peer:          resolvedPeer,
		Technology:    link.Technology,
		Description:   link.Description,
		Style:         link.Style,
		Arrow:         link.Arrow,
		Rank:          link.Rank,
		LabelPosition: link.LabelPosition,
		Color:         link.Color,
		Length:        length,
		Mirror:        link.Mirror,
	})
}
