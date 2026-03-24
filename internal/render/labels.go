package render

import (
	"strconv"
	"strings"

	"github.com/Djarvur/c4drill/internal/graph"
	"github.com/Djarvur/c4drill/internal/model"
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

	// Calculate horizontal padding to compensate for icon column
	// Use minimum of 2 rows for padding calculation to match labelMaxCharsForPerson
	effectiveRows := rowCount
	if effectiveRows < 2 {
		effectiveRows = 2
	}
	totalHeight := effectiveRows * 18
	hPadding := totalHeight / 10 // 10% for minimal padding

	var sb strings.Builder
	sb.WriteString(`<table border="0" cellpadding="0" cellspacing="0">`)

	// Calculate rowspan for emoji: always 2 (name + description) if description present, else 1
	rowspan := 1
	if label.Description != "" {
		rowspan = 2
	}

	// Row 1: Emoji (rowspan) + Name with horizontal padding
	sb.WriteString(`<tr align="center">`)

	// Emoji column with person emoji
	sb.WriteString(`<td width="36" rowspan="`)
	sb.WriteString(strconv.Itoa(rowspan))
	sb.WriteString(`" valign="middle"><font size="+4">&#x1F464;</font></td>`)

	sb.WriteString(`<td valign="bottom">`)
	sb.WriteString(`<table border="0" cellpadding="0" cellspacing="0"><tr>`)
	sb.WriteString(`<td width="`)
	sb.WriteString(strconv.Itoa(hPadding))
	sb.WriteString(`"></td><td><b>`)
	sb.WriteString(wrapAndEscape(label.Name, maxChars))
	sb.WriteString(`</b></td><td width="`)
	sb.WriteString(strconv.Itoa(hPadding))
	sb.WriteString(`"></td></tr></table></td>`)
	sb.WriteString(`</tr>`)

	// Row 2: Description (if present) with horizontal padding
	if label.Description != "" {
		sb.WriteString(`<tr align="center">`)
		sb.WriteString(`<td valign="top">`)
		sb.WriteString(`<table border="0" cellpadding="0" cellspacing="0"><tr>`)
		sb.WriteString(`<td width="`)
		sb.WriteString(strconv.Itoa(hPadding))
		sb.WriteString(`"></td><td>`)
		sb.WriteString(wrapAndEscape(label.Description, maxChars))
		sb.WriteString(`</td><td width="`)
		sb.WriteString(strconv.Itoa(hPadding))
		sb.WriteString(`"></td></tr></table></td>`)
		sb.WriteString(`</tr>`)
	}

	sb.WriteString(`</table>`)

	return sb.String()
}

// buildDbHTMLLabel generates an HTML table label for Database-type nodes.
// Format: name (bold) / [technology] italic / description
// Single-column layout without icon column.
// Note: DB nodes use cylinder shape which adds extra visual height for the 3D effect.
// We add horizontal padding to compensate and maintain the target aspect ratio.
func buildDbHTMLLabel(label *graph.Label) string {
	if label == nil {
		return ""
	}

	// Calculate max characters for word wrapping (no icon column)
	rowCount := 1 // name
	if label.Technology != "" {
		rowCount++
	}

	if label.Description != "" {
		rowCount++
	}

	// Use cylinder-specific calculation to compensate for shape's extra height
	maxChars := labelMaxCharsForCylinder(rowCount)

	// Minimal horizontal padding for visual breathing room
	totalHeight := rowCount * 18
	hPadding := totalHeight / 10 // 10% for minimal padding

	var sb strings.Builder
	sb.WriteString(`<table border="0" cellpadding="0" cellspacing="0">`)

	// Row 1: Name (bold) with horizontal padding
	sb.WriteString(`<tr align="center"><td valign="bottom">`)
	sb.WriteString(`<table border="0" cellpadding="0" cellspacing="0"><tr>`)
	sb.WriteString(`<td width="`)
	sb.WriteString(strconv.Itoa(hPadding))
	sb.WriteString(`"></td><td><b>`)
	sb.WriteString(wrapAndEscape(label.Name, maxChars))
	sb.WriteString(`</b></td><td width="`)
	sb.WriteString(strconv.Itoa(hPadding))
	sb.WriteString(`"></td></tr></table></td></tr>`)

	// Row 2: Technology (if present, italic in brackets)
	if label.Technology != "" {
		sb.WriteString(`<tr align="center"><td valign="middle">`)
		sb.WriteString(`<table border="0" cellpadding="0" cellspacing="0"><tr>`)
		sb.WriteString(`<td width="`)
		sb.WriteString(strconv.Itoa(hPadding))
		sb.WriteString(`"></td><td><i>[`)
		sb.WriteString(wrapAndEscape(label.Technology, maxChars))
		sb.WriteString(`]</i></td><td width="`)
		sb.WriteString(strconv.Itoa(hPadding))
		sb.WriteString(`"></td></tr></table></td></tr>`)
	}

	// Row 3: Description (if present)
	if label.Description != "" {
		sb.WriteString(`<tr align="center"><td valign="top">`)
		sb.WriteString(`<table border="0" cellpadding="0" cellspacing="0"><tr>`)
		sb.WriteString(`<td width="`)
		sb.WriteString(strconv.Itoa(hPadding))
		sb.WriteString(`"></td><td>`)
		sb.WriteString(wrapAndEscape(label.Description, maxChars))
		sb.WriteString(`</td><td width="`)
		sb.WriteString(strconv.Itoa(hPadding))
		sb.WriteString(`"></td></tr></table></td></tr>`)
	}

	sb.WriteString(`</table>`)

	return sb.String()
}

// buildQueueHTMLLabel generates an HTML table label for Queue-type nodes.
// Format: ASCII graphic / name (bold) / [technology] italic / description
// 4-row table with ASCII art graphic as first row.
// Note: Queue labels use horizontal padding to compensate for the ASCII graphic row
// which contributes to height but has limited width.
func buildQueueHTMLLabel(label *graph.Label) string {
	if label == nil {
		return ""
	}

	// Calculate max characters for word wrapping (no icon column)
	// Graphic row doesn't wrap but counts for proportion
	rowCount := 2 // graphic + name
	if label.Technology != "" {
		rowCount++
	}

	if label.Description != "" {
		rowCount++
	}

	maxChars := labelMaxCharsForQueue(rowCount)

	// Minimal horizontal padding for visual breathing room
	totalHeight := rowCount * 18
	hPadding := totalHeight / 10 // 10% for minimal padding

	var sb strings.Builder
	sb.WriteString(`<table border="0" cellpadding="0" cellspacing="0">`)

	// Row 1: ASCII art graphic (NOT wrapped, NOT escaped) with padding
	sb.WriteString(`<tr align="center"><td valign="middle">`)
	sb.WriteString(`<table border="0" cellpadding="0" cellspacing="0"><tr>`)
	sb.WriteString(`<td width="`)
	sb.WriteString(strconv.Itoa(hPadding))
	sb.WriteString(`"></td><td>`)
	sb.WriteString("═╦╩═╦═══")
	sb.WriteString(`</td><td width="`)
	sb.WriteString(strconv.Itoa(hPadding))
	sb.WriteString(`"></td></tr></table></td></tr>`)

	// Row 2: Name (bold) with horizontal padding
	sb.WriteString(`<tr align="center"><td valign="bottom">`)
	sb.WriteString(`<table border="0" cellpadding="0" cellspacing="0"><tr>`)
	sb.WriteString(`<td width="`)
	sb.WriteString(strconv.Itoa(hPadding))
	sb.WriteString(`"></td><td><b>`)
	sb.WriteString(wrapAndEscape(label.Name, maxChars))
	sb.WriteString(`</b></td><td width="`)
	sb.WriteString(strconv.Itoa(hPadding))
	sb.WriteString(`"></td></tr></table></td></tr>`)

	// Row 3: Technology (if present, italic in brackets)
	if label.Technology != "" {
		sb.WriteString(`<tr align="center"><td valign="middle">`)
		sb.WriteString(`<table border="0" cellpadding="0" cellspacing="0"><tr>`)
		sb.WriteString(`<td width="`)
		sb.WriteString(strconv.Itoa(hPadding))
		sb.WriteString(`"></td><td><i>[`)
		sb.WriteString(wrapAndEscape(label.Technology, maxChars))
		sb.WriteString(`]</i></td><td width="`)
		sb.WriteString(strconv.Itoa(hPadding))
		sb.WriteString(`"></td></tr></table></td></tr>`)
	}

	// Row 4: Description (if present)
	if label.Description != "" {
		sb.WriteString(`<tr align="center"><td valign="top">`)
		sb.WriteString(`<table border="0" cellpadding="0" cellspacing="0"><tr>`)
		sb.WriteString(`<td width="`)
		sb.WriteString(strconv.Itoa(hPadding))
		sb.WriteString(`"></td><td>`)
		sb.WriteString(wrapAndEscape(label.Description, maxChars))
		sb.WriteString(`</td><td width="`)
		sb.WriteString(strconv.Itoa(hPadding))
		sb.WriteString(`"></td></tr></table></td></tr>`)
	}

	sb.WriteString(`</table>`)

	return sb.String()
}

// buildSystemHTMLLabel generates an HTML table label for System-type nodes.
// Format: name (bold) / [technology] italic / description
// Single-column layout without icon column.
func buildSystemHTMLLabel(label *graph.Label) string {
	if label == nil {
		return ""
	}

	// Calculate max characters for word wrapping (no icon column)
	rowCount := 1 // name
	if label.Technology != "" {
		rowCount++
	}

	if label.Description != "" {
		rowCount++
	}

	maxChars := labelMaxCharsNoIcon(rowCount)

	var sb strings.Builder
	sb.WriteString(`<table border="0" cellpadding="0" cellspacing="0">`)

	// Row 1: Name (bold)
	sb.WriteString(`<tr align="center"><td valign="bottom"><b>`)
	sb.WriteString(wrapAndEscape(label.Name, maxChars))
	sb.WriteString(`</b></td></tr>`)

	// Row 2: Technology (if present, italic in brackets)
	if label.Technology != "" {
		sb.WriteString(`<tr align="center"><td valign="middle"><i>[`)
		sb.WriteString(wrapAndEscape(label.Technology, maxChars))
		sb.WriteString(`]</i></td></tr>`)
	}

	// Row 3: Description (if present)
	if label.Description != "" {
		sb.WriteString(`<tr align="center"><td valign="top">`)
		sb.WriteString(wrapAndEscape(label.Description, maxChars))
		sb.WriteString(`</td></tr>`)
	}

	sb.WriteString(`</table>`)

	return sb.String()
}

// buildContainerHTMLLabel generates an HTML table label for Container-type nodes.
// Format: name (bold) / [technology] italic / description
// Single-column layout without icon column.
func buildContainerHTMLLabel(label *graph.Label) string {
	if label == nil {
		return ""
	}

	// Calculate max characters for word wrapping (no icon column)
	rowCount := 1 // name
	if label.Technology != "" {
		rowCount++
	}

	if label.Description != "" {
		rowCount++
	}

	maxChars := labelMaxCharsNoIcon(rowCount)

	var sb strings.Builder
	sb.WriteString(`<table border="0" cellpadding="0" cellspacing="0">`)

	// Row 1: Name (bold)
	sb.WriteString(`<tr align="center"><td valign="bottom"><b>`)
	sb.WriteString(wrapAndEscape(label.Name, maxChars))
	sb.WriteString(`</b></td></tr>`)

	// Row 2: Technology (if present, italic in brackets)
	if label.Technology != "" {
		sb.WriteString(`<tr align="center"><td valign="middle"><i>[`)
		sb.WriteString(wrapAndEscape(label.Technology, maxChars))
		sb.WriteString(`]</i></td></tr>`)
	}

	// Row 3: Description (if present)
	if label.Description != "" {
		sb.WriteString(`<tr align="center"><td valign="top">`)
		sb.WriteString(wrapAndEscape(label.Description, maxChars))
		sb.WriteString(`</td></tr>`)
	}

	sb.WriteString(`</table>`)

	return sb.String()
}

// buildComponentHTMLLabel generates an HTML table label for Component-type nodes.
// Format: name (bold) / [technology] italic / description
// Single-column layout without icon column.
func buildComponentHTMLLabel(label *graph.Label) string {
	if label == nil {
		return ""
	}

	// Calculate max characters for word wrapping (no icon column)
	rowCount := 1 // name
	if label.Technology != "" {
		rowCount++
	}

	if label.Description != "" {
		rowCount++
	}

	maxChars := labelMaxCharsNoIcon(rowCount)

	var sb strings.Builder
	sb.WriteString(`<table border="0" cellpadding="0" cellspacing="0">`)

	// Row 1: Name (bold)
	sb.WriteString(`<tr align="center"><td valign="bottom"><b>`)
	sb.WriteString(wrapAndEscape(label.Name, maxChars))
	sb.WriteString(`</b></td></tr>`)

	// Row 2: Technology (if present, italic in brackets)
	if label.Technology != "" {
		sb.WriteString(`<tr align="center"><td valign="middle"><i>[`)
		sb.WriteString(wrapAndEscape(label.Technology, maxChars))
		sb.WriteString(`]</i></td></tr>`)
	}

	// Row 3: Description (if present)
	if label.Description != "" {
		sb.WriteString(`<tr align="center"><td valign="top">`)
		sb.WriteString(wrapAndEscape(label.Description, maxChars))
		sb.WriteString(`</td></tr>`)
	}

	sb.WriteString(`</table>`)

	return sb.String()
}

// buildBoxHTMLLabel generates an HTML table label for Box-type nodes.
// Format: name (bold) / [technology] italic / description
// Single-column layout without icon column.
// Output does NOT contain curly brackets {} unlike record labels.
func buildBoxHTMLLabel(label *graph.Label) string {
	if label == nil {
		return ""
	}

	// Calculate max characters for word wrapping (no icon column)
	rowCount := 1 // name
	if label.Technology != "" {
		rowCount++
	}

	if label.Description != "" {
		rowCount++
	}

	maxChars := labelMaxCharsNoIcon(rowCount)

	var sb strings.Builder
	sb.WriteString(`<table border="0" cellpadding="0" cellspacing="0">`)

	// Row 1: Name (bold)
	sb.WriteString(`<tr align="center"><td valign="bottom"><b>`)
	sb.WriteString(wrapAndEscape(label.Name, maxChars))
	sb.WriteString(`</b></td></tr>`)

	// Row 2: Technology (if present, italic in brackets)
	if label.Technology != "" {
		sb.WriteString(`<tr align="center"><td valign="middle"><i>[`)
		sb.WriteString(wrapAndEscape(label.Technology, maxChars))
		sb.WriteString(`]</i></td></tr>`)
	}

	// Row 3: Description (if present)
	if label.Description != "" {
		sb.WriteString(`<tr align="center"><td valign="top">`)
		sb.WriteString(wrapAndEscape(label.Description, maxChars))
		sb.WriteString(`</td></tr>`)
	}

	sb.WriteString(`</table>`)

	return sb.String()
}

// iconTypeForUnit maps unit types to icon names.
// This function is kept for backward compatibility but is no longer used
// since we replaced SVG icons with native GraphViz shapes and emoji.
// Deprecated: Will be removed in a future version.
//
//nolint:deadcode // Kept for reference
func iconTypeForUnit(_ model.UnitType) string {
	return "" // No longer needed
}
