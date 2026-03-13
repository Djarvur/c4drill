package render

import (
	"strings"

	"github.com/Djarvur/c4drill/internal/graph"
)

// buildRecordLabel generates a record-style label for a node.
// Format: "{Name|Technology|Description}" for record shapes.
// Each field is on a separate row with left alignment.
func buildRecordLabel(label *graph.Label) string {
	if label == nil {
		return ""
	}

	var parts []string

	// Build name with optional icon
	var nameBuilder strings.Builder
	if label.Icon != "" {
		nameBuilder.WriteString(label.Icon)
		nameBuilder.WriteString(" ")
	}
	nameBuilder.WriteString(label.Name)
	parts = append(parts, nameBuilder.String())

	// Add technology if present
	if label.Technology != "" {
		parts = append(parts, label.Technology)
	}

	// Add description if present
	if label.Description != "" {
		parts = append(parts, label.Description)
	}

	// Join with | for record format and wrap in {}
	return "{" + strings.Join(parts, "|") + "}"
}

// buildPersonRecordLabel generates a record-style label for Person-type nodes.
// Format: "{icon}|{name|description}" creates a two-column layout with icon
// spanning the full height on the left, and name/description stacked on the right.
//
// Visual layout:
//
//	┌──────┬──────────────────┐
//	│      │ Name             │
//	│ icon ├──────────────────┤
//	│      │ Description      │
//	└──────┴──────────────────┘
func buildPersonRecordLabel(label *graph.Label) string {
	if label == nil {
		return ""
	}

	var sb strings.Builder

	// Left cell: icon (spans full height)
	sb.WriteString("{")
	if label.Icon != "" {
		sb.WriteString(label.Icon)
	}
	sb.WriteString("}|{")

	// Right cells: name over description
	sb.WriteString(label.Name)

	if label.Description != "" {
		sb.WriteString("|")
		sb.WriteString(label.Description)
	}

	sb.WriteString("}")

	return sb.String()
}

// buildHTMLLabel generates an HTML table label for a node.
// Format:
//
//	<<TABLE BORDER="0" CELLBORDER="1" CELLSPACING="0" CELLPADDING="4">
//	  <TR><TD>{icon} {name}</TD></TR>
//	  <TR><TD>{technology}</TD></TR>  (if present)
//	  <TR><TD>{description}</TD></TR> (if present)
//	</TABLE>>
//
// Deprecated: Use buildRecordLabel for record shapes instead.
func buildHTMLLabel(label *graph.Label) string {
	if label == nil {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(`<<TABLE BORDER="0" CELLBORDER="1" CELLSPACING="0" CELLPADDING="4">`)
	sb.WriteString(`<TR><TD>`)

	// Add icon if present
	if label.Icon != "" {
		sb.WriteString(label.Icon)
		sb.WriteString(" ")
	}

	sb.WriteString(label.Name)
	sb.WriteString(`</TD></TR>`)

	// Add technology row if present
	if label.Technology != "" {
		sb.WriteString(`<TR><TD>`)
		sb.WriteString(label.Technology)
		sb.WriteString(`</TD></TR>`)
	}

	// Add description row if present
	if label.Description != "" {
		sb.WriteString(`<TR><TD>`)
		sb.WriteString(label.Description)
		sb.WriteString(`</TD></TR>`)
	}

	sb.WriteString(`</TABLE>>`)

	return sb.String()
}

// buildEdgeLabel generates a label for an edge.
// Format: [Technology]\nDescription
// - If only Technology: [Technology]
// - If only Description: Description
// - If both: [Technology]\nDescription.
func buildEdgeLabel(label *graph.EdgeLabel) string {
	if label == nil {
		return ""
	}

	var parts []string

	if label.Technology != "" {
		parts = append(parts, "["+label.Technology+"]")
	}

	if label.Description != "" {
		parts = append(parts, label.Description)
	}

	return strings.Join(parts, "\n")
}

// HTML Label Builder Functions (Wave 1 implementation)
// These stub functions are placeholders that will be implemented in plan 12-01.
// They return empty strings so tests compile but fail assertions.

// buildPersonHTMLLabel generates an HTML table label for Person-type nodes.
// Implementation pending in Wave 1 (plan 12-01).
func buildPersonHTMLLabel(label *graph.Label) string {
	if label == nil {
		return ""
	}
	// TODO: Implement in Wave 1 (plan 12-01)
	return ""
}

// buildDbHTMLLabel generates an HTML table label for Database-type nodes.
// Implementation pending in Wave 1 (plan 12-01).
func buildDbHTMLLabel(label *graph.Label) string {
	if label == nil {
		return ""
	}
	// TODO: Implement in Wave 1 (plan 12-01)
	return ""
}

// buildQueueHTMLLabel generates an HTML table label for Queue-type nodes.
// Implementation pending in Wave 1 (plan 12-01).
func buildQueueHTMLLabel(label *graph.Label) string {
	if label == nil {
		return ""
	}
	// TODO: Implement in Wave 1 (plan 12-01)
	return ""
}

// buildSystemHTMLLabel generates an HTML table label for System-type nodes.
// Implementation pending in Wave 1 (plan 12-01).
func buildSystemHTMLLabel(label *graph.Label) string {
	if label == nil {
		return ""
	}
	// TODO: Implement in Wave 1 (plan 12-01)
	return ""
}

// buildContainerHTMLLabel generates an HTML table label for Container-type nodes.
// Implementation pending in Wave 1 (plan 12-01).
func buildContainerHTMLLabel(label *graph.Label) string {
	if label == nil {
		return ""
	}
	// TODO: Implement in Wave 1 (plan 12-01)
	return ""
}

// buildComponentHTMLLabel generates an HTML table label for Component-type nodes.
// Implementation pending in Wave 1 (plan 12-01).
func buildComponentHTMLLabel(label *graph.Label) string {
	if label == nil {
		return ""
	}
	// TODO: Implement in Wave 1 (plan 12-01)
	return ""
}
