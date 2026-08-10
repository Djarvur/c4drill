---
phase: 34
status: clean
depth: standard
files_reviewed: 5
findings:
  critical: 0
  warning: 0
  info: 2
  total: 2
created: 2026-08-10
---

# Phase 34 — Code Review

**Scope (from SUMMARY.md):**
- internal/render/wrap.go
- internal/render/wrap_internal_test.go
- internal/render/labels.go
- internal/render/converter.go
- internal/render/labels_test.go

**Diff base:** parent of 0185075 (pre-phase-34)

## Summary

5 files, 63 insertions / 55 deletions. Changes deliver LABEL-01 (edge labels as borderless HTML-table rectangles via `SetLabelHTML`) and LABEL-02 (word-boundary-only wrapping, `splitLongWord` removed), with COMPAT-01 enforced by the canonicalDOT (DI-1) goldens. Full suite (`go test ./...`) green; REF-05 golden regression found in-suite and fixed (ebb4ac9).

## Verified Checks

- **wrap.go:** over-budget branch now flushes pending line and emits the whole word unsplit (D-05); `splitLongWord` deleted; `utf8` import still required (RuneCountInString in wordLen). `fitsOnLine`/`flushIfPending`/`wrapAndEscape` untouched — in-budget output unchanged.
- **labels.go:** `buildEdgeLabel` builds `<table border="0" cellpadding="0" cellspacing="0">` via the shared row helpers; escaping inherited from `wrapAndEscape` (V5); 2-row floor applied to `labelMaxCharsNoIcon` input (D-03); nil/empty returns "" (parity with prior behavior).
- **converter.go:** `SetLabelHTML` for non-empty labels (graphviz-13 HTML-ness), `SetLabel("")` shim for empty labels preserving the `label=""` attribute the v1.9 goldens pin; rationale documented inline.
- **Tests:** re-asserted `TestWrapText` (unsplit overflow) and `TestEdgeLabelGeneration` (HTML-table form, D-04 rectangle-always); no `t.Parallel()` (WASM engine); no byte-exact full-DOT `require.Equal` (canonical comparisons only).
- **COMPAT-01:** `go test ./cmd/c4drill/... ./internal/graph/...` green — canonicalDOT goldens byte-stable.

## Findings

### Info

**1. [INFO] Defensive rowCount==0 guard in buildEdgeLabel**
- File: internal/render/labels.go
- The `rowCount == 0 → return ""` guard is technically redundant with `writeTechnologyRow`/`writeDescriptionRow` (both skip empty input); without it the builder would emit an empty `<table></table>`. The guard is correct and preserves the old `""` return — keep as documentation of intent.

**2. [INFO] Empty-label SetLabel("") path is a golden-parity shim**
- File: internal/render/converter.go
- The `else { e.SetLabel("") }` branch exists solely to reproduce the `label=""` attribute the v1.9 golden pins (SafeSetHTML("") omits the attribute). If goldens are ever re-baselined, the branch could be simplified — leave as-is for COMPAT-01.

## Recommendation

No action required. Changes are minimal, well-scoped, and regression-guarded.
