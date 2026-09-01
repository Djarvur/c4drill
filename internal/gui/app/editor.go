// editor.go is the editor area's App surface: thin project-relative wrappers
// over the LSP bridge. Buffer lifecycle (didOpen/didChange/didClose) is
// tracked so the preview can render documents the editor already has open
// without re-reading them, and definition results come back in
// project-relative paths the frontend can open directly.

package app

import (
	"fmt"

	"github.com/Djarvur/c4drill/internal/lsp"
)

// DidOpen registers an editor-opened buffer with the LSP session.
func (a *App) DidOpen(rel, text string) error {
	return a.openBuffer(rel, text, 1)
}

// DidChange pushes a buffer update (full text, the server's sync mode).
func (a *App) DidChange(rel, text string, version int) error {
	return a.openBuffer(rel, text, version)
}

// DidClose drops a buffer from the LSP session.
func (a *App) DidClose(rel string) error {
	abs, err := a.absOf(rel)
	if err != nil {
		return err
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.openFiles[rel] {
		return nil
	}

	a.lsp.DidClose(abs)
	delete(a.openFiles, rel)

	return nil
}

// Completion requests completions at a position in a project file.
func (a *App) Completion(rel string, line, character uint32) (*lsp.CompletionList, error) {
	abs, err := a.absOf(rel)
	if err != nil {
		return nil, err
	}

	return a.lsp.Completion(abs, line, character)
}

// Hover requests hover content at a position (nil when nothing to show).
func (a *App) Hover(rel string, line, character uint32) (*lsp.Hover, error) {
	abs, err := a.absOf(rel)
	if err != nil {
		return nil, err
	}

	return a.lsp.Hover(abs, line, character)
}

// DefinitionResult is the definition response: the target file as a
// project-relative path (the LSP URI converted) plus its range.
type DefinitionResult struct {
	Path  string    `json:"path"`
	Range lsp.Range `json:"range"`
}

// Definition returns the definition location for the symbol at a position.
// A nil result means the server found no definition.
func (a *App) Definition(rel string, line, character uint32) (*DefinitionResult, error) {
	abs, err := a.absOf(rel)
	if err != nil {
		return nil, err
	}

	loc, err := a.lsp.Definition(abs, line, character)
	if err != nil || loc == nil {
		return nil, err
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	targetRel, rerr := a.relOf(uriToPath(string(loc.URI)))
	if rerr != nil {
		return nil, rerr
	}

	return &DefinitionResult{Path: targetRel, Range: loc.Range}, nil
}

// Symbols returns the document outline.
func (a *App) Symbols(rel string) ([]lsp.DocumentSymbol, error) {
	abs, err := a.absOf(rel)
	if err != nil {
		return nil, err
	}

	return a.lsp.Symbols(abs)
}

// FormatResult is the format response.
type FormatResult struct {
	Text string `json:"text"`
}

// Format formats the current buffer of a project file and returns the
// formatted text ("" when already formatted or not formattable).
func (a *App) Format(rel string) (*FormatResult, error) {
	abs, err := a.absOf(rel)
	if err != nil {
		return nil, err
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	text, err := a.lsp.Format(abs, 2, true)

	return &FormatResult{Text: text}, err
}

// openBuffer syncs one buffer into the LSP session under the lock (the
// shared implementation of DidOpen/DidChange; a buffer the session has not
// seen yet is opened regardless of the requested op).
func (a *App) openBuffer(rel, text string, version int) error {
	abs, err := a.absOf(rel)
	if err != nil {
		return err
	}

	if !isModelPath(abs) {
		return fmt.Errorf("%w: %s", ErrNotModelFile, rel)
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.openFiles[rel] {
		a.lsp.DidOpen(abs, text, version)
		a.openFiles[rel] = true

		return nil
	}

	a.lsp.DidChange(abs, text, version)

	return nil
}
