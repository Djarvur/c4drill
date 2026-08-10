---
status: testing
phase: 34-label-formatting-fixes-0-1-plans
source: [34-01-SUMMARY.md, 34-02-SUMMARY.md]
started: 2026-08-10T17:00:00Z
updated: 2026-08-10T17:00:00Z
---

## Current Test
<!-- OVERWRITE each test - shows where we are -->

[testing complete]

## Tests

### 1. Edge label rectangle in generated output
expected: An edge with technology + description renders its label as a borderless rectangle: `[Tech]` on its own row, wrapped description below (HTML table with border="0"), instead of a plain newline-joined string.
result: issue
reported: "looks like the word wrapping is not working as I expect: any punctuation must be considered word boundary, not just spaces"
severity: major

### 2. Tech-only edge label gets the rectangle
expected: An edge with only technology (no description) still renders the borderless rectangle around `[Tech]` — same form as full labels.
result: pass

### 3. Word-boundary-only line breaking (incl. punctuation)
expected: Lines break at word boundaries AND punctuation boundaries — hyphens (`Multi-Consumer`), `->`, `:`, and other punctuation are break opportunities, not just spaces. A long space-free token like `IMAGE_NATIVE_PROCESSED_CUDA_YUV420->EXTERNAL:...` wraps at its punctuation instead of overflowing as one unsplit line. Runs of letters/digits without any separator still stay unsplit (author may reword).
result: issue
reported: "technology string wrapped, description is not"
severity: major

### 4. No regression on existing diagrams
expected: Existing models (e.g., multilevel fixture) render unchanged vs v1.10 — unit labels identical, golden comparisons pass, no visual drift in generated diagrams.
result: pass

## Summary

total: 4
passed: 2
issues: 2
pending: 0
skipped: 0

## Gaps

- truth: "An edge with technology + description renders its label as a borderless rectangle: [Tech] on its own row, wrapped description below (HTML table with border=0), instead of a plain newline-joined string."
  status: failed
  reason: "User reported: looks like the word wrapping is not working as I expect: any punctuation must be considered word boundary, not just spaces"
  severity: major
  test: 1
  root_cause: "wrapText (internal/render/wrap.go:46) tokenizes with strings.Fields, which splits on whitespace only. Space-free tokens containing punctuation (->, :, -) receive no break opportunities, so they overflow unsplit on their own line."
  artifacts: [internal/render/wrap.go]
  missing: ["punctuation-aware tokenization: split at punctuation characters in addition to whitespace"]
- truth: "Lines break at word boundaries AND punctuation boundaries; long space-free tokens wrap at their punctuation instead of overflowing unsplit."
  status: failed
  reason: "User reported: technology string wrapped, description is not"
  severity: major
  test: 3
  root_cause: "Same root cause as test 1: whitespace-only tokenization. The technology string contains spaces (wraps normally); the space-free description token never breaks."
  artifacts: [internal/render/wrap.go]
  missing: ["punctuation-aware tokenization (shared fix with test 1)"]
