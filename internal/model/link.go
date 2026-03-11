// Package model defines the domain types for the C4 architecture model.
package model

// ArrowDirection represents the direction of an arrow on a link.
type ArrowDirection string

const (
	// ArrowForward indicates an arrow at the target end (default).
	ArrowForward ArrowDirection = "forward"
	// ArrowReverse indicates an arrow at the source end.
	ArrowReverse ArrowDirection = "reverse"
	// ArrowBidirectional indicates arrows at both ends.
	ArrowBidirectional ArrowDirection = "bidirectional"
	// ArrowNone indicates no arrow on the link.
	ArrowNone ArrowDirection = "none"
)

// RankDirection represents the ranking direction for layout purposes.
type RankDirection string

const (
	// RankForward indicates the target ranks after the source.
	RankForward RankDirection = "forward"
	// RankReverse indicates the target ranks before the source.
	RankReverse RankDirection = "reverse"
	// RankEqual indicates the target and source are on the same rank.
	RankEqual RankDirection = "equal"
)

// LabelPosition represents where the label appears on a link.
type LabelPosition string

const (
	// LabelMiddle indicates the label is in the middle of the link (default).
	LabelMiddle LabelPosition = "middle"
	// LabelTail indicates the label is at the tail (source) of the link.
	LabelTail LabelPosition = "tail"
	// LabelHead indicates the label is at the head (target) of the link.
	LabelHead LabelPosition = "head"
)

// Link represents a relationship between two units.
type Link struct {
	// Peer is the name of the linked unit (explicitly set via TOML peer field).
	Peer string `toml:"peer"`
	// Arrow indicates the arrow direction on the link.
	Arrow ArrowDirection `toml:"arrow"`
	// Rank indicates the ranking direction for layout.
	Rank RankDirection `toml:"rank"`
	// Color is the color of the link line.
	Color string `toml:"color"`
	// Style is the line style (solid, dashed, dotted).
	Style string `toml:"style"`
	// Technology describes the protocol or technology used (e.g., HTTP, TCP).
	Technology string `toml:"technology"`
	// Description describes the relationship.
	Description string `toml:"description"`
	// LabelPosition indicates where the label appears on the link.
	LabelPosition LabelPosition `toml:"labelPosition"`
}
