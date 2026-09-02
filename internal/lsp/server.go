// server.go is the transport-agnostic language-server core: LSP lifecycle,
// the open-document store, and method dispatch. Capability logic never sees
// a transport — Handle consumes one decoded JSON-RPC message and returns the
// response (or nil for notifications); server→client traffic flows through
// the notifier. The stdio transport (Serve) and in-memory test harnesses are
// thin shells around this.

package lsp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
)

// LSP lifecycle, text-document, and language-feature method names.
const (
	methodInitialize         = "initialize"
	methodCompletion         = "textDocument/completion"
	methodFormatting         = "textDocument/formatting"
	methodRenderDiagram      = "c4drill/renderDiagram"
	methodHover              = "textDocument/hover"
	methodDefinition         = "textDocument/definition"
	methodDocumentSymbol     = "textDocument/documentSymbol"
	methodSemanticTokens     = "textDocument/semanticTokens/full"
	methodInitialized        = "initialized"
	methodShutdown           = "shutdown"
	methodExit               = "exit"
	methodDidOpen            = "textDocument/didOpen"
	methodDidChange          = "textDocument/didChange"
	methodDidClose           = "textDocument/didClose"
	methodWatchedFiles       = "workspace/didChangeWatchedFiles"
	methodRegisterCapability = "client/registerCapability"
	methodPublishDiagnostics = "textDocument/publishDiagnostics"
)

// static server errors.
var (
	errNotInitialized = errors.New("server not initialized")
	errAfterShutdown  = errors.New("server is shutting down")
)

// notifier delivers a server→client notification. Set via SetNotifier; the
// default swallows traffic so an in-proc server without transport is usable.
type notifier func(method string, params any)

// document is one open text document: its URI, canonical on-disk path, and
// the current unsaved buffer contents (the source of truth for validation).
type document struct {
	URI     DocumentURI
	Path    string // canonical (filepath.Clean + filepath.Abs)
	Version int
	Text    []byte
}

// Server is the c4drill language-server core. It is safe for concurrent
// Handle calls (the stdio loop is sequential; the GUI may not be).
type Server struct {
	mu                     sync.Mutex
	docs                   map[DocumentURI]*document
	initialized            bool
	shutdown               bool
	exited                 bool
	notify                 notifier
	request                func(method string, params any) // server→client requests (dynamic registration)
	watchedFilesRegistered bool
}

// NewServer builds a server with notifications discarded.
func NewServer() *Server {
	return &Server{
		docs:   make(map[DocumentURI]*document),
		notify: func(string, any) {},
	}
}

// SetNotifier installs the sink for server→client notifications (the only
// server-initiated traffic: textDocument/publishDiagnostics). It must be
// called before the first Handle.
func (s *Server) SetNotifier(n func(method string, params any)) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.notify = n
}

// SetRequester installs the sink for server→client REQUESTS (dynamic
// capability registration). Optional: transports that cannot deliver them
// (the in-proc GUI session) simply leave it unset and the server skips
// registration. The client's response is not awaited. Must be called before
// the first Handle.
func (s *Server) SetRequester(r func(method string, params any)) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.request = r
}

// Exited reports whether the client sent the exit notification — the
// transport loop's signal to stop serving.
func (s *Server) Exited() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.exited
}

// Handle dispatches one incoming JSON-RPC message and returns the response
// message, or nil for notifications. It is the in-memory transport entry:
// the GUI app (#31) and the conformance tests drive the server through it.
func (s *Server) Handle(ctx context.Context, msg *Message) *Message {
	// A message with an id but no method is a RESPONSE to a server→client
	// request (e.g. client/registerCapability); registration is
	// fire-and-forget, so nothing is pending and the message is dropped.
	if msg.ID != nil && msg.Method == "" {
		return nil
	}

	if msg.ID == nil {
		s.handleNotification(msg)

		return nil
	}

	return s.handleRequest(ctx, msg)
}

// handleRequest processes requests (messages carrying an id); every request
// gets a response. Before initialize only `initialize` is answered; after
// shutdown everything but `exit` fails (LSP §lifecycle).
func (s *Server) handleRequest(ctx context.Context, msg *Message) *Message {
	s.mu.Lock()
	defer s.mu.Unlock()

	if msg.Method == methodInitialize {
		return s.handleInitialize(msg)
	}

	switch {
	case !s.initialized:
		return errorResponse(msg, codeServerNotInitialized, errNotInitialized.Error())
	case s.shutdown:
		return errorResponse(msg, codeInvalidRequest, errAfterShutdown.Error())
	}

	return s.dispatchRequest(ctx, msg)
}

// dispatchRequest routes an initialized-state request to its capability
// handler. Unknown methods fail with MethodNotFound.
func (s *Server) dispatchRequest(ctx context.Context, msg *Message) *Message {
	switch msg.Method {
	case methodShutdown:
		s.shutdown = true

		return okResponse(msg, json.RawMessage("null"))
	case methodCompletion:
		return s.positionRequest(msg, s.completionAt)
	case methodHover:
		return s.positionRequest(msg, s.hoverAt)
	case methodDefinition:
		return s.positionRequest(msg, s.definitionAt)
	case methodDocumentSymbol:
		return s.documentSymbolRequest(msg)
	case methodSemanticTokens:
		return s.semanticTokensRequest(msg)
	case methodFormatting:
		return s.formattingRequest(msg)
	case methodRenderDiagram:
		return s.renderDiagramRequest(ctx, msg)
	default:
		return errorResponse(msg, codeMethodNotFound, "method not found: "+msg.Method)
	}
}

// positionRequest decodes a position-bearing request, resolves its document,
// and marshals the feature result. A nil feature result answers null; an
// unknown document answers null too (nothing to report about it).
func (s *Server) positionRequest(msg *Message, feature func(*document, Position) any) *Message {
	var p TextDocumentPositionParams
	if err := json.Unmarshal(msg.Params, &p); err != nil {
		return errorResponse(msg, codeInvalidParams, "invalid params: "+err.Error())
	}

	doc, ok := s.docs[p.TextDocument.URI]
	if !ok {
		return okResponse(msg, json.RawMessage("null"))
	}

	result := feature(doc, p.Position)
	if result == nil {
		return okResponse(msg, json.RawMessage("null"))
	}

	raw, err := json.Marshal(result)
	if err != nil {
		return errorResponse(msg, codeInternalError, "marshal result: "+err.Error())
	}

	return okResponse(msg, raw)
}

// formattingRequest is textDocument/formatting's shape.
func (s *Server) formattingRequest(msg *Message) *Message {
	var p DocumentFormattingParams
	if err := json.Unmarshal(msg.Params, &p); err != nil {
		return errorResponse(msg, codeInvalidParams, "invalid params: "+err.Error())
	}

	doc, ok := s.docs[p.TextDocument.URI]
	if !ok {
		return okResponse(msg, json.RawMessage("null"))
	}

	edits := s.formatting(doc, p.Options)
	if edits == nil {
		return okResponse(msg, json.RawMessage("null")) // malformed or unsupported
	}

	if len(edits) == 0 {
		return okResponse(msg, json.RawMessage("[]")) // already formatted
	}

	raw, err := json.Marshal(edits)
	if err != nil {
		return errorResponse(msg, codeInternalError, "marshal result: "+err.Error())
	}

	return okResponse(msg, raw)
}

// renderDiagramRequest is the c4drill/renderDiagram custom method (M4): the
// live-preview primitive #27/#29/#30 and #31 build on.
func (s *Server) renderDiagramRequest(ctx context.Context, msg *Message) *Message {
	var p RenderDiagramParams
	if err := json.Unmarshal(msg.Params, &p); err != nil {
		return errorResponse(msg, codeInvalidParams, "invalid params: "+err.Error())
	}

	doc, ok := s.docs[p.TextDocument.URI]
	if !ok {
		return okResponse(msg, json.RawMessage("null"))
	}

	result := s.renderDiagram(ctx, doc, p)

	raw, err := json.Marshal(result)
	if err != nil {
		return errorResponse(msg, codeInternalError, "marshal result: "+err.Error())
	}

	return okResponse(msg, raw)
}

// documentSymbolRequest is documentSymbol's shape (document, no position).
func (s *Server) documentSymbolRequest(msg *Message) *Message {
	var p struct {
		TextDocument TextDocumentIdentifier `json:"textDocument"`
	}
	if err := json.Unmarshal(msg.Params, &p); err != nil {
		return errorResponse(msg, codeInvalidParams, "invalid params: "+err.Error())
	}

	doc, ok := s.docs[p.TextDocument.URI]
	if !ok {
		return okResponse(msg, json.RawMessage("null"))
	}

	symbols := s.documentSymbols(doc)
	if len(symbols) == 0 {
		return okResponse(msg, json.RawMessage("null"))
	}

	raw, err := json.Marshal(symbols)
	if err != nil {
		return errorResponse(msg, codeInternalError, "marshal result: "+err.Error())
	}

	return okResponse(msg, raw)
}

// semanticTokensRequest is textDocument/semanticTokens/full's shape
// (document, no position); the result is always a SemanticTokens object.
func (s *Server) semanticTokensRequest(msg *Message) *Message {
	var p struct {
		TextDocument TextDocumentIdentifier `json:"textDocument"`
	}
	if err := json.Unmarshal(msg.Params, &p); err != nil {
		return errorResponse(msg, codeInvalidParams, "invalid params: "+err.Error())
	}

	doc, ok := s.docs[p.TextDocument.URI]
	if !ok {
		return okResponse(msg, json.RawMessage(`{"data":[]}`))
	}

	raw, err := json.Marshal(s.semanticTokens(doc))
	if err != nil {
		return errorResponse(msg, codeInternalError, "marshal result: "+err.Error())
	}

	return okResponse(msg, raw)
}

// handleInitialize answers the initialize request, marking the server
// initialized and advertising the capability surface.
func (s *Server) handleInitialize(msg *Message) *Message {
	s.initialized = true

	result := InitializeResult{
		Capabilities: ServerCapabilities{
			TextDocumentSync: &TextDocumentSyncOptions{
				OpenClose: true,
				Change:    SyncFull,
				Save:      true,
			},
			CompletionProvider:         &CompletionOptions{},
			HoverProvider:              true,
			DefinitionProvider:         true,
			DocumentSymbolProvider:     true,
			DocumentFormattingProvider: true,
			SemanticTokensProvider: &SemanticTokensOptions{
				Legend: SemanticTokensLegend{TokenTypes: semTokenTypes},
				Full:   true,
			},
		},
		ServerInfo: ServerInfo{Name: "c4drill", Version: serverVersion},
	}

	raw, err := json.Marshal(result)
	if err != nil {
		return errorResponse(msg, codeInternalError,
			fmt.Sprintf("marshal initialize result: %v", err))
	}

	return okResponse(msg, raw)
}

// handleNotification processes notifications (no id): unknown or
// pre-initialize notifications are silently dropped per the LSP spec.
func (s *Server) handleNotification(msg *Message) {
	if msg.Method == methodExit {
		s.mu.Lock()
		s.exited = true
		s.mu.Unlock()

		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.initialized {
		return
	}

	switch msg.Method {
	case methodInitialized:
		// The client acknowledged initialize; negotiate the dynamic
		// capabilities (watched files) back.
		s.registerWatchedFiles()
	case methodDidOpen:
		s.onDidOpen(msg.Params)
	case methodDidChange:
		s.onDidChange(msg.Params)
	case methodDidClose:
		s.onDidClose(msg.Params)
	case methodWatchedFiles:
		s.onWatchedFiles(msg.Params)
	}
}

// registerWatchedFiles dynamically registers workspace/didChangeWatchedFiles
// (issue #33). The server has handled the notification since M1 — include
// resolution republishes on on-disk edits — but the LSP spec has no static
// server capability for it: clients only start reporting once the server
// registers. The registration is fire-and-forget: transports without a
// requester (the in-proc GUI session) skip it, and the client's response is
// tolerated to never arrive.
func (s *Server) registerWatchedFiles() {
	if s.request == nil || s.watchedFilesRegistered {
		return
	}

	s.watchedFilesRegistered = true

	s.request(methodRegisterCapability, RegistrationParams{
		Registrations: []Registration{{
			ID:     "c4drill-watched-files",
			Method: methodWatchedFiles,
			RegisterOptions: WatchedFilesRegistrationOptions{
				Watchers: []FileSystemWatcher{
					{GlobPattern: "**/*.toml"},
					{GlobPattern: "**/*.c4d"},
				},
			},
		}},
	})
}

// onDidOpen stores the document snapshot and publishes its diagnostics plus
// those of every open document that includes it.
func (s *Server) onDidOpen(raw json.RawMessage) {
	var p DidOpenTextDocumentParams
	if err := json.Unmarshal(raw, &p); err != nil {
		return // malformed notification params are dropped, never fatal
	}

	doc := &document{
		URI:     p.TextDocument.URI,
		Path:    canonicalPath(uriToPath(p.TextDocument.URI)),
		Version: p.TextDocument.Version,
		Text:    []byte(p.TextDocument.Text),
	}

	s.docs[doc.URI] = doc
	s.revalidate(doc.Path)
}

// onDidChange applies the content changes to the stored buffer and
// republishes diagnostics for the document and its dependents.
func (s *Server) onDidChange(raw json.RawMessage) {
	var p DidChangeTextDocumentParams
	if err := json.Unmarshal(raw, &p); err != nil {
		return
	}

	doc, ok := s.docs[p.TextDocument.URI]
	if !ok {
		return // didChange for an unopened document: drop (LSP §state)
	}

	doc.Version = p.TextDocument.Version
	doc.Text = applyChanges(doc.Text, p.ContentChanges)
	s.revalidate(doc.Path)
}

// onDidClose drops the document, clears its published diagnostics, and
// republishes dependents — they may have been validating against this
// document's unsaved buffer, which just reverted to disk content.
func (s *Server) onDidClose(raw json.RawMessage) {
	var p DidCloseTextDocumentParams
	if err := json.Unmarshal(raw, &p); err != nil {
		return
	}

	doc, ok := s.docs[p.TextDocument.URI]
	if !ok {
		return
	}

	delete(s.docs, p.TextDocument.URI)
	s.publish(doc.URI, nil, nil)
	s.revalidate(doc.Path)
}

// onWatchedFiles republishes dependents for every externally changed file:
// an included file edited outside the editor changes what its includers
// resolve against, and only the client can report that.
func (s *Server) onWatchedFiles(raw json.RawMessage) {
	var p DidChangeWatchedFilesParams
	if err := json.Unmarshal(raw, &p); err != nil {
		return
	}

	for _, ch := range p.Changes {
		s.revalidate(canonicalPath(uriToPath(ch.URI)))
	}
}

// applyChanges folds content changes into the buffer. Full sync sends a
// single range-less replacement; ranged edits are spliced defensively.
func applyChanges(text []byte, changes []TextDocumentContentChangeEvent) []byte {
	for _, ch := range changes {
		if ch.Range == nil {
			text = []byte(ch.Text)

			continue
		}

		text = spliceText(text, *ch.Range, ch.Text)
	}

	return text
}

// publish sends one textDocument/publishDiagnostics notification. A nil
// diagnostics slice publishes the empty array (the LSP way to clear).
func (s *Server) publish(uri DocumentURI, version *int, diags []Diagnostic) {
	if diags == nil {
		diags = []Diagnostic{}
	}

	s.notify(methodPublishDiagnostics, PublishDiagnosticsParams{
		URI:         uri,
		Version:     version,
		Diagnostics: diags,
	})
}

// okResponse builds a success response echoing the request id.
func okResponse(req *Message, result json.RawMessage) *Message {
	return &Message{JSONRPC: jsonrpcVersion, ID: req.ID, Result: result}
}

// errorResponse builds an error response echoing the request id.
func errorResponse(req *Message, code int, text string) *Message {
	return &Message{
		JSONRPC: jsonrpcVersion,
		ID:      req.ID,
		Error:   &ResponseError{Code: code, Message: text},
	}
}

// serverVersion is stamped by the CLI build via SetServerVersion; the
// package-level default keeps the in-proc server stable for tests.
//
//nolint:gochecknoglobals // build-time injection precedent: cmd/c4drill root.go `version`
var serverVersion = "dev"

// SetServerVersion overrides the advertised server version (wired from the
// CLI's build-time version variable).
func SetServerVersion(v string) {
	serverVersion = v
}

// Serve runs the server over a Content-Length framed JSON-RPC stream until
// exit, EOF, or a transport error. It is the `c4drill serve --lsp` loop:
// stdin/stdout in production, any pipe elsewhere. Notification write
// failures are tolerated — a broken pipe surfaces on the next Read.
func Serve(ctx context.Context, r io.Reader, w io.Writer) error {
	conn := NewConn(r, w)
	srv := NewServer()
	srv.SetNotifier(func(method string, params any) { _ = conn.Notify(method, params) })

	var callSeq uint64

	srv.SetRequester(func(method string, params any) {
		callSeq++
		// A failed registration write must not kill the loop — the client
		// may be gone; the next Read surfaces that.
		_ = conn.Request(method, params, callSeq)
	})

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		msg, err := conn.Read()
		switch {
		case errors.Is(err, io.EOF):
			return nil
		case errors.Is(err, errBodyParse):
			// JSON-RPC: a syntactically invalid body is answered with a
			// -32700 response (null id) and the connection stays up.
			if werr := conn.Write(errorResponse(&Message{ID: nullID()}, codeParseError, err.Error())); werr != nil {
				return fmt.Errorf("write parse-error response: %w", werr)
			}

			continue
		case err != nil:
			return fmt.Errorf("read json-rpc: %w", err)
		}

		resp := srv.Handle(ctx, msg)
		if resp != nil {
			if err := conn.Write(resp); err != nil {
				return fmt.Errorf("write response: %w", err)
			}
		}

		if srv.Exited() {
			return nil
		}
	}
}

// nullID returns the JSON null request id used on -32700 responses.
func nullID() *ID {
	id := ID("null")

	return &id
}
