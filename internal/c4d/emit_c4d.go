package c4d

import (
	"fmt"
	"slices"
	"strings"

	"github.com/Djarvur/c4drill/internal/c4d/ast"
)

// unitFieldRank is the D-23 canonical field order for unit and template-body
// fields (type lives in the header; name may appear as a body field).
//
//nolint:gochecknoglobals // rank table is immutable after init
var unitFieldRank = map[string]int{
	"name":        0,
	"description": 1,
	"technology":  2,
	"reference":   3,
	"color":       4,
	"style":       5,
	"border":      6,
	"edges":       7,
	"width":       8,
	"height":      9,
	"expanded":    10,
}

// propertyFieldRank is the D-23 canonical field order for the properties
// block (the TOML [properties] key set, D-12).
//
//nolint:gochecknoglobals // rank table is immutable after init
var propertyFieldRank = map[string]int{
	"name":        0,
	"description": 1,
	"color":       2,
	"style":       3,
	"border":      4,
	"edges":       5,
	"lineLength":  6,
	"expanded":    7,
}

// edgeOptionRank is the canonical order of edge option-block keys.
//
//nolint:gochecknoglobals // rank table is immutable after init
var edgeOptionRank = map[string]int{
	"arrow":         0,
	"rank":          1,
	"color":         2,
	"style":         3,
	"labelPosition": 4,
	"length":        5,
}

// unknownFieldRank sorts unrecognized keys after every known one.
const unknownFieldRank = 100

// stmtComments is the normalized comment placement for one statement: lead
// comments render on their own lines immediately above the statement, tail
// renders same-line after it (gofmt semantics, D-32).
type stmtComments struct {
	leads []*ast.Comment
	tail  *ast.Comment
}

// renderedStmt is one statement of a render list: its kind (top-level blank
// rules), attached comments, source line, and its renderer (which receives
// the allocated placement).
type renderedStmt struct {
	kind     string
	comments []*ast.Comment
	pos      int
	sc       stmtComments
	render   func(b *strings.Builder, sc stmtComments)
}

// EmitC4D renders doc as C4D text in the canonical style (D-33): compact
// one-line leaf blocks (no subunits, no edges, no uses, at most 3 fields,
// single-line-able values), nested units multi-line with 2-space indent per
// depth, fields in the fixed D-23 order, edges after fields, subunits last,
// and one blank line between top-level units and between statements of
// different kinds.
//
// Comments render in the self-consistent placement the grammar re-attaches
// identically (D-32, T-35-04-01): a comment at or below its statement's line
// is the statement's same-line tail, and a statement's first comment
// migrates to the previous statement's tail when that statement has none —
// exactly where re-parsing attaches it (the grammar's StmtEnd lets the first
// comment after a statement, across any whitespace, ride that statement).
// With this allocation emit(parse(emit(doc))) is byte-identical to
// emit(doc); emission also walks statement slices only, never map
// iteration, so output is deterministic.
func EmitC4D(doc *ast.Document) string {
	if doc == nil {
		return ""
	}

	top := topLevelStatements(doc)
	trailing := allocateTails(top, doc.TrailingComments)

	var b strings.Builder

	prevKind := ""

	for i, s := range top {
		if i > 0 && (s.kind != prevKind || s.kind == "unit" || prevKind == "unit") {
			b.WriteString("\n")
		}

		prevKind = s.kind
		s.render(&b, s.sc)
	}

	writeLeadComments(&b, trailing, 0)

	return b.String()
}

// topLevelStatements builds the document-level render list in the canonical
// statement order: properties, units, templates, uses, includes.
func topLevelStatements(doc *ast.Document) []renderedStmt {
	top := make([]renderedStmt, 0, len(doc.Units)+len(doc.Templates)+len(doc.UseStmts)+len(doc.Includes)+1)

	if doc.Properties != nil {
		props := doc.Properties
		top = append(top, renderedStmt{
			kind: "properties", comments: props.Comments, pos: props.Pos,
			render: func(b *strings.Builder, sc stmtComments) { emitPropertiesC4D(b, props, sc) },
		})
	}

	for _, unit := range doc.Units {
		top = append(top, unitRenderStmt(unit, 0))
	}

	for _, tmpl := range doc.Templates {
		tmplDecl := tmpl
		top = append(top, renderedStmt{
			kind: "template", comments: tmplDecl.Comments, pos: tmplDecl.Pos,
			render: func(b *strings.Builder, sc stmtComments) { emitTemplateC4D(b, tmplDecl, sc) },
		})
	}

	for _, us := range doc.UseStmts {
		useStmt := us
		top = append(top, renderedStmt{
			kind: "use", comments: useStmt.Comments, pos: useStmt.Pos,
			render: func(b *strings.Builder, sc stmtComments) {
				writeLeadComments(b, sc.leads, 0)
				b.WriteString(indentC4D(0))
				emitUseStmtC4D(b, useStmt)
				writeTailComment(b, sc.tail)
				b.WriteString("\n")
			},
		})
	}

	for _, inc := range doc.Includes {
		incStmt := inc
		top = append(top, renderedStmt{
			kind: "include", comments: incStmt.Comments, pos: incStmt.Pos,
			render: func(b *strings.Builder, sc stmtComments) {
				writeLeadComments(b, sc.leads, 0)
				b.WriteString(indentC4D(0))
				emitIncludeC4D(b, incStmt)
				writeTailComment(b, sc.tail)
				b.WriteString("\n")
			},
		})
	}

	return top
}

// unitRenderStmt wraps one unit block as a renderable statement at depth.
func unitRenderStmt(unit *ast.UnitNode, depth int) renderedStmt {
	return renderedStmt{
		kind: "unit", comments: unit.Comments, pos: unit.Pos,
		render: func(b *strings.Builder, sc stmtComments) { emitUnitC4D(b, unit, depth, sc) },
	}
}

// allocateTails computes the self-consistent comment placement for a render
// list plus its trailing comments (see EmitC4D): statements first split
// their comments Pos-aware (a comment at or below the statement's own line
// is its tail), then each statement's first lead migrates to the previous
// statement when that statement has no tail yet — where re-parsing attaches
// it. Returns the trailing comments that remain at the list's end.
func allocateTails(list []renderedStmt, trailing []*ast.Comment) []*ast.Comment {
	for i := range list {
		list[i].sc = splitStmtComments(list[i].comments, list[i].pos)
	}

	for i := 1; i < len(list); i++ {
		if len(list[i].sc.leads) == 0 || list[i-1].sc.tail != nil {
			continue
		}

		list[i-1].sc.tail = list[i].sc.leads[0]
		list[i].sc.leads = list[i].sc.leads[1:]
	}

	remaining := trailing
	if n := len(list); n > 0 && len(remaining) > 0 && list[n-1].sc.tail == nil {
		list[n-1].sc.tail = remaining[0]
		remaining = remaining[1:]
	}

	return remaining
}

// splitStmtComments splits attached comments into leads and an optional
// same-line tail: the grammar appends a statement's tail comment last, and a
// tail sits at or below the statement's own line (Pos >= stmt.Pos). Zero
// positions (hand-built ASTs) never produce tails.
func splitStmtComments(cs []*ast.Comment, stmtPos int) stmtComments {
	if n := len(cs); n > 0 && stmtPos > 0 && cs[n-1].Pos >= stmtPos {
		return stmtComments{leads: cs[:n-1], tail: cs[n-1]}
	}

	return stmtComments{leads: cs}
}

// emitPropertiesC4D writes the properties block: leads above, fields in the
// D-23 property order, the block's trailing comments inside, tail after the
// closing brace.
func emitPropertiesC4D(b *strings.Builder, props *ast.PropertiesBlock, sc stmtComments) {
	writeLeadComments(b, sc.leads, 0)
	b.WriteString("properties {\n")

	inner := make([]renderedStmt, 0, len(props.Fields))
	for _, f := range sortedFields(props.Fields, propertyFieldRank) {
		field := f
		inner = append(inner, renderedStmt{
			comments: field.Comments, pos: field.Pos,
			render: func(b *strings.Builder, sc stmtComments) {
				writeLeadComments(b, sc.leads, 1)
				b.WriteString(indentC4D(1) + field.Key + ": " + literalC4D(field.Value))
				writeTailComment(b, sc.tail)
				b.WriteString("\n")
			},
		})
	}

	trailing := allocateTails(inner, props.TrailingComments)

	for _, s := range inner {
		s.render(b, s.sc)
	}

	writeLeadComments(b, trailing, 1)
	b.WriteString("}")
	writeTailComment(b, sc.tail)
	b.WriteString("\n")
}

// emitUnitC4D writes one unit block. Leaf units emit on a single line
// (planner-pinned D-33 rule); everything else emits multi-line with fields
// (D-23 order), then edges, then use statements, then subunits — each list
// running through the same comment allocation.
func emitUnitC4D(b *strings.Builder, unit *ast.UnitNode, depth int, sc stmtComments) {
	if unit == nil {
		return
	}

	writeLeadComments(b, sc.leads, depth)
	b.WriteString(indentC4D(depth) + unitHeaderC4D(unit))

	fields := sortedFields(unit.Fields, unitFieldRank)

	if unitIsLeaf(unit, fields) {
		b.WriteString(" {")

		for i, f := range fields {
			if i > 0 {
				b.WriteString(";")
			}

			b.WriteString(" " + f.Key + ": " + literalC4D(f.Value))
		}

		b.WriteString(" }")
		writeTailComment(b, sc.tail)
		b.WriteString("\n")

		return
	}

	b.WriteString(" {\n")

	for _, s := range unitInnerStatements(unit, fields, depth+1) {
		s.render(b, s.sc)
	}

	b.WriteString(indentC4D(depth) + "}")
	writeTailComment(b, sc.tail)
	b.WriteString("\n")
}

// unitInnerStatements builds the render list of a unit or template-body
// block: canonical-order fields, then edges, then use statements, then
// subunits, with the block's trailing comments allocated at the list's end.
func unitInnerStatements(unit *ast.UnitNode, fields []*ast.FieldStmt, depth int) []renderedStmt {
	inner := make([]renderedStmt, 0, len(fields)+len(unit.Edges)+len(unit.UseStmts)+len(unit.Subunits))

	for _, f := range fields {
		field := f
		inner = append(inner, renderedStmt{
			comments: field.Comments, pos: field.Pos,
			render: func(b *strings.Builder, sc stmtComments) {
				writeLeadComments(b, sc.leads, depth)
				b.WriteString(indentC4D(depth) + field.Key + ": " + literalC4D(field.Value))
				writeTailComment(b, sc.tail)
				b.WriteString("\n")
			},
		})
	}

	for _, e := range unit.Edges {
		edge := e
		inner = append(inner, renderedStmt{
			comments: edge.Comments, pos: edge.Pos,
			render: func(b *strings.Builder, sc stmtComments) {
				writeLeadComments(b, sc.leads, depth)
				b.WriteString(indentC4D(depth))
				emitEdgeC4D(b, edge)
				writeTailComment(b, sc.tail)
				b.WriteString("\n")
			},
		})
	}

	for _, us := range unit.UseStmts {
		useStmt := us
		inner = append(inner, renderedStmt{
			comments: useStmt.Comments, pos: useStmt.Pos,
			render: func(b *strings.Builder, sc stmtComments) {
				writeLeadComments(b, sc.leads, depth)
				b.WriteString(indentC4D(depth))
				emitUseStmtC4D(b, useStmt)
				writeTailComment(b, sc.tail)
				b.WriteString("\n")
			},
		})
	}

	for _, sub := range unit.Subunits {
		inner = append(inner, unitRenderStmt(sub, depth))
	}

	allocateTails(inner, unit.TrailingComments)

	return inner
}

// unitIsLeaf reports the planner-pinned compact-leaf condition (D-33): no
// subunits, no edges, no use statements, no inner comments, at most 3 fields
// and only single-line-able values (a triple-quoted value spans lines and
// forces a multi-line block).
func unitIsLeaf(unit *ast.UnitNode, sorted []*ast.FieldStmt) bool {
	if len(unit.Subunits) > 0 || len(unit.Edges) > 0 || len(unit.UseStmts) > 0 {
		return false
	}

	if len(unit.TrailingComments) > 0 {
		return false
	}

	if len(sorted) > 3 {
		return false
	}

	for _, f := range sorted {
		if f.Value.Kind == ast.KindTriple || len(f.Comments) > 0 {
			return false
		}
	}

	return true
}

// unitHeaderC4D renders `id: type external? "Name"?` (D-02/D-04). Names
// always emit quoted, so a name containing braces cannot break out of the
// block structure (T-35-04-02).
func unitHeaderC4D(unit *ast.UnitNode) string {
	var h strings.Builder

	if unit.ID != "" {
		h.WriteString(unit.ID)
	}

	if unit.Type != "" {
		if h.Len() > 0 {
			h.WriteString(": ")
		}

		h.WriteString(unit.Type)

		if unit.External {
			h.WriteString(" external")
		}
	}

	if unit.Name != "" {
		if h.Len() > 0 {
			h.WriteString(" ")
		}

		h.WriteString(quoteC4D(unit.Name))
	}

	return h.String()
}

// emitTemplateC4D writes a template declaration: the param list, the root
// type statement when the AST records one (the D-22 text form of a TOML
// template's root `type` key — non-default root types must survive
// round-trips), then the body's fields (D-23 order), edges, uses and
// subunits — the same body machinery units use.
func emitTemplateC4D(b *strings.Builder, tmpl *ast.TemplateDecl, sc stmtComments) {
	if tmpl == nil {
		return
	}

	writeLeadComments(b, sc.leads, 0)
	fmt.Fprintf(b, "template %s(%s) {\n", tmpl.Name, strings.Join(tmpl.Params, ", "))

	if body := tmpl.Body; body != nil {
		if body.Type != "" {
			stmt := "  type: " + body.Type
			if body.External {
				stmt += " external"
			}

			b.WriteString(stmt + "\n")
		}

		for _, s := range unitInnerStatements(body, sortedFields(body.Fields, unitFieldRank), 1) {
			s.render(b, s.sc)
		}
	}

	b.WriteString("}")
	writeTailComment(b, sc.tail)
	b.WriteString("\n")
}

// emitEdgeC4D writes one edge statement: glyph, peer, the D-09 pipe-shorthand
// label and the optional trailing option block.
func emitEdgeC4D(b *strings.Builder, edge *ast.EdgeStmt) {
	if edge == nil {
		return
	}

	b.WriteString(edge.ArrowGlyph + " " + edge.Peer)

	switch {
	case edge.Technology != "" && edge.Description != "":
		b.WriteString(": " + quoteC4D(edge.Technology+" | "+edge.Description))
	case edge.Technology != "":
		b.WriteString(": " + quoteC4D(edge.Technology+" |"))
	case edge.Description != "":
		if c4dBarewordSafe(edge.Description) {
			b.WriteString(": " + edge.Description)
		} else {
			b.WriteString(": " + quoteC4D(edge.Description))
		}
	}

	if opts := sortedFields(edge.Options, edgeOptionRank); len(opts) > 0 {
		b.WriteString(" { ")

		for i, opt := range opts {
			if i > 0 {
				b.WriteString(" ")
			}

			b.WriteString(opt.Key + ": " + literalC4D(opt.Value))
		}

		b.WriteString(" }")
	}
}

// emitUseStmtC4D writes one use statement with named (`key: value`) or
// positional args in source order. The caller appends the tail comment and
// the newline.
func emitUseStmtC4D(b *strings.Builder, us *ast.UseStmt) {
	if us == nil {
		return
	}

	fmt.Fprintf(b, "use %s(", us.Template)

	for i, arg := range us.Args {
		if i > 0 {
			b.WriteString(", ")
		}

		if arg.Name != "" {
			b.WriteString(arg.Name + ": ")
		}

		b.WriteString(literalC4D(arg.Value))
	}

	b.WriteString(")")
}

// emitIncludeC4D writes one include directive; the path emits bareword when
// it matches the include-path charset, quoted otherwise. The caller appends
// the tail comment and the newline.
func emitIncludeC4D(b *strings.Builder, inc *ast.IncludeStmt) {
	if inc == nil {
		return
	}

	b.WriteString("include ")

	if includeBarewordSafe(inc.Path) {
		b.WriteString(inc.Path)
	} else {
		b.WriteString(quoteC4D(inc.Path))
	}

	if inc.Once {
		b.WriteString(" once")
	}
}

// writeLeadComments writes lead comments on their own lines immediately
// above their statement (D-32).
func writeLeadComments(b *strings.Builder, leads []*ast.Comment, depth int) {
	for _, c := range leads {
		if c.Text == "" {
			b.WriteString(indentC4D(depth) + "#\n")

			continue
		}

		b.WriteString(indentC4D(depth) + "# " + c.Text + "\n")
	}
}

// writeTailComment writes a same-line trailing comment (gofmt semantics).
func writeTailComment(b *strings.Builder, tail *ast.Comment) {
	if tail == nil {
		return
	}

	if tail.Text == "" {
		b.WriteString(" #")

		return
	}

	b.WriteString(" # " + tail.Text)
}

// sortedFields returns the fields sorted by canonical rank (unknown keys
// after known ones, alphabetical within a rank) — D-23 normalization.
func sortedFields(fields []*ast.FieldStmt, rank map[string]int) []*ast.FieldStmt {
	out := slices.Clone(fields)

	slices.SortStableFunc(out, func(a, b *ast.FieldStmt) int {
		if ra, rb := fieldRank(a.Key, rank), fieldRank(b.Key, rank); ra != rb {
			return ra - rb
		}

		return strings.Compare(a.Key, b.Key)
	})

	return out
}

// fieldRank returns the canonical rank of key, unknownFieldRank when absent.
func fieldRank(key string, rank map[string]int) int {
	if r, ok := rank[key]; ok {
		return r
	}

	return unknownFieldRank
}

// literalC4D renders a literal in its canonical form per Kind: barewords
// verbatim, quoted strings escaped, triple-quoted raw blocks (multi-line
// values, D-06), and lists as inline bracket form (D-15).
func literalC4D(lit ast.Literal) string {
	switch lit.Kind {
	case ast.KindTriple:
		return `"""` + lit.Str + `"""`
	case ast.KindList:
		items := make([]string, len(lit.List))

		for i, item := range lit.List {
			if c4dBarewordSafe(item) {
				items[i] = item
			} else {
				items[i] = quoteC4D(item)
			}
		}

		return "[" + strings.Join(items, ", ") + "]"
	case ast.KindQuoted:
		return quoteC4D(lit.Str)
	case ast.KindBareword:
		return lit.Str
	default: // zero-value Kind on hand-built literals renders as bareword
		return lit.Str
	}
}

// quoteC4D renders s as a double-quoted C4D literal, escaping quotes,
// backslashes and line breaks so the value parses back to the identical
// string (T-35-04-02). Other bytes — control characters included — pass
// through raw, which the quoted-literal grammar accepts.
func quoteC4D(s string) string {
	var b strings.Builder
	b.Grow(len(s) + 2)
	b.WriteByte('"')

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
		case '\t':
			b.WriteString(`\t`)
		default:
			b.WriteByte(c)
		}
	}

	b.WriteByte('"')

	return b.String()
}

// c4dBarewordSafe reports whether s can be emitted as a bareword value:
// non-empty, free of the grammar's bareword stop characters, and free of
// leading/trailing whitespace (barewords trim at parse, D-06).
func c4dBarewordSafe(s string) bool {
	if s == "" || strings.TrimSpace(s) != s {
		return false
	}

	return !strings.ContainsAny(s, "{};|,\"#\n\r\t")
}

// includeBarewordSafe reports whether path matches the include-path charset
// ([A-Za-z0-9_./-]+, D-14) and can emit unquoted.
func includeBarewordSafe(path string) bool {
	if path == "" {
		return false
	}

	for _, c := range path {
		switch {
		case c >= '0' && c <= '9', c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z':
		case c == '.' || c == '/' || c == '_' || c == '-':
		default:
			return false
		}
	}

	return true
}

// indentC4D renders the 2-space-per-depth indentation (D-33).
func indentC4D(depth int) string {
	return strings.Repeat("  ", depth)
}
