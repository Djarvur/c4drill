// dispatch.go is the single transport-agnostic RPC surface of the GUI
// backend: method name → handler over JSON. gui/main routes both the Wails
// bindings and the HTTP fallback through Dispatch, so the two transports
// cannot drift and the frontend speaks one protocol in dev (browser) and
// production (desktop window) alike.

package app

import (
	"encoding/json"
	"fmt"
)

// handlers is the dispatch table. Registered in newDispatch.
func (a *App) handlers() map[string]func(params json.RawMessage) (any, error) {
	return map[string]func(json.RawMessage) (any, error){
		"openProject":  a.dispatchOpenProject,
		"listFiles":    a.dispatchListFiles,
		"readFile":     a.dispatchReadFile,
		"writeFile":    a.dispatchWriteFile,
		"didOpen":      a.dispatchDidOpen,
		"didChange":    a.dispatchDidChange,
		"didClose":     a.dispatchDidClose,
		"completion":   a.dispatchCompletion,
		"hover":        a.dispatchHover,
		"definition":   a.dispatchDefinition,
		"symbols":      a.dispatchSymbols,
		"format":       a.dispatchFormat,
		"render":       a.dispatchRender,
		"resolveDrill": a.dispatchResolveDrill,
		"export":       a.dispatchExport,
		"appInfo":      a.dispatchAppInfo,
	}
}

// Dispatch routes one backend call: method name + JSON params → JSON result.
// Errors are returned, never thrown into the transport.
func (a *App) Dispatch(method string, params json.RawMessage) (json.RawMessage, error) {
	h, ok := a.handlers()[method]
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrUnknownMethod, method)
	}

	if params == nil {
		params = json.RawMessage("{}")
	}

	result, err := h(params)
	if err != nil {
		return nil, err
	}

	encoded, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("encode %s result: %w", method, err)
	}

	return encoded, nil
}

// decodeParams unmarshals params into out.
func decodeParams(params json.RawMessage, out any) error {
	if err := json.Unmarshal(params, out); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}

	return nil
}

// --- per-method adapters -------------------------------------------------

type openProjectParams struct {
	Dir string `json:"dir"`
}

func (a *App) dispatchOpenProject(params json.RawMessage) (any, error) {
	var p openProjectParams
	if err := decodeParams(params, &p); err != nil {
		return nil, err
	}

	return a.OpenProject(p.Dir)
}

func (a *App) dispatchListFiles(_ json.RawMessage) (any, error) {
	return a.ListFiles()
}

type readFileParams struct {
	Path string `json:"path"`
}

func (a *App) dispatchReadFile(params json.RawMessage) (any, error) {
	var p readFileParams
	if err := decodeParams(params, &p); err != nil {
		return nil, err
	}

	return a.ReadFile(p.Path)
}

type writeFileParams struct {
	Path string `json:"path"`
	Text string `json:"text"`
}

func (a *App) dispatchWriteFile(params json.RawMessage) (any, error) {
	var p writeFileParams
	if err := decodeParams(params, &p); err != nil {
		return nil, err
	}

	return map[string]bool{"ok": true}, a.WriteFile(p.Path, p.Text)
}

type didOpenParams struct {
	Path string `json:"path"`
	Text string `json:"text"`
}

func (a *App) dispatchDidOpen(params json.RawMessage) (any, error) {
	var p didOpenParams
	if err := decodeParams(params, &p); err != nil {
		return nil, err
	}

	return map[string]bool{"ok": true}, a.DidOpen(p.Path, p.Text)
}

type didChangeParams struct {
	Path    string `json:"path"`
	Text    string `json:"text"`
	Version int    `json:"version"`
}

func (a *App) dispatchDidChange(params json.RawMessage) (any, error) {
	var p didChangeParams
	if err := decodeParams(params, &p); err != nil {
		return nil, err
	}

	return map[string]bool{"ok": true}, a.DidChange(p.Path, p.Text, p.Version)
}

type didCloseParams struct {
	Path string `json:"path"`
}

func (a *App) dispatchDidClose(params json.RawMessage) (any, error) {
	var p didCloseParams
	if err := decodeParams(params, &p); err != nil {
		return nil, err
	}

	return map[string]bool{"ok": true}, a.DidClose(p.Path)
}

type positionParams struct {
	Path      string `json:"path"`
	Line      uint32 `json:"line"`
	Character uint32 `json:"character"`
}

func (a *App) dispatchCompletion(params json.RawMessage) (any, error) {
	var p positionParams
	if err := decodeParams(params, &p); err != nil {
		return nil, err
	}

	return a.Completion(p.Path, p.Line, p.Character)
}

func (a *App) dispatchHover(params json.RawMessage) (any, error) {
	var p positionParams
	if err := decodeParams(params, &p); err != nil {
		return nil, err
	}

	return a.Hover(p.Path, p.Line, p.Character)
}

func (a *App) dispatchDefinition(params json.RawMessage) (any, error) {
	var p positionParams
	if err := decodeParams(params, &p); err != nil {
		return nil, err
	}

	return a.Definition(p.Path, p.Line, p.Character)
}

type symbolsParams struct {
	Path string `json:"path"`
}

func (a *App) dispatchSymbols(params json.RawMessage) (any, error) {
	var p symbolsParams
	if err := decodeParams(params, &p); err != nil {
		return nil, err
	}

	return a.Symbols(p.Path)
}

type formatParams struct {
	Path string `json:"path"`
}

func (a *App) dispatchFormat(params json.RawMessage) (any, error) {
	var p formatParams
	if err := decodeParams(params, &p); err != nil {
		return nil, err
	}

	return a.Format(p.Path)
}

type renderParams struct {
	Path string        `json:"path"`
	Opts RenderOptions `json:"opts"`
}

func (a *App) dispatchRender(params json.RawMessage) (any, error) {
	var p renderParams
	if err := decodeParams(params, &p); err != nil {
		return nil, err
	}

	return a.Render(p.Path, p.Opts)
}

type resolveDrillParams struct {
	Path   string `json:"path"`
	Target string `json:"target"`
	Href   string `json:"href"`
}

func (a *App) dispatchResolveDrill(params json.RawMessage) (any, error) {
	var p resolveDrillParams
	if err := decodeParams(params, &p); err != nil {
		return nil, err
	}

	target, err := a.ResolveDrill(p.Path, p.Target, p.Href)
	if err != nil {
		return nil, err
	}

	return map[string]string{"target": target}, nil
}

type exportParams struct {
	Path   string        `json:"path"`
	Opts   RenderOptions `json:"opts"`
	Format string        `json:"format"`
	OutDir string        `json:"outDir"`
}

func (a *App) dispatchExport(params json.RawMessage) (any, error) {
	var p exportParams
	if err := decodeParams(params, &p); err != nil {
		return nil, err
	}

	return a.Export(p.Path, p.Opts, p.Format, p.OutDir)
}

// Info is the boot handshake: the frontend learns which project (if any) is
// already open (app.stutters, so the type is plain Info).
type Info struct {
	InitialDir string `json:"initialDir"`
	Version    string `json:"version"`
}

// guiVersion is stamped by gui/main via SetVersion.
//
//nolint:gochecknoglobals // build-time injection precedent: cmd/c4drill root.go `version`
var guiVersion = "dev"

// SetVersion overrides the GUI version string (build-time injection).
func SetVersion(v string) {
	guiVersion = v
}

func (a *App) dispatchAppInfo(_ json.RawMessage) (any, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	return Info{InitialDir: a.root, Version: guiVersion}, nil
}
