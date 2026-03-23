package render

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Djarvur/c4drill/internal/graph"
)

const iconPlaceholderWidth = 36

// InjectDOTIconPlaceholders post-processes DOT output to insert icon placeholder
// cells into labels for nodes/clusters that would have icons.
// This is needed because the WASM Graphviz engine has limited memory and cannot
// handle 2-column HTML table labels with rowspan for complex graphs.
// Instead, we render with simple 1-column labels, then inject the icon cells
// into the DOT text output afterward.
func InjectDOTIconPlaceholders(dotData []byte, g *graph.Graph) []byte {
	if g == nil {
		return dotData
	}

	dot := string(dotData)

	// Process top-level nodes
	for _, node := range g.Nodes {
		if node.Style != nil && node.Style.BorderColor != "" && node.Label != nil {
			dot = injectNodeDOTPlaceholder(dot, node)
		}
	}

	// Process clusters recursively
	for _, cluster := range g.Clusters {
		dot = injectClusterDOTPlaceholders(dot, cluster)
	}

	return []byte(dot)
}

// injectClusterDOTPlaceholders processes nodes and nested clusters within a cluster.
func injectClusterDOTPlaceholders(dot string, cluster *graph.Cluster) string {
	// Process nodes in this cluster
	for _, node := range cluster.Nodes {
		if node.Style != nil && node.Style.BorderColor != "" && node.Label != nil {
			dot = injectNodeDOTPlaceholder(dot, node)
		}
	}

	// Process nested clusters
	for _, nested := range cluster.Clusters {
		dot = injectClusterDOTPlaceholders(dot, nested)
	}

	return dot
}

// injectNodeDOTPlaceholder injects an icon placeholder cell into a node's HTML label.
func injectNodeDOTPlaceholder(dot string, node *graph.Node) string {
	// Find node reference in DOT text: "node.id" [... label=<<table...>>]
	// The ID is quoted in the DOT output
	nodeRef := fmt.Sprintf(`"%s"`, node.ID)

	refIdx := strings.Index(dot, nodeRef)
	if refIdx == -1 {
		return dot
	}

	// Find label within this node's attribute block
	labelStart := strings.Index(dot[refIdx:], "label=<")
	if labelStart == -1 {
		return dot
	}

	labelStart += refIdx

	// Find the end of the HTML label (matching >)
	labelEnd := findHTMLLabelEnd(dot, labelStart+len("label=<"))
	if labelEnd == -1 {
		return dot
	}

	// Extract the label content
	htmlStart := labelStart + len("label=<")
	labelHTML := dot[htmlStart:labelEnd]

	// Modify the label based on node type
	var modified string

	if graph.IsQueueType(node.Type) {
		modified = addQueueIconPlaceholder(labelHTML)
	} else {
		modified = addRowspanIconPlaceholder(labelHTML)
	}

	if modified == labelHTML {
		return dot // No change needed
	}

	return dot[:htmlStart] + modified + dot[labelEnd:]
}

// addRowspanIconPlaceholder adds a rowspan icon placeholder cell to a standard label.
// Transforms:
//
//	<table><tr align="center"><td>Name</td></tr><tr>...</tr></table>
//
// To:
//
//	<table><tr align="center"><td rowspan="N" width="36" valign="middle"></td><td>Name</td></tr><tr>...</tr></table>
func addRowspanIconPlaceholder(labelHTML string) string {
	// Count rows to determine rowspan
	rowCount := strings.Count(labelHTML, "<tr")
	if rowCount == 0 {
		return labelHTML
	}

	// Find the first <tr align="center"><td
	firstTREnd := strings.Index(labelHTML, `<tr align="center">`)
	if firstTREnd == -1 {
		return labelHTML
	}

	insertPoint := firstTREnd + len(`<tr align="center">`)

	// Find the first <td after the <tr>
	firstTD := strings.Index(labelHTML[insertPoint:], "<td")
	if firstTD == -1 {
		return labelHTML
	}

	insertAt := insertPoint + firstTD

	// Build the placeholder cell
	placeholder := `<td rowspan="` + strconv.Itoa(rowCount) +
		`" width="` + strconv.Itoa(iconPlaceholderWidth) +
		`" valign="middle"></td>`

	return labelHTML[:insertAt] + placeholder + labelHTML[insertAt:]
}

// addQueueIconPlaceholder adds width to the existing empty icon cell in a Queue label.
// Queue labels have a separate icon row: <tr align="center"><td valign="middle"></td></tr>
// Transforms:
//
//	<td valign="middle"></td>
//
// To:
//
//	<td width="36" valign="middle"></td>
func addQueueIconPlaceholder(labelHTML string) string {
	old := `<td valign="middle"></td>`
	newCell := `<td width="` + strconv.Itoa(iconPlaceholderWidth) + `" valign="middle"></td>`

	return strings.Replace(labelHTML, old, newCell, 1)
}

// findHTMLLabelEnd finds the closing > of an HTML label (label=<...>).
// HTML labels can contain nested < and > inside HTML tags, so we need to
// find the > that closes the label at the same nesting level.
func findHTMLLabelEnd(dot string, start int) int {
	depth := 1
	i := start

	for i < len(dot) && depth > 0 {
		switch dot[i] {
		case '<':
			depth++
		case '>':
			depth--

			if depth == 0 {
				return i
			}
		}

		i++
	}

	return -1
}
