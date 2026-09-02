// text.go holds document text utilities: LSP position ⇄ byte offset
// conversion (UTF-16 aware, the LSP wire encoding) and ranged-edit splicing.

package lsp

import "unicode/utf16"

// lineStarts returns the byte offset at which every line begins (line 0
// starts at 0; a trailing newline starts an empty final line, matching LSP).
func lineStarts(text string) []int {
	starts := make([]int, 1, 64)

	for i := range len(text) {
		if text[i] == '\n' {
			starts = append(starts, i+1)
		}
	}

	return starts
}

// offsetForPosition converts an LSP Position (0-based line, UTF-16 character
// column) to a byte offset in text. Positions past the end of the text (or
// of their line) clamp to the nearest legal boundary.
func offsetForPosition(text string, pos Position) int {
	starts := lineStarts(text)
	if int(pos.Line) >= len(starts) {
		return len(text)
	}

	base := starts[pos.Line]

	end := len(text)
	if int(pos.Line)+1 < len(starts) {
		end = starts[pos.Line+1] - 1 // exclude the newline
	}

	units := 0

	seg := text[base:end]
	for i, r := range seg {
		u := utf16.RuneLen(r)
		if u < 0 {
			u = 1
		}

		// The position lands within this rune's UTF-16 span (including
		// inside a surrogate pair): clamp to the rune's start.
		if units+u > int(pos.Character) {
			return base + i
		}

		units += u
	}

	return end
}

// spliceText applies one ranged replacement: bytes in [start,end) are cut,
// insert is stitched in. A collapsed or inverted range degrades to an
// insertion at start.
func spliceText(text []byte, r Range, insert string) []byte {
	s := string(text)
	start := offsetForPosition(s, r.Start)
	end := offsetForPosition(s, r.End)

	if end < start {
		end = start
	}

	return []byte(s[:start] + insert + s[end:])
}
