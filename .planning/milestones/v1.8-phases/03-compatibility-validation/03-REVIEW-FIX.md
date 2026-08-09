---
phase: 03-compatibility-validation
fixed_at: 2026-08-06T12:20:15Z
review_path: .planning/phases/03-compatibility-validation/03-REVIEW.md
iteration: 1
findings_in_scope: 2
fixed: 2
skipped: 0
status: all_fixed
---

# Phase 3: Code Review Fix Report

**Fixed at:** 2026-08-06T12:20:15Z
**Source review:** .planning/phases/03-compatibility-validation/03-REVIEW.md
**Iteration:** 1

**Summary:**
- Findings in scope: 2 (WR-01, WR-02; INFO findings out of scope by request)
- Fixed: 2
- Skipped: 0

## Fixed Issues

### WR-01: canonicalDOT off-by-one drops the last character of the last attribute — demonstrated false-pass window

**Files modified:** `internal/graph/builder_test.go`
**Commits:** `94da191` (fix), `48f4632` (TDD regression tests)
**Applied fix:** `parseDOTAttrStatement` now passes the full attribute block to `normalizeDOTAttrs` (`text[open+1:]` instead of `text[open+1:len(text)-1]`); the per-piece `TrimSpace` handles trailing whitespace. The off-by-one dropped the final character of the last attribute of every statement — empirically all 56 golden edges canonicalized as `style=soli`, and `penwidth=1` vs `penwidth=2` as a final attribute both collapsed to `penwidth=` (the false-pass window). Regression tests: `TestCanonicalDOTPreservesLastAttribute` (exact canonical form with `style=solid` intact) and `TestCanonicalDOTFinalAttributeDriftDetected` (`penwidth=1` vs `penwidth=2` final attributes must canonicalize differently) — both failed before the fix, pass after.

### WR-02: canonicalDOT statement terminators are not value-aware — `];`, `{`, `}` inside attribute values truncate and desync parsing

**Files modified:** `internal/graph/builder_test.go`
**Commits:** `94da191` (fix), `48f4632` (TDD regression tests)
**Applied fix:** Terminators are now located with quoted-string and HTML-label awareness. New helper `scanDOTValueEnd` advances past a `"`-quoted value (backslash escapes handled) or an HTML `<...>` region (nesting tracked); `findDOTAttrTerminator` uses it to locate the real `];` and `findDOTBlockOpen` to locate the subgraph `{`, so `];`, `{`, `}` inside attribute values no longer truncate the statement or shift the parse of everything after it. An unterminated statement still fails the parse loudly (`ok=false`), preserving the fail-loud invariant. Regression tests: `TestCanonicalDOTQuotedValuesDoNotTruncate` (description `"SSH [session]; admin"` and `label="uses {braces}"` survive; following statements still parsed) and `TestCanonicalDOTHTMLLabelDoesNotTruncate` (`];` inside `<b>...</b>` HTML label) — both failed before the fix, pass after.

**Verification:**
- `go test ./internal/graph/ -run BaselineDOT -count=1` — PASS (COMPAT-02 golden comparison still green on the corrected canonical form)
- `go test ./... -count=1` — all packages PASS; `go vet ./...` clean
- `golangci-lint run --new-from-rev=3d07654 ./...` (fresh cache) — 0 issues; new code kept clean (blank-line fixes for nlreturn/wsl_v5)
- Scope: only `internal/graph/builder_test.go` modified; no production code or test fixtures touched

---

_Fixed: 2026-08-06T12:20:15Z_
_Fixer: Claude (gsd-code-fixer)_
_Iteration: 1_
