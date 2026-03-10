// Package main provides the CLI entry point for c4drill.
package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

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
		Version:     version,
		Args:        cobra.ExactArgs(1),
		RunE:        runRoot,
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
	// Validate flags early (before file I/O)
	if format != "dot" && format != "svg" {
		return fmt.Errorf("invalid format %q: must be dot or svg", format)
	}

	inputPath := args[0]

	// TODO: Implement full pipeline in next step
	// For now, return nil to test flag parsing
	_ = inputPath

	return nil
}
