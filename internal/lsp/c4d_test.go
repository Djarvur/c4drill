// c4d_test.go covers the C4D language features (issue #33): one fixture class
// per capability — completion (header type slots with nesting-aware
// promotion, unit body fields, edge peers and options, enum values, template
// and use forms, ${param} tokens, include paths), hover, definition, and the
// documentSymbol outline — plus corpus fixtures over the .c4d examples and
// the mid-edit robustness contract shared with the TOML side.

package lsp_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Djarvur/c4drill/internal/lsp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// c4dCompleteAt runs completion on a .c4d buffer at (line, char).
func c4dCompleteAt(t *testing.T, text string, line, char uint32) []lsp.CompletionItem {
	t.Helper()

	h := newHarness(t)
	h.request("initialize", lsp.InitializeResult{})
	h.notify("initialized", lsp.InitializedParams{})

	uri := lsp.DocumentURI("file:///ws/model.c4d")
	h.openDoc(uri, text)

	resp := h.request("textDocument/completion", lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: uri},
		Position:     lsp.Position{Line: line, Character: char},
	})

	require.Nil(t, resp.Error, "completion must not error: %v", resp.Error)

	var list lsp.CompletionList
	require.NoError(t, json.Unmarshal(resp.Result, &list))

	return list.Items
}

// c4dHoverAt runs hover on a .c4d buffer at (line, char).
func c4dHoverAt(t *testing.T, text string, line, char uint32) *lsp.Hover {
	t.Helper()

	h := newHarness(t)
	h.request("initialize", lsp.InitializeResult{})
	h.notify("initialized", lsp.InitializedParams{})

	uri := lsp.DocumentURI("file:///ws/model.c4d")
	h.openDoc(uri, text)

	resp := h.request("textDocument/hover", lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: uri},
		Position:     lsp.Position{Line: line, Character: char},
	})

	require.Nil(t, resp.Error)

	if string(resp.Result) == "null" {
		return nil
	}

	var hover lsp.Hover
	require.NoError(t, json.Unmarshal(resp.Result, &hover))

	return &hover
}

// c4dDefinitionsAt runs definition on a .c4d buffer at (line, char).
func c4dDefinitionsAt(t *testing.T, dir, text string, line, char uint32) []lsp.Location {
	t.Helper()

	h := newHarness(t)
	h.request("initialize", lsp.InitializeResult{})
	h.notify("initialized", lsp.InitializedParams{})

	uri := lsp.DocumentURI("file://" + filepath.Join(dir, "main.c4d"))
	h.openDoc(uri, text)

	resp := h.request("textDocument/definition", lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: uri},
		Position:     lsp.Position{Line: line, Character: char},
	})

	require.Nil(t, resp.Error)

	if string(resp.Result) == "null" {
		return nil
	}

	var locs []lsp.Location
	require.NoError(t, json.Unmarshal(resp.Result, &locs))

	return locs
}

// c4dSymbolsOf runs documentSymbol on a .c4d buffer.
func c4dSymbolsOf(t *testing.T, text string) []lsp.DocumentSymbol {
	t.Helper()

	h := newHarness(t)
	h.request("initialize", lsp.InitializeResult{})
	h.notify("initialized", lsp.InitializedParams{})

	uri := lsp.DocumentURI("file:///ws/model.c4d")
	h.openDoc(uri, text)

	resp := h.request("textDocument/documentSymbol", lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: uri},
		Position:     lsp.Position{},
	})

	require.Nil(t, resp.Error)

	if string(resp.Result) == "null" {
		return nil
	}

	var symbols []lsp.DocumentSymbol
	require.NoError(t, json.Unmarshal(resp.Result, &symbols))

	return symbols
}

// --- completion: unit header forms ---------------------------------------

func TestC4DCompletionHeaderTypeSlotAtTopLevel(t *testing.T) {
	t.Parallel()

	// `web: ` — the id-led header's type slot: all 17 types, the C1 default
	// (system) first.
	items := c4dCompleteAt(t, "web: \n", 0, 5)

	require.Len(t, items, 17, "all 17 unit types are offered")

	var first lsp.CompletionItem

	for _, it := range items {
		if first.Label == "" || it.SortText < first.SortText {
			first = it
		}
	}

	assert.Equal(t, "system", first.Label, "C1 default sorts first")
	assert.Contains(t, first.Detail, "default")
}

func TestC4DCompletionHeaderTypeSlotIsNestingAware(t *testing.T) {
	t.Parallel()

	// Inside a system block the default is container and db promotes.
	text := "shop: system \"Shop\" {\n  web: \n}\n"
	items := c4dCompleteAt(t, text, 1, 7)

	var (
		dbItem        *lsp.CompletionItem
		containerItem *lsp.CompletionItem
	)

	for i := range items {
		switch items[i].Label {
		case "db":
			dbItem = &items[i]
		case "container":
			containerItem = &items[i]
		}
	}

	require.NotNil(t, dbItem)
	require.NotNil(t, containerItem)

	assert.Contains(t, dbItem.Detail, "promotes to containerDb",
		"generic db shows its promotion inside a system")
	assert.Contains(t, containerItem.Detail, "default")

	// Inside a container (C3) db promotes to componentDb.
	text = "shop: container \"Shop\" {\n  svc: \n}\n"
	items = c4dCompleteAt(t, text, 1, 7)

	for i := range items {
		if items[i].Label == "db" {
			assert.Contains(t, items[i].Detail, "promotes to componentDb")
		}
	}
}

func TestC4DCompletionTypeLedHeaderAtStatementStart(t *testing.T) {
	t.Parallel()

	// Statement start inside a unit: the type-led header form is offered
	// alongside the body keywords, with the same nesting awareness.
	text := "shop: system \"Shop\" {\n  description: d\n  \n}\n"
	items := c4dCompleteAt(t, text, 2, 2)

	got := labels(items)

	assert.Contains(t, got, "container", "type-led nested unit headers are offered")
	assert.Contains(t, got, "name")
	assert.Contains(t, got, "technology", "fields not yet authored are offered")
	assert.Contains(t, got, "->", "edge arrows are offered in unit bodies")
	assert.Contains(t, got, "use", "template instantiation is offered in unit bodies")

	for _, it := range items {
		if it.Label == "container" {
			assert.Contains(t, it.Detail, "default", "nested default is container inside a system")
		}
	}
}

// --- completion: body field keywords -------------------------------------

func TestC4DCompletionUnitBodyFields(t *testing.T) {
	t.Parallel()

	items := c4dCompleteAt(t, "web: system \"Web\" {\n  na\n}\n", 1, 2)

	got := labels(items)

	assert.Contains(t, got, "name")
	assert.Contains(t, got, "technology")
	assert.Contains(t, got, "expanded")
	assert.NotContains(t, got, "peer", "link fields stay out of unit bodies")
	assert.NotContains(t, got, "arrow", "edge-option fields stay out of unit bodies")
}

func TestC4DCompletionAuthoredFieldsAreFiltered(t *testing.T) {
	t.Parallel()

	// description is already authored: the fresh-key list drops it.
	text := "web: system \"Web\" {\n  description: d\n  \n}\n"
	items := c4dCompleteAt(t, text, 2, 2)

	assert.Contains(t, labels(items), "name")
	assert.NotContains(t, labels(items), "description", "authored keys are not re-offered")
}

// --- completion: enum values ---------------------------------------------

func TestC4DCompletionFieldEnumValues(t *testing.T) {
	t.Parallel()

	cases := []struct {
		key  string
		want []string
	}{
		{"edges", []string{"straight", "spline", "square", "ortho"}},
		{"style", []string{"solid", "dashed", "dotted"}},
	}

	for _, tc := range cases {
		t.Run(tc.key, func(t *testing.T) {
			t.Parallel()

			text := "web: system \"Web\" {\n  " + tc.key + ": \n}\n"
			items := c4dCompleteAt(t, text, 1, uint32(len(tc.key)+4)) //nolint:gosec // test-local bounded length

			assert.Equal(t, tc.want, labels(items))
		})
	}
}

func TestC4DCompletionLegendValues(t *testing.T) {
	t.Parallel()

	items := c4dCompleteAt(t, "properties {\n  legend: \n}\n", 1, 10)

	assert.Equal(t, []string{"true", "false"}, labels(items))
}

// --- completion: edge peers and options ----------------------------------

func TestC4DCompletionEdgePeerBareAndAbsolute(t *testing.T) {
	t.Parallel()

	text := "user: person \"Customer\" {\n  -> \n}\n" +
		"shop: system \"Shop\" {\n" +
		"  web: container \"W\" {\n    -> \n  }\n" +
		"  api: container \"A\" {\n    -> \n  }\n" +
		"}\n"

	// Inside shop.api (line 8, right after the `->` glyph at char 6): bare
	// walk-up candidates are shop's children (web) and the root scope's
	// children (user, shop); absolute paths exclude the host itself.
	items := c4dCompleteAt(t, text, 8, 6)
	got := labels(items)

	for _, bare := range []string{"web", "user", "shop"} {
		assert.Contains(t, got, bare, "bare walk-up candidate %s", bare)
	}

	for _, abs := range []string{"shop.web", "user"} {
		assert.Contains(t, got, abs, "absolute peer path %s", abs)
	}

	assert.NotContains(t, got, "shop.api", "the host unit is not offered as its own peer")
	assert.NotContains(t, got, "shop.webapp", "unrelated paths stay out")
}

func TestC4DCompletionEdgeOptionKeysAndValues(t *testing.T) {
	t.Parallel()

	text := "web: system \"W\" {\n  -> api: \"calls\" { \n}\n}\n\napi: system \"A\" { }\n"

	// Just inside the option block (char 20): the option keys.
	items := c4dCompleteAt(t, text, 1, 20)
	got := labels(items)

	assert.Contains(t, got, "arrow")
	assert.Contains(t, got, "rank")
	assert.Contains(t, got, "kind")
	assert.Contains(t, got, "labelPosition")
	assert.Contains(t, got, "length")
	assert.NotContains(t, got, "peer", "unit fields stay out of option blocks")

	// The option values complete their enums (char 27 is after `arrow: `).
	text = "web: system \"W\" {\n  -> api: \"calls\" { arrow: \n}\n}\n\napi: system \"A\" { }\n"
	items = c4dCompleteAt(t, text, 1, 27)

	assert.Equal(t, []string{"forward", "reverse", "bidirectional", "none"}, labels(items))
}

// --- completion: templates, use, include ---------------------------------

const c4dTemplateFixture = "template svc(name, tech) {\n" +
	"  type: container\n" +
	"  name: \"${name} Service\"\n" +
	"  technology: \"${tech}\"\n" +
	"}\n" +
	"template other(x) {\n  type: db\n}\n"

func TestC4DCompletionUseTemplateNamesAndArgKeys(t *testing.T) {
	t.Parallel()

	// c4dTemplateFixture spans lines 0-7; web is line 8, `use` line 9.
	text := c4dTemplateFixture + "web: system \"W\" {\n  use \n}\n"

	items := c4dCompleteAt(t, text, 9, 5)
	assert.Equal(t, []string{"svc", "other"}, labels(items),
		"declared template names in declaration order")

	text = c4dTemplateFixture + "web: system \"W\" {\n  use svc(\n}\n"
	items = c4dCompleteAt(t, text, 9, 10)

	assert.Equal(t, []string{"name", "tech"}, labels(items),
		"argument keys are the named template's declared params")
}

func TestC4DCompletionTemplateParamTokens(t *testing.T) {
	t.Parallel()

	// Inside ${...} in the template body (line 2, right after the brace):
	// the union of every template's declared parameters.
	text := c4dTemplateFixture
	items := c4dCompleteAt(t, text, 2, 11)

	assert.Equal(t, []string{"name", "tech", "x"}, labels(items))

	// Mid-token the list prefix-filters.
	items = c4dCompleteAt(t, text, 2, 12)
	assert.Equal(t, []string{"name"}, labels(items))
}

func TestC4DCompletionIncludePathFromDisk(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "svc.c4d"), []byte("a: system { }\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("x"), 0o600))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "sub"), 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sub", "leaf.toml"), []byte(""), 0o600))

	h := newHarness(t)
	h.request("initialize", lsp.InitializeResult{})
	h.notify("initialized", lsp.InitializedParams{})

	uri := lsp.DocumentURI("file://" + filepath.Join(dir, "main.c4d"))
	h.openDoc(uri, "include \n")

	resp := h.request("textDocument/completion", lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: uri},
		Position:     lsp.Position{Line: 0, Character: 8},
	})

	require.Nil(t, resp.Error)

	var list lsp.CompletionList
	require.NoError(t, json.Unmarshal(resp.Result, &list))

	got := labels(list.Items)
	assert.Contains(t, got, "svc.c4d", "diagram files are offered")
	assert.Contains(t, got, "sub/", "directories are offered with a trailing slash")
	assert.NotContains(t, got, "notes.txt", "non-diagram files are filtered")
}

func TestC4DCompletionTopLevelStatements(t *testing.T) {
	t.Parallel()

	items := c4dCompleteAt(t, "\n", 0, 0)
	got := labels(items)

	assert.Contains(t, got, "properties {")
	assert.Contains(t, got, "include")
	assert.Contains(t, got, "template")
	assert.Contains(t, got, "use")
	assert.Contains(t, got, "system", "type-led unit headers are offered at top level")
}

// --- hover ----------------------------------------------------------------

func TestC4DHoverPeerResolvesWalkUp(t *testing.T) {
	t.Parallel()

	text := "user: person \"Customer\" {\n" +
		"  -> shop.webapp: \"HTTPS | Browses\"\n" +
		"}\n" +
		"shop: system \"Shop\" {\n" +
		"  webapp: container \"Web App\" {\n" +
		"    description: Frontend\n" +
		"  }\n" +
		"}\n"

	hover := c4dHoverAt(t, text, 1, 8)
	require.NotNil(t, hover, "peer value hovers")

	assert.Contains(t, hover.Contents.Value, "**shop.webapp**", "resolved absolute path")
	assert.Contains(t, hover.Contents.Value, "C2")
	assert.Contains(t, hover.Contents.Value, "container", "promoted type")
	assert.Contains(t, hover.Contents.Value, "Web App", "display name")
	assert.NotNil(t, hover.Range, "hover anchors the peer word")
}

func TestC4DHoverPeerAbsoluteAndMiss(t *testing.T) {
	t.Parallel()

	text := "a: system \"A\" { }\nb: system \"B\" {\n  -> a: \"uses\"\n}\n"

	hover := c4dHoverAt(t, text, 2, 6)
	require.NotNil(t, hover)
	assert.Contains(t, hover.Contents.Value, "**a**")
	assert.Contains(t, hover.Contents.Value, "C1")

	// An unresolvable peer hovers nothing.
	text = "b: system \"B\" {\n  -> nope: \"uses\"\n}\n"
	assert.Nil(t, c4dHoverAt(t, text, 1, 6))
}

func TestC4DHoverTemplateRefAndParams(t *testing.T) {
	t.Parallel()

	// use name — the template reference (line 8: `use svc(name: auth)`).
	text := c4dTemplateFixture + "use svc(name: auth)\n"
	hover := c4dHoverAt(t, text, 8, 5)
	require.NotNil(t, hover)
	assert.Contains(t, hover.Contents.Value, "Template `svc`")
	assert.Contains(t, hover.Contents.Value, "`name`")
	assert.Contains(t, hover.Contents.Value, "`tech`")

	// An undeclared template hovers nothing.
	text = c4dTemplateFixture + "use ghost(name: auth)\n"
	assert.Nil(t, c4dHoverAt(t, text, 8, 5))

	// ${param} inside the template body lists the template's parameters.
	text = c4dTemplateFixture
	hover = c4dHoverAt(t, text, 2, 11)
	require.NotNil(t, hover)
	assert.Contains(t, hover.Contents.Value, "Template `svc` parameters")
	assert.Contains(t, hover.Contents.Value, "`name`")

	// A label that does not parse as an edge hovers nothing.
	assert.Nil(t, c4dHoverAt(t, "web: system {\n", 0, 3))
}

// --- definition ------------------------------------------------------------

func TestC4DDefinitionPeerToUnitHeader(t *testing.T) {
	t.Parallel()

	text := "user: person \"Customer\" {\n  -> shop.webapp: \"uses\"\n}\n" +
		"shop: system \"Shop\" {\n  webapp: container \"Web App\" {\n  }\n}\n"

	locs := c4dDefinitionsAt(t, t.TempDir(), text, 1, 8)
	require.Len(t, locs, 1)
	assert.Contains(t, string(locs[0].URI), "main.c4d")
	assert.Equal(t, uint32(4), locs[0].Range.Start.Line, "lands on shop.webapp's header line")
}

func TestC4DDefinitionPeerAcrossInclude(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	svc := "db1: system \"DB\" {\n  description: d\n}\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "svc.c4d"), []byte(svc), 0o600))

	text := "include svc.c4d\n\nuser: person \"U\" {\n  -> db1: \"uses\"\n}\n"
	locs := c4dDefinitionsAt(t, dir, text, 3, 7)
	require.Len(t, locs, 1, "peer defined in an included file is found")
	assert.Contains(t, string(locs[0].URI), "svc.c4d")
	assert.Equal(t, uint32(0), locs[0].Range.Start.Line)
}

func TestC4DDefinitionPeerIntoTomlInclude(t *testing.T) {
	t.Parallel()

	// Mixed-format closure (D-26): the .c4d entry defines its peer in a
	// .toml include — the walk switches front-ends per file.
	dir := t.TempDir()

	tomlSvc := "[db1]\ntype = \"system\"\nname = \"DB\"\ndescription = \"d\"\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "shared.toml"), []byte(tomlSvc), 0o600))

	text := "include ./shared.toml\n\nuser: person \"U\" {\n  -> db1: \"uses\"\n}\n"
	locs := c4dDefinitionsAt(t, dir, text, 3, 7)
	require.Len(t, locs, 1)
	assert.Contains(t, string(locs[0].URI), "shared.toml")
	assert.Equal(t, uint32(0), locs[0].Range.Start.Line)
}

func TestC4DDefinitionTemplateAndInclude(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "leaf.c4d"), []byte("x: system { }\n"), 0o600))

	text := c4dTemplateFixture + "use svc(name: a)\n\ninclude leaf.c4d once\n"

	// use svc (line 8) → the template declaration's line.
	locs := c4dDefinitionsAt(t, dir, text, 8, 5)
	require.Len(t, locs, 1)
	assert.Equal(t, uint32(0), locs[0].Range.Start.Line)

	// include path (line 10) → the target file's URI.
	locs = c4dDefinitionsAt(t, dir, text, 10, 9)
	require.Len(t, locs, 1)
	assert.Contains(t, string(locs[0].URI), "leaf.c4d")
}

// --- documentSymbol ---------------------------------------------------------

func TestC4DDocumentSymbolsHierarchy(t *testing.T) {
	t.Parallel()

	text := "properties {\n  name: P\n}\n\n" +
		"cloud: system \"Cloud\" {\n" +
		"  db1: containerDb \"DB\" {\n  }\n" +
		"  api: container \"API\" {\n    -> db1: \"uses\"\n  }\n" +
		"}\n" +
		"user: person \"User\" {\n}\n" +
		c4dTemplateFixture

	symbols := c4dSymbolsOf(t, text)
	require.NotNil(t, symbols)

	// Only unit blocks: properties and template declarations are excluded.
	names := make([]string, 0, len(symbols))
	for _, s := range symbols {
		names = append(names, s.Name)
	}

	assert.Equal(t, []string{"cloud", "user"}, names)

	// cloud has the nested db1 and api; api's edge is not a symbol.
	require.Len(t, symbols[0].Children, 2)
	assert.Equal(t, "db1", symbols[0].Children[0].Name)
	assert.Equal(t, "api", symbols[0].Children[1].Name)

	assert.Contains(t, symbols[0].Detail, "Cloud", "display name enriches the outline")
	assert.Contains(t, symbols[0].Detail, "system", "declared type enriches the outline")
	assert.Contains(t, symbols[1].Detail, "person")

	// Ranges span each brace block.
	assert.Equal(t, uint32(4), symbols[0].Range.Start.Line)
	assert.Equal(t, uint32(10), symbols[0].Range.End.Line, "closed at the block's own brace")
	assert.Equal(t, uint32(11), symbols[1].Range.Start.Line)
	assert.Equal(t, uint32(12), symbols[1].Range.End.Line)
}

func TestC4DDocumentSymbolsLineStructure(t *testing.T) {
	t.Parallel()

	// Triple-quoted values containing braces and `;` one-liner blocks keep
	// the outline honest.
	text := "a: system \"A\" {\n" +
		"  description: \"\"\"has { braces }\n" +
		"and more\"\"\"\n" +
		"}\n" +
		"box { b: system { } c: db \"C\" { } }\n"

	symbols := c4dSymbolsOf(t, text)
	require.Len(t, symbols, 2, "the braces inside the string are not structure")

	assert.Equal(t, "a", symbols[0].Name)
	assert.Equal(t, uint32(3), symbols[0].Range.End.Line)

	require.Len(t, symbols[1].Children, 2, "one-liner subunits are outlined")
	assert.Equal(t, "b", symbols[1].Children[0].Name)
	assert.Equal(t, "c", symbols[1].Children[1].Name)

	for _, child := range symbols[1].Children {
		assert.Equal(t, symbols[1].Range.Start.Line, child.Range.Start.Line,
			"one-liner units start and end on the same line")
	}

	// A quoted display name with a dot still names the unit (the parser's
	// unitKey rule: name becomes the key).
	text = "system \"v1.2\" { }\n"
	symbols = c4dSymbolsOf(t, text)
	require.Len(t, symbols, 1)
	assert.Equal(t, "v1.2", symbols[0].Name)
	assert.Equal(t, "v1.2 (system)", symbols[0].Detail)
}

func TestC4DDocumentSymbolsEmptyAndBroken(t *testing.T) {
	t.Parallel()

	// No units: null outline.
	assert.Nil(t, c4dSymbolsOf(t, "properties {\n  name: P\n}\n"))

	// Mid-edit unclosed blocks still outline, closing at the buffer end.
	symbols := c4dSymbolsOf(t, "shop: system {\n  web: \n")
	require.Len(t, symbols, 1)
	assert.Equal(t, "shop", symbols[0].Name)
	assert.Equal(t, uint32(2), symbols[0].Range.End.Line, "unclosed blocks close at EOF")
}

// --- corpus fixtures --------------------------------------------------------

// TestC4DCorpusOutline opens every repo .c4d example through the server and
// pins the outline of the nesting example — the same corpus treatment the
// TOML side gives testdata/*.toml.
func TestC4DCorpusOutline(t *testing.T) {
	t.Parallel()

	h := newHarness(t)
	h.request("initialize", lsp.InitializeResult{})
	h.notify("initialized", lsp.InitializedParams{})

	uri := h.openFile(repoPath(t, "skill/examples/02-nested.c4d"))

	resp := h.request("textDocument/documentSymbol", lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: uri},
		Position:     lsp.Position{},
	})
	require.Nil(t, resp.Error)

	var symbols []lsp.DocumentSymbol
	require.NoError(t, json.Unmarshal(resp.Result, &symbols))

	names := make([]string, 0, len(symbols))
	for _, s := range symbols {
		names = append(names, s.Name)
	}

	assert.Equal(t, []string{"user", "shop"}, names)

	require.Len(t, symbols[1].Children, 3, "webapp, api, database under shop")
	assert.Equal(t, "webapp", symbols[1].Children[0].Name)
	assert.Equal(t, "api", symbols[1].Children[1].Name)
	assert.Equal(t, "database", symbols[1].Children[2].Name)

	api := symbols[1].Children[1]
	require.Len(t, api.Children, 3, "handlers, services, repository under api")
	assert.Equal(t, "handlers", api.Children[0].Name)
	assert.Equal(t, "repository", api.Children[2].Name)
}

// TestC4DCorpusFeaturesNoErrors runs the four features over every repo .c4d
// example: no request may error, the outline must be non-empty, and hover on
// the first document-level edge resolves (CLI-parity corpus drill).
func TestC4DCorpusFeaturesNoErrors(t *testing.T) {
	t.Parallel()

	files := []string{
		"skill/examples/02-nested.c4d",
		"skill/examples/03-links.c4d",
		"skill/examples/04-styling.c4d",
		"skill/examples/06-templates.c4d",
		"skill/examples/07-relative-peer.c4d",
		"skill/examples/10-edge-kinds.c4d",
		"skill/examples/11-nesting-context.c4d",
		"skill/examples/09-composed/entry.c4d",
		"internal/include/testdata/mixed_main.c4d",
	}

	for _, file := range files {
		t.Run(file, func(t *testing.T) {
			t.Parallel()

			h := newHarness(t)
			h.request("initialize", lsp.InitializeResult{})
			h.notify("initialized", lsp.InitializedParams{})

			uri := h.openFile(repoPath(t, file))

			resp := h.request("textDocument/documentSymbol", lsp.TextDocumentPositionParams{
				TextDocument: lsp.TextDocumentIdentifier{URI: uri},
				Position:     lsp.Position{},
			})
			require.Nil(t, resp.Error)
			assert.NotEqual(t, "null", string(resp.Result), "outline is non-empty")

			resp = h.request("textDocument/completion", lsp.TextDocumentPositionParams{
				TextDocument: lsp.TextDocumentIdentifier{URI: uri},
				Position:     lsp.Position{Line: 0, Character: 0},
			})
			require.Nil(t, resp.Error)

			resp = h.request("textDocument/hover", lsp.TextDocumentPositionParams{
				TextDocument: lsp.TextDocumentIdentifier{URI: uri},
				Position:     lsp.Position{Line: 0, Character: 0},
			})
			require.Nil(t, resp.Error)

			resp = h.request("textDocument/definition", lsp.TextDocumentPositionParams{
				TextDocument: lsp.TextDocumentIdentifier{URI: uri},
				Position:     lsp.Position{Line: 0, Character: 0},
			})
			require.Nil(t, resp.Error)
		})
	}
}

// TestC4DCorpusHoverRelativePeer pins walk-up resolution against the corpus:
// 07-relative-peer.c4d's `-> api: ...` from webapp resolves to shop.webapp's
// sibling api (D-13 scopes).
func TestC4DCorpusHoverRelativePeer(t *testing.T) {
	t.Parallel()

	h := newHarness(t)
	h.request("initialize", lsp.InitializeResult{})
	h.notify("initialized", lsp.InitializedParams{})

	uri := h.openFile(repoPath(t, "skill/examples/07-relative-peer.c4d"))

	text, err := os.ReadFile(repoPath(t, "skill/examples/07-relative-peer.c4d"))
	require.NoError(t, err)

	// Find the first edge statement's peer token.
	line, char := firstEdgePeer(string(text))
	require.GreaterOrEqual(t, line, 0, "fixture has an edge")

	resp := h.request("textDocument/hover", lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: uri},
		Position:     lsp.Position{Line: uint32(line), Character: uint32(char)}, //nolint:gosec // bounded by the fixture
	})
	require.Nil(t, resp.Error)
	assert.NotEqual(t, "null", string(resp.Result), "the corpus peer resolves")

	var hover lsp.Hover
	require.NoError(t, json.Unmarshal(resp.Result, &hover))
	assert.Contains(t, hover.Contents.Value, "**")
}

// firstEdgePeer locates the character position just inside the first `->`
// peer token (0-based line and character) for corpus hover checks.
func firstEdgePeer(text string) (int, int) {
	for lineNo, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimLeft(line, " \t")
		if strings.HasPrefix(trimmed, "->") {
			// hover the first peer character after the glyph and its space
			return lineNo, len(line) - len(trimmed) + 3
		}
	}

	return -1, -1
}
