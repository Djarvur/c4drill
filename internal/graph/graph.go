// Package graph provides types and functions for constructing graph structures
// from C4 architecture views.
package graph

import "github.com/Djarvur/c4drill/internal/model"

// Shape represents the visual shape of a node in the graph.
type Shape string

const (
	// ShapeRecord is the standard record shape for nodes.
	ShapeRecord Shape = "record"
	// ShapeHTML indicates the node uses an HTML-like label for complex formatting.
	ShapeHTML Shape = "html"
	// ShapeCluster indicates a subgraph cluster containing nested nodes.
	ShapeCluster Shape = "cluster"
)

// ArrowDirection represents the direction of an arrow on an edge.
type ArrowDirection string

const (
	// ArrowForward indicates an arrow at the target end.
	ArrowForward ArrowDirection = "forward"
	// ArrowReverse indicates an arrow at the source end.
	ArrowReverse ArrowDirection = "reverse"
	// ArrowBoth indicates arrows at both ends.
	ArrowBoth ArrowDirection = "both"
	// ArrowNone indicates no arrow on the edge.
	ArrowNone ArrowDirection = "none"
)

// Graph represents a graph structure ready for DOT rendering.
type Graph struct {
	// Title is the diagram title.
	Title string
	// Direction is the layout direction (default: "TB" for top-to-bottom).
	Direction string
	// EdgeStyle is the edge routing style (straight, spline, square).
	EdgeStyle string
	// Nodes are all top-level nodes in the graph.
	Nodes []*Node
	// Edges are all edges connecting nodes.
	Edges []*Edge
	// Clusters are subgraph clusters for expanded units.
	Clusters []*Cluster
	// Legend contains legend information for the diagram.
	Legend *Legend
	// Navigation contains breadcrumb and back-link info (nil for C1).
	Navigation *Navigation
}

// Node represents a single node in the graph.
type Node struct {
	// ID is a unique identifier (the unit's full path).
	ID string
	// Label contains the formatted label parts.
	Label *Label
	// Shape is the node shape (determined by type).
	Shape Shape
	// Type is the unit type (used for special rendering like Person HTML labels).
	Type model.UnitType
	// Style contains visual styling attributes.
	Style *NodeStyle
	// IsExternal indicates if this is an external boundary node.
	IsExternal bool
	// IsInCluster indicates if this node is inside a cluster.
	IsInCluster bool
	// ExploreURL is the relative path for drill-down (empty if not expandable).
	ExploreURL string
}

// Edge represents a connection between two nodes.
type Edge struct {
	// Source is the source node ID.
	Source string
	// Target is the target node ID.
	Target string
	// Label contains edge label information.
	Label *EdgeLabel
	// Style is the line style (solid, dashed, dotted).
	Style string
	// ArrowHead is the arrow direction.
	ArrowHead ArrowDirection
	// Color is the edge line and label color (from source unit's border color or explicit override).
	Color string
	// MinLen is the minimum length (minlen attribute) for the edge.
	MinLen int
}

// EdgeLabel contains label information for an edge.
type EdgeLabel struct {
	// Technology appears in brackets (e.g., [HTTP]).
	Technology string
	// Description describes the relationship.
	Description string
	// Position is where the label appears (middle, head, tail).
	Position string
}

// Cluster represents a subgraph cluster for expanded units.
type Cluster struct {
	// ID is the cluster identifier.
	ID string
	// Label is the cluster label (parent unit info).
	Label *Label
	// Nodes are the nodes inside this cluster.
	Nodes []*Node
	// Clusters are nested clusters inside this cluster (for all-expanded mode).
	Clusters []*Cluster
	// Style contains cluster styling attributes.
	Style *NodeStyle
	// Type is the unit type for HTML label dispatch.
	Type model.UnitType
	// IsExternal indicates if this cluster represents an external unit.
	IsExternal bool
}

// Label represents a node label with multiple parts.
type Label struct {
	// Name is the primary name (with [+] indicator if expandable).
	Name string
	// Technology is the technology string (optional).
	Technology string
	// Description is the description text (optional).
	Description string
	// Icon is the emoji/icon prefix (optional).
	Icon string
}

// NodeStyle contains visual styling for a node.
type NodeStyle struct {
	// FillColor is the background color.
	FillColor string
	// BorderColor is the border color.
	BorderColor string
	// FontColor is the text color.
	FontColor string
	// BorderStyle is "solid" or "dashed".
	BorderStyle string
}

// Legend contains legend information for the diagram.
// The exact content will be defined in Phase 4 (Rendering).
type Legend struct {
	// Placeholder for legend content to be implemented in Phase 4.
}
