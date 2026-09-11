# Phase 40: Deterministic SVG Output - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-09-03
**Phase:** 40-Deterministic SVG Output
**Areas discussed:** Root-cause framing, Ordering contract, ID format, Verification contract
**Mode:** `--auto` (yolo runtime) — all selections auto-chosen and logged

---

## Root-cause framing

| Option | Description | Selected |
|--------|-------------|----------|
| Insertion-order fix in Go | Edge append order varies (map-iterated LinksFrom mirrors); GraphViz numbers `edge<N>` by insertion order | ✓ |
| Explicit SVG ids from c4drill | Write deterministic ids ourselves into DOT/SVG | |
| Post-hoc SVG rewrite | Rewrite `edge<N>` attributes in emitted SVG text | |

**User's choice:** [auto] Q: "Where does determinism enter?" → Selected: "Insertion-order fix in Go" (recommended default — scout confirmed no `id="edge` literals in Go source; ids come from GraphViz's emitter).

---

## Ordering contract

| Option | Description | Selected |
|--------|-------------|----------|
| Fix at the source + defensive stable sort | Order LinksFrom mirrors by stable content key; end buildEdges with a stable sort by Edge.Name | ✓ |
| Final sort only | Leave mirror order unstable, sort the edge slice at the end | |
| Rebuild walk around sorted maps | Convert all link collections to sorted structures everywhere | |

**User's choice:** [auto] Q: "Where to establish deterministic order?" → Selected: "Fix at the source + defensive stable sort" (recommended default — defense-in-depth without broad refactor).

---

## ID format

| Option | Description | Selected |
|--------|-------------|----------|
| Keep GraphViz `edge<N>` | Fix ordering only; zero consumer/golden churn | ✓ |
| Emit explicit content-derived ids | e.g. `edge-sys.c0-sys.c1` — self-describing but changes DOT + all goldens | |

**User's choice:** [auto] Q: "Keep `edge<N>` format?" → Selected: "Keep GraphViz `edge<N>`" (recommended default — minimal churn; id format is not the bug).

---

## Verification contract

| Option | Description | Selected |
|--------|-------------|----------|
| TDD byte-equality from issue reproducer + all-format assertion + id-only golden deltas | RED→GREEN per tdd_mode; dot/svg/html asserted; canonicalDOT goldens as no-semantic-change gate | ✓ |
| Unit test on edge slice order only | Faster but doesn't pin the user-visible guarantee | |

**User's choice:** [auto] Q: "How to pin the fix?" → Selected: "TDD byte-equality from issue reproducer + all-format assertion + id-only golden deltas" (recommended default — matches REPRO-01..03 and project TDD precedent).

---

## Claude's Discretion

- Exact stable-sort key (Name vs (Source,Target,seq)) and fix placement (validator mirror vs buildEdges vs both) — planner/researcher decide from the real data flow.
- Regression-test package placement per existing test conventions.

## Deferred Ideas

None — discussion stayed within phase scope.
