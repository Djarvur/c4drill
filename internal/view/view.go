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
	// RootTitle is the C1 root diagram title (from properties.name).
	// Used by navigation to display correct back-links text for C2/C3 views.
	RootTitle string
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
	// ExpandedUnitModel is the model unit for the expanded entity (for C2/C3 views).
	// Used by the graph builder to create the boundary cluster label.
	ExpandedUnitModel *model.Unit
	// AllExpanded indicates this view contains ALL units at all nesting levels
	// (--expanded mode). When true, edge deduplication keeps the technology+description
	// key and all edges render at penwidth 2.0 (COMPAT-02).
	AllExpanded bool
	// VisiblePaths tracks paths of subunits rendered as nodes INSIDE an expanded
	// top-level cluster (C1). They are in Units for edge building but must not be
	// rendered as top-level nodes.
	VisiblePaths map[string]bool
	// AncestorNames maps a dotted unit path to its display (pretty) Name,
	// populated by the view generators for every ancestor of the ExpandedUnit
	// (plus the ExpandedUnit itself). Used by the graph builder to render
	// breadcrumb items with human-readable names instead of raw path segments.
	AncestorNames map[string]string
	// ShowLegend reports whether the diagram carries the upper-right legend
	// (LEG-01): properties.legend absent or true. All four generators set it
	// from m.Properties so the legend reaches every diagram generation.
	ShowLegend bool
	// LegendLines lists the author-defined legend rows (LEG-03) rendered
	// after the default colour explanations.
	LegendLines []model.LegendLine
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
	// IsBoundary indicates this is a boundary node (outside the expanded unit's
	// scope) that should be rendered at the top level, outside the boundary
	// cluster. Unlike IsExternal (which reflects unit TYPE for styling),
	// IsBoundary reflects VIEW SCOPE for placement — a regular container that
	// is a sibling of the expanded unit is IsBoundary=true (render outside
	// cluster) but IsExternal=false (keep its normal blue styling).
	IsBoundary bool
	// UnfoldChain is set by the view generator when a deep-link ancestor chain
	// (CTX-02) was inserted beneath this depicted ancestor: a link target
	// deeper than the nearest visible ancestor gained visible entries down to
	// the TRUE target. The graph layer renders such entries as recursive
	// clusters exactly like IsExpanded ones — deliberately a separate flag so
	// author-expansion semantics (D-07 guard, 🔍 logic) stay untouched.
	UnfoldChain bool
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
