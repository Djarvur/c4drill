# Phase 42: Desktop GUI Binding Fix - Context

**Gathered:** 2026-09-03
**Status:** Ready for planning

<domain>
## Phase Boundary

Restore desktop-window RPC: the frontend's transport resolver calls the Wails-generated namespace that matches the actually bound Go struct (`main.desktop`), fixing the desktop transport broken since the #31/#37 restructure. `--serve` HTTP fallback untouched (issue #38).

</domain>

<decisions>
## Implementation Decisions

### Namespace alignment direction
- **D-01:** Update the FRONTEND to the real namespace (issue's "cleaner" direction): `internal/gui/frontend/src/rpc.ts` resolves `window.go.main.desktop.Dispatch`. Renaming the Go side so the generated namespace matches rpc.ts's current expectation would force the bound struct to be named `App` in package `main`/`backend` — making the Go code lie about what it is. Go binding stays as-is.

### Resolver surface
- **D-02:** Single canonical namespace, no dead fallbacks: the `Window` interface declaration (rpc.ts:11-12) and the resolver chain (rpc.ts:23) reference exactly `go.main.desktop.Dispatch`. The current `main.App` / `backend.App` alternatives never existed post-#31 (`main.gui` was the only pre-#31 name) — keeping them as fallbacks would preserve dead code that masks regressions.

### Verification contract (no GUI window in CI)
- **D-03:** Binding-shape unit test: mock `window.go.main.desktop.Dispatch` and assert the transport resolves and dispatches through it; a second case asserts dispatch fails closed (returns null transport / surfaces the miss) when the namespace is absent — this is the TDD RED test (current code fails it: resolver never finds `main.desktop`).
- **D-04:** Structural gate: zero references to the phantom namespaces remain (`main?.App`, `backend?.App`) in frontend source; frontend type-check/build green; `go build ./...` green; the existing `--serve` e2e suite stays green (GUI-02 — the HTTP path must be untouched).

### Claude's Discretion
- Test runner/file placement inside the frontend per its existing test setup (planner inspects frontend tooling).
- Whether the resolver refactor warrants extracting the namespace path into a typed constant — minor, planner's call.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Issue
- GitHub issue #38 (Djarvur/c4drill) — full bug description, namespace math (`window.go.<package>.<struct>`), fix direction

### Code
- `internal/gui/frontend/src/rpc.ts` (lines 11-12, 23) — phantom namespace declarations and resolver chain (the fix site)
- `cmd/c4drill-gui/main.go` (line 74 `type desktop struct`, `Bind: []interface{}{g}` ~line 124, `Dispatch` method) — the namespace Wails actually generates: `window.go.main.desktop`
- `cmd/c4drill-gui/main.go` SetEventSink("backend") — event channel rpc.ts already consumes correctly (unaffected)

### Project docs
- `.planning/REQUIREMENTS.md` — GUI-01..02 definitions
- Commit dc1df8b (#37 restructure) — where `main.gui` became `main.desktop` (git show dc1df8b if history needed)

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- rpc.ts `Backend` class already isolates transport selection behind `resolveWails()`-style resolution — the fix is localized to the Window typing + resolver line
- Existing serve-transport e2e harness proves the protocol layer; desktop fix rides the same `Dispatch(method, params) → string` contract

### Established Patterns
- Single `Dispatch` method carries the whole JSON protocol (desktop struct comment) — one namespace, one method: the fix cannot fragment
- Frontend tests mock `window` for transport resolution (planner verifies the exact harness)

### Integration Points
- `internal/gui/frontend/src/rpc.ts` — only frontend file referencing the phantom namespaces (roadmapper verified)
- Wails generated bindings (build-time, `window.go.*`) — never committed; the test mocks them

</code_context>

<specifics>
## Specific Ideas

From the issue: the reporter recommends "update rpc.ts to match reality" over rebinding — adopted as D-01.

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 42-Desktop GUI Binding Fix*
*Context gathered: 2026-09-03*
