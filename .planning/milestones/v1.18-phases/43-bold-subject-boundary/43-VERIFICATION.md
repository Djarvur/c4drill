---
phase: 43
slug: bold-subject-boundary
status: passed
verified: 2026-09-07
verification_method: automated
requirements_verified: [BOLD-01, BOLD-02, BOLD-03]
must_haves_checked: 8/8
---

# Phase 43 Verification — Bold Subject Boundary

## Phase Goal

Every non-expanded view shows the subject unit's boundary with a bold triple-width (penwidth 3) border; other clusters unchanged; semantic so it survives `--plain`/`--no-styles`; expanded views and C1 root untouched.

## Requirement Verification

| Req | Statement | Evidence | Status |
|-----|-----------|----------|--------|
| BOLD-01 | Every non-expanded (drill-down) view draws the boundary of the depicted element with a bold triple-width border | `TestSubjectBoundaryClusterPenwidth` (C2 `cluster_mainSystem`, C3 `cluster_mainSystem.auth`: exactly one cluster carries `penwidth=3.0`, all others none); `TestSubjectBoundaryNoCollateral` deep-link C3 (`cluster_mainSystem.iam`); `TestSubjectBoundaryCLIFlagMatrix` C2 `multilevel/mainSystem.dot` + deep-link C3 `multilevel/mainSystem/sshAuth.dot` — all PASS on raw DOT | ✅ |
| BOLD-02 | Semantic — survives `--plain`/`--no-styles`; no other cluster on the same view affected | `--plain`, `--no-styles`, `--no-colors` variant subtests assert subject `penwidth=3.0` survives; per-cluster map asserts every other cluster carries zero penwidth; no new CLI flag (git diff of `cmd/c4drill/root.go` = none) | ✅ |
| BOLD-03 | Expanded-mode views, node borders, legend, edge styling unchanged; golden updates limited to subject-boundary delta | `TestSubjectBoundaryNoCollateral` — C1 root no `penwidth=3.0`; `--expanded` canonical-equal to `multilevel.expanded.dot` (COMPAT-02/REF-05 unchanged, byte-identical); golden audit: `go test ./... -count=1` green ×3, `git diff --quiet HEAD -- cmd/c4drill/testdata/` = YES (all 7 committed goldens byte-identical, ZERO re-baselines) | ✅ |

## Plan Truth Verification (must_haves)

**43-01 truths (4/4):**
1. Raw DOT carries `penwidth=3.0` on the subject boundary cluster ONLY — ✅ per-cluster attribute extraction asserts subject `["3.0"]`, every other cluster `[]` on C2, C3, deep-link, and all flag variants
2. Emphasis on C2, C3 and deep-link drill-down views of the multi-level fixture; collapsed C1 root carries no bold penwidth — ✅ `continue` matrix: `multilevel.dot` no `penwidth=3.0`; `mainSystem.dot` / `mainSystem/sshAuth.dot` carry it
3. `--plain` and `--no-styles` keep the bold boundary; `--expanded` output canonical-identical to goldens — ✅ variant subtests + COMPAT-02 canonical equality
4. `NodeStyle.BorderWidth` follows the 0-means-renderer-default contract — ✅ `BorderWidth float64` doc-comment + `> 0` emission gate; zero attribute when 0

**43-02 truths (3/3):**
1. Every committed golden byte-unchanged from v1.17 (zero re-baselines) — ✅ 7/7 goldens byte-identical
2. PROJECT.md documents the v1.18 semantic bold boundary — ✅ `As of v1.18` paragraph (grep -c = 1, tokens `penwidth=3.0`, `--plain`, `no new flag` all present)
3. Extended/audited render paths (`--expanded`, C1 root, legend, node borders, edges) show no diff — ✅ full-suite green, canonical consumers green, cluster-scoped assertions

## Automated Checks Run

- `go test ./... -count=1` — 19/19 packages ok (run 3×: post-GREEN, golden audit, post-tracking)
- `go test ./internal/graph/ ./internal/render/ ./cmd/c4drill/ -run 'TestSubjectBoundary' -count=1 -v` — 3/3 test functions pass (9 subtests)
- Regression gate (v1.17 prior-phase tests): `TestCheck`, `TestEdgeOrderNameSortedAndStable`, `TestDeterministicByteIdenticalOutput`, `TestMirrorOrder`, `cmd/c4drill-gui` — all ok
- `go vet ./internal/graph/ ./internal/render/ ./cmd/c4drill/` — clean
- `go build ./...` — clean
- Schema drift: `verify.schema-drift` — none (Go CLI, no DB schema)
- Codebase drift: warn (221 pre-existing structural elements, none phase-43) — non-blocking
- TDD gate: `git log --grep="test(43-01)"` → `7ed953f` precedes `git log --grep="feat(43-01)"` → `bf2007d` ✅ (RED before GREEN)

## Human Verification Items

None — all phase behaviors have automated verification. The optional visual confirmation (bold cluster frame thickness in generated SVG/PNG, per 43-VALIDATION) is explicitly "optional confirmation, never a gate".

## Conclusion

All 3 requirements (BOLD-01..03) verified against the codebase; 8/8 plan truths confirmed; full suite green; zero golden drift; TDD discipline honored. Phase goal achieved.