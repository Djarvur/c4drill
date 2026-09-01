// semantictokens_test.go covers textDocument/semanticTokens/full for the
// TOML dialect (issue #33's deferred #32 scope): unit-type keys and their
// unit-type values, link-table header segments and field keys, and enum
// values — delta-encoded, in document order. Scoping fixtures pin the
// .c4d/unknown-extension behavior (empty data).

package lsp_test

import (
	"encoding/json"
	"testing"

	"github.com/Djarvur/c4drill/internal/lsp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// semanticTokensParams mirrors the request's wire shape.
type semanticTokensParams struct {
	TextDocument lsp.TextDocumentIdentifier `json:"textDocument"`
}

// semTokensFor requests semantic tokens for text at a .toml URI (or an
// alternate extension) and returns the raw data array.
func semTokensFor(t *testing.T, ext, text string) []uint32 {
	t.Helper()

	h := newHarness(t)
	h.request("initialize", lsp.InitializeResult{})
	h.notify("initialized", lsp.InitializedParams{})

	uri := lsp.DocumentURI("file:///ws/model." + ext)
	h.openDoc(uri, text)

	resp := h.request("textDocument/semanticTokens/full", semanticTokensParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: uri},
	})
	require.Nil(t, resp.Error, "semanticTokens must not error: %v", resp.Error)

	var result lsp.SemanticTokens
	require.NoError(t, json.Unmarshal(resp.Result, &result))

	return result.Data
}

// semDecoded expands a delta-encoded token array into absolute quads
// (line, char, length, typeIndex) for readable assertions.
func semDecoded(data []uint32) [][4]uint32 {
	var out [][4]uint32

	var line, char uint32

	for i := 0; i+4 < len(data); i += 5 {
		if data[i] == 0 {
			char += data[i+1]
		} else {
			line += data[i]
			char = data[i+1]
		}

		out = append(out, [4]uint32{line, char, data[i+2], data[i+3]})
	}

	return out
}

func TestSemanticTokensUnitTypeKeysAndValues(t *testing.T) {
	t.Parallel()

	// `type` key (property) + the unit-type value inside the quotes (class).
	text := "[cloud]\ntype = \"containerDb\"\nname = \"Cloud\"\n"

	data := semTokensFor(t, "toml", text)
	require.Len(t, data, 10, "two tokens = two quintuples")

	assert.Equal(t, [][4]uint32{
		{1, 0, 4, 0},  // type key
		{1, 8, 11, 1}, // containerDb value, inside the quotes
	}, semDecoded(data))
}

func TestSemanticTokensLinkTables(t *testing.T) {
	t.Parallel()

	// The [[cloud.link]] header segment and the link field keys are
	// property tokens; enum values inside the table are enumMembers.
	text := "[cloud]\ntype = \"system\"\n\n[[cloud.link]]\npeer = \"db\"\n" +
		"rank = \"reverse\"\n"

	data := semTokensFor(t, "toml", text)

	assert.Equal(t, [][4]uint32{
		{1, 0, 4, 0}, // type key
		{1, 8, 6, 1}, // system value
		{3, 8, 4, 0}, // link header segment
		{4, 0, 4, 0}, // peer key
		{5, 0, 4, 0}, // rank key
		{5, 8, 7, 2}, // reverse enum value
	}, semDecoded(data))
}

func TestSemanticTokensEnumValues(t *testing.T) {
	t.Parallel()

	// Enum vocabularies tokenize wherever their keys take them, including
	// [properties].
	text := "[properties]\nedges = \"spline\"\nstyle = \"dashed\"\n\n" +
		"[web]\ntype = \"person\"\nstyle = \"fancy\"\n"

	data := semTokensFor(t, "toml", text)

	assert.Equal(t, [][4]uint32{
		{1, 9, 6, 2}, // spline
		{2, 9, 6, 2}, // dashed
		{5, 0, 4, 0}, // type key in [web]
		{5, 8, 6, 1}, // person value
		// "fancy" is not a known style value: no token.
	}, semDecoded(data))
}

func TestSemanticTokensTemplateTypeAndUnknownExt(t *testing.T) {
	t.Parallel()

	// Template subtrees carry the root type key/value too.
	text := "[template.svc]\ntype = \"container\"\nparams = [\"name\"]\n"
	data := semTokensFor(t, "toml", text)

	assert.Equal(t, [][4]uint32{
		{1, 0, 4, 0}, // type key on the template root
		{1, 8, 9, 1},
	}, semDecoded(data))

	// .c4d documents have no TOML semantic tokens yet: empty, not null.
	assert.Empty(t, semTokensFor(t, "c4d", "web: system \"W\" { }\n"))

	// Unknown documents answer the empty result too.
	h := newHarness(t)
	h.request("initialize", lsp.InitializeResult{})
	h.notify("initialized", lsp.InitializedParams{})

	resp := h.request("textDocument/semanticTokens/full", semanticTokensParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: "file:///ws/never-opened.toml"},
	})
	require.Nil(t, resp.Error)
	assert.JSONEq(t, `{"data":[]}`, string(resp.Result))
}
