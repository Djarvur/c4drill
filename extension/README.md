# c4drill — VS Code extension

A [Visual Studio Code](https://code.visualstudio.com) extension for
[c4drill](../README.adoc) models. It is a pure **client**: every language
behavior comes from the shared Go LSP server (`c4drill serve --lsp`,
issue #32) — nothing is duplicated here. Issue #27 is the tracking issue.

## Features

| Feature | How it reaches VS Code |
| --- | --- |
| Syntax highlighting (C4D) | `c4drill-c4d` language + TextMate grammar (`syntaxes/c4d.tmLanguage.json`, authored against `internal/c4d/grammar/c4d.peg`); the same artifact is reused by the JetBrains plugin |
| TOML highlighting (scoped) | Plain `.toml` files are never touched unless they match `c4drill.toml.patterns` globs (default `**/*.architecture.toml`) or are activated per file ("C4Drill: Activate for This File") |
| Diagnostics (CLI-parity, incl. cross-file include graph) | LSP `textDocument/publishDiagnostics` on open/change/save |
| Autocomplete (17 unit types with promotion, fields, enums, peers bare+dotted, templates, include paths — TOML dialect) | LSP `textDocument/completion` |
| Hover, go-to-definition | LSP `textDocument/hover`, `textDocument/definition` |
| Format Document / format-on-save | LSP `textDocument/formatting` (both formats) |
| Diagram preview | Webview panel with debounced `c4drill/renderDiagram`, drill-through, breadcrumbs, view controls, five-format export |

## Settings

| Setting | Default | Meaning |
| --- | --- | --- |
| `c4drill.server.path` | — | Path to the `c4drill` binary; when unset it is discovered on `PATH` |
| `c4drill.toml.patterns` | `["**/*.architecture.toml"]` | Glob patterns for which `.toml` files are treated as c4drill models |
| `c4drill.preview.debounce` | `200` | Preview re-render debounce in milliseconds |

## Commands

- **C4Drill: Show Preview** (also an editor-title button) — open the live diagram preview
- **C4Drill: Validate File** — re-send the buffer and refresh diagnostics
- **C4Drill: Format Document** — LSP formatting (TOML and C4D)
- **C4Drill: Activate for This File** / **Deactivate for This File** — per-file TOML opt-in/out

## Preview

The preview calls `c4drill/renderDiagram` for the active document and
re-renders, debounced, as you type. Clicking a drill-capable node follows
the diagram's own navigation links into the child view (C1 → C2 → C3) with
breadcrumbs back; external `reference` links open in the system browser.
When the model does not parse or validate, the panel shows the same
messages the CLI prints (with line numbers) instead of a stale diagram.

Exports use the CLI conventions (`-f svg|html|dot|png|plantuml`).

## Requirements

- The `c4drill` binary on `PATH`, or a full path in `c4drill.server.path`
  (build locally with `go build ./cmd/c4drill`).
- Completion/hover/definition are currently served for the **TOML dialect
  only**; `.c4d` LSP support is tracked server-side (issue #33).

## Known limits

- Preview breadcrumbs show unit **ids** (e.g. `amazon / rds`), not display
  names — display names would need a server-side lookup beyond the current
  wire contract.
- Diagnostics for an explicitly activated *untitled* `.toml` may not render
  in the Problems view (no static selector can match untitled TOMLs);
  saved files are unaffected.

## Development

```bash
npm install
npm run compile        # tsc
npm test               # unit tests (grammar tokenizer, scoping, render targets)
npm run test:integration  # @vscode/test-electron against a real `c4drill serve --lsp`
```

The integration suite builds/uses the server binary; see `package.json`
scripts for the expected location.
