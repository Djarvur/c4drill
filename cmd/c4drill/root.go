// Package main provides the CLI entry point for c4drill.
package main

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Djarvur/c4drill/internal/graph"
	"github.com/Djarvur/c4drill/internal/model"
	"github.com/Djarvur/c4drill/internal/output"
	"github.com/Djarvur/c4drill/internal/parser"
	"github.com/Djarvur/c4drill/internal/render"
	"github.com/Djarvur/c4drill/internal/validator"
	"github.com/Djarvur/c4drill/internal/view"
	"github.com/spf13/cobra"
)

// Static errors for better error handling.
var (
	errInvalidFormat    = errors.New("invalid format: must be dot or svg")
	errValidationFailed = errors.New("validation failed")
	errGenerateView     = errors.New("failed to generate view")
	errBuildGraph       = errors.New("failed to build graph")
)

//nolint:gochecknoglobals // Cobra flags require package-level variables for PersistentFlags registration
var (
	format    string
	outputDir string
	version   = "dev"
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

	cmd.PersistentFlags().StringVarP(&format, "format", "f", "svg",
		"Output format (dot|svg)")
	cmd.PersistentFlags().StringVarP(&outputDir, "output", "o", ".",
		"Output directory")

	return cmd
}

// runRoot is the main execution function for the root command.
// It validates flags early, then orchestrates the full pipeline.
func runRoot(cmd *cobra.Command, args []string) error {
	// Show help if no input file provided
	if len(args) == 0 {
		return cmd.Help()
	}

	// Validate flags early (before file I/O)
	if format != "dot" && format != "svg" {
		return fmt.Errorf("%w: %q", errInvalidFormat, format)
	}

	inputPath := args[0]

	// Stage 1: Parse
	m, err := parser.ParseFile(inputPath)
	if err != nil {
		return fmt.Errorf("parse: %w", err)
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
	writer := output.NewWriter(outputDir)

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
// Always includes "" for C1, plus paths for expanded systems/containers.
func collectExpandedPaths(m *parser.Model) []string {
	// Preallocate with capacity for C1 + expanded units
	paths := make([]string, 0, 1+len(m.Units))
	paths = append(paths, "") // Always include C1

	// Recursively find expanded units
	for name, unit := range m.Units {
		paths = append(paths, collectExpandedUnitPaths(name, unit)...)
	}

	return paths
}

// collectExpandedUnitPaths recursively collects paths of expanded units.
func collectExpandedUnitPaths(parentPath string, unit *model.Unit) []string {
	var paths []string

	// Check if this unit is expanded (has subunits shown)
	// A unit is expanded if it has subunits AND its own name is in its Expanded list
	if len(unit.Subunits) > 0 && len(unit.Expanded) > 0 {
		// Check if the parent path itself is in the expanded list
		// For top-level units, parentPath is just the name (e.g., "mainapp")
		for _, expanded := range unit.Expanded {
			if expanded == parentPath || expanded == "" {
				paths = append(paths, parentPath)

				break
			}
		}

		// Recurse into subunits
		for subName, subUnit := range unit.Subunits {
			subPath := parentPath + "." + subName
			paths = append(paths, collectExpandedUnitPaths(subPath, subUnit)...)
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

	// Render
	data, err := render.Render(g, format)
	if err != nil {
		return fmt.Errorf("render: %w", err)
	}

	// Write
	if err := writer.Write(basename, unitPath, format, data); err != nil {
		return fmt.Errorf("write: %w", err)
	}

	return nil
}

// isC2Path checks if the path refers to a C2 (container) level unit.
func isC2Path(path string) bool {
	// C2 if the path has only one segment (no dots)
	return !strings.Contains(path, ".")
}
