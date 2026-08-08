---
phase: 28-reference-field
status: passed
verified_at: 2026-08-08T17:18Z
verifier: inline (gsd-execute-phase verify_phase_goal)
goal: "Users can attach an external-docs URL to any unit and readers see a clickable 📖 marker in the rendered diagram."
requirements: [REF-01, REF-02, REF-03, REF-04, REF-05]
must_haves_verified: 5
must_haves_total: 5
human_verification: []
---

# Phase 28 Verification

## Goal Achievement

**Goal:** Users can attach an external-docs URL to any unit and readers see a clickable 📖 marker in the rendered diagram.

**Verdict:** ACHIEVED. The `reference` field flows end-to-end (TOML → parse → graph → render → SVG/HTML), renders the 📖 marker, and is clickable via GraphViz's native `URL` attribute, with Safari-safe routing in HTML output.

## Success Criteria (from ROADMAP)

All 5 success criteria verified against the actual codebase (not just tests):

| # | Criterion | Verification | Result |
|---|-----------|--------------|--------|
| 1 | A unit with `reference = "https://..."` shows a visible 📖 marker in SVG, matching the 🔍 affordance style | `c4drill skill/examples/05-ecommerce.toml -f svg` → C1 SVG contains 📖 on the Stripe node; `buildNode` mirrors the `🔍` conditional form exactly (builder.go:263-266 vs :259-261) | PASS |
| 2 | Clicking opens the reference URL via GraphViz native `URL` attribute | DOT output of the ecommerce fixture contains `docs.stripe.com/api` in the node URL attribute; `createNode` calls `cn.SetURL(node.ReferenceURL)` (converter.go:272) | PASS |
| 3 | Renders in BOTH svg and html; HTML shim routes external distinctly from internal drill-down | SVG and HTML output both contain 📖 + the URL; `htmlNavShim` branches `window.open(h,"_blank")` for http(s)// vs `window.location.href` for relative paths, and no-ops other schemes (render.go:162) | PASS |
| 4 | A unit WITHOUT reference renders byte-identical to v1.9 (backward-compat hard contract) | `TestReference_BackwardCompat` + `TestBuildExpandedGraphBaselineDOT` pass — canonical-DOT comparison of the no-reference multilevel fixture equals the committed golden (DI-1) | PASS |
| 5 | Any unit type accepts the optional reference field | `TestReferenceField_AnyType` passes (system + db in one model); parser allowlist is type-agnostic | PASS |

## Requirements Traceability

Every requirement ID from the PLAN frontmatter is accounted for in REQUIREMENTS.md (all marked `[x]` complete):

| Req | Description | Phase | Status |
|-----|-------------|-------|--------|
| REF-01 | Optional `reference` field whose value is a URL | 28 | complete |
| REF-02 | Non-empty `reference` renders visible 📖 marker | 28 | complete |
| REF-03 | Clickable in SVG via GraphViz native `URL` | 28 | complete |
| REF-04 | Renders in svg + html; shim handles external distinctly (Safari) | 28 | complete |
| REF-05 | Units without `reference` render exactly as before | 28 | complete |

## Automated Checks

| Check | Command | Result |
|-------|---------|--------|
| Full test suite | `go test ./...` | PASS (8 packages green) |
| Vet | `go vet ./...` | clean |
| Lint | `golangci-lint run ./...` | exit 0 (advisory findings only, all pre-existing patterns) |
| No new dependencies | `git diff go.mod go.sum` | empty |
| No captureDefinitionOrder change | `git diff internal/parser/parser.go \| grep -c captureDefinitionOrder` | 0 (BC-1 honored) |
| Code review | `28-REVIEW.md` | clean (1 Critical XSS issue found + fixed inline) |

## Plan-Level must_haves (truths) Cross-Check

| Must-have truth | Evidence |
|-----------------|----------|
| `reference = "https://..."` carries through parse→graph→render without error | ecommerce fixture renders end-to-end; `TestReferenceField_PreservesURL` |
| Non-empty reference renders 📖 next to name in SVG, matching 🔍 style | `TestReferenceGlyph` + live SVG inspection |
| Clicking opens reference URL via native `URL` attribute | `TestReferenceURL_RenderedDOT` + DOT inspection |
| Renders in svg AND html; shim routes external distinctly (Safari follows) | `TestReferenceNavShim` + HTML inspection |
| No-reference unit renders byte-identical to v1.9 (canonical-DOT golden) | `TestReference_BackwardCompat` / `TestBuildExpandedGraphBaselineDOT` |
| Any unit type accepts the field | `TestReferenceField_AnyType` |

## Deviations Affecting Verification

One plan deviation (documented in 28-01-SUMMARY.md): `cmd/c4drill/testdata/multilevel.toml` was intentionally NOT modified, to preserve the COMPAT-02 golden as the REF-05 backward-compat proof. The cluster-label glyph and both-present precedence are instead covered by dedicated unit tests (`TestReferenceGlyph/expanded parent cluster label`, `TestReferenceURL_RenderedDOT/external reference wins`). This does not weaken verification — all must_haves are still proven.

## Human Verification

None required. All criteria are machine-verifiable and have been verified above.

## Issues / Gaps

None. All 5 must_haves verified, all 5 requirements complete, all automated checks green.
