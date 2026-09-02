// crossfile_test.go exercises the include-graph workspace semantics (issue #32
// acceptance 6): editing an included file republishes diagnostics for the
// including documents, through open-buffer edits and watched-file events
// alike, with include paths resolving relative to the including file (INC-02).

package lsp_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Djarvur/c4drill/internal/lsp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// validSvc defines two linked units — clean on its own and when merged.
const validSvc = `[db1]
type = "system"
name = "DB"
description = "svc store"

[api]
type = "system"
name = "API"
description = "api"

[[api.link]]
peer = "db1"
description = "uses"
`

// brokenSvc fails the TOML parse mid-header.
const brokenSvc = `[db1
type = "system"
`

// mainIncludingSvc is a clean entry including svc.toml.
const mainIncludingSvc = `[[include]]
path = "svc.toml"

[user]
type = "person"
name = "U"
description = "u"

[[user.link]]
peer = "api"
description = "uses"
`

// nestedMain includes sub/graph.toml; user links the transitively-defined svc.
const nestedMain = `[[include]]
path = "sub/graph.toml"

[user]
type = "person"
name = "U"
description = "u"

[[user.link]]
peer = "svc"
description = "uses"
`

// nestedGraph includes leaf.toml from ITS OWN directory (INC-02).
const nestedGraph = `[[include]]
path = "leaf.toml"

[svc]
type = "system"
name = "S"
description = "s"

[[svc.link]]
peer = "api"
description = "uses"
`

// cycleEntry and cycleLeaf start clean: entry includes leaf, leaf defines x
// linked from entry's user.
const (
	cycleEntry = `[[include]]
path = "b.toml"

[user]
type = "person"
name = "U"
description = "u"

[[user.link]]
peer = "x"
description = "uses"
`

	cycleLeaf = `[x]
type = "system"
name = "X"
description = "x"
`

	// cycleLeafWithLoop re-declares the entry include: including a.toml from
	// b.toml closes the cycle the CLI reports as fatal.
	cycleLeafWithLoop = `[[include]]
path = "a.toml"

[x]
type = "system"
name = "X"
description = "x"
`
)

// writeFixture drops name→content files into a fresh temp dir and returns it.
func writeFixture(t *testing.T, files map[string]string) string {
	t.Helper()

	dir := t.TempDir()
	for name, content := range files {
		p := filepath.Join(dir, name)
		require.NoError(t, os.MkdirAll(filepath.Dir(p), 0o750))
		require.NoError(t, os.WriteFile(p, []byte(content), 0o600))
	}

	return dir
}

// TestIncludedBufferEditRepublishesIncluder is the acceptance-6 happy path:
// an error in an included buffer surfaces on the INCLUDING document, and
// fixing the buffer clears it — without any disk write.
func TestIncludedBufferEditRepublishesIncluder(t *testing.T) {
	t.Parallel()

	dir := writeFixture(t, map[string]string{
		"main.toml": mainIncludingSvc,
		"svc.toml":  validSvc,
	})

	h := newHarness(t)
	h.request("initialize", lsp.InitializeResult{})
	h.notify("initialized", lsp.InitializedParams{})

	mainURI := h.openFile(filepath.Join(dir, "main.toml"))
	require.Empty(t, messagesOf(h.publishedFor(mainURI)), "fixture starts clean")

	svcURI := h.openFile(filepath.Join(dir, "svc.toml"))
	require.Empty(t, messagesOf(h.publishedFor(svcURI)))

	// Break the included BUFFER (disk stays valid): the includer must see it
	// through the open-buffer overlay.
	h.notify("textDocument/didChange", lsp.DidChangeTextDocumentParams{
		TextDocument:   lsp.VersionedTextDocumentIdentifier{URI: svcURI, Version: 2},
		ContentChanges: []lsp.TextDocumentContentChangeEvent{{Text: brokenSvc}},
	})

	mainPub := h.publishedFor(mainURI)
	require.Len(t, mainPub.Diagnostics, 1)
	assert.Contains(t, mainPub.Diagnostics[0].Message, "include not found: svc.toml")

	// Fix the buffer: includer clears without any disk change.
	h.notify("textDocument/didChange", lsp.DidChangeTextDocumentParams{
		TextDocument:   lsp.VersionedTextDocumentIdentifier{URI: svcURI, Version: 3},
		ContentChanges: []lsp.TextDocumentContentChangeEvent{{Text: validSvc}},
	})

	assert.Empty(t, messagesOf(h.publishedFor(mainURI)))
}

// TestWatchedFileChangeRepublishesIncluder covers the on-disk editing path:
// an included file changed outside the editor (reported via
// workspace/didChangeWatchedFiles) republishes its includers.
func TestWatchedFileChangeRepublishesIncluder(t *testing.T) {
	t.Parallel()

	dir := writeFixture(t, map[string]string{
		"main.toml": mainIncludingSvc,
		"svc.toml":  validSvc,
	})

	h := newHarness(t)
	h.request("initialize", lsp.InitializeResult{})
	h.notify("initialized", lsp.InitializedParams{})

	mainURI := h.openFile(filepath.Join(dir, "main.toml"))
	require.Empty(t, messagesOf(h.publishedFor(mainURI)))

	// Corrupt svc.toml ON DISK; the client reports the watched change.
	svcPath := filepath.Join(dir, "svc.toml")
	require.NoError(t, os.WriteFile(svcPath, []byte(brokenSvc), 0o600))

	h.notify("workspace/didChangeWatchedFiles", lsp.DidChangeWatchedFilesParams{
		Changes: []lsp.FileEvent{{URI: uriFor(t, svcPath), Type: lsp.FileChanged}},
	})

	mainPub := h.publishedFor(mainURI)
	require.Len(t, mainPub.Diagnostics, 1)
	assert.Contains(t, mainPub.Diagnostics[0].Message, "include not found: svc.toml")
}

// TestIncludeResolvesRelativeToIncludingFile pins INC-02 in the workspace:
// a nested includer resolves its own includes against ITS directory, and the
// graph walk finds dependents across the nesting.
func TestIncludeResolvesRelativeToIncludingFile(t *testing.T) {
	t.Parallel()

	dir := writeFixture(t, map[string]string{
		"main.toml":      nestedMain,
		"sub/graph.toml": nestedGraph,
		"sub/leaf.toml":  validSvc,
	})

	h := newHarness(t)
	h.request("initialize", lsp.InitializeResult{})
	h.notify("initialized", lsp.InitializedParams{})

	mainURI := h.openFile(filepath.Join(dir, "main.toml"))
	assert.Empty(t, messagesOf(h.publishedFor(mainURI)), "nested relative include resolves (INC-02)")

	// Editing the grandchild-included file must reach the root includer.
	leafURI := h.openFile(filepath.Join(dir, "sub", "leaf.toml"))
	require.Empty(t, messagesOf(h.publishedFor(leafURI)))

	h.notify("textDocument/didChange", lsp.DidChangeTextDocumentParams{
		TextDocument:   lsp.VersionedTextDocumentIdentifier{URI: leafURI, Version: 2},
		ContentChanges: []lsp.TextDocumentContentChangeEvent{{Text: brokenSvc}},
	})

	mainPub := h.publishedFor(mainURI)
	require.Len(t, mainPub.Diagnostics, 1, "root includer republished via the transitive graph")
	assert.Contains(t, mainPub.Diagnostics[0].Message, "include not found")
}

// TestIncludeCycleReachesIncluder: a cycle introduced in an included buffer
// surfaces as the CLI's cycle error on the including document.
func TestIncludeCycleReachesIncluder(t *testing.T) {
	t.Parallel()

	dir := writeFixture(t, map[string]string{
		"a.toml": cycleEntry,
		"b.toml": cycleLeaf,
	})

	h := newHarness(t)
	h.request("initialize", lsp.InitializeResult{})
	h.notify("initialized", lsp.InitializedParams{})

	mainURI := h.openFile(filepath.Join(dir, "a.toml"))
	require.Empty(t, messagesOf(h.publishedFor(mainURI)))

	// b alone has an orphan x (its link arrives only via the merge).
	bURI := h.openFile(filepath.Join(dir, "b.toml"))
	require.Len(t, messagesOf(h.publishedFor(bURI)), 1)

	// Introduce the cycle in b's buffer: b now includes a, which includes b.
	h.notify("textDocument/didChange", lsp.DidChangeTextDocumentParams{
		TextDocument:   lsp.VersionedTextDocumentIdentifier{URI: bURI, Version: 2},
		ContentChanges: []lsp.TextDocumentContentChangeEvent{{Text: cycleLeafWithLoop}},
	})

	mainPub := h.publishedFor(mainURI)
	require.Len(t, mainPub.Diagnostics, 1)
	assert.Contains(t, mainPub.Diagnostics[0].Message, "include cycle detected")
}

// TestClosedBufferRevertsIncluderToDisk: closing an included document with
// unsaved edits republishes includers against the reverted disk content.
func TestClosedBufferRevertsIncluderToDisk(t *testing.T) {
	t.Parallel()

	dir := writeFixture(t, map[string]string{
		"main.toml": mainIncludingSvc,
		"svc.toml":  validSvc,
	})

	h := newHarness(t)
	h.request("initialize", lsp.InitializeResult{})
	h.notify("initialized", lsp.InitializedParams{})

	mainURI := h.openFile(filepath.Join(dir, "main.toml"))
	svcURI := h.openFile(filepath.Join(dir, "svc.toml"))

	// Break svc's buffer, then close it without saving.
	h.notify("textDocument/didChange", lsp.DidChangeTextDocumentParams{
		TextDocument:   lsp.VersionedTextDocumentIdentifier{URI: svcURI, Version: 2},
		ContentChanges: []lsp.TextDocumentContentChangeEvent{{Text: brokenSvc}},
	})
	require.Len(t, h.publishedFor(mainURI).Diagnostics, 1)

	h.notify("textDocument/didClose", lsp.DidCloseTextDocumentParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: svcURI},
	})

	assert.Empty(t, messagesOf(h.publishedFor(mainURI)), "includer reverts to the clean disk state")
}

// TestMixedFormatIncludeGraph: a .c4d entry including a .toml file (D-26)
// publishes merged-model diagnostics, matching the CLI.
func TestMixedFormatIncludeGraph(t *testing.T) {
	t.Parallel()

	h := newHarness(t)
	h.request("initialize", lsp.InitializeResult{})
	h.notify("initialized", lsp.InitializedParams{})

	uri := h.openFile(repoPath(t, "internal/include/testdata/mixed_main.c4d"))

	pub := h.publishedFor(uri)
	assert.Equal(t, []string{
		`error: unit "entrySvc" has no incoming or outgoing links in entrySvc`,
		`error: unit "sharedDb" has no incoming or outgoing links in sharedDb`,
	}, messagesOf(pub), "C4D entry + TOML include merges at Model level exactly like the CLI")
}

// TestInitializedRegistersWatchedFiles (issue #33): the server has handled
// workspace/didChangeWatchedFiles since M1, but the LSP spec has no static
// server capability for it — clients only report once the server registers
// dynamically on initialized.
func TestInitializedRegistersWatchedFiles(t *testing.T) {
	t.Parallel()

	h := newHarness(t)
	h.withRequester()

	h.request("initialize", lsp.InitializeResult{})

	require.Empty(t, h.sentRequests(), "nothing registers before initialized")

	h.notify("initialized", lsp.InitializedParams{})

	reqs := h.sentRequests()
	require.Len(t, reqs, 1, "exactly one client/registerCapability request")

	var params lsp.RegistrationParams
	require.NoError(t, json.Unmarshal(reqs[0].Params, &params))

	require.Len(t, params.Registrations, 1)

	reg := params.Registrations[0]
	assert.Equal(t, "c4drill-watched-files", reg.ID)
	assert.Equal(t, "workspace/didChangeWatchedFiles", reg.Method)
	assert.Contains(t, string(reqs[0].Params), "**/*.toml")
	assert.Contains(t, string(reqs[0].Params), "**/*.c4d")

	// A repeated initialized does not register twice.
	h.notify("initialized", lsp.InitializedParams{})
	assert.Empty(t, h.sentRequests())

	// Without a requester (the in-proc GUI session shape) nothing registers.
	h2 := newHarness(t)
	h2.request("initialize", lsp.InitializeResult{})
	h2.notify("initialized", lsp.InitializedParams{})
	assert.Empty(t, h2.sentRequests())
}

// TestClientResponseIsDropped: a client response to the (fire-and-forget)
// registration request is tolerated — never answered with an error.
func TestClientResponseIsDropped(t *testing.T) {
	t.Parallel()

	h := newHarness(t)
	h.withRequester()

	h.request("initialize", lsp.InitializeResult{})
	h.notify("initialized", lsp.InitializedParams{})
	require.Len(t, h.sentRequests(), 1)

	id := lsp.ID("srv-1")
	assert.Nil(t, h.srv.Handle(context.Background(), &lsp.Message{
		JSONRPC: "2.0",
		ID:      &id,
		Result:  json.RawMessage("[]"),
	}), "a client response gets no response back")
}
