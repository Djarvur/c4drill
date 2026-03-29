package view

import (
	"slices"
	"strings"

	"github.com/Djarvur/c4drill/internal/model"
	"github.com/Djarvur/c4drill/internal/parser"
)

// GenerateExpandedView creates a view containing ALL units in the model at all nesting levels.
// This is used for the --expanded mode that shows the complete hierarchy in a single diagram.
// It includes external boundary nodes for units referenced by links but not defined.
func GenerateExpandedView(m *parser.Model) *View {
	if m == nil {
		return nil
	}

	v := &View{
		Level:     LevelC1,
		Title:     m.Properties.Name,
		Edges:     m.Properties.Edges,
		UnitOrder: make([]string, 0),
		Units:     make(map[string]*Entry),
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

	// Add external boundary nodes for referenced units not in the model
	addExternalBoundaryNodes(v, m)

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
		Level:     LevelC1,
		Title:     m.Properties.Name,
		Edges:     m.Properties.Edges,
		UnitOrder: make([]string, 0),
		Units:     make(map[string]*Entry),
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
			IsExpanded:  isUnitExpanded(unit, name),
			HasSubunits: len(unit.Subunits) > 0,
			IsExternal:  IsExternalType(unit.Type),
		}
		v.UnitOrder = append(v.UnitOrder, name)
	}

	// Add external boundary nodes for referenced units not in the model
	addExternalBoundaryNodes(v, m)

	return v
}

// isUnitExpanded checks if a unit should be expanded based on its Expanded list.
// Per the design decision: only per-unit expanded attribute is used (no global default).
func isUnitExpanded(unit *model.Unit, unitPath string) bool {
	return slices.Contains(unit.Expanded, unitPath)
}

// addExternalBoundaryNodes scans all links (including nested subunits) and adds
// boundary nodes for referenced units that are not in the current view.
func addExternalBoundaryNodes(v *View, m *parser.Model) {
	// Determine iteration order: use UnitOrder if available, otherwise fallback to map keys
	var unitOrder []string
	if len(m.UnitOrder) > 0 {
		unitOrder = m.UnitOrder
	} else {
		for name := range m.Units {
			unitOrder = append(unitOrder, name)
		}
	}

	// Iterate in definition order
	for _, name := range unitOrder {
		unit := m.Units[name]
		if unit == nil {
			continue
		}

		addExternalBoundaryNodesRecursive(v, m, name, unit)
	}
}

// addExternalBoundaryNodesRecursive recursively scans a unit and its subunits for links
// and adds boundary nodes for referenced units not in the view.
func addExternalBoundaryNodesRecursive(v *View, m *parser.Model, path string, unit *model.Unit) {
	// Check outgoing links
	for _, link := range unit.Links {
		if _, exists := v.Units[link.Peer]; !exists {
			// Create external boundary node, preserving original unit data if it exists in model
			v.Units[link.Peer] = createExternalBoundaryNode(m, link.Peer, path)
			v.UnitOrder = append(v.UnitOrder, link.Peer) // Append at end
		}
	}

	// Check incoming links (LinksFrom)
	for _, link := range unit.LinksFrom {
		if _, exists := v.Units[link.Peer]; !exists {
			// Create external boundary node, preserving original unit data if it exists in model
			v.Units[link.Peer] = createExternalBoundaryNode(m, link.Peer, path)
			v.UnitOrder = append(v.UnitOrder, link.Peer) // Append at end
		}
	}

	// Recursively check subunits in definition order (with fallback for test models)
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

		addExternalBoundaryNodesRecursive(v, m, path+"."+subName, subUnit)
	}
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
		Level:        LevelC2,
		Title:        systemUnit.Name + " - Containers",
		Edges:        systemUnit.Edges,
		Parent:       systemPath,
		ExpandedUnit: systemPath,
		UnitOrder:    make([]string, 0),
		Units:        make(map[string]*Entry),
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

	// Add external boundary nodes for links from subunits
	addExternalBoundaryNodesForSubunits(v, m, systemUnit.Subunits, subunitOrder, systemPath)

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
		Level:        LevelC3,
		Title:        title,
		Edges:        containerUnit.Edges,
		Parent:       parentPath,
		ExpandedUnit: containerPath,
		UnitOrder:    make([]string, 0),
		Units:        make(map[string]*Entry),
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

// addExternalBoundaryNodesForSubunits scans links from subunits and adds
// boundary nodes for referenced units that are not in the current view.
func addExternalBoundaryNodesForSubunits(v *View, m *parser.Model, subunits map[string]*model.Unit, subunitOrder []string, parentPath string) {
	// Iterate in definition order
	for _, name := range subunitOrder {
		unit := subunits[name]
		if unit == nil {
			continue
		}

		// Check outgoing links
		for _, link := range unit.Links {
			if _, exists := v.Units[link.Peer]; !exists {
				// Create external boundary node, preserving original unit data if it exists in model
				v.Units[link.Peer] = createExternalBoundaryNode(m, link.Peer, parentPath)
				v.UnitOrder = append(v.UnitOrder, link.Peer) // Append at end
			}
		}

		// Check incoming links (LinksFrom)
		for _, link := range unit.LinksFrom {
			if _, exists := v.Units[link.Peer]; !exists {
				// Create external boundary node, preserving original unit data if it exists in model
				v.Units[link.Peer] = createExternalBoundaryNode(m, link.Peer, parentPath)
				v.UnitOrder = append(v.UnitOrder, link.Peer) // Append at end
			}
		}
	}
}
