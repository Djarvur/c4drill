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
	TextDocumentSync           *TextDocumentSyncOptions `json:"textDocumentSync,omitempty"`
	CompletionProvider         *CompletionOptions       `json:"completionProvider,omitempty"`
	HoverProvider              bool                     `json:"hoverProvider,omitempty"`
	DefinitionProvider         bool                     `json:"definitionProvider,omitempty"`
	DocumentSymbolProvider     bool                     `json:"documentSymbolProvider,omitempty"`
	DocumentFormattingProvider bool                     `json:"documentFormattingProvider,omitempty"`
}

// CompletionOptions advertises the completion capability.
type CompletionOptions struct {
	TriggerCharacters []string `json:"triggerCharacters,omitempty"`
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

// TextDocumentPositionParams is the shared shape of the position-bearing
// requests (completion, hover, definition).
type TextDocumentPositionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

// CompletionItemKind values used by the c4drill server.
const (
	KindModule     = 2  // Module
	KindField      = 5  // Field
	KindClass      = 7  // Class
	KindValue      = 12 // Value
	KindFile       = 17 // File
	KindKeyword    = 14 // Keyword
	KindFolder     = 19 // Folder
	KindEnumMember = 20 // EnumMember
	KindObject     = 19 // Object (documentSymbol)
)

// CompletionItem is one completion proposal. Most items rely on the client's
// default word-range replacement; FilterText carries the full logical text
// where the label differs from what is being replaced (dotted paths).
type CompletionItem struct {
	Label         string `json:"label"`
	Kind          int    `json:"kind,omitempty"`
	Detail        string `json:"detail,omitempty"`
	Documentation string `json:"documentation,omitempty"`
	SortText      string `json:"sortText,omitempty"`
	FilterText    string `json:"filterText,omitempty"`
	InsertText    string `json:"insertText,omitempty"`
}

// CompletionList is the textDocument/completion result.
type CompletionList struct {
	IsIncomplete bool             `json:"isIncomplete"`
	Items        []CompletionItem `json:"items"`
}

// MarkupContent is formatted hover content.
type MarkupContent struct {
	Kind  string `json:"kind"` // "markdown" or "plaintext"
	Value string `json:"value"`
}

// Hover is the textDocument/hover result.
type Hover struct {
	Contents MarkupContent `json:"contents"`
	Range    *Range        `json:"range,omitempty"`
}

// Location points into a document.
type Location struct {
	URI   DocumentURI `json:"uri"`
	Range Range       `json:"range"`
}

// DocumentSymbol is one outline entry with hierarchy.
type DocumentSymbol struct {
	Name           string           `json:"name"`
	Detail         string           `json:"detail,omitempty"`
	Kind           int              `json:"kind"`
	Range          Range            `json:"range"`
	SelectionRange Range            `json:"selectionRange"`
	Children       []DocumentSymbol `json:"children,omitempty"`
}

// FormattingOptions are the client's formatting preferences. c4drill
// formatting is whitespace-normalizing with fixed style, so only the shape
// is accepted — the values do not alter the output (gofmt-style).
type FormattingOptions struct {
	TabSize      int  `json:"tabSize"`
	InsertSpaces bool `json:"insertSpaces"`
}

// DocumentFormattingParams is the textDocument/formatting payload.
type DocumentFormattingParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Options      FormattingOptions      `json:"options"`
}

// TextEdit is one contiguous replacement.
type TextEdit struct {
	Range   Range  `json:"range"`
	NewText string `json:"newText"`
}

// RenderDiagramParams is the c4drill/renderDiagram request (issue #32, M4 —
// the live-preview contract for #27/#29/#30/#31). Wire shape (v1):
//
//	{
//	  "textDocument": { "uri": "file:///path/model.toml" },
//	  "target": "cloud.ui",     // optional: unit path to render. "" = C1
//	                            // context; one segment = that unit's C2
//	                            // diagram; deeper = C3 — the CLI layout.
//	  "allExpanded": false,     // optional: the single all-expanded diagram
//	                            // (the CLI --expanded mode); overrides target
//	  "expanded": ["cloud"],    // optional: replacement for the model's
//	                            // [properties].expanded C1 drill-down set
//	                            // (per-unit expanded lists stay author-owned)
//	  "legend": true,           // optional: overrides properties.legend
//	  "format": "svg"           // optional, only "svg" in v1 (default)
//	}
//
// The response carries the SVG text plus the pipeline diagnostics observed
// while preparing it (see RenderDiagramResult). Validation failures return
// HTTP-style success with the diagnostics and an empty svg — the same
// information the CLI prints before refusing to render.
type RenderDiagramParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Target       string                 `json:"target,omitempty"`
	AllExpanded  bool                   `json:"allExpanded,omitempty"`
	// Expanded has no omitempty deliberately: an EMPTY slice is a
	// meaningful override ("collapse all") and must reach the server.
	Expanded []string `json:"expanded"`
	Legend   *bool    `json:"legend,omitempty"`
	Format   string   `json:"format,omitempty"`
}

// RenderDiagramResult is the c4drill/renderDiagram response: the rendered
// SVG text (empty when validation failed) and every pipeline diagnostic.
type RenderDiagramResult struct {
	SVG         string       `json:"svg"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}
