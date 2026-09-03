# Phase 42: Desktop GUI Binding Fix - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-09-03
**Phase:** 42-Desktop GUI Binding Fix
**Areas discussed:** Namespace alignment, Resolver surface, Verification contract
**Mode:** `--auto` (yolo runtime) — all selections auto-chosen and logged

---

## Namespace alignment

| Option | Description | Selected |
|--------|-------------|----------|
| Update rpc.ts to `main.desktop` | Frontend matches reality; Go binding untouched (issue's "cleaner" direction) | ✓ |
| Rename Go struct/bind to `App` | Generated namespace matches current rpc.ts, but forces the Go side to lie about identity | |

**User's choice:** [auto] Q: "Which side moves?" → Selected: "Update rpc.ts to `main.desktop`" (recommended default — scout confirmed struct `desktop` in package `main` at cmd/c4drill-gui/main.go:74).

---

## Resolver surface

| Option | Description | Selected |
|--------|-------------|----------|
| Single canonical namespace | Window typing + resolver reference only `go.main.desktop.Dispatch` | ✓ |
| Keep old names as fallback chain | `main.desktop ?? main.App ?? backend.App` — dead code masks regressions | |

**User's choice:** [auto] Q: "Fallbacks?" → Selected: "Single canonical namespace" (recommended default — neither old namespace ever existed post-#31).

---

## Verification contract

| Option | Description | Selected |
|--------|-------------|----------|
| Mocked binding-shape unit test + structural gate + builds + serve e2e | TDD RED (resolver can't find main.desktop today); no GUI window needed in CI | ✓ |
| Manual runIde-style smoke test only | Not automatable here; not repeatable | |

**User's choice:** [auto] Q: "How to verify without launching a Wails window?" → Selected: "Mocked binding-shape unit test + structural gate + builds + serve e2e" (recommended default).

---

## Claude's Discretion

- Frontend test harness/file placement; optional typed-namespace constant extraction.

## Deferred Ideas

None — discussion stayed within phase scope.
