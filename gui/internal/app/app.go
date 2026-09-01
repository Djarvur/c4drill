// Package app is the GUI backend (issue #31): the orchestration layer on top
// of the existing in-process Go packages. It owns the project workspace (the
// .toml/.c4d file set), drives the shared LSP server core (internal/lsp) over
// its in-memory transport — the same server the editor clients use — and
// exposes the live-preview render plus the export pipeline.
//
// The package is transport-agnostic: Wails bindings and the HTTP fallback in
// gui/main both call Dispatch and consume the same JSON contracts, and
// server→client traffic (diagnostics) flows through the EventSink.
package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// EventSink receives backend→frontend events (diagnostics pushes, P1 chat
// streaming). It is called from LSP notification delivery — which happens
// while the app lock may be held — so implementations must not call back
// into the App.
type EventSink func(event string, payload any)

// diagnosticsEvent is the event name diagnostics are pushed under.
const diagnosticsEvent = "diagnostics"

// modelExtToml / modelExtC4d are the accepted model-file extensions (the
// CLI's D-27 dispatch set).
const (
	modelExtToml = ".toml"
	modelExtC4d  = ".c4d"
)

// ProjectInfo describes the opened project.
type ProjectInfo struct {
	Dir   string     `json:"dir"`
	Files []FileInfo `json:"files"`
}

// FileInfo is one model file, project-relative with slash separators.
type FileInfo struct {
	Path string `json:"path"`
}

// FileContent is a file read result.
type FileContent struct {
	Path string `json:"path"`
	Text string `json:"text"`
}

// DiagnosticsEvent is the payload pushed to the frontend under the
// "diagnostics" event: one document's published diagnostics (the LSP
// publishDiagnostics notification, path made project-relative).
type DiagnosticsEvent struct {
	Path        string       `json:"path"`
	Version     *int         `json:"version,omitempty"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

// App is the GUI backend. It is safe for concurrent use.
type App struct {
	mu     sync.Mutex
	root   string // absolute project directory ("" until OpenProject)
	events EventSink

	lsp *lspBridge

	// openFiles tracks the project-relative paths currently open in the LSP
	// session so the preview can render buffers without re-reading disk.
	openFiles map[string]bool
}

// New builds an App. sink may be nil (events are dropped).
func New(sink EventSink) *App {
	if sink == nil {
		sink = func(string, any) {}
	}

	a := &App{
		events:    sink,
		openFiles: make(map[string]bool),
	}

	a.lsp = newLSPBridge(a.onDiagnostics)

	return a
}

// SetEventSink swaps the backend→frontend event sink. Wails and the HTTP
// fallback install theirs at startup; until then events are dropped.
func (a *App) SetEventSink(sink EventSink) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if sink == nil {
		sink = func(string, any) {}
	}

	a.events = sink
}

// OpenProject opens dir as the project root, (re)starts the LSP session and
// lists the model files. Previously open documents are discarded.
func (a *App) OpenProject(dir string) (*ProjectInfo, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("resolve project dir: %w", err)
	}

	st, err := os.Stat(abs)
	if err != nil {
		return nil, fmt.Errorf("open project: %w", err)
	}

	if !st.IsDir() {
		return nil, fmt.Errorf("open project: %w: %s", ErrPathOutsideProject, abs)
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	a.root = canonicalPath(abs)
	a.openFiles = make(map[string]bool)
	a.lsp.restart()

	files, err := a.listFilesLocked()
	if err != nil {
		return nil, err
	}

	return &ProjectInfo{Dir: a.root, Files: files}, nil
}

// ListFiles returns the project's model files (recursive, .toml/.c4d),
// sorted by path.
func (a *App) ListFiles() (*ProjectInfo, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.root == "" {
		return nil, errNoProjectOpen
	}

	files, err := a.listFilesLocked()
	if err != nil {
		return nil, err
	}

	return &ProjectInfo{Dir: a.root, Files: files}, nil
}

// ReadFile returns the on-disk content of a project file.
func (a *App) ReadFile(rel string) (*FileContent, error) {
	abs, err := a.absOf(rel)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(abs) //nolint:gosec // G304 is the product: the user picks which project file to read
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", rel, err)
	}

	return &FileContent{Path: rel, Text: string(data)}, nil
}

// WriteFile writes text to a project file (save). The write stays inside the
// project root.
func (a *App) WriteFile(rel, text string) error {
	abs, err := a.absOf(rel)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(abs), 0o750); err != nil {
		return fmt.Errorf("create directory for %s: %w", rel, err)
	}

	if err := os.WriteFile(abs, []byte(text), 0o600); err != nil {
		return fmt.Errorf("write %s: %w", rel, err)
	}

	// The on-disk file changed behind the LSP's buffer overlay; open buffers
	// that include this file must revalidate against fresh disk content for
	// the parts they do not hold open.
	a.lsp.watchedFileChanged(abs)

	return nil
}

// onDiagnostics translates an LSP publishDiagnostics notification into a
// frontend event with project-relative paths. Called from the LSP notifier
// (inside the server's Handle), so it must not re-enter the LSP bridge.
func (a *App) onDiagnostics(uri string, version *int, diags []Diagnostic) {
	rel, ok := a.uriToRel(uri)
	if !ok {
		return // outside the project — nothing the UI can select
	}

	a.events(diagnosticsEvent, DiagnosticsEvent{
		Path:        rel,
		Version:     version,
		Diagnostics: diags,
	})
}

// uriToRel converts an LSP file URI to a project-relative path. Caller holds
// a.mu (reads a.root).
func (a *App) uriToRel(uri string) (string, bool) {
	rel, err := a.relOf(uriToPath(uri))
	if err != nil {
		return "", false
	}

	return rel, true
}

// relOf maps an absolute path into the project, returning the slash-separated
// relative form. Paths outside the root are rejected (the GUI's edit scope is
// the opened project, issue #31). Caller holds a.mu.
func (a *App) relOf(abs string) (string, error) {
	if a.root == "" {
		return "", errNoProjectOpen
	}

	abs = canonicalPath(abs)

	rel, err := filepath.Rel(a.root, abs)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("%w: %s", ErrPathOutsideProject, abs)
	}

	return filepath.ToSlash(rel), nil
}

// absOf resolves a project-relative path to its absolute canonical form.
// Caller holds a.mu.
func (a *App) absOf(rel string) (string, error) {
	if rel == "" {
		return "", errEmptyPath
	}

	if strings.Contains(rel, "..") {
		return "", fmt.Errorf("%w: %q", ErrPathOutsideProject, rel)
	}

	return canonicalPath(filepath.Join(a.root, filepath.FromSlash(rel))), nil
}

// listFilesLocked walks the project root for .toml/.c4d files. Caller holds
// a.mu.
func (a *App) listFilesLocked() ([]FileInfo, error) {
	files := make([]FileInfo, 0, 16)

	err := filepath.WalkDir(a.root, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if name := d.Name(); name != "." && isSkippedDir(name) {
				return filepath.SkipDir
			}

			return nil
		}

		if isModelPath(p) {
			if rel, rerr := a.relOf(p); rerr == nil {
				files = append(files, FileInfo{Path: rel})
			}
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("list project files: %w", err)
	}

	return files, nil
}

// isSkippedDir reports directories the file walk must not descend into
// (VCS/hidden dirs and frontend build trees).
func isSkippedDir(name string) bool {
	return strings.HasPrefix(name, ".") || name == "node_modules" || name == "dist"
}

// canonicalPath normalizes to the absolute cleaned form (mirrors the LSP's
// own canonicalization: Clean + Abs, symlinks unresolved).
func canonicalPath(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		return filepath.Clean(path)
	}

	return filepath.Clean(abs)
}

// isModelPath reports whether the path looks like a c4drill model file.
func isModelPath(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case modelExtToml, modelExtC4d:
		return true
	default:
		return false
	}
}
