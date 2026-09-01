// render.go implements the custom c4drill/renderDiagram method (issue #32,
// M4): the CLI render pipeline in-process — parse → include.Resolve →
// template.Expand → peer.Resolve → validate → view → graph → render SVG —
// so the SVG bytes are identical to `c4drill <file> -f svg` for the same
// model and options, and the diagnostics ride along in the response.
//
// Wire contract: see RenderDiagramParams / RenderDiagramResult in
// protocol.go — this is the live-preview primitive for #27/#29/#30/#31.
//
// Concurrency: internal/render's WASM graphviz engine is not thread-safe and
// serializes inside render() on wasmMutex. Render calls go through the
// exported render functions (RenderSVG) — never around them.

package lsp

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Djarvur/c4drill/internal/graph"
	"github.com/Djarvur/c4drill/internal/include"
	"github.com/Djarvur/c4drill/internal/parser"
	"github.com/Djarvur/c4drill/internal/peer"
	"github.com/Djarvur/c4drill/internal/render"
	"github.com/Djarvur/c4drill/internal/template"
	"github.com/Djarvur/c4drill/internal/validator"
	"github.com/Djarvur/c4drill/internal/view"
)

// renderFormatSVG is the only render output format in v1.
const renderFormatSVG = "svg"

// static render-pipeline errors (message-shaped like the CLI's stage wraps).
var (
	errGenerateView            = errors.New("failed to generate view")
	errBuildGraph              = errors.New("failed to build graph")
	errUnsupportedRenderFormat = errors.New("unsupported render format")
)

// pipelineModel aliases the hub type the composition pipeline produces.
type pipelineModel = parser.Model

// unsupportedFormatDiagnostic / renderErrorDiagnostic wrap render-time
// failures into the diagnostics channel of the response.
func unsupportedFormatDiagnostic(format string) Diagnostic {
	return Diagnostic{
		Severity: SeverityError,
		Source:   diagnosticSource,
		Message:  "render: " + errUnsupportedRenderFormat.Error() + ": " + format + " (supported: svg)",
	}
}

func renderErrorDiagnostic(err error) Diagnostic {
	return Diagnostic{
		Severity: SeverityError,
		Source:   diagnosticSource,
		Message:  "render: " + err.Error(),
	}
}

// renderDiagram is the c4drill/renderDiagram feature entry. The result
// always carries the pipeline diagnostics; the SVG is empty when validation
// fails (the CLI refuses to render an invalid model the same way).
func (s *Server) renderDiagram(ctx context.Context, doc *document, p RenderDiagramParams) *RenderDiagramResult {
	diags, m := s.pipelineForRender(doc)
	if len(diags) > 0 {
		return &RenderDiagramResult{SVG: "", Diagnostics: diags}
	}

	// Expanded-set override: replaces the model's [properties].expanded C1
	// drill-down set (per-unit expanded lists stay author-owned).
	if p.Expanded != nil {
		m.Properties.Expanded = p.Expanded
	}

	// Legend override: nil keeps the model default (properties.legend).
	if p.Legend != nil {
		m.Properties.Legend = p.Legend
	}

	format := p.Format
	if format == "" {
		format = renderFormatSVG
	}

	if format != renderFormatSVG {
		return &RenderDiagramResult{
			SVG:         "",
			Diagnostics: append(diags, unsupportedFormatDiagnostic(format)),
		}
	}

	svg, err := renderView(ctx, m, p, basenameOf(doc.Path))
	if err != nil {
		return &RenderDiagramResult{
			SVG:         "",
			Diagnostics: append(diags, renderErrorDiagnostic(err)),
		}
	}

	return &RenderDiagramResult{SVG: string(svg), Diagnostics: diags}
}

// pipelineForRender runs the composition pipeline on the open buffer and
// returns its diagnostics along with the merged model (nil when parsing
// failed — the diagnostics then name the failure, mirroring the CLI).
func (s *Server) pipelineForRender(doc *document) ([]Diagnostic, *pipelineModel) {
	m, err := parseByExt(doc.Path, doc.Text)
	if err != nil {
		return []Diagnostic{stageDiagnostic("parse: ", err)}, nil
	}

	m, err = include.ResolveWithReader(m, filepath.Dir(doc.Path), doc.Path, s.overlayRead())
	if err != nil {
		return []Diagnostic{stageDiagnostic("include: ", err)}, nil
	}

	m, err = template.Expand(m)
	if err != nil {
		return []Diagnostic{stageDiagnostic("expand: ", err)}, nil
	}

	if err := peer.Resolve(m); err != nil {
		return []Diagnostic{stageDiagnostic("resolve peers: ", err)}, nil
	}

	valErrors := validator.Validate(m)

	diags := make([]Diagnostic, 0, len(valErrors))
	for _, ve := range valErrors {
		diags = append(diags, Diagnostic{
			Range:    Range{},
			Severity: SeverityError,
			Source:   diagnosticSource,
			Message:  ve.Error(),
		})
	}

	return diags, m
}

// renderView generates the requested view, builds the graph, and renders
// SVG — the same stages processView runs in the CLI (root.go), with default
// presentation flags (no --plain/--no-colors equivalents exist over LSP v1).
func renderView(ctx context.Context, m *pipelineModel, p RenderDiagramParams, basename string) ([]byte, error) {
	if p.AllExpanded {
		return renderExpanded(ctx, m)
	}

	target := p.Target

	var v *view.View

	switch {
	case target == "":
		v = view.GenerateC1View(m)
	case isC2Target(target):
		v = view.GenerateC2View(m, target)
	default:
		v = view.GenerateC3View(m, target)
	}

	if v == nil {
		return nil, errGenerateView
	}

	g := graph.BuildGraphWithPath(v, target, basename, renderFormatSVG)
	if g == nil {
		return nil, errBuildGraph
	}

	return renderSVG(ctx, g)
}

// renderExpanded is the --expanded mode: one all-nested diagram.
func renderExpanded(ctx context.Context, m *pipelineModel) ([]byte, error) {
	v := view.GenerateExpandedView(m)
	if v == nil {
		return nil, errGenerateView
	}

	g := graph.BuildExpandedGraph(v)
	if g == nil {
		return nil, errBuildGraph
	}

	return renderSVG(ctx, g)
}

// renderSVG serializes through internal/render's exported functions (the
// WASM engine locks internally on wasmMutex) and honors client cancellation
// before the render starts.
//
//nolint:contextcheck // internal/render v1 takes no context and owns its serialization
func renderSVG(ctx context.Context, g *graph.Graph) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err //nolint:wrapcheck // the caller maps this to a render diagnostic
	}

	svg, err := render.RenderSVG(g)
	if err != nil {
		return nil, fmt.Errorf("render svg: %w", err)
	}

	return svg, nil
}

// isC2Target reports a one-segment target (a unit's C2 diagram).
func isC2Target(target string) bool {
	return !strings.Contains(target, ".")
}

// basenameOf strips the extension (the CLI's diagram-naming input).
func basenameOf(path string) string {
	base := path
	if i := strings.LastIndexByte(base, '/'); i >= 0 {
		base = base[i+1:]
	}

	if dot := strings.LastIndexByte(base, '.'); dot > 0 {
		base = base[:dot]
	}

	return base
}
