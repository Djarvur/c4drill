package render

import (
	"html"
	"strconv"
	"strings"

	"github.com/Djarvur/c4drill/internal/graph"
	"github.com/Djarvur/c4drill/internal/model"
	"github.com/Djarvur/c4drill/internal/render/icons"
)

// iconReserve is a sentinel value used as iconRelPath to signal that the icon
// cell should be allocated (with width="36") but no <img> tag should be emitted.
// Used for SVG rendering where the layout must reserve space for post-render
// icon injection.
const iconReserve = "\x00icon"

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

// iconTypeForUnit maps unit types to icon names.
// Per D-07: All 6 types get icons, Box uses container icon.
func iconTypeForUnit(t model.UnitType) string {
	switch {
	case graph.IsPersonType(t):
		return icons.TypePerson
	case graph.IsDbType(t):
		return icons.TypeDb
	case graph.IsQueueType(t):
		return icons.TypePipe
	case graph.IsSystemType(t):
		return icons.TypeSystem
	case graph.IsContainerType(t), t == model.TypeBox:
		return icons.TypeContainer
	case graph.IsComponentType(t):
		return icons.TypeComponent
	default:
		return icons.TypeContainer // Fallback
	}
}

// HTML Label Builder Functions (Wave 1 implementation)
// These stub functions are placeholders that will be implemented in plan 12-01.
// They return empty strings so tests compile but fail assertions.

// buildPersonHTMLLabel generates an HTML table label for Person-type nodes.
// Format: icon (rowspan=2) | name (bold) / description
// Per CONTEXT.md: Person has NO technology field.
func buildPersonHTMLLabel(label *graph.Label, iconRelPath string) string {
	if label == nil {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(`<table border="0" cellpadding="0" cellspacing="0">`)

	// Calculate rowspan for icon: always 2 (name + description) if description present, else 1
	rowspan := 1
	if label.Description != "" {
		rowspan = 2
	}

	// Row 1: Icon (rowspan) + Name
	sb.WriteString(`<tr align="center">`)

	if iconRelPath != "" {
		sb.WriteString(`<td width="36" rowspan="`)
		sb.WriteString(strconv.Itoa(rowspan))
		sb.WriteString(`" valign="middle">`)

		if iconRelPath != iconReserve {
			sb.WriteString(`<img src="`)
			sb.WriteString(iconRelPath)
			sb.WriteString(`" width="32" height="32"/>`)
		}

		sb.WriteString(`</td>`)
	}

	sb.WriteString(`<td valign="bottom"><b>`)
	sb.WriteString(html.EscapeString(label.Name))
	sb.WriteString(`</b></td>`)
	sb.WriteString(`</tr>`)

	// Row 2: Description (if present)
	if label.Description != "" {
		sb.WriteString(`<tr align="center">`)
		sb.WriteString(`<td valign="top">`)
		sb.WriteString(html.EscapeString(label.Description))
		sb.WriteString(`</td>`)
		sb.WriteString(`</tr>`)
	}

	sb.WriteString(`</table>`)

	return sb.String()
}

// buildDbHTMLLabel generates an HTML table label for Database-type nodes.
// Format: icon (rowspan) | name (bold) / [technology] italic / description
// Per CONTEXT.md: icon is U+26C1 (golf flag in hole). used as cylinder proxy.
func buildDbHTMLLabel(label *graph.Label, iconRelPath string) string {
	if label == nil {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(`<table border="0" cellpadding="0" cellspacing="0">`)

	// Calculate rowspan for icon: count of present fields (name always present)
	rowspan := 1 // name
	if label.Technology != "" {
		rowspan++
	}

	if label.Description != "" {
		rowspan++
	}

	// Row 1: Icon (rowspan) + Name
	sb.WriteString(`<tr align="center">`)

	if iconRelPath != "" {
		sb.WriteString(`<td width="36" rowspan="`)
		sb.WriteString(strconv.Itoa(rowspan))
		sb.WriteString(`" valign="middle">`)

		if iconRelPath != iconReserve {
			sb.WriteString(`<img src="`)
			sb.WriteString(iconRelPath)
			sb.WriteString(`" width="32" height="32"/>`)
		}

		sb.WriteString(`</td>`)
	}

	sb.WriteString(`<td valign="bottom"><b>`)
	sb.WriteString(html.EscapeString(label.Name))
	sb.WriteString(`</b></td>`)
	sb.WriteString(`</tr>`)

	// Row 2: Technology (if present)
	if label.Technology != "" {
		sb.WriteString(`<tr align="center">`)
		sb.WriteString(`<td valign="middle"><i>[`)
		sb.WriteString(html.EscapeString(label.Technology))
		sb.WriteString(`]</i></td>`)
		sb.WriteString(`</tr>`)
	}

	// Row 3: Description (if present)
	if label.Description != "" {
		sb.WriteString(`<tr align="center">`)
		sb.WriteString(`<td valign="top">`)
		sb.WriteString(html.EscapeString(label.Description))
		sb.WriteString(`</td>`)
		sb.WriteString(`</tr>`)
	}

	sb.WriteString(`</table>`)

	return sb.String()
}

// buildQueueHTMLLabel generates an HTML table label for Queue-type nodes.
// Format: icon (rowspan) | name (bold) / [technology] italic / description
// Same layout as other unit types.
func buildQueueHTMLLabel(label *graph.Label, iconRelPath string) string {
	if label == nil {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(`<table border="0" cellpadding="0" cellspacing="0">`)

	// Calculate rowspan for icon: count of present fields (name always present)
	rowspan := 1 // name
	if label.Technology != "" {
		rowspan++
	}

	if label.Description != "" {
		rowspan++
	}

	// Row 1: Icon (rowspan) + Name
	sb.WriteString(`<tr align="center">`)

	if iconRelPath != "" {
		sb.WriteString(`<td width="36" rowspan="`)
		sb.WriteString(strconv.Itoa(rowspan))
		sb.WriteString(`" valign="middle">`)

		if iconRelPath != iconReserve {
			sb.WriteString(`<img src="`)
			sb.WriteString(iconRelPath)
			sb.WriteString(`" width="32" height="32"/>`)
		}

		sb.WriteString(`</td>`)
	}

	sb.WriteString(`<td valign="bottom"><b>`)
	sb.WriteString(html.EscapeString(label.Name))
	sb.WriteString(`</b></td>`)
	sb.WriteString(`</tr>`)

	// Row 2: Technology (if present)
	if label.Technology != "" {
		sb.WriteString(`<tr align="center">`)
		sb.WriteString(`<td valign="middle"><i>[`)
		sb.WriteString(html.EscapeString(label.Technology))
		sb.WriteString(`]</i></td>`)
		sb.WriteString(`</tr>`)
	}

	// Row 3: Description (if present)
	if label.Description != "" {
		sb.WriteString(`<tr align="center">`)
		sb.WriteString(`<td valign="top">`)
		sb.WriteString(html.EscapeString(label.Description))
		sb.WriteString(`</td>`)
		sb.WriteString(`</tr>`)
	}

	sb.WriteString(`</table>`)

	return sb.String()
}

// buildSystemHTMLLabel generates an HTML table label for System-type nodes.
// Format: icon (rowspan) | name (bold) / [technology] italic / description
// Per CONTEXT.md: icon replaces SYS label.
func buildSystemHTMLLabel(label *graph.Label, iconRelPath string) string {
	if label == nil {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(`<table border="0" cellpadding="0" cellspacing="0">`)

	// Calculate rowspan for icon: count of present fields (name always present)
	rowspan := 1 // name
	if label.Technology != "" {
		rowspan++
	}

	if label.Description != "" {
		rowspan++
	}

	// Row 1: Icon (rowspan) + Name
	sb.WriteString(`<tr align="center">`)

	if iconRelPath != "" {
		sb.WriteString(`<td width="36" rowspan="`)
		sb.WriteString(strconv.Itoa(rowspan))
		sb.WriteString(`" valign="middle">`)

		if iconRelPath != iconReserve {
			sb.WriteString(`<img src="`)
			sb.WriteString(iconRelPath)
			sb.WriteString(`" width="32" height="32"/>`)
		}

		sb.WriteString(`</td>`)
	}

	sb.WriteString(`<td valign="bottom"><b>`)
	sb.WriteString(html.EscapeString(label.Name))
	sb.WriteString(`</b></td>`)
	sb.WriteString(`</tr>`)

	// Row 2: Technology (if present)
	if label.Technology != "" {
		sb.WriteString(`<tr align="center">`)
		sb.WriteString(`<td valign="middle"><i>[`)
		sb.WriteString(html.EscapeString(label.Technology))
		sb.WriteString(`]</i></td>`)
		sb.WriteString(`</tr>`)
	}

	// Row 3: Description (if present)
	if label.Description != "" {
		sb.WriteString(`<tr align="center">`)
		sb.WriteString(`<td valign="top">`)
		sb.WriteString(html.EscapeString(label.Description))
		sb.WriteString(`</td>`)
		sb.WriteString(`</tr>`)
	}

	sb.WriteString(`</table>`)

	return sb.String()
}

// buildContainerHTMLLabel generates an HTML table label for Container-type nodes.
// Format: icon (rowspan) | name (bold) / [technology] italic / description
// Per CONTEXT.md: icon replaces CONT label. Also used for Box type.
func buildContainerHTMLLabel(label *graph.Label, iconRelPath string) string {
	if label == nil {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(`<table border="0" cellpadding="0" cellspacing="0">`)

	// Calculate rowspan for icon: count of present fields (name always present)
	rowspan := 1 // name
	if label.Technology != "" {
		rowspan++
	}

	if label.Description != "" {
		rowspan++
	}

	// Row 1: Icon (rowspan) + Name
	sb.WriteString(`<tr align="center">`)

	if iconRelPath != "" {
		sb.WriteString(`<td width="36" rowspan="`)
		sb.WriteString(strconv.Itoa(rowspan))
		sb.WriteString(`" valign="middle">`)

		if iconRelPath != iconReserve {
			sb.WriteString(`<img src="`)
			sb.WriteString(iconRelPath)
			sb.WriteString(`" width="32" height="32"/>`)
		}

		sb.WriteString(`</td>`)
	}

	sb.WriteString(`<td valign="bottom"><b>`)
	sb.WriteString(html.EscapeString(label.Name))
	sb.WriteString(`</b></td>`)
	sb.WriteString(`</tr>`)

	// Row 2: Technology (if present)
	if label.Technology != "" {
		sb.WriteString(`<tr align="center">`)
		sb.WriteString(`<td valign="middle"><i>[`)
		sb.WriteString(html.EscapeString(label.Technology))
		sb.WriteString(`]</i></td>`)
		sb.WriteString(`</tr>`)
	}

	// Row 3: Description (if present)
	if label.Description != "" {
		sb.WriteString(`<tr align="center">`)
		sb.WriteString(`<td valign="top">`)
		sb.WriteString(html.EscapeString(label.Description))
		sb.WriteString(`</td>`)
		sb.WriteString(`</tr>`)
	}

	sb.WriteString(`</table>`)

	return sb.String()
}

// buildComponentHTMLLabel generates an HTML table label for Component-type nodes.
// Format: icon (rowspan) | name (bold) / [technology] italic / description
// Per CONTEXT.md: icon replaces COMP label.
func buildComponentHTMLLabel(label *graph.Label, iconRelPath string) string {
	if label == nil {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(`<table border="0" cellpadding="0" cellspacing="0">`)

	// Calculate rowspan for icon: count of present fields (name always present)
	rowspan := 1 // name
	if label.Technology != "" {
		rowspan++
	}

	if label.Description != "" {
		rowspan++
	}

	// Row 1: Icon (rowspan) + Name
	sb.WriteString(`<tr align="center">`)

	if iconRelPath != "" {
		sb.WriteString(`<td width="36" rowspan="`)
		sb.WriteString(strconv.Itoa(rowspan))
		sb.WriteString(`" valign="middle">`)

		if iconRelPath != iconReserve {
			sb.WriteString(`<img src="`)
			sb.WriteString(iconRelPath)
			sb.WriteString(`" width="32" height="32"/>`)
		}

		sb.WriteString(`</td>`)
	}

	sb.WriteString(`<td valign="bottom"><b>`)
	sb.WriteString(html.EscapeString(label.Name))
	sb.WriteString(`</b></td>`)
	sb.WriteString(`</tr>`)

	// Row 2: Technology (if present)
	if label.Technology != "" {
		sb.WriteString(`<tr align="center">`)
		sb.WriteString(`<td valign="middle"><i>[`)
		sb.WriteString(html.EscapeString(label.Technology))
		sb.WriteString(`]</i></td>`)
		sb.WriteString(`</tr>`)
	}

	// Row 3: Description (if present)
	if label.Description != "" {
		sb.WriteString(`<tr align="center">`)
		sb.WriteString(`<td valign="top">`)
		sb.WriteString(html.EscapeString(label.Description))
		sb.WriteString(`</td>`)
		sb.WriteString(`</tr>`)
	}

	sb.WriteString(`</table>`)

	return sb.String()
}
