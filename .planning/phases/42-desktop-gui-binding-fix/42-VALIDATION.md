---
phase: 42
slug: desktop-gui-binding-fix
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-09-03
---

# Phase 42 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | vitest 4.1.11 (frontend) + go test / testify (backend e2e) |
| **Config file** | none — vitest defaults, `internal/gui/frontend/tsconfig.json` (strict) |
| **Quick run command** | `cd internal/gui/frontend && npx vitest run src/rpc.test.ts` |
| **Full suite command** | `cd internal/gui/frontend && npm test && npm run typecheck` && `go build ./... && go test ./cmd/c4drill-gui/...` |
| **Estimated runtime** | ~20 seconds |

---

## Sampling Rate

- **After every task commit:** Run `cd internal/gui/frontend && npx vitest run`
- **After every plan wave:** Run full suite command (frontend gates + go build + go e2e)
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 42-01-01 | 01 | 1 | GUI-01 | — | N/A | unit (TDD RED) | `cd internal/gui/frontend && npx vitest run src/rpc.test.ts` | ❌ W0 | ⬜ pending |
| 42-01-02 | 01 | 1 | GUI-01 | — | Resolver narrowed to one canonical path | unit (GREEN) + source grep | `cd internal/gui/frontend && npx vitest run && ! grep -rn "main?\.App\|backend?\.App" src/` | ✅ (test file from 42-01-01) | ⬜ pending |
| 42-01-03 | 01 | 1 | GUI-01, GUI-02 | — | N/A | typecheck + compile + e2e | `cd internal/gui/frontend && npm run typecheck` && `go build ./... && go test ./cmd/c4drill-gui/...` | ✅ (`cmd/c4drill-gui/main_test.go`) | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/gui/frontend/src/rpc.test.ts` — binding-shape tests for GUI-01 (resolve + dispatch through mocked `window.go.main.desktop.Dispatch`; fail-closed when namespace absent) — the TDD RED test

*Otherwise: existing infrastructure covers all phase requirements (vitest configured, Go e2e exists).*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Real desktop window: editor/preview/chat/export function through Wails bindings | GUI-01 | `wails` CLI + CGO/webview unavailable in CI (D-03/D-04 define the CI-verifiable substitute) | With wails toolchain installed: `wails build` in `cmd/c4drill-gui`, run binary, open a project, edit → preview renders, chat + export work |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 30s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
