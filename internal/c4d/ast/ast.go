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
	// UseStmts holds `use` instantiations inside the block (D-16); the
	// grammar lands in a later plan, the slot exists so conversion and the
	// formatter can account for it.
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

// UseStmt is a placeholder for `use name(args)` instantiation statements
// (D-13, D-16); the grammar and semantics land in later plans.
type UseStmt struct {
	// Pos is the 1-indexed line the statement starts on.
	Pos int
	// Template is the template name.
	Template string
	// Comments are the comments attached to this statement.
	Comments []*Comment
}
