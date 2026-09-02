// render_test.go covers the c4drill/renderDiagram custom method (issue #32,
// M4): SVG byte-parity with the CLI render pipeline, render-time diagnostics,
// target/expanded/legend options, and the wire contract on unknown documents.

package lsp_test

import (
	"encoding/json"
	"regexp"
	"strings"
	"testing"

	"github.com/Djarvur/c4drill/internal/graph"
	"github.com/Djarvur/c4drill/internal/lsp"
	"github.com/Djarvur/c4drill/internal/parser"
	"github.com/Djarvur/c4drill/internal/peer"
	"github.com/Djarvur/c4drill/internal/render"
	"github.com/Djarvur/c4drill/internal/validator"
	"github.com/Djarvur/c4drill/internal/view"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mustParseModel parses clean fixture text (helper for the CLI-parity builds).
func mustParseModel(t *testing.T, text string) *parser.Model {
	t.Helper()

	m, err := parser.Parse([]byte(text))
	require.NoError(t, err)

	return m
}

// renderModel is a small clean model with drill-down structure.
const renderModel = `[properties]
name = "Render Fixture"
description = "r"
expanded = ["cloud"]

[user]
type = "person"
name = "User"
description = "u"

[[user.link]]
peer = "cloud.api"
description = "uses"

[cloud]
type = "system"
name = "Cloud"
description = "c"

[cloud.api]
type = "container"
name = "API"
description = "a"

[cloud.db1]
type = "containerDb"
name = "DB"
description = "d"

[[cloud.api.link]]
peer = "db1"
description = "stores"
`

func renderAt(t *testing.T, text string, mutate func(*lsp.RenderDiagramParams)) *lsp.RenderDiagramResult {
	t.Helper()

	h := newHarness(t)
	h.request("initialize", lsp.InitializeResult{})
	h.notify("initialized", lsp.InitializedParams{})

	uri := lsp.DocumentURI("file:///ws/model.toml")
	h.openDoc(uri, text)

	params := lsp.RenderDiagramParams{TextDocument: lsp.TextDocumentIdentifier{URI: uri}}
	if mutate != nil {
		mutate(&params)
	}

	resp := h.request("c4drill/renderDiagram", params)
	require.Nil(t, resp.Error, "renderDiagram must not error: %+v", resp.Error)

	var result lsp.RenderDiagramResult
	require.NoError(t, json.Unmarshal(resp.Result, &result))

	return &result
}

// TestRenderDiagramSVGBtyeParityWithCLI pins acceptance 5: the returned SVG
// is byte-identical to building the same view through the CLI's own stages.
func TestRenderDiagramSVGBtyeParityWithCLI(t *testing.T) {
	t.Parallel()

	result := renderAt(t, renderModel, nil)
	require.Empty(t, result.Diagnostics, "clean model renders without diagnostics")
	require.NotEmpty(t, result.SVG)
	assert.True(t, strings.HasPrefix(strings.TrimSpace(result.SVG), "<svg") ||
		strings.HasPrefix(strings.TrimSpace(result.SVG), "<?xml"),
		"result is SVG text")

	// The CLI stages for the same model and options (root.go: parse ->
	// include/template/peer passes -> Validate -> C1 view ->
	// BuildGraphWithPath -> RenderSVGWithOutput, which is RenderSVG).
	// Validate MUTATES the model (populateIncomingLinks) and runs before
	// view generation — the boundary nodes in the render depend on it.
	m := mustParseModel(t, renderModel)
	require.NoError(t, peer.Resolve(m))

	_ = validator.Validate(m) // the CLI discards the errors after reporting

	v := view.GenerateC1View(m)
	require.NotNil(t, v)
	g := graph.BuildGraphWithPath(v, "", "model", "svg")
	require.NotNil(t, g)

	want, err := render.RenderSVGWithOutput(g, "")
	require.NoError(t, err)

	assert.Equal(t, string(want), result.SVG, "byte-identical to the CLI pipeline")
}

func TestRenderDiagramTargetC2(t *testing.T) {
	t.Parallel()

	result := renderAt(t, renderModel, func(p *lsp.RenderDiagramParams) {
		p.Target = "cloud"
	})

	require.NotEmpty(t, result.SVG)

	// Byte parity for the C2 target through the CLI stages, too.
	m := mustParseModel(t, renderModel)
	require.NoError(t, peer.Resolve(m))

	_ = validator.Validate(m) // the CLI discards the errors after reporting

	v := view.GenerateC2View(m, "cloud")
	require.NotNil(t, v)
	g := graph.BuildGraphWithPath(v, "cloud", "model", "svg")
	require.NotNil(t, g)

	want, err := render.RenderSVG(g)
	require.NoError(t, err)

	// The WASM graphviz engine numbers anchor ids (a_graph0_N) from a
	// process-global counter, so repeated in-process renders drift by index
	// while geometry and content stay identical (the CLI renders once per
	// process and never sees this). Byte-compare modulo the counter.
	assert.Equal(t, normalizeAnchorIDs(string(want)), normalizeAnchorIDs(result.SVG))
}

// anchorIDRe matches the engine's anchor id suffix.
var anchorIDRe = regexp.MustCompile(`a_graph0_\d+`)

// normalizeAnchorIDs collapses the engine's anchor counter drift.
func normalizeAnchorIDs(svg string) string {
	return anchorIDRe.ReplaceAllString(svg, "a_graph0")
}

func TestRenderDiagramAllExpanded(t *testing.T) {
	t.Parallel()

	result := renderAt(t, renderModel, func(p *lsp.RenderDiagramParams) {
		p.AllExpanded = true
	})

	require.NotEmpty(t, result.SVG)

	// The expanded graph shows nested units the C1 view hides: its SVG
	// differs from the plain C1 render.
	c1 := renderAt(t, renderModel, nil)
	assert.NotEqual(t, c1.SVG, result.SVG, "all-expanded differs from C1")
}

func TestRenderDiagramExpandedSetOverride(t *testing.T) {
	t.Parallel()

	// The model's [properties].expanded = ["cloud"] drills into cloud on C1;
	// an empty override collapses it.
	collapsed := renderAt(t, renderModel, func(p *lsp.RenderDiagramParams) {
		p.Expanded = []string{}
	})
	defaultRender := renderAt(t, renderModel, nil)

	require.NotEmpty(t, collapsed.SVG)
	assert.NotEqual(t, defaultRender.SVG, collapsed.SVG,
		"the expanded override changes the C1 drill-down")
}

func TestRenderDiagramLegendToggle(t *testing.T) {
	t.Parallel()

	withLegend := renderAt(t, renderModel, func(p *lsp.RenderDiagramParams) {
		p.Legend = boolPtr(true)
	})
	withoutLegend := renderAt(t, renderModel, func(p *lsp.RenderDiagramParams) {
		p.Legend = boolPtr(false)
	})

	require.NotEmpty(t, withLegend.SVG)
	require.NotEmpty(t, withoutLegend.SVG)
	assert.NotEqual(t, withLegend.SVG, withoutLegend.SVG, "the legend toggle changes the render")
}

func boolPtr(b bool) *bool { return &b }

func TestRenderDiagramInvalidModelCarriesDiagnostics(t *testing.T) {
	t.Parallel()

	invalid := "[user]\ntype = \"person\"\nname = \"U\"\n"
	result := renderAt(t, invalid, nil)

	assert.Empty(t, result.SVG, "an invalid model renders nothing (CLI parity)")
	require.Len(t, result.Diagnostics, 1)
	assert.Equal(t,
		`error: unit "user" has no incoming or outgoing links in user`,
		result.Diagnostics[0].Message)

	// A parse failure lands in diagnostics the same way.
	broken := "[web\ntype = \"system\"\n"
	result = renderAt(t, broken, nil)
	require.Len(t, result.Diagnostics, 1)
	assert.Contains(t, result.Diagnostics[0].Message, "parse: parse error at line 1:")
}

func TestRenderDiagramUnknownDocumentIsNull(t *testing.T) {
	t.Parallel()

	h := newHarness(t)
	h.request("initialize", lsp.InitializeResult{})
	h.notify("initialized", lsp.InitializedParams{})

	resp := h.request("c4drill/renderDiagram", lsp.RenderDiagramParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: "file:///ws/never-opened.toml"},
	})
	require.Nil(t, resp.Error)
	assert.JSONEq(t, "null", string(resp.Result))
}

func TestRenderDiagramUnsupportedFormatIsDiagnostic(t *testing.T) {
	t.Parallel()

	result := renderAt(t, renderModel, func(p *lsp.RenderDiagramParams) {
		p.Format = "png"
	})

	assert.Empty(t, result.SVG)
	require.Len(t, result.Diagnostics, 1)
	assert.Contains(t, result.Diagnostics[0].Message, "unsupported render format")
}
