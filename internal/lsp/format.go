// format.go implements textDocument/formatting (issue #32, M3) with exact
// `c4drill fmt` parity: .toml through internal/tomlfmt (comment-preserving,
// key order preserved), .c4d through the canonical C4D printer (ParseAST +
// EmitC4D — comments ride the AST). Both formats re-apply the fmt command's
// safety gate: a candidate that does not re-parse to the original Model is
// refused, never offered as an edit. Malformed documents format to null.

package lsp

import (
	"bytes"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/Djarvur/c4drill/internal/c4d"
	"github.com/Djarvur/c4drill/internal/parser"
	"github.com/Djarvur/c4drill/internal/tomlfmt"
)

// formatting is the textDocument/formatting feature entry. The client's
// options are accepted per the wire contract but do not alter the output:
// c4drill formatting is deterministic, gofmt-style, with no style knobs
// (same as the fmt command).
func (s *Server) formatting(doc *document, _ FormattingOptions) []TextEdit {
	formatted, ok := formatDocument(doc.Text, doc.Path)
	if !ok {
		return nil // malformed or unsupported: no edit offered
	}

	if bytes.Equal(formatted, doc.Text) {
		return []TextEdit{} // already formatted: fmt --check parity (clean)
	}

	return []TextEdit{{
		Range:   wholeDocumentRange(string(doc.Text)),
		NewText: string(formatted),
	}}
}

// formatDocument mirrors cmd/c4drill fmt's per-file candidate construction
// and T-35-08-01 safety gate. ok is false when the document cannot be
// formatted safely.
func formatDocument(data []byte, path string) ([]byte, bool) {
	switch ext := filepath.Ext(path); ext {
	case extToml:
		orig, err := parser.Parse(data)
		if err != nil {
			return nil, false
		}

		formatted, err := tomlfmt.Format(data)
		if err != nil {
			return nil, false // tomlfmt fails closed on malformed input
		}

		return gated(formatted, orig, parser.Parse)
	case extC4d:
		orig, err := c4d.Parse(data)
		if err != nil {
			return nil, false
		}

		// Emission AST: a separate parse, mirroring fmt's two-parse rule.
		emitDoc, err := c4d.ParseAST(data)
		if err != nil {
			return nil, false
		}

		return gated([]byte(c4d.EmitC4D(emitDoc)), orig, c4d.Parse)
	default:
		return nil, false
	}
}

// gated re-parses the candidate through reparse and compares Models — the
// fmt safety gate: a rewrite that would break parsing or change semantics is
// refused (the file equivalent of leaving the file untouched).
func gated(formatted []byte, orig *parser.Model, reparse func([]byte) (*parser.Model, error)) ([]byte, bool) {
	gatedModel, err := reparse(formatted)
	if err != nil || !reflect.DeepEqual(orig, gatedModel) {
		return nil, false
	}

	return formatted, true
}

// wholeDocumentRange spans the entire text (the single-edit replacement
// range textDocument/formatting results use).
func wholeDocumentRange(text string) Range {
	lines := strings.Count(text, "\n")

	endChar := uint32(utf16Units(text[lastNewline(text)+1:])) //nolint:gosec // bounded by text length

	return Range{
		Start: Position{Line: 0, Character: 0},
		End:   Position{Line: uint32(lines), Character: endChar}, //nolint:gosec // line count is non-negative
	}
}

// lastNewline returns the index of the final '\n' (or -1).
func lastNewline(text string) int {
	for i := len(text) - 1; i >= 0; i-- {
		if text[i] == '\n' {
			return i
		}
	}

	return -1
}
