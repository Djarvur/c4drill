// symbols.go implements textDocument/documentSymbol (issues #32/#33): the
// document's units as an outline with hierarchy — each plain-table unit
// section (TOML) or unit block (C4D, closed at its own brace) becomes a
// symbol nested under its parent unit, with the display name and type as
// detail.

package lsp

import (
	"path/filepath"
	"strings"
)

// documentSymbols builds the unit outline for a document; nil for documents
// without unit sections.
func (s *Server) documentSymbols(doc *document) []DocumentSymbol {
	if filepath.Ext(doc.Path) == extC4d {
		return s.c4dDocumentSymbols(doc)
	}

	return s.tomlDocumentSymbols(doc)
}

// tomlDocumentSymbols is the TOML-dialect outline body: plain-table unit
// sections nested by their dotted paths.
func (s *Server) tomlDocumentSymbols(doc *document) []DocumentSymbol {
	text := string(doc.Text)

	headers := scanHeaders(text)

	b := newOutlineBuilder(text)

	for _, h := range headers {
		if h.isArray || !isUnitTablePath(h.path) {
			continue
		}

		b.add(h)
	}

	return b.finish(strings.Count(text, "\n"))
}

// outlineBuilder assembles the symbol tree, closing section ranges as the
// nesting stack unwinds.
type outlineBuilder struct {
	types map[string]string
	names map[string]string
	roots []DocumentSymbol
	stack []outlineFrame
}

// outlineFrame is one open section: its path and the symbol to close later.
type outlineFrame struct {
	path    []string
	created *DocumentSymbol
}

// newOutlineBuilder precomputes the per-section details from the buffer.
func newOutlineBuilder(text string) *outlineBuilder {
	return &outlineBuilder{
		types: declaredTypes(text),
		names: sectionDisplayNames(text),
	}
}

// add folds one unit header into the tree.
func (b *outlineBuilder) add(h tomlHeader) {
	segments := strings.Split(h.path, ".")

	// Close every open section the new header is not nested under.
	for len(b.stack) > 0 && !isPrefixPath(b.stack[len(b.stack)-1].path, segments) {
		top := b.stack[len(b.stack)-1]
		b.stack = b.stack[:len(b.stack)-1]

		b.closeAt(top, h.line)
	}

	sym := b.newSymbol(h, segments)

	if len(b.stack) == 0 {
		b.roots = append(b.roots, sym)
		b.stack = append(b.stack, outlineFrame{path: segments, created: &b.roots[len(b.roots)-1]})
	} else {
		parent := b.stack[len(b.stack)-1].created
		parent.Children = append(parent.Children, sym)
		b.stack = append(b.stack, outlineFrame{
			path:    segments,
			created: &parent.Children[len(parent.Children)-1],
		})
	}
}

// finish closes any sections still open at EOF and returns the roots.
func (b *outlineBuilder) finish(lastLine int) []DocumentSymbol {
	for i := len(b.stack) - 1; i >= 0; i-- {
		b.closeAt(b.stack[i], lastLine)
	}

	return b.roots
}

// closeAt bounds an open section's range at the closing line.
func (b *outlineBuilder) closeAt(f outlineFrame, line int) {
	f.created.Range.End = Position{Line: uint32(line)} //nolint:gosec // line indices are non-negative
}

// newSymbol builds one unit symbol with its detail line.
func (b *outlineBuilder) newSymbol(h tomlHeader, segments []string) DocumentSymbol {
	detail := b.types[h.path]
	if display := b.names[h.path]; display != "" {
		detail = display + " (" + detail + ")"
	}

	ln := uint32(h.line) //nolint:gosec // line indices are non-negative

	return DocumentSymbol{
		Name:           segments[len(segments)-1],
		Detail:         detail,
		Kind:           KindObject,
		Range:          Range{Start: Position{Line: ln}, End: Position{Line: ln}},
		SelectionRange: Range{Start: Position{Line: ln}, End: Position{Line: ln}},
	}
}

// isPrefixPath reports whether prefix is a proper ancestor of path.
func isPrefixPath(prefix, path []string) bool {
	if len(prefix) >= len(path) {
		return false
	}

	for i := range prefix {
		if prefix[i] != path[i] {
			return false
		}
	}

	return true
}

// sectionDisplayNames maps unit table path -> its authored `name = "..."`.
func sectionDisplayNames(text string) map[string]string {
	out := map[string]string{}

	current := ""

	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)

		if isTableHeader(trimmed) {
			current = headerTablePath(trimmed)

			continue
		}

		if k, v, found := cutKeyValue(trimmed); found && k == "name" && current != "" {
			if _, taken := out[current]; !taken {
				out[current] = unquote(v)
			}
		}
	}

	return out
}
