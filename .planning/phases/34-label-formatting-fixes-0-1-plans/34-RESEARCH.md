# Phase 34: Label formatting fixes - Research

**Researched:** 2026-08-10
**Domain:** Go / GraphViz DOT HTML-label rendering (word wrapping, aspect-ratio sizing)
**Confidence:** HIGH

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

#### Edge label layout
- **D-01:** Edge label keeps its current text structure inside the rectangle: `[Technology]` on its own row, then a line separator, then the Description wrapped below. NOT inline `[Tech] Description` — the user explicitly corrected the initial inline suggestion: "it must be [Tech] + line separator + Description".
- **D-02:** The rectangle is the same HTML table form unit labels use: `<table border="0">` (invisible borders), wrapped text via the existing `wrapAndEscape` machinery.

#### Ratio sizing rows
- **D-03:** Edge-label width calc uses a **2-row floor** (like `labelMaxCharsForPerson`): the aspect-ratio width is always computed from at least 2 content rows, so short labels still get a wide-enough rectangle; multi-line text wraps naturally within it.

#### Tech-only edges
- **D-04:** Rectangle always — any edge label renders the borderless rectangle, including technology-only edges (no description) and description-only edges (no technology).

#### Long-word rule
- **D-05:** A word longer than the per-line budget starts its own line and stays **unsplit, overflowing the width** — no character-level fallback, no safety cap, no hyphen-point splitting. The document author may reword (user's explicit stance). This removes the `splitLongWord` fallback from `wrapText`.

### Claude's Discretion
- Row font styling inside the edge rectangle (e.g., smaller font for the `[Tech]` row, matching unit-label technology rows vs uniform font) — planner's call, keep consistent with unit-label conventions.
- Exact maxChars function variant for edge labels (`labelMaxCharsNoIcon` is the closest fit — no icon column); verify row-count handling per D-03.

### Deferred Ideas (OUT OF SCOPE)
None — discussion stayed within phase scope.

</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| LABEL-01 | Edge labels render as wrapped rectangles with aspect-ratio sizing — an edge's label (`[Technology]` + Description) is formatted like a unit label: an HTML label with invisible borders (`border="0"`), text wrapped to fit the configured aspect ratio (`LabelRatio`), instead of the current plain `\n`-joined string | §Implementation Surface: `buildEdgeLabel` (labels.go:48) + `writeLabelTableStart`/`writeNameRow`/`writeTechnologyRow`/`writeDescriptionRow` + `Edge.SetLabelHTML` in the pinned fork (attribute.go:1163) + `labelMaxCharsNoIcon` (wrap.go:185) + `labelMaxCharsForPerson` 2-row floor (wrap.go:237) |
| LABEL-02 | Lines break at word boundaries only — the wrapping procedure never splits a word mid-word; words longer than the per-line character budget overflow the budget on their own line (the document author may reword instead of the tool force-splitting) | §Implementation Surface: `wrapText` over-budget branch (wrap.go:82-89) calls `splitLongWord` (wrap.go:123); D-05 removes the fallback so the over-budget word goes on its own line unsplit |
| COMPAT-01 | Existing diagrams render without regression — unit labels are unchanged, and golden comparisons (canonicalDOT, DI-1) still pass | §Golden Impact: multilevel fixture has no edge labels and no over-budget words; `canonical.Canonical` (internal/testutil/canonical, DI-1) used by cmd/c4drill/root_test.go and internal/graph/builder_test.go — expected byte-stable; `TestWrapText`/`TestEdgeLabelGeneration` re-assertions budgeted (§Test Updates) |

</phase_requirements>

## Summary

Phase 34 is a focused bug fix in the `internal/render` package with two coordinated changes. **LABEL-01** converts edge labels from the current plain `\n`-joined string (`buildEdgeLabel`, labels.go:48) to the same HTML-table rectangle unit labels use (`<table border="0" cellpadding="0" cellspacing="0">` with name/technology/description rows), sized by the configured `LabelRatio` via the existing `labelMaxChars*` machinery — the `[Technology]` row and the wrapped Description stay on separate rows per D-01, the rectangle is always emitted per D-04 (tech-only and description-only edges included), and the width calculation uses a 2-row floor per D-03. **LABEL-02** removes the `splitLongWord` character-level fallback from `wrapText` (wrap.go:82-89), so an over-budget word goes on its own line unsplit, overflowing the budget — no character splitting anywhere (D-05).

The critical API finding: the pinned go-graphviz fork (`github.com/onokonem/go-graphviz v0.0.0-20260810112110-d19e8171ebc7`) exposes `Edge.SetLabelHTML` (cgraph/attribute.go:1163-1168, `SafeSetHTML`), which converter.go already uses for nodes (`cn.SetLabelHTML` at converter.go:239). Since graphviz 13, setting an HTML label through the plain `SetLabel` round-trip loses HTML-ness (documented in converter.go:235-239), so the edge conversion MUST switch converter.go:459 from `e.SetLabel(...)` to `e.SetLabelHTML(...)`. Without this, the `<table>` markup would render as escaped text.

COMPAT-01 is enforced by the existing canonicalDOT (DI-1) goldens: the committed multilevel fixture (`cmd/c4drill/testdata/multilevel.toml`) has no edge labels and no over-budget words in unit labels, so goldens are expected byte-stable. Two test suites assert the removed/old behavior and must be re-asserted as part of this phase: `TestWrapText` cases "forced character break" and "multi-byte unicode" (wrap_internal_test.go:30-52) and `TestEdgeLabelGeneration` "Technology and Description with newline" `checkNewline` (labels_test.go:118-153).

**Primary recommendation:** Build the edge-label rectangle with a new `buildEdgeHTMLLabel`-style builder in labels.go reusing `writeLabelTableStart`/`writeTechnologyRow`/`writeDescriptionRow`/`writeLabelTableEnd` with `labelMaxCharsNoIcon` (2-row floor applied per D-03), switch converter.go:459 to `SetLabelHTML`, and delete the `splitLongWord` fallback branch from `wrapText` — all changes land in `internal/render` in one coordinated phase (STATE.md v1.11 single-phase decision).

## Implementation Surface (verified in codebase)

### Current edge-label path
- `buildEdgeLabel` — internal/render/labels.go:48-64. Plain `strings.Join(parts, "\n")`; parts = `"[Tech]"` and Description. Called from exactly ONE site: internal/render/converter.go:459 `e.SetLabel(buildEdgeLabel(edge.Label))` inside `createEdge` (converter.go:442-491).
- `graph.EdgeLabel` — internal/graph/graph.go:96-103: `{Technology, Description, Position string}`; `graph.Edge.Label *EdgeLabel` (graph.go:82).

### Unit-label HTML table machinery (reusable, D-02)
All in internal/render/labels.go:
- `writeLabelTableStart` (labels.go:220-222): `<table border="0" cellpadding="0" cellspacing="0">` — the invisible-border rectangle form.
- `writeTechnologyRow` (labels.go:234-242): `<tr align="center"><td valign="middle"><i>[` + `wrapAndEscape(tech, maxChars)` + `]</i></td></tr>` — skips empty technology. This is exactly the D-01 `[Technology]` row form.
- `writeDescriptionRow` (labels.go:244-252): `<tr align="center"><td valign="top">` + `wrapAndEscape(desc, maxChars)` + `</td></tr>` — skips empty description.
- `writeLabelTableEnd` (labels.go:224-226): `</table>`.
- `buildNoIconHTMLLabel` (labels.go:188-203): the no-icon-column table builder (name/technology/description rows) used by System/Container/Component/Box labels — the closest structural analog; edge labels differ in having no Name row (D-01: tech row first).

### maxChars machinery (D-02, D-03)
All in internal/render/wrap.go:
- `LabelRatio` (wrap.go:33-36): package var, default `defaultLabelRatio = 1.6`, set by CLI before rendering.
- `labelMaxCharsNoIcon(rowCount)` (wrap.go:185-195): `totalHeight = rowCount * pointsPerRow` (18), `totalWidth = int(totalHeight * LabelRatio)`, `chars = totalWidth / pointsPerChar` (8), min 20. No icon-column subtraction. The closest fit for edge labels (no icon column).
- `labelMaxCharsForPerson(rowCount)` (wrap.go:237-253): enforces `effectiveRows = max(rowCount, 2)` BEFORE `calculateTextWidth` — the D-03 2-row-floor precedent. NOTE: it uses `calculateTextWidth` (which subtracts `iconColumnWidth = 36`); the no-icon variant must NOT subtract an icon column, so the D-03 floor applies to `labelMaxCharsNoIcon` directly (effectiveRows = max(rowCount, 2) on line 1 of the function).
- `wrapText` (wrap.go:41-97): word-boundary wrapping, `<BR/>` line separator. Over-budget branch (wrap.go:82-89) iterates `splitLongWord(word, maxChars)` — the D-05 removal target. After removal, the over-budget word must go on its own line unsplit: i.e., `lines = flushIfPending(...)`, then write the whole word.
- `splitLongWord` (wrap.go:123-138): rune-chunking helper — delete entirely per D-05 (or leave unused; deletion is cleaner and verified by `TestWrapText` re-assertion).
- `wrapAndEscape` (wrap.go:142-156): wraps then `html.EscapeString` per `<BR/>`-split part — used by all HTML label rows.

### HTML-ness API requirement (critical finding)
- `Edge.SetLabelHTML(v string)` exists in the pinned fork: cgraph/attribute.go:1163-1168, implemented via `SafeSetHTML` (same `agsafeset_html` path as `Node.SetLabelHTML`, converter.go:235-239 comment: "Since graphviz 13, the StrdupHTML+SetLabel round-trip loses HTML-ness"). `e.SetLabel` (attribute.go:1155-1162) is plain `SafeSet`.
- Therefore converter.go:459 must become `e.SetLabelHTML(buildEdgeLabel(edge.Label))`. If it stays `SetLabel`, the `<table>` label renders as escaped literal text — the exact bug class LABEL-01 is meant to fix.

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Edge label text assembly (`[Tech]` + Description rows) | internal/render (labels.go) | — | Pure presentation-layer string building; single consumer is converter.go `createEdge` |
| Word wrapping / line breaking | internal/render (wrap.go) | — | Shared by unit labels and edge labels; `wrapText` + `wrapAndEscape` are the single wrap machinery |
| Aspect-ratio width calc | internal/render (wrap.go) | — | `labelMaxChars*` family derives per-line char budgets from `LabelRatio` |
| DOT emission of HTML labels | internal/render (converter.go) | go-graphviz fork (cgraph) | `SetLabelHTML` preserves HTML-ness since graphviz 13 |
| Golden regression enforcement | internal/testutil/canonical + cmd/c4drill/root_test.go, internal/graph/builder_test.go | — | DI-1 canonicalDOT comparisons |

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go stdlib (strings, html, unicode/utf8) | go 1.26.1 (go.mod) | Word wrap, HTML escaping, rune counting | Already used by wrap.go/labels.go — no new deps [VERIFIED: codebase] |
| github.com/onokonem/go-graphviz (pinned fork) | v0.0.0-20260810112110-d19e8171ebc7 | DOT graph construction, HTML label attributes | Pinned in go.mod; provides `Edge.SetLabelHTML` [VERIFIED: module cache cgraph/attribute.go:1163] |
| github.com/stretchr/testify | (existing) | assert/require in render tests | Existing test convention [VERIFIED: codebase test files] |

No new external packages. Phase installs nothing — Package Legitimacy Audit not applicable (see below).

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Reuse `writeTechnologyRow`/`writeDescriptionRow` | Custom inline table markup in buildEdgeLabel | Reuse guarantees D-02 "same form as unit labels" and inherits escaping via `wrapAndEscape`; custom markup risks divergence |
| `labelMaxCharsNoIcon` with 2-row floor | New bespoke `labelMaxCharsForEdge` | D-03 explicitly asks for the person-label floor pattern; a dedicated function duplicates ~8 lines and adds a parallel maintenance surface — planner may choose either; discretion area |

## Package Legitimacy Audit

> Not applicable — this phase introduces zero external package installs (Go stdlib + already-pinned go-graphviz fork + existing testify). No slopcheck run required; nothing to verify on any registry.

## Architecture Patterns

### Pattern 1: HTML-table rectangle label (unit-label form, D-02)
**What:** A label rendered as `<table border="0" cellpadding="0" cellspacing="0">` with one `<tr align="center"><td ...>` per content row; text in cells goes through `wrapAndEscape(text, maxChars)` which emits `<BR/>` line breaks; the whole table string is passed to `SetLabelHTML` (not `SetLabel`).
**When to use:** Any label that must wrap at word boundaries into a proportional rectangle (all unit labels today; edge labels from this phase).
**Example (from buildNoIconHTMLLabel, labels.go:188-203):**
```
<task action> writeLabelTableStart(&sb); writeTechnologyRow(&sb, tech, maxChars); writeDescriptionRow(&sb, desc, maxChars); writeLabelTableEnd(&sb) </task>
```
Row helper contract: `writeTechnologyRow` emits `<i>[` + escaped/wrapped tech + `]</i>`; `writeDescriptionRow` emits plain escaped/wrapped text; both skip empty input.

### Pattern 2: Ratio-derived per-line budget (D-03)
**What:** `maxChars = labelMaxCharsNoIcon(max(rowCount, 2))` where rowCount = number of content rows (tech? + desc?, min 1). The 2-row floor (labelMaxCharsForPerson precedent, wrap.go:237-253) keeps short labels wide enough to preserve the aspect ratio.
**Edge-label row count:** tech row (if Technology != "") + description row (if Description != "") — NOTE: unlike unit labels there is no Name row; the minimum content is 1 row (tech-only or desc-only per D-04) and the floor lifts it to 2.

### Pattern 3: Unsplit overflow line (D-05)
**What:** In `wrapText`, when `wordLen > maxChars`, flush the pending line, then emit the entire word on its own line (`currentLine.WriteString(word); currentLen = wordLen`), no chunking. This is exactly what the existing "word fits as start of new line" branch does for `wordLen <= maxChars` — the over-budget case just skips the `<=` guard.

### Anti-Patterns to Avoid
- **Keeping `e.SetLabel` for the HTML edge label:** renders `<table>` as escaped literal text under graphviz 13 (converter.go:235-239 documents the round-trip loss). Must use `e.SetLabelHTML`.
- **Character-level splitting in any branch of `wrapText`:** violates LABEL-02/D-05 — no `splitLongWord`, no safety cap, no hyphen-point splitting anywhere.
- **Re-asserting old behavior instead of new:** `TestWrapText` "forced character break"/"multi-byte unicode" and `TestEdgeLabelGeneration` `checkNewline` currently pin the removed behavior — updating them is part of this phase, not a side-effect to avoid (STATE.md carry-forward).
- **Calling `t.Parallel()` in render tests:** WASM engine not concurrency-safe (TESTING.md:77, 91 annotations) — new/updated tests must carry `//nolint:paralleltest // go-graphviz WASM engine has concurrency issues`.
- **Byte-exact `require.Equal` on full DOT goldens:** go-graphviz layout is byte-nondeterministic (STATE.md DI-1) — new edge-label goldens must use `canonical.Canonical`.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| HTML table label markup | New markup in buildEdgeLabel | `writeLabelTableStart`/`writeTechnologyRow`/`writeDescriptionRow`/`writeLabelTableEnd` (labels.go:220-252) | Same form as unit labels (D-02), escaping via `wrapAndEscape` already handled |
| Word wrapping | New wrapping logic | `wrapText` + `wrapAndEscape` (wrap.go:41-156) | Single wrap machinery shared by unit + edge labels; LABEL-02 semantics change here once |
| Rune-aware char budget | Manual byte math | `utf8.RuneCountInString` (already in wrap.go) | Multi-byte safety (TestWrapText unicode case) |
| HTML-ness of DOT label attr | Manual `label=<...>` string tricks | `Edge.SetLabelHTML` (pinned fork attribute.go:1163) | `agsafeset_html` path; plain SetLabel loses HTML-ness since graphviz 13 |

**Key insight:** the codebase already contains every building block this phase needs — the phase is a *rewiring* (route edge labels through existing table row helpers + SetLabelHTML) plus one *deletion* (splitLongWord fallback), not new machinery.

## Common Pitfalls

### Pitfall 1: HTML label rendered as escaped text
**What goes wrong:** Edge label shows literal `<table border="0">...` text in the diagram.
**Why it happens:** `e.SetLabel` stores the string via `SafeSet` (plain); graphviz 13 keys the string dict on is_html and the round-trip loses HTML-ness (converter.go:235-239 documents the identical node-side issue).
**How to avoid:** Use `e.SetLabelHTML` in createEdge (converter.go:459).
**Warning signs:** Render output contains `&lt;table` or escaped markup; unit labels (which use SetLabelHTML) look right but edges don't.

### Pitfall 2: D-01 row structure silently lost
**What goes wrong:** Edge label becomes a single wrapped paragraph or inline `[Tech] Description` instead of tech row + separator + description rows.
**Why it happens:** Reusing `buildNoIconHTMLLabel` directly would add a Name row the edge label doesn't have; inlining the markup invites structure drift.
**How to avoid:** Build the edge table explicitly: table start → technology row (if any) → description row (if any) → table end; D-04 rectangle-always for the tech-only/desc-only cases.
**Warning signs:** TestEdgeLabelGeneration contains-only assertions pass while the row structure regresses — keep row-structure assertions (`<tr`, `<td`, `<i>[...]`) in the re-asserted tests.

### Pitfall 3: Over-budget word still split
**What goes wrong:** LABEL-02 not actually delivered — words still chunked.
**Why it happens:** Removing `splitLongWord` but leaving a loop/branch structure that implies chunking; or the function is deleted but TestWrapText still asserts `abcde<BR/>fghij`.
**How to avoid:** Replace the over-budget branch with the flush-and-emit-whole-word path; re-assert TestWrapText cases to the overflow-on-own-line expectation.
**Warning signs:** `splitLongWord` still referenced anywhere (`grep splitLongWord` hits), or TestWrapText asserts chunked output.

### Pitfall 4: Golden regression from unit-label output changes
**What goes wrong:** COMPAT-01 fails — unit label bytes change for in-budget text.
**Why it happens:** Refactoring `wrapText` beyond the over-budget branch (e.g., changing the fits-on-line math or the `<BR/>` separator).
**How to avoid:** Only touch the `wordLen > maxChars` branch (wrap.go:82-89) and delete `splitLongWord`; run the full canonicalDOT suite (cmd/c4drill + internal/graph) to confirm byte stability.
**Warning signs:** Any canonicalDOT golden diff in the execute-phase verification for reasons other than the LABEL-02 over-budget case or LABEL-01 edge-label form.

## Code Examples

### Edge label call site today (converter.go:458-461)
```go
if edge.Label != nil {
    e.SetLabel(buildEdgeLabel(edge.Label))
    e.SetFontSize(fontSizeEdge)
}
```
→ Change `e.SetLabel` to `e.SetLabelHTML` (keep `SetFontSize`).

### Table row helpers contract (labels.go:220-252)
```go
func writeLabelTableStart(sb *strings.Builder) { sb.WriteString(`<table border="0" cellpadding="0" cellspacing="0">`) }
func writeLabelTableEnd(sb *strings.Builder)   { sb.WriteString(`</table>`) }
func writeTechnologyRow(sb *strings.Builder, technology string, maxChars int) // skips empty; <i>[wrapAndEscape(tech)]</i>
func writeDescriptionRow(sb *strings.Builder, description string, maxChars int) // skips empty; wrapAndEscape(desc)
```

### 2-row floor precedent (wrap.go:237-244, labelMaxCharsForPerson)
```go
effectiveRows := rowCount
if effectiveRows < 2 { effectiveRows = 2 }
```
Apply the same floor to the no-icon budget used for edge labels (labelMaxCharsNoIcon, wrap.go:185-195 — no icon-column subtraction).

### Edge.SetLabelHTML signature (pinned fork, cgraph/attribute.go:1163-1168)
```go
func (e *Edge) SetLabelHTML(v string) *Edge { e.SafeSetHTML(string(labelAttr), v, "\\E"); return e }
```

## Golden Impact (COMPAT-01)

- Committed fixture `cmd/c4drill/testdata/multilevel.toml` (+ `multilevel.expanded.dot`, `expanded.dot` goldens) contains NO edge labels and no over-budget words in unit labels → canonicalDOT goldens in cmd/c4drill/root_test.go (TestBuildExpandedGraphBaselineDOT, DI-1, root_test.go:706,1269-1331) and internal/graph/builder_test.go are expected byte-stable after both changes [VERIFIED: fixture scan].
- The ONLY permitted output deltas per STATE.md: (1) LABEL-02 over-budget-word case (unit AND edge labels), (2) LABEL-01 edge-label form (plain `\n` string → HTML table).
- New edge-label golden assertions, if any, MUST use `canonical.Canonical` (DI-1), never byte-exact `require.Equal`.

## Test Updates Required (budgeted in planning)

- `internal/render/wrap_internal_test.go` `TestWrapText` (lines 30-52):
  - "forced character break" `abcdefghij`/5: expected `abcde<BR/>fghij` → re-assert overflow-on-own-line: `abcdefghij` (whole word on one line, unsplit).
  - "multi-byte unicode" `日本語テスト文字列`/4: expected `日本語テ<BR/>スト文字<BR/>列` → re-assert `日本語テスト文字列` unsplit on one line.
- `internal/render/labels_test.go` `TestEdgeLabelGeneration` (lines 88-156):
  - "Technology and Description with newline" (`checkNewline: true`, asserts `\n`) → re-assert the HTML table form: contains `<table border="0"`, `<BR/>` separator inside cells, `[gRPC]` in `<i>`, "Protocol buffers" text; drop/repurpose `checkNewline`.
  - "Technology only" and "Description only" cases → D-04 rectangle-always: assert `<table border="0"` presence for these too.
- Render tests must not call `t.Parallel()` (`//nolint:paralleltest // go-graphviz WASM engine has concurrency issues`).

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Edge label = plain `\n`-joined string via `e.SetLabel` | HTML-table rectangle via `e.SetLabelHTML` + row helpers | This phase (v1.11) | LABEL-01 |
| Over-budget words force-split by `splitLongWord` | Over-budget word unsplit on its own line (author rewords) | This phase (v1.11) | LABEL-02, unit labels too |

**Deprecated/outdated:**
- `splitLongWord` (wrap.go:123-138): character-level fallback — removed per D-05.

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | The multilevel fixture has no edge labels and no over-budget words (fixture scan confirms no `label=` attrs with tech/desc on edges) | Golden Impact | If wrong, goldens legitimately change and COMPAT-01 re-baselining would be needed — CONTEXT.md Notes make the same claim; verify at execute time via the canonicalDOT suite run |

All other claims verified directly against the codebase and the pinned go-graphviz fork (module cache).

## Open Questions (RESOLVED)

1. **Does the pinned fork expose `Edge.SetLabelHTML`?** — RESOLVED: yes, cgraph/attribute.go:1163-1168 (`SafeSetHTML`), verified in module cache at github.com/onokonem/go-graphviz@v0.0.0-20260810112110-d19e8171ebc7.
2. **Which maxChars variant for edge labels (D-03 discretion)?** — RESOLVED (recommendation): `labelMaxCharsNoIcon` (no icon column; wrap.go:185) with the 2-row floor from `labelMaxCharsForPerson` applied to its rowCount input. `labelMaxCharsForPerson` itself is not used because it subtracts `iconColumnWidth` (36pt) via `calculateTextWidth` — wrong for an icon-less edge label. Planner may still choose a dedicated wrapper function; discretion area.
3. **Does removing `splitLongWord` change unit-label output?** — RESOLVED: only when a word exceeds maxChars (the intended LABEL-02 semantic, STATE.md carry-forward); in-budget unit labels are byte-identical.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go toolchain | build + tests | ✓ | go1.26.5 darwin/arm64 (go.mod: go 1.26.1) | — |
| go-graphviz pinned fork | DOT rendering | ✓ | v0.0.0-20260810112110-d19e8171ebc7 (go.mod) | — |
| testify | tests | ✓ | (existing) | — |

**Missing dependencies with no fallback:** none. Phase is code/test-only — Step 2.6 full audit skipped (no external services, CLIs, or runtimes beyond the Go toolchain).

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing (built-in) + testify assert/require |
| Config file | none — Go conventions (`*_test.go` beside source) |
| Quick run command | `go test ./internal/render/ -run 'TestWrapText|TestEdgeLabelGeneration'` |
| Full suite command | `go test ./...` (canonicalDOT goldens live in `./cmd/c4drill` and `./internal/graph`) |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| LABEL-01 | Edge label renders as HTML table with `border="0"`, `[Tech]` row + `<BR/>`-wrapped Description | unit (render) | `go test ./internal/render/ -run TestEdgeLabelGeneration -v` | ✅ (re-asserted in this phase) |
| LABEL-01 | Edge label maxChars derives from `LabelRatio` via 2-row floor (D-03) | unit (render, internal) | `go test ./internal/render/ -run 'TestLabelMaxChars|TestBuildEdgeLabel'` (new internal tests) | ❌ Wave 0 — add internal assertions |
| LABEL-02 | Over-budget word stays unsplit on its own line | unit (render, internal) | `go test ./internal/render/ -run TestWrapText` | ✅ (re-asserted in this phase) |
| LABEL-02 | `splitLongWord` gone — no char-level splitting anywhere | static | `grep -rn "splitLongWord" internal/` → 0 hits | ✅ (command check) |
| COMPAT-01 | canonicalDOT goldens pass unchanged (DI-1) | integration/golden | `go test ./cmd/c4drill/... ./internal/graph/...` | ✅ (existing) |
| COMPAT-01 | In-budget unit labels byte-identical | unit | `go test ./internal/render/` (existing label tests unchanged) | ✅ (existing) |

### Sampling Rate
- **Per task commit:** `go test ./internal/render/ -run '<affected tests>'`
- **Per wave merge:** `go test ./internal/render/... ./internal/graph/... ./cmd/c4drill/...`
- **Phase gate:** `go test ./...` green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `internal/render/wrap_internal_test.go` — re-assert TestWrapText "forced character break" + "multi-byte unicode" to unsplit-overflow (this phase's RED step, TDD)
- [ ] `internal/render/labels_test.go` — re-assert TestEdgeLabelGeneration for HTML table form + D-04 rectangle-always cases (this phase's RED step, TDD)
- [ ] No framework install — Go testing built-in

*(No infrastructure gaps — existing test infrastructure covers all phase requirements)*

## Security Domain

> Required when `security_enforcement` is enabled (absent = enabled in .planning/config.json).

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | no | — |
| V3 Session Management | no | — |
| V4 Access Control | no | — |
| V5 Input Validation | yes | `wrapAndEscape` → `html.EscapeString` per line (existing, wrap.go:142-156) — user-authored label text is HTML-escaped before embedding in `label=<...>` |
| V6 Cryptography | no | — |

### Known Threat Patterns for Go CLI rendering

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| HTML injection via user-authored label text (Technology/Description in `<table>` label) | Tampering | `wrapAndEscape` escapes `&<>"'` via `html.EscapeString` before emission (wrap.go:150-155); keep escaping in the new edge-label path — reuse the row helpers so escaping is inherited |
| DOT attribute breakout via `];`/`{`/`}` in label text | Tampering | Existing: user text goes through HTML-table label form (canonical parser is HTML-aware); no new input surface in this phase — labels already flow through the same `wrapAndEscape` machinery unit labels use |

No new attack surface: this phase changes the *format* of edge label emission (HTML table vs plain string) and removes a fallback branch; both changes route through the existing escaping machinery. Threat model for the plans: accept residual risk on label styling (cosmetic), keep escaping mitigation for V5.

## Sources

### Primary (HIGH confidence)
- Codebase: internal/render/labels.go, internal/render/wrap.go, internal/render/converter.go, internal/render/wrap_internal_test.go, internal/render/labels_test.go, internal/graph/graph.go, internal/testutil/canonical/canonical.go, cmd/c4drill/root_test.go, .planning/STATE.md, .planning/REQUIREMENTS.md, .planning/codebase/TESTING.md — all read directly
- Pinned go-graphviz fork (module cache): cgraph/attribute.go `Edge.SetLabelHTML`/`Edge.SetLabel` — read directly at github.com/onokonem/go-graphviz@v0.0.0-20260810112110-d19e8171ebc7

### Secondary (MEDIUM confidence)
- 34-CONTEXT.md decisions D-01..D-05 + canonical refs (user decisions, gathered 2026-08-10)

### Tertiary (LOW confidence)
- None

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - no new dependencies; everything verified in go.mod/module cache/codebase
- Architecture: HIGH - implementation surface fully mapped (function-level) from direct reads
- Pitfalls: HIGH - includes the SetLabelHTML graphviz-13 trap with in-repo documentation (converter.go:235-239)

**Research date:** 2026-08-10
**Valid until:** 2026-09-09 (stable — no fast-moving dependencies)
