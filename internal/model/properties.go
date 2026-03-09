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
}
