// diagnostics.go converts the CLI composition pipeline's outcomes into LSP
// diagnostics. The pass threading mirrors cmd/c4drill/root.go's runRoot
// exactly — include.Resolve → template.Expand → peer.Resolve →
// validator.Validate — so published messages and line numbers match
// `c4drill <file>` message-for-message and line-for-line:
//
//   - parse stage failures   → one diagnostic: "parse: <err>", line from the
//     ParseError (1-based → 0-based for LSP);
//   - include stage failures → "include: <err>" (same ParseError shapes);
//   - expand stage failures  → "expand: <err>";
//   - peer resolution        → "resolve peers: <err>";
//   - validation             → one diagnostic per validator error, text
//     exactly the CLI's "error: ..." line (the validator reports paths, not
//     line numbers, so those diagnostics anchor at line 0 — the CLI prints
//     no line for them either).
//
// Humanize is not a separate CLI stage: it runs inside view generation at
// render time, after validation, so there is nothing to mirror here.

package lsp

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Djarvur/c4drill/internal/c4d"
	"github.com/Djarvur/c4drill/internal/include"
	"github.com/Djarvur/c4drill/internal/parser"
	"github.com/Djarvur/c4drill/internal/peer"
	"github.com/Djarvur/c4drill/internal/template"
	"github.com/Djarvur/c4drill/internal/validator"
)

// diagnosticSource attributes every published diagnostic.
const diagnosticSource = "c4drill"

// Accepted input extensions (D-27) — the same dispatch the CLI uses.
const (
	extToml = ".toml"
	extC4d  = ".c4d"
)

// errUnsupportedExt mirrors the CLI's unsupported-extension error so the
// diagnostic text matches `c4drill <file>` for unknown extensions too.
var errUnsupportedExt = errors.New("unsupported input extension")

// parseByExt parses document bytes through the front-end its extension
// selects (D-27): .toml → TOML front-end, .c4d → C4D front-end. The C4D
// front-end is given the file path so parse errors carry it — identical to
// what `c4drill <file>` prints. Unknown extensions fail closed, mirroring
// the CLI's parseInput.
func parseByExt(path string, data []byte) (*parser.Model, error) {
	switch ext := filepath.Ext(path); ext {
	case extToml:
		return parser.Parse(data) //nolint:wrapcheck // stageDiagnostic attaches the parse: stage prefix
	case extC4d:
		return c4d.ParseNamed(path, data) //nolint:wrapcheck // stageDiagnostic attaches the parse: stage prefix
	default:
		return nil, fmt.Errorf("%w %q (accepted: %s, %s)", errUnsupportedExt, ext, extToml, extC4d)
	}
}

// computeDiagnostics runs the full pipeline over one document and returns
// its diagnostics (empty for a clean model).
func (s *Server) computeDiagnostics(doc *document) []Diagnostic {
	m, err := parseByExt(doc.Path, doc.Text)
	if err != nil {
		return []Diagnostic{stageDiagnostic("parse: ", err)}
	}

	// Include resolution reads included files through the buffer overlay so
	// unsaved edits in open documents are what includers see; the entry
	// document itself is already the open buffer.
	m, err = include.ResolveWithReader(m, filepath.Dir(doc.Path), doc.Path, s.overlayRead())
	if err != nil {
		return []Diagnostic{stageDiagnostic("include: ", err)}
	}

	m, err = template.Expand(m)
	if err != nil {
		return []Diagnostic{stageDiagnostic("expand: ", err)}
	}

	if err := peer.Resolve(m); err != nil {
		return []Diagnostic{stageDiagnostic("resolve peers: ", err)}
	}

	valErrors := validator.Validate(m)

	diags := make([]Diagnostic, 0, len(valErrors))
	for _, ve := range valErrors {
		diags = append(diags, Diagnostic{
			Range:    Range{},
			Severity: SeverityError,
			Source:   diagnosticSource,
			Message:  ve.Error(),
		})
	}

	return diags
}

// stageDiagnostic builds the single diagnostic a failed pipeline stage
// produces: the CLI's stage-prefixed message text anchored at the error's
// line when the error carries one (ParseError.Line, 1-based), else line 0.
func stageDiagnostic(stage string, err error) Diagnostic {
	line := uint32(lineForLSP(parseErrorLine(err))) //nolint:gosec // lineForLSP is never negative

	return Diagnostic{
		Range:    Range{Start: Position{Line: line}},
		Severity: SeverityError,
		Source:   diagnosticSource,
		Message:  stage + err.Error(),
	}
}

// parseErrorLine extracts the 1-based line a *parser.ParseError carries
// (0 when the error has no line information — most pipeline-stage errors
// attribute by file path instead).
func parseErrorLine(err error) int {
	var perr *parser.ParseError
	if errors.As(err, &perr) {
		return perr.Line
	}

	return 0
}

// lineForLSP converts a 1-based source line to the LSP's 0-based form; a
// non-positive input (no line) maps to 0.
func lineForLSP(line int) int {
	if line <= 0 {
		return 0
	}

	return line - 1
}

// overlayRead returns the reader include.Resolve uses for included-file
// bytes: open documents win (their unsaved buffer is the source of truth),
// disk fills in everything else. Keys are canonical paths, matching the
// resolver's own canonicalization.
func (s *Server) overlayRead() func(path string) ([]byte, error) {
	return func(path string) ([]byte, error) {
		if doc, ok := s.docByPath(path); ok {
			return doc.Text, nil
		}

		//nolint:gosec // G304: the resolver-controlled include path is the read target by design
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err //nolint:wrapcheck // include.Resolve attributes read failures itself
		}

		return data, nil
	}
}
