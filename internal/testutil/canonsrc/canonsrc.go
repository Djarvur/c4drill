// Package canonsrc provides canonical-equivalent normalizers for C4Drill
// source texts, realizing decision D-22 (the DI-1 canonicalDOT precedent
// applied to source formats): NormalizeTOML and NormalizeC4D parse a source
// document and re-serialize it through a canonical form in which whitespace,
// comments, literal quoting/representation choices, non-semantic key order
// and explicit defaults (arrow = forward, rank = forward, labelPosition =
// middle) normalize away, so round-trip tests compare canonical forms
// instead of bytes.
//
// The TOML path routes through parser.Parse — the semantic hub — so two TOML
// texts normalize equal iff they parse to canonically-equal Models. The C4D
// path routes through the exported AST entry c4d.ParseAST, dropping
// comments/positions and serializing statements in the AST's canonical
// section order with sorted non-semantic keys. Values containing newlines
// serialize through one canonical representation per format (D-06/D-22): a
// TOML multi-line basic string and a C4D triple-quoted block, so escaped-\n,
// multi-line and triple-quoted spellings of the same value compare equal.
// Both outputs are valid source that re-normalizes identically (the fixpoint
// property pinned by tests).
//
// This package is test-only by convention but lives outside a _test.go file
// so the toolchain exposes it on the importable package surface (the same
// layout as internal/testutil/canonical). It is pure (no I/O, no global
// state); testing.T is used only for t.Helper()/t.Fatalf error reporting —
// callers assert on the returned canonical form.
package canonsrc

import (
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/Djarvur/c4drill/internal/c4d"
	"github.com/Djarvur/c4drill/internal/c4d/ast"
	"github.com/Djarvur/c4drill/internal/model"
	"github.com/Djarvur/c4drill/internal/parser"
)

// NormalizeTOML normalizes TOML source text into its canonical form (D-22).
// Malformed input fails the test via t.Fatalf — the same contract as
// canonical.Canonical. The returned string is deterministic and valid TOML.
func NormalizeTOML(t *testing.T, src string) string {
	t.Helper()

	out, err := normalizeTOML(src)
	if err != nil {
		t.Fatalf("canonsrc: failed to normalize TOML source: %v", err)
	}

	return out
}

// NormalizeC4D normalizes C4D source text into its canonical form (D-22).
// Malformed input fails the test via t.Fatalf. The returned string is
// deterministic and valid C4D.
func NormalizeC4D(t *testing.T, src string) string {
	t.Helper()

	out, err := normalizeC4D(src)
	if err != nil {
		t.Fatalf("canonsrc: failed to normalize C4D source: %v", err)
	}

	return out
}

// normalizeTOML is the error-returning core behind NormalizeTOML.
func normalizeTOML(src string) (string, error) {
	m, err := parser.Parse([]byte(src))
	if err != nil {
		return "", fmt.Errorf("canonsrc: TOML parse: %w", err)
	}

	return serializeModelTOML(m), nil
}

// normalizeC4D is the error-returning core behind NormalizeC4D.
func normalizeC4D(src string) (string, error) {
	doc, err := c4d.ParseAST([]byte(src))
	if err != nil {
		return "", fmt.Errorf("canonsrc: C4D parse: %w", err)
	}

	return serializeDocC4D(doc), nil
}

// serializeModelTOML renders the Model as canonical TOML: [properties],
// unit tables in UnitOrder (recursively, links BEFORE subunit tables — TOML
// array-table headers attach to the most recent table), sorted-name
// [template.*] tables, [[use]] entries in order, then [[include]] entries.
func serializeModelTOML(m *parser.Model) string {
	var b strings.Builder

	writePropertiesCanonTOML(&b, m.Properties)

	for _, name := range m.UnitOrder {
		if unit, ok := m.Units[name]; ok {
			writeUnitCanonTOML(&b, name, unit, "")
		}
	}

	for _, name := range slices.Sorted(maps.Keys(m.Templates)) {
		writeTemplateCanonTOML(&b, name, m.Templates[name])
	}

	for _, inst := range m.Instantiations {
		writeUseCanonTOML(&b, "use", inst, true)
	}

	for _, inc := range m.Includes {
		b.WriteString("[[include]]\n")
		fmt.Fprintf(&b, "path = %s\n", canonicalTOMLValue(inc.Path))

		if inc.Once {
			b.WriteString("once = true\n")
		}
	}

	return b.String()
}

// writePropertiesCanonTOML writes the [properties] table with sorted keys.
func writePropertiesCanonTOML(b *strings.Builder, props model.Properties) {
	fields := make(map[string]string, 8)
	putString(fields, "name", props.Name)
	putString(fields, "description", props.Description)
	putString(fields, "color", props.Color)
	putString(fields, "style", props.Style)
	putString(fields, "border", props.Border)
	putString(fields, "edges", props.Edges)

	if props.LineLength != 0 {
		fields["lineLength"] = strconv.Itoa(props.LineLength)
	}

	if len(props.Expanded) > 0 {
		fields["expanded"] = canonicalTOMLArray(props.Expanded)
	}

	// LEG-01: legend nil == true — emit only when explicitly false, so the
	// canonical forms of a defaulting model and an explicit-true model match.
	if props.Legend != nil && !*props.Legend {
		fields["legend"] = "false"
	}

	if len(props.LegendLines) > 0 {
		items := make([]string, 0, len(props.LegendLines))

		for _, line := range props.LegendLines {
			item := line.Label + "|" + line.Color
			if line.Style != "" {
				item += "|" + line.Style
			}

			items = append(items, item)
		}

		fields["legendLine"] = canonicalTOMLArray(items)
	}

	if len(fields) == 0 {
		return
	}

	b.WriteString("[properties]\n")
	writeSortedFieldsTOML(b, fields)
}

// writeUnitCanonTOML writes one [path] table, its [[path.link]]/[[path.linkFrom]]
// tables, then its subunit tables recursively.
func writeUnitCanonTOML(b *strings.Builder, name string, unit *model.Unit, prefix string) {
	if unit == nil {
		return
	}

	path := joinDottedCanon(prefix, name)
	b.WriteString("[" + path + "]\n")
	writeSortedFieldsTOML(b, unitFieldMapTOML(unit))
	writeLinksCanonTOML(b, path, unit)

	for _, sub := range unit.SubunitOrder {
		if s, ok := unit.Subunits[sub]; ok {
			writeUnitCanonTOML(b, sub, s, path)
		}
	}
}

// unitFieldMapTOML collects a unit's present scalar/array/numeric fields.
// Width and height ride the canonical form like every other unit field, so
// round-trip comparisons catch a converter dropping them (F-02).
func unitFieldMapTOML(unit *model.Unit) map[string]string {
	fields := make(map[string]string, 11)
	putString(fields, "type", string(unit.Type))
	putString(fields, "name", unit.Name)
	putString(fields, "description", unit.Description)
	putString(fields, "technology", unit.Technology)
	putString(fields, "reference", unit.Reference)
	putString(fields, "color", unit.Color)
	putString(fields, "style", unit.Style)
	putString(fields, "border", unit.Border)
	putString(fields, "edges", unit.Edges)

	if unit.Width != 0 {
		fields["width"] = strconv.FormatFloat(unit.Width, 'f', -1, 64)
	}

	if unit.Height != 0 {
		fields["height"] = strconv.FormatFloat(unit.Height, 'f', -1, 64)
	}

	if len(unit.Expanded) > 0 {
		fields["expanded"] = canonicalTOMLArray(unit.Expanded)
	}

	return fields
}

// writeLinksCanonTOML writes the unit's link/linkFrom array tables.
// Validator-synthesized mirror links are bookkeeping and skipped.
func writeLinksCanonTOML(b *strings.Builder, path string, unit *model.Unit) {
	for _, link := range unit.Links {
		if !link.IsMirror() {
			writeLinkCanonTOML(b, joinDottedCanon(path, "link"), link)
		}
	}

	for _, link := range unit.LinksFrom {
		if !link.IsMirror() {
			writeLinkCanonTOML(b, joinDottedCanon(path, "linkFrom"), link)
		}
	}
}

// writeLinkCanonTOML writes one [[header]] table with sorted keys and the
// D-22 explicit defaults (arrow/rank/labelPosition) normalized away.
func writeLinkCanonTOML(b *strings.Builder, header string, link model.Link) {
	fields := make(map[string]string, 9)
	putString(fields, "peer", link.Peer)

	if link.Arrow != "" && link.Arrow != model.ArrowForward {
		fields["arrow"] = canonicalTOMLValue(string(link.Arrow))
	}

	if link.Rank != "" && link.Rank != model.RankForward {
		fields["rank"] = canonicalTOMLValue(string(link.Rank))
	}

	putString(fields, "kind", string(link.Kind))
	putString(fields, "color", link.Color)
	putString(fields, "style", link.Style)
	putString(fields, "technology", link.Technology)
	putString(fields, "description", link.Description)

	if link.LabelPosition != "" && link.LabelPosition != model.LabelMiddle {
		fields["labelPosition"] = canonicalTOMLValue(string(link.LabelPosition))
	}

	if link.Length != 0 {
		fields["length"] = strconv.Itoa(link.Length)
	}

	b.WriteString("[[" + header + "]]\n")
	writeSortedFieldsTOML(b, fields)
}

// writeTemplateCanonTOML writes the [template.<name>] table (params + root
// fields + root-level uses), then the subunit walk.
func writeTemplateCanonTOML(b *strings.Builder, name string, tmpl *parser.TemplateDef) {
	if tmpl == nil || tmpl.Unit == nil {
		return
	}

	header := joinDottedCanon("template", name)
	b.WriteString("[" + header + "]\n")

	fields := unitFieldMapTOML(tmpl.Unit)
	if len(tmpl.Params) > 0 {
		fields["params"] = canonicalTOMLArray(tmpl.Params)
	}

	writeSortedFieldsTOML(b, fields)

	for _, inst := range tmpl.Instantiations {
		if inst.Parent == "" {
			writeUseCanonTOML(b, header+".use", inst, false)
		}
	}

	writeTemplateSubunitsCanonTOML(b, name, "", tmpl.Unit, tmpl.Instantiations)
}

// writeTemplateSubunitsCanonTOML walks the template's subunit subtree writing
// each [template.<name>.<path>] table and the root-relative uses under it.
func writeTemplateSubunitsCanonTOML(
	b *strings.Builder,
	tmplName, relPath string,
	unit *model.Unit,
	insts []parser.Instantiation,
) {
	for _, subName := range unit.SubunitOrder {
		sub, ok := unit.Subunits[subName]
		if !ok {
			continue
		}

		subRel := joinDottedCanon(relPath, subName)
		header := "template." + tmplName + "." + subRel
		b.WriteString("[" + header + "]\n")
		writeSortedFieldsTOML(b, unitFieldMapTOML(sub))
		writeLinksCanonTOML(b, header, sub)

		for _, inst := range insts {
			if inst.Parent == subRel {
				writeUseCanonTOML(b, header+".use", inst, false)
			}
		}

		writeTemplateSubunitsCanonTOML(b, tmplName, subRel, sub, insts)
	}
}

// writeUseCanonTOML writes one [[header]] use table with sorted keys.
func writeUseCanonTOML(b *strings.Builder, header string, inst parser.Instantiation, withParent bool) {
	b.WriteString("[[" + header + "]]\n")

	fields := make(map[string]string, len(inst.Params)+2)
	putString(fields, "template", inst.Template)

	if withParent && inst.Parent != "" {
		putString(fields, "parent", inst.Parent)
	}

	for key, value := range inst.Params {
		fields[key] = canonicalTOMLValue(value)
	}

	writeSortedFieldsTOML(b, fields)
}

// writeSortedFieldsTOML writes a field map with sorted keys — the key-order
// normalization of D-22.
func writeSortedFieldsTOML(b *strings.Builder, fields map[string]string) {
	for _, key := range slices.Sorted(maps.Keys(fields)) {
		fmt.Fprintf(b, "%s = %s\n", key, fields[key])
	}
}

// putString records a non-empty field value in canonical quoted form.
func putString(fields map[string]string, key, value string) {
	if value != "" {
		fields[key] = canonicalTOMLValue(value)
	}
}

// joinDottedCanon joins two dotted path segments, treating "" as identity.
func joinDottedCanon(base, rel string) string {
	if base == "" {
		return rel
	}

	return base + "." + rel
}

// canonicalTOMLValue renders a value through the ONE canonical TOML
// representation: newline-containing values become multi-line basic strings
// (leading newline trimmed on re-parse, trailing newline escaped so it stays
// clear of the closing delimiter); everything else is a single-line basic
// string with full control escaping (T-35-04-02 policy).
func canonicalTOMLValue(s string) string {
	if strings.ContainsAny(s, "\n\r") {
		return `"""` + "\n" + escapeTOMLBody(s) + `"""`
	}

	return `"` + escapeTOMLBody(s) + `"`
}

// escapeTOMLBody escapes quotes, backslashes and control bytes, keeping
// newlines (and tabs) literal for the multi-line form; a TRAILING newline
// escapes so the value cannot bleed into the closing delimiter handling.
func escapeTOMLBody(s string) string {
	var b strings.Builder

	for i := range len(s) {
		c := s[i]

		switch {
		case c == '"':
			b.WriteString(`\"`)
		case c == '\\':
			b.WriteString(`\\`)
		case c == '\n' && i == len(s)-1:
			b.WriteString(`\n`)
		case c == '\n' || c == '\t':
			b.WriteByte(c)
		case c < 0x20 || c == 0x7f:
			fmt.Fprintf(&b, `\u00%02X`, c)
		default:
			b.WriteByte(c)
		}
	}

	return b.String()
}

// canonicalTOMLArray renders items as a canonical array of basic values.
func canonicalTOMLArray(items []string) string {
	parts := make([]string, len(items))
	for i, item := range items {
		parts[i] = canonicalTOMLValue(item)
	}

	return "[" + strings.Join(parts, ", ") + "]"
}

// serializeDocC4D renders the AST document canonically: properties, units,
// templates, uses, includes — the AST's canonical section order — with
// comments and positions dropped and non-semantic keys sorted.
func serializeDocC4D(doc *ast.Document) string {
	var b strings.Builder

	if doc.Properties != nil {
		writePropsCanonC4D(&b, doc.Properties)
	}

	for _, unit := range doc.Units {
		writeUnitCanonC4D(&b, unit, 0)
	}

	for _, tmpl := range doc.Templates {
		writeTemplateCanonC4D(&b, tmpl)
	}

	for _, us := range doc.UseStmts {
		writeUseLineC4D(&b, us, "")
	}

	for _, inc := range doc.Includes {
		writeIncludeCanonC4D(&b, inc)
	}

	return b.String()
}

// writePropsCanonC4D writes the properties block with sorted field keys.
func writePropsCanonC4D(b *strings.Builder, props *ast.PropertiesBlock) {
	b.WriteString("properties {\n")
	writeFieldsCanonC4D(b, props.Fields, "  ")
	b.WriteString("}\n")
}

// writeUnitCanonC4D writes one unit block; an empty body collapses to the
// single-line `{ }` form. Subunit and statement order is AST order.
func writeUnitCanonC4D(b *strings.Builder, node *ast.UnitNode, depth int) {
	if node == nil {
		return
	}

	inner := &strings.Builder{}
	writeFieldsCanonC4D(inner, node.Fields, "  ")
	writeEdgesCanonC4D(inner, node.Edges, "  ")

	for _, us := range node.UseStmts {
		writeUseLineC4D(inner, us, "  ")
	}

	for _, sub := range node.Subunits {
		writeUnitCanonC4D(inner, sub, 1)
	}

	indent := strings.Repeat("  ", depth)
	b.WriteString(indent + unitHeaderCanonC4D(node) + " ")

	if inner.Len() == 0 {
		b.WriteString("{ }\n")

		return
	}

	b.WriteString("{\n")
	b.WriteString(inner.String())
	b.WriteString(indent + "}\n")
}

// unitHeaderCanonC4D renders `id`, `id: type external?`, `type external?`
// or the full `id: type external? "Name"` header (grammar-shaped so the
// canonical form re-parses).
func unitHeaderCanonC4D(node *ast.UnitNode) string {
	var h strings.Builder

	if node.ID != "" {
		h.WriteString(node.ID)
	}

	if node.Type != "" {
		if h.Len() > 0 {
			h.WriteString(": ")
		}

		h.WriteString(node.Type)

		if node.External {
			h.WriteString(" external")
		}
	}

	if node.Name != "" {
		if h.Len() > 0 {
			h.WriteString(" ")
		}

		h.WriteString(canonicalC4DValue(node.Name))
	}

	return h.String()
}

// writeFieldsCanonC4D writes field statements sorted by key (the key-order
// normalization); literal KINDS drop away — the value is canonical.
func writeFieldsCanonC4D(b *strings.Builder, fields []*ast.FieldStmt, indent string) {
	sorted := slices.Clone(fields)
	slices.SortFunc(sorted, func(a, b *ast.FieldStmt) int {
		return strings.Compare(a.Key, b.Key)
	})

	for _, f := range sorted {
		if f != nil {
			fmt.Fprintf(b, "%s%s: %s\n", indent, f.Key, literalCanonC4D(f.Value))
		}
	}
}

// literalCanonC4D canonicalizes an AST literal by VALUE: lists render as
// bracket arrays, scalars through canonicalC4DValue regardless of the
// authored literal kind.
func literalCanonC4D(lit ast.Literal) string {
	if lit.Kind == ast.KindList {
		return canonicalC4DArray(lit.List)
	}

	return canonicalC4DValue(lit.Str)
}

// writeEdgesCanonC4D writes edge statements in AST order: glyph, peer, the
// D-09 pipe-shorthand label, then the option block with sorted keys and the
// D-22 defaults (arrow = forward, rank = forward, labelPosition = middle)
// normalized away.
func writeEdgesCanonC4D(b *strings.Builder, edges []*ast.EdgeStmt, indent string) {
	for _, edge := range edges {
		if edge == nil {
			continue
		}

		fmt.Fprintf(b, "%s%s %s", indent, edge.ArrowGlyph, edge.Peer)

		if label := edgeLabelCanonC4D(edge); label != "" {
			b.WriteString(": " + label)
		}

		if opts := edgeOptsCanonC4D(edge.Options); len(opts) > 0 {
			b.WriteString(" { " + strings.Join(opts, " ") + " }")
		}

		b.WriteByte('\n')
	}
}

// edgeLabelCanonC4D renders the pipe-shorthand label: `"tech | desc"`,
// `"tech |"` or `"desc"`; the quoted form keeps embedded pipes round-trip
// safe (splitPipeLabel splits on the FIRST pipe only).
func edgeLabelCanonC4D(edge *ast.EdgeStmt) string {
	switch {
	case edge.Technology != "" && edge.Description != "":
		return canonicalC4DValue(edge.Technology + " | " + edge.Description)
	case edge.Technology != "":
		return canonicalC4DValue(edge.Technology + " |")
	case edge.Description != "":
		return canonicalC4DValue(edge.Description)
	default:
		return ""
	}
}

// edgeOptsCanonC4D renders the option block entries sorted by key, skipping
// D-22 explicit defaults.
func edgeOptsCanonC4D(options []*ast.FieldStmt) []string {
	entries := make(map[string]string, len(options))

	for _, opt := range options {
		if opt == nil {
			continue
		}

		value := literalCanonC4D(opt.Value)
		if isCanonDefaultEdgeOpt(opt.Key, value) {
			continue
		}

		entries[opt.Key] = opt.Key + ": " + value
	}

	out := make([]string, 0, len(entries))
	for _, key := range slices.Sorted(maps.Keys(entries)) {
		out = append(out, entries[key])
	}

	return out
}

// isCanonDefaultEdgeOpt reports whether an edge option carries a D-22
// explicit default (equal to the omitted field).
func isCanonDefaultEdgeOpt(key, canonicalValue string) bool {
	switch key {
	case "arrow":
		return canonicalValue == `"forward"`
	case "rank":
		return canonicalValue == `"forward"`
	case "labelPosition":
		return canonicalValue == `"middle"`
	default:
		return false
	}
}

// writeTemplateCanonC4D writes a template declaration: params in declared
// order (semantic), then the body — root type statement first when the AST
// records one, sorted fields, edges, uses, subunits.
func writeTemplateCanonC4D(b *strings.Builder, tmpl *ast.TemplateDecl) {
	if tmpl == nil {
		return
	}

	fmt.Fprintf(b, "template %s(%s) {\n", tmpl.Name, strings.Join(tmpl.Params, ", "))

	if body := tmpl.Body; body != nil {
		writeTemplateRootTypeC4D(b, body)
		writeFieldsCanonC4D(b, body.Fields, "  ")
		writeEdgesCanonC4D(b, body.Edges, "  ")

		for _, us := range body.UseStmts {
			writeUseLineC4D(b, us, "  ")
		}

		for _, sub := range body.Subunits {
			writeUnitCanonC4D(b, sub, 1)
		}
	}

	b.WriteString("}\n")
}

// writeTemplateRootTypeC4D renders the template root's type statement when
// the AST records one (non-default root types; default roots stay implicit).
func writeTemplateRootTypeC4D(b *strings.Builder, body *ast.UnitNode) {
	if body.Type == "" {
		return
	}

	stmt := "  type: " + body.Type
	if body.External {
		stmt += " external"
	}

	b.WriteString(stmt + "\n")
}

// writeUseLineC4D writes one use statement with args sorted by (name, value)
// — the Model stores named params as a map, so name order is non-semantic.
func writeUseLineC4D(b *strings.Builder, us *ast.UseStmt, indent string) {
	if us == nil {
		return
	}

	type argPair struct{ name, value string }

	pairs := make([]argPair, 0, len(us.Args))
	for _, arg := range us.Args {
		pairs = append(pairs, argPair{name: arg.Name, value: literalCanonC4D(arg.Value)})
	}

	slices.SortFunc(pairs, func(a, b argPair) int {
		if c := strings.Compare(a.name, b.name); c != 0 {
			return c
		}

		return strings.Compare(a.value, b.value)
	})

	parts := make([]string, len(pairs))
	for i, p := range pairs {
		if p.name != "" {
			parts[i] = p.name + ": " + p.value
		} else {
			parts[i] = p.value
		}
	}

	fmt.Fprintf(b, "%suse %s(%s)\n", indent, us.Template, strings.Join(parts, ", "))
}

// writeIncludeCanonC4D writes one include statement with a quoted path.
func writeIncludeCanonC4D(b *strings.Builder, inc *ast.IncludeStmt) {
	if inc == nil {
		return
	}

	stmt := "include " + canonicalC4DValue(inc.Path)
	if inc.Once {
		stmt += " once"
	}

	b.WriteString(stmt + "\n")
}

// canonicalC4DValue renders a scalar through the ONE canonical C4D
// representation (D-06/D-22): newline-containing values become raw
// triple-quoted blocks (falling back to the escaped quoted form when the
// value itself contains the delimiter); everything else is a double-quoted
// string with quote/backslash/newline escapes.
func canonicalC4DValue(s string) string {
	if strings.ContainsAny(s, "\n\r") && !strings.Contains(s, `"""`) {
		return `"""` + s + `"""`
	}

	return `"` + escapeC4DBody(s) + `"`
}

// escapeC4DBody escapes the characters the DoubleQuoted grammar would
// otherwise consume (quotes, backslashes) plus newlines; non-EOL control
// bytes stay raw — the grammar accepts them and unescapeDoubleQuoted
// reproduces them verbatim.
func escapeC4DBody(s string) string {
	var b strings.Builder

	for i := range len(s) {
		switch c := s[i]; c {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		default:
			b.WriteByte(c)
		}
	}

	return b.String()
}

// canonicalC4DArray renders items as a canonical bracket list.
func canonicalC4DArray(items []string) string {
	parts := make([]string, len(items))
	for i, item := range items {
		parts[i] = canonicalC4DValue(item)
	}

	return "[" + strings.Join(parts, ", ") + "]"
}
