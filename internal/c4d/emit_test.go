package c4d_test

import (
	"os"
	"strings"
	"testing"

	"github.com/Djarvur/c4drill/internal/c4d"
	"github.com/Djarvur/c4drill/internal/model"
	"github.com/Djarvur/c4drill/internal/parser"
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
				Expanded:    []string{"api"},
			},
		},
	}

	out, err := c4d.EmitTOML(m)
	require.NoError(t, err, "EmitTOML() should not error")

	// D-23 fixed order: type, name, description, technology, reference,
	// color, style, border, edges, expanded.
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
		"expanded = ",
	)
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
					Peer:           "db",
					Arrow:          model.ArrowReverse,
					Rank:           model.RankReverse,
					Color:          "green",
					Style:          "dashed",
					Technology:     "SQL",
					Description:    "Queries",
					LabelPosition:  model.LabelHead,
					Length:         2,
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
