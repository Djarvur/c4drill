package render

import (
	"bytes"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/Djarvur/c4drill/internal/graph"
)

const (
	iconPadding    = 4  // pixels from left edge
	iconSize       = 32 // width and height of icons
	dataAttrLength = 3  // length of `d="` attribute prefix
	halfDivider    = 2  // divide by 2 for centering
)

// InjectSVGIcons post-processes SVG output to inject icons as base64-encoded images.
// It finds nodes by their title and adds icon images at the appropriate positions.
func InjectSVGIcons(svgData []byte, g *graph.Graph, iconExtractor *IconExtractor) ([]byte, error) {
	if iconExtractor == nil || g == nil {
		return svgData, nil
	}

	svg := string(svgData)

	// Process top-level nodes
	for _, node := range g.Nodes {
		if node.Style != nil && node.Style.BorderColor != "" && node.Label != nil {
			svg = injectNodeIcon(svg, node, iconExtractor)
		}
	}

	// Process nodes in clusters recursively
	for _, cluster := range g.Clusters {
		svg = injectClusterIcons(svg, cluster, iconExtractor)
	}

	return []byte(svg), nil
}

// injectClusterIcons processes nodes and nested clusters within a cluster.
func injectClusterIcons(svg string, cluster *graph.Cluster, iconExtractor *IconExtractor) string {
	// Process nodes in this cluster
	for _, node := range cluster.Nodes {
		if node.Style != nil && node.Style.BorderColor != "" && node.Label != nil {
			svg = injectNodeIcon(svg, node, iconExtractor)
		}
	}

	// Process nested clusters
	for _, nested := range cluster.Clusters {
		svg = injectClusterIcons(svg, nested, iconExtractor)
	}

	return svg
}

// injectNodeIcon injects an icon into a node's SVG representation.
func injectNodeIcon(svg string, node *graph.Node, iconExtractor *IconExtractor) string {
	// Find the node's title element
	titleTag := fmt.Sprintf("<title>%s</title>", node.ID)

	titleIdx := strings.Index(svg, titleTag)
	if titleIdx == -1 {
		return svg
	}

	// Find the containing <g> element (go backwards to find the opening tag)
	gStart := strings.LastIndex(svg[:titleIdx], `<g id="node`)
	if gStart == -1 {
		return svg
	}

	// Find the first <path> element after the title
	pathStart := strings.Index(svg[titleIdx:], `<path`)
	if pathStart == -1 {
		return svg
	}

	pathStart += titleIdx

	// Find the end of the path element
	pathEnd := findPathEnd(svg, pathStart)
	if pathEnd == -1 {
		return svg
	}

	// Get icon type and color
	iconType := iconTypeForUnit(node.Type)

	hexColor := node.Style.BorderColor
	if !strings.HasPrefix(hexColor, "#") {
		hexColor = "#" + hexColor
	}

	hexClean := strings.TrimPrefix(hexColor, "#")

	// Get base64 icon data
	iconData, err := iconExtractor.ExtractSVGBase64(iconType, hexClean)
	if err != nil {
		return svg // Skip on error
	}

	// Extract bounding box from the path to position icon correctly
	left, top, _, bottom := extractPathBoundingBox(svg[pathStart:])
	if left == "" || top == "" || bottom == "" {
		return svg
	}

	// Find the first text element to align icon with text content
	textY := extractFirstTextY(svg[titleIdx:])
	if textY == "" {
		return svg // Skip if no text found
	}

	// Position icon on the left edge with small padding
	// Icon is aligned with the text content, constrained to node bounds
	x := addPadding(left, iconPadding)
	y := calculateIconYConstrained(top, bottom, textY, iconSize)

	// Create the image element
	imageEl := fmt.Sprintf(`<image href="%s" x="%s" y="%s" width="32" height="32" preserveAspectRatio="xMidYMid meet"/>`,
		iconData, x, y)

	// Insert image after the path element
	return svg[:pathEnd] + "\n" + imageEl + svg[pathEnd:]
}

// findPathEnd finds the end of a path element (handles both /> and </path>).
func findPathEnd(svg string, pathStart int) int {
	// Look for />
	selfClose := strings.Index(svg[pathStart:], "/>")
	if selfClose != -1 {
		return pathStart + selfClose + len("/>")
	}

	// Look for </path>
	fullClose := strings.Index(svg[pathStart:], "</path>")
	if fullClose != -1 {
		return pathStart + fullClose + len("</path>")
	}

	return -1
}

// extractPathBoundingBox extracts the bounding box coordinates from a node's path.
// Returns left, top, right, bottom coordinates.
// Path format is like: d="M77.88,-41.6C77.88,-41.6 12,-41.6 12,-41.6...".
func extractPathBoundingBox(pathStr string) (left, top, right, bottom string) {
	// Find d=" attribute
	dIdx := strings.Index(pathStr, `d="`)
	if dIdx == -1 {
		return "", "", "", ""
	}

	// Extract the path data
	dataStart := dIdx + dataAttrLength

	dataEnd := strings.Index(pathStr[dataStart:], `"`)
	if dataEnd == -1 {
		return "", "", "", ""
	}

	pathData := pathStr[dataStart : dataStart+dataEnd]

	// Parse all coordinate pairs from the path
	// Find min x (left), max x (right), min y (top - more negative), max y (bottom - less negative)
	minX, maxX := "", ""
	minY, maxY := "", ""

	// Extract numbers from path (including negative decimals)
	i := 0
	for i < len(pathData) {
		// Skip non-numeric characters
		for i < len(pathData) && !isDigitOrSign(pathData[i]) {
			i++
		}

		if i >= len(pathData) {
			break
		}

		// Extract number
		start := i
		for i < len(pathData) && (isDigitOrSign(pathData[i]) || pathData[i] == '.') {
			i++
		}

		num := pathData[start:i]

		// Skip comma if present
		if i < len(pathData) && pathData[i] == ',' {
			i++
		}

		// Extract second number (y coordinate)
		start = i
		for i < len(pathData) && (isDigitOrSign(pathData[i]) || pathData[i] == '.') {
			i++
		}

		num2 := pathData[start:i]

		// Update bounds
		if num != "" && num2 != "" {
			if minX == "" || compareFloats(num, minX) < 0 {
				minX = num
			}

			if maxX == "" || compareFloats(num, maxX) > 0 {
				maxX = num
			}

			if minY == "" || compareFloats(num2, minY) < 0 {
				minY = num2
			}

			if maxY == "" || compareFloats(num2, maxY) > 0 {
				maxY = num2
			}
		}
	}

	return minX, minY, maxX, maxY
}

// isDigitOrSign checks if a character is a digit or negative sign.
func isDigitOrSign(c byte) bool {
	return (c >= '0' && c <= '9') || c == '-' || c == '+'
}

// compareFloats compares two float strings numerically.
func compareFloats(a, b string) int {
	// Parse the actual float values for correct comparison
	aVal, _ := strconv.ParseFloat(a, 64)
	bVal, _ := strconv.ParseFloat(b, 64)

	if aVal < bVal {
		return -1
	} else if aVal > bVal {
		return 1
	}
	return 0
}

// extractFirstTextY finds the y position of the first text element in the SVG.
// This is used to align the icon with the text content, not the node center.
func extractFirstTextY(svg string) string {
	// Find the first <text element
	textIdx := strings.Index(svg, `<text`)
	if textIdx == -1 {
		return ""
	}

	// Extract the y attribute from the text element
	yIdx := strings.Index(svg[textIdx:], `y="`)
	if yIdx == -1 {
		return ""
	}

	// Calculate absolute position of y value
	// yIdx is relative to svg[textIdx:], so add textIdx to get absolute position
	yStart := textIdx + yIdx + 3 // 3 = len(`y="`)
	yEnd := strings.Index(svg[yStart:], `"`)
	if yEnd == -1 {
		return ""
	}

	return svg[yStart : yStart+yEnd]
}

// calculateIconYConstrained calculates the icon y position, aligning with text
// while ensuring the icon stays within node bounds.
// top and bottom are the node bounding box y coordinates.
// textY is the y position of the first text baseline.
// iconHeight is the height of the icon in pixels.
func calculateIconYConstrained(top, bottom, textY string, iconHeight int) string {
	// Parse coordinates
	topVal, _ := strconv.ParseFloat(top, 64)
	bottomVal, _ := strconv.ParseFloat(bottom, 64)
	textYVal, _ := strconv.ParseFloat(textY, 64)

	// Calculate ideal icon position (centered with text)
	textCenterOffset := 5.0 // Approximate offset from baseline to text visual center
	iconCenterOffset := float64(iconHeight) / halfDivider
	idealIconTopY := textYVal - iconCenterOffset - textCenterOffset

	// Calculate node height (in SVG, bottom > top since y increases downward)
	nodeHeight := bottomVal - topVal
	iconHeightFloat := float64(iconHeight)

	// Constrain icon to stay within node bounds
	minY := topVal + 2 // +2 pixel padding from node top
	maxY := bottomVal - iconHeightFloat - 2

	if iconHeightFloat <= nodeHeight {
		// Icon fits - constrain to node bounds
		if idealIconTopY < minY {
			idealIconTopY = minY
		} else if idealIconTopY > maxY {
			idealIconTopY = maxY
		}
	} else {
		// Icon doesn't fit - center vertically within node
		idealIconTopY = topVal + (nodeHeight-iconHeightFloat)/halfDivider
	}

	// Format result with 2 decimal places
	return strconv.FormatFloat(idealIconTopY, 'f', 2, 64)
}

// addPadding adds pixel padding to a coordinate value.
func addPadding(coord string, padding int) string {
	val, _ := strconv.ParseFloat(coord, 64)

	return strconv.FormatFloat(val+float64(padding), 'f', 2, 64)
}

// HasIcons checks if the graph has any nodes that need icons.
func HasIcons(g *graph.Graph) bool {
	if g == nil {
		return false
	}

	for _, node := range g.Nodes {
		if node.Style != nil && node.Style.BorderColor != "" && node.Label != nil {
			return true
		}
	}

	return slices.ContainsFunc(g.Clusters, clusterHasIcons)
}

func clusterHasIcons(c *graph.Cluster) bool {
	for _, node := range c.Nodes {
		if node.Style != nil && node.Style.BorderColor != "" && node.Label != nil {
			return true
		}
	}

	return slices.ContainsFunc(c.Clusters, clusterHasIcons)
}

// Verify the implementation satisfies the interface.
var _ = bytes.TrimSpace(nil) // Just to ensure bytes is imported if needed
