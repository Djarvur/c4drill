package model

import "slices"

// UnitType represents the type of a C4 architecture unit.
type UnitType string

// C1 Context level unit types.
const (
	TypePerson         UnitType = "person"
	TypePersonExternal UnitType = "personExternal"
	TypeSystem         UnitType = "system"
	TypeSystemExternal UnitType = "systemExternal"
	TypeDb             UnitType = "db"
	TypeDbExternal     UnitType = "dbExternal"
	TypeQueue          UnitType = "queue"
	TypeQueueExternal  UnitType = "queueExternal"
	TypeBox            UnitType = "box"
)

// C2 Container level unit types.
const (
	TypeContainer      UnitType = "container"
	TypeContainerDb    UnitType = "containerDb"
	TypeContainerQueue UnitType = "containerQueue"
	TypeContainerBox   UnitType = "containerBox"
)

// C3 Component level unit types.
const (
	TypeComponent      UnitType = "component"
	TypeComponentDb    UnitType = "componentDb"
	TypeComponentQueue UnitType = "componentQueue"
	TypeComponentBox   UnitType = "componentBox"
)

// String returns the string representation of the UnitType.
func (t UnitType) String() string {
	return string(t)
}

// Unit represents a C4 architecture unit at any level (C1, C2, or C3).
type Unit struct {
	// Type is the discriminator that identifies the unit type.
	Type UnitType `toml:"type"`
	// Name is the display name of the unit.
	Name string `toml:"name"`
	// Description provides a brief description of the unit.
	Description string `toml:"description"`
	// Technology describes the technology used (NOT for person types).
	Technology string `toml:"technology"`
	// Reference is an optional external documentation URL (📖). Empty when unset.
	Reference string `toml:"reference"`
	// Color is the background color of the unit.
	Color string `toml:"color"`
	// Style is the visual style of the unit.
	Style string `toml:"style"`
	// Border is the border color/style of the unit.
	Border string `toml:"border"`
	// Edges specifies the edge style (cascades from parent).
	Edges string `toml:"edges"`
	// Width is the explicit width (0 = auto).
	Width float64 `toml:"width"`
	// Height is the explicit height (0 = auto).
	Height float64 `toml:"height"`
	// Expanded lists subunits that should be expanded by default.
	Expanded []string `toml:"expanded"`
	// Links contains outgoing relationships as a slice of Link structs.
	Links []Link `toml:"link"`
	// LinksFrom contains incoming relationships as a slice of Link structs.
	LinksFrom []Link `toml:"linkFrom"`
	// SubunitOrder tracks the definition order of subunit names (not from TOML).
	SubunitOrder []string `toml:"-"`
	// Subunits contains nested units within this unit.
	Subunits map[string]*Unit `toml:",inline"`
}

// Clone returns a deep copy of u. Value fields are copied; slice fields
// (Expanded, SubunitOrder, Links, LinksFrom) are deep-copied so the clone's
// slices have disjoint backing arrays; Subunits is a fresh map with every
// child *Unit recursively cloned (pointer-disjoint).
//
// The unexported Link.Mirror field is PRESERVED by the element-wise Link copy
// (model.Link has only value-type fields — no pointers/maps/slices — so a
// value copy duplicates each Link wholesale, Mirror included). This is the
// load-bearing HS-1 mitigation (Plan 31-02): the validator mutates
// Unit.LinksFrom in place (internal/validator/index.go:70-81), so a Clone that
// dropped Mirror (e.g. reflect/gob/json copiers, which cannot see unexported
// fields) would corrupt multiplicity counting for every template instantiation
// after the first.
//
// Clone on a nil *Unit returns nil (no panic).
func (u *Unit) Clone() *Unit {
	if u == nil {
		return nil
	}

	clone := *u // shallow struct copy: value fields + slice/map headers

	// Deep-copy slice fields. slices.Clone copies the backing array; for Links
	// and LinksFrom we do an explicit element-wise copy because each Link is a
	// value we want independently owned (defensive even though Link has no
	// reference-type fields).
	clone.Expanded = slices.Clone(u.Expanded)
	clone.SubunitOrder = slices.Clone(u.SubunitOrder)
	clone.Links = cloneLinks(u.Links)
	clone.LinksFrom = cloneLinks(u.LinksFrom)

	// Deep-copy Subunits: fresh map, each child *Unit recursively cloned.
	if u.Subunits != nil {
		clone.Subunits = make(map[string]*Unit, len(u.Subunits))
		for k, child := range u.Subunits {
			clone.Subunits[k] = child.Clone()
		}
	}

	return &clone
}

// cloneLinks returns a deep copy of a Link slice: a new backing array with
// each Link value-copied. Because model.Link has ONLY value-type fields
// (strings, ints, enums, and the unexported Mirror bool — no pointers, maps,
// or slices), a value copy fully duplicates each Link including Mirror.
func cloneLinks(links []Link) []Link {
	if links == nil {
		return nil
	}

	out := make([]Link, len(links))
	copy(out, links) // value copy preserves Mirror for every element

	return out
}
