// Package main provides the CLI entry point for c4drill.
// This is a simple development testing CLI for Phase 1.
// Full CLI with flags and help comes in Phase 6.
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/Djarvur/c4drill/internal/model"
	"github.com/Djarvur/c4drill/internal/parser"
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
		fmt.Fprintf(os.Stderr, "Error parsing %s: %v\n", path, err)
		os.Exit(1)
	}

	// Print model for verification (simple debug output)
	//nolint:forbidigo // Debug output for Phase 1 CLI
	fmt.Printf("Properties: %s\n", model.Properties.Name)

	//nolint:forbidigo // Debug output for Phase 1 CLI
	fmt.Printf("Units: %d\n", len(model.Units))

	for name, unit := range model.Units {
		printUnit(name, unit, 0)
	}
}

// printUnit recursively prints a unit and its subunits with indentation.
func printUnit(name string, unit *model.Unit, depth int) {
	var sb strings.Builder

	for range depth {
		sb.WriteString("  ")
	}

	indent := sb.String()

	//nolint:forbidigo // Debug output for Phase 1 CLI
	fmt.Printf("%s- %s: %s (%s)\n", indent, name, unit.Name, unit.Type)

	if len(unit.Subunits) > 0 {
		for subname, subunit := range unit.Subunits {
			printUnit(subname, subunit, depth+1)
		}
	}
}
