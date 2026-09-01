// lsp.go is the in-memory LSP client half of the GUI: it drives the shared
// c4drill server core (internal/lsp) through Handle — the same transport-
// agnostic entry the stdio clients wrap — so editor behavior (diagnostics
// wording, completion, formatting) is byte-identical to the CLI clients by
// construction (issue #31: one server, four clients).
//
// The bridge serializes requests (the server core is Handle-safe but the GUI
// issues them from one UI-driven goroutine at a time) and fans server
// notifications out through the diagnostics callback.

package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/Djarvur/c4drill/internal/lsp"
)

// Diagnostic mirrors the LSP diagnostic for the frontend (same wire shape as
// internal/lsp.Diagnostic; aliased to keep one contract).
type Diagnostic = lsp.Diagnostic

// documentSymbolParams matches the server's documentSymbol decode shape (the
// protocol type lives server-side unexported).
type documentSymbolParams struct {
	TextDocument lsp.TextDocumentIdentifier `json:"textDocument"`
}

// lspBridge is the client side of the in-memory LSP session.
type lspBridge struct {
	mu      sync.Mutex // serializes request ids and dispatch
	nextID  int
	server  *lsp.Server
	onDiags func(uri string, version *int, diags []Diagnostic)
}

// newLSPBridge starts a fresh server session (initialize handshake included)
// with diagnostics flowing to onDiags.
func newLSPBridge(onDiags func(uri string, version *int, diags []Diagnostic)) *lspBridge {
	b := &lspBridge{onDiags: onDiags}
	b.start()

	return b
}

// DidOpen pushes a document snapshot into the LSP session.
func (b *lspBridge) DidOpen(absPath, text string, version int) {
	b.notify(methodDidOpen, lsp.DidOpenTextDocumentParams{
		TextDocument: lsp.TextDocumentItem{
			URI:        pathToURI(absPath),
			LanguageID: languageIDOf(absPath),
			Version:    version,
			Text:       text,
		},
	})
}

// DidChange pushes a full-buffer update (the server's advertised sync mode).
func (b *lspBridge) DidChange(absPath, text string, version int) {
	b.notify(methodDidChange, lsp.DidChangeTextDocumentParams{
		TextDocument: lsp.VersionedTextDocumentIdentifier{
			URI:     pathToURI(absPath),
			Version: version,
		},
		ContentChanges: []lsp.TextDocumentContentChangeEvent{{Text: text}},
	})
}

// DidClose drops the document from the session.
func (b *lspBridge) DidClose(absPath string) {
	b.notify(methodDidClose, lsp.DidCloseTextDocumentParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: pathToURI(absPath)},
	})
}

// Completion requests completions at a zero-based line/character position.
func (b *lspBridge) Completion(absPath string, line, character uint32) (*lsp.CompletionList, error) {
	var result lsp.CompletionList

	err := b.call(context.Background(), methodCompletion, lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: pathToURI(absPath)},
		Position:     lsp.Position{Line: line, Character: character},
	}, &result)

	return &result, err
}

// Hover requests hover markdown at a position (nil when nothing to show).
func (b *lspBridge) Hover(absPath string, line, character uint32) (*lsp.Hover, error) {
	var result *lsp.Hover

	err := b.call(context.Background(), methodHover, lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: pathToURI(absPath)},
		Position:     lsp.Position{Line: line, Character: character},
	}, &result)

	return result, err
}

// Definition requests the definition location at a position (nil when none).
func (b *lspBridge) Definition(absPath string, line, character uint32) (*lsp.Location, error) {
	var result []*lsp.Location

	err := b.call(context.Background(), methodDefinition, lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: pathToURI(absPath)},
		Position:     lsp.Position{Line: line, Character: character},
	}, &result)
	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, nil //nolint:nilnil // "no definition at position" is a valid LSP outcome, not a failure
	}

	return result[0], nil
}

// Symbols requests the document outline.
func (b *lspBridge) Symbols(absPath string) ([]lsp.DocumentSymbol, error) {
	var result []lsp.DocumentSymbol

	err := b.call(context.Background(), methodDocumentSymbol, documentSymbolParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: pathToURI(absPath)},
	}, &result)

	return result, err
}

// Format requests a full-document format of the current buffer and returns
// the formatted text (empty string when already formatted or unsupported).
func (b *lspBridge) Format(absPath string, tabSize int, insertSpaces bool) (string, error) {
	var edits []lsp.TextEdit

	err := b.call(context.Background(), methodFormatting, lsp.DocumentFormattingParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: pathToURI(absPath)},
		Options:      lsp.FormattingOptions{TabSize: tabSize, InsertSpaces: insertSpaces},
	}, &edits)
	if err != nil {
		return "", err
	}

	return applyEdits(edits), nil
}

// RenderDiagram calls the custom c4drill/renderDiagram method — the exact
// live-preview primitive the other clients use.
func (b *lspBridge) RenderDiagram(absPath string, p lsp.RenderDiagramParams) (*lsp.RenderDiagramResult, error) {
	p.TextDocument = lsp.TextDocumentIdentifier{URI: pathToURI(absPath)}

	var result lsp.RenderDiagramResult

	err := b.call(context.Background(), methodRenderDiagram, p, &result)

	return &result, err
}

// watchedFileChanged reports an on-disk change so cross-file diagnostics
// revalidate (workspace/didChangeWatchedFiles).
func (b *lspBridge) watchedFileChanged(absPath string) {
	b.notify(methodWatchedFiles, lsp.DidChangeWatchedFilesParams{
		Changes: []lsp.FileEvent{{URI: pathToURI(absPath), Type: lsp.FileChanged}},
	})
}

// restart drops the current session and starts a new one (project switch).
func (b *lspBridge) restart() {
	b.start()
}

// start builds the server, wires the notifier, and runs the initialize
// handshake. Callers must treat the old session as dropped (restart).
func (b *lspBridge) start() {
	b.mu.Lock()
	defer b.mu.Unlock()

	srv := lsp.NewServer()
	srv.SetNotifier(b.dispatchNotification)
	b.server = srv

	if err := b.callLocked(context.Background(), methodInitialize, nil, nil); err != nil {
		// The in-process server cannot fail initialize; a failure means a
		// programming error upstream. Surface it on stderr for diagnosability.
		fmt.Fprintf(os.Stderr, "gui: lsp initialize failed: %v\n", err)
	}

	_ = b.callLocked(context.Background(), methodInitialized, lsp.InitializedParams{}, nil)
}

// dispatchNotification routes one server→client notification. It runs inside
// the server's Handle (the server lock is held there) — it must never call
// back into b.server.
func (b *lspBridge) dispatchNotification(method string, params any) {
	if method != methodPublishDiagnostics {
		return
	}

	p, ok := params.(lsp.PublishDiagnosticsParams)
	if !ok {
		return
	}

	if b.onDiags != nil {
		b.onDiags(string(p.URI), p.Version, p.Diagnostics)
	}
}

// call sends one request and decodes the response result into out (nil to
// discard it). out must be a pointer.
func (b *lspBridge) call(ctx context.Context, method string, params, out any) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.callLocked(ctx, method, params, out)
}

// callLocked is call with b.mu held (used by the handshake).
func (b *lspBridge) callLocked(ctx context.Context, method string, params, out any) error {
	var raw json.RawMessage

	if params != nil {
		enc, err := json.Marshal(params)
		if err != nil {
			return fmt.Errorf("marshal %s params: %w", method, err)
		}

		raw = enc
	}

	b.nextID++

	id := lsp.ID(json.Number(strconv.Itoa(b.nextID)))

	resp := b.server.Handle(ctx, &lsp.Message{
		JSONRPC: "2.0",
		ID:      &id,
		Method:  method,
		Params:  raw,
	})
	if resp == nil {
		return fmt.Errorf("%w: %s", errLSPNoResponse, method)
	}

	if resp.Error != nil {
		return fmt.Errorf("%w: %s: code %d: %s", errLSPRequestFailed, method, resp.Error.Code, resp.Error.Message)
	}

	if out != nil && len(resp.Result) > 0 && string(resp.Result) != "null" {
		if err := json.Unmarshal(resp.Result, out); err != nil {
			return fmt.Errorf("decode %s result: %w", method, err)
		}
	}

	return nil
}

// notify sends one client→server notification (no response).
func (b *lspBridge) notify(method string, params any) {
	b.mu.Lock()
	defer b.mu.Unlock()

	raw, err := json.Marshal(params)
	if err != nil {
		return // malformed client params are a programming error; drop
	}

	b.server.Handle(context.Background(), &lsp.Message{
		JSONRPC: "2.0",
		Method:  method,
		Params:  raw,
	})
}

// applyEdits folds a full-document edit set into a replacement text. The
// server only emits whole-document replacements, but the fold is defensive.
func applyEdits(edits []lsp.TextEdit) string {
	if len(edits) == 0 {
		return ""
	}

	// Whole-document replacement fast path (the c4drill formatter's shape).
	for _, e := range edits {
		if e.Range.Start.Line == 0 && e.Range.Start.Character == 0 &&
			e.Range.End.Line == 0 && e.Range.End.Character == 0 {
			return e.NewText
		}
	}

	return edits[len(edits)-1].NewText
}

// languageIDOf maps a file extension to an LSP language id.
func languageIDOf(path string) string {
	if strings.EqualFold(filepath.Ext(path), ".c4d") {
		return "c4d"
	}

	return "toml"
}

// uriToPath converts a file:// URI to a filesystem path (lenient fallback to
// the literal text, mirroring the server's own conversion).
func uriToPath(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "file" {
		return raw
	}

	p, err := url.PathUnescape(u.Path)
	if err != nil {
		return u.Path
	}

	if runtime.GOOS == "windows" && strings.HasPrefix(p, "/") {
		p = strings.TrimPrefix(p, "/")
	}

	return p
}

// pathToURI converts a filesystem path to a file:// document URI.
func pathToURI(path string) lsp.DocumentURI {
	u := url.URL{Scheme: "file", Path: filepath.ToSlash(path)}

	return lsp.DocumentURI(u.String())
}

// LSP method names (the server's constants are unexported).
const (
	methodInitialize         = "initialize"
	methodInitialized        = "initialized"
	methodCompletion         = "textDocument/completion"
	methodHover              = "textDocument/hover"
	methodDefinition         = "textDocument/definition"
	methodDocumentSymbol     = "textDocument/documentSymbol"
	methodFormatting         = "textDocument/formatting"
	methodRenderDiagram      = "c4drill/renderDiagram"
	methodDidOpen            = "textDocument/didOpen"
	methodDidChange          = "textDocument/didChange"
	methodDidClose           = "textDocument/didClose"
	methodWatchedFiles       = "workspace/didChangeWatchedFiles"
	methodPublishDiagnostics = "textDocument/publishDiagnostics"
)
