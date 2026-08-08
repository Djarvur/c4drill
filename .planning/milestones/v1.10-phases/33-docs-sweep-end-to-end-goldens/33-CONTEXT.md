# Phase 33: Docs sweep + end-to-end goldens - Context

**Gathered:** 2026-08-08
**Status:** Ready for planning

## Phase Boundary

Document all four v1.10 features (include, templates, ergonomics, reference) with runnable examples, complete DOC-01 (document omittable `type`), prove the features compose correctly end-to-end (XC-05: multi-file + templates + relative-peers ≡ equivalent single-file), and enforce the pipeline ordering as a regression test (XC-01). This is the milestone's integration + documentation phase — it ships no new feature code, only docs, fixtures, and tests.

This phase delivers DOC-01, DOC-02, DOC-03, XC-01, XC-05. It depends on Phases 30–32 (all feature work done).

## Implementation Decisions

### DOC-03 example-fixture set

- **D-17:** Two tiers of fixtures:
  1. **Per-feature standalone examples** in `skill/examples/` for learning:
     - `06-templates.toml` — template define (`[template.*]`) + instantiate (`[[use]]`), demonstrating `${param}` substitution, fixed link set, subunit subtree (TMPL-04).
     - `07-relative-peer.toml` — bare peers resolving via walk-up (D-13/D-14/D-15), demonstrating sibling/aunt/root resolution + absolute fallback (ERGO-02).
     - `08-include/` — a small multi-file set: entry file + 2 included files, demonstrating transitive include, `once=true`, and cross-file subunits (D-10).
  2. **One composed multi-file example** that uses templates + include + relative-peer + reference together — this doubles as the XC-05 golden fixture (both the multi-file and its hand-expanded equivalent single-file ship as committed artifacts, so the "≡ single-file" test has real fixtures, not synthetics).
  - Reference, omittable-name, and omittable-type are already demonstrated in `05-ecommerce.toml` (Phase 28 added Stripe's `reference`; Phase 29 documented humanization). No new fixture files needed for those three.

### XC-05 golden test structure

- **D-18:** **Reusable canonicalDOT helper + E2E test on the composed fixtures.** *(Best judgment — user did not answer.)*
  - Build a reusable canonicalizer (extract the `builder_test.go:1245-1259` approach per STATE.md DI-1 into a shared test helper, likely `internal/render/canonical_test.go` or a testutil package) that: parses DOT, strips layout geometry (`bb`/`pos`/`lp`/`lheight`/`lwidth`/`height`/`width`), sorts statements and attributes recursively.
  - Test (in `cmd/c4drill/`, since it's an end-to-end CLI test): (1) render the composed multi-file fixture from D-17 through the full pipeline (`c4drill` entry → `include.Resolve` → `template.Expand` → `peer.Resolve` → `validate` → views → render); (2) render the equivalent hand-expanded single-file fixture the same way; (3) canonicalize both DOT outputs; (4) assert equality.
  - **Why E2E over model-level:** catches render-layer composition bugs (e.g. a templated unit's `reference` URL not rendering, a relative peer inside a template resolving to the wrong target) that a `parser.Model` deep-equal would miss.
  - The reusable helper is built once here; future phases/milestones with golden tests reuse it. (Plan 32 explicitly deferred SVG goldens to Phase 33 expecting this helper.)
  - Must use order-insensitive comparison (STATE.md DI-1) — go-graphviz layout nondeterminism + an added ordering axis from multi-file/templates makes byte-exact `require.Equal` fail spuriously.

### DOC-02 depth + README/SKILL split

- **D-19:** **Fill gaps, don't rewrite.** *(Best judgment — user did not answer.)* Phase 28 + 29 executors already added reference-field and humanization docs to both README.md and skill/SKILL.md in a consistent style. Phase 33 matches that style for the three missing features (templates, include, relative-peer) — no rework of shipped docs.
  - **README.md (user-facing):** per feature, a syntax block + one short example + a pointer to the `skill/examples/` fixture. Plus DOC-01 (document `type` is omittable, with the inference-rules table from `parser.go:250`/`:276` and a before/after example).
  - **skill/SKILL.md (agent-facing schema reference):** per feature, the full field table + validation rules + a pipeline-ordering note + interactions with other features. Plus DOC-01's inference-rules table.

### XC-01 pipeline-ordering enforcement mechanism

- **D-20:** **Behavioral test only.** *(Best judgment — user did not answer.)*
  - Craft an input (or reuse the XC-05 composed fixture from D-17/D-18) where pipeline order is load-bearing — a templated unit with a relative peer that only resolves after `template.Expand`, plus an include that defines the template before a `[[use]]`. Render it; assert correct output.
  - Robust to refactors that move the pipeline calls into helper functions (a source-scan test would break on such a refactor; a behavioral test would not).
  - Pairs naturally with the XC-05 composed fixture — the golden already *implicitly* enforces correct ordering (skip or reorder any pass and the multi-file templated relative-peer case breaks visibly). So XC-01 and XC-05 can share the same fixture with different assertions.
  - Rejected the source-scan test (fragile to legitimate refactors) and the pinned-comment-grep (weakest — catches comment edits, not silent call reordering).

### Already-locked decisions (carried from REQUIREMENTS / prior phases, NOT re-discussed)

- **DOC-01:** Document `type` is omittable with the full inference rules (default type by parent at `parser.go:250`; generic `db`/`queue` promotion by nesting level at `parser.go:276`) and a before/after example. Lands in README + SKILL (per D-19).
- **XC-05 comparator:** order-insensitive canonicalDOT (STATE.md DI-1, now realized as the reusable helper per D-18).
- **Pipeline ordering:** `include → template-expand → relative-peer-resolve → humanize → validate → generate-views → render` — load-bearing for correctness (template peers must expand before relative resolution; include must resolve before templates defined in included files are visible). Documented in every feature phase's CONTEXT; Phase 33 adds the regression test (D-20) and a user-facing note in SKILL.md (D-19).

### Claude's Discretion

- Exact location of the reusable canonicalDOT helper (`internal/render/canonical_test.go` vs a new `internal/testutil/` package vs `cmd/c4drill/` test helper). Lean toward wherever `builder_test.go:1245-1259` lives most naturally alongside.
- Whether XC-01 and XC-05 share a single composed fixture with two assertion blocks, or ship as two tests reading the same fixtures. (Lean toward shared fixture, separate assertions, for clarity.)
- The composed fixture's domain (a small synthetic, or a sliced-down real model). Lean small synthetic — clearer for a golden, no external-model baggage.
- Whether to also assert the rendered SVG (not just DOT) in XC-05. Lean DOT-only — the SVG-to-DOT path is already tested; SVG adds go-graphviz nondeterminism without strengthening the composition proof.

## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Project planning (load-bearing context)
- `.planning/REQUIREMENTS.md` — DOC-01, DOC-02, DOC-03, XC-01, XC-05 (Phase 33 requirements). §Traceability maps each to Phase 33.
- `.planning/ROADMAP.md` — Phase 33 section: goal, requirements, 4 success criteria (incl. criterion 3: canonicalDOT order-insensitive comparison; criterion 4: pipeline ordering enforced + reorder regression test).
- `.planning/STATE.md` — DI-1 (canonicalDOT approach: sort-normalize, strip layout geometry, NOT byte-exact); the prior-art reference to `builder_test.go:1245-1259`.
- `.planning/todos/pending/2026-08-08-document-type-inference-omittable.md` — DOC-01 original intent (tagged resolves_phase: 33).

### Prior phase CONTEXTs (consolidate the pipeline + interactions)
- `.planning/phases/31-template-expansion/31-CONTEXT.md` — D-01..D-08 (templates; pipeline position; BC-1 parser prerequisite).
- `.planning/phases/32-include-directive-multi-file/32-CONTEXT.md` — D-09..D-12 (include; runs first in pipeline).
- `.planning/phases/30-relative-peer-resolution/30-CONTEXT.md` — D-13..D-16 (relative-peer; runs after template.Expand).

### Research
- `.planning/research/SUMMARY.md` — §4 pipeline + unchanged consumers; §9 golden-comparison watch-out (canonicalDOT, not byte-exact).
- `.planning/research/ARCHITECTURE-v1.10.md` — pipeline insertion points in `cmd/c4drill/root.go` (the gap the XC-01 behavioral test exercises).

### Code (integration points)
- `cmd/c4drill/root.go` — the pipeline the XC-01 test exercises end-to-end.
- `internal/graph/builder_test.go:1245-1259` — prior-art canonicalDOT approach (extract into reusable helper per D-18).
- `internal/render/` — where the canonicalDOT helper most likely lives.

## Existing Code Insights

### Reusable Assets
- `builder_test.go:1245-1259` canonicalDOT prior art — the exact comparison approach STATE.md DI-1 endorses. Extract + reuse (D-18).
- The existing `skill/examples/01-05` graduated tutorial set — match its style for the new `06-08` fixtures (D-17).

### Established Patterns
- **canonicalDOT, not byte-exact** (STATE.md DI-1) — every golden in this phase uses order-insensitive comparison. Non-negotiable; go-graphviz nondeterminism + multi-file ordering variance.
- **skill/examples/ as the fixture home** — the cleanest references per the format-mapping research. Per-feature standalones live here (D-17); the composed XC-05 fixture likely lives in `skill/examples/08-composed/` or `cmd/c4drill/testdata/` (planner's call).

### Integration Points
- **Docs**: `README.md` + `skill/SKILL.md` — fill gaps for templates/include/relative-peer + add DOC-01 omittable-type (D-19).
- **Tests**: `cmd/c4drill/` (E2E XC-05 + XC-01) + `internal/render/` (canonicalDOT helper, D-18).
- **Fixtures**: `skill/examples/06-templates.toml`, `07-relative-peer.toml`, `08-include/` + one composed multi-file set (D-17).

## Specific Ideas

- The composed XC-05 fixture shape (sketch for the planner):
  - `composed/entry.toml` — `[properties]`, `[[include]] path="templates.toml" once=true`, `[[include]] path="domains/auth.toml"`, `[[use]] template="microservice" parent="mainSystem" name="auth" ...` with a relative peer.
  - `composed/templates.toml` — `[template.microservice]` with a subunit + a `${param}` reference URL + a fixed link with a parametrized peer.
  - `composed/domains/auth.toml` — cross-file subunits appending to `[mainSystem]`.
  - `composed/single-file-equivalent.toml` — the hand-expanded equivalent (templates inlined, includes spliced, relative peers absolute).
  - XC-05 test renders `entry.toml` and `single-file-equivalent.toml` through the full pipeline, canonicalizes, asserts equal.
  - XC-01 test reuses the same `entry.toml` (or a narrower crafted input) and asserts the specific ordering-dependent behavior (relative peer inside a template resolves to instantiation-site sibling, not template-site).

## Deferred Ideas

- **SVG-level golden assertion in XC-05** — DOT-only is sufficient for the composition proof; SVG adds nondeterminism without strengthening coverage. Deferred (Claude discretion, D-18 note).
- **User-facing tutorial walk-throughs** (beyond README syntax blocks) — deferred; the `skill/examples/` fixtures serve as the tutorials. A future docs phase could add a dedicated guide if users need more hand-holding.
- **API/CLI reference page** — out of scope; README's CLI section already covers it.

---

*Phase: 33-Docs sweep + end-to-end goldens*
*Context gathered: 2026-08-08*
