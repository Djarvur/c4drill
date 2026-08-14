package c4d_test

import (
	"os"
	"strings"
	"testing"

	"github.com/Djarvur/c4drill/internal/c4d"
	"github.com/Djarvur/c4drill/internal/c4d/ast"
	"github.com/Djarvur/c4drill/internal/model"
	"github.com/Djarvur/c4drill/internal/parser"
	"github.com/Djarvur/c4drill/internal/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Emitters (35-04): Model->TOML in the fixed canonical field order (D-23),
// Model->AST->C4D in the compact-leaf style (D-33). Black-box tests build
// *parser.Model values by hand or parse TOML fixtures; the C4D fixpoint runs
// purely at the AST level (the same-wave 35-05 API state never leaks in).

// assertOrdered asserts each needle appears in out strictly after the
// previous one (canonical order pinning, D-23).
func assertOrdered(t *testing.T, out string, needles ...string) {
	t.Helper()

	last := -1

	for _, needle := range needles {
		idx := strings.Index(out, needle)
		require.NotEqual(t, -1, idx, "emitted output should contain %q:\n%s", needle, out)
		assert.Greater(t, idx, last, "%q must follow the previous canonical entry (got index %d, last %d):\n%s",
			needle, idx, last, out)
		last = idx
	}
}

// sectionOf returns the body of the TOML section starting at header, up to
// the next table header (TOML sections end where the next '[' line begins).
func sectionOf(t *testing.T, out, header string) string {
	t.Helper()

	start := strings.Index(out, header)
	require.NotEqual(t, -1, start, "emitted output should contain %q:\n%s", header, out)

	rest := out[start+len(header):]
	if end := strings.Index(rest, "\n["); end != -1 {
		return rest[:end]
	}

	return rest
}

func TestEmitTOMLUnitOrderFollowsUnitOrder(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		UnitOrder: []string{"second", "first"},
		Units: map[string]*model.Unit{
			"first":  {Type: model.TypeSystem, Name: "First"},
			"second": {Type: model.TypeSystem, Name: "Second"},
		},
	}

	out, err := c4d.EmitTOML(m)
	require.NoError(t, err, "EmitTOML() should not error")

	assertOrdered(t, out, "[second]", "[first]")
}

func TestEmitTOMLSubunitOrderFollowsSubunitOrder(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		UnitOrder: []string{"platform"},
		Units: map[string]*model.Unit{
			"platform": {
				Type:         model.TypeSystem,
				Name:         "Platform",
				SubunitOrder: []string{"zeta", "alpha"},
				Subunits: map[string]*model.Unit{
					"zeta":  {Type: model.TypeContainer, Name: "Zeta"},
					"alpha": {Type: model.TypeContainer, Name: "Alpha"},
				},
			},
		},
	}

	out, err := c4d.EmitTOML(m)
	require.NoError(t, err, "EmitTOML() should not error")

	assertOrdered(t, out, "[platform]", "[platform.zeta]", "[platform.alpha]")
}

func TestEmitTOMLCanonicalUnitFieldOrder(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		UnitOrder: []string{"webapp"},
		Units: map[string]*model.Unit{
			"webapp": {
				Type:        model.TypeSystem,
				Name:        "Web App",
				Description: "Serves",
				Technology:  "Go",
				Reference:   "https://example.com",
				Color:       "blue",
				Style:       "dashed",
				Border:      "black",
				Edges:       "spline",
				Width:       300,
				Height:      200,
				Expanded:    []string{"api"},
			},
		},
	}

	out, err := c4d.EmitTOML(m)
	require.NoError(t, err, "EmitTOML() should not error")

	// D-23 fixed order: type, name, description, technology, reference,
	// color, style, border, edges, width, height, expanded.
	assertOrdered(t, out,
		"[webapp]",
		"type = ",
		"name = ",
		"description = ",
		"technology = ",
		"reference = ",
		"color = ",
		"style = ",
		"border = ",
		"edges = ",
		"width = ",
		"height = ",
		"expanded = ",
	)
	assert.Contains(t, out, `width = 300`, "width emits as an integer spelling")
	assert.Contains(t, out, `height = 200`, "height emits as an integer spelling")
	assert.Contains(t, out, `expanded = ["api"]`, "expanded emits as an array")
}

func TestEmitTOMLLinkFieldOrderAndTables(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		UnitOrder: []string{"api"},
		Units: map[string]*model.Unit{
			"api": {
				Type: model.TypeSystem,
				Name: "API",
				Links: []model.Link{{
					Peer:          "db",
					Arrow:         model.ArrowReverse,
					Rank:          model.RankReverse,
					Color:         "green",
					Style:         "dashed",
					Technology:    "SQL",
					Description:   "Queries",
					LabelPosition: model.LabelHead,
					Length:        2,
				}},
				LinksFrom: []model.Link{{
					Peer:        "webapp",
					Technology:  "HTTPS",
					Description: "Calls",
				}},
			},
		},
	}

	out, err := c4d.EmitTOML(m)
	require.NoError(t, err, "EmitTOML() should not error")

	require.Contains(t, out, "[[api.link]]", "links emit as [[api.link]] array tables")
	require.Contains(t, out, "[[api.linkFrom]]", "LinksFrom emit as [[api.linkFrom]] array tables")

	link := sectionOf(t, out, "[[api.link]]")
	for _, needle := range []string{
		`peer = "db"`,
		`arrow = "reverse"`,
		`rank = "reverse"`,
		`color = "green"`,
		`style = "dashed"`,
		`technology = "SQL"`,
		`description = "Queries"`,
		`labelPosition = "head"`,
		`length = 2`,
	} {
		require.Contains(t, link, needle, "[[api.link]] body")
	}

	// D-23 link order: peer, arrow, rank, color, style, technology,
	// description, labelPosition, length.
	assertOrdered(t, link,
		"peer = ",
		"arrow = ",
		"rank = ",
		"color = ",
		"style = ",
		"technology = ",
		"description = ",
		"labelPosition = ",
		"length = ",
	)

	from := sectionOf(t, out, "[[api.linkFrom]]")
	assert.Contains(t, from, `peer = "webapp"`, "[[api.linkFrom]] body")
	assert.Contains(t, from, `technology = "HTTPS"`, "[[api.linkFrom]] body")
}

func TestEmitTOMLTemplatesUsesIncludes(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "T"},
		UnitOrder:  []string{"platform"},
		Units: map[string]*model.Unit{
			"platform": {Type: model.TypeSystem, Name: "Platform"},
		},
		Templates: map[string]*parser.TemplateDef{
			"microservice": {
				Params: []string{"name", "tech"},
				Unit: &model.Unit{
					Type:        model.TypeContainer,
					Name:        "${name} Service",
					Technology:  "${tech}",
					Description: "${name} handles its domain",
				},
			},
		},
		Instantiations: []parser.Instantiation{
			{Template: "microservice", Parent: "platform", Params: map[string]string{"tech": "Go", "name": "auth"}},
			{Template: "microservice", Params: map[string]string{"name": "edge"}},
		},
		Includes: []parser.IncludeDirective{
			{Path: "auth.toml"},
			{Path: "billing.toml", Once: true},
		},
	}

	out, err := c4d.EmitTOML(m)
	require.NoError(t, err, "EmitTOML() should not error")

	// [template.<name>] with params = [...] and the unit body fields.
	require.Contains(t, out, "[template.microservice]\n", "[template.microservice] table")
	assert.Contains(t, out, `params = ["name", "tech"]`, "declared params in order")
	assert.Contains(t, out, `name = "${name} Service"`, "template unit name")
	assert.Contains(t, out, `technology = "${tech}"`, "template unit technology")

	// [[use]] with template/parent/params — Parent omitted when empty.
	assert.Contains(t, out,
		"[[use]]\ntemplate = \"microservice\"\nparent = \"platform\"\nname = \"auth\"\ntech = \"Go\"\n",
		"[[use]] with parent")
	assert.Contains(t, out,
		"[[use]]\ntemplate = \"microservice\"\nname = \"edge\"\n",
		"[[use]] without parent omits the parent key")

	// [[include]] with path and once = true only when set.
	assert.Contains(t, out, "[[include]]\npath = \"auth.toml\"\n", "[[include]] without once")
	assert.Contains(t, out, "[[include]]\npath = \"billing.toml\"\nonce = true\n", "[[include]] with once")
}

func TestEmitTOMLTemplateSubunitsLinksAndBodyUses(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Templates: map[string]*parser.TemplateDef{
			"dataService": {
				Params: []string{"name"},
				Unit: &model.Unit{
					Type:         model.TypeContainer,
					Name:         "${name} Data",
					SubunitOrder: []string{"cache"},
					Subunits: map[string]*model.Unit{
						"cache": {Type: model.TypeComponentDb, Name: "Cache"},
					},
					Links: []model.Link{{Peer: "bus", Description: "Publishes"}},
				},
				Instantiations: []parser.Instantiation{
					{Template: "helper", Params: map[string]string{"x": "1"}},
					{Template: "helper", Parent: "cache", Params: map[string]string{"x": "2"}},
				},
			},
		},
	}

	out, err := c4d.EmitTOML(m)
	require.NoError(t, err, "EmitTOML() should not error")

	// Template root table, its link table, then subunit tables.
	assertOrdered(t, out,
		"[template.dataService]",
		"[[template.dataService.link]]",
		"[template.dataService.cache]",
	)

	// Template-body uses: root-relative parent "" -> [[template.<name>.use]];
	// parent under a subunit path -> [[template.<name>.<path>.use]]. The
	// root-level use MUST precede the subunit table (TOML table context).
	assertOrdered(t, out,
		"[[template.dataService.use]]",
		"[template.dataService.cache]",
		"[[template.dataService.cache.use]]",
	)
	assert.Contains(t, out, "template = \"helper\"\nx = \"1\"\n", "template-body use entry params")
}

func TestEmitTOMLStringQuoting(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		UnitOrder: []string{"webapp"},
		Units: map[string]*model.Unit{
			"webapp": {
				Type:        model.TypeSystem,
				Name:        `Ünïcode "quoted" \ back`,
				Description: "plain",
			},
		},
	}

	out, err := c4d.EmitTOML(m)
	require.NoError(t, err, "EmitTOML() should not error")

	assert.Contains(t, out, `name = "Ünïcode \"quoted\" \\ back"`,
		"quotes and backslashes escape; non-ASCII passes through")
}

func TestEmitTOMLOmitsEmptyFields(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		UnitOrder: []string{"bare"},
		Units: map[string]*model.Unit{
			"bare": {Type: model.TypeDb},
		},
	}

	out, err := c4d.EmitTOML(m)
	require.NoError(t, err, "EmitTOML() should not error")

	assert.Contains(t, out, "[bare]\ntype = \"db\"\n", "only non-empty fields emit")
	assert.NotContains(t, out, "name = \"\"", "empty name never emits")
	assert.NotContains(t, out, "description = ", "empty description never emits")
	assert.NotContains(t, out, "expanded = ", "empty expanded never emits")
	assert.NotContains(t, out, "[[bare.link]]", "no link table for a unit without links")
}

func TestEmitTOMLNewlineValueSingleLine(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		UnitOrder: []string{"webapp"},
		Units: map[string]*model.Unit{
			"webapp": {
				Type:        model.TypeSystem,
				Name:        "Web",
				Description: "line1\nline2",
			},
		},
	}

	out, err := c4d.EmitTOML(m)
	require.NoError(t, err, "EmitTOML() should not error")

	descLine := ""

	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "description = ") {
			descLine = line
		}
	}

	require.NotEmpty(t, descLine, "emitted output should carry a description line:\n%s", out)

	// D-06 pinned rule: newline-containing values emit as a single-line basic
	// string with the escaped \n (two characters) — never a raw line break
	// inside the quoted string (that would be invalid TOML).
	assert.Contains(t, descLine, `\n`, "escaped newline (two characters)")
	assert.Equal(t, "description = \"line1\\nline2\"", descLine, "single-line escaped form")
	assert.Len(t, strings.Split(descLine, "\n"), 1, "the description value stays on one line")

	// Round-trip prerequisite: the emitted TOML re-parses to the identical value.
	m2, err := parser.Parse([]byte(out))
	require.NoError(t, err, "emitted TOML must re-parse")
	assert.Equal(t, "line1\nline2", m2.Units["webapp"].Description, "re-parsed newline value")
}

func TestEmitTOMLDeterministic(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("../../testdata/valid.toml")
	require.NoError(t, err, "failed to read test fixture")

	m, err := parser.Parse(data)
	require.NoError(t, err, "parser.Parse(fixture)")

	first, err := c4d.EmitTOML(m)
	require.NoError(t, err, "EmitTOML() first call")
	second, err := c4d.EmitTOML(m)
	require.NoError(t, err, "EmitTOML() second call")

	assert.Exactly(t, first, second, "two EmitTOML calls must be byte-identical")
}

func TestEmitTOMLValidFixture(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("../../testdata/valid.toml")
	require.NoError(t, err, "failed to read test fixture")

	m, err := parser.Parse(data)
	require.NoError(t, err, "parser.Parse(fixture)")

	out, err := c4d.EmitTOML(m)
	require.NoError(t, err, "EmitTOML() should not error")

	// [properties] first, units in UnitOrder (fixture section order).
	assertOrdered(t, out, "[properties]", "[user]", "[webapp]")

	// The emitted TOML re-parses without error and reproduces the model.
	m2, err := parser.Parse([]byte(out))
	require.NoError(t, err, "emitted TOML must re-parse (round-trip prerequisite)")

	assert.Equal(t, m.Properties, m2.Properties, "Properties round-trip")
	assert.Equal(t, m.UnitOrder, m2.UnitOrder, "UnitOrder round-trip")

	for _, name := range m.UnitOrder {
		assert.Equal(t, m.Units[name], m2.Units[name], "unit %q round-trips", name)
	}
}

func TestEmitTOMLNilModelError(t *testing.T) {
	t.Parallel()

	_, err := c4d.EmitTOML(nil)
	require.Error(t, err, "EmitTOML(nil) must error")
}

// --- FromModel: Model -> AST (35-04 Task 2) ---

func TestFromModelMapsDocument(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Demo"},
		UnitOrder:  []string{"webapp", "user"},
		Units: map[string]*model.Unit{
			"webapp": {
				Type:        model.TypeSystem,
				Name:        "Web App",
				Description: "Serves",
				Links:       []model.Link{{Peer: "db", Technology: "SQL", Description: "Queries"}},
			},
			"user": {Type: model.TypePerson, Name: "User"},
		},
		Templates: map[string]*parser.TemplateDef{
			"svc": {Params: []string{"name"}, Unit: &model.Unit{Type: model.TypeSystem, Name: "${name} Service"}},
		},
		Instantiations: []parser.Instantiation{
			{Template: "svc", Params: map[string]string{"name": "auth"}},
		},
		Includes: []parser.IncludeDirective{{Path: "./shared.c4d", Once: true}},
	}

	doc := c4d.FromModel(m)

	// Units in UnitOrder with id/type/name from the Model.
	require.Len(t, doc.Units, 2, "doc.Units in UnitOrder")
	assert.Equal(t, "webapp", doc.Units[0].ID, "Units[0].ID")
	assert.Equal(t, "system", doc.Units[0].Type, "Units[0].Type")
	assert.Equal(t, "Web App", doc.Units[0].Name, "Units[0].Name (header slot)")
	assert.Equal(t, "user", doc.Units[1].ID, "Units[1].ID")

	// Links map to edge statements.
	require.Len(t, doc.Units[0].Edges, 1, "Edges from Links")
	assert.Equal(t, "db", doc.Units[0].Edges[0].Peer, "Edges[0].Peer")
	assert.Equal(t, "SQL", doc.Units[0].Edges[0].Technology, "Edges[0].Technology")
	assert.Equal(t, "Queries", doc.Units[0].Edges[0].Description, "Edges[0].Description")

	// Templates, uses, includes map 1:1.
	require.Len(t, doc.Templates, 1, "doc.Templates")
	assert.Equal(t, "svc", doc.Templates[0].Name, "Templates[0].Name")
	assert.Equal(t, []string{"name"}, doc.Templates[0].Params, "Templates[0].Params")
	require.Len(t, doc.UseStmts, 1, "doc.UseStmts")
	assert.Equal(t, "svc", doc.UseStmts[0].Template, "UseStmts[0].Template")
	require.Len(t, doc.Includes, 1, "doc.Includes")
	assert.Equal(t, "./shared.c4d", doc.Includes[0].Path, "Includes[0].Path")
	assert.True(t, doc.Includes[0].Once, "Includes[0].Once")

	// Properties map to the properties block.
	require.NotNil(t, doc.Properties, "doc.Properties")
	require.Len(t, doc.Properties.Fields, 1, "Properties.Fields")
	assert.Equal(t, "name", doc.Properties.Fields[0].Key, "Properties.Fields[0].Key")
	assert.Equal(t, "Demo", doc.Properties.Fields[0].Value.Str, "Properties.Fields[0].Value")
}

func TestFromModelArrowGlyphs(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		arrow    model.ArrowDirection
		glyph    string
		arrowOpt string
	}{
		{"omitted default", "", "->", ""},
		{"forward states the default explicitly", model.ArrowForward, "->", "forward"},
		{"reverse rides the arrow option", model.ArrowReverse, "->", "reverse"},
		{"bidirectional", model.ArrowBidirectional, "<->", ""},
		{"none", model.ArrowNone, "--", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			m := &parser.Model{
				UnitOrder: []string{"api"},
				Units: map[string]*model.Unit{
					"api": {Type: model.TypeSystem, Name: "API", Links: []model.Link{{Peer: "db", Arrow: tc.arrow}}},
				},
			}

			doc := c4d.FromModel(m)

			require.Len(t, doc.Units[0].Edges, 1, "one edge per link")
			assert.Equal(t, tc.glyph, doc.Units[0].Edges[0].ArrowGlyph, "arrow glyph inverse mapping")

			if tc.arrowOpt == "" {
				assertNoEdgeOption(t, doc.Units[0].Edges[0], "arrow")
			} else {
				require.Len(t, doc.Units[0].Edges[0].Options, 1, "one option")
				assert.Equal(t, "arrow", doc.Units[0].Edges[0].Options[0].Key, "arrow option key")
				assert.Equal(t, tc.arrowOpt, doc.Units[0].Edges[0].Options[0].Value.Str, "arrow option value")
			}
		})
	}

	t.Run("linksFrom emit as incoming statements", func(t *testing.T) {
		t.Parallel()

		m := &parser.Model{
			UnitOrder: []string{"db"},
			Units: map[string]*model.Unit{
				"db": {Type: model.TypeDb, Name: "DB", LinksFrom: []model.Link{{Peer: "api", Description: "Queries"}}},
			},
		}

		doc := c4d.FromModel(m)

		require.Len(t, doc.Units[0].Edges, 1, "one edge per linkFrom entry")
		assert.Equal(t, "<-", doc.Units[0].Edges[0].ArrowGlyph, "LinksFrom statement form")
		assert.Equal(t, "api", doc.Units[0].Edges[0].Peer, "LinksFrom peer")
		assertNoEdgeOption(t, doc.Units[0].Edges[0], "arrow")

		m.Units["db"].LinksFrom[0].Arrow = model.ArrowReverse

		doc = c4d.FromModel(m)

		require.Len(t, doc.Units[0].Edges, 1, "one edge per explicit-arrow linkFrom")
		assert.Equal(t, "<-", doc.Units[0].Edges[0].ArrowGlyph, "LinksFrom statement form")
		require.Len(t, doc.Units[0].Edges[0].Options, 1, "arrow option present")
		assert.Equal(t, "reverse", doc.Units[0].Edges[0].Options[0].Value.Str, "arrow option value")
	})
}

// assertNoEdgeOption asserts the edge carries no option with the given key.
func assertNoEdgeOption(t *testing.T, edge *ast.EdgeStmt, key string) {
	t.Helper()

	for _, opt := range edge.Options {
		assert.NotEqual(t, key, opt.Key, "edge must not carry a %s option", key)
	}
}

func TestFromModelInstantiationPlacement(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		UnitOrder: []string{"platform"},
		Units: map[string]*model.Unit{
			"platform": {Type: model.TypeSystem, Name: "Platform"},
		},
		Instantiations: []parser.Instantiation{
			{Template: "svc", Parent: "platform", Params: map[string]string{"name": "auth"}},
			{Template: "svc", Params: map[string]string{"name": "edge"}},
		},
	}

	doc := c4d.FromModel(m)

	// D-16: a use with a Parent lands INSIDE that unit's block; a parentless
	// use stays at top level.
	require.Len(t, doc.Units, 1, "doc.Units")
	require.Len(t, doc.Units[0].UseStmts, 1, "parented use attaches inside the unit block")
	assert.Equal(t, "svc", doc.Units[0].UseStmts[0].Template, "nested use template")
	require.Len(t, doc.UseStmts, 1, "parentless use stays top level")
	assert.Equal(t, "svc", doc.UseStmts[0].Template, "top-level use template")

	out := c4d.EmitC4D(doc)
	assert.Contains(t, out, "platform: system \"Platform\" {\n  use svc(name: auth)\n}\n",
		"nested use emits inside the block")
	assert.Contains(t, out, "\nuse svc(name: edge)\n", "top-level use emits after the units")
}

func TestFromModelExternalType(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		UnitOrder: []string{"pay"},
		Units: map[string]*model.Unit{
			"pay": {Type: model.TypeSystemExternal, Name: "Pay"},
		},
	}

	doc := c4d.FromModel(m)

	require.Len(t, doc.Units, 1, "doc.Units")
	assert.Equal(t, "system", doc.Units[0].Type, "external variants split to the base keyword")
	assert.True(t, doc.Units[0].External, "External records the modifier (D-04)")

	out := c4d.EmitC4D(doc)
	assert.Equal(t, "pay: system external \"Pay\" { }\n", out, "external modifier in the header")
}

func TestFromModelNewlineValueTripleQuoted(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		UnitOrder: []string{"webapp"},
		Units: map[string]*model.Unit{
			"webapp": {Type: model.TypeSystem, Name: "W", Description: "line1\nline2"},
		},
	}

	out := c4d.EmitC4D(c4d.FromModel(m))

	// D-06 inverse: newline-containing values emit as a triple-quoted block.
	assert.Contains(t, out, `"""`, "triple-quoted block")
	assert.Contains(t, out, "line1\nline2", "raw newline inside the triple block")

	// The single-line leaf rule requires single-line-able values — a unit
	// carrying a triple value NEVER emits as a one-line leaf.
	assert.Contains(t, out, "webapp: system \"W\" {\n", "multi-line header (not a one-line leaf)")
}

func TestFromModelValidFixtureToC4D(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("../../testdata/valid.toml")
	require.NoError(t, err, "failed to read test fixture")

	m, err := parser.Parse(data)
	require.NoError(t, err, "parser.Parse(fixture)")

	out := c4d.EmitC4D(c4d.FromModel(m))

	// Properties block first with the D-23 property field order.
	assertOrdered(t, out, "properties {", "name: Test System", "description: A test architecture",
		"color: transparent", "edges: straight", "lineLength: 40", "expanded: [webapp]")

	// Units in UnitOrder.
	assertOrdered(t, out, "properties {", "user: person", "webapp: system")

	// Compact-leaf one-liners with the D-23 body field order (description
	// before technology) and safe-value quoting ("Go, React" carries a comma).
	assert.Contains(t, out, "user: person \"User\" { description: End user of the system }\n")
	assert.Contains(t, out,
		"webapp: system \"Web Application\" { description: Main web application; technology: \"Go, React\" }\n")
}

func TestEmittersSkipMirrorLinks(t *testing.T) {
	t.Parallel()

	// The validator synthesizes mirror LinksFrom entries in place — the only
	// way to observe Mirror outside package model. Emitters must skip them.
	m := &parser.Model{
		UnitOrder: []string{"user", "webapp"},
		Units: map[string]*model.Unit{
			"user":   {Type: model.TypePerson, Name: "User", Links: []model.Link{{Peer: "webapp", Description: "Uses"}}},
			"webapp": {Type: model.TypeSystem, Name: "Web"},
		},
	}
	require.Empty(t, validator.Validate(m), "fixture model must validate")
	require.Len(t, m.Units["webapp"].LinksFrom, 1, "validator synthesized a LinksFrom entry")
	require.True(t, m.Units["webapp"].LinksFrom[0].IsMirror(), "the synthesized entry is a mirror")

	out, err := c4d.EmitTOML(m)
	require.NoError(t, err, "EmitTOML() should not error")
	assert.NotContains(t, out, "[[webapp.linkFrom]]", "mirror links never emit as TOML")

	doc := c4d.FromModel(m)

	for _, unit := range doc.Units {
		if unit.ID != "webapp" {
			continue
		}

		assert.Empty(t, unit.Edges, "mirror links never map to C4D edges (webapp carries only the synthesized LinksFrom)")
	}
}

// --- EmitC4D: AST -> text, compact-leaf style (D-33) ---

func TestEmitC4DCompactLeaf(t *testing.T) {
	t.Parallel()

	// A multi-line-authored leaf normalizes to the one-line form (D-33).
	doc := parseDoc(t, "db: db {\n\tdescription: cache\n}\n")

	out := c4d.EmitC4D(doc)

	assert.Equal(t, "db: db { description: cache }\n", out, "one-line leaf block")

	// No newline between the braces of a leaf.
	line := strings.TrimSuffix(strings.Split(out, "\n")[0], "")
	assert.Contains(t, line, "{ description: cache }", "leaf fields stay on the header line")
}

func TestEmitC4DNestedMultiLineIndent(t *testing.T) {
	t.Parallel()

	doc := parseDoc(t, `platform: system "Platform" {
	description: hosts services
	auth: container {
		technology: Go
		-> db: calls
	}
}`+"\n")

	out := c4d.EmitC4D(doc)

	lines := strings.Split(strings.TrimSuffix(out, "\n"), "\n")
	require.Len(t, lines, 7, "multi-line block shape:\n%s", out)
	assert.Equal(t, `platform: system "Platform" {`, lines[0], "header at depth 0")
	assert.Equal(t, "  description: hosts services", lines[1], "2 spaces per depth (D-33)")
	assert.Equal(t, "  auth: container {", lines[2], "subunit header at depth 1")
	assert.Equal(t, "    technology: Go", lines[3], "nested fields at depth 2")
	assert.Equal(t, "    -> db: calls", lines[4], "nested edge at depth 2")
	assert.Equal(t, "  }", lines[5], "nested closing brace at depth 1")
	assert.Equal(t, "}", lines[6], "closing brace at the header's depth")
}

func TestEmitC4DEdgeForms(t *testing.T) {
	t.Parallel()

	doc := parseDoc(t, `api: system {
	-> db: "SQL | queries orders"
	-> cache: "Redis |"
	-> queue: publishes events
	-> analytics: "HTTP POST | sends" { color: gray }
}`+"\n")

	out := c4d.EmitC4D(doc)

	// Both parts, tech-only trailing pipe, desc-only without pipe, and the
	// trailing option block (D-09 inverse).
	assert.Contains(t, out, "\n  -> db: \"SQL | queries orders\"\n", "tech | desc")
	assert.Contains(t, out, "\n  -> cache: \"Redis |\"\n", "tech-only trailing pipe")
	assert.Contains(t, out, "\n  -> queue: publishes events\n", "description-only omits the pipe")
	assert.Contains(t, out, "\n  -> analytics: \"HTTP POST | sends\" { color: gray }\n", "option block")
}

func TestEmitC4DTemplateUseInclude(t *testing.T) {
	t.Parallel()

	src := "template svc(name, tech) {\n\ttechnology: ${tech}\n}\n" +
		"use svc(name: auth, tech: Go)\ninclude ./shared.c4d once\n"

	out := c4d.EmitC4D(parseDoc(t, src))

	assert.Contains(t, out, "template svc(name, tech) {\n", "template declaration")
	assert.Contains(t, out, "\n  technology: ${tech}\n", "${param} token rides verbatim")
	assert.Contains(t, out, "\nuse svc(name: auth, tech: Go)\n", "use with named args")
	assert.Contains(t, out, "\ninclude ./shared.c4d once\n", "include with once")
}

func TestEmitC4DBlankLines(t *testing.T) {
	t.Parallel()

	out := c4d.EmitC4D(parseDoc(t, "a: system { }\n\nb: system { }\n"))

	assert.Equal(t, "a: system { }\n\nb: system { }\n", out, "one blank line between top-level units")
}

func TestEmitC4DComments(t *testing.T) {
	t.Parallel()

	src := "# lead comment\nwebapp: system {\n\t# field comment\n\tdescription: serves\n\t# orphan\n}\n"

	out := c4d.EmitC4D(parseDoc(t, src))

	// Leads emit on their own line immediately before their statement; a
	// comment at or below its statement's line is its same-line tail (gofmt
	// semantics, D-32) — the self-consistent placement the grammar
	// re-attaches identically.
	assert.Equal(t,
		"# lead comment\nwebapp: system {\n  # field comment\n  description: serves # orphan\n}\n",
		out, "comment placement (D-32)")

	// The placed text is stable: re-parsing and re-emitting is byte-identical.
	assert.Equal(t, out, c4d.EmitC4D(parseDoc(t, out)), "comment placement is fixpoint-stable")
}

func TestEmitC4DFixpointASTLevel(t *testing.T) {
	t.Parallel()

	src := `# Diagram header
properties {
	name: Demo
	expanded: [webapp]
}

# Serves traffic
webapp: system "Web App" {
	description: serves traffic
	-> db: "SQL | queries" { color: red }
	db: db {
		technology: PostgreSQL
	}
}

pay: system external "Pay" { }

template svc(name, tech) {
	technology: ${tech}
	-> bus: publishes ${name} events
}

use svc(name: auth, tech: Go)
include ./shared.c4d once
# tail comment
`

	// AST-level fixpoint (T-35-04-01): emit, re-parse through the AST entry,
	// re-emit — byte-identical. Only *ast.Document is consumed (same-wave
	// safe; never a Model-returning entry).
	doc1 := parseDoc(t, src)
	emit1 := c4d.EmitC4D(doc1)

	doc2 := parseDoc(t, emit1)
	emit2 := c4d.EmitC4D(doc2)

	assert.Exactly(t, emit1, emit2, "emit(parse(emit(doc))) must equal emit(doc)")
}

func TestEmitC4DNilDocument(t *testing.T) {
	t.Parallel()

	assert.Empty(t, c4d.EmitC4D(nil), "nil document emits empty text")
}

// TestEmitC4DLeafFieldLimit pins the planner-pinned compact-leaf rule: at
// most 3 fields stay on one line; a 4-field unit goes multi-line.
func TestEmitC4DLeafFieldLimit(t *testing.T) {
	t.Parallel()

	three := c4d.EmitC4D(parseDoc(t, "db: db { description: a; color: red; style: solid }\n"))
	assert.Contains(t, three, "db: db { description: a; color: red; style: solid }\n", "3 fields stay one-line")

	four := c4d.EmitC4D(parseDoc(t, "db: db { description: a; color: red; style: solid; border: black }\n"))
	assert.Contains(t, four, "db: db {\n", "4 fields force multi-line")
	assert.Contains(t, four, "\n  border: black\n", "fields at depth 1")
}

// TestEmitC4DFieldsSortedCanonically pins D-23 at the AST level: fields emit
// in the fixed canonical order regardless of source order.
func TestEmitC4DFieldsSortedCanonically(t *testing.T) {
	t.Parallel()

	doc := parseDoc(t, "db: db { color: red; description: a; technology: sql }\n")

	out := c4d.EmitC4D(doc)

	assertOrdered(t, out, "description: a", "technology: sql", "color: red")
}

// TestEmitC4DQuotingNormalization pins the D-06 value-form rules: emission
// is Kind-preserving (quoted safe values stay quoted — the fmt contract),
// and FromModel's canonical literal rule quotes unsafe values.
func TestEmitC4DQuotingNormalization(t *testing.T) {
	t.Parallel()

	t.Run("kind-preserving render", func(t *testing.T) {
		t.Parallel()

		doc := &ast.Document{
			Units: []*ast.UnitNode{{
				ID:   "x",
				Type: "system",
				Fields: []*ast.FieldStmt{
					{Key: "description", Value: ast.Literal{Kind: ast.KindQuoted, Str: "kept quoted"}},
					{Key: "technology", Value: ast.Literal{Kind: ast.KindBareword, Str: "bare"}},
					{Key: "color", Value: ast.Literal{Kind: ast.KindQuoted, Str: `has "quotes" \ and, commas`}},
				},
			}},
		}

		out := c4d.EmitC4D(doc)

		assert.Contains(t, out, `description: "kept quoted"`, "KindQuoted renders quoted")
		assert.Contains(t, out, "technology: bare", "KindBareword renders verbatim")
		assert.Contains(t, out, `color: "has \"quotes\" \\ and, commas"`, "escapes keep the value parseable")
	})

	t.Run("frommodel quotes unsafe values", func(t *testing.T) {
		t.Parallel()

		m := &parser.Model{
			UnitOrder: []string{"x"},
			Units: map[string]*model.Unit{
				"x": {Type: model.TypeSystem, Name: "X", Description: "has, comma"},
			},
		}

		out := c4d.EmitC4D(c4d.FromModel(m))

		assert.Contains(t, out, `description: "has, comma"`, "comma values cannot ride barewords")
	})
}
