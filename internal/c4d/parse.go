package c4d

import (
	"errors"
	"fmt"
	"os"

	"github.com/Djarvur/c4drill/internal/c4d/ast"
	"github.com/Djarvur/c4drill/internal/c4d/grammar"
	"github.com/Djarvur/c4drill/internal/parser"
)

// Parse parses C4D data into a *parser.Model (D-21) — the DSL parses
// DIRECTLY to the Model hub, so include.Resolve, template.Expand,
// peer.Resolve and validator.Validate run unchanged downstream. All
// failures (grammar and conversion) are *parser.ParseError with DSL-native
// line numbers, so callers handle C4D input exactly like TOML input.
//
// Parse composes the exported AST entry: ParseAST then ToModel.
func Parse(data []byte) (*parser.Model, error) {
	return ParseNamed("", data)
}

// ParseNamed is Parse with error attribution: grammar errors carry name in
// their rendered message (pigeon prefixes the parser name) and conversion
// errors ride it in ParseError.Context — exactly what ParseFile produces
// for the same bytes, minus the disk read. Callers that hold file content
// from a non-disk source (the LSP's open-editor buffers) use it so their
// diagnostics stay message-for-message identical to the CLI.
func ParseNamed(name string, data []byte) (*parser.Model, error) {
	doc, err := parse(name, data)
	if err != nil {
		return nil, err
	}

	m, err := ToModel(doc)
	if err != nil {
		// Attribution: conversion errors carry the statement line; the
		// caller-supplied name rides in Context when the converter did not
		// set one (empty name adds nothing, mirroring Parse).
		var perr *parser.ParseError

		if name != "" && errors.As(err, &perr) && perr.Context == "" {
			perr.Context = name
		}

		return nil, err
	}

	return m, nil
}

// ParseFile reads a C4D file and parses it into a *parser.Model (D-21).
//
//nolint:gosec // G304: Path is provided by caller, this is intentional for CLI tool
func ParseFile(path string) (*parser.Model, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, &parser.ParseError{Message: "failed to read file", Context: path, Cause: err}
	}

	return ParseNamed(path, data)
}

// ParseAST parses C4D data into the typed AST document — the exported
// AST-level entry (D-21/D-32). Cross-package consumers that need positions
// and comments rather than the Model (cmd/c4drill fmt in 35-08,
// internal/testutil canonsrc in 35-06) compile against THIS function; Parse
// is the Model-level composition of it.
func ParseAST(data []byte) (*ast.Document, error) {
	return parse("", data)
}

// ParseASTFile reads a C4D file into the typed AST document (the exported
// file-level AST entry — see ParseAST).
//
//nolint:gosec // G304: Path is provided by caller, this is intentional for CLI tool
func ParseASTFile(path string) (*ast.Document, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, &parser.ParseError{Message: "failed to read file", Context: path, Cause: err}
	}

	return parse(path, data)
}

// maxExpressions caps pigeon expression evaluations so pathological input
// fails with an error instead of hanging the parser (T-35-01-01).
const maxExpressions = 1_000_000

// parse runs the generated grammar with the fixed parser options and
// asserts the untyped result to the typed AST.
func parse(name string, data []byte) (*ast.Document, error) {
	result, err := grammar.Parse(name, data,
		grammar.Memoize(true),
		grammar.MaxExpressions(maxExpressions),
	)
	if err != nil {
		return nil, wrapPigeonError(err, name)
	}

	doc, ok := result.(*ast.Document)
	if !ok {
		return nil, &parser.ParseError{
			Message: fmt.Sprintf("internal error: unexpected parse result type %T", result),
			Context: name,
		}
	}

	return doc, nil
}
