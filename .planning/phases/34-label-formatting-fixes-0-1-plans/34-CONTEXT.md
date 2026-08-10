# Phase 34: Label formatting fixes - Context

**Gathered:** 2026-08-10
**Status:** Ready for planning

<domain>
## Phase Boundary

Generated diagram labels render with proper word wrapping and aspect-ratio sizing:
1. **LABEL-01** — Edge labels render as wrapped rectangles: the unit-label HTML table form (`border="0"`, invisible borders) with width derived from the configured aspect ratio (`LabelRatio`).
2. **LABEL-02** — Lines break at word boundaries only; no mid-word splits anywhere (unit labels and edge labels).
3. **COMPAT-01** — No regression: unit labels byte-identical absent over-budget words; canonicalDOT goldens pass unchanged.

</domain>

<decisions>
## Implementation Decisions

### Edge label layout
- **D-01:** Edge label keeps its current text structure inside the rectangle: `[Technology]` on its own row, then a line separator, then the Description wrapped below. NOT inline `[Tech] Description` — the user explicitly corrected the initial inline suggestion: "it must be [Tech] + line separator + Description".
- **D-02:** The rectangle is the same HTML table form unit labels use: `<table border="0">` (invisible borders), wrapped text via the existing `wrapAndEscape` machinery.

### Ratio sizing rows
- **D-03:** Edge-label width calc uses a **2-row floor** (like `labelMaxCharsForPerson`): the aspect-ratio width is always computed from at least 2 content rows, so short labels still get a wide-enough rectangle; multi-line text wraps naturally within it.

### Tech-only edges
- **D-04:** Rectangle always — any edge label renders the borderless rectangle, including technology-only edges (no description) and description-only edges (no technology).

### Long-word rule
- **D-05:** A word longer than the per-line budget starts its own line and stays **unsplit, overflowing the width** — no character-level fallback, no safety cap, no hyphen-point splitting. The document author may reword (user's explicit stance). This removes the `splitLongWord` fallback from `wrapText`.

### Claude's Discretion
- Row font styling inside the edge rectangle (e.g., smaller font for the `[Tech]` row, matching unit-label technology rows vs uniform font) — planner's call, keep consistent with unit-label conventions.
- Exact maxChars function variant for edge labels (`labelMaxCharsNoIcon` is the closest fit — no icon column); verify row-count handling per D-03.

### Folded Todos
- **Wrap edge labels like unit labels in generated TOML** (2026-08-10) — original problem: edge labels in generated output contain no line breaks; must be formatted like unit labels (rectangle with specified aspect ratio, invisible borders). This is LABEL-01, the core of this phase.
- **Wrap labels at word boundaries only** (2026-08-10) — original problem: long words get split mid-word by the alignment procedure; line division must happen at word boundaries; if the unit shape is unacceptable the author uses a different word. This is LABEL-02.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Source of truth (this phase)
- `internal/render/labels.go` — `buildEdgeLabel` (line ~48, plain `\n`-joined string), `buildHTMLLabelForType` dispatch, `buildPersonHTMLLabel`/`buildNoIconHTMLLabel` unit-label table builders, `writeLabelTableStart`/`writeNameRow`/`writeTechnologyRow`/`writeDescriptionRow` row helpers
- `internal/render/wrap.go` — `wrapText` (word-boundary wrap with `splitLongWord` fallback at line ~123 to be removed per D-05), `wrapAndEscape`, `labelMaxCharsNoIcon` (line ~185), `labelMaxCharsForPerson` (2-row floor precedent, line ~237), `LabelRatio` var (default 1.6)
- `internal/render/converter.go` — `createEdge` single call site (`e.SetLabel(buildEdgeLabel(edge.Label))`, line ~459)
- `internal/graph/graph.go` — `EdgeLabel` struct (Technology, Description, Position), `Edge.Label`
- `internal/testutil/canonical` — `canonical.Canonical` (DI-1 order-insensitive golden helper; COMPAT-01 enforcement)

### Test conventions
- `internal/render/wrap_test.go` — `TestWrapText` cases "forced character break" and "multi-byte unicode" assert the removed `splitLongWord` behavior (must be re-asserted per D-05)
- `internal/render/labels_test.go` — `TestEdgeLabelGeneration` "Technology and Description with newline" asserts plain `\n` via `checkNewline` (must be re-asserted for the HTML form)
- `.planning/codebase/TESTING.md` — golden/canonicalDOT conventions; render tests never `t.Parallel()` (WASM engine)

No external specs — requirements fully captured in decisions above.

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- Unit-label HTML table builders (`buildPersonHTMLLabel`, `buildNoIconHTMLLabel`, row helpers): produce `<table border="0">` labels — the exact form edge labels must reuse (D-02)
- `wrapAndEscape` + `labelMaxChars*` functions: word-wrap with ratio-derived per-line budgets
- `labelMaxCharsForPerson`'s 2-row floor pattern: precedent for D-03
- `canonical.Canonical` (from Phase 33): order-insensitive golden comparisons for COMPAT-01

### Established Patterns
- Edge labels are plain string labels via `SetLabel`; unit labels are HTML labels via `SetLabelHTML`
- WASM render engine is not concurrency-safe — render tests must not call `t.Parallel()` (91 annotations)
- Integration-by-construction testing: real models → real output assertions, no mocks

### Integration Points
- `internal/render/converter.go:459` (`createEdge`) — switch edge labels from `SetLabel` (plain) to the HTML form
- `internal/render/labels.go` `buildEdgeLabel` — replace the string join with the rectangle builder
- `internal/render/wrap.go` `wrapText` — remove the `splitLongWord` fallback branch (affects unit labels too — intended per LABEL-02)

</code_context>

<specifics>
## Specific Ideas

- User's exact words on layout: edge labels "должны быть отформатированы так же, как и unit labels: в прямоугольники с указанным соотношением сторон (границы такого прямоугольника должны быть невидимы)" — formatted like unit labels: rectangles with the specified aspect ratio, borders invisible.
- User's exact words on long words: "деление на строки происходило по границам слов. если форма юнита окажется неприемлемой - автор документа всегда может использовать другое слово" — line division at word boundaries; if the unit shape is unacceptable the document author can use a different word.

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 34-Label formatting fixes*
*Context gathered: 2026-08-10*
