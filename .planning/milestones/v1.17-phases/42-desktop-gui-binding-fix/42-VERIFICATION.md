---
phase: 42-desktop-gui-binding-fix
status: human_needed
score: 7/7
verified: 2026-09-03
method: inline goal-backward verification (no verifier subagent available in runtime)
requirements_covered: [GUI-01, GUI-02]
human_verification_count: 1
gaps: []
---

# Phase 42: Desktop GUI Binding Fix — Verification

**Goal:** Desktop-window mode works again — the frontend RPC layer calls the Wails-generated namespace that matches the actually bound Go struct (`main.desktop`).

## Observable Truths (roadmap SCs + PLAN must_haves merged)

| # | Truth | Evidence | Status |
|---|-------|----------|--------|
| 1 | rpc.ts calls `window.go.main.desktop.Dispatch`; no phantom `main.App` / `backend.App` references remain | rpc.ts:22 `w.go?.main?.desktop?.Dispatch ?? null`; `! grep -rn "App?" internal/gui/frontend/src/` → zero matches | VERIFIED |
| 2 | Desktop RPC works: every method rpc.ts invokes exists on the generated `main.desktop` binding | rpc.ts invokes exactly one binding function — `Dispatch(method, params)` (rpc.ts:38-41), which IS the bound method (cmd/c4drill-gui/main.go:86); binding-shape test proves resolve+dispatch through mocked `window.go.main.desktop.Dispatch` (vitest 27/27) | VERIFIED |
| 3 | `--serve` HTTP fallback unchanged, e2e green | git diff of rpc.ts limited to go-declaration + resolver lines (fetch branch byte-identical); `go test ./cmd/c4drill-gui/...` → ok (TestHTTPSmokeE2E, TestEventHubBroadcast) | VERIFIED |
| 4 | Editor/preview/chat/export RPC flows through `main.desktop.Dispatch` in desktop mode | single transport resolver (`wailsBinding`) feeds `Backend.dispatch`; UI layers call `backend.dispatch`/`call` only — no per-feature transport code | VERIFIED |
| 5 | Missing desktop binding resolves to null transport (observable miss) | rpc.test.ts "fails closed when the desktop binding is absent" — `isDesktop()` false with `window.go = {}` | VERIFIED |
| 6 | Raw `""`/`"null"` binding results normalize to null | rpc.test.ts "normalizes raw null and empty-string binding results to null" (both raws) | VERIFIED |
| 7 | No phantom namespace anywhere in frontend source | structural grep `App?` over `internal/gui/frontend/src/` → zero matches (D-04 gate) | VERIFIED |

**Score: 7/7 verified**

## Artifacts

| Artifact | Provides | Contains-check | Status |
|----------|----------|----------------|--------|
| internal/gui/frontend/src/rpc.ts | Desktop transport resolver bound to real Wails namespace | `main?.desktop?.Dispatch` (line 22) | PASS (implemented, exercised by tests) |
| internal/gui/frontend/src/rpc.test.ts | Binding-shape unit tests: resolve+dispatch, fail-closed | `main.desktop` (2 occurrences) | PASS (implemented, passing) |

## Key Links (wiring)

| From | To | Via | Pattern check | Status |
|------|----|----|---------------|--------|
| rpc.test.ts | rpc.ts | vitest import + vi.stubGlobal window mock | `from "./rpc"` (line 9), `vi.stubGlobal("window", ...)` (line 16) | WIRED |
| rpc.ts wailsBinding() | Wails-generated binding | optional chain | `go?.main?.desktop?.Dispatch` (line 22) | WIRED |

## Requirements Coverage

| Requirement | Definition | Status |
|-------------|-----------|--------|
| GUI-01 | Desktop RPC calls the Wails-generated `main.desktop` namespace | Complete (REQUIREMENTS.md, plan 42-01) |
| GUI-02 | `--serve` HTTP fallback unchanged, e2e green | Complete (REQUIREMENTS.md, plan 42-01) |

## Gate Results (D-04 + plan verification block)

- `npx vitest run` (frontend) — 27/27 passed
- `npm run typecheck` (tsc --noEmit, strict) — exit 0
- `npm run build` — exit 0 (dist regenerated for go:embed)
- `! grep -rn "App?" internal/gui/frontend/src/` — zero matches
- `go build ./...` — exit 0
- `go test ./cmd/c4drill-gui/...` — ok
- Debt-marker scan (TBD/FIXME/XXX) over phase files — zero markers

## Anti-Patterns

None found in phase-modified files.

## Concurrent-Phase Observation

A full `go test ./...` run during this phase's execution showed failures confined to `cmd/c4drill/check_test.go` (Phase 41 surface, `check` command — consistent with that phase's in-flight TDD cycle) and an untracked `internal/render/deterministic_test.go` (Phase 40 RED artifact). None touch this phase's surface; this phase's scoped gates are all green.

## Human Verification Items

### 1. Real desktop-window smoke test (GUI-01 end-to-end)
expected: With the wails toolchain installed: `wails build` in cmd/c4drill-gui, run the binary, open a project — editor input renders preview, chat and export work (all RPC flows through window.go.main.desktop.Dispatch).
why manual: `wails` CLI + CGO/webview unavailable in CI; D-03/D-04 define the CI-verifiable substitute (binding-shape tests + structural gates, all green).
result: [pending — persisted to 42-HUMAN-UAT.md]

## Conclusion

All automatable checks pass (7/7 truths). One manual verification remains: real desktop-window smoke with the wails toolchain — persisted as UAT. Status: **human_needed** (human items present; no gaps).
