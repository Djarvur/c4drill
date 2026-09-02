// tomlctx.go is the line-level TOML source analyzer the language features
// share: table-header scanning, cursor context classification (which table,
// which key, value prefix), declared-type extraction, and unit-path
// derivation. It works on the CURRENT buffer — mid-edit documents that no
// longer parse must still complete and outline, so nothing here requires a
// successful model parse.

package lsp

import (
	"regexp"
	"strings"
)

// tomlHeader is one table header found by scanning lines.
type tomlHeader struct {
	path    string // dotted path as authored (unquoted segments assumed)
	isArray bool   // [[path]] array-of-tables
	line    int    // 0-based line index
}

// reserved table names and link-table suffixes the unit outline must skip.
const (
	tblProperties = "properties"
	tblInclude    = "include"
	tblUse        = "use"
	tblTemplate   = "template"
	tblLink       = "link"
	tblLinkFrom   = "linkFrom"
)

// scanHeaders returns the document's table headers in source order.
// Quoted key segments are out of scope for the analyzer (c4drill section
// names are bare identifiers), so a naive bracket/scan suffices; malformed
// header lines are skipped rather than guessed.
func scanHeaders(text string) []tomlHeader {
	var headers []tomlHeader

	for i, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "[") || !strings.HasSuffix(trimmed, "]") {
			continue
		}

		isArray := strings.HasPrefix(trimmed, "[[")

		inner := strings.TrimSuffix(strings.TrimPrefix(trimmed, "["), "]")
		if isArray {
			inner = strings.TrimSuffix(strings.TrimPrefix(inner, "["), "]")
		}

		path := normalizeTablePath(inner)
		if path == "" || strings.ContainsAny(path, "[]#") {
			continue
		}

		headers = append(headers, tomlHeader{path: path, isArray: isArray, line: i})
	}

	return headers
}

// normalizeTablePath trims spaces and one layer of quotes per segment.
func normalizeTablePath(inner string) string {
	parts := strings.Split(inner, ".")
	for j, p := range parts {
		p = strings.TrimSpace(p)
		p = strings.TrimSuffix(strings.TrimPrefix(p, `"`), `"`)

		parts[j] = p
	}

	return strings.Join(parts, ".")
}

// declaredTypes maps plain-table path -> the `type = "..."` value authored in
// that table's section. Only the first type line per table counts.
func declaredTypes(text string) map[string]string {
	types := make(map[string]string)

	current := ""

	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)

		if isTableHeader(trimmed) {
			current = headerTablePath(trimmed)

			continue
		}

		key, value, found := cutKeyValue(trimmed)
		if !found || key != "type" || current == "" {
			continue
		}

		if _, taken := types[current]; !taken {
			types[current] = unquote(value)
		}
	}

	return types
}

// unitPaths returns the document's hand-authored unit paths: plain-table
// headers that are not reserved tables, link tables, use tables, or template
// bodies.
func unitPaths(headers []tomlHeader) []string {
	var paths []string

	for _, h := range headers {
		if h.isArray || !isUnitTablePath(h.path) {
			continue
		}

		paths = append(paths, h.path)
	}

	return paths
}

// isUnitTablePath reports whether a plain table path denotes a unit section
// (not [properties], not a template subtree, not a [[use]]/[[include]]).
func isUnitTablePath(path string) bool {
	segments := strings.Split(path, ".")
	switch segments[0] {
	case tblProperties, tblTemplate:
		return false
	case tblUse, tblInclude:
		return len(segments) == 1 // [[use]]/[[include]] themselves; unit names win below
	}

	return true
}

// lineContext classifies the cursor position in a TOML document.
type lineContext struct {
	// tablePath is the enclosing table's dotted path ("" before any header).
	tablePath string
	// isArray reports an array-of-tables header ([[link]] and friends).
	isArray bool
	// key is the current line's key when the cursor sits in value position.
	key string
	// inValue reports value position (after the line's '=').
	inValue bool
	// valuePrefix is the typed value text at the cursor (quotes stripped).
	valuePrefix string
	// fullValue is the complete value on the cursor's line (quotes
	// stripped) — hover and definition want the whole value, completion
	// the typed prefix.
	fullValue string
	// inTemplateRef reports a cursor inside a ${param} token.
	inTemplateRef bool
	// templateRefPrefix is the param text already typed inside the ${...}.
	templateRefPrefix string
	// line and byteIdx locate the cursor for word extraction (byteIdx is
	// within the line); lineNo is the 0-based line index.
	line    string
	byteIdx int
	lineNo  int
}

// analyzeLine classifies the cursor location (0-based line, byte-safe).
func analyzeLine(text string, pos Position) lineContext {
	starts := lineStarts(text)
	byteIdx := offsetForPosition(text, pos)

	lineStart := 0
	if int(pos.Line) < len(starts) {
		lineStart = starts[pos.Line]
	}

	fullLine := text[lineStart:]
	if nl := strings.IndexByte(fullLine, '\n'); nl >= 0 {
		fullLine = fullLine[:nl]
	}

	idx := byteIdx - lineStart
	if idx < 0 {
		idx = 0
	}

	if idx > len(fullLine) {
		idx = len(fullLine)
	}

	ctx := lineContext{
		line:    fullLine,
		byteIdx: idx,
		lineNo:  int(pos.Line),
	}

	ctx.classify()

	// Enclosing table: the last header at or above the cursor's line.
	for _, h := range scanHeaders(text) {
		if h.line <= int(pos.Line) {
			ctx.tablePath = h.path
			ctx.isArray = h.isArray
		}
	}

	return ctx
}

// classify fills the key/value/template-ref fields from the cursor prefix.
func (c *lineContext) classify() {
	before := c.line[:c.byteIdx]

	if eq := indexOutsideQuotes(before, '='); eq >= 0 {
		c.inValue = true
		c.key = strings.TrimSpace(before[:eq])
		c.valuePrefix = stripValueQuotes(strings.TrimSpace(before[eq+1:]))
	}

	if eqFull := indexOutsideQuotes(c.line, '='); eqFull >= 0 {
		c.fullValue = stripValueQuotes(strings.TrimSpace(c.line[eqFull+1:]))
	}

	if m := templateRefRe.FindStringSubmatch(before); m != nil {
		c.inTemplateRef = true
		c.templateRefPrefix = m[1]
	}
}

// hostUnitPath returns the dotted path of the unit whose section (or link
// table) the cursor is in: [[web.link]] hosts unit "web"; [a.b] hosts "a.b".
func (c *lineContext) hostUnitPath() string {
	path := c.tablePath
	if path == "" {
		return ""
	}

	if c.isArray {
		segments := strings.Split(path, ".")

		last := segments[len(segments)-1]
		if last == tblLink || last == tblLinkFrom {
			path = strings.Join(segments[:len(segments)-1], ".")
		}
	}

	return path
}

// isLinkTable reports a [[...link]]/[[...linkFrom]] cursor context.
func (c *lineContext) isLinkTable() bool {
	if !c.isArray {
		return false
	}

	segments := strings.Split(c.tablePath, ".")

	last := segments[len(segments)-1]

	return last == tblLink || last == tblLinkFrom
}

// tableKind classifies the enclosing table for field completion.
type tableKind int

const (
	kindUnit       tableKind = iota // [unit] or [parent.child]
	kindLink                        // [[x.link]] / [[x.linkFrom]]
	kindProperties                  // [properties]
	kindInclude                     // [[include]]
	kindUse                         // [[use]] / [[x.use]] / [[template.t.p.use]]
	kindTemplate                    // [template.<name>] subtree
	kindOther                       // pre-table or unrecognized
)

// kind classifies the cursor's enclosing table.
func (c *lineContext) kind() tableKind {
	if c.tablePath == "" {
		return kindOther
	}

	switch {
	case c.isLinkTable():
		return kindLink
	case c.isArray && strings.HasSuffix(c.tablePath, "."+tblUse):
		return kindUse
	case c.isArray && c.tablePath == tblInclude:
		return kindInclude
	}

	segments := strings.Split(c.tablePath, ".")
	switch segments[0] {
	case tblProperties:
		return kindProperties
	case tblUse:
		return kindUse
	case tblInclude:
		return kindInclude
	case tblTemplate:
		return kindTemplate
	}

	return kindUnit
}

// templateRefRe matches a cursor position inside an unterminated ${param}
// token: the prefix ends somewhere within "${" ... no closing brace.
var templateRefRe = regexp.MustCompile(`\$\{([^}]*)$`)

// indexOutsideQuotes returns the index of the first occurrence of target in
// s that is not inside a single- or double-quoted region, or -1.
func indexOutsideQuotes(s string, target byte) int {
	var inSingle, inDouble bool

	for i := range len(s) {
		switch s[i] {
		case '\'':
			if !inDouble {
				inSingle = !inSingle
			}
		case '"':
			if !inSingle {
				inDouble = !inDouble
			}
		case target:
			if !inSingle && !inDouble {
				return i
			}
		}
	}

	return -1
}

// cutKeyValue splits a `key = value` line into its parts; found is false for
// lines without an unquoted '='.
func cutKeyValue(line string) (string, string, bool) {
	eq := indexOutsideQuotes(line, '=')
	if eq < 0 {
		return "", "", false
	}

	return strings.TrimSpace(line[:eq]), strings.TrimSpace(line[eq+1:]), true
}

// unquote strips one layer of surrounding double or single quotes.
func unquote(s string) string {
	if len(s) >= 2 && (s[0] == '"' && s[len(s)-1] == '"' || s[0] == '\'' && s[len(s)-1] == '\'') {
		return s[1 : len(s)-1]
	}

	return s
}

// stripValueQuotes strips the quotes around a value the cursor is typing:
// the closing quote is usually absent mid-edit, so each quote side strips
// independently.
func stripValueQuotes(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, `"`)
	s = strings.TrimSuffix(s, `"`)
	s = strings.TrimPrefix(s, "'")
	s = strings.TrimSuffix(s, "'")

	return s
}
