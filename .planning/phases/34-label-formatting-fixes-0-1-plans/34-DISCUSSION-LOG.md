# Phase 34: Label formatting fixes - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-08-10
**Phase:** 34-Label formatting fixes
**Areas discussed:** Edge label layout, Ratio sizing rows, Tech-only edges, Long-word rule

---

## Edge label layout

| Option | Description | Selected |
|--------|-------------|----------|
| Inline [Tech] + desc (Recommended) | Single cell, one wrapped text: `[HTTP] Fetches user data...` — classic C4 convention | |
| Separate tech row | Tech on its own row in smaller font (like unit labels' technology row), description wrapped below | ✓ |
| Side-by-side columns | Two-column: technology left, description right in the same row | |

**User's choice:** Initially selected "Inline [Tech] + desc", then **corrected in free text**: "not about sizing calc, but about edge labels: it must be [Tech] + line separator + Description".
**Notes:** The user's correction overrides the initial selection — the edge label preserves its current `[Technology]\nDescription` structure inside the rectangle. Confirmed via follow-up ("Yes, two rows").

## Ratio sizing rows

| Option | Description | Selected |
|--------|-------------|----------|
| 2-row floor (Recommended) | Width always computed from at least 2 content rows (like person labels) — short labels still get a wide-enough rectangle | ✓ |
| Actual row count | Width from actual content rows — closer to text, but single-row labels get a narrower rectangle | |
| Fixed budget, no ratio | Fixed chars-per-line regardless of ratio — simplest, but not ratio-driven | |

**User's choice:** 2-row floor (Recommended)
**Notes:** The user's free-text correction about layout arrived while this area was presented; sizing question itself was answered with the recommendation.

## Tech-only edges

| Option | Description | Selected |
|--------|-------------|----------|
| Rectangle always (Recommended) | Any edge label (tech-only, desc-only, or both) renders the borderless rectangle | ✓ |
| Only with description | Edges without a description keep plain text | |

**User's choice:** Rectangle always (Recommended)
**Notes:** None.

## Long-word rule

| Option | Description | Selected |
|--------|-------------|----------|
| Unsplit overflow (Recommended) | Word longer than the budget starts its own line and stays unsplit, overflowing — no cap, no fallback; author rewording is the remedy | ✓ |
| Safety cap | Split only words longer than 2× the budget | |
| Hyphen-point splitting | Allow splitting at hyphens/punctuation only | |

**User's choice:** Unsplit overflow (Recommended)
**Notes:** Consistent with the user's original todo: "автор документа всегда может использовать другое слово" (the document author can always use a different word).

---

## Claude's Discretion

- Row font styling inside the edge rectangle (smaller font for the `[Tech]` row vs uniform) — keep consistent with unit-label conventions
- Exact maxChars function variant for edge labels (`labelMaxCharsNoIcon` closest fit — no icon column); verify row-count handling per the 2-row floor decision

## Deferred Ideas

None — discussion stayed within phase scope.
