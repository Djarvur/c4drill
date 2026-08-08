// Package main provides the CLI entry point for c4drill.
package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Djarvur/c4drill/internal/graph"
	"github.com/Djarvur/c4drill/internal/model"
	"github.com/Djarvur/c4drill/internal/output"
	"github.com/Djarvur/c4drill/internal/parser"
	"github.com/Djarvur/c4drill/internal/peer"
	"github.com/Djarvur/c4drill/internal/render"
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
)

const (
	formatDot  = "dot"
	formatSVG  = "svg"
	formatHTML = "html"
)

//nolint:gochecknoglobals // Cobra flags require package-level variables for PersistentFlags registration
var (
	format    string
	outputDir string
	expanded   bool
	labelRatio float64
	version    = "dev"
)

// NewRootCmd creates the root command for c4drill.
// It configures flags, validation, and the main execution function.
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "c4drill <input.toml>",
		Short: "Generate C4 architecture diagrams from TOML",
		Long: `Generate C4 architecture diagrams from TOML definitions.

Supports C1 (Context), C2 (Container), and C3 (Component) level diagrams
with drill-down navigation between levels.

Examples:
  c4drill architecture.toml
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
	cmd.PersistentFlags().Float64Var(&labelRatio, "label-ratio", 0,
		"Width:height ratio for unit labels (default: 1.6, credit card proportions)")

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
	outDir := outputDir
	if outDir == "" {
		outDir = filepath.Dir(inputPath)
	}

	// Stage 1: Parse
	m, err := parser.ParseFile(inputPath)
	if err != nil {
		return fmt.Errorf("parse: %w", err)
	}

	// Stage 1.6: Resolve relative peers (Phase 30; runs after future
	// template.Expand, before humanize + Validate). Rewrites every bare
	// Link.Peer to an absolute dotted path so the validator sees only
	// absolute paths. A miss at root is a hard error naming the peer + host.
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
	basename := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
	if basename == "" {
		basename = "diagram"
	}

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
		paths = append(paths, collectExpandableUnitPaths(m, name, unit)...)
	}

	return paths
}

// collectExpandableUnitPaths recursively collects paths of units that should have sub-diagrams.
// A unit gets a sub-diagram if it has subunits (auto-detect).
func collectExpandableUnitPaths(m *parser.Model, parentPath string, unit *model.Unit) []string {
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
			paths = append(paths, collectExpandableUnitPaths(m, subPath, subUnit)...)
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
