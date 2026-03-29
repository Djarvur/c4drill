---
phase: 17-units-labels-cells-must-be-word-wrapped-to-make-the-unit-shape-proportions-as-close-as-possible-to-credit-card-proportions
plan: 01
status: complete
---

# Phase 17 Plan 01 Summary

## What Was Built

Word-wrapping system for HTML label cells in C4 diagram units, achieving configurable width:height ratios (default 1.6:1, credit card proportions).

## Files Created

- `internal/render/wrap.go` — Word-wrap engine with `wrapText()`, `wrapAndEscape()`, `estimateCharsFromWidth()`, `calculateTextWidth()`, `labelMaxChars()`, exported `LabelRatio` global
- `internal/render/wrap_test.go` — 14 unit tests covering word-boundary wrapping, character-fallback, unicode, and dimension calculations

## Files Modified

- `internal/render/labels.go` — All 6 HTML label builders (`buildPersonHTMLLabel`, `buildDbHTMLLabel`, `buildQueueHTMLLabel`, `buildSystemHTMLLabel`, `buildContainerHTMLLabel`, `buildComponentHTMLLabel`) now use `wrapAndEscape()` instead of `html.EscapeString()`, with dynamic `maxChars` based on row count and ratio
- `cmd/c4drill/root.go` — Added `--label-ratio` CLI flag, `C4DRILL_LABEL_RATIO` env var support, `getLabelRatio()` helper, sets `render.LabelRatio` before rendering

## Key Decisions

- Used package-level `render.LabelRatio` global (set by CLI) instead of threading ratio through 14+ function signatures — pragmatic for CLI tool architecture
- Added `minCharsPerLine = 20` constant to prevent excessive wrapping of short text
- `wrapAndEscape()` splits by `<BR/>`, escapes each segment individually, then rejoins — ensures HTML entities don't break line boundaries
- `wrapText()` uses `strings.Fields()` for word splitting and `[]rune` for Unicode-safe character counting

## Patterns Established

- Word-wrapping via `<BR/>` tags in GraphViz HTML labels (Graphviz `<td>` doesn't natively word-wrap)
- Dynamic width calculation: `textWidth = (rowCount × pointsPerRow × ratio) - iconColumnWidth`
- CLI flag with env var fallback pattern: flag > env > default

## Verification

- All 14 wrap tests pass
- All existing HTML label tests pass (backward compatible)
- `--label-ratio` flag visible in `--help` output
- Visual verification approved by user
