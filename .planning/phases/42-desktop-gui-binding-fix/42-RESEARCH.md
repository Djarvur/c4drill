# Phase 42: Desktop GUI Binding Fix - Research

**Researched:** 2026-09-03
**Domain:** Wails v2 desktop bindings (frontend RPC namespace alignment), TypeScript transport resolver, vitest unit testing
**Confidence:** HIGH

## Summary

The desktop-window RPC break is fully diagnosed and narrowly scoped. Wails v2.15.0 (pinned in `go.mod`) generates runtime bindings at `window.go.<package>.<struct>.<Method>` — verified against official v2 docs, whose generated wrapper reads `window["go"]["main"]["App"]["Greet"](arg1)`. The Go side binds `*desktop` from package `main` (`cmd/c4drill-gui/main.go:74`, `Bind: []interface{}{g}` at line 124), so the only real namespace is `window.go.main.desktop.Dispatch`. The frontend resolver (`internal/gui/frontend/src/rpc.ts:23`) probes `main?.App` and `backend?.App` — both phantom names. `backend.App` is a v1 artifact: Wails v1 exposed bindings at `window.backend`, and v2's migration guide explicitly notes v2 moved to `window.go`; no v2 app ever had `window.go.backend.App`. The break dates to the #31 GUI introduction and was confirmed pre-existing during #37 (commit dc1df8b moved `main.gui` → `main.desktop`).

The fix is a two-line surface change in one file plus tests: rewrite the `WailsWindow` interface (rpc.ts:9-19) and the resolver chain (rpc.ts:23) to reference exactly `go.main.desktop.Dispatch`, per locked decisions D-01/D-02 (no dead fallbacks — neither alternative ever existed post-#31, and keeping them masks regressions). D-03 requires a vitest binding-shape unit test: mock `window.go.main.desktop.Dispatch`, assert the transport resolves and dispatches; assert fail-closed (null transport) when the namespace is absent — this is the TDD RED test since current code never finds `main.desktop`. D-04 defines the structural gate: zero phantom-namespace references in frontend source, `tsc --noEmit` green, `go build ./...` green, and the existing `--serve` e2e (`TestHTTPSmokeE2E`, `cmd/c4drill-gui/main_test.go`) green. Repo-wide grep confirms rpc.ts is the only file referencing phantom namespaces (excluding `node_modules`/`dist` build output).

Real desktop-window verification (`wails build` + run) is out of reach in CI — the wails CLI is not installed and needs CGO/webview — so D-03/D-04 deliberately define the CI-verifiable contract. Issue #38's "loud runtime failure" acceptance note is already distilled into D-03's fail-closed case (resolver returns null → `isDesktop()` false → transport miss surfaces instead of silently dead UI).

**Primary recommendation:** Update rpc.ts's `WailsWindow.go` declaration and `wailsBinding()` resolver to the single canonical path `w.go?.main?.desktop?.Dispatch ?? null`; add `src/rpc.test.ts` with the RED binding-shape tests before the fix; verify with `npm run typecheck`, `npx vitest run`, `go build ./...`, `go test ./cmd/c4drill-gui/...`.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- **D-01:** Update the FRONTEND to the real namespace (issue's "cleaner" direction): `internal/gui/frontend/src/rpc.ts` resolves `window.go.main.desktop.Dispatch`. Renaming the Go side so the generated namespace matches rpc.ts's current expectation would force the bound struct to be named `App` in package `main`/`backend` — making the Go code lie about what it is. Go binding stays as-is.
- **D-02:** Single canonical namespace, no dead fallbacks: the `Window` interface declaration (rpc.ts:11-12) and the resolver chain (rpc.ts:23) reference exactly `go.main.desktop.Dispatch`. The current `main.App` / `backend.App` alternatives never existed post-#31 (`main.gui` was the only pre-#31 name) — keeping them as fallbacks would preserve dead code that masks regressions.
- **D-03:** Binding-shape unit test: mock `window.go.main.desktop.Dispatch` and assert the transport resolves and dispatches through it; a second case asserts dispatch fails closed (returns null transport / surfaces the miss) when the namespace is absent — this is the TDD RED test (current code fails it: resolver never finds `main.desktop`).
- **D-04:** Structural gate: zero references to the phantom namespaces remain (`main?.App`, `backend?.App`) in frontend source; frontend type-check/build green; `go build ./...` green; the existing `--serve` e2e suite stays green (GUI-02 — the HTTP path must be untouched).

### Claude's Discretion
- Test runner/file placement inside the frontend per its existing test setup (planner inspects frontend tooling).
- Whether the resolver refactor warrants extracting the namespace path into a typed constant — minor, planner's call.

### Deferred Ideas (OUT OF SCOPE)
None — discussion stayed within phase scope.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| GUI-01 | In desktop-window mode the frontend RPC layer calls the Wails-generated binding namespace that matches the actually bound struct (`main.desktop`), so desktop-window RPC works again | Wails v2 namespace math verified (`window.go.main.desktop.Dispatch`); fix site isolated to rpc.ts:9-24; binding-shape vitest test design per D-03 |
| GUI-02 | The `--serve` HTTP fallback path is unchanged and the existing e2e suite stays green | HTTP path (`newHandler`, `/api/dispatch`, `/api/events` in main.go:155-175) untouched by a frontend-only fix; e2e identified: `TestHTTPSmokeE2E` via `go test ./cmd/c4drill-gui/...` |
</phase_requirements>

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Desktop RPC transport resolution | Browser / Client (webview frontend, rpc.ts) | — | Wails injects `window.go.*` bindings into the webview; only client JS can see them |
| RPC dispatch protocol (method+JSON params → JSON result) | API / Backend (Go `desktop.Dispatch` → `app.App.Dispatch`) | Browser / Client | Go owns execution; frontend only resolves the binding handle and stringifies params |
| Serve-mode HTTP transport (`POST /api/dispatch`, SSE `/api/events`) | API / Backend (main.go `newHandler`) | Browser / Client (fetch + EventSource in rpc.ts) | Untouched by this phase (GUI-02) — resolver change must not alter the fetch fallback |
| Backend→frontend events (`backend` channel / SSE) | API / Backend (SetEventSink / eventHub) | Browser / Client (`startEvents`) | Already correct in rpc.ts (EventsOn "backend"); explicitly out of scope |

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Wails v2 | v2.15.0 (go.mod, `github.com/wailsapp/wails/v2`) | Desktop shell; generates `window.go.<pkg>.<struct>` bindings at build time | Project-pinned; official framework |
| TypeScript | ^5.9.2 (frontend devDependency) | Strict typing of the `WailsWindow` declaration | Project-pinned; `tsc --noEmit` is the type gate |
| vitest | ^4.1.11 (frontend devDependency) | Unit tests for the binding-shape contract | Already the frontend test runner (`npm test` = `vitest run`; existing `src/language/c4d.test.ts`) |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|--------------|
| vite | ^7.1.5 | Frontend build (assets embedded via `go:embed`) | `npm run build` after the rpc.ts change regenerates `dist/` |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Update rpc.ts to `main.desktop` (D-01, chosen) | Rename Go struct to `App` so bindings match old rpc.ts | Rejected by D-01: makes Go code lie (`desktop` describes the bound app); larger blast radius |
| Direct property access `w.go.main.desktop.Dispatch` | Optional chaining `w.go?.main?.desktop?.Dispatch ?? null` | Optional chaining is the existing style and required for the fail-closed contract (D-03) in browser/serve mode where `window.go` is undefined |

**Installation:** None — no new packages. All tooling already pinned in `internal/gui/frontend/package.json` and `go.mod`.

**Package Legitimacy Audit:** Not applicable — this phase installs zero external packages (frontend-only edit + tests against existing tooling).

## Architecture Patterns

### System Architecture Diagram

```
                       ┌────────────────────────────────────────────┐
                       │            Frontend (webview / browser)    │
                       │                                            │
   UI modules ────────►│  rpc.ts Backend.dispatch(method, params)   │
   (chat, editor,      │        │                                   │
   preview)            │        ▼                                   │
                       │  wailsBinding() resolver                   │
                       │   window.go.main.desktop.Dispatch ?        │
                       │   └── yes ──► binding(method, JSON) ───┐   │
                       │   └── no  ──► fetch POST /api/dispatch │   │
                       │               + EventSource /api/events│   │
                       └────────────────────────────────────────┼───┘
                                                                │
              ┌─────────────────────────────────────────────────┘
              ▼
   ┌─────────────────────────────┐        ┌──────────────────────────────┐
   │ Desktop mode (Wails v2)     │        │ Serve mode (--serve HTTP)    │
   │ window.go.main.desktop      │        │ /api/dispatch (POST JSON)    │
   │   .Dispatch(method, params) │        │ /api/events  (SSE)           │
   │ cmd/c4drill-gui/main.go:86  │        │ cmd/c4drill-gui/main.go:157+ │
   │ Bind: []interface{}{g} :124 │        │ TestHTTPSmokeE2E covers this │
   └──────────────┬──────────────┘        └──────────────┬───────────────┘
                  ▼                                      ▼
        ┌────────────────────────────────────────────────────────┐
        │        internal/gui/app.App.Dispatch (shared core)     │
        └────────────────────────────────────────────────────────┘
```

Both transports converge on the same `Dispatch(method, params) → string` JSON protocol; only the resolver line decides which path is live. Fixing the resolver cannot fragment the protocol (one namespace, one method).

### Recommended Project Structure

```
cmd/c4drill-gui/
├── main.go            # desktop struct (L74), Dispatch (L86), Bind (L124), newHandler (L157)
├── main_test.go       # TestHTTPSmokeE2E — the GUI-02 e2e gate
└── wails.json         # wails v2 config; frontend:build = npm run build
internal/gui/frontend/
├── package.json       # vitest 4.1.11, tsc --noEmit gates
├── tsconfig.json      # strict: true, noEmit, include: ["src"]
└── src/
    ├── rpc.ts         # THE FIX SITE: WailsWindow interface (L9-19), wailsBinding() (L21-24)
    ├── rpc.test.ts    # NEW (Wave 0): binding-shape unit tests (D-03)
    └── language/c4d.test.ts   # existing vitest suite (pattern reference)
```

### Pattern 1: Wails v2 binding namespace resolution
**What:** Wails v2 exposes bound struct methods on `window.go.<package>.<struct>.<Method>`; generated wrappers literally call `window["go"]["main"]["App"]["Greet"](arg1)`.
**When to use:** Any frontend code touching desktop bindings.
**Example (verified — wails.io/docs/howdoesitwork):**
```javascript
// Generated by Wails v2 (wailsjs/go/main/App.js)
export function Greet(arg1) {
  return window["go"]["main"]["App"]["Greet"](arg1);
}
```
For this project: package `main` + struct `desktop` ⇒ `window.go.main.desktop.Dispatch`.

### Pattern 2: Fail-closed transport resolution (D-03)
**What:** The resolver must return `null` (not throw, not fall through to a phantom fallback) when the binding is absent, so `isDesktop()` is false and the miss is observable.
**When to use:** Any environment-probing resolver that has a browser fallback.
**Example:**
```typescript
function wailsBinding(): ((method: string, params: string) => Promise<string>) | null {
  const w = window as unknown as WailsWindow;
  return w.go?.main?.desktop?.Dispatch ?? null;
}
```
Behavior contract: in serve/browser mode `window.go` is `undefined` → resolver returns `null` → fetch fallback runs (GUI-02 preserved). In desktop mode the binding exists → desktop path runs.

### Pattern 3: vitest window stubbing (D-03 test harness)
**What:** vitest's default environment is `node` (no vitest.config.ts exists; none needed). Stub `window` per test with `vi.stubGlobal`/`vi.unstubAllGlobals` and assert through the exported `backend.dispatch` / `isDesktop()`.
**When to use:** The new `src/rpc.test.ts`.
**Example:**
```typescript
import { afterEach, describe, expect, it, vi } from "vitest";

afterEach(() => vi.unstubAllGlobals());

it("resolves the desktop binding and dispatches through it", async () => {
  const dispatch = vi.fn().mockResolvedValue(JSON.stringify({ ok: true }));
  vi.stubGlobal("window", { go: { main: { desktop: { Dispatch: dispatch } } } });
  expect(isDesktop()).toBe(true);
  await expect(call({ ok: 1 } as never)).resolves.toEqual({ ok: true });
  expect(dispatch).toHaveBeenCalledWith("someMethod", JSON.stringify({ ok: 1 }));
});

it("fails closed when the namespace is absent", () => {
  vi.stubGlobal("window", {});
  expect(isDesktop()).toBe(false);
});
```
(Exact shape per planner/executor; existing `c4d.test.ts` shows the project's vitest style.)

### Anti-Patterns to Avoid
- **Dead fallback chains:** Keeping `main?.App` / `backend?.App` as fallbacks after the fix. They never resolve post-#31 and mask future namespace regressions (explicitly rejected by D-02).
- **Renaming the Go struct to satisfy the frontend:** Would make package `main` expose struct `App` — the Go code would misdescribe itself (rejected by D-01).
- **Importing `wailsjs/go/...` generated modules:** Wails generates those only in dev mode and they are not committed; rpc.ts correctly probes the runtime `window.go` object instead. Keep that approach.
- **Testing through a real webview:** `wails build` needs the wails CLI + CGO/webview (unavailable in CI). The binding-shape mock test (D-03) is the agreed substitute.
- **Touching the fetch fallback or SSE code:** GUI-02 forbids it; the `Backend.dispatch` fetch branch and `startEvents()` must remain byte-identical.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Test runner + assertions | Custom test harness | vitest 4.1.11 (already configured, `npm test`) | Existing suite + zero setup |
| Type checking | Ad-hoc tsc flags | `npm run typecheck` (`tsc --noEmit`, strict) | Project's canonical gate |
| Window stubbing | Manual global juggling with cleanup | `vi.stubGlobal` / `vi.unstubAllGlobals` | Leak-proof per-test isolation |
| Go compile + e2e verification | Custom scripts | `go build ./...` + `go test ./cmd/c4drill-gui/...` | Existing Makefile-less convention |

**Key insight:** This phase's entire risk is a two-identifier mismatch; every tool needed to prove the fix already exists in the repo. Adding anything new is scope creep.

## Runtime State Inventory

> Included: this is a namespace rename/alignment phase (refactor trigger).

| Category | Items Found | Action Required |
|----------|-------------|------------------|
| Stored data | None — verified by repo grep: no datastore, fixture, or config stores `main.App` / `backend.App` / `main.desktop` strings; the namespace exists only inside Wails' runtime injection | none |
| Live service config | None — no external services involved; wails.json carries no namespace strings | none |
| OS-registered state | None — no launchd/systemd/Task Scheduler registrations; wails.json is a build config, not OS state | none |
| Secrets/env vars | None — no env vars reference the namespace (verified by grep) | none |
| Build artifacts | `internal/gui/frontend/dist/assets/*.js` (local, gitignored — only `.gitkeep` is committed) may contain the old resolver from a prior `npm run build`; `go:embed` bakes whatever `dist/` holds at `go build` time | Regenerate via `npm run build` (or at minimum typecheck + vitest; CI regenerates dist before go build) |

**Canonical question answered:** After all files are updated, no runtime system holds the old string — `window.go.*` is injected fresh by Wails at app start, and `dist/` is rebuilt from source. The only stale-artifact risk is a local, uncommitted `dist/` build.

## Common Pitfalls

### Pitfall 1: Guessing the namespace instead of deriving it
**What goes wrong:** Writing `window.go.main.Desktop` (capitalized) or `window.go.desktop.Dispatch` — Wails uses the literal Go package name and the literal struct name.
**Why it happens:** Case conventions differ between Go (type names PascalCase) and the struct here (`desktop`, deliberately lowercase).
**How to avoid:** Derive from source of truth: package `main` (main.go:15) + `type desktop struct` (main.go:74) ⇒ `window.go.main.desktop.Dispatch`. All three identifiers are lowercase.
**Warning signs:** Binding-shape test fails to resolve even after the "fix".

### Pitfall 2: `backend.App` assumed to be a v2 alternative
**What goes wrong:** Treating `go.backend.App` as a legitimate second namespace worth keeping as a fallback.
**Why it happens:** Wails v1 exposed bindings at `window.backend`; the name is a v1 fossil that never existed in any v2 app.
**How to avoid:** D-02 — single canonical path, no fallbacks.
**Warning signs:** Reviewer asks why two paths remain.

### Pitfall 3: vitest environment surprise (`window` undefined)
**What goes wrong:** Test crashes with `window is not defined` instead of cleanly exercising the resolver.
**Why it happens:** No vitest.config.ts → default `node` environment; rpc.ts reads `window` at call time.
**How to avoid:** Stub with `vi.stubGlobal("window", {...})` in each test; restore with `vi.unstubAllGlobals()` in `afterEach`.
**Warning signs:** First test run errors before any assertion.

### Pitfall 4: Stale local `dist/` masking a green build
**What goes wrong:** `go build ./...` passes because `go:embed` accepts whatever `dist/` holds — but the embedded JS still carries the phantom resolver.
**Why it happens:** `dist/` is a build artifact, not tracked source (only `.gitkeep` is committed).
**How to avoid:** Run `npm run build` (which runs grammar + `tsc --noEmit` + vite build) before/with the go build in verification; acceptance criteria assert on **source** (rpc.ts contains `main.desktop`, zero `App?` phantom refs), never on dist output.
**Warning signs:** Desktop binary still broken after a "green" pipeline.

### Pitfall 5: Breaking the serve fallback while editing the resolver
**What goes wrong:** Refactor accidentally changes `Backend.dispatch`'s fetch branch or `isDesktop()` semantics, breaking browser/serve mode.
**Why it happens:** Resolver and fallback live in the same small file.
**How to avoid:** GUI-02 e2e (`go test ./cmd/c4drill-gui/...`) plus the D-03 fail-closed test (stub `window` as `{}` → expect fetch path selection / `isDesktop() === false`) pin both behaviors.
**Warning signs:** e2e regression or the fail-closed test flipping.

## Code Examples

### Current broken resolver (rpc.ts:9-24 — the exact fix site)
```typescript
interface WailsWindow {
  go?: {
    main?: { App?: { Dispatch?: (method: string, params: string) => Promise<string> } };      // phantom
    backend?: { App?: { Dispatch?: (method: string, params: string) => Promise<string> } };   // phantom (v1 fossil)
  };
  // runtime?: {...} — unchanged, already correct
}

function wailsBinding(): ((method: string, params: string) => Promise<string>) | null {
  const w = window as unknown as WailsWindow;
  return w.go?.main?.App?.Dispatch ?? w.go?.backend?.App?.Dispatch ?? null;   // never matches post-#31
}
```

### Target state (per D-01/D-02)
```typescript
interface WailsWindow {
  go?: {
    main?: { desktop?: { Dispatch?: (method: string, params: string) => Promise<string> } };
  };
  // runtime?: {...} unchanged
}

function wailsBinding(): ((method: string, params: string) => Promise<string>) | null {
  const w = window as unknown as WailsWindow;
  return w.go?.main?.desktop?.Dispatch ?? null;
}
```

### Go binding (source of truth — do not modify)
```go
// cmd/c4drill-gui/main.go
package main                                    // L15

type desktop struct { ... }                     // L74 — lowercase struct name
func (g *desktop) Dispatch(method, params string) (string, error) { ... }  // L86
Bind: []interface{}{g},                         // L124 — the only bound value
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Wails v1 `window.backend` bindings | Wails v2 `window.go.<pkg>.<struct>` bindings | v2.0 (2022) | `go.backend.*` could never exist in this v2.15 project — pure fossil |
| `main.gui` bound struct (pre-#31) | `main.desktop` (commit dc1df8b, #37 restructure) | 2026-09-02 | Namespace the frontend must target |
| Wails v2 runtime `window.go` probing | Wails v3 generated SDK (`bindings/` modules, `$Call.ByID`) | v3 (current) | Not applicable — project pins v2.15.0; do not adopt v3 patterns |

**Deprecated/outdated:**
- `window.go.main.App` / `window.go.backend.App` in this repo: never valid post-#31.

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | vitest 4.1.11 runs in default `node` environment (no vitest.config.ts present; none required for stubGlobal-based tests) | Pattern 3, Validation Architecture | LOW — worst case a `// @vitest-environment jsdom`-style directive or minimal config is added; no jsdom dependency is installed (discretion area) |
| A2 | Existing `c4d.test.ts` passes today (baseline green assumed; will be confirmed by the phase's vitest run) | Validation Architecture | LOW — if red, it's a pre-existing failure to report, not caused by this phase |

All other claims verified against the codebase (`[VERIFIED: repo grep/read]`) or official Wails docs (`[CITED: wails.io/docs/howdoesitwork]` via Context7).

## Open Questions

1. **Typed constant for the namespace path?**
   - What we know: D-02 requires exactly `go.main.desktop.Dispatch` in interface + resolver; CONTEXT leaves a typed constant to planner discretion.
   - What's unclear: whether the indirection earns its keep in a 2-line surface.
   - Recommendation: optional; prefer the plain optional chain matching existing style unless the executor sees duplication.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go | `go build ./...`, `go test ./cmd/c4drill-gui/...` | ✓ | go1.26.5 darwin/arm64 | — |
| Node.js | vitest, tsc, vite | ✓ | v26.8.1 | — |
| npm | frontend scripts | ✓ | 11.19.0 | — |
| wails CLI | real desktop `wails build` verification | ✗ | — | Not needed — D-03/D-04 define the CI-verifiable contract (binding-shape mock tests + structural grep); real desktop run is a manual smoke |

**Missing dependencies with no fallback:** None that block this phase (wails CLI absence is accommodated by D-03/D-04).

**Missing dependencies with fallback:** wails CLI — manual desktop smoke after merge, when the operator has the toolchain.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | vitest ^4.1.11 (frontend), Go testing + testify (backend) |
| Config file | none (vitest defaults; tsconfig.json strict) |
| Quick run command | `cd internal/gui/frontend && npx vitest run src/rpc.test.ts` |
| Full suite command | `cd internal/gui/frontend && npm test && npm run typecheck && npm run build` + `go build ./... && go test ./cmd/c4drill-gui/...` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| GUI-01 | Resolver finds `window.go.main.desktop.Dispatch` and dispatches through it | unit (TDD RED→GREEN) | `cd internal/gui/frontend && npx vitest run src/rpc.test.ts` | ❌ Wave 0 (`src/rpc.test.ts`) |
| GUI-01 | Fail-closed: absent namespace → null transport / `isDesktop() === false` | unit | same as above | ❌ Wave 0 (same file) |
| GUI-01 | Structural: zero phantom refs (`main?.App`, `backend?.App`, `App?:`) in frontend source | source assertion (grep) | `! grep -rn "main?\.App\|backend?\.App" internal/gui/frontend/src/` | n/a (command) |
| GUI-02 | Serve e2e stays green (HTTP dispatch, SSE, static assets) | integration e2e | `go test ./cmd/c4drill-gui/...` | ✅ `cmd/c4drill-gui/main_test.go` (TestHTTPSmokeE2E, TestEventHubBroadcast) |
| GUI-01/02 | Types compile under strict mode | compile gate | `cd internal/gui/frontend && npm run typecheck` | n/a (command) |

### Sampling Rate
- **Per task commit:** `cd internal/gui/frontend && npx vitest run`
- **Per wave merge:** frontend `npm test && npm run typecheck` + `go build ./... && go test ./cmd/c4drill-gui/...`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `internal/gui/frontend/src/rpc.test.ts` — covers GUI-01 (binding-shape resolve + dispatch, fail-closed) — the TDD RED test

*(Otherwise none — Go e2e and toolchain already exist.)*

## Security Domain

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | no | Local desktop app; no auth surface touched |
| V3 Session Management | no | No sessions in this path |
| V4 Access Control | no | Single-user local process; bindings localhost-only |
| V5 Input Validation | unchanged | Existing: `backend.Dispatch` validates method/params (unknown methods rejected — e2e asserts 400); this phase changes transport resolution only, not input handling |
| V6 Cryptography | no | No crypto involved |
| V14 File/Download | unchanged | Export path untouched (covered by e2e) |

### Known Threat Patterns for Wails v2 desktop + local HTTP fallback

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Overly broad binding exposure (binding more of the app than needed) | Elevation | Already mitigated: single `Dispatch` method funnels through app-layer validation; phase does not expand the bound surface |
| Resolver silently degrading to the wrong transport | Tampering (silent behavior divergence) | D-03 fail-closed contract: missing namespace ⇒ observable null transport, not silent fallback confusion |
| localhost HTTP fallback exposure | Information Disclosure | Pre-existing design (127.0.0.1 bind); untouched by this phase; e2e keeps it green |

**Phase security verdict:** No new attack surface. The fix strictly narrows behavior (removes two never-valid resolution paths). ASVS L1: no additional controls required.

## Sources

### Primary (HIGH confidence)
- Context7 `/websites/wails_io` → wails.io/docs/howdoesitwork — Wails v2 binding namespace (`window["go"]["main"]["App"]["Greet"]`), `Bind` option, v1→v2 `window.backend`→`window.go` migration note
- Context7 `/websites/v3_wails_io` — Wails v3 binding model checked and explicitly ruled out (project pins v2)
- Repo reads: `internal/gui/frontend/src/rpc.ts`, `cmd/c4drill-gui/main.go`, `cmd/c4drill-gui/main_test.go`, `cmd/c4drill-gui/wails.json`, `internal/gui/frontend/package.json`, `tsconfig.json`, `vite.config.ts`, `go.mod`
- Repo greps: phantom-namespace census (rpc.ts is the only source file referencing them), runtime-state census, dist tracking status
- GitHub issue #38 (`gh issue view 38`) — bug report, fix direction, acceptance criteria

### Secondary (MEDIUM confidence)
- None needed

### Tertiary (LOW confidence)
- A1 (vitest node-environment default) — standard vitest behavior, confirmed by absence of config; validated by running the tests in Wave 0

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — everything project-pinned and read from manifests; no new packages
- Architecture: HIGH — namespace math verified against official docs AND the actual bound struct; fix site isolated by repo-wide grep
- Pitfalls: HIGH — each pitfall drawn from this repo's actual state (v1 fossil name, missing vitest config, untracked dist)

**Research date:** 2026-09-03
**Valid until:** 2026-10-03 (stable — project-pinned toolchain; Wails v2 namespace semantics are frozen by the v2.15.0 pin)
