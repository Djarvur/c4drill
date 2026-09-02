// export.go is the export menu's backend: it produces byte-identical
// artifacts to `c4drill <file> -f <format> -o <dir>` for the same model and
// view, by running the same pipeline stages the CLI runs (parse →
// include.Resolve → template.Expand → peer.Resolve → validate → view →
// graph) and writing through internal/output's Writer conventions.
//
// The pipeline here mirrors internal/lsp's renderDiagram stages; the LSP v1
// method speaks SVG only, so non-SVG export formats run the stages directly
// (a documented package gap: if c4drill/renderDiagram gains a format
// parameter beyond "svg", this file should collapse into it).

package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Djarvur/c4drill/internal/c4d"
	"github.com/Djarvur/c4drill/internal/graph"
	"github.com/Djarvur/c4drill/internal/include"
	"github.com/Djarvur/c4drill/internal/output"
	"github.com/Djarvur/c4drill/internal/parser"
	"github.com/Djarvur/c4drill/internal/peer"
	"github.com/Djarvur/c4drill/internal/render"
	"github.com/Djarvur/c4drill/internal/template"
	"github.com/Djarvur/c4drill/internal/validator"
	"github.com/Djarvur/c4drill/internal/view"
)

// Export formats — the CLI's -f set (the preview export menu mirrors these).
const (
	FormatSVG      = "svg"
	FormatHTML     = "html"
	FormatDOT      = "dot"
	FormatPNG      = "png"
	FormatPlantUML = "plantuml"
)

// ExportResult names the files an export produced (paths relative to the
// output directory).
type ExportResult struct {
	Format string   `json:"format"`
	Files  []string `json:"files"`
}

// Export writes the current view (target/allExpanded per opts) of rel into
// outDir using the CLI's naming layout. The model is read from disk (the UI
// saves before exporting). Includes resolve from disk as well — export is an
// artifact-writing action, not an editor interaction.
func (a *App) Export(rel string, opts RenderOptions, format, outDir string) (*ExportResult, error) {
	if !validExportFormat(format) {
		return nil, ErrInvalidExportFormat
	}

	abs, err := a.absOf(rel)
	if err != nil {
		return nil, err
	}

	if !isModelPath(abs) {
		return nil, fmt.Errorf("export: %w: %s", ErrNotModelFile, rel)
	}

	if outDir == "" {
		outDir = a.projectExportDir()
	}

	if err := os.MkdirAll(outDir, 0o750); err != nil {
		return nil, fmt.Errorf("create export dir: %w", err)
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	m, err := a.pipelineFromDisk(abs)
	if err != nil {
		return nil, err
	}

	files, err := exportView(m, a.basenameOf(rel), opts, format, output.NewWriter(outDir))
	if err != nil {
		return nil, err
	}

	return &ExportResult{Format: format, Files: files}, nil
}

// projectExportDir is the default export destination: <project>/diagrams.
func (a *App) projectExportDir() string {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.root == "" {
		return "."
	}

	return filepath.Join(a.root, "diagrams")
}

// pipelineFromDisk runs the composition pipeline over the on-disk file:
// parse → include.Resolve → template.Expand → peer.Resolve → validate. It
// mirrors the CLI/LSP stage order; a stage failure returns the stage-named
// error the CLI prints. Caller holds a.mu.
func (a *App) pipelineFromDisk(abs string) (*parser.Model, error) {
	data, err := os.ReadFile(abs) //nolint:gosec // G304 is the product: the user picked this project file
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", filepath.Base(abs), err)
	}

	m, err := parseByExt(abs, data)
	if err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}

	m, err = include.Resolve(m, filepath.Dir(abs), abs)
	if err != nil {
		return nil, fmt.Errorf("include: %w", err)
	}

	m, err = template.Expand(m)
	if err != nil {
		return nil, fmt.Errorf("expand: %w", err)
	}

	if err := peer.Resolve(m); err != nil {
		return nil, fmt.Errorf("resolve peers: %w", err)
	}

	if valErrors := validator.Validate(m); len(valErrors) > 0 {
		msgs := make([]string, 0, len(valErrors))
		for _, ve := range valErrors {
			msgs = append(msgs, ve.Error())
		}

		return nil, fmt.Errorf("%w:\n%s", errValidationFailed, strings.Join(msgs, "\n"))
	}

	return m, nil
}

// exportView renders one view in the requested format and writes it with the
// CLI's writer conventions, returning the written file paths.
func exportView(
	m *parser.Model, basename string, opts RenderOptions, format string, writer *output.Writer,
) ([]string, error) {
	unitPath := opts.Target

	var g *graph.Graph

	switch {
	case opts.AllExpanded:
		v := view.GenerateExpandedView(m)
		if v == nil {
			return nil, errGenerateView
		}

		g = graph.BuildExpandedGraph(v)
		unitPath = "" // WriteExpanded handles naming
	default:
		v := viewForTarget(m, unitPath)
		if v == nil {
			return nil, errGenerateView
		}

		g = graph.BuildGraphWithPath(v, unitPath, basename, format)
	}

	if g == nil {
		return nil, errBuildGraph
	}

	// The same render entry points the CLI uses (root.go renderView), so the
	// artifacts are byte-identical for the same model and options.
	switch {
	case opts.AllExpanded:
		return writeExpanded(g, basename, format, writer)
	case format == FormatPNG:
		return writePNGView(g, basename, unitPath, writer)
	default:
		data, rerr := renderExport(g, format)
		if rerr != nil {
			return nil, rerr
		}

		if werr := writer.Write(basename, unitPath, format, data); werr != nil {
			return nil, fmt.Errorf("write: %w", werr)
		}

		return []string{writtenPath(basename, unitPath, format)}, nil
	}
}

// viewForTarget generates the view for a render target: "" → C1, one segment
// → that unit's C2, deeper → C3 (the CLI/LSP dispatch rule).
func viewForTarget(m *parser.Model, target string) *view.View {
	switch {
	case target == "":
		return view.GenerateC1View(m)
	case isC2Target(target):
		return view.GenerateC2View(m, target)
	default:
		return view.GenerateC3View(m, target)
	}
}

// renderExport dispatches to the format-specific render function (the CLI's
// switch, minus the png special case handled above).
func renderExport(g *graph.Graph, format string) ([]byte, error) {
	var (
		data []byte
		err  error
	)

	switch format {
	case FormatSVG:
		data, err = render.RenderSVG(g)
	case FormatHTML:
		data, err = render.RenderHTML(g)
	default:
		data, err = render.Render(g, format)
	}

	if err != nil {
		return nil, fmt.Errorf("render: %w", err)
	}

	return data, nil
}

// writePNGView mirrors the CLI's writePNGView: the raster plus its sibling
// HTML navigation doc.
func writePNGView(g *graph.Graph, basename, unitPath string, writer *output.Writer) ([]string, error) {
	pngData, err := render.RenderPNG(g)
	if err != nil {
		return nil, fmt.Errorf("render: %w", err)
	}

	htmlDoc, err := render.RenderHTMLForPNG(g, output.PNGImageName(basename, unitPath))
	if err != nil {
		return nil, fmt.Errorf("render: %w", err)
	}

	if err := writer.Write(basename, unitPath, FormatPNG, pngData); err != nil {
		return nil, fmt.Errorf("write: %w", err)
	}

	if err := writer.Write(basename, unitPath, FormatHTML, htmlDoc); err != nil {
		return nil, fmt.Errorf("write: %w", err)
	}

	return []string{writtenPath(basename, unitPath, FormatPNG), writtenPath(basename, unitPath, FormatHTML)}, nil
}

// writeExpanded mirrors the CLI's expanded output naming
// ({basename}.expanded.{ext}; PNG gains its sibling HTML nav doc).
func writeExpanded(g *graph.Graph, basename, format string, writer *output.Writer) ([]string, error) {
	if format == FormatPNG {
		return writeExpandedPNG(g, basename, writer)
	}

	data, err := renderExport(g, format)
	if err != nil {
		return nil, err
	}

	if err := writer.WriteExpanded(basename, format, data); err != nil {
		return nil, fmt.Errorf("write: %w", err)
	}

	return []string{basename + ".expanded." + format}, nil
}

// writeExpandedPNG is writePNGView for the --expanded single diagram, per
// the CLI's writeExpandedPNGView.
func writeExpandedPNG(g *graph.Graph, basename string, writer *output.Writer) ([]string, error) {
	pngData, err := render.RenderPNG(g)
	if err != nil {
		return nil, fmt.Errorf("render: %w", err)
	}

	htmlDoc, err := render.RenderHTMLForPNG(g, output.ExpandedPNGImageName(basename))
	if err != nil {
		return nil, fmt.Errorf("render: %w", err)
	}

	if err := writer.WriteExpanded(basename, FormatPNG, pngData); err != nil {
		return nil, fmt.Errorf("write: %w", err)
	}

	if err := writer.WriteExpanded(basename, FormatHTML, htmlDoc); err != nil {
		return nil, fmt.Errorf("write: %w", err)
	}

	return []string{basename + ".expanded." + FormatPNG, basename + ".expanded." + FormatHTML}, nil
}

// writtenPath recomputes the writer's relative path for reporting (mirror of
// output.Writer.Write's layout).
func writtenPath(basename, unitPath, format string) string {
	if unitPath == "" {
		return fmt.Sprintf("%s.%s", basename, format)
	}

	return basename + "/" + strings.ReplaceAll(unitPath, ".", "/") + "." + format
}

// validExportFormat reports whether format is one of the CLI's -f values.
func validExportFormat(format string) bool {
	switch format {
	case FormatSVG, FormatHTML, FormatDOT, FormatPNG, FormatPlantUML:
		return true
	default:
		return false
	}
}

// isC2Target reports a one-segment target (a unit's C2 diagram) — the LSP's
// own dispatch rule.
func isC2Target(target string) bool {
	return !strings.Contains(target, ".")
}

// parseByExt is the extension-based front-end dispatch (D-27, mirroring the
// CLI and the LSP).
func parseByExt(path string, data []byte) (*parser.Model, error) {
	switch ext := filepath.Ext(path); {
	case strings.EqualFold(ext, modelExtToml):
		return parser.Parse(data) //nolint:wrapcheck // stageDiagnostic-style callers attach the stage prefix
	case strings.EqualFold(ext, modelExtC4d):
		return c4d.ParseNamed(path, data) //nolint:wrapcheck // same: the error rides the export stage wrap
	default:
		return nil, fmt.Errorf("%w %q (accepted: %s, %s)", ErrNotModelFile, ext, modelExtToml, modelExtC4d)
	}
}
