package view

import (
	"slices"

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
