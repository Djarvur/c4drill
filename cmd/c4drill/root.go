// Package main provides the CLI entry point for c4drill.
package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Djarvur/c4drill/internal/c4d"
	"github.com/Djarvur/c4drill/internal/graph"
	"github.com/Djarvur/c4drill/internal/include"
	"github.com/Djarvur/c4drill/internal/model"
	"github.com/Djarvur/c4drill/internal/output"
	"github.com/Djarvur/c4drill/internal/parser"
	"github.com/Djarvur/c4drill/internal/peer"
	"github.com/Djarvur/c4drill/internal/render"
	"github.com/Djarvur/c4drill/internal/template"
	"github.com/Djarvur/c4drill/internal/validator"
	"github.com/Djarvur/c4drill/internal/view"
	"github.com/spf13/cobra"
)

// Static errors for better error handling.
var (
	errInvalidFormat    = errors.New("invalid format: must be dot, svg, or html")
	errValidationFailed = errors.New("validation failed")
	errGenerateView     = errors.New("failed to generate view")
	errBuildGraph       = errors.New("failed to build graph")
	errUnsupportedExt   = errors.New("unsupported input extension")
)

const (
	formatDot  = "dot"
	formatSVG  = "svg"
	formatHTML = "html"
)

// Accepted input extensions (D-27): dispatch is extension-based and fails
// closed — .toml -> TOML front-end, .c4d -> C4D front-end, anything else is a
// hard parse error naming the accepted extensions (no content sniffing).
const (
	extToml = ".toml"
	extC4d  = ".c4d"
)

//nolint:gochecknoglobals // Cobra flags require package-level variables for PersistentFlags registration
var (
	format     string
	outputDir  string
	expanded   bool
	plain      bool
	noColors   bool
	noStyles   bool
	noLength   bool
	noRank     bool
	labelRatio float64
	version    = "dev"
)

// NewRootCmd creates the root command for c4drill.
// It configures flags, validation, and the main execution function.
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "c4drill <input.toml|input.c4d>",
		Short: "Generate C4 architecture diagrams from TOML and C4D",
		Long: `Generate C4 architecture diagrams from TOML or C4D definitions.

Input dispatch is by extension (D-27): .toml files parse through the TOML
front-end, .c4d files through the C4D DSL front-end — both render directly
through the full pipeline (C4D is a first-class authoring format, D-29).
Any other extension is a hard error.

Supports C1 (Context), C2 (Container), and C3 (Component) level diagrams
with drill-down navigation between levels.

Examples:
  c4drill architecture.toml
  c4drill architecture.c4d
  c4drill architecture.toml -o ./docs/diagrams
  c4drill architecture.toml -f dot -o ./output

Output:
  - C1 diagram: {basename}.{format}
  - C2 diagrams: {basename}/{system}.{format}
  - C3 diagrams: {basename}/{system}/{container}.{format}`,
		Version:      version,
		Args:         cobra.MaximumNArgs(1),
		RunE:         runRoot,
		SilenceUsage: true,
	}

	cmd.PersistentFlags().StringVarP(&format, "format", "f", formatSVG,
		"Output format (dot|svg|html)")
	cmd.PersistentFlags().StringVarP(&outputDir, "output", "o", "",
		"Output directory (default: same as input file)")
	cmd.PersistentFlags().BoolVar(&expanded, "expanded", false,
		"Generate all-expanded diagram showing all units")
	cmd.PersistentFlags().BoolVar(&plain, "plain", false,
		"Ignore author-custom formatting: default unit/edge styling, spacing and ranking, plain-text labels")
	cmd.PersistentFlags().BoolVar(&noColors, "no-colors", false,
		"Suppress colouring only: author unit/link colors and kind-derived edge colours")
	cmd.PersistentFlags().BoolVar(&noStyles, "no-styles", false,
		"Suppress line styles only: author unit/link style overrides")
	cmd.PersistentFlags().BoolVar(&noLength, "no-length", false,
		"Suppress link spacing only: link length no longer sets minlen")
	cmd.PersistentFlags().BoolVar(&noRank, "no-rank", false,
		"Suppress ranking hints only: link rank reverse/equal ignored")
	cmd.PersistentFlags().Float64Var(&labelRatio, "label-ratio", 0,
		"Width:height ratio for unit labels (default: 1.6, credit card proportions)")

	// Subcommands (Plan 35-07): convert between TOML and C4D formats.
	cmd.AddCommand(newConvertCmd())

	// Subcommands (Plan 35-08): format both authoring formats in place.
	cmd.AddCommand(newFMTCmd())

	return cmd
}

// runRoot is the main execution function for the root command.
// It validates flags early, then orchestrates the full pipeline.
func runRoot(cmd *cobra.Command, args []string) error {
	// Show help if no input file provided
	if len(args) == 0 {
		if err := cmd.Help(); err != nil {
			return fmt.Errorf("show help: %w", err)
		}

		return nil
	}

	// Validate flags early (before file I/O)
	if format != formatDot && format != formatSVG && format != formatHTML {
		return fmt.Errorf("%w: %q", errInvalidFormat, format)
	}

	// Set label ratio for word-wrapping
	render.LabelRatio = getLabelRatio()

	inputPath := args[0]

	// Default output directory to input file's directory
	outDir := resolveOutDir(inputPath)

	// Stage 1: Parse — extension dispatch (D-27): .toml -> TOML front-end,
	// .c4d -> C4D front-end so a .c4d document renders directly through the
	// rest of the pipeline unchanged (D-29). Unknown extensions fail closed
	// with a hard error naming the accepted ones — no fallback parsing.
	m, err := parseInput(inputPath)
	if err != nil {
		return fmt.Errorf("parse: %w", err)
	}

	// Stage 1a: Resolve includes (Phase 32; runs FIRST — before
	// template.Expand and Validate). Walks every [[include]] directive,
	// recursively ParseFile+merges the transitively-included files into one
	// *parser.Model per D-09/D-10/D-11/INC-08. A no-op (returns m unchanged)
	// when the model has no [[include]] — guaranteeing no regression for
	// single-file input. Pipeline ordering is load-bearing: include must run
	// before template.Expand so templates defined in included files are visible
	// to [[use]] in the entry file (XC-02).
	if m, err = include.Resolve(m, filepath.Dir(inputPath), inputPath); err != nil {
		return fmt.Errorf("include: %w", err)
	}

	// Stage 1.5: Expand templates (Phase 31; runs after Parse, before
	// peer.Resolve + Validate). Turns every [[use]] instantiation into a
	// concrete, parametrized unit subtree drained into m.Units/m.UnitOrder,
	// producing a model structurally indistinguishable from a hand-authored
	// one. A no-op (returns m unchanged) when the model has no templates —
	// guaranteeing no regression for hand-authored-only input. Pipeline
	// ordering: Parse -> template.Expand -> peer.Resolve -> Validate.
	m, err = template.Expand(m)
	if err != nil {
		return fmt.Errorf("expand: %w", err)
	}

	// Stage 1.6: Resolve relative peers (Phase 30; runs after
	// template.Expand, before Validate). Rewrites every bare Link.Peer to an
	// absolute dotted path so the validator sees only absolute paths. A miss
	// at root is a hard error naming the peer + host.
	if err := peer.Resolve(m); err != nil {
		return fmt.Errorf("resolve peers: %w", err)
	}

	// Stage 2: Validate
	valErrors := validator.Validate(m)
	if len(valErrors) > 0 {
		validator.ReportErrors(valErrors, cmd.OutOrStderr())

		return errValidationFailed
	}

	// Derive basename from input file
	basename := deriveBasename(inputPath)

	// Create output writer
	writer := output.NewWriter(outDir)

	// Handle --expanded mode (skip normal C1/C2/C3 generation)
	if expanded {
		return processExpandedView(m, basename, writer)
	}

	// Collect all expanded unit paths (C1 + expanded C2/C3)
	expandedPaths := collectExpandedPaths(m)

	// Stage 3-6: Generate all views and write
	for _, unitPath := range expandedPaths {
		if err := processView(m, unitPath, basename, writer); err != nil {
			return err
		}
	}

	return nil // Success - silent per spec
}

// parseInput parses the input file through the front-end its extension
// selects (D-27): .toml -> TOML front-end, .c4d -> C4D front-end (D-29 — a
// .c4d document parses straight into the Model hub the pipeline consumes).
// Any other extension is a hard error naming the accepted ones; the caller
// prefixes all failures with the parse stage name.
//
//nolint:wrapcheck // runRoot wraps the returned error with the parse: stage prefix
func parseInput(inputPath string) (*parser.Model, error) {
	switch ext := filepath.Ext(inputPath); ext {
	case extToml:
		return parser.ParseFile(inputPath)
	case extC4d:
		return c4d.ParseFile(inputPath)
	default:
		return nil, fmt.Errorf("%w %q (accepted: %s, %s)", errUnsupportedExt, ext, extToml, extC4d)
	}
}

// resolveOutDir defaults the output directory to the input file's directory.
func resolveOutDir(inputPath string) string {
	if outputDir != "" {
		return outputDir
	}

	return filepath.Dir(inputPath)
}

// deriveBasename strips the extension from the input file name, falling back
// to "diagram" for dotfiles with no basename.
func deriveBasename(inputPath string) string {
	basename := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
	if basename == "" {
		return "diagram"
	}

	return basename
}

// collectExpandedPaths returns all unit paths that need diagrams.
// Always includes "" for C1, plus paths for any unit that has subunits.
// Units are auto-detected (any unit with subunits gets a sub-diagram).
// This can be overridden by properties.expanded or per-unit expanded lists.
func collectExpandedPaths(m *parser.Model) []string {
	// Preallocate with capacity for C1 + expanded units
	paths := make([]string, 0, 1+len(m.Units))
	paths = append(paths, "") // Always include C1

	// Recursively find units with subunits (auto-detect expandable units)
	for name, unit := range m.Units {
		paths = append(paths, collectExpandableUnitPaths(name, unit)...)
	}

	return paths
}

// collectExpandableUnitPaths recursively collects paths of units that should have sub-diagrams.
// A unit gets a sub-diagram if it has subunits (auto-detect).
func collectExpandableUnitPaths(parentPath string, unit *model.Unit) []string {
	var paths []string

	// Auto-detect: any unit with subunits should have a sub-diagram
	if len(unit.Subunits) > 0 {
		paths = append(paths, parentPath)

		// Recurse into subunits (they may also have subunits)
		var subOrder []string
		if len(unit.SubunitOrder) > 0 {
			subOrder = unit.SubunitOrder
		} else {
			for subName := range unit.Subunits {
				subOrder = append(subOrder, subName)
			}
		}

		for _, subName := range subOrder {
			subUnit := unit.Subunits[subName]
			if subUnit == nil {
				continue
			}

			subPath := parentPath + "." + subName
			paths = append(paths, collectExpandableUnitPaths(subPath, subUnit)...)
		}
	}

	return paths
}

// processView generates a view, builds graph, renders, and writes.
func processView(m *parser.Model, unitPath, basename string, writer *output.Writer) error {
	// Generate appropriate view based on path
	var v *view.View

	switch {
	case unitPath == "":
		v = view.GenerateC1View(m)
	case isC2Path(unitPath):
		v = view.GenerateC2View(m, unitPath)
	default:
		v = view.GenerateC3View(m, unitPath)
	}

	if v == nil {
		return fmt.Errorf("%w: %q", errGenerateView, unitPath)
	}

	// PLAIN-01: thread --plain onto every generated view so the graph builder
	// suppresses author-custom formatting (PLAIN-02). KEY-01: the granular
	// switches thread the same way onto every view.
	v.Plain = plain
	v.NoColors = noColors
	v.NoStyles = noStyles
	v.NoLength = noLength
	v.NoRank = noRank

	// Build graph with navigation
	g := graph.BuildGraphWithPath(v, unitPath, basename, format)
	if g == nil {
		return fmt.Errorf("%w: %q", errBuildGraph, unitPath)
	}

	// Render with output directory for icon extraction (SVG/HTML use the same
	// graphviz SVG pipeline; HTML post-processes the bytes into a wrapper doc)
	var (
		data []byte
		err  error
	)

	switch format {
	case formatSVG:
		data, err = render.RenderSVGWithOutput(g, writer.BaseDir())
	case formatHTML:
		data, err = render.RenderHTML(g)
	default:
		data, err = render.Render(g, format)
	}

	if err != nil {
		return fmt.Errorf("render: %w", err)
	}

	// Write
	if err := writer.Write(basename, unitPath, format, data); err != nil {
		return fmt.Errorf("write: %w", err)
	}

	return nil
}

// processExpandedView generates a single all-expanded diagram showing all units.
// This is used for the --expanded flag which produces a single file with nested clusters.
func processExpandedView(m *parser.Model, basename string, writer *output.Writer) error {
	// Generate expanded view with all units at all nesting levels
	v := view.GenerateExpandedView(m)
	if v == nil {
		return fmt.Errorf("%w: expanded view", errGenerateView)
	}

	// PLAIN-01: --plain x --expanded — the expanded view gets the flag too
	// (BuildExpandedGraph copies View.Plain into Graph.Opts.Plain).
	// KEY-01: the granular switches thread the same way.
	v.Plain = plain
	v.NoColors = noColors
	v.NoStyles = noStyles
	v.NoLength = noLength
	v.NoRank = noRank

	// Build graph with nested clusters (no navigation for expanded view)
	g := graph.BuildExpandedGraph(v)
	if g == nil {
		return fmt.Errorf("%w: expanded graph", errBuildGraph)
	}

	// Render with icon extraction for SVG format
	var (
		data []byte
		err  error
	)

	if format == formatSVG {
		data, err = render.RenderSVGWithOutput(g, writer.BaseDir())
	} else {
		data, err = render.Render(g, format)
	}

	if err != nil {
		return fmt.Errorf("render: %w", err)
	}

	// Write to {basename}.expanded.{format}
	if err := writer.WriteExpanded(basename, format, data); err != nil {
		return fmt.Errorf("write: %w", err)
	}

	return nil
}

// isC2Path checks if the path refers to a C2 (container) level unit.
func isC2Path(path string) bool {
	// C2 if the path has only one segment (no dots)
	return !strings.Contains(path, ".")
}

// getLabelRatio returns the label ratio from CLI flag, env var, or default.
// Priority: CLI flag > C4DRILL_LABEL_RATIO env var > default (1.6).
func getLabelRatio() float64 {
	if labelRatio > 0 {
		return labelRatio
	}

	if envVal := os.Getenv("C4DRILL_LABEL_RATIO"); envVal != "" {
		if v, err := strconv.ParseFloat(envVal, 64); err == nil && v > 0 {
			return v
		}
	}

	return 1.6
}
