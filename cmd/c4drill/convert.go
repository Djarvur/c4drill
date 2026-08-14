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

	"github.com/Djarvur/c4drill/internal/c4d"
	"github.com/Djarvur/c4drill/internal/include"
	"github.com/Djarvur/c4drill/internal/parser"
	"github.com/Djarvur/c4drill/internal/peer"
	"github.com/Djarvur/c4drill/internal/template"
	"github.com/Djarvur/c4drill/internal/validator"
	"github.com/spf13/cobra"
)

// Static errors for better error handling.
var errWrongDirection = errors.New("wrong input extension for direction")

// Conversion directions (D-28): the direction names the TARGET format.
const (
	dirToTOML = "to-toml"
	dirToC4D  = "to-c4d"
)

//nolint:gochecknoglobals // Cobra flags require package-level variables (root.go precedent)
var convertOutDir string

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
	if err := validateSourceForConvert(cmd, inputPath); err != nil {
		return err
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
