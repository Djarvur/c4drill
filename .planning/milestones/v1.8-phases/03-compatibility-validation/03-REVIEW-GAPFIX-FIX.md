---
phase: 03-compatibility-validation
plan: 04
fixed_at: 2026-08-06T13:49:32Z
review_path: .planning/phases/03-compatibility-validation/03-REVIEW-GAPFIX.md
iteration: 1
findings_in_scope: 3
fixed: 3
skipped: 0
status: all_fixed
---

# Phase 3 (plan 03-04): Code Review Fix Report — Gap-Closure Delta

**Fixed at:** 2026-08-06T13:49:32Z
**Source review:** .planning/phases/03-compatibility-validation/03-REVIEW-GAPFIX.md
**Iteration:** 1

**Summary:**
- Findings in scope: 3 (WR-01, WR-02, WR-03; INFO findings IN-01/IN-02 out of scope by request)
- Fixed: 3
- Skipped: 0

## Fixed Issues

### WR-01: Navigation `<TD HREF>` URLs are not HTML-attribute-escaped; a `&` in a unit name silently drops the navigation bar

**Files modified:** `internal/render/navigation.go`, `internal/render/navigation_test.go`
**Commits:** `1c69ac1` (TDD failing tests), `fff0a52` (fix)
**Applied fix:** HTML-escape the URL when placing it into the `HREF="..."` attribute at both TD HREF sites: the back-link in `navigationTDs` (`internal/render/navigation.go` back-link block) and the breadcrumb in `breadcrumbItemTD` (`internal/render/navigation.go`). `url.PathEscape` does not escape `&` (verified: `url.PathEscape("a&b") == "a&b"`), and GraphViz's HTML-like label parser rejects a raw `&` that is not part of a valid entity, silently dropping the navigation anchor — empirically, a nav URL with `&` yields zero `<a xlink:href>` anchors in SVG. `html.EscapeString` escapes exactly `<`, `>`, `&`, `'`, `"`; already-encoded URL segments (`%20`, `%28`) contain none of those and are left intact, so legitimate URLs are unaffected. Node explore URLs via `SetURL` (a GraphViz attribute, not an HTML label) were NOT touched. The package doc comment on `BuildNavigationLabel` was updated to note URLs are now also HTML-escaped.

**TDD:** Failing tests first — `backlink_URL_with_ampersand_is_HTML-escaped_in_HREF_(WR-01)`, `breadcrumb_URL_with_ampersand_is_HTML-escaped_in_HREF_(WR-01)` (both asserted the raw `HREF="../r&d.svg"` was absent and `HREF="../r&amp;d.svg"` present; both failed RED with the raw `&` present), plus a regression guard `already-encoded_URL_is_unaffected_by_HTML-escape_(WR-01)` proving `api%20(v2).svg` is untouched. All GREEN after the fix.

### WR-02: `ComputeExploreURL` self-link guard is asymmetric and the empty-target cases emit broken `".svg"` (latent robustness gap)

**Files modified:** `internal/graph/path.go`, `internal/graph/path_test.go`
**Commits:** `1c69ac1` (TDD failing tests), `fff0a52` (fix)
**Applied fix:** Changed the guard from `currentPath != "" && targetPath == currentPath` to `if targetPath == "" || targetPath == currentPath { return "" }`. The original guard only fired when `currentPath` was non-empty, so empty-target (`ComputeExploreURL("a", "", ...)`) and both-empty (`ComputeExploreURL("", "", ...)`) cases collapsed to broken `".svg"` / `"basename/.svg"` URLs. The new guard is symmetric and rejects empty targets explicitly, so the function cannot emit a degenerate URL for any input. Unreachable through current callers (target IDs/ancestor paths are never empty) but matches the Gap 1 symptom B bug class and closes the trap for future callers. Doc comment updated to describe the generalized guard.

**TDD:** Failing tests first — `empty_target_with_non-empty_current_returns_empty_(WR-02)` (failed RED returning `".svg"`) and `empty_target_and_empty_current_returns_empty_(WR-02)` (failed RED returning `"diagram/.svg"`). Both GREEN after the fix. All pre-existing `ComputeExploreURL` subtests (ancestor, sibling, descendant, self-link) remain GREEN — the guard generalization does not regress any resolved case.

### WR-03: `TestComputeExploreURL` exceeds funlen limit (103 > 60) without the repo's `//nolint:funlen` directive

**Files modified:** `internal/graph/path_test.go`
**Commits:** `1c69ac1`
**Applied fix:** Added `//nolint:funlen // Test functions with many table-driven subtests are naturally longer` directly above the `TestComputeExploreURL` declaration, matching the established repo convention (`internal/graph/builder_test.go` uses the same directive on 9 functions; `internal/render/labels_test.go` uses `//nolint:funlen,paralleltest`). Note: this finding grew in severity because the WR-02 subtests were added by this same fix report (test commit), so the directive now covers 105 lines (was 103). Verified: `golangci-lint run ./internal/graph/...` no longer reports `TestComputeExploreURL` under funlen (only the two pre-existing, out-of-scope `builder_test.go` funlen findings remain).

**Verification (all three findings):**
- `go build ./...` — clean
- `go test -race ./...` — all 8 packages PASS
- `go vet ./internal/graph/ ./internal/render/` — clean
- `golangci-lint run --new-from-rev=b8831b0 ./...` (pre-03-04 baseline) — **0 issues**
- `golangci-lint run ./internal/graph/... ./internal/render/...` — no findings on `path.go`, `navigation.go`, `path_test.go`, `navigation_test.go` (the two pre-existing `builder_test.go` funlen findings are out of scope and predate this delta)
- Scope: only the 4 in-scope files modified (`internal/render/navigation.go`, `internal/render/navigation_test.go`, `internal/graph/path.go`, `internal/graph/path_test.go`); no production code outside those touched, no fixtures or other files changed

---

_Fixed: 2026-08-06T13:49:32Z_
_Fixer: Claude (gsd-code-fixer)_
_Iteration: 1_
