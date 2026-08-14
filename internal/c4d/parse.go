package c4d

import (
	"fmt"
	"os"

	"github.com/Djarvur/c4drill/internal/c4d/ast"
	"github.com/Djarvur/c4drill/internal/c4d/grammar"
	"github.com/Djarvur/c4drill/internal/parser"
)

// Parse parses C4D data into a typed AST document. All failures are
// *parser.ParseError with DSL-native line numbers (D-21), so callers handle
// C4D input exactly like TOML input.
func Parse(data []byte) (*ast.Document, error) {
	return parse("", data)
}

// ParseFile reads a C4D file and parses it into a typed AST document.
//
//nolint:gosec // G304: Path is provided by caller, this is intentional for CLI tool
func ParseFile(path string) (*ast.Document, error) {
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
