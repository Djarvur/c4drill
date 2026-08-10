package render

import (
	"html"
	"math"
	"strings"
	"unicode"
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

// wrapText wraps text at word and punctuation boundaries to fit within
// maxChars per line. Uses "<BR/>" as line separator (GraphViz HTML label
// line break). Break opportunities are whitespace AND any punctuation or
// symbol character (per UAT: "any punctuation must be considered word
// boundary, not just spaces") — e.g. "Multi-Consumer", "YUV420->EXTERNAL",
// "IMAGE_NATIVE_PROCESSED" all wrap at their separators. Punctuation
// attaches to the word it belongs to, so rejoin on the same line never
// inserts a space. A token longer than maxChars (a pure letter/digit run
// with no separator) starts its own line and stays unsplit, overflowing
// the width (D-05): the document author may reword instead — the tool
// never splits inside a letter/digit run.
func wrapText(text string, maxChars int) string {
	if text == "" || maxChars <= 0 {
		return text
	}

	tokens := tokenizeWrapText(text)
	if len(tokens) == 0 {
		return ""
	}

	var lines []string

	var currentLine strings.Builder

	currentLen := 0

	for _, token := range tokens {
		tokenLen := utf8.RuneCountInString(token.text)

		// Token fits on current line (with space if whitespace-separated
		// from the previous token)
		if fitsOnLine(currentLen, tokenLen, maxChars, token.spaceBefore) {
			if token.spaceBefore {
				currentLine.WriteRune(' ')
			}

			currentLine.WriteString(token.text)

			currentLen += runeCountWithSpace(tokenLen, token.spaceBefore)

			continue
		}

		// Token fits as the start of a new line
		if tokenLen <= maxChars {
			lines = flushIfPending(lines, &currentLine, &currentLen)

			currentLine.WriteString(token.text)

			currentLen = tokenLen

			continue
		}

		// Token exceeds maxChars - emit it unsplit on its own line (D-05):
		// no character-level fallback, no safety cap. The document author
		// may reword instead.
		lines = flushIfPending(lines, &currentLine, &currentLen)

		currentLine.WriteString(token.text)

		currentLen = tokenLen
	}

	if currentLen > 0 {
		lines = append(lines, currentLine.String())
	}

	return strings.Join(lines, htmlLineBreak)
}

// wrapToken is a breakable unit of label text: a letter/digit run with any
// attached separator run (trailing attachment; leading attachment for a
// separator run at the start of the string).
type wrapToken struct {
	// text is the token content, including attached separators.
	text string
	// spaceBefore reports whether whitespace separated this token from the
	// previous one (a space is re-inserted when rejoined on the same line).
	spaceBefore bool
}

// tokenizeWrapText splits text into wrap tokens at whitespace and
// punctuation/symbol boundaries. A separator is any rune that is neither a
// letter/digit (unicode.IsLetter/IsDigit) nor whitespace (unicode.IsSpace) —
// note this covers Unicode symbols like '>' in "->", which unicode.IsPunct
// would miss. A separator run attaches to the word-like run that precedes
// it ("Multi-", "YUV420->", "IMAGE_"); a leading separator run attaches to
// the following word ("[CGF").
func tokenizeWrapText(text string) []wrapToken {
	runes := []rune(text)

	var tokens []wrapToken

	var current strings.Builder

	var pendingSpace bool

	hasWord := false

	lastWasSeparator := false

	flush := func(spaceBefore bool) {
		if current.Len() > 0 {
			tokens = append(tokens, wrapToken{text: current.String(), spaceBefore: spaceBefore})
			current.Reset()
		}
	}

	for i := 0; i < len(runes); {
		r := runes[i]

		switch {
		case isWordChar(r):
			// A word char after a trailing separator run ends the previous
			// token ("Multi-" | "Consumer") — unless the current token holds
			// only separators, which attach to this word ("[CGF").
			if hasWord && lastWasSeparator {
				flush(pendingSpace)
				pendingSpace = false
			}
			current.WriteRune(r)
			hasWord = true
			lastWasSeparator = false
			i++

		case unicode.IsSpace(r):
			// Whitespace ends the current token and marks the next one as
			// space-separated.
			flush(pendingSpace)
			pendingSpace = true
			hasWord = false
			lastWasSeparator = false
			// Consume the whole whitespace run.
			for i < len(runes) && unicode.IsSpace(runes[i]) {
				i++
			}

		default:
			// Separator (punctuation or symbol): attach to the current token
			// (trailing attachment), or start a pending token that will
			// attach to the following word (leading attachment).
			current.WriteRune(r)
			lastWasSeparator = true
			i++
		}
	}

	flush(pendingSpace)

	return tokens
}

// isWordChar reports whether r is part of a word-like run (letter or digit).
func isWordChar(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}

// runeCountWithSpace returns the line cost of a token: its rune count plus
// one when it is whitespace-separated and not the first token on the line.
func runeCountWithSpace(tokenLen int, spaceBefore bool) int {
	if spaceBefore {
		return 1 + tokenLen
	}

	return tokenLen
}

// fitsOnLine reports whether the token fits on the current line (with a
// preceding space when whitespace-separated and the line is non-empty).
func fitsOnLine(currentLen, tokenLen, maxChars int, spaceBefore bool) bool {
	if currentLen == 0 {
		return false
	}

	return currentLen+runeCountWithSpace(tokenLen, spaceBefore) <= maxChars
}

// flushIfPending appends the current line to lines and resets it when it is
// non-empty.
func flushIfPending(lines []string, currentLine *strings.Builder, currentLen *int) []string {
	if *currentLen == 0 {
		return lines
	}

	lines = append(lines, currentLine.String())

	currentLine.Reset()

	*currentLen = 0

	return lines
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

// labelMaxCharsForText computes the per-line character budget for a label
// whose rendered height is fixedRows content rows plus the wrapped text
// lines, such that the resulting rectangle's width:height approximates the
// given ratio (self-consistent aspect-ratio sizing, UAT 34-04).
//
// Closed-form solution (no iteration — a line-count loop oscillates because
// of ceil rounding): with width = (fixedRows + textLen·pointsPerChar/width)
// · pointsPerRow · ratio, the positive root is
//
//	width = (B + √(B² + 4C)) / 2,  B = fixedRows·pointsPerRow·ratio,
//	C = textLen·pointsPerChar·pointsPerRow·ratio
//
// Returns chars per line = width / pointsPerChar, floored at 6 so tiny
// labels never degenerate to a single-char column.
func labelMaxCharsForText(fixedRows, textLen int, ratio float64) int {
	b := float64(fixedRows) * pointsPerRow * ratio
	c := float64(textLen) * pointsPerChar * pointsPerRow * ratio
	width := (b + math.Sqrt(b*b+4*c)) / 2
	chars := int(width / pointsPerChar)
	if chars < 6 {
		return 6
	}

	return chars
}

// labelMaxCharsNoIcon calculates the maximum characters per line for a label
// without an icon column, so the rectangle's ratio approximates LabelRatio
// (self-consistent: fixed rows plus the wrapped description lines).
func labelMaxCharsNoIcon(fixedRows, textLen int) int {
	return labelMaxCharsForText(fixedRows, textLen, LabelRatio)
}

// labelMaxCharsForCylinder calculates the maximum characters per line for DB labels.
// DB nodes use cylinder shape which adds extra visual height for the 3D effect.
// Uses a higher effective ratio to fill the wider space and maintain target aspect ratio.
func labelMaxCharsForCylinder(fixedRows, textLen int) int {
	// Use higher ratio to compensate for cylinder shape overhead (~37% extra height)
	return labelMaxCharsForText(fixedRows, textLen, LabelRatio*2.2)
}

// labelMaxCharsForQueue calculates the maximum characters per line for Queue labels.
// Queue nodes have an ASCII graphic row that adds height but limited width.
// Uses a higher effective ratio to fill the wider space.
func labelMaxCharsForQueue(fixedRows, textLen int) int {
	// Use higher ratio to compensate for ASCII graphic row overhead
	return labelMaxCharsForText(fixedRows, textLen, LabelRatio*1.6)
}

// personRatioFactor compensates the person label's icon column (36pt of
// width that adds no rows) plus the token-packing overhead of wrapped
// descriptions, so the rendered rectangle approximates LabelRatio instead of
// collapsing toward square for long descriptions (UAT 34-05).
const personRatioFactor = 1.5

// labelMaxCharsForPerson calculates the maximum characters per line for Person labels.
// Icon-aware closed form: the 36pt icon column adds to the width but not to
// the row count, so the text column solves
//
//	textCol² + iconColumnWidth·textCol − C = 0,
//	C = textLen·pointsPerChar·pointsPerRow·LabelRatio·personRatioFactor
//
// (height rows = textLen·pointsPerChar/textCol).
func labelMaxCharsForPerson(fixedRows, textLen int) int {
	c := float64(textLen) * pointsPerChar * pointsPerRow * LabelRatio * personRatioFactor
	textCol := (-iconColumnWidth + math.Sqrt(float64(iconColumnWidth*iconColumnWidth)+4*c)) / 2
	chars := int(textCol / pointsPerChar)
	// Floor of 10 keeps short actor names on one line ("Actor A"); the
	// icon-aware formula takes over for longer descriptions.
	if chars < 10 {
		return 10
	}

	return chars
}
