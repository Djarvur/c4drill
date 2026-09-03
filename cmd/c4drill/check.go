package main

// The check subcommand (issue #41): render-free model validation. check runs
// the SAME pipeline front-half as render — extension dispatch, [[include]]
// resolution, template expansion, relative peer resolution, then full
// semantic validation — through the shared parseValidatedModel helper (D-02),
// reports exactly the validation errors render would report (D-03), and
// stops: no views, no graphviz, no files written, no output directory
// required (CHECK-01). Silent exit 0 on a valid model (D-04); exit 1 with
// the render-identical errors on an invalid one (CHECK-02). It defines no
// render flags and hides the inherited ones from its help (D-05).

import (
	"github.com/spf13/cobra"
)

// checkLong is check's long help: the same-front-half-as-render contract, the
// exit codes, and the no-output guarantee.
const checkLong = `Validate a C4 model without rendering anything.

check runs the same pipeline front-half as render — input dispatch by
extension, [[include]] resolution, template expansion, relative peer
resolution, then full semantic validation — and reports exactly the errors
render would report, then stops. Nothing is written: no output directory
is needed and no diagram files are produced, so it fits CI gates and
edit-loop validation.

Exit codes:
  0  the model is valid (silent)
  1  the model is invalid — validation errors on stderr, identical to
     what render prints for the same file

Examples:
  c4drill check architecture.toml
  c4drill check architecture.c4d`

// renderOnlyFlags are the root command's render-only persistent flags
// (D-05): meaningless without rendering, so check must not advertise them.
// Cobra inherits them parseable onto every subcommand — the same is already
// true of fmt/convert/serve — so they stay parseable-but-ignored here; only
// check's help hides them.
//nolint:gochecknoglobals // flag-name list for the cobra inherited-flag surface (root.go:56 precedent)
var renderOnlyFlags = []string{
	"format", "output", "expanded", "plain", "edges", "no-colors",
	"no-styles", "no-length", "no-rank", "no-labels", "label-ratio",
}

// newCheckCmd creates the check subcommand: render-free model validation
// (issue #41).
func newCheckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check <file.toml|file.c4d>",
		Short: "Validate a model without rendering",
		Long:  checkLong,
		Args:  cobra.ExactArgs(1),
		RunE:  runCheck,
		// SilenceUsage: a validation failure is the answer, not a usage
		// mistake (same choice as the root and fmt commands).
		SilenceUsage: true,
	}

	// D-05: render check's help without the inherited render flags. A help
	// func scoped to check is required because cobra shares one flag object
	// between the root and every subcommand — permanently marking the
	// inherited flags Hidden would also strip them from root --help.
	cmd.SetHelpFunc(helpCheck)

	return cmd
}

// helpCheck renders check's help with the inherited render flags hidden
// (D-05). The hidden state is restored as soon as the template is rendered,
// so root and sibling commands keep advertising their flags.
func helpCheck(c *cobra.Command, args []string) {
	setRenderFlagsHidden(c, true)
	defer setRenderFlagsHidden(c, false)

	c.Parent().HelpFunc()(c, args)
}

// setRenderFlagsHidden shows or hides the render-only flags in cmd's
// inherited flag set.
func setRenderFlagsHidden(cmd *cobra.Command, hidden bool) {
	for _, name := range renderOnlyFlags {
		if f := cmd.InheritedFlags().Lookup(name); f != nil {
			f.Hidden = hidden
		}
	}
}

// runCheck validates the input model and returns. Success is silent (D-04):
// the validated model is discarded — check never renders and never writes
// (CHECK-01). Failures return the shared front-half's error verbatim —
// stage-prefixed parse/include/expand/peer errors or the
// errValidationFailed sentinel after the render-identical report — which
// main() turns into exit 1 (CHECK-02).
func runCheck(cmd *cobra.Command, args []string) error {
	_, err := parseValidatedModel(cmd, args[0])

	return err //nolint:wrapcheck // check surfaces the shared front-half error verbatim
}
