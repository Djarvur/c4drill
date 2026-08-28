package grammar

import (
	"fmt"
	"slices"

	c4dparser "github.com/Djarvur/c4drill/internal/parser"
	"github.com/Djarvur/c4drill/internal/validator"
)

// reservedKeywords is the closed set of words a unit id cannot collide with
// (D-19): every internal/parser isBuiltinField key plus the statement
// keywords once/use/include/template/properties. The isBuiltinField strings
// are copied verbatim from internal/parser/parser.go — keep both lists in
// sync when either grows.
//nolint:gochecknoglobals // Pinned closed set per D-19, immutable after init
var reservedKeywords = []string{
	// isBuiltinField (internal/parser/parser.go)
	"type", "name", "description", "technology",
	"reference",
	"color", "style", "border", "edges",
	"width", "height", "expanded",
	"link", "linkFrom",
	"use",
	// statement keywords (c4d.peg)
	"include", "template", "properties", "once",
}

// fieldKeywords are the keys accepted where field statements parse (unit
// bodies, edge options, properties blocks). They serve as suggestion
// candidates when an unknown field key errors (D-19).
//nolint:gochecknoglobals // Pinned closed set per D-19, immutable after init
var fieldKeywords = []string{
	// FieldKey + OptionKey + PropertyKey (c4d.peg) + width/height (TOML
	// unit fields) — union, alphabetical.
	"arrow", "border", "color", "description", "edges", "expanded",
	"height", "kind", "labelPosition", "length", "lineLength", "name", "rank",
	"reference", "style", "technology", "type", "width",
}

// ReservedKeywords returns the words reserved against unit ids (D-19): the
// internal/parser isBuiltinField set plus the statement keywords. 19 total.
func ReservedKeywords() []string {
	return slices.Clone(reservedKeywords)
}

// isReservedKeyword reports whether id exactly collides with a reserved
// word. Near-misses are NOT reserved — only exact matches error.
func isReservedKeyword(id string) bool {
	return slices.Contains(reservedKeywords, id)
}

// CheckUnitID returns a *c4dparser.ParseError carrying a Levenshtein
// suggestion when id collides with a reserved word (D-19), nil otherwise.
// pos is the DSL-native line the id appears on.
func CheckUnitID(id string, pos int) error {
	if !isReservedKeyword(id) {
		return nil
	}

	return &c4dparser.ParseError{
		Message: fmt.Sprintf("unit id %q is a reserved word%s", id, validator.FormatSuggestion(id, ReservedKeywords())),
		Line:    pos,
	}
}

// UnknownFieldError returns a *c4dparser.ParseError for a field key outside
// the known set (D-19). Suggestion candidates are the reserved set plus the
// accepted field keywords.
func UnknownFieldError(key string, pos int) error {
	candidates := append(ReservedKeywords(), fieldKeywords...)

	return &c4dparser.ParseError{
		Message: fmt.Sprintf("unknown field key %q%s", key, validator.FormatSuggestion(key, candidates)),
		Line:    pos,
	}
}
