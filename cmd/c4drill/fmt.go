package main

// fmt subcommand (Plan 35-08): gofmt-style formatter for BOTH authoring
// formats (D-31). .c4d files format through the trivia-aware AST and the
// EmitC4D canonical printer (D-32/D-33 — comments ride the AST; fmt is NOT a
// Model re-emit), .toml files through internal/tomlfmt (position-aware,
// comment-preserving). Files rewrite IN PLACE; --check reports misformatted
// files without writing and exits 1 — the CI gate (T-35-08-04).
//
// SAFETY GATE (T-35-08-01): before any in-place rewrite, the candidate
// output must re-parse to a Model exactly equal to the ORIGINAL file's —
// fmt can never corrupt a file past the semantic gate. A rewrite that would
// break parsing or change semantics is a hard error and the file is left
// untouched.

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"

	"github.com/Djarvur/c4drill/internal/c4d"
	"github.com/Djarvur/c4drill/internal/parser"
	"github.com/Djarvur/c4drill/internal/tomlfmt"
	"github.com/spf13/cobra"
)

// Static errors for better error handling.
var (
	errFmtNoTargets   = errors.New("no .toml or .c4d files found")
	errFmtGate        = errors.New("formatted output failed the semantic safety gate")
	errFmtCheckNeeded = errors.New("files need formatting")
)

// reparseFunc re-parses formatted bytes into a Model for the safety gate —
// c4d.Parse for .c4d candidates, parser.Parse for .toml candidates.
type reparseFunc func([]byte) (*parser.Model, error)

//nolint:gochecknoglobals // Cobra flags require package-level variables (root.go precedent)
var fmtCheck bool

// newFMTCmd builds the fmt command.
func newFMTCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fmt [--check] <file|dir>...",
		Short: "Format TOML and C4D diagram files in place (gofmt-style)",
		Long: `Format diagram files in place, gofmt-style.

Each .c4d file re-emits through the canonical C4D printer (comments
preserved, compact one-line leaf blocks); each .toml file normalizes
whitespace, indentation and blank-line grouping with comments preserved.
Key order is the author's in both formats — fmt never reorders.

Arguments may be files or directories; directories walk recursively,
formatting every *.c4d and *.toml found (other files are ignored).

Use --check to report misformatted files without writing anything: the
offending paths print one per line and the command exits 1 (CI gate).`,
		Args:         cobra.MinimumNArgs(1),
		RunE:         runFMT,
		SilenceUsage: true,
	}

	cmd.Flags().BoolVar(&fmtCheck, "check", false,
		"Report misformatted files without writing; exit 1 if any")

	return cmd
}

// runFMT expands the file/dir arguments, formats each target and, in
// --check mode, lists the differing files and fails (T-35-08-04).
func runFMT(cmd *cobra.Command, args []string) error {
	files, err := expandFMTArgs(args)
	if err != nil {
		return err
	}

	if len(files) == 0 {
		return fmt.Errorf("fmt: %w", errFmtNoTargets)
	}

	var differing []string

	for _, path := range files {
		dirty, err := formatFile(path)
		if err != nil {
			return err
		}

		if dirty {
			differing = append(differing, path)
		}
	}

	if fmtCheck && len(differing) > 0 {
		for _, p := range differing {
			if _, err := fmt.Fprintln(cmd.OutOrStdout(), p); err != nil {
				return fmt.Errorf("fmt: report %s: %w", p, err)
			}
		}

		return fmt.Errorf("fmt: %w: %d file(s)", errFmtCheckNeeded, len(differing))
	}

	return nil
}

// expandFMTArgs resolves the argument list to diagram files (a file is used
// directly, a directory walks recursively — T-35-08-02).
func expandFMTArgs(args []string) ([]string, error) {
	var files []string

	for _, arg := range args {
		collected, err := expandFMTArg(arg)
		if err != nil {
			return nil, err
		}

		files = append(files, collected...)
	}

	return files, nil
}

// expandFMTArg resolves ONE argument: a file is used directly (its
// extension must be .toml or .c4d — fmt cannot be pointed at arbitrary file
// types, T-35-08-02), a directory walks recursively collecting matching
// extensions only.
func expandFMTArg(arg string) ([]string, error) {
	info, err := os.Stat(arg)
	if err != nil {
		return nil, fmt.Errorf("fmt: stat %s: %w", arg, err)
	}

	if !info.IsDir() {
		if ext := filepath.Ext(arg); ext != extToml && ext != extC4d {
			return nil, fmt.Errorf("fmt: %w %q (accepted: %s, %s)",
				errUnsupportedExt, ext, extToml, extC4d)
		}

		return []string{arg}, nil
	}

	files, err := walkFMTDir(arg)
	if err != nil {
		return nil, fmt.Errorf("fmt: walk %s: %w", arg, err)
	}

	return files, nil
}

// walkFMTDir collects every *.c4d and *.toml under dir via filepath.WalkDir
// (which never follows symlinked directories — T-35-08-02).
func walkFMTDir(dir string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		switch {
		case err != nil:
			return err
		case d.IsDir():
			return nil
		}

		if ext := filepath.Ext(path); ext == extToml || ext == extC4d {
			files = append(files, path)
		}

		return nil
	})
	if err != nil {
		return nil, err //nolint:wrapcheck // expanded by the caller with the walk prefix
	}

	return files, nil
}

// formatFile formats one diagram file: parse the original (the safety
// gate's baseline Model), build the candidate output (.c4d via ParseAST +
// EmitC4D — comments ride the AST; .toml via tomlfmt.Format), then hand
// both to the gate. Returns whether the file's content differs from the
// candidate.
func formatFile(path string) (bool, error) {
	//nolint:gosec // G304: caller-provided path, intentional for CLI tool
	data, err := os.ReadFile(path)
	if err != nil {
		return false, fmt.Errorf("fmt: read %s: %w", path, err)
	}

	var (
		formatted []byte
		orig      *parser.Model
		reparse   reparseFunc
	)

	switch ext := filepath.Ext(path); ext {
	case extC4d:
		// Gate baseline: the original must survive its own Model parse
		// (grammar-valid but Model-refused documents are hard errors).
		orig, err = c4d.Parse(data)
		if err != nil {
			return false, fmt.Errorf("fmt: parse %s: %w", path, err)
		}

		// Emission AST: a SEPARATE parse — the Model conversion and the
		// emitter never share a doc, mirroring convert's two-parse rule.
		doc, astErr := c4d.ParseAST(data)
		if astErr != nil {
			return false, fmt.Errorf("fmt: parse %s: %w", path, astErr)
		}

		formatted = []byte(c4d.EmitC4D(doc))
		reparse = c4d.Parse
	case extToml:
		orig, err = parser.Parse(data)
		if err != nil {
			return false, fmt.Errorf("fmt: parse %s: %w", path, err)
		}

		formatted, err = tomlfmt.Format(data)
		if err != nil {
			return false, fmt.Errorf("fmt: format %s: %w", path, err)
		}

		reparse = parser.Parse
	default:
		return false, fmt.Errorf("fmt: %w %q (accepted: %s, %s)",
			errUnsupportedExt, ext, extToml, extC4d)
	}

	return applyFormatted(path, data, formatted, orig, reparse)
}

// applyFormatted runs the T-35-08-01 safety gate, then either reports the
// file as differing (--check — nothing is written) or performs the in-place
// rewrite. The gate: the candidate output must re-parse (through reparse)
// to a Model exactly equal to the original file's — a rewrite that would
// break parsing or change semantics is a hard error and the file is left
// untouched.
func applyFormatted(
	path string,
	data, formatted []byte,
	orig *parser.Model,
	reparse reparseFunc,
) (bool, error) {
	gated, err := reparse(formatted)
	if err != nil {
		return false, fmt.Errorf("fmt: %s: %w: %w", path, errFmtGate, err)
	}

	if !reflect.DeepEqual(orig, gated) {
		return false, fmt.Errorf("fmt: %s: %w", path, errFmtGate)
	}

	if bytes.Equal(data, formatted) {
		return false, nil
	}

	if fmtCheck {
		return true, nil
	}

	//nolint:gosec // G306: 0644 is the documented fmt output permission
	if err := os.WriteFile(path, formatted, 0o644); err != nil {
		return false, fmt.Errorf("fmt: write %s: %w", path, err)
	}

	return true, nil
}
