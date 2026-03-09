package view

import (
	"slices"
	"strings"

	"github.com/Djarvur/c4drill/internal/model"
	"github.com/Djarvur/c4drill/internal/parser"
)

// GenerateC1View creates a C1 (Context) level view showing all top-level units.
// It includes external boundary nodes for units referenced by links but not defined.
func GenerateC1View(m *parser.Model) *View {
	if m == nil {
		return nil
	}

	v := &View{
		Level:  LevelC1,
		Title:  m.Properties.Name,
		Edges:  m.Properties.Edges,
		Units:  make(map[string]*ViewUnit),
	}

	// Add all top-level units
	for name, unit := range m.Units {
		v.Units[name] = &ViewUnit{
			Unit:        unit,
			FullPath:    name,
			IsExpanded:  isUnitExpanded(unit, name),
			HasSubunits: len(unit.Subunits) > 0,
			IsExternal:  IsExternalType(unit.Type),
		}
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

// addExternalBoundaryNodes scans all links and adds boundary nodes for
// referenced units that are not in the current view.
func addExternalBoundaryNodes(v *View, m *parser.Model) {
	for name, unit := range m.Units {
		// Check outgoing links
		for target := range unit.Links {
			if _, exists := v.Units[target]; !exists {
				// Create external boundary node
				v.Units[target] = createExternalBoundaryNode(target, name)
			}
		}

		// Check incoming links (LinksFrom)
		for source := range unit.LinksFrom {
			if _, exists := v.Units[source]; !exists {
				// Create external boundary node
				v.Units[source] = createExternalBoundaryNode(source, name)
			}
		}
	}
}

// createExternalBoundaryNode creates a ViewUnit representing an external boundary node.
// It infers the appropriate external type based on the context.
func createExternalBoundaryNode(name string, _ string) *ViewUnit {
	// Default to external system type for boundary nodes
	return &ViewUnit{
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
		Units:        make(map[string]*ViewUnit),
	}

	// Add subunits (containers) of the expanded system
	for name, unit := range systemUnit.Subunits {
		fullPath := systemPath + "." + name
		v.Units[fullPath] = &ViewUnit{
			Unit:        unit,
			FullPath:    fullPath,
			IsExpanded:  isUnitExpanded(systemUnit, name),
			HasSubunits: len(unit.Subunits) > 0,
			IsExternal:  IsExternalType(unit.Type),
		}
	}

	// Add external boundary nodes for links from subunits
	addExternalBoundaryNodesForSubunits(v, systemUnit.Subunits, systemPath)

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
		Units:        make(map[string]*ViewUnit),
	}

	// Add subunits (components) of the expanded container
	for name, unit := range containerUnit.Subunits {
		fullPath := containerPath + "." + name
		v.Units[fullPath] = &ViewUnit{
			Unit:        unit,
			FullPath:    fullPath,
			IsExpanded:  isUnitExpanded(containerUnit, name),
			HasSubunits: len(unit.Subunits) > 0,
			IsExternal:  IsExternalType(unit.Type),
		}
	}

	// Add external boundary nodes for links from subunits
	addExternalBoundaryNodesForSubunits(v, containerUnit.Subunits, containerPath)

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
func addExternalBoundaryNodesForSubunits(v *View, subunits map[string]*model.Unit, parentPath string) {
	for _, unit := range subunits {
		// Check outgoing links
		for target := range unit.Links {
			if _, exists := v.Units[target]; !exists {
				// Create external boundary node
				v.Units[target] = createExternalBoundaryNode(target, parentPath)
			}
		}

		// Check incoming links (LinksFrom)
		for source := range unit.LinksFrom {
			if _, exists := v.Units[source]; !exists {
				// Create external boundary node
				v.Units[source] = createExternalBoundaryNode(source, parentPath)
			}
		}
	}
}
