// hover.go implements textDocument/hover (issue #32/#33): at peer references
// the resolved absolute unit path, its C level, and its (post-promotion)
// type; at ${param} tokens and template references, the template's parameter
// info. The TOML classes live here; the .c4d classes in c4dlang.go.

package lsp

import (
	"path/filepath"
	"strings"
	"unicode/utf16"

	"github.com/Djarvur/c4drill/internal/include"
	"github.com/Djarvur/c4drill/internal/model"
	"github.com/Djarvur/c4drill/internal/parser"
)

// mergedModel parses the document and resolves its includes through the
// buffer overlay — the model the CLI pipeline itself validates. When include
// resolution fails (a broken include), the entry's own units still resolve.
func (s *Server) mergedModel(doc *document) *parser.Model {
	m, err := parseByExt(doc.Path, doc.Text)
	if err != nil {
		return nil
	}

	m, err = include.ResolveWithReader(m, filepath.Dir(doc.Path), doc.Path, s.overlayRead())
	if err != nil {
		return m
	}

	return m
}

// hoverAt is the textDocument/hover feature entry.
func (s *Server) hoverAt(doc *document, pos Position) any {
	switch ext := filepath.Ext(doc.Path); ext {
	case extC4d:
		return s.c4dHover(doc, pos)
	case extToml:
		return s.tomlHover(doc, pos)
	default:
		return nil
	}
}

// tomlHover is the TOML-dialect hover body.
func (s *Server) tomlHover(doc *document, pos Position) any {
	text := string(doc.Text)
	ctx := analyzeLine(text, pos)

	switch {
	case ctx.inTemplateRef:
		return templateParamHover(text, ctx)
	case ctx.inValue && ctx.key == "peer":
		return s.peerHover(doc, text, ctx, pos)
	case ctx.inValue && ctx.key == "template":
		return templateRefHover(text, ctx)
	default:
		return nil
	}
}

// peerHover resolves the peer value under the cursor the way peer.Resolve
// would (read-only) and reports the target's absolute path, level, and type.
func (s *Server) peerHover(doc *document, text string, ctx lineContext, pos Position) *Hover {
	m := s.mergedModel(doc)
	if m == nil {
		return nil // a buffer that does not parse has nothing resolvable
	}

	target := resolvePeerReadOnly(m, ctx.hostUnitPath(), ctx.fullValue)
	if target == "" {
		return nil
	}

	unit := findUnit(m, target)
	if unit == nil {
		return nil
	}

	return &Hover{
		Contents: MarkupContent{Kind: "markdown", Value: peerHoverMarkdown(target, unit)},
		Range:    wordRangeAt(text, pos),
	}
}

// peerHoverMarkdown renders the resolved-peer hover text shared by both
// dialects: absolute path, C level, promoted type, display name, description.
func peerHoverMarkdown(target string, unit *model.Unit) string {
	var b strings.Builder

	b.WriteString("**" + target + "**\n\n")
	b.WriteString(levelLabel(target) + " unit — type `" + string(unit.Type) + "`")

	if unit.Name != "" {
		b.WriteString("\n\n" + unit.Name)
	}

	if unit.Description != "" {
		b.WriteString(" — " + unit.Description)
	}

	return b.String()
}

// templateParamListHover renders the ${param} hover for a template's declared
// parameters (shared by both dialects).
func templateParamListHover(name string, params []string) *Hover {
	if len(params) == 0 {
		return nil
	}

	var b strings.Builder

	b.WriteString("Template `" + name + "` parameters:")

	for _, p := range params {
		b.WriteString("\n- `" + p + "`")
	}

	return &Hover{Contents: MarkupContent{Kind: "markdown", Value: b.String()}}
}

// templateRefListHover renders the template-reference hover (shared by both
// dialects).
func templateRefListHover(name string, params []string) *Hover {
	if len(params) == 0 {
		return nil
	}

	var b strings.Builder

	b.WriteString("Template `" + name + "` — parameters:")

	for _, p := range params {
		b.WriteString(" `" + p + "`")
	}

	return &Hover{Contents: MarkupContent{Kind: "markdown", Value: b.String()}}
}

// templateParamHover reports the parameter list of the template the cursor's
// ${...} token belongs to.
func templateParamHover(text string, ctx lineContext) *Hover {
	name := enclosingTemplate(ctx)
	if name == "" {
		return nil
	}

	return templateParamListHover(name, templateParams(text, name))
}

// templateRefHover reports the referenced template's parameter info at a
// `template = "name"` value.
func templateRefHover(text string, ctx lineContext) *Hover {
	return templateRefListHover(ctx.fullValue, templateParams(text, ctx.fullValue))
}

// levelLabel derives C1/C2/C3 from the unit path depth (the view levels).
func levelLabel(path string) string {
	switch strings.Count(path, ".") + 1 {
	case 1:
		return "C1"
	case 2:
		return "C2"
	default:
		return "C3"
	}
}

// resolvePeerReadOnly mirrors internal/peer's D-13/D-14/D-15/D-16 walk-up
// WITHOUT mutating the model: dotted peers are absolute; bare peers search
// the host's ancestor scopes nearest-first (immediate parent's children,
// then each grandparent's, ending at root's m.Units).
func resolvePeerReadOnly(m *parser.Model, hostPath, peer string) string {
	if m == nil || peer == "" {
		return ""
	}

	if strings.Contains(peer, ".") {
		return absolutePeer(m, peer)
	}

	return walkedUpPeer(m, hostPath, peer)
}

// absolutePeer validates a dotted peer against the model.
func absolutePeer(m *parser.Model, peer string) string {
	if unit := findUnit(m, peer); unit != nil {
		return peer
	}

	return ""
}

// walkedUpPeer resolves a bare peer through the host's ancestor scopes,
// nearest-first, exactly the walk internal/peer performs before Validate.
func walkedUpPeer(m *parser.Model, hostPath, peer string) string {
	segments := strings.Split(hostPath, ".")

	for i := len(segments) - 1; i >= 0; i-- {
		scopePath := strings.Join(segments[:i], ".")
		if childScope(m, scopePath, peer) {
			if scopePath == "" {
				return peer // root match is the identity rewrite (D-15)
			}

			return scopePath + "." + peer
		}
	}

	return ""
}

// childScope reports whether the scope at scopePath ("", root, else a unit
// path) directly contains a unit named peer.
func childScope(m *parser.Model, scopePath, peer string) bool {
	if scopePath == "" {
		_, ok := m.Units[peer]

		return ok
	}

	scope := findUnit(m, scopePath)
	if scope == nil {
		return false
	}

	_, ok := scope.Subunits[peer]

	return ok
}

// findUnit locates the unit at a dotted path (nil when absent).
func findUnit(m *parser.Model, path string) *model.Unit {
	if m == nil || path == "" {
		return nil
	}

	current := m.Units

	segments := strings.Split(path, ".")

	var unit *model.Unit

	for i, seg := range segments {
		unit = current[seg]
		if unit == nil {
			return nil
		}

		if i < len(segments)-1 {
			current = unit.Subunits
		}
	}

	return unit
}

// wordRangeAt returns the range of the identifier-like word (letters, digits,
// `_`, `-`, `.`) at the position, or nil when the cursor is not on one.
func wordRangeAt(text string, pos Position) *Range {
	starts := lineStarts(text)
	if int(pos.Line) >= len(starts) {
		return nil
	}

	lineStart := starts[pos.Line]

	lineEnd := len(text)
	if int(pos.Line)+1 < len(starts) {
		lineEnd = starts[pos.Line+1] - 1
	}

	line := text[lineStart:lineEnd]
	idx := offsetForPosition(text, pos) - lineStart

	isWord := func(b byte) bool {
		switch {
		case b >= 'a' && b <= 'z', b >= 'A' && b <= 'Z', b >= '0' && b <= '9':
			return true
		case b == '_' || b == '-' || b == '.':
			return true
		default:
			return false
		}
	}

	start := idx
	for start > 0 && start-1 < len(line) && isWord(line[start-1]) {
		start--
	}

	end := idx
	for end < len(line) && isWord(line[end]) {
		end++
	}

	if start >= end {
		return nil
	}

	return &Range{
		Start: Position{Line: pos.Line, Character: uint32(utf16Units(line[:start]))}, //nolint:gosec // bounded by line length
		End:   Position{Line: pos.Line, Character: uint32(utf16Units(line[:end]))},   //nolint:gosec // bounded by line length
	}
}

// utf16Units counts the UTF-16 code units of s (the LSP column encoding).
func utf16Units(s string) int {
	n := 0

	for _, r := range s {
		u := utf16.RuneLen(r)
		if u < 0 {
			u = 1
		}

		n += u
	}

	return n
}
