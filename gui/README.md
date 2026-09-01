# c4drill GUI (issue #31)

The c4drill desktop app: the full authoring loop in one window — edit the
architecture model, watch the diagram re-render live, ask the AI assistant to
change things, export.

```
┌──────────┬───────────────────┬──────────────────────────┬──────────────┐
│ files    │ editor (CodeMirror│ preview (live SVG)       │ AI chat (P1) │
│ (.toml/  │ 6, TOML + C4D)    │  toolbar: level/expanded/│  settings,   │
│  .c4d)   │  completion,      │  legend/pause/zoom/export│  streaming,  │
│ tabs     │  diagnostics,     │  breadcrumbs, drill-down │  diff-confirmed│
│          │  hover, format    │  zoom/pan, error panel   │  edits       │
└──────────┴───────────────────┴──────────────────────────┴──────────────┘
```

## Architecture

- **One backend, two transports.** `gui/internal/app` is the orchestration
  layer over the existing in-process Go packages (parser, validator, LSP
  core, render, output). Both the Wails desktop window (`gui/main.go`) and
  the plain-HTTP fallback speak the same `Dispatch(method, params)` JSON
  protocol, so behavior cannot drift between them.
- **One LSP server, four clients.** The editor area drives the shared
  `internal/lsp` server core through its in-memory `Handle` transport — the
  same entry point the VS Code / JetBrains / Zed / stdio clients wrap — so
  diagnostics wording, completion, formatting and the `c4drill/renderDiagram`
  live-preview method are identical to the CLI and the other editors by
  construction.
- **C4D highlighting via a real grammar.** The editor parses `.c4d` buffers
  with a dedicated CodeMirror Lezer grammar (`frontend/src/language/`),
  authored against `internal/c4d/grammar/c4d.peg` with the #27 TextMate and
  #30 tree-sitter grammars as references for the disambiguation cases. It
  provides syntax highlighting, code folding and brace indentation; the P0
  StreamLanguage fallback is gone (TOML highlighting still uses the legacy
  mode).
- **Preview fidelity.** The preview renders through `c4drill/renderDiagram`,
  i.e. the exact CLI pipeline. Export writes through `internal/output`'s
  Writer with the CLI's `-f svg|html|dot|png|plantuml` conventions; tests
  assert the exported SVG is byte-identical to the preview SVG.
- **Drill-down.** SVG internal links encode the CLI's output file layout
  (`{basename}.svg`, `{basename}/{unit}.svg`, dotted paths as directories);
  the backend inverts that layout back to render targets (click → drill,
  breadcrumbs → back up, alt-click → in-place C1 expand override).

## Build & run

Prerequisites: Go ≥ 1.26, Node/npm (frontend build), and for the desktop
window a platform webview (macOS: Xcode CLT; Linux: webkit2gtk; Windows:
WebView2). CGO is required for the Wails window.

```sh
# frontend (once, before the first desktop build)
cd gui/frontend && npm install && npm run build

# desktop app (system webview window)
go build ./gui && ./gui --dir /path/to/project

# webview-less mode: same UI in a regular browser
go run ./gui --serve --addr 127.0.0.1:5278 --dir /path/to/project

# frontend dev mode (vite HMR against the HTTP backend)
go run ./gui --serve &          # backend on :5278
cd gui/frontend && npm run dev  # vite on :5279, /api proxied
```

Without a prior `npm run build`, `go build` still succeeds: the committed
`gui/frontend/dist/.gitkeep` keeps `go:embed` valid and the app serves a
placeholder page (tests and the backend API work; the UI needs the vite
build).

The `wails` CLI (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)
is optional — only needed for `wails dev`/packaging; plain `go build` works.

## Tests

```sh
go test ./gui/...            # backend binding logic + HTTP smoke e2e
go test -race ./gui/...      # chat streams across goroutines
cd gui/frontend && npm test           # vitest: Lezer grammar token tests
cd gui/frontend && npx tsc --noEmit   # frontend typecheck (also in npm run build)
```

Coverage highlights: in-memory LSP round trip (didOpen → clean diagnostics,
broken buffer → CLI-parity messages, cross-file include republish),
completion/format/hover, live render contract (no stale diagram on invalid
models), drill-target algebra (the real href set from C1/C2/C3 outputs),
export layout + byte parity, chat streaming/proposals/confirmation gating
for both providers (OpenAI-compatible and Anthropic Messages-API mocked via
httptest, including the abort/cancellation path), and a full HTTP-transport
smoke e2e. The frontend's vitest suite parses every `.c4d` example fixture
with the Lezer grammar (no error nodes) and asserts the highlight spans and
folding/indentation props.

## AI chat (P1)

Settings (⚙ in the chat header): provider, base URL, model, API key, optional
extra system prompt. Two providers ship (issue #36): **openai-compatible**
(OpenAI, Ollama `http://localhost:11434/v1`, LM Studio, most proxies) and a
native **Anthropic** Messages-API adapter (`https://api.anthropic.com` —
`x-api-key`/`anthropic-version` headers, top-level `system` field, typed SSE
events). Settings persist per provider, so switching never clobbers the other
slot. The key is stored in the local user-config file only
(`<UserConfigDir>/c4drill/gui.json`, 0600) and is never echoed back to the UI.

While an answer streams, the composer's **Stop** button aborts the request:
the backend cancels the provider stream end-to-end (request context), the
partial answer stays in the transcript visibly marked "stopped — answer is
partial", and the composer resets. Aborted answers never carry edit proposals.

The assistant's system prompt is seeded from a build-time snapshot of
`plugins/c4drill/skills/c4drill-toml/SKILL.md` (`gui/internal/ai/skill_seed.md`
— `go:embed` cannot reach outside `gui/`, so regenerate the snapshot when the
skill changes). Edit proposals arrive as fenced `c4drill-edit path=…` blocks,
render as add/remove diffs, and are written **only** on explicit Apply, with
the scope (opened project, model files only) re-checked at apply time.

Not yet (honest list): streaming tool-calls. Known grammar trade-offs (all
documented in `frontend/src/language/c4d.grammar`): inside an id-led header
the id and type words share the plain-text style, and unquoted value words
that start with identifier characters lex as identifiers — so a value like
`https://x` splits at the colon (quote URLs). Nothing in the example
corpus trips either case.

## Layout map

```
gui/
  main.go              Wails shell + HTTP fallback (one Dispatch protocol)
  internal/app/        backend: workspace, LSP bridge, render/drill, export,
                       chat orchestration, dispatch table (+ tests)
  internal/ai/         provider clients (OpenAI-compatible + Anthropic
                       Messages-API), prompt assembly, edit proposals/diff
                       (+ tests, skill_seed.md snapshot)
  frontend/            vite + TypeScript: CodeMirror 6 editor (TOML legacy
                       mode + the C4D Lezer grammar), preview, toolbar, chat
                       panel; dist/ is embedded via go:embed
  build/appicon.png    app icon placeholder (Wails packaging convention)
```
