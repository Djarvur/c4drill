// render.go is the preview area's backend: live rendering through the shared
// LSP c4drill/renderDiagram method (so the SVG is identical to what the CLI
// and the editor clients produce), drill-target resolution for SVG-internal
// navigation links, and breadcrumb computation.
//
// SVG links encode the CLI's output file layout ({basename}.svg for C1,
// {basename}/{unit}.{svg} with dotted paths as directories below it — see
// internal/graph ComputeExploreURL); ResolveDrillTarget inverts that layout
// back to a unit path.

package app

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/Djarvur/c4drill/internal/lsp"
)

// RenderOptions are the preview's render controls (toolbar state). The zero
// value renders the model's own defaults.
type RenderOptions struct {
	// Target is the unit path to render: "" = C1 context, one segment = that
	// unit's C2, deeper = C3 (the CLI layout). Ignored when AllExpanded.
	Target string `json:"target"`
	// AllExpanded renders the single --expanded-style diagram.
	AllExpanded bool `json:"allExpanded"`
	// Expanded, when non-nil, replaces the model's [properties].expanded C1
	// drill-down set (per-unit expanded lists stay author-owned). An empty
	// slice means "collapse all" and is distinct from nil (model default).
	Expanded []string `json:"expanded,omitempty"`
	// Legend overrides properties.legend when non-nil.
	Legend *bool `json:"legend,omitempty"`
}

// Breadcrumb is one crumb of the preview's navigation bar.
type Breadcrumb struct {
	// Name is the display label (the model/file name for the C1 crumb, the
	// unit's own name below it).
	Name string `json:"name"`
	// Target is the render target the crumb navigates to ("" for C1).
	Target string `json:"target"`
}

// RenderResult is the preview render response.
type RenderResult struct {
	SVG         string       `json:"svg"`
	Diagnostics []Diagnostic `json:"diagnostics"`
	// Target echoes the target actually rendered.
	Target      string       `json:"target"`
	AllExpanded bool         `json:"allExpanded"`
	Breadcrumbs []Breadcrumb `json:"breadcrumbs"`
}

// Render renders the preview for the given project file via the shared LSP
// render method. The file is opened in the LSP session on first use (from
// disk) if the editor has not opened it yet.
func (a *App) Render(rel string, opts RenderOptions) (*RenderResult, error) {
	abs, err := a.absOf(rel)
	if err != nil {
		return nil, err
	}

	if !isModelPath(abs) {
		return nil, fmt.Errorf("render: %w: %s", ErrNotModelFile, rel)
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.openFiles[rel] {
		data, rerr := os.ReadFile(abs) //nolint:gosec // G304 is the product: the user picked this project file
		if rerr != nil {
			return nil, fmt.Errorf("read %s: %w", rel, rerr)
		}

		a.lsp.DidOpen(abs, string(data), 0)
		a.openFiles[rel] = true
	}

	res, err := a.lsp.RenderDiagram(abs, lsp.RenderDiagramParams{
		Target:      opts.Target,
		AllExpanded: opts.AllExpanded,
		Expanded:    opts.Expanded,
		Legend:      opts.Legend,
		Format:      "svg",
	})
	if err != nil {
		return nil, fmt.Errorf("render: %w", err)
	}

	return &RenderResult{
		SVG:         res.SVG,
		Diagnostics: res.Diagnostics,
		Target:      opts.Target,
		AllExpanded: opts.AllExpanded,
		Breadcrumbs: BreadcrumbsFor(a.basenameOf(rel), opts.Target, opts.AllExpanded),
	}, nil
}

// ResolveDrill maps an SVG navigation href (clicked in the preview) to the
// render target it navigates to, relative to the currently rendered target.
// External URLs (http/https/mailto/...) error — the frontend lets the
// runtime handle those itself.
func (a *App) ResolveDrill(rel, currentTarget, href string) (string, error) {
	if !isInternalHref(href) {
		return "", fmt.Errorf("%w: %q", errNotInternalLink, href)
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	return resolveDrillTarget(a.basenameOf(rel), currentTarget, href)
}

// BreadcrumbsFor computes the navigation crumbs for a target:
// C1 (file), then one crumb per unit-path segment.
func BreadcrumbsFor(basename, target string, allExpanded bool) []Breadcrumb {
	crumbs := []Breadcrumb{{Name: basename, Target: ""}}
	if allExpanded || target == "" {
		return crumbs
	}

	acc := ""

	for _, seg := range strings.Split(target, ".") {
		acc = joinTarget(acc, seg)
		crumbs = append(crumbs, Breadcrumb{Name: seg, Target: acc})
	}

	return crumbs
}

// basenameOf returns the file's base name without extension (the CLI's
// diagram-naming input). rel is project-relative, so no lock is needed.
func (a *App) basenameOf(rel string) string {
	base := filepath.Base(rel)

	if dot := strings.LastIndexByte(base, '.'); dot > 0 {
		base = base[:dot]
	}

	return base
}

// resolveDrillTarget is the pure core of ResolveDrill (unit-testable without
// an App): href is interpreted relative to the virtual file layout the CLI
// writes — current "" lives at {basename}.svg, "a" at {basename}/a.svg,
// "a.b" at {basename}/a/b.svg.
func resolveDrillTarget(basename, current, href string) (string, error) {
	if basename == "" {
		return "", errEmptyBasename
	}

	curFile := fileOfTarget(basename, current)
	resolved := path.Clean(path.Join(path.Dir(curFile), href))

	if resolved == "." || resolved == ".." || strings.HasPrefix(resolved, "../") {
		return "", fmt.Errorf("%w: %q", errBadNavigation, href)
	}

	target, ok := targetOfFile(basename, resolved)
	if !ok {
		return "", fmt.Errorf("%w: %q", errBadNavigation, href)
	}

	return target, nil
}

// fileOfTarget maps a render target to its virtual output file (the CLI
// writer layout).
func fileOfTarget(basename, target string) string {
	if target == "" {
		return basename + ".svg"
	}

	return basename + "/" + strings.ReplaceAll(target, ".", "/") + ".svg"
}

// targetOfFile maps a virtual output file back to its render target. ok is
// false for files outside the diagram set.
func targetOfFile(basename, file string) (string, bool) {
	if !strings.HasSuffix(file, ".svg") {
		return "", false
	}

	stem := strings.TrimSuffix(file, ".svg")

	switch {
	case stem == basename:
		return "", true
	case strings.HasPrefix(stem, basename+"/"):
		return strings.ReplaceAll(strings.TrimPrefix(stem, basename+"/"), "/", "."), true
	default:
		return "", false
	}
}

// isInternalHref reports whether href is a relative diagram link (as opposed
// to the external reference URLs the renderer also emits).
func isInternalHref(href string) bool {
	if href == "" {
		return false
	}

	lower := strings.ToLower(href)

	for _, scheme := range []string{"http://", "https://", "mailto:", "#"} {
		if strings.HasPrefix(lower, scheme) {
			return false
		}
	}

	return strings.HasSuffix(lower, ".svg")
}

// joinTarget appends one segment to a dotted unit path.
func joinTarget(prefix, seg string) string {
	if prefix == "" {
		return seg
	}

	return prefix + "." + seg
}
