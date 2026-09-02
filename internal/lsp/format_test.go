// format_test.go covers textDocument/formatting (issue #32, M3): tomlfmt
// parity for .toml, canonical C4D printer parity for .c4d, the fmt safety
// gate, and the clean-document no-op.

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

// formatAt runs textDocument/formatting on text at uri and returns the raw
// result JSON (null / [] / edit array).
func formatAt(t *testing.T, uri lsp.DocumentURI, text string) json.RawMessage {
	t.Helper()

	h := newHarness(t)
	h.request("initialize", lsp.InitializeResult{})
	h.notify("initialized", lsp.InitializedParams{})

	h.openDoc(uri, text)

	resp := h.request("textDocument/formatting", lsp.DocumentFormattingParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: uri},
		Options:      lsp.FormattingOptions{TabSize: 4, InsertSpaces: true},
	})

	require.Nil(t, resp.Error)

	return resp.Result
}

func editsOf(t *testing.T, raw json.RawMessage) []lsp.TextEdit {
	t.Helper()

	require.NotEqual(t, "null", string(raw), "expected an edit array")

	var edits []lsp.TextEdit
	require.NoError(t, json.Unmarshal(raw, &edits))

	return edits
}

func TestFormattingNormalizesTOMLWhitespace(t *testing.T) {
	t.Parallel()

	// Sloppy spacing/indentation normalizes; values, order and comments stay.
	text := "# header comment\n[web]\ntype=\"system\"\n      name    =  \"Web\"\n\n\n\ndescription = \"w\"  # trailing\n"
	raw := formatAt(t, lsp.DocumentURI("file:///ws/model.toml"), text)

	edits := editsOf(t, raw)
	require.Len(t, edits, 1, "one whole-document replacement")

	want := "# header comment\n[web]\ntype = \"system\"\nname = \"Web\"\n\ndescription = \"w\" # trailing\n"
	assert.Equal(t, want, edits[0].NewText)

	assert.Equal(t, uint32(0), edits[0].Range.Start.Line)
	assert.Equal(t, uint32(8), edits[0].Range.End.Line, "spans the whole document")
}

func TestFormattingCleanTOMLIsNoOp(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile(repoPath(t, "testdata/links.toml"))
	require.NoError(t, err)

	// links.toml is clean per `c4drill fmt --check`: the result is [].
	raw := formatAt(t, lsp.DocumentURI("file:///ws/links.toml"), string(data))
	assert.Equal(t, "[]", string(raw), "clean documents produce no edits")
}

func TestFormattingTOMLKeyOrderAndCommentsPreserved(t *testing.T) {
	t.Parallel()

	// Misformatted spacing, but the b-before-a section order and the comment
	// are the author's: formatting must fix spacing only.
	text := "[b]\ntype=\"system\"\n\n[a]\ntype = \"system\"\n# keep me\n"
	want := "[b]\ntype = \"system\"\n\n[a]\ntype = \"system\"\n# keep me\n"

	raw := formatAt(t, lsp.DocumentURI("file:///ws/model.toml"), text)

	edits := editsOf(t, raw)
	require.Len(t, edits, 1)
	assert.Equal(t, want, edits[0].NewText,
		"author key order and comments survive formatting")
}

func TestFormattingMalformedTOMLIsNull(t *testing.T) {
	t.Parallel()

	raw := formatAt(t, lsp.DocumentURI("file:///ws/model.toml"), "[broken\n")
	assert.Equal(t, "null", string(raw), "malformed documents offer no edit")
}

func TestFormattingC4DUsesCanonicalPrinter(t *testing.T) {
	t.Parallel()

	// The canonical printer compacts leaf blocks (D-32/D-33) — the same
	// output `c4drill fmt` writes for .c4d files.
	text := "web: system \"Web\" {\n  description: The Web\n}\n"
	raw := formatAt(t, lsp.DocumentURI("file:///ws/model.c4d"), text)

	edits := editsOf(t, raw)
	require.Len(t, edits, 1)
	assert.Equal(t, "web: system \"Web\" { description: The Web }\n", edits[0].NewText)
}

func TestFormattingC4DSemanticGate(t *testing.T) {
	t.Parallel()

	// A grammar-valid but model-refused document: fmt refuses it through the
	// safety gate, so formatting answers null (never a corrupting edit).
	raw := formatAt(t, lsp.DocumentURI("file:///ws/model.c4d"), "sysetm web { }\n")
	assert.Equal(t, "null", string(raw))
}

func TestFormattingUnknownExtensionIsNull(t *testing.T) {
	t.Parallel()

	raw := formatAt(t, lsp.DocumentURI("file:///ws/model.xyz"), "anything")
	assert.Equal(t, "null", string(raw))
}

func TestFormattingAdvertisedByInitialize(t *testing.T) {
	t.Parallel()

	h := newHarness(t)

	resp := h.request("initialize", lsp.InitializeResult{})

	var result lsp.InitializeResult
	require.NoError(t, json.Unmarshal(resp.Result, &result))

	assert.True(t, result.Capabilities.DocumentFormattingProvider)
}

// TestFormattingMatchesFmtCheckOnCorpus pins fmt --check parity: every
// already-formatted repo fixture produces zero edits.
func TestFormattingMatchesFmtCheckOnCorpus(t *testing.T) {
	t.Parallel()

	for _, f := range []string{
		"testdata/links.toml",
		"testdata/nested.toml",
		"testdata/template_basic.toml",
	} {
		t.Run(f, func(t *testing.T) {
			t.Parallel()

			data, err := os.ReadFile(repoPath(t, f))
			require.NoError(t, err)

			raw := formatAt(t, lsp.DocumentURI("file://"+filepath.Join("/ws", filepath.Base(f))), string(data))
			assert.Equal(t, "[]", string(raw), "%s is fmt --check clean", f)
		})
	}
}
