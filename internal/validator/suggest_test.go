package validator_test

import (
	"testing"

	"github.com/Djarvur/c4drill/internal/validator"
	"github.com/stretchr/testify/assert"
)

func TestSuggestSimilar_ExactMatch(t *testing.T) {
	t.Parallel()

	candidates := []string{"api", "db", "web"}
	result := validator.SuggestSimilar("api", candidates)

	assert.Equal(t, "api", result)
}

func TestSuggestSimilar_OneCharTypo(t *testing.T) {
	t.Parallel()

	candidates := []string{"api", "db", "web"}
	result := validator.SuggestSimilar("apu", candidates) // 'u' instead of 'i'

	assert.Equal(t, "api", result)
}

func TestSuggestSimilar_TwoCharTypo(t *testing.T) {
	t.Parallel()

	candidates := []string{"database", "api", "web"}
	result := validator.SuggestSimilar("databas", candidates) // missing 'e'

	assert.Equal(t, "database", result)
}

func TestSuggestSimilar_NoMatchExceedsThreshold(t *testing.T) {
	t.Parallel()

	candidates := []string{"api", "db", "web"}
	result := validator.SuggestSimilar("xyz", candidates) // distance > 2 from all

	assert.Empty(t, result)
}

func TestSuggestSimilar_ShortNameReturnsEmpty(t *testing.T) {
	t.Parallel()

	candidates := []string{"db", "ab", "cd"}
	result := validator.SuggestSimilar("xb", candidates) // 2 chars < minNameLengthForSuggestion

	assert.Empty(t, result)
}

func TestSuggestSimilar_MinLengthThree(t *testing.T) {
	t.Parallel()

	candidates := []string{"api", "db", "web"}
	result := validator.SuggestSimilar("apx", candidates) // 3 chars, distance 1 from api

	assert.Equal(t, "api", result)
}

func TestSuggestSimilar_EmptyCandidates(t *testing.T) {
	t.Parallel()

	candidates := []string{}
	result := validator.SuggestSimilar("api", candidates)

	assert.Empty(t, result)
}

func TestSuggestSimilar_CaseSensitive(t *testing.T) {
	t.Parallel()

	candidates := []string{"API", "Api", "api"}
	result := validator.SuggestSimilar("api", candidates)

	// Should match exact case first
	assert.Equal(t, "api", result)
}

func TestSuggestSimilar_ReturnsBestMatch(t *testing.T) {
	t.Parallel()

	candidates := []string{"handler", "handler2", "handler3"}
	result := validator.SuggestSimilar("handlar", candidates) // distance 1 from handler

	assert.Equal(t, "handler", result)
}

func TestFormatSuggestion_WithMatch(t *testing.T) {
	t.Parallel()

	candidates := []string{"api", "db", "web"}
	result := validator.FormatSuggestion("apu", candidates)

	expected := ` (did you mean "api"?)`
	assert.Equal(t, expected, result)
}

func TestFormatSuggestion_NoMatch(t *testing.T) {
	t.Parallel()

	candidates := []string{"api", "db", "web"}
	result := validator.FormatSuggestion("xyz", candidates)

	assert.Empty(t, result)
}

func TestFormatSuggestion_ShortName(t *testing.T) {
	t.Parallel()

	candidates := []string{"db", "ab", "cd"}
	result := validator.FormatSuggestion("xb", candidates)

	assert.Empty(t, result)
}

func TestFormatSuggestion_EmptyCandidates(t *testing.T) {
	t.Parallel()

	candidates := []string{}
	result := validator.FormatSuggestion("api", candidates)

	assert.Empty(t, result)
}
