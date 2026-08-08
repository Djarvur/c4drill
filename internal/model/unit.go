package model

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
