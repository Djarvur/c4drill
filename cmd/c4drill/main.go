// Package main provides the CLI entry point for c4drill.
// This is a minimal CLI for Phase 3 that parses, validates, and generates views/graphs.
// Full CLI with flags and help comes in Phase 6.
package main

import (
	"fmt"
	"os"

	"github.com/Djarvur/c4drill/internal/graph"
	"github.com/Djarvur/c4drill/internal/parser"
	"github.com/Djarvur/c4drill/internal/validator"
	"github.com/Djarvur/c4drill/internal/view"
)

const minArgs = 2

func main() {
	if len(os.Args) < minArgs {
		fmt.Fprintln(os.Stderr, "Usage: c4drill <input.toml>")
		os.Exit(1)
	}

	path := os.Args[1]

	model, err := parser.ParseFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	errors := validator.Validate(model)

	exitCode := validator.ReportErrors(errors, os.Stderr)
	if exitCode != 0 {
		os.Exit(exitCode)
	}

	// Generate C1 view and build graph
	// This demonstrates the view-to-graph pipeline works correctly
	// Full rendering output comes in Phase 4
	c1View := view.GenerateC1View(model)
	if c1View == nil {
		fmt.Fprintln(os.Stderr, "error: failed to generate C1 view")
		os.Exit(1)
	}

	_, _ = fmt.Fprintf(os.Stdout, "Generated C1 view: %q with %d units\n", c1View.Title, len(c1View.Units))

	g := graph.BuildGraph(c1View)
	if g == nil {
		fmt.Fprintln(os.Stderr, "error: failed to build graph")
		os.Exit(1)
	}

	_, _ = fmt.Fprintf(os.Stdout, "Built graph: %d nodes, %d edges, %d clusters\n",
		len(g.Nodes), len(g.Edges), len(g.Clusters))
}
