package graph

import (
	"github.com/Djarvur/c4drill/internal/model"
)

// ShapeForType returns the appropriate shape for a unit type.
// Per CONTEXT.md decision: all types use HTML-like labels for proper cell formatting.
func ShapeForType(_ model.UnitType) Shape {
	return ShapeHTML
}

// IconForType returns the emoji icon for a unit type.
// Per CONTEXT.md design:
// - Person: U+1F464 (person emoji)
// - DB: U+26C1 (flag in hole, used as cylinder proxy)
// - Queue: U+255F/U+2562 (box drawing characters for bars)
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
	default:
		return "" // System, Container, Component, Box
	}
}

// IsExternalType returns true if the type is an external variant.
func IsExternalType(t model.UnitType) bool {
	return t == model.TypePersonExternal ||
		t == model.TypeSystemExternal ||
		t == model.TypeDbExternal ||
		t == model.TypeQueueExternal
}

// LevelForType returns the C4 level (1, 2, or 3) for a unit type.
func LevelForType(t model.UnitType) int {
	switch t {
	case model.TypePerson, model.TypePersonExternal,
		model.TypeSystem, model.TypeSystemExternal,
		model.TypeDb, model.TypeDbExternal,
		model.TypeQueue, model.TypeQueueExternal,
		model.TypeBox:
		return 1 // C1
	case model.TypeContainer, model.TypeContainerDb,
		model.TypeContainerQueue:
		return 2 // C2
	case model.TypeComponent, model.TypeComponentDb,
		model.TypeComponentQueue:
		return 3 // C3
	default:
		return 1 // Default to C1
	}
}

// GetStyleForType returns styling based on unit type and external status.
// Per CONTEXT.md decisions:
// - No inheritance: each unit's style is independent
// - External nodes: same size, dashed border, external palette colors
func GetStyleForType(t model.UnitType, isExternal bool) *NodeStyle {
	if isExternal {
		return getExternalStyle(t)
	}
	return getLevelStyle(t)
}

// getLevelStyle returns the style for internal nodes based on their C4 level.
func getLevelStyle(t model.UnitType) *NodeStyle {
	level := LevelForType(t)

	switch level {
	case 1:
		return &NodeStyle{
			FillColor:   model.SystemBackground,
			BorderColor: model.SystemBorder,
			FontColor:   model.FontColorC1C2,
			BorderStyle: "solid",
		}
	case 2:
		return &NodeStyle{
			FillColor:   model.ContainerBackground,
			BorderColor: model.ContainerBorder,
			FontColor:   model.FontColorC1C2,
			BorderStyle: "solid",
		}
	case 3:
		return &NodeStyle{
			FillColor:   model.ComponentBackground,
			BorderColor: model.ComponentBorder,
			FontColor:   model.FontColorC3,
			BorderStyle: "solid",
		}
	default:
		return &NodeStyle{
			FillColor:   model.SystemBackground,
			BorderColor: model.SystemBorder,
			FontColor:   model.FontColorC1C2,
			BorderStyle: "solid",
		}
	}
}

// getExternalStyle returns the style for external boundary nodes.
// Per CONTEXT.md: dashed border, external palette colors.
func getExternalStyle(t model.UnitType) *NodeStyle {
	level := LevelForType(t)

	switch level {
	case 1:
		return &NodeStyle{
			FillColor:   model.SystemExternalBackground,
			BorderColor: model.SystemExternalBorder,
			FontColor:   model.FontColorC1C2,
			BorderStyle: "dashed",
		}
	case 2:
		return &NodeStyle{
			FillColor:   model.ContainerExternalBackground,
			BorderColor: model.ContainerExternalBorder,
			FontColor:   model.FontColorC1C2,
			BorderStyle: "dashed",
		}
	case 3:
		return &NodeStyle{
			FillColor:   model.ComponentExternalBackground,
			BorderColor: model.ComponentExternalBorder,
			FontColor:   model.FontColorC3,
			BorderStyle: "dashed",
		}
	default:
		return &NodeStyle{
			FillColor:   model.SystemExternalBackground,
			BorderColor: model.SystemExternalBorder,
			FontColor:   model.FontColorC1C2,
			BorderStyle: "dashed",
		}
	}
}
