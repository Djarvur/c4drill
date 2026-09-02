package main

// serve subcommand (issue #32): runs the c4drill language server — the
// shared foundation the VS Code (#27), JetBrains (#29) and Zed (#30) plugins
// and the GUI app (#31) all talk to. The transport is LSP-over-stdio
// (Content-Length framed JSON-RPC on stdin/stdout); the server core in
// internal/lsp is transport-agnostic, which is what the GUI's in-memory
// transport will drive. Stdout carries ONLY protocol traffic — the CLI is
// silent on success and the server never logs to it.

import (
	"errors"
	"fmt"

	"github.com/Djarvur/c4drill/internal/lsp"
	"github.com/spf13/cobra"
)

// errServeMode is returned when serve is invoked without a transport mode.
var errServeMode = errors.New("serve requires --lsp (the only transport mode currently)")

// newServeCmd builds the serve command.
func newServeCmd() *cobra.Command {
	var lspMode bool

	cmd := &cobra.Command{
		Use:   "serve --lsp",
		Short: "Run the c4drill language server (LSP over stdio)",
		Long: `Run the c4drill language server.

Speaks the Language Server Protocol over stdio (Content-Length framed
JSON-RPC) for both authoring formats: the c4drill TOML dialect and C4D.
Diagnostics run the exact CLI pipeline, so message text and line numbers
match c4drill <file> — including cross-file [[include]] graphs, where
editing an included file republishes diagnostics for every including
document.

Editors launch this command as: c4drill serve --lsp`,
		Args:         cobra.NoArgs,
		RunE:         runServe(&lspMode),
		SilenceUsage: true,
	}

	cmd.Flags().BoolVar(&lspMode, "lsp", false,
		"Serve LSP over stdio (stdin/stdout carry the protocol)")

	return cmd
}

// runServe wires the --lsp flag into the server loop (indirection keeps the
// flag captured at RunE time and the command testable).
func runServe(lspMode *bool) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		if !*lspMode {
			return errServeMode
		}

		lsp.SetServerVersion(version)

		if err := lsp.Serve(cmd.Context(), cmd.InOrStdin(), cmd.OutOrStdout()); err != nil {
			return fmt.Errorf("lsp: %w", err)
		}

		return nil
	}
}
