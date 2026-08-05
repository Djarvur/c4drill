# External Integrations

**Analysis Date:** 2026-08-05

## APIs & External Services

**Network APIs:**
- None. The application makes zero outbound network calls. It is a fully self-contained offline CLI tool. Verified by grep: no `net/http`, no third-party SDK imports beyond rendering/parsing libraries.

**Embedded engine (the only "service"):**
- Graphviz - Diagram layout and rendering engine, compiled to WASM and executed via wazero. This is the core rendering capability.
  - SDK/Client: `github.com/goccy/go-graphviz`, replaced in `go.mod` (line 28) by the `github.com/onokonem/go-graphviz` fork v0.0.0-20260321130544-f364b5235161
  - Engine: WASM Graphviz binary run by `github.com/tetratelabs/wazero` v1.10.1
  - Concurrency: WASM engine is NOT thread-safe; all render calls are serialized via a package-global `wasmMutex` in `internal/render/render.go` (lines 15-20)
  - Usage: `internal/render/render.go` (`graphviz.New`, `buildCgraph`, `gv.Render`) and `internal/render/converter.go` (`buildCgraph`, node/cluster/edge construction against the `cgraph` API)
  - Output formats: XDOT (`graphviz.XDOT`) and SVG (`graphviz.SVG`)

## Data Storage

**Databases:**
- None. No SQL, no ORM, no persistent storage.

**File Storage:**
- Local filesystem only.
  - Input: one TOML file (`parser.ParseFile` in `internal/parser/parser.go:322`)
  - Output: rendered `.svg`/`.dot` files written to a directory hierarchy by `internal/output/writer.go` (C1: `{basename}.{format}`; C2/C3: `{basename}/{dotted-path}.{format}`; expanded: `{basename}.expanded.{format}`). Directories created `0o750`, files written `0o600`

**Caching:**
- None. Each run re-parses and re-renders from scratch.

## Authentication & Identity

**Auth Provider:**
- Not applicable. Offline CLI with no authentication, no user accounts, no sessions.

## Monitoring & Observability

**Error Tracking:**
- None. Errors propagate as Go errors; CLI prints them to stderr and exits with code 1 (`cmd/c4drill/main.go`).

**Logs:**
- None beyond error output. Success is silent by design (`cmd/c4drill/root.go:148` comment: "Success - silent per spec"). Exit codes: 0 success, 1 error (documented in `README.md`).

## CI/CD & Deployment

**Hosting:**
- GitHub (repo `Djarvur/c4drill`). Distributed as Go source (`go install`) or built binary; no hosted service.

**CI Pipeline:**
- GitHub Actions: `.github/workflows/validate-skill-examples.yml` - triggered on push/PR touching `skill/**`. Steps: `actions/checkout@v4`, `actions/setup-go@v5` (version from `go.mod`), `go build`, then validates every `skill/examples/*.toml` by running `./c4drill <file> -f dot`.

## Environment Configuration

**Required env vars:**
- None required for operation.

**Optional env vars:**
- `C4DRILL_LABEL_RATIO` - overrides default label width:height ratio (1.6). Read at `cmd/c4drill/root.go:301`. Precedence: `--label-ratio` CLI flag > env var > default.

**Secrets location:**
- Not applicable. No secrets, API keys, or credentials are used anywhere in the codebase.

## Webhooks & Callbacks

**Incoming:**
- None. No HTTP endpoints.

**Outgoing:**
- None. No webhook calls. (The only "callbacks" are Graphviz node `URL` attributes embedded in SVG output for drill-down navigation between C1/C2/C3 diagrams — see `cn.SetURL(node.ExploreURL)` in `internal/render/converter.go:245` and `internal/graph/navigation.go`.)

## Local Data Assets (non-service)

**Legacy icon assets:**
- `data/*.svg` (person, system, container, component, db, pipe) - historical per-type icon files from an earlier phase. NOT referenced by current Go source (verified: no `data/` imports, no `go:embed` anywhere). Icons are now rendered as Graphviz native shapes plus emoji via `IconForType` in `internal/graph/shapes.go` (person U+1F464, db U+26C1, queue U+255F/U+2562). Safe to ignore for integration purposes.

**Private test data:**
- `cyp-auth-infra/` - large TOML/SVG test fixtures, gitignored (`.gitignore` line 72: "keep local, never commit").

**Sample/test fixtures:**
- `testdata/` (valid.toml, nested.toml, links.toml, links.dot, invalid_*.toml) - committed fixtures used by tests across `internal/parser`, `internal/validator`, `internal/render`, `internal/graph`, `internal/view`, `internal/output`, and `cmd/c4drill`.
- `skill/examples/*.toml` - TOML examples consumed by the CI validation workflow.

---

*Integration audit: 2026-08-05*
