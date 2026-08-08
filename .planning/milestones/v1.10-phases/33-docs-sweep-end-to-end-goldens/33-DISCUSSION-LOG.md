# Phase 33: Docs sweep + end-to-end goldens - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-08-08
**Phase:** 33-Docs sweep + end-to-end goldens
**Areas discussed:** DOC-03 fixture set, XC-05 golden structure, DOC-02 depth, XC-01 enforcement mechanism

---

## DOC-03 example-fixture set

| Option | Description | Selected |
|--------|-------------|----------|
| Per-feature standalones + one composed XC-05 fixture | Two tiers: 06-templates, 07-relative-peer, 08-include standalones in skill/examples/ + one composed multi-file example doubling as XC-05 golden. Reference/name/type already in 05-ecommerce. | ✓ |
| Standalones only; XC-05 synthetic | Per-feature standalones; XC-05 uses inline synthetic (t.TempDir). Simpler; composition tested but not demonstrated as a doc artifact. | |
| Composed only; no standalones | One cumulative multi-file example; no per-feature standalones. Compelling composition proof but harder to learn each feature. | |

**User's choice:** Per-feature standalones + one composed XC-05 fixture

---

## XC-05 golden test structure

| Option | Description | Selected |
|--------|-------------|----------|
| Reusable canonicalDOT helper + E2E test on composed fixtures | Extract builder_test.go:1245-1259 approach into shared helper; render composed multi-file + equivalent single-file through full pipeline; canonicalize; assert equal. Reusable; catches render-layer bugs. | ✓ (best judgment) |
| Inline canonicalizer, no shared helper | Same assertion, canonicalizer local to test. Less reusable. | |
| Model-level equivalence (no rendering) | Deep-equal parser.Model post-passes. Faster but misses render-layer bugs. | |

**User's choice:** (no response — Claude applied best judgment: reusable helper + E2E test)
**Notes:** Plan 32 explicitly deferred SVG goldens to Phase 33 expecting this helper; E2E catches render-layer composition bugs (templated reference URLs, relative-peer-in-template resolution) that model-level would miss.

---

## DOC-02 depth + README/SKILL split

| Option | Description | Selected |
|--------|-------------|----------|
| Fill gaps; README=syntax+example, SKILL=full reference | Match 28/29's existing style for the three missing features. README: syntax block + short example + pointer to skill/examples/. SKILL: full field table + rules + interactions. Plus DOC-01 in both. | ✓ (best judgment) |
| Rewrite/consolidate into unified guide | Produce a unified 'v1.10 Authoring Guide' rewriting 28/29 additions. Higher consistency, more rework. | |
| Minimal — README syntax only | Syntax blocks only in README; SKILL unchanged except three feature tables. Least work. | |

**User's choice:** (no response — Claude applied best judgment: fill gaps)
**Notes:** 28/29's docs are consistent and shipped; rewriting is wasteful. Fill-gaps matches the two docs' established purposes.

---

## XC-01 pipeline-ordering enforcement mechanism

| Option | Description | Selected |
|--------|-------------|----------|
| Behavioral test only | Craft/ reuse an input where order is load-bearing (templated unit with relative peer resolving only after Expand, include defining template before use). Assert correct output. Refactor-safe; pairs with XC-05 fixture. | ✓ (best judgment) |
| Source-scan test only | Read root.go, assert the four pre-processing calls appear in order. Direct but fragile to helper refactors. | |
| Both source-scan + behavioral | Belt-and-suspenders. Two tests for one requirement. | |

**User's choice:** (no response — Claude applied best judgment: behavioral only)
**Notes:** XC-05's composed fixture already implicitly enforces ordering (reorder any pass → multi-file templated relative-peer breaks visibly). Behavioral test is refactor-safe; source-scan breaks on legitimate helper extraction.

---

## Claude's Discretion

- Exact location of canonicalDOT helper (`internal/render/canonical_test.go` vs `internal/testutil/` vs `cmd/c4drill/`).
- Whether XC-01/XC-05 share one fixture with two assertion blocks or ship as two tests.
- Composed fixture's domain (small synthetic vs sliced real model) — lean synthetic.
- SVG-level assertion in XC-05 (lean DOT-only).

## Deferred Ideas

- SVG-level golden in XC-05 (DOT-only sufficient).
- User-facing tutorial walk-throughs beyond README (skill/examples/ serve as tutorials).
- Dedicated API/CLI reference page (README's CLI section covers it).
