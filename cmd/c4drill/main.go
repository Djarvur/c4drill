// Package main provides the CLI entry point for c4drill.
// This is a minimal CLI for Phase 2 that parses and validates TOML files.
// Full CLI with flags and help comes in Phase 6.
package main

import (
	"fmt"
	"os"

	"github.com/Djarvur/c4drill/internal/parser"
	"github.com/Djarvur/c4drill/internal/validator"
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
	os.Exit(exitCode)
}
