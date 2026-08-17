// Package ast defines the typed syntax tree for the C4D DSL. It is produced
// by the pigeon-generated parser (internal/c4d/grammar) and consumed by the
// c4d front-end; unlike *parser.Model it carries source positions and
// attached comments, which the formatter (D-32) needs to preserve trivia.
package ast

// LiteralKind discriminates the literal forms of the DSL (D-06, D-15).
type LiteralKind string

// Literal kinds: barewords, double-quoted strings (escape-aware),
// triple-quoted raw multi-line strings, and bracketed lists.
const (
	// KindBareword is an unquoted value; edge whitespace is trimmed.
	KindBareword LiteralKind = "bareword"
	// KindQuoted is a double-quoted value; escapes processed, whitespace
	// preserved verbatim.
	KindQuoted LiteralKind = "quoted"
	// KindTriple is a triple-quoted raw multi-line value.
	KindTriple LiteralKind = "triple"
	// KindList is a bracketed list of items (inline or one-per-line, D-15).
	KindList LiteralKind = "list"
)

// Literal is a field or option value in one of the DSL literal forms.
type Literal struct {
	// Kind discriminates which literal form was written.
	Kind LiteralKind
	// Str holds the unquoted value for KindBareword, KindQuoted and
	// KindTriple.
	Str string
	// List holds the item values for KindList.
	List []string
}

// Comment is a `#` line comment. Comments attach to the following
// statement (D-32: the formatter needs trivia in the tree); comments with
// no following statement attach to the enclosing node's trailing list.
type Comment struct {
	// Pos is the 1-indexed line the comment starts on.
	Pos int
	// Text is the comment content without the leading '#' and trimmed.
	Text string
}

// Document is the root of a parsed .c4d file.
type Document struct {
	// Pos is the 1-indexed line the document starts on.
	Pos int
	// Properties is the `properties { }` block, nil when absent.
	Properties *PropertiesBlock
	// Units holds the top-level units in statement order.
	Units []*UnitNode
	// Templates holds the `template name(params) { ... }` declarations in
	// statement order (D-13).
	Templates []*TemplateDecl
	// UseStmts holds the top-level `use name(args)` instantiations in
	// statement order (D-13); uses inside unit blocks and template bodies
	// live on those nodes instead.
	UseStmts []*UseStmt
	// Includes holds the `include path [once]` directives in statement
	// order (D-14); resolution is include.Resolve's job at pipeline time.
	Includes []*IncludeStmt
	// TrailingComments holds comments at the end of the document with no
	// following statement.
	TrailingComments []*Comment
}

// PropertiesBlock is the top-level `properties { }` block (D-12) with the
// same key set as the TOML [properties] table.
type PropertiesBlock struct {
	// Pos is the 1-indexed line the block starts on.
	Pos int
	// Fields holds the key/value statements in source order.
	Fields []*FieldStmt
	// Comments are the comments attached to this block.
	Comments []*Comment
	// TrailingComments holds comments inside the block with no following
	// statement.
	TrailingComments []*Comment
}

// UnitNode is a brace-block unit (D-01, D-02). ID is empty for type-led
// headers and Type is empty when omitted — both are inferred later, when
// the AST is converted to a *parser.Model (never at parse time).
type UnitNode struct {
	// Pos is the 1-indexed line the unit header starts on.
	Pos int
	// ID is the unit identifier ([A-Za-z0-9_-]+, D-07); empty when the
	// header was type-led.
	ID string
	// Type is the exact TOML type keyword (D-03); empty when omitted.
	Type string
	// External records the `external` modifier (D-04).
	External bool
	// Name is the quoted display name from the header; empty when omitted.
	Name string
	// Fields holds the field statements in source order.
	Fields []*FieldStmt
	// Edges holds the edge statements in source order (D-08).
	Edges []*EdgeStmt
	// Subunits holds the nested units in source order (D-01).
	Subunits []*UnitNode
	// UseStmts holds `use` instantiations inside the block (D-16) — the
	// enclosing unit is the instantiation's parent.
	UseStmts []*UseStmt
	// Comments are the comments attached to this unit.
	Comments []*Comment
	// TrailingComments holds comments inside the block with no following
	// statement.
	TrailingComments []*Comment
}

// EdgeStmt is an in-block edge statement (D-05, D-08, D-09). ArrowGlyph is
// the raw source glyph; the mapping to model.ArrowDirection happens in the
// Model conversion, not here.
type EdgeStmt struct {
	// Pos is the 1-indexed line the statement starts on.
	Pos int
	// ArrowGlyph is one of "->", "<-", "<->", "--".
	ArrowGlyph string
	// Peer is the peer reference — bare name or dotted path (D-07, D-10).
	Peer string
	// Technology is the tech part of the pipe shorthand; empty when absent.
	Technology string
	// Description is the description part; a single un-piped value lands
	// here (D-09, user's explicit choice).
	Description string
	// Options holds the trailing brace-block options in source order.
	Options []*FieldStmt
	// Comments are the comments attached to this statement.
	Comments []*Comment
}

// FieldStmt is a key/value statement — unit fields, properties fields, and
// edge options all share this shape.
type FieldStmt struct {
	// Pos is the 1-indexed line the statement starts on.
	Pos int
	// Key is the field keyword (e.g. description, color, rank).
	Key string
	// Value is the literal value.
	Value Literal
	// Comments are the comments attached to this statement.
	Comments []*Comment
}

// UseStmt is a `use name(args)` instantiation statement (D-13). It is legal
// in three positions: top level (Document.UseStmts), inside unit blocks
// (UnitNode.UseStmts, D-16 — the enclosing unit is the parent) and inside
// template bodies (TemplateDecl.Body.UseStmts, D-17 — the parent is relative
// to the template's unit root). The fields map 1:1 onto
// parser.Instantiation: Template -> Template, the enclosing position ->
// Parent, Args -> Params.
type UseStmt struct {
	// Pos is the 1-indexed line the statement starts on.
	Pos int
	// Template is the template name.
	Template string
	// Args holds the supplied arguments in source order. This single ordered
	// representation is canonical for BOTH authoring forms (the plan's
	// one-representation decision): named args carry their key in Name,
	// positional args carry an empty Name and pair positionally with
	// TemplateDecl.Params at expansion time. Positional values containing a
	// ':' must be quoted — the named form wins on a bare `key: value` shape.
	Args []Arg
	// Comments are the comments attached to this statement.
	Comments []*Comment
}

// Arg is one argument of a use statement. Named args (name: "value") carry
// the key; positional args ("value") carry an empty Name.
type Arg struct {
	// Name is the argument key; empty for positional args.
	Name string
	// Value is the argument literal — any Literal form is accepted (D-13).
	Value Literal
}

// TemplateDecl is a `template name(p1, p2) { ... }` declaration (D-13). The
// body reuses *UnitNode (the plan's documented choice): the template body is
// exactly a unit body, so Fields, Edges, Subunits and UseStmts carry the
// body statements in source order; the Body node's own ID/Type/External and
// Name header slots stay empty. `${param}` tokens inside values are captured
// verbatim — substitution is the TemplateDef/Expand contract, never parse's.
type TemplateDecl struct {
	// Pos is the 1-indexed line the declaration starts on.
	Pos int
	// Name is the template identifier.
	Name string
	// Params holds the declared parameter names in order.
	Params []string
	// Body carries the body statements (full unit grammar, D-13).
	Body *UnitNode
	// Comments are the comments attached to this declaration.
	Comments []*Comment
}

// IncludeStmt is an `include path [once]` directive (D-14). The path text is
// captured as written (bareword or quoted); resolution relative to the
// including file and once-dedup happen in include.Resolve at pipeline time,
// never at parse time.
type IncludeStmt struct {
	// Pos is the 1-indexed line the statement starts on.
	Pos int
	// Path is the include path text as written.
	Path string
	// Once records the `once` modifier (at-most-once inclusion).
	Once bool
	// Comments are the comments attached to this statement.
	Comments []*Comment
}
