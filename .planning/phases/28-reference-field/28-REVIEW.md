---
phase: 28-reference-field
reviewed_phase: "28"
depth: standard
status: clean
reviewer: inline (gsd-execute-phase code_review_gate)
reviewed_at: 2026-08-08T17:15Z
files_reviewed:
  - internal/model/unit.go
  - internal/parser/parser.go
  - internal/parser/parser_test.go
  - internal/graph/graph.go
  - internal/graph/builder.go
  - internal/graph/builder_test.go
  - internal/render/converter.go
  - internal/render/render.go
  - internal/render/render_test.go
  - README.md
  - skill/SKILL.md
  - skill/examples/05-ecommerce.toml
summary: "Phase 28 reference-field changes reviewed. One Critical security issue found and fixed inline (XSS via scheme-prefixed reference URLs). All requirements verified."
---

# Phase 28 Code Review

## Scope

Reviewed all source/test/doc changes across the 4 phase-28 commits:
`45e3d90` (model+parser), `5370d45` (graph+render), `03a2b61` (docs+fixture),
`7df332b` (XSS fix).

## Findings

### Critical

**1. XSS via scheme-prefixed reference URLs in htmlNavShim (T-28-02 violation) — FIXED**

- **File:** `internal/render/render.go` (`htmlNavShim`)
- **Found during:** code_review_gate (inline review)
- **Issue:** The initial REF-04 shim used `/^(?:[a-z]+:)?\/\//i` as the "has a scheme" detector for the no-op branch. This regex REQUIRES `://`, so scheme-prefixed URIs without a slash pair fell through to `window.location.href=h`:
  - `javascript:alert(1)` → executed in the reader's browser
  - `data:text/html,<script>...` → rendered
  - `vbscript:msgbox(1)` → executed (legacy IE)
  - This directly violated the plan's threat-model T-28-02 hardening requirement ("for any href that has a scheme other than http(s), the click handler MUST no-op").
- **Root cause:** The regex conflated "has a scheme" with "has `://`". Most dangerous schemes (`javascript:`, `data:`, `vbscript:`) do not use `//`.
- **Fix:** Replaced the detector with `/^[a-z][a-z0-9+.\-]*:/i`, which matches ANY RFC-3986 scheme prefix. Classification is now:
  1. `^https?://` or leading `//` → `window.open(h, "_blank")` (external reference)
  2. any other `scheme:` prefix → no-op (preventDefault, no navigation)
  3. everything else (relative path) → `window.location.href=h` (internal drill-down)
- **Verification:** Confirmed classification against 15 cases including `javascript:`, `data:`, `vbscript:`, `mailto:`, `ftp:`, `file:`, protocol-relative, and relative paths. Strengthened `TestReferenceNavShim` to assert the generic scheme detector substring is present (the original Test K only checked for `window.open` + `http`, which is why the bug slipped through). `go test ./...` green.
- **Commit:** `7df332b` (`fix(28-01): no-op non-http(s) scheme reference URLs (T-28-02 XSS hardening)`)

### Warning

None.

### Info

**1. Boundary-cluster label does not carry the 📖 glyph (consistency, out of scope)**

- **File:** `internal/graph/builder.go` (`buildBoundaryCluster` at :96-130)
- **Observation:** `buildBoundaryCluster` constructs its label inline (builder.go:102-107) and does NOT call `buildClusterLabel`, so the 📖 glyph is not added to the C2/C3 boundary-cluster label when the expanded unit itself carries a reference. This mirrors the pre-existing 🔍 behavior (boundary clusters also never got 🔍). Phase 28 ships Option A only and this is cosmetic; the unit's reference URL still renders on its in-cluster node representations via `buildNode`. Flagging for awareness, not action.

**2. `createNode` URL precedence is single-slot by GraphViz design (documented)**

- **File:** `internal/render/converter.go` (`createNode` :267-275)
- **Observation:** When a node has both `ReferenceURL` and `ExploreURL`, only the external reference is emitted as the GraphViz `URL=` attribute; the drill-down URL is silently dropped for that node. This is the locked decision (ARCHITECTURE-v1.10.md §6 (6) Option A) and is documented inline with a pointer to the single-URL-per-node limitation. The 📖 + 🔍 glyphs remain as visual affordances. Correct as specified.

## Positive Observations

- The 📖 glyph implementation mirrors the existing 🔍 collapsed-cluster affordance pattern exactly (same conditional form, same `Label.Name += " <glyph>"` style), keeping the builder consistent and low-risk.
- The `isBuiltinField` addition is correctly scoped as a leaf-field one-liner; `captureDefinitionOrder` is deliberately untouched (verified: `git diff internal/parser/parser.go | grep -c captureDefinitionOrder` → 0), honoring BC-1.
- The REF-05 backward-compat contract is guarded by both the pre-existing `TestBuildExpandedGraphBaselineDOT` (COMPAT-02) and a new `TestReference_BackwardCompat` alias, both using the order-insensitive `canonicalDOT` comparison (DI-1).
- The deviation from the plan's Task 3 (not modifying `multilevel.toml`) is the correct call — it preserves the COMPAT-02 golden as the REF-05 proof and is fully documented in the SUMMARY.
- No new dependencies (`go.mod`/`go.sum` unchanged).

## Status: clean

All Critical findings fixed and verified. No Warning-level issues. Info findings are documented design choices. Phase 28 is ready for verification.
