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
