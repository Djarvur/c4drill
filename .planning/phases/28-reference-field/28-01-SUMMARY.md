---
phase: 28-reference-field
plan: 01
subsystem: model
tags: [model-field, parser, render, graphviz, backward-compat, toml]

# Dependency graph
requires: []
provides:
  - "Unit.Reference optional external-docs URL field (TOML key `reference`)"
  - "📖 glyph rendering on node and expanded-cluster labels"
  - "Node.ReferenceURL plumbing through graph → render"
  - "External-wins URL precedence in createNode (GraphViz single-URL slot)"
  - "htmlNavShim branches external http(s)// vs internal drill-down vs no-op for non-http(s) schemes"
affects: [31-templates, 32-include-directive-multi-file]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Leaf-field isBuiltinField allowlist addition (no captureDefinitionOrder change needed)"
    - "Glyph-as-affordance appended to Label.Name (mirrors 🔍 collapsed-cluster precedent)"
    - "GraphViz single-URL-slot precedence — external reference wins over drill-down explore URL"

key-files:
  created:
    - ".planning/phases/28-reference-field/28-01-SUMMARY.md"
  modified:
    - "internal/model/unit.go"
    - "internal/parser/parser.go"
    - "internal/parser/parser_test.go"
    - "internal/graph/graph.go"
    - "internal/graph/builder.go"
    - "internal/graph/builder_test.go"
    - "internal/render/converter.go"
    - "internal/render/render.go"
    - "internal/render/render_test.go"
    - "README.md"
    - "skill/SKILL.md"
    - "skill/examples/05-ecommerce.toml"

key-decisions:
  - "Field name is `reference` (collision-free; `link`/`linkFrom` taken) — leaf field, not a reserved table, so isBuiltinField is the only parser change and NO captureDefinitionOrder modification is needed (research §9 BC-1)"
  - "Glyph is 📖 appended to Label.Name, matching the 🔍 collapsed-cluster affordance at builder.go:258-261 (ARCHITECTURE-v1.10.md §6 (6) Option A)"
  - "Clickability via GraphViz native URL attribute. A GraphViz node has a SINGLE URL slot, so external ReferenceURL wins over drill-down ExploreURL when both are present (precedence documented in converter.go)"
  - "No URL scheme validation at the model layer — empty-equals-omitted; the renderer's shim branches on http(s):// for routing, not for rejection"
  - "htmlNavShim hardened (T-28-02): non-http(s) schemes (javascript:, data:) are a no-op (preventDefault without navigation), not a fall-through to window.location.href"

patterns-established:
  - "Leaf-field allowlist: new scalar Unit fields join isBuiltinField as a one-liner; reserved tables (template/include/use) still need captureDefinitionOrder in Phase 31/32"
  - "Single-URL-slot precedence rule: when a node would carry both an external reference and an internal drill-down URL, external wins and the glyph is the visible affordance"

requirements-completed: [REF-01, REF-02, REF-03, REF-04, REF-05]

# Metrics
duration: 15min
completed: 2026-08-08
---

# Phase 28: Reference Field Summary

**Optional per-unit `reference` external-docs URL rendered as a clickable 📖 marker, wired through GraphViz's native URL attribute with a Safari-safe HTML shim that routes external links to a new tab.**

## Performance

- **Duration:** ~15 min
- **Started:** 2026-08-08T16:5XZ (phase begin recorded in STATE.md at 17:03:51Z)
- **Completed:** 2026-08-08T17:11Z
- **Tasks:** 3
- **Files modified:** 12 (10 source/test + 2 docs + 1 fixture)

## Accomplishments
- Added the optional `reference` field to `Unit` and registered it in the parser's `isBuiltinField` allowlist so it is decoded as a leaf field rather than treated as a phantom subunit (REF-01, BC-1). Any unit type accepts it.
- Rendered the 📖 glyph on both regular nodes (`buildNode`) and expanded-cluster labels (`buildClusterLabel`), and plumbed `Node.ReferenceURL` from `Unit.Reference` (REF-02, REF-03).
- Wired the external reference URL into GraphViz's single node URL slot with external-wins precedence over drill-down `ExploreURL`, documented inline (REF-03).
- Hardened the HTML nav shim to branch external `http(s)//` URLs to a new tab via `window.open`, keep internal `.html` drill-down in the same tab, and no-op non-http(s) schemes for XSS safety (REF-04, T-28-02).
- Proved backward compatibility: the committed `multilevel.expanded.dot` golden (no reference fields) still renders semantically identical to v1.9 via the canonical-DOT order-insensitive comparison (REF-05, DI-1).
- Documented the field in README and skill/SKILL.md and added a real docs URL (`https://docs.stripe.com/api`) to the Stripe unit in the ecommerce example, verified end-to-end (📖 glyph + URL present in the rendered C1 SVG).

## Task Commits

Each task was committed atomically:

1. **Task 1: Add Reference field to Unit + isBuiltinField entry** — `45e3d90` (feat) — RED parser tests → GREEN model+parser change
2. **Task 2: Render 📖 glyph + wire external URL through Node.ReferenceURL** — `5370d45` (feat) — RED graph/render tests → GREEN graph+render changes
3. **Task 3: Document reference field + ecommerce example** — `03a2b61` (docs)

**Plan metadata:** committed separately with SUMMARY/STATE/ROADMAP/REQUIREMENTS.

Tasks 1 and 2 were TDD (RED → GREEN). Task 3 was docs/fixture-only.

## Files Created/Modified
- `internal/model/unit.go` — added `Reference string` field on Unit after Technology.
- `internal/parser/parser.go` — added `"reference"` to the `isBuiltinField` allowlist; `captureDefinitionOrder` deliberately untouched (BC-1).
- `internal/parser/parser_test.go` — 4 new behavior tests (preserve URL, empty-equals-omitted, no phantom subunit, any-type).
- `internal/graph/graph.go` — added `ReferenceURL string` field on Node after ExploreURL.
- `internal/graph/builder.go` — `buildNode` appends 📖 + populates ReferenceURL; `buildClusterLabel` appends 📖.
- `internal/graph/builder_test.go` — new tests: TestReferenceGlyph (present/absent/cluster), TestReferenceURL_RenderedDOT (URL present + external-wins precedence), TestReference_BackwardCompat (REF-05 golden alias).
- `internal/render/converter.go` — `createNode` URL precedence: external reference wins the single URL slot over explore URL.
- `internal/render/render.go` — `htmlNavShim` branches external (window.open new tab) vs internal (window.location.href) vs non-http(s) (no-op); `svgHrefSuffix` unchanged.
- `internal/render/render_test.go` — new TestReferenceNavShim (REF-04 shim routing assertions).
- `README.md` — new "Reference (External Documentation URL)" subsection in the Unit Types section.
- `skill/SKILL.md` — `reference` row in the Unit Definition block + dedicated reference subsection.
- `skill/examples/05-ecommerce.toml` — `reference = "https://docs.stripe.com/api"` on the Stripe unit.

## Verification Results

Phase-level `<verification>` checks (all PASS):

1. `go test ./...` — entire suite green (cmd, graph, model, output, parser, render, validator, view).
2. `go vet ./...` — clean.
3. `golangci-lint run ./...` — exit 0 (project `.golangci.yml` treats findings as advisory; no new failures in modified files beyond the pre-existing `funlen`/`errcheck`/`gocognit` patterns already present in these files).
4. REF-01: parser carries the URL through to `Unit.Reference` (TestReferenceField_*).
5. REF-02: referenced unit's Node.Label.Name contains 📖; cluster label too; non-referenced does not (TestReferenceGlyph).
6. REF-03: rendered DOT carries the external URL via `cn.SetURL`; external-wins when both URLs present (TestReferenceURL_RenderedDOT).
7. REF-04: htmlNavShim branches `window.open` for external vs `window.location.href` for internal and no-ops non-http(s) (TestReferenceNavShim + T-28-02 hardening).
8. REF-05: canonical-DOT golden on the no-reference multilevel fixture is identical to the committed v1.9 baseline (TestReference_BackwardCompat / TestBuildExpandedGraphBaselineDOT).
9. `git diff go.mod go.sum` — empty (no new dependencies).
10. `git diff internal/parser/parser.go | grep -c captureDefinitionOrder` — 0 (no captureDefinitionOrder change, BC-1 honored).

End-to-end sanity: `c4drill skill/examples/05-ecommerce.toml -f svg` emits a C1 SVG containing both the 📖 glyph and the `docs.stripe.com/api` URL on the Stripe node.

## Decisions Made
None beyond the locked decisions already in the plan — all five locked decisions (field name `reference`, 📖 glyph via Label.Name append, GraphViz native URL attribute, no URL over-validation, canonical-DOT backward-compat contract) were followed as specified.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Conflicting plan instructions] Did NOT modify cmd/c4drill/testdata/multilevel.toml**
- **Found during:** Task 3 (document reference field + example fixtures)
- **Issue:** The plan's Task 3 instructed adding a `reference` to `cmd/c4drill/testdata/multilevel.toml` (to exercise the cluster-label glyph + both-present precedence in an integration test), but that fixture is the very one used by the plan's own Test L / the existing COMPAT-02 golden test (`TestBuildExpandedGraphBaselineDOT`), which compares rendered DOT against the committed `multilevel.expanded.dot`. Modifying the fixture would (a) break that golden, requiring regeneration, and (b) eliminate the no-reference fixture that REF-05's backward-compat contract depends on — directly contradicting the plan's REF-05 requirement and the STATE.md DI-1 canonical-DOT contract.
- **Fix:** Kept `cmd/c4drill/testdata/multilevel.toml` reference-free so the existing COMPAT-02 golden serves as the REF-05 backward-compat proof (it already renders byte-identical to v1.9). The cluster-label glyph path (Test G) and the both-present precedence (Test J) are instead covered by dedicated self-contained unit tests in `builder_test.go` using hand-built models, and the end-to-end feature is demonstrated on the real ecommerce example fixture (Stripe unit) which is NOT consumed by any committed golden. Added a `TestReference_BackwardCompat` alias pinning the REF-05 contract name to the same comparison.
- **Files modified:** `internal/graph/builder_test.go` (added the dedicated cluster-label + precedence tests; added the REF-05 alias). `cmd/c4drill/testdata/multilevel.toml` left UNCHANGED (one fewer file modified than the plan's `files_modified` frontmatter listed — the plan's done-criterion "multilevel.toml has exactly one reference" is intentionally not met, by design).
- **Verification:** All Task 2 behavior tests (E–L) pass; the COMPAT-02 golden (`TestBuildExpandedGraphBaselineDOT`) and its REF-05 alias (`TestReference_BackwardCompat`) both pass without regenerating any golden; `go test ./...` green.
- **Committed in:** `5370d45` (Task 2 commit).

---

**Total deviations:** 1 auto-fixed (Rule 1 — conflicting plan instructions resolved to preserve the REF-05 / DI-1 backward-compat contract).
**Impact on plan:** All five requirements (REF-01..05) are fully met. The deviation strictly narrows the integration-fixture choice to avoid breaking the plan's own backward-compat golden; no scope change, no feature loss. The `files_modified` frontmatter listed `cmd/c4drill/testdata/multilevel.toml` as modified — it is intentionally NOT modified (see deviation above).

## Issues Encountered
None.

## User Setup Required
None — no external service configuration required. The `reference` field is a pure local TOML → SVG/HTML rendering feature with no runtime dependencies beyond the existing GraphViz stack.

## Next Phase Readiness
- Phase 28 ships independently of all other v1.10 phases (28/29/30 are independent & parallelizable per STATE.md build order).
- The `internal/model/unit.go` and `internal/parser/parser.go` changes are the only v1.10 touches to those files so far; Phase 31 (templates, carries BC-1 parser change) has been held back to avoid colliding and now has a clean foundation. The `reference` leaf-field addition confirms the BC-1 precedent: scalar fields are a safe one-liner in `isBuiltinField`, while reserved tables (`template`/`include`/`use`) will still need the `captureDefinitionOrder` change in Phase 31/32.
- No blockers. The `cmd/c4drill/testdata/multilevel.expanded.svg` untracked file present in the working tree at phase start is pre-existing and unrelated to this phase.

---
*Phase: 28-reference-field*
*Completed: 2026-08-08*

## Self-Check: PASSED
