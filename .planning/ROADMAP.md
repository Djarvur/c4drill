# Roadmap: C4Drill

## Milestones

- ✅ **v1.0 Initial Release** — Phases 1-6 (shipped 2026-03-10)
- ✅ **v1.7 Queue Label Fix & Visual Improvements** — Phases 7-18 (shipped 2026-03-29)
- ✅ **v1.8 Proper C1/C2/C3 View Generation** — Phases 19-26 (shipped 2026-08-06) → [archive](milestones/v1.8-ROADMAP.md)
- ✅ **v1.9 C3 Boundary Node Fix** — Phase 27 (shipped 2026-08-06) → [archive](milestones/v1.9-ROADMAP.md)
- ✅ **v1.10 Model Composition** — Phases 28-33 (shipped 2026-08-08) → [archive](milestones/v1.10-ROADMAP.md)
- 🚧 **v1.11 Label Formatting Fixes** — Phase 34 (in progress)

## Phases

<details>
<summary>✅ v1.9 C3 Boundary Node Fix (Phase 27) — SHIPPED 2026-08-06</summary>

- [x] Phase 27: C3 Boundary Node Fix (1/1 plan) — completed 2026-08-06

Full details: [milestones/v1.9-ROADMAP.md](milestones/v1.9-ROADMAP.md)

</details>

<details>
<summary>✅ v1.10 Model Composition (Phases 28-33) — SHIPPED 2026-08-08</summary>

**Goal:** Expand C4Drill's authoring model from a single static TOML file into a composable, parametrized, multi-file format. Four additive features form a strict runtime pipeline: `include → template-expand → relative-peer-resolve → humanize → validate → generate-views → render`.

- [x] Phase 28: Reference field (📖) (1/1 plan) — completed 2026-08-08
- [x] Phase 29: Optional name humanization (2/2 plans) — completed 2026-08-08
- [x] Phase 30: Relative-peer resolution (2/2 plans) — completed 2026-08-08
- [x] Phase 31: Template expansion (2/2 plans) — completed 2026-08-08
- [x] Phase 32: Include directive (multi-file) (2/2 plans) — completed 2026-08-08
- [x] Phase 33: Docs sweep + end-to-end goldens (4/4 plans) — completed 2026-08-08

**Stats:** 6 phases, 13 plans, 35 tasks, 119 files changed (+17,703/−626). All 39 requirements validated.

Full details: [milestones/v1.10-ROADMAP.md](milestones/v1.10-ROADMAP.md)

</details>

<details>
<summary>🚧 v1.11 Label Formatting Fixes (Phase 34) — IN PROGRESS</summary>

**Milestone Goal:** Generated diagram labels render with proper word wrapping and aspect-ratio sizing — edge labels formatted like unit labels (wrapped rectangle with `LabelRatio` aspect ratio, invisible borders), and line breaks at word boundaries only (no mid-word splits).

- [ ] Phase 34: Label formatting fixes (0/1 plans) — not started

**Granularity:** standard (1 phase — small bug-fix milestone; both fixes live in `internal/render` and share the wrap machinery, so they must land coordinated)
**Coverage:** 3/3 v1.11 requirements mapped ✓

### Phase 34: Label formatting fixes

**Goal**: Users see properly formatted labels: edge labels render as wrapped rectangles with invisible borders and aspect-ratio sizing (the same HTML table form unit labels use), and no label text is ever split mid-word.
**Depends on**: Phase 33 (canonicalDOT helper `internal/testutil/canonical` enforces the COMPAT goldens)
**Requirements**: LABEL-01, LABEL-02, COMPAT-01
**Success Criteria** (what must be TRUE):

  1. An edge with `[Technology]` and/or Description renders its label as an HTML table label with `border="0"` (invisible borders) — the same table-label form unit labels use — instead of the current plain `\n`-joined string (LABEL-01)
  2. Edge label text wraps to fit the configured aspect ratio (`LabelRatio`) — per-line width derived from the same maxChars machinery as unit labels (`labelMaxChars*` + `wrapAndEscape`), so the label's width:height stays proportional (LABEL-01)
  3. Lines break at word boundaries only — a word longer than the per-line character budget appears unsplit on its own line (overflowing the budget), never character-split; the author may reword instead (LABEL-02)
  4. No regression: unit labels render byte-identically to v1.10 wherever no word exceeds the per-line budget, and all existing goldens (canonicalDOT, DI-1) pass unchanged — the only permitted output changes are the LABEL-02 long-word case and the LABEL-01 edge-label form (COMPAT-01)

**Notes**:

- **LABEL-01 implementation surface:** `buildEdgeLabel` (internal/render/labels.go:48) currently emits a plain string `[Technology]\nDescription` and is called from the single site converter.go:459 (`e.SetLabel(buildEdgeLabel(edge.Label))`). Reuse the unit-label HTML table machinery: `<table border="0">` start/end + name/technology/description rows (`writeLabelTableStart`/`writeNameRow`/`writeTechnologyRow`/`writeDescriptionRow`), with maxChars derived from `LabelRatio` via the `labelMaxChars*` functions (edge labels have no icon column — `labelMaxCharsNoIcon` is the closest fit; verify row-count handling).
- **LABEL-02 implementation surface:** `wrapText` (internal/render/wrap.go:41) falls back to character-level breaking via `splitLongWord` (wrap.go:123) for words exceeding maxChars. Remove that fallback branch — an over-budget word goes on its own line unsplit. **`splitLongWord` removal also changes UNIT label output** for the pathological over-budget case, which is the intended LABEL-02 semantic, not a regression.
- **Test updates required (plan-phase must budget for them):** `TestWrapText` cases "forced character break" (`abcdefghij`/5 → `abcde<BR/>fghij`) and "multi-byte unicode" (`日本語テスト文字列`/4) assert the removed splitLongWord behavior — re-assert to the overflow-on-own-line behavior. `TestEdgeLabelGeneration` "Technology and Description with newline" uses `checkNewline` asserting a plain `\n` separator — the HTML table form emits `<BR/>` inside `<td>` cells instead; re-assert the HTML form.
- **Golden impact:** the committed multilevel fixture (`cmd/c4drill/testdata/multilevel.toml`) has NO edge labels (links carry no technology/description) and no over-budget words in unit labels, so the COMPAT-02/REF-05 canonicalDOT goldens should stay byte-stable — run the full canonicalDOT suite to confirm. Any new edge-label golden must use `canonical.Canonical` (DI-1), never byte-exact `require.Equal`.

**Plans**: TBD
**UI hint**: yes

</details>

## Progress

| Phase | Milestone | Plans Complete | Status | Completed |
|-------|-----------|----------------|--------|-----------|
| 28. Reference field | v1.10 | 1/1 | Complete | 2026-08-08 |
| 29. Optional name humanization | v1.10 | 2/2 | Complete | 2026-08-08 |
| 30. Relative-peer resolution | v1.10 | 2/2 | Complete | 2026-08-08 |
| 31. Template expansion | v1.10 | 2/2 | Complete | 2026-08-08 |
| 32. Include directive | v1.10 | 2/2 | Complete | 2026-08-08 |
| 33. Docs sweep + goldens | v1.10 | 4/4 | Complete | 2026-08-08 |
| 34. Label formatting fixes | v1.11 | 0/1 | Not started | - |
