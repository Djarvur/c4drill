// Package c4d_test — ToModel: typed AST -> *parser.Model with full parity
// hooks (Plan 35-05 Task 2).
//
// The load-bearing contract is D-02/D-21 parity: for equivalent documents
// the C4D front-end produces the SAME *parser.Model the TOML front-end
// produces — same inference (DefaultTypeForParent/InferGenericType), same
// Humanize name derivation, same order slices, same Link shapes. The twin
// tests assert require.Equal on the parsed structures, not field spot
// checks, so any drift fails loudly.
package c4d_test

import (
	"os"
	"testing"

	"github.com/Djarvur/c4drill/internal/c4d"
	"github.com/Djarvur/c4drill/internal/model"
	"github.com/Djarvur/c4drill/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// toModel runs the C4D front-end end to end (ParseAST + ToModel via
// c4d.Parse) and returns the Model.
func toModel(t *testing.T, src string) *parser.Model {
	t.Helper()

	m, err := c4d.Parse([]byte(src))
	require.NoError(t, err, "c4d.Parse() should not error")

	return m
}

// toModelErr asserts that src fails the full front-end (parse or convert).
//
//nolint:wrapcheck // tests inspect the *parser.ParseError unwrapped
func toModelErr(t *testing.T, src string) error {
	t.Helper()

	m, err := c4d.Parse([]byte(src))
	require.Error(t, err, "c4d.Parse() should error on %q", src)
	assert.Nil(t, m, "no model on error")

	return err
}

// TestToModelFullDocument maps every statement kind into the corresponding
// parser.Model field with UnitOrder/SubunitOrder reflecting AST statement
// order (D-21).
func TestToModelFullDocument(t *testing.T) {
	t.Parallel()

	m := toModel(t, `properties {
	name: Demo
	lineLength: 42
	expanded: [gateway]
}

gateway: system "Gateway" {
	description: serves traffic
	api: container {
		-> db: sql | queries
	}
	db: db { }
}

template svc(name) {
	technology: ${name}
}

use svc(name: x)

include ./other.c4d once
`)

	assert.Equal(t, "Demo", m.Properties.Name, "Properties.Name")
	assert.Equal(t, 42, m.Properties.LineLength, "Properties.LineLength")
	assert.Equal(t, []string{"gateway"}, m.Properties.Expanded, "Properties.Expanded (list literal)")

	require.Len(t, m.UnitOrder, 1, "UnitOrder")
	require.Contains(t, m.Units, "gateway", "Units[gateway]")

	gw := m.Units["gateway"]
	assert.Equal(t, model.TypeSystem, gw.Type, "gateway.Type")
	assert.Equal(t, "Gateway", gw.Name, "gateway.Name (header name)")
	assert.Equal(t, "serves traffic", gw.Description, "gateway.Description")
	assert.Equal(t, []string{"api", "db"}, gw.SubunitOrder, "SubunitOrder (statement order)")

	api := gw.Subunits["api"]
	require.NotNil(t, api, "Subunits[api]")
	assert.Equal(t, model.TypeContainer, api.Type, "api.Type")
	require.Len(t, api.Links, 1, "api.Links")
	assert.Equal(t, "db", api.Links[0].Peer, "api.Links[0].Peer (verbatim, D-10)")

	db := gw.Subunits["db"]
	require.NotNil(t, db, "Subunits[db]")
	assert.Equal(t, model.TypeContainerDb, db.Type, "db.Type (generic inference in system)")

	require.Contains(t, m.Templates, "svc", "Templates[svc]")
	assert.Equal(t, []string{"name"}, m.Templates["svc"].Params, "TemplateDef.Params")

	require.Len(t, m.Instantiations, 1, "Instantiations")
	assert.Equal(t, parser.Instantiation{
		Template: "svc",
		Parent:   "",
		Params:   map[string]string{"name": "x"},
	}, m.Instantiations[0], "Instantiation")

	require.Equal(t, []parser.IncludeDirective{{Path: "./other.c4d", Once: true}}, m.Includes,
		"Includes (path + once)")
}

// TestToModelInferenceParity pins D-02: the C4D document `sys: system { api { } }`
// and its TOML twin produce IDENTICAL *model.Unit structs — Type inferred via
// DefaultTypeForParent + InferGenericType, Name humanized from the id, order
// slices identical, Subunits nil-ness identical.
func TestToModelInferenceParity(t *testing.T) {
	t.Parallel()

	tomlModel, err := parser.Parse([]byte(`
[sys]
type = "system"

[sys.api]
`))
	require.NoError(t, err, "parser.Parse(TOML twin)")

	dslModel := toModel(t, "sys: system { api { } }")

	require.Contains(t, tomlModel.Units, "sys", "TOML sys unit")
	require.Contains(t, dslModel.Units, "sys", "C4D sys unit")

	require.Equal(t, tomlModel.Units["sys"], dslModel.Units["sys"],
		"identical *model.Unit for equivalent documents (inference + humanize parity, D-02)")

	assert.Equal(t, model.TypeContainer, dslModel.Units["sys"].Subunits["api"].Type,
		"api.Type inferred from system parent")
	assert.Equal(t, "Api", dslModel.Units["sys"].Subunits["api"].Name,
		"api.Name humanized from the identifier")
}

// TestToModelExternalModifier maps the `external` modifier to the *External
// UnitType variants (D-04 at Model level) and hard-errors on types without
// an external variant.
func TestToModelExternalModifier(t *testing.T) {
	t.Parallel()

	t.Run("person external type-led", func(t *testing.T) {
		t.Parallel()

		m := toModel(t, `person external "X" { }`)

		require.Len(t, m.UnitOrder, 1, "one unit")
		// Type-led header without id: the quoted display name is the unit key
		// (the TOML twin is the quoted table ["X"]).
		unit := m.Units[m.UnitOrder[0]]
		assert.Equal(t, model.TypePersonExternal, unit.Type, "Type personExternal")
		assert.Equal(t, "X", unit.Name, "Name from header")
	})

	t.Run("system external with id", func(t *testing.T) {
		t.Parallel()

		m := toModel(t, "partner: system external { }")

		assert.Equal(t, model.TypeSystemExternal, m.Units["partner"].Type, "Type systemExternal")
	})

	t.Run("box external has no variant", func(t *testing.T) {
		t.Parallel()

		err := toModelErr(t, "b: box external { }")

		var perr *parser.ParseError
		require.ErrorAs(t, err, &perr, "*parser.ParseError")
		assert.Contains(t, err.Error(), "external", "message names the modifier problem")
	})
}

// TestToModelEdges pins the glyph->Link mapping: `->` Links with
// ArrowForward, `<-` LinksFrom (TOML linkFrom semantics, arrowhead at the
// owner), `<->` ArrowBidirectional, `--` ArrowNone; option block keys map to
// Link fields; the arrow option overrides the glyph default; peer strings
// ride verbatim (bare or dotted — peer.Resolve resolves, D-10).
func TestToModelEdges(t *testing.T) {
	t.Parallel()

	m := toModel(t, `a: system {
	-> db: sql | queries { color: red rank: equal labelPosition: head length: 2 }
	<- other
	<-> both
	-- plain
	-> dotted.path
	-> rev { arrow: reverse }
}
`)

	unit := m.Units["a"]

	require.Equal(t, []model.Link{
		{
			Peer: "db", Arrow: model.ArrowForward, Technology: "sql", Description: "queries",
			Color: "red", Rank: model.RankEqual, LabelPosition: model.LabelHead, Length: 2,
		},
		{Peer: "both", Arrow: model.ArrowBidirectional},
		{Peer: "plain", Arrow: model.ArrowNone},
		{Peer: "dotted.path", Arrow: model.ArrowForward},
		{Peer: "rev", Arrow: model.ArrowReverse},
	}, unit.Links, "Links (statement order, glyph mapping)")

	require.Equal(t, []model.Link{
		{Peer: "other", Arrow: model.ArrowForward},
	}, unit.LinksFrom, "LinksFrom (`<-` — arrowhead at the owner, mirror-consistent)")
}

// TestToModelDuplicateEdgeError pins D-11: two edge statements for the same
// (unit, peer) pair in one block are a hard error naming the peer; opposite
// directions do not collide (they live in different lists).
func TestToModelDuplicateEdgeError(t *testing.T) {
	t.Parallel()

	t.Run("duplicate forward", func(t *testing.T) {
		t.Parallel()

		err := toModelErr(t, "a: system { -> db\n-> db: other label }")

		var perr *parser.ParseError
		require.ErrorAs(t, err, &perr, "*parser.ParseError")
		assert.Contains(t, err.Error(), "duplicate", "message names the duplicate")
		assert.Contains(t, err.Error(), `"db"`, "message names the peer")
		assert.Positive(t, perr.Line, "ParseError.Line (AST position)")
	})

	t.Run("duplicate incoming", func(t *testing.T) {
		t.Parallel()

		err := toModelErr(t, "a: system { <- db\n<- db }")

		assert.Contains(t, err.Error(), "duplicate", "duplicate in LinksFrom errors too")
	})

	t.Run("opposite directions do not collide", func(t *testing.T) {
		t.Parallel()

		m := toModel(t, "a: system { -> db\n<- other }")

		require.Len(t, m.Units["a"].Links, 1, "Links")
		require.Len(t, m.Units["a"].LinksFrom, 1, "LinksFrom — separate lists, no collision")
	})

	t.Run("same peer both directions is not a duplicate", func(t *testing.T) {
		t.Parallel()

		m := toModel(t, "a: system { -> db\n<- db }")

		require.Len(t, m.Units["a"].Links, 1, "Links")
		require.Len(t, m.Units["a"].LinksFrom, 1, "LinksFrom — different lists, legal pair")
	})
}

// TestToModelDuplicateUnitError: two units claiming the same path in one
// document are a hard error — the TOML front-end errors on the duplicate
// table twin, so the C4D front-end must too (fail-closed parity).
func TestToModelDuplicateUnitError(t *testing.T) {
	t.Parallel()

	t.Run("duplicate top level", func(t *testing.T) {
		t.Parallel()

		err := toModelErr(t, "a: system { }\na: db { }")

		assert.Contains(t, err.Error(), "duplicate", "message names the duplicate")
		assert.Contains(t, err.Error(), `"a"`, "message names the path")
	})

	t.Run("duplicate subunit", func(t *testing.T) {
		t.Parallel()

		err := toModelErr(t, "a: system { b { }\nb: db { } }")

		assert.Contains(t, err.Error(), "duplicate", "message names the duplicate")
	})
}

// TestToModelUseParents pins the three use positions: top level -> Parent "",
// inside a unit block -> Parent is the enclosing dotted path (D-16), inside a
// template body -> TemplateDef.Instantiations with root-relative Parent (D-17).
//
// Ordering note: the AST stores top-level statements in canonical sections
// (units before uses, the EmitC4D canonical order), so unit-nested uses
// surface in m.Instantiations BEFORE top-level uses regardless of source
// position — the TOML twin lists the [[unit.use]] tables before the
// top-level [[use]].
func TestToModelUseParents(t *testing.T) {
	t.Parallel()

	m := toModel(t, `template svc(name) { }

use svc(name: top)

a: system {
	b: box {
		use svc(name: nested)
	}
}

template outer(p) {
	api: container {
		use svc(name: inner)
	}
}
`)

	require.Len(t, m.Instantiations, 2, "document-level instantiations")

	assert.Equal(t, parser.Instantiation{
		Template: "svc",
		Parent:   "a.b",
		Params:   map[string]string{"name": "nested"},
	}, m.Instantiations[0], "unit-nested use -> Parent a.b (D-16)")

	assert.Equal(t, parser.Instantiation{
		Template: "svc",
		Parent:   "",
		Params:   map[string]string{"name": "top"},
	}, m.Instantiations[1], "top-level use -> Parent \"\"")

	require.Contains(t, m.Templates, "outer", "Templates[outer]")
	require.Len(t, m.Templates["outer"].Instantiations, 1, "template-body use (D-17)")

	assert.Equal(t, parser.Instantiation{
		Template: "svc",
		Parent:   "api",
		Params:   map[string]string{"name": "inner"},
	}, m.Templates["outer"].Instantiations[0], "template-body use Parent is root-relative")
}

// TestToModelUsePositionalArgs: positional args pair with the template's
// declared params in order; named args map directly; positional args for a
// template not declared in the same file are a hard error (pairing is a
// conversion-time decision — named args always work).
func TestToModelUsePositionalArgs(t *testing.T) {
	t.Parallel()

	t.Run("positional pairs with declared params", func(t *testing.T) {
		t.Parallel()

		m := toModel(t, "template svc(name, tech) { }\nuse svc(\"api\", \"Go\")")

		require.Len(t, m.Instantiations, 1, "Instantiations")
		assert.Equal(t, map[string]string{"name": "api", "tech": "Go"},
			m.Instantiations[0].Params, "positional args paired with Params order")
	})

	t.Run("too many positional args", func(t *testing.T) {
		t.Parallel()

		err := toModelErr(t, "template svc(name) { }\nuse svc(\"api\", \"Go\")")

		assert.Contains(t, err.Error(), "positional", "message names the problem")
	})

	t.Run("positional without same-file template", func(t *testing.T) {
		t.Parallel()

		err := toModelErr(t, "use nosuch(\"api\")")

		assert.Contains(t, err.Error(), "named", "message suggests named args")
	})

	t.Run("duplicate named arg", func(t *testing.T) {
		t.Parallel()

		err := toModelErr(t, "use svc(name: a, name: b)\ntemplate svc(name) { }")

		assert.Contains(t, err.Error(), "name", "message names the duplicated key")
	})
}

// TestToModelTemplateTokens: template bodies carry literal ${param} tokens
// into TemplateDef.Unit — never substituted at conversion time (the
// TemplateDef/Expand contract).
func TestToModelTemplateTokens(t *testing.T) {
	t.Parallel()

	m := toModel(t, "template svc(name) {\n\tdescription: ${name} handles\n\t-> cache: ${name} publishes\n}")

	tmpl := m.Templates["svc"]
	require.NotNil(t, tmpl, "Templates[svc]")
	require.NotNil(t, tmpl.Unit, "TemplateDef.Unit")

	assert.Equal(t, "${name} handles", tmpl.Unit.Description, "token rides verbatim")
	require.Len(t, tmpl.Unit.Links, 1, "Unit.Links")
	assert.Equal(t, "cache", tmpl.Unit.Links[0].Peer, "link peer")
	assert.Equal(t, "${name} publishes", tmpl.Unit.Links[0].Description, "link token rides verbatim")
}

// TestParseASTExported pins the exported AST-level entry points (35-06
// canonsrc and 35-08 fmt consume them from other packages): ParseAST and
// ParseASTFile return (*ast.Document, error) with comments attached, and
// c4d.Parse composes them (Parse == ToModel(ParseAST), D-21/D-32).
func TestParseASTExported(t *testing.T) {
	t.Parallel()

	doc, err := c4d.ParseAST([]byte("# lead\na: system { }"))

	require.NoError(t, err, "ParseAST() should not error")
	require.NotNil(t, doc, "ParseAST() document")
	require.Len(t, doc.Units, 1, "doc.Units")
	require.Len(t, doc.Units[0].Comments, 1, "comments attached (D-32)")
	assert.Equal(t, "lead", doc.Units[0].Comments[0].Text, "lead comment text")

	src := "properties { name: Demo }\na: system { -> b: sql | q }\nb: db { }"

	viaParse, err := c4d.Parse([]byte(src))
	require.NoError(t, err, "c4d.Parse composition")

	astDoc, err := c4d.ParseAST([]byte(src))
	require.NoError(t, err, "c4d.ParseAST")

	viaToModel, err := c4d.ToModel(astDoc)
	require.NoError(t, err, "ToModel(ParseAST(data))")

	require.Equal(t, viaParse, viaToModel, "Parse(data) == ToModel(ParseAST(data))")
}

// TestParseASTFileExported: ParseASTFile keeps the file-level AST entry.
func TestParseASTFileExported(t *testing.T) {
	t.Parallel()

	path := t.TempDir() + "/doc.c4d"
	require.NoError(t, os.WriteFile(path, []byte("a: system { }"), 0o600), "write fixture")

	doc, err := c4d.ParseASTFile(path)

	require.NoError(t, err, "ParseASTFile() success path")
	require.NotNil(t, doc, "ParseASTFile() document")
	require.Len(t, doc.Units, 1, "doc.Units")
}

// TestToModelTemplateBasicTwin: the C4D twin of testdata/template_basic.toml
// produces a require.Equal Model. The glyph `->` carries the arrow default
// explicitly (D-22: explicit defaults normalize away), so the TOML twin
// states `arrow = "forward"` where the fixture omits the key — the parsed
// Links are identical.
func TestToModelTemplateBasicTwin(t *testing.T) {
	t.Parallel()

	// testdata/template_basic.toml with the implicit arrow default made
	// explicit (semantically identical — forward IS the model default).
	tomlSrc := `
[properties]
name = "Basic Template Test"

[template.svc]
params = ["name", "tech"]
name = "${name} Service"
type = "system"
technology = "${tech}"
description = "${name} handles requests"
color = "${name}-color"

[[template.svc.link]]
peer = "messageBus"
description = "Publishes ${name} events"
technology = "${tech}"
arrow = "forward"

[[use]]
template = "svc"
name = "auth"
tech = "Go"

[messageBus]
type = "queue"
name = "Message Bus"
`

	tomlModel, err := parser.Parse([]byte(tomlSrc))
	require.NoError(t, err, "parser.Parse(TOML twin)")

	dslSrc := `properties { name: Basic Template Test }

messageBus: queue "Message Bus" { }

template svc(name, tech) {
	name: "${name} Service"
	technology: ${tech}
	description: ${name} handles requests
	color: ${name}-color
	-> messageBus: ${tech} | Publishes ${name} events { arrow: forward }
}

use svc(name: auth, tech: Go)
`

	dslModel := toModel(t, dslSrc)

	require.Equal(t, tomlModel.Properties, dslModel.Properties, "Properties equal")
	require.Equal(t, tomlModel.UnitOrder, dslModel.UnitOrder, "UnitOrder equal")
	require.Equal(t, tomlModel.Units, dslModel.Units, "Units equal")
	require.Equal(t, tomlModel.Templates, dslModel.Templates, "Templates equal (params, body, links)")
	require.Equal(t, tomlModel.Instantiations, dslModel.Instantiations, "Instantiations equal")
	require.Equal(t, tomlModel.Includes, dslModel.Includes, "Includes equal")
}

// TestToModelTemplateRootTypeFromModel closes the 35-04 gap: FromModel
// records a template root's type on Body.Type (the grammar has no
// template-root-type syntax), and ToModel honors it —
// ToModel(FromModel(m)) preserves the template root type.
func TestToModelTemplateRootTypeFromModel(t *testing.T) {
	t.Parallel()

	tomlModel, err := parser.Parse([]byte(`
[template.actor]
type = "person"
name = "External ${name}"

[[use]]
template = "actor"
name = "user"
`))
	require.NoError(t, err, "parser.Parse")

	require.NotNil(t, tomlModel.Templates["actor"], "template parsed")
	assert.Equal(t, model.TypePerson, tomlModel.Templates["actor"].Unit.Type, "root type person")

	doc := c4d.FromModel(tomlModel)
	require.Len(t, doc.Templates, 1, "FromModel templates")

	roundTripped, err := c4d.ToModel(doc)
	require.NoError(t, err, "ToModel(FromModel(m))")

	require.Equal(t, tomlModel.Templates, roundTripped.Templates,
		"template root type survives Model->AST->Model (35-04 deferred gap closed)")
	require.Equal(t, tomlModel.Instantiations, roundTripped.Instantiations, "instantiations survive")
}
