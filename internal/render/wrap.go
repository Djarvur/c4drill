package render

import (
	"html"
	"strings"
	"unicode/utf8"
)

const (
	// pointsPerChar is the approximate width of a character in points.
	pointsPerChar = 8

	// pointsPerRow is the approximate height of a label row in points (12pt font + 6pt padding).
	pointsPerRow = 18

	// iconColumnWidth is the fixed width of the icon column in points.
	iconColumnWidth = 36

	// minTextWidth is the minimum text column width in points.
	minTextWidth = 50

	// htmlLineBreak is the GraphViz HTML label line break tag.
	htmlLineBreak = "<BR/>"

	// minCharsPerLine is the minimum characters per line to prevent excessive wrapping of short text.
	minCharsPerLine = 20

	// defaultLabelRatio is the default width:height ratio for unit labels (credit card proportions).
	defaultLabelRatio = 1.6
)

// LabelRatio is the width:height ratio for unit labels.
// Set by CLI before rendering. Default is 1.6 (8/5, approximately credit card proportions).
//
//nolint:gochecknoglobals // CLI configuration set before render calls
var LabelRatio = defaultLabelRatio

// wrapText wraps text at word boundaries to fit within maxChars per line.
// Uses "<BR/>" as line separator (GraphViz HTML label line break).
// Falls back to character-level breaking for words exceeding maxChars.
func wrapText(text string, maxChars int) string {
	if text == "" || maxChars <= 0 {
		return text
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return ""
	}

	var lines []string

	var currentLine strings.Builder

	currentLen := 0

	for _, word := range words {
		wordLen := utf8.RuneCountInString(word)

		// Word fits on current line (with space if not first word on line)
		if currentLen > 0 && currentLen+1+wordLen <= maxChars {
			currentLine.WriteRune(' ')
			currentLine.WriteString(word)
			currentLen += 1 + wordLen

			continue
		}

		// Word fits as the start of a new line
		if wordLen <= maxChars {
			if currentLen > 0 {
				lines = append(lines, currentLine.String())
				currentLine.Reset()
			}

			currentLine.WriteString(word)
			currentLen = wordLen

			continue
		}

		// Word exceeds maxChars - force character-level break
		if currentLen > 0 {
			lines = append(lines, currentLine.String())
			currentLine.Reset()
			currentLen = 0
		}

		runes := []rune(word)
		for len(runes) > 0 {
			chunkSize := maxChars
			if chunkSize > len(runes) {
				chunkSize = len(runes)
			}

			if currentLen > 0 {
				lines = append(lines, currentLine.String())
				currentLine.Reset()
			}

			currentLine.WriteString(string(runes[:chunkSize]))
			currentLen = chunkSize
			runes = runes[chunkSize:]
		}
	}

	if currentLen > 0 {
		lines = append(lines, currentLine.String())
	}

	return strings.Join(lines, htmlLineBreak)
}

// wrapAndEscape wraps text at word boundaries, then HTML-escapes each line.
// Returns HTML with <BR/> line breaks suitable for GraphViz HTML labels.
func wrapAndEscape(text string, maxChars int) string {
	if text == "" {
		return ""
	}

	wrapped := wrapText(text, maxChars)

	// Split by <BR/>, escape each part, rejoin
	parts := strings.Split(wrapped, htmlLineBreak)
	for i, part := range parts {
		parts[i] = html.EscapeString(part)
	}

	return strings.Join(parts, htmlLineBreak)
}

// estimateCharsFromWidth estimates how many characters fit in the given width (in points).
func estimateCharsFromWidth(widthPoints int) int {
	if widthPoints <= 0 {
		return 0
	}

	return widthPoints / pointsPerChar
}

// calculateTextWidth calculates the text column width in points for a label.
// rowCount is the number of content rows, ratio is the desired width:height ratio.
// Returns width in points, minimum 50 for readability.
func calculateTextWidth(rowCount int, ratio float64) int {
	totalHeight := rowCount * pointsPerRow
	totalWidth := int(float64(totalHeight) * ratio)
	textWidth := totalWidth - iconColumnWidth

	if textWidth < minTextWidth {
		return minTextWidth
	}

	return textWidth
}

// labelMaxChars calculates the maximum characters per line for a label
// based on the number of content rows and the current LabelRatio.
func labelMaxChars(rowCount int) int {
	textWidth := calculateTextWidth(rowCount, LabelRatio)
	chars := estimateCharsFromWidth(textWidth)

	if chars < minCharsPerLine {
		return minCharsPerLine
	}

	return chars
}
