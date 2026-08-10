# Phase 34 Plan 04: self-consistent aspect-ratio sizing Summary

**Label rectangles now size their width from the actual text: a closed-form quadratic (width² − B·width − C = 0) replaces the fixed content-row-count model, so the rectangle's ratio approximates LabelRatio instead of collapsing to a tall narrow strip for long labels**

## Performance

- **Duration:** ~40 min
- **Started:** 2026-08-10T18:05:00Z
- **Completed:** 2026-08-10T18:45:00Z
- **Tasks:** 3 (TDD: RED/GREEN/REFACTOR-verify + 1 follow-up fix)
- **Files modified:** 4 (+1 golden re-baselined, 6 example SVGs regenerated)

## Root Cause Closed (UAT test 1, minor)

`buildEdgeLabel` sized the width from a fixed row count (`labelMaxCharsNoIcon(max(rowCount, 2))`), but the wrapped text's line count drives the rendered height — a 196-char label at the 2-row floor produced a rectangle ≈ 0.2:1 instead of the configured LabelRatio (1.6). User: "the label text rectangle ratio is pretty far from the desired". User scoped: **edge + unit labels**.

## Implementation

- **`labelMaxCharsForText(fixedRows, textLen, ratio)`** (wrap.go): closed-form positive root of width² − B·width − C = 0 with B = fixedRows·pointsPerRow·ratio, C = textLen·pointsPerChar·pointsPerRow·ratio. Deliberately NOT an iterative line-count loop — that oscillates (ceil sawtooth: 16 lines → 57 chars → 2 lines → 7 chars → …). Floor 6 for tiny labels.
- **All label builders rewired** (labels.go): edge (fixedRows=0, textLen = tech + desc), no-icon units (name + tech + desc), queue (fixedRows=1 for the ASCII graphic row), person (icon column subtracted). The `minCharsPerLine` (20) floor at the call sites prevents over-wrapping of short labels.
- **Edge font correction** (follow-up fix): edge labels render at `fontSizeEdge` 10pt, not the 12pt the pointsPerRow=18 model assumes — without scaling, the measured rectangle was 2.4:1 (too wide). `edgeRatio = LabelRatio × fontSizeEdge / pointsPerRow` aligns the modeled height with the rendered height → 1.13:1.
- The shape ratio multipliers (cylinder ×2.2, queue ×1.6) preserved.

## Task Commits (TDD)

1. **RED** `8937756` — `test(34-04): add labelMaxCharsForText ratio-sizing cases` (6 cases: 21/15/7/6 + monotonicity)
2. **GREEN** `ee2efe6` — `feat(34-04): self-consistent aspect-ratio sizing for label rectangles`
3. **REFACTOR** `422f828` — `test(34-04): re-baseline expanded DOT golden for ratio sizing (COMPAT-02)`
4. **FIX** `3ac58a0` — `fix(34-04): scale edge-label ratio by fontSizeEdge/pointsPerRow for rendered row height`
5. **DOCS** `729c692` + `4ac2a7b` — regenerated example SVGs

## COMPAT-01 / Golden Impact

`multilevel.expanded.dot` re-baselined for the intended re-wraps only: cylinder/unit labels' name and description line budgets changed (e.g. "Dynamic User Storage" description 5→4 lines; "Policy and Lock Storage" name wraps at the new 20-char budget). All other deltas are layout geometry (bb/pos/lp), stripped by canonicalDOT. Full suite: 12/12 packages green.

## Measured Result (overflow-test edge label, SVG)

| State | Rows | Width | Height | Ratio |
|-------|------|-------|--------|-------|
| Before 34-04 (2-row floor) | ~11 | ~217pt | ~256pt | ≈ 0.85 |
| After 34-04 (quadratic, no font fix) | 9 | ~217pt | ~90pt | ≈ 2.4 |
| **Final (font-scaled ratio + floor 20)** | **14** | **~164pt** | **~146pt** | **≈ 1.13** |

Remaining gap to 1.6 is content-driven: underscore-token descriptions ("IMAGE_NATIVE_", "PROCESSED_CUDA_") pack into short rows that no single chars-per-line can fill evenly. Short labels (1-2 rows) are inherently wide-thin and unchanged in spirit.

## Decisions

- Closed-form quadratic over iterative line-count loop (oscillation)
- fixedRows counts only non-wrapping rows (edge: 0; queue: 1 graphic)
- Edge ratio scaled by fontSizeEdge/pointsPerRow (10/18) so the model matches rendered 10pt rows
- minCharsPerLine (20) floor retained at call sites; person floor 10 (old precedent)

## Requirements

- LABEL-01 (re-asserted): rectangles approximate the configured aspect ratio ✓
- COMPAT-01: goldens green after documented re-baseline; full suite 12/12 ✓
