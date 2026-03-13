package render

import (
	"html"
	"strconv"
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
// Format: icon (rowspan=2) | name (bold) / description
// Per CONTEXT.md: Person has NO technology field.
func buildPersonHTMLLabel(label *graph.Label) string {
	if label == nil {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("<<table>")

	// Calculate rowspan for icon: always 2 (name + description) if description present, else 1
	rowspan := 1
	if label.Description != "" {
		rowspan = 2
	}

	// Row 1: Icon (rowspan) + Name
	sb.WriteString(`<tr align="center">`)
	sb.WriteString(`<td rowspan="`)
	sb.WriteString(strconv.Itoa(rowspan))
	sb.WriteString(`" valign="middle"><font point-size="20">`)
	sb.WriteString("\U0001F464") // person emoji
	sb.WriteString(`</font></td>`)
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

	sb.WriteString(`</table>>`)

	return sb.String()
}

// buildDbHTMLLabel generates an HTML table label for Database-type nodes.
// Format: icon (rowspan) | name (bold) / [technology] italic / description
// Per CONTEXT.md: icon is U+26C1 (golf flag in hole), used as cylinder proxy.
func buildDbHTMLLabel(label *graph.Label) string {
	if label == nil {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("<<table>")

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
	sb.WriteString(`<td rowspan="`)
	sb.WriteString(strconv.Itoa(rowspan))
	sb.WriteString(`" valign="middle"><font point-size="20">`)
	sb.WriteString("\u26C1") // db cylinder icon (using golf flag as proxy)
	sb.WriteString(`</font></td>`)
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

	sb.WriteString(`</table>>`)

	return sb.String()
}

// buildQueueHTMLLabel generates an HTML table label for Queue-type nodes.
// Format: 4 separate rows (NO rowspan): graphics, name (bold), [technology] italic, description
// Per CONTEXT.md: Queue has NO rowspan - 4 separate single-cell rows.
// Graphics: ═╦╩═╦══ (Unicode box drawing characters)
func buildQueueHTMLLabel(label *graph.Label) string {
	if label == nil {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("<<table>")

	// Row 1: Graphics (NO rowspan - separate row)
	sb.WriteString(`<tr align="center">`)
	sb.WriteString(`<td valign="middle">`)
	sb.WriteString("═╦╩═╦══") // queue graphics
	sb.WriteString(`</td>`)
	sb.WriteString(`</tr>`)

	// Row 2: Name (bold)
	sb.WriteString(`<tr align="center">`)
	sb.WriteString(`<td valign="bottom"><b>`)
	sb.WriteString(html.EscapeString(label.Name))
	sb.WriteString(`</b></td>`)
	sb.WriteString(`</tr>`)

	// Row 3: Technology (if present)
	if label.Technology != "" {
		sb.WriteString(`<tr align="center">`)
		sb.WriteString(`<td valign="middle"><i>[`)
		sb.WriteString(html.EscapeString(label.Technology))
		sb.WriteString(`]</i></td>`)
		sb.WriteString(`</tr>`)
	}

	// Row 4: Description (if present)
	if label.Description != "" {
		sb.WriteString(`<tr align="center">`)
		sb.WriteString(`<td valign="top">`)
		sb.WriteString(html.EscapeString(label.Description))
		sb.WriteString(`</td>`)
		sb.WriteString(`</tr>`)
	}

	sb.WriteString(`</table>>`)

	return sb.String()
}

// buildSystemHTMLLabel generates an HTML table label for System-type nodes.
// Format: SYS label (rowspan, monospace) | name (bold) / [technology] italic / description
// Per CONTEXT.md: SYS label in <tt> tags.
func buildSystemHTMLLabel(label *graph.Label) string {
	if label == nil {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("<<table>")

	// Calculate rowspan for SYS label: count of present fields (name always present)
	rowspan := 1 // name
	if label.Technology != "" {
		rowspan++
	}
	if label.Description != "" {
		rowspan++
	}

	// Row 1: SYS (rowspan) + Name
	sb.WriteString(`<tr align="center">`)
	sb.WriteString(`<td rowspan="`)
	sb.WriteString(strconv.Itoa(rowspan))
	sb.WriteString(`" valign="middle"><tt>SYS</tt></td>`)
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

	sb.WriteString(`</table>>`)

	return sb.String()
}

// buildContainerHTMLLabel generates an HTML table label for Container-type nodes.
// Format: CONT label (rowspan, monospace) | name (bold) / [technology] italic / description
// Per CONTEXT.md: CONT label in <tt> tags. Also used for Box type.
func buildContainerHTMLLabel(label *graph.Label) string {
	if label == nil {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("<<table>")

	// Calculate rowspan for CONT label: count of present fields (name always present)
	rowspan := 1 // name
	if label.Technology != "" {
		rowspan++
	}
	if label.Description != "" {
		rowspan++
	}

	// Row 1: CONT (rowspan) + Name
	sb.WriteString(`<tr align="center">`)
	sb.WriteString(`<td rowspan="`)
	sb.WriteString(strconv.Itoa(rowspan))
	sb.WriteString(`" valign="middle"><tt>CONT</tt></td>`)
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

	sb.WriteString(`</table>>`)

	return sb.String()
}

// buildComponentHTMLLabel generates an HTML table label for Component-type nodes.
// Format: COMP label (rowspan, monospace) | name (bold) / [technology] italic / description
// Per CONTEXT.md: COMP label in <tt> tags.
func buildComponentHTMLLabel(label *graph.Label) string {
	if label == nil {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("<<table>")

	// Calculate rowspan for COMP label: count of present fields (name always present)
	rowspan := 1 // name
	if label.Technology != "" {
		rowspan++
	}
	if label.Description != "" {
		rowspan++
	}

	// Row 1: COMP (rowspan) + Name
	sb.WriteString(`<tr align="center">`)
	sb.WriteString(`<td rowspan="`)
	sb.WriteString(strconv.Itoa(rowspan))
	sb.WriteString(`" valign="middle"><tt>COMP</tt></td>`)
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

	sb.WriteString(`</table>>`)

	return sb.String()
}
