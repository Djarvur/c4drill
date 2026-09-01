// features_test.go covers textDocument/hover, textDocument/definition, and
// textDocument/documentSymbol (issue #32 M2).

package lsp_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Djarvur/c4drill/internal/lsp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// hoverAt runs hover on text at (line, char) and returns the decoded Hover.
func hoverAt(t *testing.T, text string, line, char uint32) *lsp.Hover {
	t.Helper()

	h := newHarness(t)
	h.request("initialize", lsp.InitializeResult{})
	h.notify("initialized", lsp.InitializedParams{})

	uri := lsp.DocumentURI("file:///ws/model.toml")
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

// definitionsAt runs definition and returns decoded locations (nil on null).
func definitionsAt(t *testing.T, dir, text string, line, char uint32) []lsp.Location {
	t.Helper()

	h := newHarness(t)
	h.request("initialize", lsp.InitializeResult{})
	h.notify("initialized", lsp.InitializedParams{})

	uri := lsp.DocumentURI("file://" + filepath.Join(dir, "main.toml"))
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

func TestHoverPeerResolvesWalkUp(t *testing.T) {
	t.Parallel()

	// Hovering "db1" in [[cloud.api.link]]: bare peer resolves UP to
	// cloud.db1 (sibling scope, D-13) — and its parsed type is containerDb.
	text := `[cloud]
type = "system"
name = "Cloud"

[cloud.db1]
type = "containerDb"
name = "DB"
description = "store"

[cloud.api]
type = "container"
name = "API"

[[cloud.api.link]]
peer = "db1"
description = "uses"
`

	hover := hoverAt(t, text, 14, 9)
	require.NotNil(t, hover, "peer value hovers")

	assert.Contains(t, hover.Contents.Value, "**cloud.db1**", "resolved absolute path")
	assert.Contains(t, hover.Contents.Value, "C2")
	assert.Contains(t, hover.Contents.Value, "containerDb", "promoted type")
	assert.Contains(t, hover.Contents.Value, "DB", "display name")
	assert.NotNil(t, hover.Range, "hover anchors the peer word")
}

func TestHoverPeerAbsoluteAndMiss(t *testing.T) {
	t.Parallel()

	text := "[a]\ntype = \"system\"\n\n[b]\ntype = \"system\"\n\n[[b.link]]\npeer = \"a\"\n"

	hover := hoverAt(t, text, 7, 9)
	require.NotNil(t, hover)
	assert.Contains(t, hover.Contents.Value, "**a**")
	assert.Contains(t, hover.Contents.Value, "C1")

	// An unresolvable peer hovers nothing.
	text = "[b]\ntype = \"system\"\n\n[[b.link]]\npeer = \"nope\"\n"
	assert.Nil(t, hoverAt(t, text, 4, 9))
}

func TestHoverTemplateParamInfo(t *testing.T) {
	t.Parallel()

	// ${param} token inside a template body.
	text := "[template.svc]\nparams = [\"name\", \"tech\"]\n\ntechnology = \"${tech}\"\n"
	hover := hoverAt(t, text, 3, 16)
	require.NotNil(t, hover)
	assert.Contains(t, hover.Contents.Value, "Template `svc`")
	assert.Contains(t, hover.Contents.Value, "`name`")
	assert.Contains(t, hover.Contents.Value, "`tech`")

	// template = "svc" reference.
	text = "[template.svc]\nparams = [\"name\", \"tech\"]\n\n[[use]]\ntemplate = \"svc\"\n"
	hover = hoverAt(t, text, 4, 12)
	require.NotNil(t, hover)
	assert.Contains(t, hover.Contents.Value, "Template `svc`")

	// A peer that does not parse: no hover, no crash.
	assert.Nil(t, hoverAt(t, "[broken\n", 0, 3))
}

func TestDefinitionPeerToUnitSection(t *testing.T) {
	t.Parallel()

	text := `[cloud]
type = "system"

[cloud.db1]
type = "containerDb"

[cloud.api]
type = "container"

[[cloud.api.link]]
peer = "cloud.db1"
`
	locs := definitionsAt(t, t.TempDir(), text, 10, 7)
	require.Len(t, locs, 1)
	assert.Contains(t, string(locs[0].URI), "main.toml")
	assert.Equal(t, uint32(3), locs[0].Range.Start.Line, "lands on [cloud.db1]'s line")
}

func TestDefinitionPeerAcrossInclude(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	svc := "[db1]\ntype = \"system\"\nname = \"DB\"\ndescription = \"d\"\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "svc.toml"), []byte(svc), 0o600))

	text := "[[include]]\npath = \"svc.toml\"\n\n[user]\ntype = \"person\"\n\n[[user.link]]\npeer = \"db1\"\n"
	locs := definitionsAt(t, dir, text, 7, 7)
	require.Len(t, locs, 1, "peer defined in an included file is found")
	assert.Contains(t, string(locs[0].URI), "svc.toml")
	assert.Equal(t, uint32(0), locs[0].Range.Start.Line)
}

func TestDefinitionTemplateAndInclude(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	shared := "[x]\ntype = \"system\"\nname = \"X\"\ndescription = \"d\"\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "shared.toml"), []byte(shared), 0o600))

	text := "[template.svc]\nparams = [\"name\"]\n\n[[use]]\ntemplate = \"svc\"\n" +
		"\n[[include]]\npath = \"shared.toml\"\n"

	// template = "svc" → [template.svc] line 0.
	locs := definitionsAt(t, dir, text, 4, 12)
	require.Len(t, locs, 1)
	assert.Equal(t, uint32(0), locs[0].Range.Start.Line)

	// include path → the target file's URI.
	locs = definitionsAt(t, dir, text, 7, 7)
	require.Len(t, locs, 1)
	assert.Contains(t, string(locs[0].URI), "shared.toml")
}

func TestDocumentSymbolsHierarchy(t *testing.T) {
	t.Parallel()

	h := newHarness(t)
	h.request("initialize", lsp.InitializeResult{})
	h.notify("initialized", lsp.InitializedParams{})

	text := `[properties]
name = "P"

[cloud]
type = "system"
name = "Cloud"

[cloud.db1]
type = "containerDb"
name = "DB"

[[cloud.db1.link]]
peer = "x"

[[include]]
path = "other.toml"

[user]
type = "person"
name = "User"
`
	uri := lsp.DocumentURI("file:///ws/model.toml")
	h.openDoc(uri, text)

	resp := h.request("textDocument/documentSymbol", lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: uri},
		Position:     lsp.Position{},
	})

	require.Nil(t, resp.Error)

	var symbols []lsp.DocumentSymbol
	require.NoError(t, json.Unmarshal(resp.Result, &symbols))

	// Only unit sections: properties, include, and link tables are excluded.
	names := make([]string, 0, len(symbols))
	for _, s := range symbols {
		names = append(names, s.Name)
	}

	assert.Equal(t, []string{"cloud", "user"}, names)

	// cloud has the nested db1; link tables do not appear.
	require.Len(t, symbols[0].Children, 1)
	assert.Equal(t, "db1", symbols[0].Children[0].Name)
	require.Empty(t, symbols[0].Children[0].Children)

	assert.Contains(t, symbols[0].Detail, "Cloud", "display name enriches the outline")
	assert.Contains(t, symbols[1].Detail, "person", "type enriches the outline")

	// Ranges cover the section span.
	assert.Equal(t, uint32(3), symbols[0].Range.Start.Line)
	assert.Equal(t, uint32(17), symbols[0].Range.End.Line, "closed where [user] begins")
	assert.Equal(t, uint32(20), symbols[1].Range.End.Line, "last section closes at the EOF line")
}

func TestLanguageFeaturesAdvertiseAndScopeToToml(t *testing.T) {
	t.Parallel()

	h := newHarness(t)

	resp := h.request("initialize", lsp.InitializeResult{})

	var result lsp.InitializeResult
	require.NoError(t, json.Unmarshal(resp.Result, &result))

	assert.NotNil(t, result.Capabilities.CompletionProvider)
	assert.True(t, result.Capabilities.HoverProvider)
	assert.True(t, result.Capabilities.DefinitionProvider)
	assert.True(t, result.Capabilities.DocumentSymbolProvider)

	// A .c4d document: features scope to TOML, so they answer empty/null.
	c4dURI := lsp.DocumentURI("file:///ws/model.c4d")

	h.openDoc(c4dURI, "web: system \"Web\" { }\n")

	comp := h.request("textDocument/completion", lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: c4dURI},
		Position:     lsp.Position{Line: 0, Character: 0},
	})
	require.Nil(t, comp.Error)

	var list lsp.CompletionList
	require.NoError(t, json.Unmarshal(comp.Result, &list))
	assert.Empty(t, list.Items, ".c4d completion is intentionally empty in M2")

	hover := h.request("textDocument/hover", lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: c4dURI},
		Position:     lsp.Position{Line: 0, Character: 1},
	})
	assert.JSONEq(t, "null", string(hover.Result), ".c4d hover is null in M2")

	symbols := h.request("textDocument/documentSymbol", lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: c4dURI},
		Position:     lsp.Position{},
	})
	assert.JSONEq(t, "null", string(symbols.Result), ".c4d symbols are null in M2")
}
