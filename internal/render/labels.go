package render

import (
	"strings"
	"unicode/utf8"

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

// buildEdgeLabel generates an HTML-table label for an edge.
// Form (D-01/D-02/D-04): a borderless rectangle <table border="0"
// cellpadding="0" cellspacing="0"> with the [Technology] row (italic) and
// the wrapped Description row below it.
// - If only Technology: table with the [Technology] row only
// - If only Description: table with the Description row only
// - If both: [Technology] row, then the Description wrapped below
// The width derives from LabelRatio via labelMaxCharsForText — self-
// consistent aspect-ratio sizing: both the technology and the description
// wrap, so the total text length drives the width (supersedes the D-03
// 2-row floor). A minCharsPerLine floor prevents over-wrapping of short
// labels.
func buildEdgeLabel(label *graph.EdgeLabel) string {
	if label == nil {
		return ""
	}

	if label.Technology == "" && label.Description == "" {
		return ""
	}

	textLen := utf8.RuneCountInString(label.Technology) +
		utf8.RuneCountInString(label.Description)

	maxChars := max(labelMaxCharsForText(0, textLen, LabelRatio), minCharsPerLine)

	var sb strings.Builder
	writeLabelTableStart(&sb)
	writeTechnologyRow(&sb, label.Technology, maxChars)
	writeDescriptionRow(&sb, label.Description, maxChars)
	writeLabelTableEnd(&sb)

	return sb.String()
}

// HTML Label Builder Functions

// buildPersonHTMLLabel generates an HTML table label for Person-type nodes.
// Format: emoji (rowspan=2) | name (bold) / description
// Per CONTEXT.md: Person has NO technology field.
func buildPersonHTMLLabel(label *graph.Label) string {
	if label == nil {
		return ""
	}

	// Calculate max characters for word wrapping: the name and description
	// both wrap, so the total text length drives the width (self-consistent
	// ratio sizing, icon column subtracted).
	textLen := utf8.RuneCountInString(label.Name) +
		utf8.RuneCountInString(label.Description)

	maxChars := labelMaxCharsForPerson(0, textLen)

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
// Format: name (bold) / [technology] italic / description. No icon row — the
// queue identity is carried by the pipe shape drawn around the node
// (internal/render/pipe.go); a text-bar graphic inside a pipe would be
// double graphics.
func buildQueueHTMLLabel(label *graph.Label) string {
	if label == nil {
		return ""
	}

	textLen := utf8.RuneCountInString(label.Name) +
		utf8.RuneCountInString(label.Technology) +
		utf8.RuneCountInString(label.Description)

	maxChars := max(labelMaxCharsForQueue(1, textLen), minCharsPerLine)

	var sb strings.Builder
	writeLabelTableStart(&sb)
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
func buildNoIconHTMLLabel(label *graph.Label, maxCharsFor func(fixedRows, textLen int) int) string {
	if label == nil {
		return ""
	}

	// Name, technology and description all wrap; their total text length
	// drives the width (self-consistent ratio sizing). The minCharsPerLine
	// floor prevents over-wrapping of short labels.
	textLen := utf8.RuneCountInString(label.Name) +
		utf8.RuneCountInString(label.Technology) +
		utf8.RuneCountInString(label.Description)

	maxChars := max(maxCharsFor(0, textLen), minCharsPerLine)

	var sb strings.Builder
	writeLabelTableStart(&sb)
	writeNameRow(&sb, label.Name, maxChars)
	writeTechnologyRow(&sb, label.Technology, maxChars)
	writeDescriptionRow(&sb, label.Description, maxChars)
	writeLabelTableEnd(&sb)

	return sb.String()
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
