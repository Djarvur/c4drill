// definition.go implements textDocument/definition (issues #32/#33):
//   - peer values resolve (walk-up, read-only) to the target unit's
//     [section] header / unit header line — in this document or, via the
//     include closure, in the included file that defines it (both formats);
//   - `template = "name"` / `use name(...)` values jump to the
//     [template.name] header / template declaration;
//   - `path =` / `include path` values jump to the target file itself.
// The TOML classes live here; the .c4d classes in c4dlang.go.

package lsp

import (
	"os"
	"path/filepath"
	"strings"
)

// definitionAt is the textDocument/definition feature entry; the LSP result
// is Location[] (null when nothing is found).
func (s *Server) definitionAt(doc *document, pos Position) any {
	switch ext := filepath.Ext(doc.Path); ext {
	case extC4d:
		return s.c4dDefinition(doc, pos)
	case extToml:
		return s.tomlDefinition(doc, pos)
	default:
		return nil
	}
}

// tomlDefinition is the TOML-dialect definition body.
func (s *Server) tomlDefinition(doc *document, pos Position) any {
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

	return s.searchIncludesFrom(doc.Path, text, defTarget{unitPath: target}, map[string]bool{})
}

// templateDefinition jumps from a `template = "name"` value to its
// [template.name] header.
func (s *Server) templateDefinition(doc *document, text string, ctx lineContext) *Location {
	name := ctx.fullValue
	if name == "" {
		return nil
	}

	return s.searchIncludesFrom(doc.Path, text, defTarget{template: name}, map[string]bool{})
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

// searchIncludesFrom walks the include closure (open buffers first, then
// disk) looking for the target's defining line. The entry file is searched
// first; the closure walks whichever front-end each file's extension selects,
// so a .c4d entry can define peers in .toml files and vice versa.
func (s *Server) searchIncludesFrom(path, text string, target defTarget, visited map[string]bool) *Location {
	canonical := canonicalPath(path)
	if visited[canonical] {
		return nil
	}

	visited[canonical] = true

	var r *Range

	if filepath.Ext(canonical) == extC4d {
		r = target.findInC4D(text)
	} else {
		r = target.findInToml(text)
	}

	if r != nil {
		return &Location{URI: pathToURI(canonical), Range: *r}
	}

	dir := filepath.Dir(canonical)

	for _, includeTarget := range includeDirectives(canonical, text) {
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

// includeDirectives extracts the include paths a document (of either format)
// directives at, in source order.
func includeDirectives(path, text string) []string {
	if filepath.Ext(path) == extC4d {
		includes := c4dScanDocument(text).includes
		paths := make([]string, 0, len(includes))

		for _, inc := range includes {
			paths = append(paths, inc.path)
		}

		return paths
	}

	var paths []string

	for _, h := range scanHeaders(text) {
		if !h.isArray || !strings.HasSuffix(h.path, "."+tblInclude) && h.path != tblInclude {
			continue
		}

		if target := includeDirectiveTarget(text, h); target != "" {
			paths = append(paths, target)
		}
	}

	return paths
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
