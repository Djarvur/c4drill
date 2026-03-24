package graph

import (
	"github.com/Djarvur/c4drill/internal/model"
)

// C4 level constants.
const (
	levelC1 = 1
	levelC2 = 2
	levelC3 = 3
)

// ShapeForType returns the appropriate shape for a unit type.
// Per CONTEXT.md decision: collapsed units use record shape.
// Expanded units are rendered as clusters (subgraphs), not nodes.
func ShapeForType(_ model.UnitType) Shape {
	return ShapeRecord
}

// IconForType returns the emoji icon for a unit type.
// Per CONTEXT.md design:
// - Person: U+1F464 (person emoji)
// - DB: U+26C1 (flag in hole, used as cylinder proxy)
// - Queue: U+255F/U+2562 (box drawing characters for bars).
func IconForType(t model.UnitType) string {
	switch t {
	case model.TypePerson, model.TypePersonExternal:
		return "\U0001F464" // person emoji
	case model.TypeDb, model.TypeDbExternal,
		model.TypeContainerDb, model.TypeComponentDb:
		return "\u26C1" // db cylinder icon (using golf flag as proxy)
	case model.TypeQueue, model.TypeQueueExternal,
		model.TypeContainerQueue, model.TypeComponentQueue:
		return "\u255F\n\u2562" // queue bars
	case model.TypeSystem, model.TypeSystemExternal,
		model.TypeContainer, model.TypeContainerBox,
		model.TypeComponent, model.TypeComponentBox,
		model.TypeBox:
		return "" // No icon for these types
	default:
		return ""
	}
}

// IsExternalType returns true if the type is an external variant.
func IsExternalType(t model.UnitType) bool {
	return t == model.TypePersonExternal ||
		t == model.TypeSystemExternal ||
		t == model.TypeDbExternal ||
		t == model.TypeQueueExternal
}

// IsPersonType returns true if the type is a person type (internal or external).
func IsPersonType(t model.UnitType) bool {
	return t == model.TypePerson || t == model.TypePersonExternal
}

// IsDbType returns true if the type is a database type (any level).
func IsDbType(t model.UnitType) bool {
	return t == model.TypeDb || t == model.TypeDbExternal ||
		t == model.TypeContainerDb || t == model.TypeComponentDb
}

// IsQueueType returns true if the type is a queue type (any level).
func IsQueueType(t model.UnitType) bool {
	return t == model.TypeQueue || t == model.TypeQueueExternal ||
		t == model.TypeContainerQueue || t == model.TypeComponentQueue
}

// IsSystemType returns true if the type is a system type.
func IsSystemType(t model.UnitType) bool {
	return t == model.TypeSystem || t == model.TypeSystemExternal
}

// IsContainerType returns true if the type is a container or containerBox type.
func IsContainerType(t model.UnitType) bool {
	return t == model.TypeContainer || t == model.TypeContainerBox
}

// IsComponentType returns true if the type is a component or componentBox type.
func IsComponentType(t model.UnitType) bool {
	return t == model.TypeComponent || t == model.TypeComponentBox
}

// IsBoxType returns true if the type is a box variant (box, containerBox, componentBox).
func IsBoxType(t model.UnitType) bool {
	return t == model.TypeBox || t == model.TypeContainerBox || t == model.TypeComponentBox
}

// HasExternalSubunits returns true if the unit has any external subunits.
// Returns false if unit is nil or has no subunits.
func HasExternalSubunits(unit *model.Unit) bool {
	if unit == nil {
		return false
	}

	for _, subunit := range unit.Subunits {
		if IsExternalType(subunit.Type) {
			return true
		}
	}

	return false
}

// GetBoxStyleByContents returns the style for a C1 box based on its contents.
// - Boxes with external subunits: grey border (PersonExternalBorder)
// - Boxes with only non-external subunits: dark blue border (PersonBorder)
// - Both cases: dashed border style
func GetBoxStyleByContents(unit *model.Unit) *NodeStyle {
	if HasExternalSubunits(unit) {
		return &NodeStyle{
			FillColor:   "",                              // Transparent background
			BorderColor: model.PersonExternalBorder,      // Grey for external boxes
			FontColor:   model.PersonExternalBorder,      // Font color matches border color
			BorderStyle: "dashed",
		}
	}

	return &NodeStyle{
		FillColor:   "",                     // Transparent background
		BorderColor: model.PersonBorder,     // Dark blue for internal boxes
		FontColor:   model.PersonBorder,     // Font color matches border color
		BorderStyle: "dashed",
	}
}

// LevelForType returns the C4 level (1, 2, or 3) for a unit type.
func LevelForType(t model.UnitType) int {
	switch t {
	case model.TypePerson, model.TypePersonExternal,
		model.TypeSystem, model.TypeSystemExternal,
		model.TypeDb, model.TypeDbExternal,
		model.TypeQueue, model.TypeQueueExternal,
		model.TypeBox:
		return levelC1
	case model.TypeContainer, model.TypeContainerDb,
		model.TypeContainerQueue, model.TypeContainerBox:
		return levelC2
	case model.TypeComponent, model.TypeComponentDb,
		model.TypeComponentQueue, model.TypeComponentBox:
		return levelC3
	default:
		return levelC1
	}
}

// GetStyleForType returns styling based on unit type and external status.
// Per CONTEXT.md decisions:
// - No inheritance: each unit's style is independent
// - External nodes: same size, solid border, external palette colors
// - Transparent backgrounds: FillColor is empty for all units.
func GetStyleForType(t model.UnitType, isExternal bool) *NodeStyle {
	if isExternal {
		return getExternalStyle(t)
	}

	return getLevelStyle(t)
}

// getLevelStyle returns the style for internal nodes based on their C4 level.
// Per user decision: all unit types at each level use the same color:
// - C1: dark blue (PersonBorder)
// - C2: blue (ContainerBorder)
// - C3: light blue (ComponentBorder)
// Box types use dashed borders to differentiate from other unit types.
func getLevelStyle(t model.UnitType) *NodeStyle {
	level := LevelForType(t)

	// Box types get dashed borders
	borderStyle := "solid"
	if IsBoxType(t) {
		borderStyle = "dashed"
	}

	switch level {
	case levelC1:
		return &NodeStyle{
			FillColor:   "", // Transparent background
			BorderColor: model.PersonBorder, // Dark blue for all C1 units
			FontColor:   model.PersonBorder, // Font color matches border color
			BorderStyle: borderStyle,
		}
	case levelC2:
		return &NodeStyle{
			FillColor:   "", // Transparent background
			BorderColor: model.ContainerBorder, // Blue for all C2 units
			FontColor:   model.ContainerBorder, // Font color matches border color
			BorderStyle: borderStyle,
		}
	case levelC3:
		return &NodeStyle{
			FillColor:   "", // Transparent background
			BorderColor: model.ComponentBorder, // Light blue for all C3 units
			FontColor:   model.ComponentBorder, // Font color matches border color
			BorderStyle: borderStyle,
		}
	default:
		return &NodeStyle{
			FillColor:   "", // Transparent background
			BorderColor: model.PersonBorder,
			FontColor:   model.PersonBorder, // Font color matches border color
			BorderStyle: borderStyle,
		}
	}
}

// getExternalStyle returns the style for external boundary nodes.
// Per user decision: all unit types at each level use the same color:
// - C1 external: dark gray (PersonExternalBorder)
// - C2 external: medium gray (ContainerExternalBorder)
// - C3 external: light gray (ComponentExternalBorder)
// Box types use dashed borders to differentiate from other unit types.
func getExternalStyle(t model.UnitType) *NodeStyle {
	level := LevelForType(t)

	// Box types get dashed borders
	borderStyle := "solid"
	if IsBoxType(t) {
		borderStyle = "dashed"
	}

	switch level {
	case levelC1:
		return &NodeStyle{
			FillColor:   "", // Transparent background
			BorderColor: model.PersonExternalBorder, // Dark gray for all C1 external units
			FontColor:   model.PersonExternalBorder, // Font color matches border color
			BorderStyle: borderStyle,
		}
	case levelC2:
		return &NodeStyle{
			FillColor:   "", // Transparent background
			BorderColor: model.ContainerExternalBorder, // Medium gray for all C2 external units
			FontColor:   model.ContainerExternalBorder, // Font color matches border color
			BorderStyle: borderStyle,
		}
	case levelC3:
		return &NodeStyle{
			FillColor:   "", // Transparent background
			BorderColor: model.ComponentExternalBorder, // Light gray for all C3 external units
			FontColor:   model.ComponentExternalBorder, // Font color matches border color
			BorderStyle: borderStyle,
		}
	default:
		return &NodeStyle{
			FillColor:   "", // Transparent background
			BorderColor: model.PersonExternalBorder,
			FontColor:   model.PersonExternalBorder, // Font color matches border color
			BorderStyle: borderStyle,
		}
	}
}
