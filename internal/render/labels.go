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

// HTML Label Builder Functions

// buildPersonHTMLLabel generates an HTML table label for Person-type nodes.
// Format: emoji (rowspan=2) | name (bold) / description
// Per CONTEXT.md: Person has NO technology field.
func buildPersonHTMLLabel(label *graph.Label) string {
	if label == nil {
		return ""
	}

	// Calculate max characters for word wrapping
	rowCount := 1 // name
	if label.Description != "" {
		rowCount++
	}

	maxChars := labelMaxCharsForPerson(rowCount)

	var sb strings.Builder
	sb.WriteString(`<table border="0" cellpadding="0" cellspacing="0">`)

	// Calculate rowspan for emoji: always 2 (name + description) if description present, else 1
	rowspan := 1
	if label.Description != "" {
		rowspan = 2
	}

	// Row 1: Emoji (rowspan) + Name
	sb.WriteString(`<tr align="center">`)

	// Emoji column with person emoji
	sb.WriteString(`<td rowspan="`)
	sb.WriteString(rowspanString(rowspan))
	sb.WriteString(`" valign="middle"><font POINT-SIZE="32">&#x1F464;</font></td>`)

	sb.WriteString(`<td valign="bottom"><b>`)
	sb.WriteString(wrapAndEscape(label.Name, maxChars))
	sb.WriteString(`</b></td></tr>`)

	// Row 2: Description (if present)
	if label.Description != "" {
		sb.WriteString(`<tr align="center"><td valign="top">`)
		sb.WriteString(wrapAndEscape(label.Description, maxChars))
		sb.WriteString(`</td></tr>`)
	}

	sb.WriteString(`</table>`)

	return sb.String()
}

// rowspanString converts rowspan int to string without importing strconv.
func rowspanString(n int) string {
	if n == 1 {
		return "1"
	}

	return "2"
}

// buildDbHTMLLabel generates an HTML table label for Database-type nodes.
// Format: name (bold) / [technology] italic / description
// Single-column layout without icon column.
func buildDbHTMLLabel(label *graph.Label) string {
	return buildNoIconHTMLLabel(label, labelMaxCharsForCylinder)
}

// buildQueueHTMLLabel generates an HTML table label for Queue-type nodes.
// Format: ASCII graphic / name (bold) / [technology] italic / description
// 4-row table with ASCII art graphic as first row.
func buildQueueHTMLLabel(label *graph.Label) string {
	if label == nil {
		return ""
	}

	// Graphic row doesn't wrap but counts for proportion.
	maxChars := labelMaxCharsForQueue(labelRowCount(label) + 1)

	var sb strings.Builder
	writeLabelTableStart(&sb)
	sb.WriteString(`<tr align="center"><td valign="middle">`)
	sb.WriteString("═╦╩═╦═══")
	sb.WriteString(`</td></tr>`)
	writeNameRow(&sb, label.Name, maxChars)
	writeTechnologyRow(&sb, label.Technology, maxChars)
	writeDescriptionRow(&sb, label.Description, maxChars)
	writeLabelTableEnd(&sb)

	return sb.String()
}

// buildSystemHTMLLabel generates an HTML table label for System-type nodes.
// Format: name (bold) / [technology] italic / description
// Single-column layout without icon column.
func buildSystemHTMLLabel(label *graph.Label) string {
	return buildNoIconHTMLLabel(label, labelMaxCharsNoIcon)
}

// buildContainerHTMLLabel generates an HTML table label for Container-type nodes.
// Format: name (bold) / [technology] italic / description
// Single-column layout without icon column.
func buildContainerHTMLLabel(label *graph.Label) string {
	return buildNoIconHTMLLabel(label, labelMaxCharsNoIcon)
}

// buildComponentHTMLLabel generates an HTML table label for Component-type nodes.
// Format: name (bold) / [technology] italic / description
// Single-column layout without icon column.
func buildComponentHTMLLabel(label *graph.Label) string {
	return buildNoIconHTMLLabel(label, labelMaxCharsNoIcon)
}

// buildBoxHTMLLabel generates an HTML table label for Box-type nodes.
// Format: name (bold) / [technology] italic / description
// Single-column layout without icon column.
// Output does NOT contain curly brackets {} unlike record labels.
func buildBoxHTMLLabel(label *graph.Label) string {
	return buildNoIconHTMLLabel(label, labelMaxCharsNoIcon)
}

// buildNoIconHTMLLabel builds the standard name/technology/description table
// label without an icon column, using maxCharsFor for word-wrapping.
func buildNoIconHTMLLabel(label *graph.Label, maxCharsFor func(rowCount int) int) string {
	if label == nil {
		return ""
	}

	maxChars := maxCharsFor(labelRowCount(label))

	var sb strings.Builder
	writeLabelTableStart(&sb)
	writeNameRow(&sb, label.Name, maxChars)
	writeTechnologyRow(&sb, label.Technology, maxChars)
	writeDescriptionRow(&sb, label.Description, maxChars)
	writeLabelTableEnd(&sb)

	return sb.String()
}

// labelRowCount counts the content rows: name plus optional technology and
// description.
func labelRowCount(label *graph.Label) int {
	rowCount := 1 // name
	if label.Technology != "" {
		rowCount++
	}

	if label.Description != "" {
		rowCount++
	}

	return rowCount
}

func writeLabelTableStart(sb *strings.Builder) {
	sb.WriteString(`<table border="0" cellpadding="0" cellspacing="0">`)
}

func writeLabelTableEnd(sb *strings.Builder) {
	sb.WriteString(`</table>`)
}

func writeNameRow(sb *strings.Builder, name string, maxChars int) {
	sb.WriteString(`<tr align="center"><td valign="bottom"><b>`)
	sb.WriteString(wrapAndEscape(name, maxChars))
	sb.WriteString(`</b></td></tr>`)
}

func writeTechnologyRow(sb *strings.Builder, technology string, maxChars int) {
	if technology == "" {
		return
	}

	sb.WriteString(`<tr align="center"><td valign="middle"><i>[`)
	sb.WriteString(wrapAndEscape(technology, maxChars))
	sb.WriteString(`]</i></td></tr>`)
}

func writeDescriptionRow(sb *strings.Builder, description string, maxChars int) {
	if description == "" {
		return
	}

	sb.WriteString(`<tr align="center"><td valign="top">`)
	sb.WriteString(wrapAndEscape(description, maxChars))
	sb.WriteString(`</td></tr>`)
}
