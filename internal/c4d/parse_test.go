package c4d_test

import (
	"testing"

	"github.com/Djarvur/c4drill/internal/c4d/ast"
	"github.com/Djarvur/c4drill/internal/c4d/grammar"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// parseDoc runs the pigeon-generated grammar over src and returns the typed
// AST document. Task 2 tests exercise the grammar layer directly; Task 3
// re-routes the front-end through c4d.Parse.
func parseDoc(t *testing.T, src string) *ast.Document {
	t.Helper()

	result, err := grammar.Parse("", []byte(src), grammar.Memoize(true), grammar.MaxExpressions(1000000))
	require.NoError(t, err, "grammar.Parse() should not error")

	doc, ok := result.(*ast.Document)
	require.True(t, ok, "grammar.Parse() result should be *ast.Document, got %T", result)

	return doc
}

// parseErr asserts that src fails to parse.
func parseErr(t *testing.T, src string) {
	t.Helper()

	_, err := grammar.Parse("", []byte(src), grammar.Memoize(true), grammar.MaxExpressions(1000000))
	require.Error(t, err, "grammar.Parse() should error on invalid input %q", src)
}

// firstUnit returns the single top-level unit of a parsed document.
func firstUnit(t *testing.T, src string) *ast.UnitNode {
	t.Helper()

	doc := parseDoc(t, src)
	require.Len(t, doc.Units, 1, "doc.Units")

	return doc.Units[0]
}

// fieldByKey returns the field statement with the given key, failing the test
// when absent.
func fieldByKey(t *testing.T, unit *ast.UnitNode, key string) *ast.FieldStmt {
	t.Helper()

	for _, f := range unit.Fields {
		if f.Key == key {
			return f
		}
	}

	t.Fatalf("no field with key %q in unit %q", key, unit.ID)

	return nil
}

func TestParseUnitTypeAndName(t *testing.T) {
	t.Parallel()

	unit := firstUnit(t, `system "Payment Gateway" { description: processes cards }`)

	assert.Empty(t, unit.ID, "UnitNode.ID (omitted in type-led header)")
	assert.Equal(t, "system", unit.Type, "UnitNode.Type")
	assert.Equal(t, "Payment Gateway", unit.Name, "UnitNode.Name")
	assert.False(t, unit.External, "UnitNode.External")
	require.Len(t, unit.Fields, 1, "UnitNode.Fields")
	assert.Equal(t, "description", unit.Fields[0].Key, "Fields[0].Key")
	assert.Equal(t, ast.Literal{Kind: ast.KindBareword, Str: "processes cards"}, unit.Fields[0].Value, "Fields[0].Value")
}

func TestParseUnitOmittedType(t *testing.T) {
	t.Parallel()

	unit := firstUnit(t, "api { }")

	assert.Equal(t, "api", unit.ID, "UnitNode.ID")
	assert.Empty(t, unit.Type, "UnitNode.Type (omitted — inference happens in toModel, not here)")
	assert.Empty(t, unit.Subunits, "UnitNode.Subunits")
}

func TestParseUnitExternalModifier(t *testing.T) {
	t.Parallel()

	unit := firstUnit(t, `system external "Partner" { }`)

	assert.Equal(t, "system", unit.Type, "UnitNode.Type")
	assert.Equal(t, "Partner", unit.Name, "UnitNode.Name")
	assert.True(t, unit.External, "UnitNode.External (D-04 external modifier)")
}

func TestParseUnitIdLedHeader(t *testing.T) {
	t.Parallel()

	unit := firstUnit(t, `api: system "API Gateway" { technology: Go }`)

	assert.Equal(t, "api", unit.ID, "UnitNode.ID")
	assert.Equal(t, "system", unit.Type, "UnitNode.Type")
	assert.Equal(t, "API Gateway", unit.Name, "UnitNode.Name")
	assert.Equal(t, "Go", fieldByKey(t, unit, "technology").Value.Str, "technology value")
}

func TestParseNesting(t *testing.T) {
	t.Parallel()

	unit := firstUnit(t, `box "Platform" { api: system { } db: db { } }`)

	assert.Equal(t, "box", unit.Type, "UnitNode.Type")
	assert.Equal(t, "Platform", unit.Name, "UnitNode.Name")
	require.Len(t, unit.Subunits, 2, "UnitNode.Subunits (statement order, D-01)")
	assert.Equal(t, "api", unit.Subunits[0].ID, "Subunits[0].ID")
	assert.Equal(t, "system", unit.Subunits[0].Type, "Subunits[0].Type")
	assert.Equal(t, "db", unit.Subunits[1].ID, "Subunits[1].ID")
	assert.Equal(t, "db", unit.Subunits[1].Type, "Subunits[1].Type")
}

func TestParseArrows(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		glyph string
	}{
		{name: "forward", glyph: "->"},
		{name: "reverse", glyph: "<-"},
		{name: "bidirectional", glyph: "<->"},
		{name: "none", glyph: "--"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			unit := firstUnit(t, "a: system { "+tt.glyph+" peers }")

			require.Len(t, unit.Edges, 1, "UnitNode.Edges")
			assert.Equal(t, tt.glyph, unit.Edges[0].ArrowGlyph, "EdgeStmt.ArrowGlyph")
			assert.Equal(t, "peers", unit.Edges[0].Peer, "EdgeStmt.Peer")
		})
	}
}

func TestParseEdgeLabelShorthand(t *testing.T) {
	t.Parallel()

	// The D-09 triple: single un-piped value is the DESCRIPTION (user's
	// explicit override — 35-RESEARCH Risk 3), trailing pipe is tech-only,
	// pipe with both sides carries both.
	tests := []struct {
		name string
		src  string
		tech string
		desc string
	}{
		{name: "desc only", src: "a: system { -> db: queries orders }", tech: "", desc: "queries orders"},
		{name: "tech only pipe", src: "a: system { -> db: sql | }", tech: "sql", desc: ""},
		{name: "both", src: "a: system { -> db: sql | queries }", tech: "sql", desc: "queries"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			unit := firstUnit(t, tt.src)

			require.Len(t, unit.Edges, 1, "UnitNode.Edges")
			assert.Equal(t, tt.tech, unit.Edges[0].Technology, "EdgeStmt.Technology")
			assert.Equal(t, tt.desc, unit.Edges[0].Description, "EdgeStmt.Description")
		})
	}
}

func TestParseEdgeLabelQuotedWithPipe(t *testing.T) {
	t.Parallel()

	unit := firstUnit(t, `a: system { -> db: "sql | q" { color: red rank: equal } }`)

	require.Len(t, unit.Edges, 1, "UnitNode.Edges")
	assert.Equal(t, "sql", unit.Edges[0].Technology, "EdgeStmt.Technology (quoted label splits on pipe)")
	assert.Equal(t, "q", unit.Edges[0].Description, "EdgeStmt.Description (quoted label splits on pipe)")
	require.Len(t, unit.Edges[0].Options, 2, "EdgeStmt.Options (trailing brace block, D-09)")
	assert.Equal(t, "color", unit.Edges[0].Options[0].Key, "Options[0].Key")
	assert.Equal(t, "red", unit.Edges[0].Options[0].Value.Str, "Options[0].Value")
	assert.Equal(t, "rank", unit.Edges[0].Options[1].Key, "Options[1].Key")
	assert.Equal(t, "equal", unit.Edges[0].Options[1].Value.Str, "Options[1].Value")
}

func TestParseLiteralsBareword(t *testing.T) {
	t.Parallel()

	unit := firstUnit(t, "a: system { description: simple value here }")

	assert.Equal(t,
		ast.Literal{Kind: ast.KindBareword, Str: "simple value here"},
		fieldByKey(t, unit, "description").Value,
		"bareword value (edge whitespace trimmed)")
}

func TestParseLiteralsURLBareword(t *testing.T) {
	t.Parallel()

	unit := firstUnit(t, "a: system { reference: https://example.com/docs }")

	assert.Equal(t,
		ast.Literal{Kind: ast.KindBareword, Str: "https://example.com/docs"},
		fieldByKey(t, unit, "reference").Value,
		"scheme-prefixed URL as bareword (D-06)")
}

func TestParseLiteralsDoubleQuoted(t *testing.T) {
	t.Parallel()

	unit := firstUnit(t, `a: system { description: "has { } : |  and spaces " }`)

	assert.Equal(t,
		ast.Literal{Kind: ast.KindQuoted, Str: "has { } : |  and spaces "},
		fieldByKey(t, unit, "description").Value,
		"double-quoted value (structural chars + edge whitespace preserved, D-06)")
}

func TestParseLiteralsDoubleQuotedEscapes(t *testing.T) {
	t.Parallel()

	unit := firstUnit(t, `a: system { name: "say \"hi\"" }`)

	assert.Equal(t,
		ast.Literal{Kind: ast.KindQuoted, Str: `say "hi"`},
		fieldByKey(t, unit, "name").Value,
		"escaped quote inside double-quoted value")
}

func TestParseLiteralsTripleQuoted(t *testing.T) {
	t.Parallel()

	unit := firstUnit(t, "a: system { description: \"\"\"\nline one\nline two\n\"\"\" }")

	assert.Equal(t,
		ast.Literal{Kind: ast.KindTriple, Str: "\nline one\nline two\n"},
		fieldByKey(t, unit, "description").Value,
		"triple-quoted multi-line value (D-06)")
}

func TestParseCommentLines(t *testing.T) {
	t.Parallel()

	// Leading comments attach to the following statement; a same-line
	// trailing comment attaches to the preceding statement (gofmt
	// semantics); an orphan comment attaches to the enclosing document.
	doc := parseDoc(t, "# top comment\na: system { } # trailing comment\n# orphan\n")

	require.Len(t, doc.Units, 1, "doc.Units")
	require.Len(t, doc.Units[0].Comments, 2, "unit.Comments (lead + same-line tail)")
	assert.Equal(t, "top comment", doc.Units[0].Comments[0].Text, "lead Comment.Text")
	assert.Equal(t, 1, doc.Units[0].Comments[0].Pos, "lead Comment.Pos")
	assert.Equal(t, "trailing comment", doc.Units[0].Comments[1].Text, "tail Comment.Text")
	assert.Equal(t, 2, doc.Units[0].Comments[1].Pos, "tail Comment.Pos")
	require.Len(t, doc.TrailingComments, 1, "doc.TrailingComments (orphan)")
	assert.Equal(t, "orphan", doc.TrailingComments[0].Text, "orphan Comment.Text")
}

func TestParseCommentAttachmentInsideBlock(t *testing.T) {
	t.Parallel()

	unit := firstUnit(t, "a: system {\n\t# describes a\n\tdescription: hello\n}")

	f := fieldByKey(t, unit, "description")
	require.Len(t, f.Comments, 1, "FieldStmt.Comments (attached to following statement)")
	assert.Equal(t, "describes a", f.Comments[0].Text, "Comment.Text")
	assert.Equal(t, 2, f.Comments[0].Pos, "Comment.Pos (comment's own line)")
	assert.Equal(t, 3, f.Pos, "FieldStmt.Pos (statement's own line)")
}

func TestParseIdentifiers(t *testing.T) {
	t.Parallel()

	unit := firstUnit(t, "my-unit_2: system { -> billing.svc }")

	assert.Equal(t, "my-unit_2", unit.ID, "UnitNode.ID ([A-Za-z0-9_-]+, D-07)")
	require.Len(t, unit.Edges, 1, "UnitNode.Edges")
	assert.Equal(t, "billing.svc", unit.Edges[0].Peer, "EdgeStmt.Peer (dotted path, D-07/D-10)")
}

func TestParseProperties(t *testing.T) {
	t.Parallel()

	doc := parseDoc(t, "properties {\n\tname: Demo\n\tdescription: Demo diagram\n\tcolor: red\n\tlineLength: 42\n}")

	require.NotNil(t, doc.Properties, "Document.Properties (D-12)")
	require.Len(t, doc.Properties.Fields, 4, "PropertiesBlock.Fields")
	assert.Equal(t, "name", doc.Properties.Fields[0].Key, "Fields[0].Key")
	assert.Equal(t, "Demo", doc.Properties.Fields[0].Value.Str, "Fields[0].Value")
	assert.Equal(t, "lineLength", doc.Properties.Fields[3].Key, "Fields[3].Key")
	assert.Equal(t, "42", doc.Properties.Fields[3].Value.Str, "Fields[3].Value")
}

func TestParsePropertiesListValues(t *testing.T) {
	t.Parallel()

	// D-15: inline and one-per-line list forms both accepted.
	tests := []struct {
		name string
		src  string
	}{
		{name: "inline", src: "properties { expanded: [a, b] }"},
		{name: "one per line", src: "properties {\n\texpanded: [\n\t\ta\n\t\tb\n\t]\n}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			doc := parseDoc(t, tt.src)

			require.NotNil(t, doc.Properties, "Document.Properties")
			require.Len(t, doc.Properties.Fields, 1, "PropertiesBlock.Fields")
			assert.Equal(t,
				ast.Literal{Kind: ast.KindList, List: []string{"a", "b"}},
				doc.Properties.Fields[0].Value,
				"expanded list value")
		})
	}
}

func TestParseSemicolonSeparator(t *testing.T) {
	t.Parallel()

	unit := firstUnit(t, "api: container { technology: Go; db: db { } }")

	assert.Equal(t, "Go", fieldByKey(t, unit, "technology").Value.Str, "technology value (before ; D-18)")
	require.Len(t, unit.Subunits, 1, "Subunits (after ; D-18)")
	assert.Equal(t, "db", unit.Subunits[0].ID, "Subunits[0].ID")
}

func TestParseEmptyBlock(t *testing.T) {
	t.Parallel()

	unit := firstUnit(t, "x: system { }")

	assert.Equal(t, "x", unit.ID, "UnitNode.ID")
	assert.Empty(t, unit.Fields, "Fields")
	assert.Empty(t, unit.Edges, "Edges")
	assert.Empty(t, unit.Subunits, "Subunits")
}

func TestParseEmptyDocument(t *testing.T) {
	t.Parallel()

	doc := parseDoc(t, "")

	assert.Empty(t, doc.Units, "Units")
	assert.Nil(t, doc.Properties, "Properties")
}

func TestParseReservedKeywordsNotUnitIds(t *testing.T) {
	t.Parallel()

	// D-19 groundwork: statement keywords cannot parse as unit ids at the
	// grammar level (full Levenshtein-suggestion errors land in Plan 03).
	for _, src := range []string{
		"use svc(api) { }",
		"include ./shared.c4d",
		"template base(p) { }",
	} {
		parseErr(t, src)
	}
}

func TestParseSyntaxErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		src  string
	}{
		{name: "unclosed brace", src: "a: system {"},
		{name: "missing separator", src: "a: system b: db { }"},
		{name: "stray arrow glyph", src: "a: system { => db }"},
		{name: "unknown field key", src: "a: system { bogus: value }"},
		{name: "unknown edge option", src: "a: system { -> db { bogus: x } }"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			parseErr(t, tt.src)
		})
	}
}
