package c4d_test

import (
	"testing"

	"github.com/Djarvur/c4drill/internal/c4d/ast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Composition statements: template declarations, use instantiations (three
// positions), include directives, list forms (35-03 Task 1) and the
// reserved-keyword rule with suggestions (35-03 Task 2).

func TestParseTemplateDecl(t *testing.T) {
	t.Parallel()

	doc := parseDoc(t, "template service(name, tech) {\n\tsvc: system { technology: ${tech} }\n}\n")

	require.Len(t, doc.Templates, 1, "doc.Templates")
	tmpl := doc.Templates[0]
	assert.Equal(t, "service", tmpl.Name, "TemplateDecl.Name")
	assert.Equal(t, []string{"name", "tech"}, tmpl.Params, "TemplateDecl.Params")
	require.NotNil(t, tmpl.Body, "TemplateDecl.Body")

	// The body carries the full unit grammar; ${param} tokens are captured
	// VERBATIM — substitution is TemplateDef's contract, never parse's.
	require.Len(t, tmpl.Body.Subunits, 1, "Body.Subunits")
	assert.Equal(t, "svc", tmpl.Body.Subunits[0].ID, "Body.Subunits[0].ID")
	assert.Equal(t, "system", tmpl.Body.Subunits[0].Type, "Body.Subunits[0].Type")

	tech := fieldByKey(t, tmpl.Body.Subunits[0], "technology")
	assert.Equal(t, ast.KindBareword, tech.Value.Kind, "technology literal kind")
	assert.Equal(t, "${tech}", tech.Value.Str, "technology value must carry the raw token (no substitution at parse)")
}

func TestParseUseTopLevelPositional(t *testing.T) {
	t.Parallel()

	doc := parseDoc(t, `use service("api", "Go")` + "\n")

	require.Len(t, doc.UseStmts, 1, "doc.UseStmts")
	us := doc.UseStmts[0]
	assert.Equal(t, "service", us.Template, "UseStmt.Template")
	assert.Equal(t, []ast.Arg{
		{Name: "", Value: ast.Literal{Kind: ast.KindQuoted, Str: "api"}},
		{Name: "", Value: ast.Literal{Kind: ast.KindQuoted, Str: "Go"}},
	}, us.Args, "positional args recorded in order with empty Name")
}

func TestParseUseTopLevelNamed(t *testing.T) {
	t.Parallel()

	doc := parseDoc(t, `use service(name: "api", tech: "Go")` + "\n")

	require.Len(t, doc.UseStmts, 1, "doc.UseStmts")
	us := doc.UseStmts[0]
	assert.Equal(t, "service", us.Template, "UseStmt.Template")
	assert.Equal(t, []ast.Arg{
		{Name: "name", Value: ast.Literal{Kind: ast.KindQuoted, Str: "api"}},
		{Name: "tech", Value: ast.Literal{Kind: ast.KindQuoted, Str: "Go"}},
	}, us.Args, "named args recorded as key/value pairs in source order")
}

func TestParseUseInsideUnit(t *testing.T) {
	t.Parallel()

	doc := parseDoc(t, "platform: system {\n\tuse billing(name: \"b1\")\n}\n")

	require.Len(t, doc.Units, 1, "doc.Units")
	unit := doc.Units[0]
	require.Len(t, unit.UseStmts, 1, "UnitNode.UseStmts (D-16 authoring surface)")
	assert.Equal(t, "billing", unit.UseStmts[0].Template, "UseStmt.Template")
	require.Len(t, unit.UseStmts[0].Args, 1, "UseStmt.Args")
	assert.Equal(t, "name", unit.UseStmts[0].Args[0].Name, "Args[0].Name")
	assert.Equal(t, "b1", unit.UseStmts[0].Args[0].Value.Str, "Args[0].Value")
}

func TestParseUseInsideTemplateBody(t *testing.T) {
	t.Parallel()

	doc := parseDoc(t, "template outer(a) {\n\tuse helper(x: 1)\n\tapi: container { }\n}\n")

	require.Len(t, doc.Templates, 1, "doc.Templates")
	body := doc.Templates[0].Body
	require.NotNil(t, body, "TemplateDecl.Body")
	require.Len(t, body.UseStmts, 1, "Body.UseStmts (D-17 surface)")
	assert.Equal(t, "helper", body.UseStmts[0].Template, "UseStmt.Template")
	require.Len(t, body.UseStmts[0].Args, 1, "UseStmt.Args")
	assert.Equal(t, "x", body.UseStmts[0].Args[0].Name, "Args[0].Name")
	assert.Equal(t, "1", body.UseStmts[0].Args[0].Value.Str, "Args[0].Value")
	require.Len(t, body.Subunits, 1, "Body.Subunits still parses alongside use")
}

func TestParseInclude(t *testing.T) {
	t.Parallel()

	t.Run("bare path", func(t *testing.T) {
		t.Parallel()

		doc := parseDoc(t, "include ./shared.c4d\n")

		require.Len(t, doc.Includes, 1, "doc.Includes")
		assert.Equal(t, "./shared.c4d", doc.Includes[0].Path, "IncludeStmt.Path")
		assert.False(t, doc.Includes[0].Once, "IncludeStmt.Once")
	})

	t.Run("bare path with once", func(t *testing.T) {
		t.Parallel()

		doc := parseDoc(t, "include ./shared.c4d once\n")

		require.Len(t, doc.Includes, 1, "doc.Includes")
		assert.Equal(t, "./shared.c4d", doc.Includes[0].Path, "IncludeStmt.Path")
		assert.True(t, doc.Includes[0].Once, "IncludeStmt.Once (D-14 modifier)")
	})

	t.Run("quoted path", func(t *testing.T) {
		t.Parallel()

		doc := parseDoc(t, "include \"shared/common.c4d\"\n")

		require.Len(t, doc.Includes, 1, "doc.Includes")
		assert.Equal(t, "shared/common.c4d", doc.Includes[0].Path, "IncludeStmt.Path")
	})
}

func TestParseListFieldFormsUnit(t *testing.T) {
	t.Parallel()

	inline := firstUnit(t, "api: system { expanded: [c1, c2] }\n")
	perLine := firstUnit(t, "api: system {\n\texpanded: [\n\t\tc1\n\t\tc2\n\t]\n}\n")

	want := ast.Literal{Kind: ast.KindList, List: []string{"c1", "c2"}}
	assert.Equal(t, want, fieldByKey(t, inline, "expanded").Value, "inline list form (D-15)")
	assert.Equal(t, want, fieldByKey(t, perLine, "expanded").Value, "one-per-line list form (D-15)")
}

func TestParseTemplateBodyFullGrammar(t *testing.T) {
	t.Parallel()

	doc := parseDoc(t, `template dataService(name) {
	description: ${name} owns a private cache
	cache: componentDb { technology: Redis }
	-> cache publishes ${name} cache invalidations
	inner: containerBox {
		web: component { }
	}
}
`)

	require.Len(t, doc.Templates, 1, "doc.Templates")
	body := doc.Templates[0].Body
	require.NotNil(t, body, "TemplateDecl.Body")

	// Direct body fields parse (D-13 full grammar).
	desc := fieldByKey(t, body, "description")
	assert.Equal(t, "${name} owns a private cache", desc.Value.Str, "body field with embedded token")

	// Body edges with relative peers parse (D-13).
	require.Len(t, body.Edges, 1, "Body.Edges")
	assert.Equal(t, "cache", body.Edges[0].Peer, "Edges[0].Peer (relative peer)")
	assert.Equal(t, "publishes ${name} cache invalidations", body.Edges[0].Description, "Edges[0].Description")

	// Body nesting parses to any depth (D-13).
	require.Len(t, body.Subunits, 2, "Body.Subunits")
	assert.Equal(t, "cache", body.Subunits[0].ID, "Subunits[0].ID")
	require.Len(t, body.Subunits[1].Subunits, 1, "nested subunit inside template body")
	assert.Equal(t, "web", body.Subunits[1].Subunits[0].ID, "Subunits[1].Subunits[0].ID")
}
