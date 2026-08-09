// Package canonical provides an order-insensitive semantic comparator for DOT
// (Graphviz) output, realizing STATE.md decision DI-1 (the "canonicalDOT"
// contract) as a reusable helper importable from any _test.go file in the repo.
//
// The pinned go-graphviz fork emits map-order-dependent sibling statement order
// and layout geometry, so raw byte comparison against committed golden baselines
// is impossible. Canonical normalizes both sides to a sorted, geometry-stripped
// semantic form (the D-02 equivalence contract) before comparison.
//
// This package is test-only by convention but lives outside a _test.go file so
// Go's toolchain exposes it on the importable package surface (Go excludes
// _test.go files from the import graph). It is the D-18 extraction of the prior
// art originally inlined at internal/graph/builder_test.go:1249-1527 (Phase 26
// WR-01/WR-02 hardening + Phase 28 REF-05 backward-compat golden). The 4 WR-01/
// WR-02 regression tests move with it (see canonical_test.go).
//
// Canonical is a pure function (no I/O, no global state), deterministic, and
// depends only on stdlib sort + strings. It uses testing.T only for t.Helper()
// marking; callers assert on the returned canonical form.
package canonical

import (
	"sort"
	"strings"
	"testing"
)

// Canonical normalizes DOT output into an order-insensitive semantic
// serialization (DI-1). The pinned go-graphviz fork emits map-order-dependent
// sibling statement order and layout geometry, so raw byte comparison against
// the committed golden baseline is impossible; the normalized form keeps
// statement kind, head, sorted attributes with layout geometry stripped, and
// recursively sorted children — the D-02 equivalence contract.
func Canonical(t *testing.T, dot string) string {
	t.Helper()

	stmts, ok := parseDOTStatements(dot)
	if !ok {
		t.Fatalf("canonical: failed to parse DOT output for canonical comparison")
	}

	return serializeDOTStatements(stmts)
}

// dotStatement is one parsed statement of an XDOT document: an attribute block
// (graph/node/edge defaults, node or edge statement) or a nested subgraph.
type dotStatement struct {
	kind     string
	head     string
	attrs    []string
	children []dotStatement
}

// parseDOTStatements skips the "digraph ... {" header and parses the statement
// list until the closing brace.
func parseDOTStatements(dot string) ([]dotStatement, bool) {
	idx := strings.Index(dot, "{")
	if idx < 0 {
		return nil, false
	}

	stmts, _, ok := parseDOTBlock(dot, idx+1)

	return stmts, ok
}

// parseDOTBlock parses statements starting at pos until the matching "}".
// It returns the parsed statements and the position just past the closing brace.
func parseDOTBlock(dot string, pos int) ([]dotStatement, int, bool) {
	var stmts []dotStatement

	for pos < len(dot) {
		pos = skipDOTWhitespace(dot, pos)
		if pos >= len(dot) {
			return stmts, pos, false
		}

		if dot[pos] == '}' {
			return stmts, pos + 1, true
		}

		var (
			stmt dotStatement
			ok   bool
		)

		if strings.HasPrefix(dot[pos:], "subgraph") {
			stmt, pos, ok = parseDOTSubgraph(dot, pos)
		} else {
			stmt, pos, ok = parseDOTAttrStatement(dot, pos)
		}

		if !ok {
			return stmts, pos, false
		}

		stmts = append(stmts, stmt)
	}

	return stmts, pos, true
}

// skipDOTWhitespace advances pos past spaces, tabs and newlines.
func skipDOTWhitespace(dot string, pos int) int {
	for pos < len(dot) && (dot[pos] == ' ' || dot[pos] == '\t' || dot[pos] == '\n' || dot[pos] == '\r') {
		pos++
	}

	return pos
}

// parseDOTSubgraph parses a "subgraph ... { ... }" statement, recursing into
// its children. It returns the statement and the position past its closing "}".
// The opening "{" is located with quoted-string and HTML label awareness so a
// head containing those characters cannot truncate the statement early.
func parseDOTSubgraph(dot string, pos int) (dotStatement, int, bool) {
	open := findDOTBlockOpen(dot, pos)
	if open < 0 {
		return dotStatement{}, pos, false
	}

	head := strings.TrimSpace(dot[pos:open])

	children, next, ok := parseDOTBlock(dot, open+1)
	if !ok {
		return dotStatement{}, next, false
	}

	return dotStatement{kind: "subgraph", head: head, children: children}, next, true
}

// parseDOTAttrStatement parses one attribute statement (graph/node/edge
// defaults, node or edge statement). It returns the statement and the position
// past its "];" terminator. Terminators are located with quoted-string and
// HTML label awareness: attribute values are user-authored (descriptions,
// technologies) and may contain "];", "{" or "}" inside quotes or HTML labels,
// so a raw "];" search would truncate the statement and silently shift the
// parse of everything after it. The ";" alone is not a safe terminator because
// HTML entity values (e.g. "&#x1F464;") contain semicolons.
func parseDOTAttrStatement(dot string, pos int) (dotStatement, int, bool) {
	end := findDOTAttrTerminator(dot, pos)
	if end < 0 {
		return dotStatement{}, pos, false
	}

	text := dot[pos:end]
	next := end + 2

	open := strings.Index(text, "[")
	if open < 0 {
		return dotStatement{kind: "bare", head: strings.TrimSpace(text)}, next, true
	}

	return dotStatement{
		kind:  "attr",
		head:  strings.TrimSpace(text[:open]),
		attrs: normalizeDOTAttrs(text[open+1:]),
	}, next, true
}

// scanDOTValueEnd advances pos past a DOT value region starting at pos: a
// double-quoted string (with backslash escapes) or an HTML label (<...>,
// which may nest). It returns the position just past the closing quote or ">",
// or len(dot) when the region is unterminated.
func scanDOTValueEnd(dot string, pos int) int {
	if dot[pos] == '"' {
		for i := pos + 1; i < len(dot); i++ {
			if dot[i] == '\\' {
				i++ // skip the escaped character

				continue
			}

			if dot[i] == '"' {
				return i + 1
			}
		}

		return len(dot)
	}

	depth := 1

	for i := pos + 1; i < len(dot); i++ {
		switch dot[i] {
		case '<':
			depth++
		case '>':
			depth--
			if depth == 0 {
				return i + 1
			}
		}
	}

	return len(dot)
}

// findDOTAttrTerminator returns the absolute index of the "];" that terminates
// the attribute statement starting at pos, skipping quoted strings and HTML
// labels inside attribute values. It returns -1 when the document ends before
// a terminator.
func findDOTAttrTerminator(dot string, pos int) int {
	for i := pos; i < len(dot); i++ {
		switch dot[i] {
		case '"', '<':
			i = scanDOTValueEnd(dot, i) - 1
		case ']':
			if i+1 < len(dot) && dot[i+1] == ';' {
				return i
			}
		}
	}

	return -1
}

// findDOTBlockOpen returns the absolute index of the "{" that opens the block
// of the subgraph statement starting at pos, skipping quoted strings and HTML
// labels in the head. It returns -1 when the document ends before an opening
// brace.
func findDOTBlockOpen(dot string, pos int) int {
	for i := pos; i < len(dot); i++ {
		switch dot[i] {
		case '"', '<':
			i = scanDOTValueEnd(dot, i) - 1
		case '{':
			return i
		}
	}

	return -1
}

// isGeometryAttr reports whether an attribute is layout geometry emitted by the
// go-graphviz engine. Its values flip between runs and carry no D-02 semantic
// content, so it is stripped from the canonical form.
func isGeometryAttr(name string) bool {
	switch name {
	case "bb", "pos", "lp", "lheight", "lwidth", "height", "width":
		return true
	}

	return false
}

// normalizeDOTAttrs splits an attribute block (between "[" and "]") into
// individual attributes, strips layout geometry, and sorts the remainder.
// Attributes are one per line in xdot output, so ",\n" is the separator;
// block content never contains a newline inside a value.
func normalizeDOTAttrs(block string) []string {
	raw := strings.Split(block, ",\n")
	attrs := make([]string, 0, len(raw))

	for _, piece := range raw {
		attr := strings.TrimSpace(piece)
		if attr == "" {
			continue
		}

		name, _, _ := strings.Cut(attr, "=")
		if isGeometryAttr(name) {
			continue
		}

		attrs = append(attrs, attr)
	}

	sort.Strings(attrs)

	return attrs
}

// serializeDOTStatements serializes a statement list canonically: statements
// are ordered by their canonical form so sibling order flips (DI-1) compare
// equal, preserving duplicates for multiset semantics.
func serializeDOTStatements(stmts []dotStatement) string {
	sorted := make([]string, 0, len(stmts))
	for _, stmt := range stmts {
		sorted = append(sorted, serializeDOTStatement(stmt))
	}

	sort.Strings(sorted)

	return strings.Join(sorted, "\x00")
}

// serializeDOTStatement serializes one statement: kind, head, sorted
// attributes, then recursively serialized children.
func serializeDOTStatement(stmt dotStatement) string {
	var b strings.Builder

	b.WriteString(stmt.kind)
	b.WriteByte('\x00')
	b.WriteString(stmt.head)

	for _, attr := range stmt.attrs {
		b.WriteByte('\x00')
		b.WriteString(attr)
	}

	if len(stmt.children) > 0 {
		b.WriteByte('\x00')
		b.WriteString(serializeDOTStatements(stmt.children))
	}

	return b.String()
}
