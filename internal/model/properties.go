package model

// Properties represents the root-level [properties] section in a TOML file.
type Properties struct {
	// Name is the project or diagram name.
	Name string `toml:"name"`
	// Description provides a brief description of the project.
	Description string `toml:"description"`
	// Color is the default background color.
	Color string `toml:"color"`
	// Style is the default visual style.
	Style string `toml:"style"`
	// Border is the default border color/style.
	Border string `toml:"border"`
	// Edges specifies the default edge style.
	Edges string `toml:"edges"`
	// LineLength is the maximum line length before wrapping (0 = auto).
	LineLength int `toml:"lineLength"`
	// Expanded lists units that should be expanded by default.
	Expanded []string `toml:"expanded"`
	// Legend controls the diagram legend (upper-right). Nil means enabled —
	// the Go zero value (false) cannot express "default on", so absence is
	// read as true. Explicitly false disables the legend for the whole model.
	Legend *bool `toml:"legend"`
	// LegendLines lists author-defined legend rows rendered after the
	// default colour explanations.
	LegendLines []LegendLine `toml:"legendLine"`
}

// LegendLine is one author-defined legend row.
type LegendLine struct {
	// Label is the explanation text for the row.
	Label string `toml:"label"`
	// Color is the swatch/line colour of the row.
	Color string `toml:"color"`
	// Style is the optional line style (solid, dashed, dotted) for rows that
	// demonstrate a line variant; empty rows render a colour swatch.
	Style string `toml:"style"`
}
