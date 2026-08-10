---
phase: 34-label-formatting-fixes-0-1-plans
plan: 02
subsystem: render
tags: [graphviz, dot, html-label, edge-label, setlabelhtml]

# Dependency graph
requires:
  - phase: 34 (plan 01)
    provides: word-boundary-only wrapText semantics (LABEL-02) consumed by wrapAndEscape in edge-label rows
  - phase: 33
    provides: canonical.Canonical (DI-1) golden helper enforcing COMPAT goldens
provides:
  - edge labels as borderless HTML-table rectangles (LABEL-01): [Technology] row + wrapped Description row, LabelRatio sizing with 2-row floor, rectangle always emitted (D-01..D-04)
  - e.SetLabelHTML edge emission (graphviz-13 HTML-ness preserved)
affects: [future label styling work, SVG/HTML output snapshots]

# Tech tracking
tech-stack:
  added: []
  patterns: [HTML-table rectangle edge labels via writeLabelTableStart/writeTechnologyRow/writeDescriptionRow/writeLabelTableEnd + labelMaxCharsNoIcon with 2-row floor]

key-files:
  created: []
  modified: [internal/render/labels.go, internal/render/converter.go, internal/render/labels_test.go]

key-decisions:
  - "buildEdgeLabel now emits the borderless HTML-table rectangle (D-01/D-02/D-04); maxChars = labelMaxCharsNoIcon(max(rowCount,2)) per D-03"
  - "createEdge uses SetLabelHTML for non-empty HTML labels (graphviz-13 HTML-ness); empty labels keep SetLabel(\"\") to preserve the label=\"\" attribute emission the v1.9 goldens pin (COMPAT-01)"
  - "TestEdgeLabelGeneration re-asserted: checkNewline removed; all four cases assert <table border=\"0\" (D-04 rectangle-always)"

patterns-established:
  - "Edge label rectangle: writeLabelTableStart → writeTechnologyRow(tech) → writeDescriptionRow(desc) → writeLabelTableEnd"

requirements-completed: [LABEL-01, COMPAT-01]

# Metrics
duration: 20min
completed: 2026-08-10
---

# Phase 34 Plan 02: edge labels as borderless HTML-table rectangles Summary

**Edge labels now render as borderless HTML-table rectangles: `[Technology]` row + `<BR/>`-wrapped Description row, width derived from LabelRatio via labelMaxCharsNoIcon with the 2-row floor (D-01..D-04), emitted through e.SetLabelHTML — with the v1.9 golden's empty-label attribute emission preserved (COMPAT-01)**

## Performance

- **Duration:** 20 min
- **Started:** 2026-08-10T13:32:30Z
- **Completed:** 2026-08-10T13:52:00Z
- **Tasks:** 3 (TDD: RED/GREEN/REFACTOR-verify)
- **Files modified:** 3

## Accomplishments
- Re-asserted `TestEdgeLabelGeneration`: removed the `checkNewline` (`\n`) assertion, all four cases now assert the HTML-table form (`<table border="0"`, `<i>[gRPC]</i>`, `<i>[TCP]</i>`) — D-04 rectangle-always for tech-only/desc-only edges (RED)
- `buildEdgeLabel` rewritten: borderless rectangle via `writeLabelTableStart`/`writeTechnologyRow`/`writeDescriptionRow`/`writeLabelTableEnd`; `maxChars = labelMaxCharsNoIcon(max(rowCount, 2))` (D-03 2-row floor); empty EdgeLabel → "" (GREEN)
- `createEdge` switched to `e.SetLabelHTML(...)` — HTML-ness survives the graphviz-13 string-dict round-trip (same `agsafeset_html` path nodes use)
- COMPAT-01 confirmed: `go test ./internal/render/... ./cmd/c4drill/... ./internal/graph/...` all green — canonicalDOT goldens byte-stable (REF-05 `TestReference_BackwardCompat` passes)

## Task Commits

Each task was committed atomically:

1. **Task 1 (RED): Re-assert TestEdgeLabelGeneration for the HTML-table form** - `acb77b7` (test)
2. **Task 2 (GREEN): buildEdgeLabel emits the borderless HTML-table rectangle; createEdge uses SetLabelHTML** - `1ebdafa` (feat)
3. **Task 3 (REFACTOR + COMPAT-01): full package and golden regression run** - `ebb4ac9` (fix, deviation — see below)

**Plan metadata:** (pending `docs(34-02): complete plan`)

## TDD Gate Compliance

| Gate | Commit | Status |
|------|--------|--------|
| RED | `test(34-02): re-assert edge labels as borderless HTML-table rectangles` (acb77b7) | ✓ — all 4 cases failed for the right reason (plain-string labels, no `<table border="0"`) |
| GREEN | `feat(34-02): render edge labels as borderless HTML-table rectangles` (1ebdafa) | ✓ — all 4 edge-label cases + unit-label guard tests pass |
| REFACTOR | `fix(34-02): preserve empty edge label emission...` (ebb4ac9) | ✓ — deviation auto-fix (below); tests still pass |

## Files Created/Modified
- `internal/render/labels.go` - `buildEdgeLabel` rewritten: borderless HTML-table rectangle (tech row + wrapped description row), `labelMaxCharsNoIcon` with 2-row floor, empty-label → ""
- `internal/render/converter.go` - `createEdge`: `SetLabelHTML` for non-empty HTML labels, `SetLabel("")` for empty labels (golden parity)
- `internal/render/labels_test.go` - `TestEdgeLabelGeneration` re-asserted (HTML-table form, D-04 rectangle-always, `checkNewline` removed)

## Decisions Made
- Empty-edge-label emission preserved as `label=""` (SetLabel path) — required for COMPAT-01 byte-stability of the v1.9 goldens
- Row styling follows unit-label conventions (reused `writeTechnologyRow`'s `<i>` italic form — discretion area resolved in favor of consistency)

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Empty edge labels dropped the `label=""` attribute (REF-05 golden regression)**
- **Found during:** Task 3 (REFACTOR + COMPAT-01 golden run)
- **Issue:** `SetLabelHTML("")` (SafeSetHTML) omits the label attribute entirely for non-nil empty EdgeLabels. The committed v1.9 golden (`multilevel.expanded.dot`) pins explicit `label=""` on 56 label-less edges and `label="\E"` on 1 nil-label edge; the canonicalDOT comparison failed (COMPAT-01 violated by my own GREEN change).
- **Fix:** `createEdge` now calls `SetLabelHTML` only when `buildEdgeLabel` returns non-empty; empty labels keep the plain `SetLabel("")` path — byte-identical emission restored.
- **Files modified:** internal/render/converter.go
- **Verification:** `go test ./internal/graph/ -run TestReference_BackwardCompat -v` passes; full `./cmd/c4drill/... ./internal/graph/... ./internal/render/...` suites green
- **Committed in:** ebb4ac9

---

**Total deviations:** 1 auto-fixed (1 bug)
**Impact on plan:** The fix was necessary for COMPAT-01 (the phase's own regression contract). No scope creep; behavior for labeled edges unchanged.

## Issues Encountered
- REF-05 canonicalDOT regression caught in-suite (not by a human): `SafeSetHTML("")` drops the attribute, unlike `SafeSet("")` which emits `label=""`. Resolved by routing empty labels through `SetLabel`. Documented in the deviation above.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Both v1.11 fixes delivered: LABEL-01 (edge-label rectangles) + LABEL-02 (word-boundary-only wrapping, plan 01); COMPAT-01 green across all goldens
- Phase 34 complete — ready for phase verification and close-out

---
*Phase: 34-label-formatting-fixes-0-1-plans*
*Completed: 2026-08-10*
