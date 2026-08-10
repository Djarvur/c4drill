# Requirements: v1.11 Label Formatting Fixes

**Status:** Active
**Milestone:** v1.11 Label Formatting Fixes
**Last updated:** 2026-08-10

C4Drill's generated diagram labels must render with proper word wrapping and aspect-ratio sizing. Two user-observed defects in the render package: edge labels are emitted as plain newline-joined strings with no wrapping and no aspect-ratio sizing, and the label wrap procedure splits long words mid-word. Both fixes are additive-safe: unit label output must remain byte-identical where no long words are involved, and existing diagrams must not regress (goldens enforced via canonicalDOT, DI-1).

Source: captured todos (2026-08-10), `.planning/todos/pending/`.

---

## v1.11 Requirements

### LABEL — Label rendering

- [ ] **LABEL-01**: Edge labels render as wrapped rectangles with aspect-ratio sizing — an edge's label (`[Technology]` + Description) is formatted like a unit label: an HTML label with invisible borders (`border="0"`), text wrapped to fit the configured aspect ratio (`LabelRatio`), instead of the current plain `\n`-joined string.
- [x] **LABEL-02**: Lines break at word boundaries only — the wrapping procedure never splits a word mid-word; words longer than the per-line character budget overflow the budget on their own line (the document author may reword instead of the tool force-splitting).

### COMPAT — Backward compatibility

- [x] **COMPAT-01**: Existing diagrams render without regression — unit labels are unchanged, and golden comparisons (canonicalDOT, DI-1) still pass.

---

## Out of Scope

| Feature | Reason |
|---------|--------|
| Edge label style/styling options (colors, fonts, line styles) | Not requested; edge styling already handled by link attributes |
| Manual label positioning | GraphViz auto-layout is a core design decision (see PROJECT.md) |
| Hard-wrapping *into* the aspect ratio (scaling text to force-fit) | Violates LABEL-02 — the author chooses wording, not the tool |

---

## Traceability

*Phase mapping filled by roadmapper (2026-08-10). 3/3 v1.11 requirements mapped.*

| REQ-ID | Phase | Notes |
|--------|-------|-------|
| LABEL-01 | 34 | edge labels as wrapped HTML-table rectangles (`border="0"`), `LabelRatio` sizing |
| LABEL-02 | 34 | word-boundary-only breaking; remove `splitLongWord` char-level fallback (wrap.go) |
| COMPAT-01 | 34 | unit labels byte-identical (absent over-budget words); canonicalDOT (DI-1) goldens pass |
