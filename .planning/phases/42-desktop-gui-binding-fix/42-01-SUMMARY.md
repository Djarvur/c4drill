---
phase: 42-desktop-gui-binding-fix
plan: 01
subsystem: ui
tags: [wails, desktop, rpc, vitest, tdd, frontend]

# Dependency graph
requires:
  - phase: 32-desktop-gui (issue #31/#37 restructure)
    provides: Wails v2 shell binding the `desktop` struct in package main; rpc.ts dual-transport resolver
provides:
  - Desktop transport resolver bound to the real Wails-generated namespace window.go.main.desktop.Dispatch
  - Binding-shape vitest suite (rpc.test.ts) proving resolve+dispatch, fail-closed, and null normalization
  - Structural gate: zero phantom-namespace references (main?.App / backend?.App) in frontend source
affects: [desktop-gui, frontend-rpc, future wails-bound surfaces]

# Tech tracking
tech-stack:
  added: []
  patterns: [vi.stubGlobal window mocking for transport resolution, fail-closed namespace probing]

key-files:
  created:
    - internal/gui/frontend/src/rpc.test.ts
  modified:
    - internal/gui/frontend/src/rpc.ts

key-decisions:
  - "Single canonical namespace go?.main?.desktop?.Dispatch ?? null — no typed constant, no dead fallbacks (D-02; planner discretion resolved to plain optional chain)"

patterns-established:
  - "window-stub unit testing: vi.stubGlobal('window', {go}) + vi.unstubAllGlobals() in afterEach; no jsdom, no new deps"

requirements-completed: [GUI-01, GUI-02]

# Metrics
duration: 4min
completed: 2026-09-03
---

# Phase 42 Plan 01: Desktop GUI Binding Fix Summary

**Frontend transport resolver now calls the Wails-generated window.go.main.desktop.Dispatch (matching the actually bound Go struct), developed RED-first with a binding-shape vitest suite; serve HTTP path byte-identical**

## Performance

- **Duration:** 4 min
- **Started:** 2026-09-03T19:12:20Z
- **Completed:** 2026-09-03T19:16:36Z
- **Tasks:** 3
- **Files modified:** 2 (1 created, 1 modified)

## Accomplishments

- Desktop-window RPC restored: `wailsBinding()` resolves exactly `window.go.main.desktop.Dispatch` — the namespace Wails v2 generates for `package main` + `type desktop struct` + `Bind: []interface{}{g}` (cmd/c4drill-gui/main.go L15/L74/L86/L124)
- Phantom namespaces (`main?.App`, `backend?.App`) eradicated from frontend source; structural grep `App?` over internal/gui/frontend/src returns zero matches (D-04)
- Fail-closed contract pinned by tests: missing binding ⇒ `isDesktop()` false and null transport — the miss is observable, never silent dead UI (D-03)

## Task Commits

Each task was committed atomically:

1. **Task 1: RED — binding-shape tests for the desktop transport** - `ee88c49` (test)
2. **Task 2: GREEN — align resolver to window.go.main.desktop.Dispatch** - `e743d48` (feat)
3. **Task 3: structural gate + full regression** - verification-only (no code delta; gates all green)

_Note: TDD plan — RED (test) → GREEN (feat); no REFACTOR commit needed, the GREEN diff is the minimal 2-line resolver change._

## Files Created/Modified

- `internal/gui/frontend/src/rpc.test.ts` — binding-shape tests (GUI-01): resolve+dispatch through mocked window.go.main.desktop.Dispatch, fail-closed when absent, raw "null"/"" normalization
- `internal/gui/frontend/src/rpc.ts` — WailsWindow.go declaration and wailsBinding() resolver narrowed to the single canonical namespace; fetch fallback, startEvents(), runtime block, and all exports byte-identical (GUI-02)

## Decisions Made

- Plain optional chain `w.go?.main?.desktop?.Dispatch ?? null` instead of a typed constant — research question resolved during planning; the chain mirrors the existing style and adds no indirection
- Case 3 (normalization) exercises the binding mock through the public `backend.dispatch` surface only; no direct `wailsBinding()` export was added (kept the module surface unchanged)

## Deviations from Plan

### Auto-fixed Issues

**1. [Observation - plan forecast correction] Case 3 is RED pre-fix, not pre-passing**
- **Found during:** Task 1 (RED run)
- **Issue:** Plan forecast "Cases 2 and 3 pass already" against unmodified rpc.ts; actually Case 3 fails pre-fix because the desktop mock is unreachable through the phantom resolver, so dispatch falls through to the HTTP fetch branch, which throws (`TypeError: Failed to parse URL from /api/dispatch`) in the node test environment
- **Fix:** No code change — this is additional RED evidence of the exact issue #38 symptom ("every desktop call falls through to a failing HTTP fetch"); Case 3 goes GREEN with the resolver fix like Case 1
- **Files modified:** none beyond planned scope
- **Verification:** RED run exit=1 (2 failed | 1 passed); GREEN run 27/27 passed
- **Committed in:** ee88c49 (test) / e743d48 (fix)

---

**Total deviations:** 1 forecast correction (no auto-fix required)
**Impact on plan:** None — strictly stronger RED evidence; final state matches all plan acceptance criteria.

## Issues Encountered

- Transient shell interpolation: the GREEN commit message initially lost the word `desktop` to zsh command substitution of backticks; amended the message in place (`e743d48`) using a message file. No code impact.
- Executed inline-sequential (orchestrator runtime has no subagent Agent tool) instead of Pattern A worktree isolation — same task order, same gates; commits explicitly file-scoped because Phases 40/41 run concurrently in this working tree. No concurrent-phase failures appeared in `go build ./...` or the serve e2e during this run.

## User Setup Required

None - no external service configuration required.

## Verification (plan-level `<verification>` + D-04 gates)

- `cd internal/gui/frontend && npx vitest run` — 27/27 passed (3 binding-shape + 24 grammar)
- `cd internal/gui/frontend && npm run typecheck` — exit 0 (tsc --noEmit, strict)
- `cd internal/gui/frontend && npm run build` — exit 0 (grammar + tsc + vite build; dist regenerated for go:embed)
- `! grep -rn "App?" internal/gui/frontend/src/` — zero matches
- `go build ./...` — exit 0
- `go test ./cmd/c4drill-gui/...` — ok (TestHTTPSmokeE2E + TestEventHubBroadcast green; GUI-02 serve path untouched)
- Manual (post-merge, per 42-VALIDATION.md): `wails build` + desktop-window smoke when the operator has the wails toolchain — CI-verifiable contract is D-03/D-04 by design

## Self-Check: PASSED

- key-files exist on disk (rpc.ts, rpc.test.ts)
- `git log --grep="42-01"` returns 2 production commits (ee88c49, e743d48)
- All task acceptance criteria re-run and passing (see Verification)

## Next Phase Readiness

- Phase 42 complete pending verification: GUI-01 satisfied (desktop RPC via real namespace), GUI-02 satisfied (serve path untouched, e2e green)
- Real desktop-window smoke remains the one manual verification item (wails toolchain not present here, by design)

---
*Phase: 42-desktop-gui-binding-fix*
*Completed: 2026-09-03*
