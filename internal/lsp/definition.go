// definition.go implements textDocument/definition (issue #32):
//   - peer values resolve (walk-up, read-only) to the target unit's
//     [section] header — in this document or, via the include closure, in
//     the included file that defines it;
//   - `template = "name"` values jump to the [template.name] header;
//   - `path =` values in [[include]] jump to the target file itself.
// TOML dialect only.

package lsp

import (
	"os"
	"path/filepath"
	"strings"
)

// definitionAt is the textDocument/definition feature entry; the LSP result
// is Location[] (null when nothing is found).
func (s *Server) definitionAt(doc *document, pos Position) any {
	if filepath.Ext(doc.Path) != extToml {
		return nil
	}

	text := string(doc.Text)
	ctx := analyzeLine(text, pos)

	if !ctx.inValue {
		return nil
	}

	var loc *Location

	switch ctx.key {
	case "peer":
		loc = s.peerDefinition(doc, text, ctx)
	case "template":
		loc = s.templateDefinition(doc, text, ctx)
	case "path":
		loc = s.includeDefinition(doc, ctx)
	}

	if loc == nil {
		return nil
	}

	return []Location{*loc}
}

// peerDefinition finds the unit section a peer value refers to, searching
// this document first and then the include closure.
func (s *Server) peerDefinition(doc *document, text string, ctx lineContext) *Location {
	m := s.mergedModel(doc)
	if m == nil {
		return nil
	}

	target := resolvePeerReadOnly(m, ctx.hostUnitPath(), ctx.fullValue)
	if target == "" {
		return nil
	}

	// This document first, then every transitively included file.
	if r := headerRange(scanHeaders(text), target, false); r != nil {
		return &Location{URI: doc.URI, Range: *r}
	}

	return s.searchIncludedFiles(doc, target)
}

// templateDefinition jumps from a `template = "name"` value to its
// [template.name] header.
func (s *Server) templateDefinition(doc *document, text string, ctx lineContext) *Location {
	name := ctx.fullValue
	if name == "" {
		return nil
	}

	target := tblTemplate + "." + name

	if r := headerRange(scanHeaders(text), target, false); r != nil {
		return &Location{URI: doc.URI, Range: *r}
	}

	return s.searchIncludedFiles(doc, target)
}

// includeDefinition jumps from a [[include]] path to the target file.
func (s *Server) includeDefinition(doc *document, ctx lineContext) *Location {
	if ctx.fullValue == "" {
		return nil
	}

	target := filepath.Join(filepath.Dir(doc.Path), ctx.fullValue)
	if !fileExists(target) {
		return nil
	}

	return &Location{
		URI:   pathToURI(target),
		Range: Range{},
	}
}

// searchIncludedFiles walks the include closure (open buffers first, then
// disk) looking for the plain-table header defining path.
func (s *Server) searchIncludedFiles(doc *document, path string) *Location {
	visited := map[string]bool{}

	return s.searchIncludesFrom(doc.Path, string(doc.Text), path, visited)
}

// searchIncludesFrom is the recursive worker over one file's text.
func (s *Server) searchIncludesFrom(path, text, target string, visited map[string]bool) *Location {
	canonical := canonicalPath(path)
	if visited[canonical] {
		return nil
	}

	visited[canonical] = true

	headers := scanHeaders(text)
	if r := headerRange(headers, target, false); r != nil {
		loc := &Location{URI: pathToURI(canonical), Range: *r}

		return loc
	}

	dir := filepath.Dir(canonical)

	for _, h := range headers {
		if !h.isArray || !strings.HasSuffix(h.path, "."+tblInclude) && h.path != tblInclude {
			continue
		}

		includeTarget := includeDirectiveTarget(text, h)
		if includeTarget == "" {
			continue
		}

		includedPath := canonicalPath(filepath.Join(dir, includeTarget))

		includedText, err := s.readForWalk(includedPath)
		if err != nil {
			continue
		}

		if loc := s.searchIncludesFrom(includedPath, string(includedText), target, visited); loc != nil {
			return loc
		}
	}

	return nil
}

// includeDirectiveTarget extracts the `path = "..."` value authored inside
// the [[include]] section starting at header h.
func includeDirectiveTarget(text string, h tomlHeader) string {
	lines := strings.Split(text, "\n")

	for i := h.line + 1; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])

		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			return "" // next section began without a path
		}

		if k, v, found := cutKeyValue(trimmed); found && k == "path" {
			return unquote(v)
		}
	}

	return ""
}

// headerRange finds the plain-table header with exactly path and returns
// its header-line range. isArray selects [[...]] headers instead.
func headerRange(headers []tomlHeader, path string, isArray bool) *Range {
	for _, h := range headers {
		if h.path == path && h.isArray == isArray {
			r := Range{
				Start: Position{Line: uint32(h.line)}, //nolint:gosec // header line is non-negative
				End:   Position{Line: uint32(h.line)}, //nolint:gosec // header line is non-negative
			}

			r.End.Character = 1 // represent a non-empty selection span

			return &r
		}
	}

	return nil
}

// fileExists reports whether path exists on disk.
func fileExists(path string) bool {
	_, err := os.Stat(path)

	return err == nil
}
