// server_test.go drives the transport-agnostic server core with raw JSON-RPC
// messages (issue #32 conformance suite): one fixture per lifecycle step, plus
// golden parity tests pinning every diagnostics class to the exact output
// `c4drill <file>` prints (captured from the CLI on the testdata corpus).

package lsp_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/Djarvur/c4drill/internal/lsp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// harness drives the server core in-proc (the GUI app's in-memory transport
// shape): requests via Handle, notifications recorded in arrival order.
type harness struct {
	t    *testing.T
	srv  *lsp.Server
	next int

	mu   sync.Mutex
	sent []lsp.Message
}

func newHarness(t *testing.T) *harness {
	t.Helper()

	h := &harness{t: t, srv: lsp.NewServer()}
	h.srv.SetNotifier(func(method string, params any) {
		raw, err := json.Marshal(params)
		require.NoError(t, err)

		h.mu.Lock()
		defer h.mu.Unlock()

		h.sent = append(h.sent, lsp.Message{
			JSONRPC: "2.0",
			Method:  method,
			Params:  raw,
		})
	})

	return h
}

// request sends one JSON-RPC request and returns the raw response message.
func (h *harness) request(method string, params any) *lsp.Message {
	h.t.Helper()

	h.mu.Lock()
	h.next++
	id := h.next
	h.mu.Unlock()

	return h.srv.Handle(context.Background(), h.message(&id, method, params))
}

// notify sends one JSON-RPC notification (no id, no response).
func (h *harness) notify(method string, params any) {
	h.t.Helper()
	h.srv.Handle(context.Background(), h.message(nil, method, params))
}

// message builds a raw JSON-RPC envelope with the params marshaled exactly
// as a client would put them on the wire.
func (h *harness) message(id *int, method string, params any) *lsp.Message {
	h.t.Helper()

	raw, err := json.Marshal(params)
	require.NoError(h.t, err)

	msg := &lsp.Message{JSONRPC: "2.0", Method: method, Params: raw}

	if id != nil {
		rawID := lsp.ID(strconv.Itoa(*id))
		msg.ID = &rawID
	}

	return msg
}

// published drains the recorded notifications into publishDiagnostics params.
func (h *harness) published() []lsp.PublishDiagnosticsParams {
	h.t.Helper()

	h.mu.Lock()
	defer h.mu.Unlock()

	out := make([]lsp.PublishDiagnosticsParams, 0, len(h.sent))
	for i := range h.sent {
		require.Equal(h.t, "textDocument/publishDiagnostics", h.sent[i].Method)

		var p lsp.PublishDiagnosticsParams
		require.NoError(h.t, json.Unmarshal(h.sent[i].Params, &p))

		out = append(out, p)
	}

	h.sent = nil

	return out
}

// publishedFor returns the LAST diagnostics publication for uri.
func (h *harness) publishedFor(uri lsp.DocumentURI) lsp.PublishDiagnosticsParams {
	h.t.Helper()

	pubs := h.published()

	var last lsp.PublishDiagnosticsParams

	found := false

	for _, p := range pubs {
		if p.URI == uri {
			last, found = p, true
		}
	}

	require.True(h.t, found, "no publishDiagnostics for %s in %d publications", uri, len(pubs))

	return last
}

// messagesOf extracts sorted diagnostic messages (validation errors arrive in
// map order; the CLI's own printed order is equally arbitrary for them).
func messagesOf(p lsp.PublishDiagnosticsParams) []string {
	out := make([]string, 0, len(p.Diagnostics))
	for _, d := range p.Diagnostics {
		out = append(out, d.Message)
	}

	slices.Sort(out)

	return out
}

// uriFor builds a file:// URI over a path.
func uriFor(t *testing.T, path string) lsp.DocumentURI {
	t.Helper()

	abs, err := filepath.Abs(path)
	require.NoError(t, err)

	return lsp.DocumentURI("file://" + abs)
}

// repoPath resolves a repo-relative fixture path from this package's dir.
func repoPath(t *testing.T, rel string) string {
	t.Helper()

	abs, err := filepath.Abs(filepath.Join("..", "..", rel))
	require.NoError(t, err)

	return abs
}

// openDoc didOpens text at uri with version 1.
func (h *harness) openDoc(uri lsp.DocumentURI, text string) {
	h.t.Helper()

	h.notify("textDocument/didOpen", lsp.DidOpenTextDocumentParams{
		TextDocument: lsp.TextDocumentItem{
			URI: uri, LanguageID: "toml", Version: 1, Text: text,
		},
	})
}

func (h *harness) openFile(path string) lsp.DocumentURI {
	h.t.Helper()

	data, err := os.ReadFile(path) //nolint:gosec // test reads repo fixtures by path
	require.NoError(h.t, err)

	uri := uriFor(h.t, path)
	h.openDoc(uri, string(data))

	return uri
}

// --- lifecycle fixtures -------------------------------------------------

func TestInitializeAdvertisesCapabilities(t *testing.T) {
	t.Parallel()

	h := newHarness(t)

	resp := h.request("initialize", lsp.InitializeResult{}) // params content is opaque

	require.Nil(t, resp.Error, "initialize must succeed: %v", resp.Error)

	var result lsp.InitializeResult
	require.NoError(t, json.Unmarshal(resp.Result, &result))

	require.NotNil(t, result.Capabilities.TextDocumentSync)
	assert.True(t, result.Capabilities.TextDocumentSync.OpenClose)
	assert.Equal(t, lsp.SyncFull, result.Capabilities.TextDocumentSync.Change)
	assert.Equal(t, "c4drill", result.ServerInfo.Name)
}

func TestRequestBeforeInitializeFails(t *testing.T) {
	t.Parallel()

	h := newHarness(t)

	resp := h.request("shutdown", nil)
	require.NotNil(t, resp.Error)
	assert.Equal(t, -32002, resp.Error.Code)
}

func TestNotificationBeforeInitializeIsDropped(t *testing.T) {
	t.Parallel()

	h := newHarness(t)

	h.openFile(repoPath(t, "testdata/links.toml"))
	assert.Empty(t, h.published(), "pre-initialize notifications are dropped")
}

func TestUnknownRequestMethodFails(t *testing.T) {
	t.Parallel()

	h := newHarness(t)
	h.request("initialize", lsp.InitializeResult{})
	h.notify("initialized", lsp.InitializedParams{})

	resp := h.request("textDocument/unknown", nil)
	require.NotNil(t, resp.Error)
	assert.Equal(t, -32601, resp.Error.Code)
}

func TestShutdownThenExitLifecycle(t *testing.T) {
	t.Parallel()

	h := newHarness(t)
	h.request("initialize", lsp.InitializeResult{})
	h.notify("initialized", lsp.InitializedParams{})

	resp := h.request("shutdown", nil)
	require.Nil(t, resp.Error)
	assert.JSONEq(t, "null", string(resp.Result), "shutdown result is null")

	resp = h.request("shutdown", nil) // any request after shutdown
	require.NotNil(t, resp.Error)
	assert.Equal(t, -32600, resp.Error.Code)

	assert.False(t, h.srv.Exited())
	h.notify("exit", nil)
	assert.True(t, h.srv.Exited())
}

// --- diagnostics fixtures ------------------------------------------------

func TestDidOpenPublishesEmptyDiagnosticsForCleanModel(t *testing.T) {
	t.Parallel()

	h := newHarness(t)
	h.request("initialize", lsp.InitializeResult{})
	h.notify("initialized", lsp.InitializedParams{})

	uri := h.openFile(repoPath(t, "testdata/links.toml"))

	pub := h.publishedFor(uri)
	assert.Empty(t, pub.Diagnostics, "links.toml is clean per the CLI")
	require.NotNil(t, pub.Version)
	assert.Equal(t, 1, *pub.Version)
}

// golden parity corpus: file → the exact lines `c4drill <file>` prints
// (captured from the CLI; validation errors sorted because the validator's
// map iteration order is not deterministic across rules' unit walks).
var messagesGolden = map[string][]string{ //nolint:gochecknoglobals // golden table
	"testdata/invalid_links.toml": {
		`error: unit "parent" has subunits and cannot be linked to directly in other`,
		`error: unit "parent" has subunits and cannot have direct links in parent`,
		`error: unit "parent.child" has no incoming or outgoing links in parent.child`,
	},
	"testdata/invalid_subunits.toml": {
		`error: unit "person_with_subunits" has type person which cannot have subunits in person_with_subunits`,
		`error: unit "person_with_subunits.child" has no incoming or outgoing links in person_with_subunits.child`,
	},
	"testdata/template_forward_ref.toml": {
		`error: unit "auth" has no incoming or outgoing links in auth`,
		`error: unit "auth" has type container which is not allowed at top level (C1 types only) in auth`,
	},
	"testdata/nested.toml": {
		`error: unit "externals" has no incoming or outgoing links in externals`,
		`error: unit "mainapp.api.handler" has no incoming or outgoing links in mainapp.api.handler`,
		`error: unit "mainapp.db" has no incoming or outgoing links in mainapp.db`,
	},
}

func TestValidatorParityPerCorpusFile(t *testing.T) {
	t.Parallel()

	for file, want := range messagesGolden {
		t.Run(file, func(t *testing.T) {
			t.Parallel()

			h := newHarness(t)
			h.request("initialize", lsp.InitializeResult{})
			h.notify("initialized", lsp.InitializedParams{})

			uri := h.openFile(repoPath(t, file))

			pub := h.publishedFor(uri)
			assert.Equal(t, want, messagesOf(pub))

			for _, d := range pub.Diagnostics {
				assert.Equal(t, 1, d.Severity)
				assert.Equal(t, "c4drill", d.Source)
			}
		})
	}
}

func TestPipelineStageParityPerCorpusFile(t *testing.T) {
	t.Parallel()

	// file → the single stage-failure line `c4drill <file>` prints. For
	// invalid_references.toml the CLI's line is one of TWO valid strings:
	// peer.Resolve fails fast on whichever of the two bad units the model
	// map walk hits first, so the golden accepts both CLI-printable lines.
	golden := map[string][]string{
		"testdata/invalid_references.toml": {
			`resolve peers: cannot resolve peer "undefined_db" from unit "user"`,
			`resolve peers: cannot resolve peer "missing_system" from unit "app"`,
		},
		"testdata/template_duplicate_path.toml": {"expand: template expand: duplicate unit path at [[use]] #2 " +
			`(name="auth", template="svc"): path "auth" is already claimed by instantiation [[use]] #1 ` +
			`(name="auth", template="svc")`},
		"testdata/template_missing_param.toml": {"expand: template expand: missing parameter at [[use]] #1 " +
			`(name="auth", template="svc"): template "svc" requires parameter "tech" which is not supplied`},
	}

	for file, want := range golden {
		t.Run(file, func(t *testing.T) {
			t.Parallel()

			h := newHarness(t)
			h.request("initialize", lsp.InitializeResult{})
			h.notify("initialized", lsp.InitializedParams{})

			uri := h.openFile(repoPath(t, file))

			pub := h.publishedFor(uri)
			require.Len(t, pub.Diagnostics, 1, "stage failures produce exactly one diagnostic")
			assert.Contains(t, want, pub.Diagnostics[0].Message)
			assert.Equal(t, uint32(0), pub.Diagnostics[0].Range.Start.Line,
				"stage errors carry no CLI line number")
		})
	}
}

func TestParseErrorParityWithLine(t *testing.T) {
	t.Parallel()

	h := newHarness(t)
	h.request("initialize", lsp.InitializeResult{})
	h.notify("initialized", lsp.InitializedParams{})

	// A second-pass decode error carries a real line: the CLI prints
	// "parse: parse error at line 4: toml: table a already exists" for this
	// text (the first unstable pass also reports header problems at line 1).
	text := "[a]\nx = 1\n\n[a]\ny = 2\n"
	uri := lsp.DocumentURI("file:///ws/broken.toml")
	h.openDoc(uri, text)

	pub := h.publishedFor(uri)
	require.Len(t, pub.Diagnostics, 1)

	d := pub.Diagnostics[0]
	assert.Equal(t, "parse: parse error at line 4: toml: table a already exists", d.Message)
	assert.Equal(t, uint32(3), d.Range.Start.Line, "1-based CLI line 4 → 0-based LSP line 3")
}

func TestC4DParseErrorParity(t *testing.T) {
	t.Parallel()

	h := newHarness(t)
	h.request("initialize", lsp.InitializeResult{})
	h.notify("initialized", lsp.InitializedParams{})

	// C4D front-end via extension dispatch; the CLI embeds the file path in
	// pigeon's message, so the diagnostic must carry the same path.
	text := "sysetm web { }\n"
	path := "/ws/broken.c4d"
	uri := lsp.DocumentURI("file://" + path)
	h.openDoc(uri, text)

	pub := h.publishedFor(uri)
	require.Len(t, pub.Diagnostics, 1)

	d := pub.Diagnostics[0]
	assert.True(t, strings.HasPrefix(d.Message, "parse: parse error: "+path+":1:"),
		"c4d parse error names the file and line, got: %s", d.Message)
	assert.Contains(t, d.Message, "no match found",
		"expected pigeon no-match text, got: %s", d.Message)
}

func TestUnknownExtensionFailsClosed(t *testing.T) {
	t.Parallel()

	h := newHarness(t)
	h.request("initialize", lsp.InitializeResult{})
	h.notify("initialized", lsp.InitializedParams{})

	uri := lsp.DocumentURI("file:///ws/model.xyz")
	h.openDoc(uri, "anything")

	pub := h.publishedFor(uri)
	require.Len(t, pub.Diagnostics, 1)
	assert.Equal(t, `parse: unsupported input extension ".xyz" (accepted: .toml, .c4d)`,
		pub.Diagnostics[0].Message)
}

// twoLinkedUnits is a minimal clean model; twoLinkedUnitsBrokenPeer corrupts
// the link so the resolve-peers stage fails.
const (
	twoLinkedUnits = "[a]\ntype = \"system\"\nname = \"A\"\ndescription = \"d\"\n\n" +
		"[b]\ntype = \"system\"\nname = \"B\"\ndescription = \"d\"\n\n" +
		"[[a.link]]\npeer = \"b\"\ndescription = \"uses\"\n"

	twoLinkedUnitsBrokenPeer = "[a]\ntype = \"system\"\nname = \"A\"\ndescription = \"d\"\n\n" +
		"[b]\ntype = \"system\"\nname = \"B\"\ndescription = \"d\"\n\n" +
		"[[a.link]]\npeer = \"nope\"\ndescription = \"uses\"\n"
)

func TestDidChangeRepublishesWithVersion(t *testing.T) {
	t.Parallel()

	h := newHarness(t)
	h.request("initialize", lsp.InitializeResult{})
	h.notify("initialized", lsp.InitializedParams{})

	uri := lsp.DocumentURI("file:///ws/model.toml")
	h.openDoc(uri, twoLinkedUnits)

	h.notify("textDocument/didChange", lsp.DidChangeTextDocumentParams{
		TextDocument: lsp.VersionedTextDocumentIdentifier{URI: uri, Version: 7},
		ContentChanges: []lsp.TextDocumentContentChangeEvent{
			// Corrupt the peer: resolve-peers stage fails, one diagnostic.
			{Text: twoLinkedUnitsBrokenPeer},
		},
	})

	pub := h.publishedFor(uri)
	require.Len(t, pub.Diagnostics, 1)
	require.NotNil(t, pub.Version)
	assert.Equal(t, 7, *pub.Version)
	assert.Contains(t, pub.Diagnostics[0].Message, `cannot resolve peer "nope"`)
}

func TestDidCloseClearsDiagnostics(t *testing.T) {
	t.Parallel()

	h := newHarness(t)
	h.request("initialize", lsp.InitializeResult{})
	h.notify("initialized", lsp.InitializedParams{})

	uri := h.openFile(repoPath(t, "testdata/invalid_links.toml"))
	require.NotEmpty(t, h.publishedFor(uri).Diagnostics)

	h.notify("textDocument/didClose", lsp.DidCloseTextDocumentParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: uri},
	})

	pub := h.publishedFor(uri)
	assert.Empty(t, pub.Diagnostics, "didClose clears published diagnostics")
	assert.Nil(t, pub.Version, "the clear carries no document version")
}
