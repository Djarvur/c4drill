package main

// convert subcommand (Plan 35-07): converts diagram files between the TOML
// and C4D formats (D-28) with validate-first semantics (D-24) and
// swapped-extension output placement (D-30).
//
// The load-bearing shape is the TWO-PARSE rule: the pipeline stages up to
// Validate (include.Resolve -> template.Expand -> peer.Resolve) MUTATE the
// model in place — Includes drained to nil, templates consumed, bare peers
// absolutized — so a gated model can never be emitted losslessly. The gate
// runs on a DISCARDED copy; emission always re-parses the source fresh so
// include directives, template declarations, use instantiations and authored
// relative peers survive verbatim (D-25 single-file default, D-22 round-trip
// parity with the 35-06 parity contract).

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/Djarvur/c4drill/internal/c4d"
	"github.com/Djarvur/c4drill/internal/include"
	"github.com/Djarvur/c4drill/internal/parser"
	"github.com/Djarvur/c4drill/internal/peer"
	"github.com/Djarvur/c4drill/internal/template"
	"github.com/Djarvur/c4drill/internal/validator"
	"github.com/spf13/cobra"
)

// Static errors for better error handling.
var (
	errWrongDirection = errors.New("wrong input extension for direction")
	errConvertCycle   = errors.New("include cycle detected")
	errConvertDepth   = errors.New("include depth exceeded")
)

// Conversion directions (D-28): the direction names the TARGET format.
const (
	dirToTOML = "to-toml"
	dirToC4D  = "to-c4d"
)

//nolint:gochecknoglobals // Cobra flags require package-level variables (root.go precedent)
var (
	convertOutDir         string
	convertFollowIncludes bool
)

// maxConvertDepth caps the --follow-includes traversal as defense-in-depth
// against unbounded recursion (T-35-07-02), mirroring internal/include's
// maxIncludeDepth. The D-24 gate already rejects cyclic graphs (include's
// own cycle detection) before the walk runs, so the cap only guards against
// pathological acyclic-but-deep graphs.
const maxConvertDepth = 100

// newConvertCmd builds the convert command with one subcommand per
// direction (D-28's literal names) sharing the -o output-directory flag.
func newConvertCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "convert <to-toml|to-c4d> <file>",
		Short: "Convert diagrams between the TOML and C4D formats",
		Long: `Convert a diagram file between the TOML and C4D formats.

The direction names the target format:
  to-toml  <file.c4d>  writes <file>.toml
  to-c4d   <file.toml> writes <file>.c4d

The source is VALIDATED first (parse -> include.Resolve -> template.Expand
-> peer.Resolve -> validate, the same stage composition as the render
pipeline): an invalid model is a hard error and NO output file is written.

By default a single file converts alone — include directives, template
declarations and use instantiations are preserved verbatim in the twin (the
twin re-parses to the same model). Use --follow-includes to convert the
whole include graph, rewriting include paths to the twin extension.

Output lands next to the input with the extension swapped; -o overrides the
directory.`,
		SilenceUsage: true,
	}

	cmd.PersistentFlags().StringVarP(&convertOutDir, "output", "o", "",
		"Output directory (default: same as input file)")
	cmd.PersistentFlags().BoolVar(&convertFollowIncludes, "follow-includes", false,
		"Convert the whole include graph, rewriting include paths to the target extension")

	cmd.AddCommand(newDirectionCmd(dirToTOML, extC4d, extToml))
	cmd.AddCommand(newDirectionCmd(dirToC4D, extToml, extC4d))

	return cmd
}

// newDirectionCmd builds one direction subcommand. srcExt is the only
// accepted input extension; targetExt is the twin's extension.
func newDirectionCmd(direction, srcExt, targetExt string) *cobra.Command {
	return &cobra.Command{
		Use:   fmt.Sprintf("%s <file%s>", direction, srcExt),
		Short: fmt.Sprintf("Convert a %s diagram to %s", srcExt, targetExt),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConvert(cmd, args[0], direction, srcExt, targetExt)
		},
		SilenceUsage: true,
	}
}

// runConvert converts inputPath into the direction's target format. Order
// is load-bearing (D-24): direction gate -> validation gate -> fresh
// re-parse -> emit -> write. No output is written unless every gate passed.
func runConvert(cmd *cobra.Command, inputPath, direction, srcExt, targetExt string) error {
	// Direction gate: the input must already be in the SOURCE format.
	if got := filepath.Ext(inputPath); got != srcExt {
		return fmt.Errorf("%w: %s converts %s files, got %q",
			errWrongDirection, direction, srcExt, got)
	}

	// D-24 gate: the source must survive the render pipeline's stages up to
	// Validate. The gated model is discarded — these stages mutate in place.
	// In --follow-includes mode the gate applies to the ORIGINAL entry graph
	// as authored (D-24): include.Resolve merges the whole graph and only the
	// merged entry model is validated — included fragments may carry bare
	// peers and template fragments that are not standalone-valid, exactly as
	// the normal render pipeline treats them.
	if err := validateSourceForConvert(cmd, inputPath); err != nil {
		return err
	}

	// Graph mode (D-25): convert every file in the include graph.
	if convertFollowIncludes {
		return convertGraph(inputPath, targetExt)
	}

	// Emission model: a FRESH parse of the untouched source (D-25).
	m, err := parseInput(inputPath)
	if err != nil {
		return fmt.Errorf("parse: %w", err)
	}

	text, err := emitConverted(m, targetExt)
	if err != nil {
		return err
	}

	// D-30: next to the input with the extension swapped; -o overrides the
	// directory (created when missing).
	dst := convertOutputPath(inputPath, targetExt)

	if err := os.MkdirAll(filepath.Dir(dst), 0o750); err != nil {
		return fmt.Errorf("write: create output directory: %w", err)
	}

	//nolint:gosec // G306: 0644 is the documented convert output permission
	if err := os.WriteFile(dst, []byte(text), 0o644); err != nil {
		return fmt.Errorf("write: %w", err)
	}

	return nil
}

// validateSourceForConvert runs the D-24 convert gate: the source must
// survive the SAME stage composition the render pipeline applies up to
// Validate — include.Resolve -> template.Expand -> peer.Resolve ->
// validator.Validate (ReportErrors + hard error, no output on failure).
// Error wrapping mirrors runRoot's stage prefixes exactly.
func validateSourceForConvert(cmd *cobra.Command, inputPath string) error {
	m, err := parseInput(inputPath)
	if err != nil {
		return fmt.Errorf("parse: %w", err)
	}

	if m, err = include.Resolve(m, filepath.Dir(inputPath), inputPath); err != nil {
		return fmt.Errorf("include: %w", err)
	}

	if m, err = template.Expand(m); err != nil {
		return fmt.Errorf("expand: %w", err)
	}

	if err := peer.Resolve(m); err != nil {
		return fmt.Errorf("resolve peers: %w", err)
	}

	valErrors := validator.Validate(m)
	if len(valErrors) > 0 {
		validator.ReportErrors(valErrors, cmd.ErrOrStderr())

		return errValidationFailed
	}

	return nil
}

// emitConverted serializes the (fresh, unmutated) model into the target
// format.
func emitConverted(m *parser.Model, targetExt string) (string, error) {
	if targetExt == extToml {
		text, err := c4d.EmitTOML(m)
		if err != nil {
			return "", fmt.Errorf("emit: %w", err)
		}

		return text, nil
	}

	return c4d.EmitC4D(c4d.FromModel(m)), nil
}

// convertOutputPath places the twin next to the input with the extension
// swapped (D-30). The filename is derived from the input's basename ONLY —
// no user-controlled filename portion (T-35-07-01). -o overrides the
// directory.
func convertOutputPath(inputPath, targetExt string) string {
	dir := filepath.Dir(inputPath)
	if convertOutDir != "" {
		dir = convertOutDir
	}

	return filepath.Join(dir, deriveBasename(inputPath)+targetExt)
}

// convertGraph converts every file in the entry's include graph into the
// target format (D-25 --follow-includes). Per file, the twin is emitted from
// THAT file's fresh parse (the same preservation rule as single-file mode:
// no include.Resolve/template.Expand/peer.Resolve ever touches an emission
// model); only the include-directive PATH STRINGS are rewritten to the
// target extension. Files already in the target format are skipped —
// conversion is additive (originals are never deleted), so their untouched
// include directives keep resolving against the original files.
func convertGraph(entryPath, targetExt string) error {
	entryDir := filepath.Dir(entryPath)

	absEntry, err := filepath.Abs(entryPath)
	if err != nil {
		return fmt.Errorf("include: canonicalize entry path: %w", err)
	}

	files, err := walkIncludeGraph(absEntry)
	if err != nil {
		return fmt.Errorf("include: %w", err)
	}

	for _, path := range files {
		if filepath.Ext(path) == targetExt {
			continue // already in the target format — no twin needed
		}

		m, err := parseInput(path)
		if err != nil {
			return fmt.Errorf("parse: %w", err)
		}

		for i := range m.Includes {
			m.Includes[i].Path = retargetExt(m.Includes[i].Path, targetExt)
		}

		text, err := emitConverted(m, targetExt)
		if err != nil {
			return err
		}

		dst := graphTwinPath(entryDir, path, targetExt)

		if err := os.MkdirAll(filepath.Dir(dst), 0o750); err != nil {
			return fmt.Errorf("write: create output directory: %w", err)
		}

		//nolint:gosec // G306: 0644 is the documented convert output permission
		if err := os.WriteFile(dst, []byte(text), 0o644); err != nil {
			return fmt.Errorf("write: %w", err)
		}
	}

	return nil
}

// walkIncludeGraph walks the include graph deterministically (depth-first
// in include-directive order, matching internal/include's traversal order)
// and returns every reachable file including the entry. This replicates
// include's canonical-path stack/visited pattern (its internals are
// unexported; cycle behavior is pinned by TestConvertGraphCycle): canonical
// absolute paths key the ancestor stack (cycle detection) and the visited
// set (diamond dedup — each file converts once). The D-24 gate runs BEFORE
// this walk and already rejects cycles, so the guards here are
// defense-in-depth (T-35-07-02/T-35-07-03).
func walkIncludeGraph(entryAbs string) ([]string, error) {
	var files []string

	visited := map[string]bool{}

	var walk func(path string, ancestors []string) error

	walk = func(path string, ancestors []string) error {
		if len(ancestors) > maxConvertDepth {
			return fmt.Errorf("%w %d (cycle or pathological graph)",
				errConvertDepth, maxConvertDepth)
		}

		if visited[path] {
			return nil
		}

		visited[path] = true
		files = append(files, path)

		m, err := parseInput(path)
		if err != nil {
			return fmt.Errorf("parse: %w", err)
		}

		next := append(slices.Clone(ancestors), path)

		targets, err := collectIncludeTargets(m, path, next)
		if err != nil {
			return err
		}

		for _, abs := range targets {
			if err := walk(abs, next); err != nil {
				return err
			}
		}

		return nil
	}

	if err := walk(entryAbs, nil); err != nil {
		return nil, err
	}

	return files, nil
}

// collectIncludeTargets canonicalizes one file's include directives and
// cycle-checks them against the ancestor chain, returning the canonical
// targets in directive order.
func collectIncludeTargets(m *parser.Model, includingPath string, ancestors []string) ([]string, error) {
	targets := make([]string, 0, len(m.Includes))

	for _, inc := range m.Includes {
		abs, err := canonicalizeGraphPath(includingPath, inc.Path)
		if err != nil {
			return nil, err
		}

		if slices.Contains(ancestors, abs) {
			return nil, fmt.Errorf("%w: %s", errConvertCycle, inc.Path)
		}

		targets = append(targets, abs)
	}

	return targets, nil
}

// canonicalizeGraphPath resolves an include directive's path against the
// including file's directory into a canonical absolute path (filepath.Abs
// does not resolve symlinks — the internal/include canonicalize precedent,
// T-35-07-03: a symlink loop still collapses into the stack's cycle error).
func canonicalizeGraphPath(includingPath, incPath string) (string, error) {
	joined := filepath.Join(filepath.Dir(includingPath), incPath)

	abs, err := filepath.Abs(joined)
	if err != nil {
		return "", fmt.Errorf("canonicalize include path %q: %w", incPath, err)
	}

	return filepath.Clean(abs), nil
}

// retargetExt rewrites an include-directive path to the twin extension,
// preserving the authored relative form ("./x.c4d" -> "./x.toml",
// "domains/auth.toml" -> "domains/auth.c4d"). Paths already in the target
// format (mixed graphs, D-26) are returned unchanged — an identity rewrite.
func retargetExt(path, targetExt string) string {
	ext := filepath.Ext(path)
	if ext == "" || ext == targetExt {
		return path
	}

	return strings.TrimSuffix(path, ext) + targetExt
}

// graphTwinPath places one graph twin: next to its source by default, or
// under convertOutDir preserving the source's directory structure relative
// to the entry's directory (domains/auth.toml stays at domains/auth.c4d
// under -o). The filename is the source basename + swapped extension ONLY
// (T-35-07-01).
func graphTwinPath(entryDir, srcPath, targetExt string) string {
	dir := filepath.Dir(srcPath)

	if convertOutDir != "" {
		dir = convertOutDir

		if rel, err := filepath.Rel(entryDir, filepath.Dir(srcPath)); err == nil &&
			rel != "." && !strings.HasPrefix(rel, "..") {
			dir = filepath.Join(convertOutDir, rel)
		}
	}

	return filepath.Join(dir, deriveBasename(srcPath)+targetExt)
}
