// protocol.go holds the subset of LSP structures the c4drill server speaks:
// diagnostics publishing, text-document sync, and the capability surface
// advertised by initialize. Field names follow the LSP wire format
// (camelCase JSON, camelCase method strings).

package lsp

// DocumentURI is an LSP document URI (file:///...).
type DocumentURI string

// Position is a zero-based line/character offset in a document.
type Position struct {
	Line      uint32 `json:"line"`
	Character uint32 `json:"character"`
}

// Range is a document span between two positions.
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// DiagnosticSeverity values (LSP DiagnosticSeverity).
const (
	SeverityError       = 1
	SeverityWarning     = 2
	SeverityInformation = 3
	SeverityHint        = 4
)

// Diagnostic is one published problem report. The c4drill server only emits
// errors, all attributed to Source "c4drill".
type Diagnostic struct {
	Range    Range  `json:"range"`
	Severity int    `json:"severity,omitempty"`
	Source   string `json:"source,omitempty"`
	Message  string `json:"message"`
}

// PublishDiagnosticsParams is the textDocument/publishDiagnostics payload.
type PublishDiagnosticsParams struct {
	URI         DocumentURI  `json:"uri"`
	Version     *int         `json:"version,omitempty"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

// TextDocumentSyncKind values (LSP TextDocumentSyncKind).
const (
	SyncNone      = 0
	SyncFull      = 1
	SyncIncrement = 2
)

// TextDocumentSyncOptions advertises how documents sync to the server.
type TextDocumentSyncOptions struct {
	OpenClose bool `json:"openClose"`
	Change    int  `json:"change"`
	Save      bool `json:"save"`
}

// ServerCapabilities is the capability list advertised by initialize. Each
// milestone adds its entries: M1 text sync, M2 language features, M3
// formatting, M4 the c4drill render method (via Experimental).
type ServerCapabilities struct {
	TextDocumentSync *TextDocumentSyncOptions `json:"textDocumentSync,omitempty"`
}

// ServerInfo names the server in initialize results.
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

// InitializeResult is the initialize response.
type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`
	ServerInfo   ServerInfo         `json:"serverInfo"`
}

// InitializedParams is the (empty) initialized notification payload.
type InitializedParams struct{}

// TextDocumentItem is a didOpen document snapshot.
type TextDocumentItem struct {
	URI        DocumentURI `json:"uri"`
	LanguageID string      `json:"languageId"`
	Version    int         `json:"version"`
	Text       string      `json:"text"`
}

// TextDocumentIdentifier references a document by URI.
type TextDocumentIdentifier struct {
	URI DocumentURI `json:"uri"`
}

// VersionedTextDocumentIdentifier references a document by URI and version.
type VersionedTextDocumentIdentifier struct {
	URI     DocumentURI `json:"uri"`
	Version int         `json:"version"`
}

// DidOpenTextDocumentParams is the textDocument/didOpen payload.
type DidOpenTextDocumentParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}

// TextDocumentContentChangeEvent is one didChange edit. Full sync (the
// server's advertised mode) sends a range-less full replacement; ranged
// edits are applied as splices defensively.
type TextDocumentContentChangeEvent struct {
	Range *Range `json:"range,omitempty"`
	Text  string `json:"text"`
}

// DidChangeTextDocumentParams is the textDocument/didChange payload.
type DidChangeTextDocumentParams struct {
	TextDocument   VersionedTextDocumentIdentifier  `json:"textDocument"`
	ContentChanges []TextDocumentContentChangeEvent `json:"contentChanges"`
}

// DidCloseTextDocumentParams is the textDocument/didClose payload.
type DidCloseTextDocumentParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// FileEvent type values (LSP FileChangeType).
const (
	FileCreated = 1
	FileChanged = 2
	FileDeleted = 3
)

// FileEvent is one external filesystem change.
type FileEvent struct {
	URI  DocumentURI `json:"uri"`
	Type int         `json:"type"`
}

// DidChangeWatchedFilesParams is the workspace/didChangeWatchedFiles payload:
// the client's report of on-disk changes the server cannot see otherwise
// (an included file edited outside the editor, for instance).
type DidChangeWatchedFilesParams struct {
	Changes []FileEvent `json:"changes"`
}
