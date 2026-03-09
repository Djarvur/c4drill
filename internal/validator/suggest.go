package validator

import (
	"fmt"

	"github.com/agnivade/levenshtein"
)

const (
	// maxSuggestionDistance is the maximum Levenshtein distance for a valid suggestion.
	maxSuggestionDistance = 2
	// minNameLengthForSuggestion is the minimum name length to offer suggestions.
	// Short names (e.g., "db", "id") produce too many false positives.
	minNameLengthForSuggestion = 3
)

// SuggestSimilar returns the closest matching name from candidates using Levenshtein distance.
// Returns empty string if:
//   - typo length < minNameLengthForSuggestion (3 chars)
//   - no candidate within maxSuggestionDistance (2 edits)
//   - candidates list is empty
func SuggestSimilar(typo string, candidates []string) string {
	// Skip suggestions for short names - too many false positives
	if len(typo) < minNameLengthForSuggestion {
		return ""
	}

	if len(candidates) == 0 {
		return ""
	}

	bestMatch := ""
	bestDistance := maxSuggestionDistance + 1 // Start above threshold

	for _, candidate := range candidates {
		dist := levenshtein.ComputeDistance(typo, candidate)
		if dist < bestDistance {
			bestDistance = dist
			bestMatch = candidate
		}
	}

	if bestDistance <= maxSuggestionDistance {
		return bestMatch
	}
	return ""
}

// FormatSuggestion returns a formatted suggestion string for error messages.
// Returns empty string if no suitable suggestion is found.
func FormatSuggestion(typo string, candidates []string) string {
	if suggestion := SuggestSimilar(typo, candidates); suggestion != "" {
		return fmt.Sprintf(` (did you mean "%s"?)`, suggestion)
	}
	return ""
}
