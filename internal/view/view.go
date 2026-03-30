// Package view provides types and functions for generating scoped C4 architecture views.
package view

import "github.com/Djarvur/c4drill/internal/model"

// Level represents the C4 hierarchy level.
type Level int

const (
	// LevelC1 is the Context level (top-level systems and actors).
	LevelC1 Level = iota
	// LevelC2 is the Container level (subunits of expanded systems).
	LevelC2
	// LevelC3 is the Component level (subunits of expanded containers).
	LevelC3
)

// View represents a scoped view of the architecture model.
// It contains the units visible at a specific C4 level.
type View struct {
	// Level indicates the C4 level (C1, C2, or C3).
	Level Level
	// Title is the diagram title (from properties.name or parent name).
	Title string
	// UnitOrder tracks the definition order of unit paths.
	UnitOrder []string
	// Units are the units visible in this view (keyed by full path).
	Units map[string]*Entry
	// Edges is the edge routing style for this view.
	Edges string
	// Parent is the parent unit path for C2/C3 views (empty for C1).
	Parent string
	// ExpandedUnit is the unit being expanded (for C2/C3 views).
	ExpandedUnit string
}

// Entry represents a unit entry within a view.
// It wraps a model.Unit with additional view-specific metadata.
type Entry struct {
	// Unit is the underlying model unit.
	Unit *model.Unit
	// FullPath is the dotted path (e.g., "mainapp.api").
	FullPath string
	// IsExpanded indicates if this unit is expanded (shows subunits).
	IsExpanded bool
	// IsExternal indicates if this is an external boundary node.
	IsExternal bool
	// HasSubunits indicates if this unit has children (for [+] indicator).
	HasSubunits bool
	// ResolvedLinks contains outgoing links resolved for the current view level.
	// When non-nil, the graph builder uses these instead of Unit.Links.
	ResolvedLinks []model.Link
	// ResolvedLinksFrom contains incoming links resolved for the current view level.
	// When non-nil, the graph builder uses these instead of Unit.LinksFrom.
	ResolvedLinksFrom []model.Link
}

// IsExternalType returns true if the unit type is an external variant.
// External types represent systems or actors outside the current scope.
func IsExternalType(t model.UnitType) bool {
	return t == model.TypePersonExternal ||
		t == model.TypeSystemExternal ||
		t == model.TypeDbExternal ||
		t == model.TypeQueueExternal
}
