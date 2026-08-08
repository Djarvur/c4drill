# Phase 33: Docs sweep + end-to-end goldens - Research

**Researched:** 2026-08-08
**Domain:** Go test infrastructure + canonicalization helper extraction + TOML/docs authoring
**Confidence:** HIGH

## Summary

Phase 33 is the v1.10 milestone's integration-and-documentation phase. It ships **no new feature code** — only (a) reusable test infrastructure (the `canonicalDOT` order-insensitive comparator extracted from `internal/graph/builder_test.go:1249-1527` into a shared helper), (b) four end-to-end golden/behavioral tests proving the four v1.10 features (include, templates, relative-peer, reference) compose correctly through the full pipeline, (c) three new runnable example fixtures in `skill/examples/` plus one composed multi-file set, and (d) documentation gaps filled in `README.md` (user-facing) and `skill/SKILL.md` (agent-facing schema reference) covering templates, include, relative-peer, and the omittable-`type` inference rules (DOC-01). Phases 28 (reference) and 29 (optional-name humanization) already documented their features in both files; Phase 33 matches that established style and does not rework it (D-19).

The mechanical work is small in LOC but high in coordination value. The reusable `canonicalDOT` helper (D-18) is the linchpin: it already exists in two test call sites (`internal/graph/builder_test.go:1245` and `:2754`) and is needed by the new XC-05 E2E test in `cmd/c4drill/`. Extracting it once into a shared `internal/testutil/` package (recommended location — see Architecture Patterns) eliminates the duplication and makes future milestone goldens trivial. The XC-05 golden test then renders both the composed multi-file fixture and its hand-expanded single-file equivalent through the full pipeline (`c4drill` entry → `include.Resolve` → `template.Expand` → `peer.Resolve` → `validate` → views → render) and asserts the canonicalized DOT outputs are equal — proving include+templates+relative-peer compose to an indistinguishable result (XC-05) AND implicitly proving the pipeline ordering is correct (XC-01, per D-20).

**Primary recommendation:** Land the canonicalDOT helper extraction as Plan 1 (Wave 1) since every other test in this phase depends on it. Then fixtures (Wave 1, parallel), then docs (Wave 1, parallel — no code dependency), then the three E2E tests (Wave 2, depend on fixtures + helper). The behavioral XC-01 test and the golden XC-05 test share the composed fixture with separate assertion blocks (per CONTEXT D-20 discretion lean).

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**D-17 — DOC-03 example-fixture set (two tiers):**
1. Per-feature standalone examples in `skill/examples/`:
   - `06-templates.toml` — template define (`[template.*]`) + instantiate (`[[use]]`), demonstrating `${param}` substitution, fixed link set, subunit subtree (TMPL-04).
   - `07-relative-peer.toml` — bare peers resolving via walk-up (D-13/D-14/D-15), demonstrating sibling/aunt/root resolution + absolute fallback (ERGO-02).
   - `08-include/` — a small multi-file set: entry file + 2 included files, demonstrating transitive include, `once=true`, and cross-file subunits (D-10).
2. One composed multi-file example that uses templates + include + relative-peer + reference together — this doubles as the XC-05 golden fixture (both the multi-file AND its hand-expanded equivalent single-file ship as committed artifacts, so the "≡ single-file" test has real fixtures, not synthetics).
- Reference, omittable-name, and omittable-type are already demonstrated in `05-ecommerce.toml` (Phase 28 added Stripe's `reference`; Phase 29 documented humanization). No new fixture files needed for those three.

**D-18 — XC-05 golden test structure (reusable canonicalDOT helper + E2E test):**
- Build a reusable canonicalizer (extract `builder_test.go:1245-1259` approach per STATE.md DI-1 into a shared test helper, likely `internal/render/canonical_test.go` or a testutil package) that: parses DOT, strips layout geometry (`bb`/`pos`/`lp`/`lheight`/`lwidth`/`height`/`width`), sorts statements and attributes recursively.
- Test (in `cmd/c4drill/`, since it's an end-to-end CLI test): (1) render the composed multi-file fixture through the full pipeline; (2) render the equivalent hand-expanded single-file fixture the same way; (3) canonicalize both DOT outputs; (4) assert equality.
- E2E over model-level because it catches render-layer composition bugs (a templated unit's `reference` URL not rendering, a relative peer inside a template resolving to the wrong target) that a `parser.Model` deep-equal would miss.
- The reusable helper is built once here; future phases/milestones with golden tests reuse it. Plan 32 explicitly deferred SVG goldens to Phase 33 expecting this helper.
- Must use order-insensitive comparison (STATE.md DI-1).

**D-19 — DOC-02 depth + README/SKILL split (fill gaps, don't rewrite):**
- Phase 28 + 29 executors already added reference-field and humanization docs to both README.md and skill/SKILL.md in a consistent style. Phase 33 matches that style for the three missing features (templates, include, relative-peer) — no rework of shipped docs.
- **README.md (user-facing):** per feature, a syntax block + one short example + a pointer to the `skill/examples/` fixture. Plus DOC-01 (document `type` is omittable, with the inference-rules table from `parser.go:250`/`:276` and a before/after example).
- **skill/SKILL.md (agent-facing schema reference):** per feature, the full field table + validation rules + a pipeline-ordering note + interactions with other features. Plus DOC-01's inference-rules table.

**D-20 — XC-01 pipeline-ordering enforcement (behavioral test only):**
- Craft an input (or reuse the XC-05 composed fixture) where pipeline order is load-bearing — a templated unit with a relative peer that only resolves after `template.Expand`, plus an include that defines the template before a `[[use]]`. Render it; assert correct output.
- Robust to refactors that move pipeline calls into helper functions (a source-scan test would break; behavioral test would not).
- Pairs naturally with the XC-05 composed fixture — the golden already implicitly enforces correct ordering. So XC-01 and XC-05 can share the same fixture with different assertions.
- Rejected source-scan test (fragile to legitimate refactors) and pinned-comment-grep (weakest).

**Already-locked (carried from REQUIREMENTS / prior phases, NOT re-discussed):**
- **DOC-01:** Document `type` is omittable with the full inference rules (default type by parent at `parser.go:250`; generic `db`/`queue` promotion by nesting level at `parser.go:276`) and a before/after example. Lands in README + SKILL (per D-19).
- **XC-05 comparator:** order-insensitive canonicalDOT (STATE.md DI-1, realized as reusable helper per D-18).
- **Pipeline ordering:** `include → template-expand → relative-peer-resolve → humanize → validate → generate-views → render` — load-bearing for correctness (template peers must expand before relative resolution; include must resolve before templates defined in included files are visible). Documented in every feature phase's CONTEXT; Phase 33 adds the regression test (D-20) and a user-facing note in SKILL.md (D-19).

### Claude's Discretion

- Exact location of the reusable canonicalDOT helper (`internal/render/canonical_test.go` vs a new `internal/testutil/` package vs `cmd/c4drill/` test helper). Lean toward wherever `builder_test.go:1245-1259` lives most naturally alongside.
- Whether XC-01 and XC-05 share a single composed fixture with two assertion blocks, or ship as two tests reading the same fixtures. (Lean toward shared fixture, separate assertions, for clarity.)
- The composed fixture's domain (small synthetic, or sliced-down real model). Lean small synthetic — clearer for a golden, no external-model baggage.
- Whether to also assert rendered SVG (not just DOT) in XC-05. Lean DOT-only — the SVG-to-DOT path is already tested; SVG adds go-graphviz nondeterminism without strengthening the composition proof.

### Deferred Ideas (OUT OF SCOPE)

- **SVG-level golden assertion in XC-05** — DOT-only is sufficient for the composition proof; SVG adds nondeterminism without strengthening coverage.
- **User-facing tutorial walk-throughs** (beyond README syntax blocks) — the `skill/examples/` fixtures serve as the tutorials. A future docs phase could add a dedicated guide if users need more hand-holding.
- **API/CLI reference page** — out of scope; README's CLI section already covers it.
- **Reworking Phase 28/29 docs** — D-19 explicitly fills gaps only; reference + humanization sections stay as shipped.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| DOC-01 | README.md and skill/SKILL.md document that `type` is optional, with full type-inference rules + before/after example | Code inspection confirms `defaultTypeForParent` at `parser.go:250` (verified read) and `inferGenericType` at `parser.go:276` (verified read). Both functions are the single source of truth; the inference-rules table is a direct transcription. |
| DOC-02 | README.md and skill/SKILL.md document all four features (include, templates, ergonomics, reference) with syntax and examples | Phase 28 (reference) + 29 (optional-name) already documented in both files. Phase 33 fills templates, include, relative-peer in the same style. Established doc structure documented in Code Examples below. |
| DOC-03 | New example fixtures demonstrate each feature | Fixture design for 06/07/08 + composed set documented in Specific Ideas. Style template = `skill/examples/05-ecommerce.toml`. |
| XC-01 | Pipeline ordering enforced in code + test detects reordering as regression | D-20 behavioral test design in Specific Ideas. Shares composed fixture with XC-05. |
| XC-05 | Multi-file model (include + templates + relative peers) produces same rendered output (canonicalDOT) as equivalent hand-expanded single-file | D-18 E2E test design + reusable helper extraction documented in Architecture Patterns + Code Examples. |
</phase_requirements>

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| canonicalDOT helper extraction | `internal/testutil/` (new) | `internal/graph/` (current home) | Test-only code shared across two packages (`internal/graph` + `cmd/c4drill`) must live in a `*_test.go`-only or `internal/` importable testutil package; Go's test-package boundary forbids cross-package `_test.go` imports. |
| XC-05 E2E golden test | `cmd/c4drill/` | `internal/testutil/` (helper) | End-to-end CLI test — exercises the full pipeline in `root.go`. The existing `TestReference_BackwardCompat` at `internal/graph/builder_test.go:2734` is the precedent for golden-style end-to-end assertions, but it stops at `graph.BuildExpandedGraph`; the XC-05 test must go through `root.go`'s pipeline entry to catch include/template/peer composition. |
| XC-01 behavioral ordering test | `cmd/c4drill/` | composed fixture | Behavioral assertion reuses the XC-05 fixture; ordering correctness is proven by the fixture requiring the documented order to render correctly. |
| Fixture artifacts (06/07/08/composed) | `skill/examples/` | — | Established fixture home; the existing `01-05` graduated set lives here. The composed set needs a subdirectory (`skill/examples/08-include/` already; composed set likely `skill/examples/09-composed/` or under `cmd/c4drill/testdata/` per planner's call). |
| README.md docs | `README.md` (user-facing) | — | Per D-19: syntax block + short example + fixture pointer per feature. |
| skill/SKILL.md docs | `skill/SKILL.md` (agent-facing) | — | Per D-19: full field table + validation rules + pipeline-ordering note per feature. |

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| (stdlib `testing`) | go1.22+ | Test framework | Project already uses stdlib `testing` everywhere; `go test ./...` is the runner. [VERIFIED: go.mod + every `*_test.go` in repo] |
| testify (assert/require) | current (`require`/`assert` already imported) | Assertions | Already used in `internal/graph/builder_test.go:1235`, `cmd/c4drill/root_test.go:14`. [VERIFIED: import block at builder_test.go:6-14] |
| stdlib `sort`, `strings` | stdlib | canonicalDOT internals | The existing helper (`builder_test.go:1255-1527`) is pure stdlib — no new deps. [VERIFIED: code read] |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| (none new) | — | — | This phase adds ZERO dependencies. Pure refactor (extract) + new test files + new fixture TOML + doc edits. |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `internal/testutil/` package | keep helper duplicated in each `_test.go` | Duplication means copy-paste drift risk; the helper already exists in 2 call sites and a 3rd is needed. Single source of truth wins. Rejected. |
| `internal/render/canonical_test.go` | `internal/testutil/canonical/canonical_test.go` | `internal/render/` is an implementation package; a `*_test.go` there is only importable by `internal/render`'s own tests. The helper must be importable by `cmd/c4drill/` AND `internal/graph/` tests, so it needs a real (non-test) package. A `testutil` subpackage whose files are still `*_test.go`-named is the Go idiom for "test-only but importable". |

**Installation:**
```bash
# No installation — zero new dependencies.
go test ./...   # verify build + run all tests
```

**Version verification:** Not applicable — no packages installed. Go toolchain version confirmed via `go.mod` (go 1.22+ required; canonicalDOT uses only `sort`/`strings` which have been stable since Go 1.0).

## Package Legitimacy Audit

> Phase installs NO external packages. Audit N/A.

No packages added — this is a docs + test-refactor phase. Slopscheck/slopcheck gate does not apply.

## Architecture Patterns

### System Architecture Diagram

```
                  ┌─────────────────────────────────────────┐
                  │         Phase 33 deliverables           │
                  └─────────────────────────────────────────┘
                                   │
        ┌──────────────┬───────────┼───────────┬──────────────┐
        ▼              ▼           ▼           ▼              ▼
 ┌─────────────┐ ┌──────────┐ ┌────────┐ ┌──────────┐ ┌──────────┐
 │ canonicalDOT│ │  XC-05   │ │  XC-01 │ │ fixtures │ │   docs   │
 │   helper    │ │ E2E test │ │ behav. │ │ 06/07/08 │ │ README + │
 │ extraction  │ │(uses helper)│ │(shares │ │ + composed│ │ SKILL.md │
 │  (D-18)     │ │          │ │ fix.)  │ │  (D-17)  │ │ (D-19)   │
 └──────┬──────┘ └────┬─────┘ └───┬────┘ └────┬─────┘ └────┬─────┘
        │               │          │          │            │
        │  imports      │ renders  │ renders  │ read by    │ edits
        ▼               ▼          ▼          ▼            ▼
 ┌─────────────────────────────────────────────────────────────┐
 │  FULL PIPELINE (cmd/c4drill/root.go, exercised end-to-end)  │
 │  ParseFile → include.Resolve → template.Expand → peer.Resolve│
 │       → humanize → Validate → GenerateViews → RenderDOT      │
 │  (Phases 30/31/32 insert the three middle passes)            │
 └─────────────────────────────────────────────────────────────┘
```

The diagram shows the key insight: **the canonicalDOT helper is the shared dependency** for both E2E tests, and **the composed fixture is the shared input** for XC-01 + XC-05. Everything else is independent (docs, standalone fixtures, helper extraction have no cross-dependency within the phase).

### Recommended Project Structure

```
internal/
├── testutil/
│   └── canonical/              # NEW — reusable canonicalDOT helper (D-18)
│       ├── canonical_test.go   # the public Canonical() func (exported)
│       ├── parse.go            # parseDOTStatements + parseDOTBlock etc.
│       └── canonical_test.go   # (the existing TestCanonicalDOT* tests move here)
├── graph/
│   └── builder_test.go         # CHANGED — deletes the local copy, imports testutil/canonical
└── render/                     # unchanged
cmd/c4drill/
├── root_test.go                # CHANGED — adds XC-05 + XC-01 E2E tests
└── testdata/
    └── composed/               # NEW (D-17) — composed multi-file set + single-file equiv
        ├── entry.toml
        ├── templates.toml
        ├── domains/
        │   └── auth.toml
        └── single-file-equivalent.toml
skill/
├── examples/
│   ├── 06-templates.toml       # NEW (D-17)
│   ├── 07-relative-peer.toml   # NEW (D-17)
│   └── 08-include/             # NEW (D-17)
│       ├── entry.toml
│       ├── auth.toml
│       └── billing.toml
├── SKILL.md                    # CHANGED — adds templates/include/relative-peer/DOC-01 sections
README.md                       # CHANGED — adds same four sections in user-facing style
```

**Note on the `internal/testutil/` package boundary:** Go forbids importing `_test.go` files across packages. The canonicalDOT helper currently lives in `internal/graph/builder_test.go` (a `_test.go` file), so it is only reachable from within `internal/graph`. To make it importable by `cmd/c4drill/` tests, it must move to a package whose files are NOT all `*_test.go`. The idiom: a subpackage named `testutil/canonical` where the source files are named `canonical.go` (not `_test.go`) but the package itself is only ever imported from tests (enforced by living under `internal/`). This keeps the build clean (no production binary bloat) while allowing cross-package test import. The CONTEXT D-18 discretion leans "wherever it lives most naturally alongside" — `internal/testutil/canonical/` is the correct answer because `internal/render/`'s `*_test.go` files are not importable and duplicating into `cmd/c4drill/` defeats D-18's "reusable" goal.

### Pattern 1: Extract-and-Delegate (canonicalDOT helper)
**What:** Move the canonicalDOT functions + `dotStatement` type from `internal/graph/builder_test.go:1249-1527` into `internal/testutil/canonical/`, export them (`Canonical`, `dotStatement`→`Statement`), and have `builder_test.go` import the new package.
**When to use:** When the same test helper is needed in 2+ test packages.
**Why:** Eliminates the latent duplication risk (the helper is already referenced from 2 call sites in `builder_test.go` and a 3rd is needed in `cmd/c4drill/`).

### Pattern 2: Render-then-canonicalize golden (XC-05)
**What:** Render both fixtures through the full pipeline to DOT, canonicalize both with `canonical.Canonical(t, dot)`, assert equal with `require.Equal`.
**When to use:** Any "≡ equivalent" golden comparison where the render layer introduces nondeterministic ordering (go-graphviz map-order-dependent sibling order) or layout geometry.
**Why:** STATE.md DI-1 — byte-exact comparison fails spuriously on the pinned go-graphviz fork. The existing precedent is `builder_test.go:1245` (COMPAT-02) and `:2754` (REF-05). XC-05 follows the exact same pattern, just one layer up (full pipeline instead of `graph.BuildExpandedGraph` directly).

### Pattern 3: Behavioral ordering test (XC-01)
**What:** Reuse the XC-05 composed fixture (which is only renderable if the pipeline runs in the documented order — template defined in an included file must be visible before `[[use]]`; relative peer inside a template must resolve after `template.Expand`). Assert specific order-dependent properties of the output.
**When to use:** When the requirement is "detect reordering as a regression" but a source-scan test would be fragile to legitimate refactors (moving pipeline calls into helper functions).
**Why:** D-20 — the golden already implicitly enforces correct ordering (skip or reorder any pass and the multi-file templated relative-peer case breaks visibly). The behavioral test makes the ordering assertion explicit and named.

### Anti-Patterns to Avoid
- **Byte-exact golden comparison:** `require.Equal(t, expectedDOT, actualDOT)` will fail spuriously on go-graphviz layout nondeterminism. ALWAYS canonicalize first. (STATE.md DI-1.)
- **Source-scan pipeline ordering test:** Grep-ing `root.go` for the call sequence breaks the moment a refactor extracts the pipeline into a helper function. D-20 explicitly rejected this.
- **Duplicating the canonicalDOT helper:** Copy-pasting the ~280 LOC into `cmd/c4drill/root_test.go` defeats D-18's "reusable helper" goal and guarantees drift. Extract once.
- **Rewriting Phase 28/29 docs:** D-19 — fill gaps only. The reference and humanization sections in README/SKILL stay as shipped.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| canonicalDOT comparator | Re-derive the parse/serialize/sort logic | Extract the existing `builder_test.go:1249-1527` verbatim into `internal/testutil/canonical/` | The logic already exists, is tested (3 regression tests at `builder_test.go:1529-1591`), and encodes subtle invariants (WR-01 last-attribute preservation, WR-2 quoted-value non-truncation). Re-deriving risks regressing those. |
| Pipeline execution for the E2E test | Re-implement the pipeline inline in the test | Call through `root.go`'s entry (or the closest testable seam — see Code Examples) | The test's value is exercising the REAL composition; a reimplementation would test the test, not the pipeline. |
| Type-inference rules table | Paraphrase the rules | Transcribe `defaultTypeForParent` (`parser.go:250`) and `inferGenericType` (`parser.go:276`) directly | DOC-01 requires the FULL rules; paraphrasing risks omission. The code IS the source of truth. |

**Key insight:** Every deliverable in this phase has a verified prior-art source in the repo. The phase is extraction + transcription + composition, not greenfield authoring.

## Runtime State Inventory

> Phase 33 involves no rename/refactor/migration of production code (only test-helper extraction and doc/fixture additions). The canonicalDOT extraction moves code BETWEEN test files, not production runtime state. Skip.

N/A — no production runtime state touched. The only "state" is test fixtures (committed TOML files) and the canonicalDOT helper (test-only code moving between test packages). Neither has runtime persistence.

## Common Pitfalls

### Pitfall 1: Cross-package `_test.go` import boundary
**What goes wrong:** Moving the canonicalDOT helper to `internal/render/canonical_test.go` (a `_test.go` file) makes it unimportable from `cmd/c4drill/root_test.go`. Go's toolchain excludes `_test.go` files from the importable package surface.
**Why it happens:** Intuitive but wrong assumption that "test helper → `_test.go` file". The `_test.go` suffix restricts to same-package tests only.
**How to avoid:** Put the helper in a real (non-`_test.go`) package under `internal/` — e.g. `internal/testutil/canonical/canonical.go`. The `internal/` prefix prevents production import; the non-`_test.go` filename permits cross-package test import.
**Warning signs:** `cannot import` error during `go test ./cmd/c4drill/` after the extraction.

### Pitfall 2: Forgetting that Phases 30/31/32 must ship BEFORE this phase executes
**What goes wrong:** Running the XC-05/XC-01 tests before `include.Resolve`, `template.Expand`, `peer.Resolve` exist in `root.go` → tests fail to compile (undefined symbols) or fail semantically (pipeline doesn't have the passes).
**Why it happens:** The plan-phase pipeline runs now (planning is safe), but EXECUTION must wait for 30/31/32. The CONTEXT explicitly notes "30 just started, 31/32 planned".
**How to avoid:** Plans must declare `depends_on: phase-30,phase-31,phase-32` (or equivalent cross-phase dependency) so execute-phase refuses to run until those ship. Record the execution-wait in STATE.md. The plan-checker should verify the dependency is declared.
**Warning signs:** Tests referencing `include.Resolve`/`template.Expand`/`peer.Resolve` fail to compile at execution time.

### Pitfall 3: Fixture uses a feature the executing pipeline doesn't support yet
**What goes wrong:** The composed fixture (`composed/entry.toml`) uses `[[include]]` + `[template.*]` + `[[use]]` + bare relative peers. If authored against the v1.10 spec but the executing code is still v1.9, the parser rejects the reserved tables (BC-1 not yet landed) or the pipeline lacks the passes.
**Why it happens:** Same root cause as Pitfall 2 — feature code lands in 30/31/32, not 33.
**How to avoid:** Fixture TOML must conform to the FINAL v1.10 grammar (after 30/31/32 ship). The fixtures are authored now (planning) but only validated at execution time (after 30/31/32). This is fine as long as execution waits.
**Warning signs:** `unknown table 'template'` parse errors when running the fixture.

### Pitfall 4: Composed fixture's single-file equivalent is not actually equivalent
**What goes wrong:** The XC-05 test asserts `canonical(composed) == canonical(single-file-equivalent)`. If the hand-expanded single-file fixture has a subtle difference (a link that the template would have generated but the author forgot to inline; a relative peer that resolves differently), the test fails — and the failure mode is "the golden is wrong", not "the code is wrong".
**Why it happens:** Hand-expanding templates + includes + relative peers is error-prone — it's exactly the kind of mechanical work humans get wrong.
**How to avoid:** Author the composed fixture FIRST, render it (after 30/31/32 ship), capture the DOT, then hand-author the single-file equivalent and verify it canonicalizes equal BEFORE committing the golden. The single-file equivalent is the "expected" side; it must be correct. Consider generating it from the composed fixture's rendered output as a cross-check, then committing the hand-authored version (so the fixture stays human-readable).
**Warning signs:** XC-05 test fails on first run after 30/31/32 ship — investigate whether the fixture or the code is wrong before "fixing" either.

### Pitfall 5: canonicalDOT helper extraction breaks the existing COMPAT-02/REF-05 goldens
**What goes wrong:** Moving the helper changes its import path; if the move isn't clean (e.g. the helper was depending on a test-only fixture in `internal/graph/`), the existing `TestBuildExpandedGraphBaselineDOT` and `TestReference_BackwardCompat` break.
**Why it happens:** The helper is self-contained (pure stdlib) so this is LOW risk, but the extraction must preserve the exact function signatures and the `dotStatement` type.
**How to avoid:** After extraction, run `go test ./internal/graph/` and confirm both goldens still pass BEFORE touching anything else. The extraction is the first task; its acceptance criterion is "existing tests still green".
**Warning signs:** `TestBuildExpandedGraphBaselineDOT` or `TestReference_BackwardCompat` fail after extraction.

## Code Examples

### Existing canonicalDOT helper (the D-18 extraction source)
Source: `internal/graph/builder_test.go:1249-1527` [VERIFIED: code read]

The helper comprises:
- `canonicalDOT(t *testing.T, dot string) string` (line 1255) — the entry point
- `dotStatement` struct (line 1266) — `kind`, `head`, `attrs`, `children`
- `parseDOTStatements` (line 1275) — skips `digraph {` header
- `parseDOTBlock` (line 1288) — recursive statement-list parser
- `skipDOTWhitespace` (line 1323)
- `parseDOTSubgraph` (line 1335) — handles `subgraph ... { ... }`
- `parseDOTAttrStatement` (line 1359) — handles attr/node/edge statements
- `scanDOTValueEnd` (line 1384) — quoted-string + HTML-label aware
- `findDOTAttrTerminator` (line 1422) — finds `];` skipping quoted/HTML values
- `findDOTBlockOpen` (line 1441) — finds `{` skipping quoted/HTML in head
- `isGeometryAttr` (line 1457) — the `bb`/`pos`/`lp`/`lheight`/`lwidth`/`height`/`width` strip list
- `normalizeDOTAttrs` (line 1470) — split, strip geometry, sort
- `serializeDOTStatements` (line 1496) — recursive canonical serialization
- `serializeDOTStatement` (line 1509)

Plus three regression tests that MUST move with the helper:
- `TestCanonicalDOTPreservesLastAttribute` (line 1529) — WR-01
- `TestCanonicalDOTFinalAttributeDriftDetected` (line 1544) — WR-01 false-pass guard
- `TestCanonicalDOTQuotedValuesDoNotTruncate` (line 1563) — WR-2

**Extraction target signature:**
```go
// Package canonical provides an order-insensitive DOT canonicalizer for
// golden comparisons (STATE.md DI-1). The pinned go-graphviz fork emits
// map-order-dependent sibling order and layout geometry, so byte-exact
// comparison against committed goldens is impossible; this normalizes both
// sides to a sorted, geometry-stripped semantic form.
package canonical

// Canonical normalizes a DOT document into an order-insensitive semantic
// serialization. It is the public entry point; t is used for Helper() marking.
func Canonical(t *testing.T, dot string) string { ... }
```

`internal/graph/builder_test.go` then replaces its local `canonicalDOT(t, ...)` calls with `canonical.Canonical(t, ...)` and deletes lines 1249-1591.

### Existing E2E golden pattern (the XC-05 precedent)
Source: `internal/graph/builder_test.go:2720-2756` (`TestReference_BackwardCompat`) [VERIFIED: code read]

```go
func TestReference_BackwardCompat(t *testing.T) {
    m, err := parser.ParseFile("../../cmd/c4drill/testdata/multilevel.toml")
    require.NoError(t, err)
    valErrors := validator.Validate(m)
    require.Empty(t, valErrors, "model should be valid")
    v := view.GenerateExpandedView(m)
    g := graph.BuildExpandedGraph(v)
    dotData, err := render.RenderDOT(g)
    require.NoError(t, err)
    expected, err := os.ReadFile("../../cmd/c4drill/testdata/multilevel.expanded.dot")
    require.NoError(t, err)
    require.Equal(t, canonical.Canonical(t, string(expected)), canonical.Canonical(t, string(dotData)),
        "REF-05: a no-reference model must render identical to the v1.9 golden baseline")
}
```

**XC-05 difference:** This precedent stops at `graph.BuildExpandedGraph` (skips include/template/peer). The XC-05 test must invoke the FULL pipeline. Two options for the "full pipeline" seam:
1. **Pure-Go through the pipeline functions** (preferred — deterministic, no subprocess): call `parser.ParseFile` → `include.Resolve` → `template.Expand` → `peer.Resolve` → (humanize, which runs inside Parse or as a pass per Phase 29) → `validator.Validate` → `view.Generate*` → `graph.Build*Graph` → `render.RenderDOT`. This mirrors how `root.go`'s `runRoot` orchestrates them, just without the cobra/CLI scaffolding.
2. **Subprocess through the `c4drill` binary** (`go run ./cmd/c4drill entry.toml -f dot`): more truly "end-to-end" but slower and requires building the binary. The existing test style in this repo is pure-Go (option 1).

Recommend option 1 — it matches the existing `TestReference_BackwardCompat` style and is what "full pipeline" means in this codebase. The test lives in `cmd/c4drill/root_test.go` (alongside the existing CLI tests) because it's testing the pipeline composition that `root.go` owns.

### DOC-01 type-inference rules table (transcription source)
Source: `internal/parser/parser.go:250-317` [VERIFIED: code read]

The inference rules to document (for both README and SKILL):

**Default type by parent** (`defaultTypeForParent`, `parser.go:250`):
| Parent type | Default child type | Level |
|-------------|-------------------|-------|
| (none — root) | `system` | C1 |
| `system` | `container` | C2 |
| `box` | `system` | C1 (same-level grouping) |
| `container` | `component` | C3 |
| `containerBox` | `container` | C2 (same-level grouping) |
| `componentBox` | `component` | C3 (same-level grouping) |
| (other: db, queue, etc.) | `system` | C1 fallback |

**Generic db/queue promotion by nesting level** (`inferGenericType`, `parser.go:276`):
| Parent type | `db` becomes | `queue` becomes | Level |
|-------------|-------------|-----------------|-------|
| (none) or `box` | `db` | `queue` | C1 (unchanged) |
| `system` or `containerBox` | `containerDb` | `containerQueue` | C2 |
| `container` or `componentBox` | `componentDb` | `componentQueue` | C3 |

**Before/after example** (for both README and SKILL):
```toml
# BEFORE — explicit types (verbose)
[platform]
type = "system"
[platform.webapp]
type = "container"
[platform.webapp.cache]
type = "componentDb"

# AFTER — type omitted, inferred (DOC-01)
[platform]
# type omitted → inferred "system" (no parent)
[platform.webapp]
# type omitted → inferred "container" (parent is system)
[platform.webapp.cache]
# type omitted → inferred "componentDb" (parent is container, generic db promoted)
```

### Pipeline-ordering note (for SKILL.md, per D-19)
```
The v1.10 features compose through a fixed runtime pipeline:
  include.Resolve → template.Expand → peer.Resolve → humanize → validate → views → render
Ordering is load-bearing: include must run first (so templates defined in
included files are visible to [[use]]); template.Expand must run before
peer.Resolve (so relative peers authored in templates resolve at the
instantiation site); humanize runs after expand (so it sees the substituted
name, not ${param}) and before validate (so error messages show final names).
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| canonicalDOT duplicated per test package | Reusable `internal/testutil/canonical/` package (D-18) | Phase 33 | Future milestone goldens import one helper; no copy-paste drift. |
| DOC-01 type-inference undocumented | Full rules table in README + SKILL | Phase 33 | Users can omit `type` confidently; matches code at `parser.go:250`/`:276`. |
| No end-to-end composition proof | XC-05 golden + XC-01 behavioral (D-18/D-20) | Phase 33 | Refactors that break feature composition fail CI immediately. |

**Deprecated/outdated:**
- The local `canonicalDOT` copy in `internal/graph/builder_test.go:1249-1591` is superseded by the extracted `internal/testutil/canonical/` package. The local copy is DELETED as part of Plan 1 (the existing tests are rewritten to import the new package).

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | The pipeline seam for the E2E test is best invoked pure-Go (call the pass functions directly) rather than subprocess. | Code Examples | LOW — if subprocess is preferred, the test is slightly slower but otherwise identical. The existing test style (TestReference_BackwardCompat) is pure-Go, so this is the consistent choice. |
| A2 | `humanize` runs inside `Parse`/`ParseFile` (per Phase 29) rather than as a separate pipeline pass. | Code Examples | MEDIUM — if Phase 29 shipped humanize as a standalone pass between peer.Resolve and Validate, the E2E test must call it explicitly. The planner/executor must verify Phase 29's actual implementation site before finalizing the test's pipeline call sequence. (Phase 29 is Complete per ROADMAP; its SUMMARY confirms the site.) |
| A3 | The composed fixture's domain should be small synthetic (CONTEXT discretion lean). | Specific Ideas | LOW — a real-model slice would also work but carries external-model baggage. Synthetic is clearer for a golden. |
| A4 | The canonicalDOT helper should live in `internal/testutil/canonical/` (not `internal/render/`). | Architecture Patterns | LOW — the Go `_test.go` import boundary makes `internal/render/canonical_test.go` unimportable from `cmd/c4drill/`. `internal/testutil/canonical/` is the only location that satisfies D-18's "reusable across packages" goal. |

**If this table is empty:** N/A — 4 assumptions flagged for executor verification. A2 is the only one with non-trivial risk; the others are low-risk and align with CONTEXT discretion leans.

## Open Questions (RESOLVED — disposition in plans)

Both items are executor-verifiable at execution time (they depend on Phases 30-32 shipped code, which is not yet executing when this plan is authored). Neither blocks planning; both have concrete resolution paths encoded in Plan 04.

1. **Exact Phase 29 humanize implementation site**
   - What we know: Phase 29 is Complete (ROADMAP). XC-04 (humanize-after-expand) is assigned to Phase 31 with "end-to-end test in 33". The pipeline note says humanize runs after expand, before validate.
   - What's unclear: Is humanize a standalone pass callable as `humanize.Apply(m)`, or does it run inside `parser.ParseFile`/`parser.Parse`? This affects whether the E2E test calls it explicitly.
   - Recommendation: The executor reads Phase 29's SUMMARY.md (`/.planning/phases/29-*/SUMMARY.md`) and the actual `internal/parser/` or `internal/humanize/` code to confirm the call site before finalizing the XC-05 pipeline sequence. Non-blocking for planning.

2. **Composed fixture location: `skill/examples/09-composed/` vs `cmd/c4drill/testdata/composed/`**
   - What we know: CONTEXT D-17 says the composed set "doubles as the XC-05 golden fixture". D-18 says the test renders it. The CONTEXT discretion note says "likely lives in `skill/examples/08-composed/` or `cmd/c4drill/testdata/` (planner's call)".
   - What's unclear: Should the composed fixture live in `skill/examples/` (user-facing tutorial value) AND be read by the test, or in `cmd/c4drill/testdata/` (test-only) with a copy/symlink in `skill/examples/`?
   - Recommendation: Put it in `skill/examples/09-composed/` (user-facing — it's the best tutorial for "all four features together") and have the XC-05 test read from there via a relative path (`../../skill/examples/09-composed/entry.toml`). This matches the existing precedent where `TestReference_BackwardCompat` reads `../../cmd/c4drill/testdata/multilevel.toml` from `internal/graph/`. Cross-directory test reads are already established. The single-file-equivalent lives alongside it (it's part of the fixture pair).

## Environment Availability

> Phase 33 has no external dependencies beyond the Go toolchain already in use.

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go toolchain | all tests + build | ✓ | go1.22+ (per go.mod) | — |
| `go test` runner | all test tasks | ✓ | bundled | — |
| go-graphviz (pinned fork) | render layer (exercised by XC-05) | ✓ | already in go.mod | — |

**Missing dependencies with no fallback:** None.
**Missing dependencies with fallback:** None.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go stdlib `testing` + testify (`require`/`assert`) |
| Config file | none (Go convention — `go test` auto-discovers `*_test.go`) |
| Quick run command | `go test ./internal/testutil/canonical/ -run TestCanonical -x` |
| Full suite command | `go test ./...` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| XC-05 | composed multi-file ≡ single-file (canonicalDOT) | e2e golden | `go test ./cmd/c4drill/ -run TestXC05_ComposedEquivSingleFile -x` | ❌ Wave 2 (new) |
| XC-01 | pipeline ordering enforced (behavioral) | e2e behavioral | `go test ./cmd/c4drill/ -run TestXC01_PipelineOrdering -x` | ❌ Wave 2 (new) |
| DOC-01 | type-inference rules documented | doc assertion | `grep -c 'defaultTypeForParent\|inferGenericType' README.md skill/SKILL.md` (planner defines exact assertion) | ❌ Wave 1 (docs) |
| DOC-02 | all four features documented | doc assertion | `grep` for feature headings in README + SKILL | ❌ Wave 1 (docs) |
| DOC-03 | example fixtures exist | file existence | `ls skill/examples/06-templates.toml 07-relative-peer.toml 08-include/ 09-composed/` | ❌ Wave 1 (fixtures) |
| (helper regression) | extracted canonicalDOT still passes WR-01/WR-2 regression tests | unit | `go test ./internal/testutil/canonical/ -x` | ❌ Wave 1 (moved from internal/graph) |
| (backward-compat) | existing COMPAT-02/REF-05 goldens still pass after helper extraction | e2e golden | `go test ./internal/graph/ -run 'TestBuildExpandedGraphBaselineDOT|TestReference_BackwardCompat' -x` | ✅ exists (must stay green) |

### Sampling Rate
- **Per task commit:** `go test ./internal/testutil/canonical/ ./internal/graph/ ./cmd/c4drill/ -x` (the three packages this phase touches)
- **Per wave merge:** `go test ./...` (full suite — catches accidental cross-package breakage)
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `internal/testutil/canonical/canonical.go` — the extracted helper (Wave 1, Plan 1)
- [ ] `cmd/c4drill/testdata/composed/` OR `skill/examples/09-composed/` — the composed fixture set (Wave 1, but only validated in Wave 2 tests; see Pitfall 3)

*(Framework install: none needed — Go toolchain + testify already present.)*

## Security Domain

> Phase 33 adds no production code, no network calls, no user input handling, no crypto. The canonicalDOT helper is test-only and operates on already-rendered DOT strings. Test fixtures are committed TOML (no untrusted input). Security enforcement is enabled but no ASVS category applies to a docs+test phase.

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | no | N/A — no auth in this phase |
| V3 Session Management | no | N/A |
| V4 Access Control | no | N/A |
| V5 Input Validation | no | N/A — fixtures are committed, not user-supplied; the canonicalDOT parser operates on trusted test output |
| V6 Cryptography | no | N/A |

### Known Threat Patterns for docs+test phase

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| (none applicable) | — | The phase touches no production attack surface. The only "input" is committed fixture TOML parsed by the already-shipped v1.10 parser (Phases 30/31/32). |

## Sources

### Primary (HIGH confidence)
- `internal/graph/builder_test.go:1249-1591` — the canonicalDOT helper source + 3 regression tests [VERIFIED: code read, this session]
- `internal/parser/parser.go:250-317` — `defaultTypeForParent` + `inferGenericType` (DOC-01 transcription source) [VERIFIED: code read, this session]
- `cmd/c4drill/root.go:112-118` — the pipeline gap (Parse → Validate) where 30/31/32 insert passes [VERIFIED: code read, this session]
- `cmd/c4drill/root_test.go` — existing E2E/CLI test patterns [VERIFIED: code read, this session]
- `.planning/phases/33-*/33-CONTEXT.md` — D-17/D-18/D-19/D-20 locked decisions [VERIFIED: read, this session]
- `.planning/phases/30,31,32-*/CONTEXT.md` — pipeline ordering + feature interactions [VERIFIED: read, this session]
- `.planning/REQUIREMENTS.md` — DOC-01/02/03, XC-01/05 full text [VERIFIED: read, this session]
- `.planning/STATE.md` DI-1 — canonicalDOT order-insensitive mandate [VERIFIED: read, this session]
- `skill/examples/05-ecommerce.toml` — fixture style template [VERIFIED: read, this session]
- `README.md` + `skill/SKILL.md` — Phase 28/29 doc sections (style template + gap identification) [VERIFIED: read, this session]

### Secondary (MEDIUM confidence)
- `.planning/research/SUMMARY.md` §4/§9 — pipeline + canonicalDOT watch-out [VERIFIED: read, this session]

### Tertiary (LOW confidence)
- (none — all claims verified against code or planning artifacts in this session)

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — zero new deps; everything verified in go.mod / code.
- Architecture (helper extraction): HIGH — Go `_test.go` import boundary is well-documented; the extraction target is forced by the language.
- Pipeline/E2E test design: HIGH — direct precedent in `TestReference_BackwardCompat`; the only delta is calling the three new passes.
- Fixture design: HIGH — style template (`05-ecommerce.toml`) read; v1.10 grammar fully specified in REQUIREMENTS + prior CONTEXTs.
- Pitfalls: HIGH — all derived from verified code/CONTEXT facts, not speculation.
- Open Question A2 (humanize site): MEDIUM — depends on Phase 29's actual implementation, not re-verified in this session (Phase 29 SUMMARY not re-read; flagged for executor).

**Research date:** 2026-08-08
**Valid until:** 2026-09-07 (30 days — stable; the phase is docs+test, no fast-moving dependencies)
