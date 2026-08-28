package model

// Base colors from C4-PlantUML.
const (
	// ElementFontColor is the default font color for element labels.
	ElementFontColor = "#FFFFFF"
	// ArrowColor is the default color for arrows/links.
	ArrowColor = "#666666"
	// BoundaryColor is the default color for boundaries.
	BoundaryColor = "#444444"
)

// C1 Context level colors (background, border).
const (
	// PersonBackground is the background color for person elements.
	PersonBackground = "#08427B"
	// PersonBorder is the border color for person elements.
	PersonBorder = "#073B6F"
	// PersonExternalBackground is the background color for external person elements.
	PersonExternalBackground = "#686868"
	// PersonExternalBorder is the border color for external person elements.
	PersonExternalBorder = "#8A8A8A"
	// SystemBackground is the background color for system elements.
	SystemBackground = "#1168BD"
	// SystemBorder is the border color for system elements.
	SystemBorder = "#3C7FC0"
	// SystemExternalBackground is the background color for external system elements.
	SystemExternalBackground = "#999999"
	// SystemExternalBorder is the border color for external system elements.
	SystemExternalBorder = "#8A8A8A"
)

// C2 Container level colors (background, border).
const (
	// ContainerBackground is the background color for container elements.
	ContainerBackground = "#438DD5"
	// ContainerBorder is the border color for container elements.
	ContainerBorder = "#3C7FC0"
	// ContainerExternalBackground is the background color for external container elements.
	ContainerExternalBackground = "#B3B3B3"
	// ContainerExternalBorder is the border color for external container elements.
	ContainerExternalBorder = "#A6A6A6"
)

// C3 Component level colors (background, border).
const (
	// ComponentBackground is the background color for component elements.
	ComponentBackground = "#85BBF0"
	// ComponentBorder is the border color for component elements.
	ComponentBorder = "#78A8D8"
	// ComponentExternalBackground is the background color for external component elements.
	ComponentExternalBackground = "#CCCCCC"
	// ComponentExternalBorder is the border color for external component elements.
	ComponentExternalBorder = "#BFBFBF"
)

// Font colors by level.
const (
	// FontColorC1C2 is the font color for C1 and C2 level elements (white).
	FontColorC1C2 = "#FFFFFF"
	// FontColorC3 is the font color for C3 level elements (black).
	FontColorC3 = "#000000"
)

// Link-kind colours (KIND-01). Dark enough for legible edge-label text
// (fontcolor rides the same value) on the white SVG background; all three
// are distinct from the blue C4 node palette (#08427B/#1168BD/#438DD5/#85BBF0),
// the gray external palette (#999999/#B3B3B3/#CCCCCC), and the default edge
// colors (border blues + #666666). The legend (LEG-02) consumes these same
// constants so it can never drift from the renderer.
const (
	// LinkReadColour is the edge colour for kind = "read" (green, ~5.1:1 on white).
	LinkReadColour = "#2E7D32"
	// LinkWriteColour is the edge colour for kind = "write" (red, ~5.9:1 on white).
	LinkWriteColour = "#C62828"
	// LinkReadWriteColour is the edge colour for kind = "read-write" (purple —
	// hue-distinct from BOTH green and red, ~7.4:1 on white; a literal green-red
	// blend would be an illegible muddy brown).
	LinkReadWriteColour = "#6A1B9A"
)
