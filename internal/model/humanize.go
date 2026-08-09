package model

import (
	"strings"
	"unicode"
)

// Humanize derives a readable display name from a single TOML identifier
// segment (camelCase or PascalCase) via a dumb split on case boundaries.
// It Title-cases each resulting word and joins them with single spaces.
//
// Per ERGO-04, acronym preservation is an ANTI-feature: there is no acronym
// allowlist. The dumb split does, however, leave a TRAILING all-uppercase run
// intact (so "localIDP" → "Local IDP", "sessionAPI" → "Session API"), while a
// leading or mid-string upper-run is lowercased ("IDPToken" → "Idp Token") and
// a lone leading lowercase glued to an upper-run stays one word ("gRPC" →
// "Grpc"). Authors who want full control supply an explicit `name =` in TOML,
// which always wins (ERGO-05).
//
// Phase 31's XC-04 will relocate the caller from the parse-time hook to a
// post-template-expansion pass; this function's contract is the stable
// artifact that relocation reuses unchanged.
//
// Empty input returns empty output. Boundaries are inserted on:
//   - lower→upper transition when the preceding lowercase run has length ≥ 2
//     (so "linuxSystem" → "Linux System", but the single leading lowercase in
//     "gRPC" stays glued to the upper-run as one word → "Grpc");
//   - upper→upper→lower transition, so the last upper of a run begins the next
//     word ("IDPToken" → "Idp Token").
//
// Each word is lowercased and its first rune capitalized, EXCEPT a trailing
// word that is entirely uppercase, which is preserved verbatim.
func Humanize(segment string) string {
	if segment == "" {
		return ""
	}

	runes := []rune(segment)
	words := splitWords(runes)

	parts := make([]string, len(words))
	for i, w := range words {
		parts[i] = titleWord(w, isLastPureUpper(w, i == len(words)-1))
	}

	return strings.Join(parts, " ")
}

// splitWords slices the runes into words at the case boundaries described in
// Humanize's godoc.
func splitWords(runes []rune) [][]rune {
	var words [][]rune

	start := 0
	lowerRun := 0 // length of the current trailing lowercase run

	for i := 1; i < len(runes); i++ {
		prev := runes[i-1]
		cur := runes[i]

		split := false

		switch {
		case unicode.IsLower(prev) && unicode.IsUpper(cur):
			// lower→upper: split only when the lowercase run has length ≥ 2,
			// so a lone leading lowercase (e.g. the "g" in "gRPC") stays glued
			// to the following upper-run as one word.
			split = lowerRun >= 2
		case unicode.IsUpper(prev) && unicode.IsUpper(cur) &&
			i+1 < len(runes) && unicode.IsLower(runes[i+1]) &&
			hasLowerRun(runes, i+1, 2):
			// upper→upper→lower: the current upper begins a new word so the
			// preceding upper-run forms its own word ("IDPToken" → IDP|Token).
			// Require at least two lowercase letters to follow, so a trailing
			// singular lowercase (the "s" in "APIs") does NOT tear the acronym
			// apart ("grpcAPIs" → grpc|APIs, not grpc|AP|Is).
			split = true
		}

		if split {
			words = append(words, runes[start:i])
			start = i
			lowerRun = 0
		}

		if unicode.IsLower(cur) {
			lowerRun++
		} else if unicode.IsUpper(cur) {
			lowerRun = 0
		}
	}

	words = append(words, runes[start:])

	return words
}

// hasLowerRun reports whether at least min consecutive lowercase runes begin at
// runes[from]. Used to gate the upper→upper→lower split on a substantial
// lowercase word following the upper run.
func hasLowerRun(runes []rune, from, minCount int) bool {
	count := 0

	for i := from; i < len(runes); i++ {
		if !unicode.IsLower(runes[i]) {
			break
		}

		count++
	}

	return count >= minCount
}

// titleWord formats a word for display. When preserveUpper is true the word is
// returned unchanged (trailing pure-acronym case); otherwise the word is
// lowercased and its first rune capitalized.
func titleWord(word []rune, preserveUpper bool) string {
	if len(word) == 0 {
		return ""
	}

	if preserveUpper {
		return string(word)
	}

	runes := append([]rune(nil), word...) // copy so ToLower does not mutate caller
	for i := range runes {
		runes[i] = unicode.ToLower(runes[i])
	}

	runes[0] = unicode.ToTitle(runes[0])

	return string(runes)
}

// isLastPureUpper reports whether a word should be kept uppercase: it must be
// the final word of the segment and consist entirely of uppercase letters with
// length at least 2 (a trailing acronym like "IDP" or "API"). Single-character
// uppercase words are treated as normal Title-cased words.
func isLastPureUpper(word []rune, isLast bool) bool {
	if !isLast || len(word) < 2 {
		return false
	}

	for _, r := range word {
		if !unicode.IsUpper(r) {
			return false
		}
	}

	return true
}
