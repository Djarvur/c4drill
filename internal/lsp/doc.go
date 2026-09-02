// Package lsp implements the c4drill language server (issue #32): a
// transport-agnostic LSP core for the c4drill TOML dialect and the C4D DSL,
// with a stdio JSON-RPC transport for the editor clients (#27/#29/#30) and an
// in-proc Handle entry the GUI app (#31) can drive over an in-memory
// transport without touching capability logic.
//
// The server reuses the CLI's composition pipeline (include.Resolve →
// template.Expand → peer.Resolve → validate) so diagnostics match
// `c4drill <file>` message-for-message and line-for-line.
package lsp
