# c4drill — Zed extension

A [Zed](https://zed.dev) editor extension for [c4drill](../README.adoc)
models. It is a pure **client**: every language behavior comes from the
shared Go LSP server (`c4drill serve --lsp`, issue #32) — nothing is
duplicated here. Issue #30 is the tracking issue.

## Features

| Feature | How it reaches Zed |
| --- | --- |
| Syntax highlighting (C4D) | `tree-sitter-c4d` grammar + `highlights.scm` (Zed has no TextMate support) |
| TOML highlighting (scoped) | "C4Drill TOML" language reusing Zed's built-in `toml` grammar |
| Diagnostics (CLI-parity, incl. cross-file include graph) | LSP `textDocument/publishDiagnostics` on open/change/close |
| Autocomplete (17 unit types with promotion, fields, enums, peers bare+dotted, templates, include paths — TOML dialect) | LSP `textDocument/completion` |
| Hover, go-to-definition | LSP `textDocument/hover`, `textDocument/definition` |
| Outline (TOML models) | LSP `textDocument/documentSymbol`; C4D uses `outline.scm` |
| Format Document / format-on-save | LSP `textDocument/formatting` (both formats) |
| Diagram preview | **Fallback surface** — extension-provided tasks that run the `c4drill` CLI (see below) |

## Layout

```
zed-extension/
├── extension.toml            # manifest: languages, grammar, language server
├── Cargo.toml + src/         # WASM shim (zed_extension_api), compiled to wasm32-wasip2
├── grammars/c4d/             # the standalone tree-sitter-c4d grammar repo
│   ├── grammar.js            #   mirrors internal/c4d/grammar/c4d.peg
│   ├── src/                  #   generated parser (committed — Zed compiles it directly)
│   └── test/corpus/          #   corpus tests (tree-sitter test)
├── languages/c4d/            # C4D DSL: config, queries, preview tasks
├── languages/c4drill-toml/   # scoped TOML dialect: config, queries, preview tasks
└── scripts/bootstrap-grammar.sh
```

## Languages and TOML scoping (the joint decision)

Two languages ship in one extension:

1. **C4D** — the DSL, claimed unconditionally via `path_suffixes = ["c4d"]`.
2. **C4Drill TOML** — the TOML dialect, which deliberately claims **no
   suffixes** (`path_suffixes = []`). Zed has no per-extension glob
   claiming, and plain `.toml` files must stay untouched (#30 acceptance
   criterion 2), so activation is pattern-based and user-controlled: the
   standard `file_types` settings glob maps chosen paths to the language:

   ```json
   {
     "file_types": {
       "C4Drill TOML": ["models/**/*.toml", "architecture/**/*.toml"]
     }
   }
   ```

   or per buffer via `editor: change language`. Plain TOML files never see
   c4drill tooling unless matched by such a glob. This is the same joint
   decision as the VS Code client (#27, whose `c4drill.toml.patterns`
   defaults to `**/*.architecture.toml`) — only the mechanism differs,
   because Zed's extension system has no document selector; its analogue
   is exactly the `file_types` settings glob. A matching default:
   `"file_types": { "C4Drill TOML": ["**/*.architecture.toml"] }`.
   The language reuses Zed's built-in `toml` grammar (no vendored copy; if
   the bundled TOML extension is disabled the language still works — just
   without syntax coloring).

The same two-surface split is what the VS Code (#27) and JetBrains (#29)
clients implement, so the scoping decision stays joint.

## Language server wiring

The WASM shim (`src/lib.rs`) implements the standard Zed pattern:

1. **Settings override** — `"lsp": { "c4drill": { "binary": { "path":
   "/opt/tools/c4drill", "arguments": ["serve", "--lsp"],
   "env": {"C4DRILL_LOG": "debug"} }, "settings": {...} } }`. A custom
   `binary.arguments` replaces the default `serve --lsp` entirely.
2. **Host probe** — `sh -c 'command -v -- c4drill'` (unix) or
   `cmd /C where c4drill` (windows) executed through Zed's process host.
3. **Bare name** — `c4drill`, resolved against `PATH` by the Zed host when
   spawning (covers systems without `which`/`where`).

The user's `lsp.c4drill.settings` block is forwarded to the server as
workspace configuration. The current #32 server accepts and ignores it;
the plumbing is future-proof.

## Diagram preview — the fallback path (read this)

**The WebView path did not land.** As of `zed_extension_api` 0.7.0 there is
**no WebView/panel API** — the `Extension` trait surface is language-server
command/configuration, labels, slash commands, context servers and DAP
only. A visual-extension API (webviews, sidebar panels) is still an open
proposal: zed-industries/zed discussion #53403. The issue anticipated this:
*"if the WebView API is unavailable, the fallback is regenerate-on-save …
the auto-refresh promise must hold in whichever path lands."*

What ships instead — and what holds the promise:

- **Live error state**: the LSP republishes CLI-identical diagnostics on
  every keystroke, so the editor always reflects the current model.
- **One-keystroke regeneration**: extension-provided task templates
  (visible in `task: spawn` when a `.c4d` / C4Drill TOML buffer is open):
  - `c4drill: render diagram` — writes the full SVG site next to the model
    (`{name}.svg`, `{name}/{system}.svg`, …). Because it *is* the CLI, the
    output is byte-for-byte identical to `c4drill <file>` (acceptance
    criterion), and it carries **drill-down links and breadcrumbs inside
    the SVG** — the renderer's own `xlink:href` navigation plus the graph
    breadcrumb bar — so drill-click navigation and breadcrumbs work in the
    browser exactly as the issue describes, with external `reference`
    links opening in the browser too.
  - `c4drill: render diagram (all expanded)` — the `--expanded` single diagram.
  - `c4drill: render diagram and open in browser` — renders, then opens the
    C1 SVG (`open`/`xdg-open`; on Windows open `{name}.svg` manually).
  - `c4drill: export dot` / `c4drill: export PlantUML` — the export actions.
  All tasks set `save: current`, so the render always reflects the buffer
  content ("regenerate-on-save" semantics).

Bind the render task to a key for a live-preview loop:

```json
{
  "context": "Editor && language == C4D",
  "bindings": {
    "ctrl-alt-d": ["task::Spawn", { "task_name": "c4drill: render diagram" }]
  }
}
```

**When Zed ships a WebView surface**, the client-side wiring is already
specified by the server: `c4drill/renderDiagram` (params on
`RenderDiagramParams` in `internal/lsp/protocol.go`) takes the document URI
plus `target`/`allExpanded`/`expanded`/`legend` and returns SVG text plus
render diagnostics. The SVG's relative drill links map back to unit paths
via the documented file layout (`{name}.svg` = C1, `{name}/{unit}.svg` =
C2, `{name}/{unit}/{sub}.svg` = C3 — `internal/graph/path.go`
`ComputeExploreURL`). A future shim version would issue that request
debounced on change, intercept link clicks, and host the result in a panel
— no server changes needed.

## Grammar: tree-sitter-c4d

`grammars/c4d/grammar.js` mirrors the authoritative PEG
(`../internal/c4d/grammar/c4d.peg`): brace-block units in all three header
forms, the four ASCII arrows with `peer: label { options }`, field keywords,
`properties`/`template`/`use`/`include`, `${param}` tokens, strings incl.
`"""triple"""`, lists, barewords, `#` comments, `;` one-liners.

Design notes (details inline in `grammar.js`):

- **Type words are not keyword tokens.** Models legally use them as unit ids
  (`queue: containerQueue "Platform Queue" { … }`), and a keyword token
  would win the lexer race and break the id-led header. Type slots are
  aliased identifiers/barewords; the id-led vs type-led choice is resolved
  structurally by the GLR on the `:` — the PEG's ordered choice. Highlighting
  captures the slot positionally.
- **Reserved words** (`properties`/`template`/`use`/`include`/`once`/
  `external`) stay literals; string literals outrank identifiers, so
  reserved-id units fail to parse — the same D-19 rejection the PEG
  produces (the LSP still emits the friendly diagnostic).
- **Tolerant by design**: unknown field/option/property keys parse as
  ordinary fields, so highlighting survives partially-written code (the
  server's diagnostics report them).

The grammar is self-contained and extraction-ready: push `grammars/c4d/`
to a standalone `tree-sitter-c4d` repository and update the manifest URL —
no grammar changes required.

### Grammar tests

```sh
cd grammars/c4d
npm install
npx tree-sitter test        # corpus tests (test/corpus, generated from test/inputs)
```

Additionally, every checked-in example parses cleanly:
`skill/examples/**/*.c4d` and `internal/include/testdata/*.c4d` (verified
with `tree-sitter parse`).

## Building and installing the extension

Prerequisites: Rust via rustup (the `wasm32-wasip2` target is installed
automatically by Zed), node (only for grammar regeneration), and a
`c4drill` binary on `PATH` (or a `binary.path` settings override).

```sh
# 1. make the in-repo grammar cloneable by Zed (offline, local tweak)
zed-extension/scripts/bootstrap-grammar.sh --local

# 2. quality gates
cd zed-extension
cargo check --target wasm32-wasip2     # the WASM shim compiles
cargo test                             # host-side discovery/config tests

# 3. Zed: Extensions > Install as Dev Extension > select zed-extension/
```

`--local` rewrites the manifest's grammar `repository` to a `file://` URL —
a **local, uncommitted** tweak, because Zed clones the grammar repository
by URL and cannot yet resolve an in-repo directory otherwise. Once the
grammar is extracted (pushed to `https://github.com/Djarvur/tree-sitter-c4d`),
run `scripts/bootstrap-grammar.sh --canonical`, keep the canonical URL in
`extension.toml`, and pin a SHA as `rev` for registry publishing.

### Packaging for the Zed extension registry

1. Push `grammars/c4d/` to the dedicated grammar repository; set
   `[grammars.c4d] rev` to the pushed commit SHA.
2. From this directory, package with the registry's standard script:

   ```sh
   git clone https://github.com/zed-industries/extensions.git /tmp/zed-extensions
   /tmp/zed-extensions/package_extensions.sh \
       --extension-path "$PWD" --output /tmp/zed-dist
   ```

   (This compiles the Rust shim to `wasm32-wasip2` and each grammar to a
   Tree-sitter WASM artifact, producing the uploadable archive.)
3. Submit a PR to `zed-industries/extensions` adding the archive +
   `extensions.json` entry (community-extension flow). Bump `version` in
   `extension.toml` for each release; re-verify against each Zed release —
   `zed_extension_api` churn is a known risk (pinned at 0.7.0 here).

## Server gaps observed (reported, not patched here)

- `textDocument/completion` returns empty for `.c4d` documents (the TOML
  dialect only). The client wires it regardless; C4D completions arrive
  server-side later.
- The server ignores workspace configuration; forwarded settings are inert
  today.
- The server does not advertise `workspace/didChangeWatchedFiles`
  registration, so include-file edits made outside Zed may not revalidate
  dependents until the includer is touched.
- No server-side semantic tokens; the optional TOML dialect semantics
  layer would ride `semantic_tokens` (off by default in Zed) plus
  `semantic_token_rules.json`, not tree-sitter.
