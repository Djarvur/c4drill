# Deferred Items — Phase 01 (Fix C1 View Scoping)

Out-of-scope discoveries logged during plan execution (per executor scope boundary:
pre-existing issues in unrelated files are NOT fixed by plan 01-01).

## 1. Pre-existing golangci-lint debt (121 issues)

**Found during:** Plan 01-01 Task 3 (lint gate)
**Status:** Pre-existing — NOT caused by plan 01-01 changes.

- `mise lint` fails on the whole repo: 121 issues at plan completion vs 125 at the
  plan base commit (66ace06) — plan 01-01 introduced 0 new issues (verified with
  `golangci-lint run --new-from-rev=HEAD~2`).
- Breakdown (current): goconst 50, wsl_v5 16, gocognit 10, mnd 6, nlreturn 6, lll 5,
  modernize 4, testifylint 3, gosmopolitan 2, nestif 2, unused 2, plus gocritic,
  godot, intrange, maintidx, testpackage, unparam, wrapcheck, errcheck, funlen,
  exhaustive (1 each).
- Files with the most debt: internal/view/view_test.go, internal/graph/builder_test.go
  (pre-existing test functions), internal/render/labels_test.go, internal/graph/path.go,
  internal/parser/parser.go, internal/graph/shapes.go, internal/view/scope.go,
  internal/render/wrap.go.
- **Warning:** `mise lint` runs `golangci-lint run --fix` as its `lint-fix` dependency.
  Auto-fixing the whole repo would modify files outside the executing plan's scope —
  a dedicated cleanup plan (or fixing the mise task to drop lint-fix) is required.

## 2. `View.AllExpanded` not yet set by GenerateExpandedView

**Found during:** Plan 01-01 (D-02 mode discriminator)
**Status:** Planned handoff, not a bug — do NOT fix in 01-01.

- Until plan 01-02 sets `v.AllExpanded = true` in `GenerateExpandedView`
  (scope.go:14), `--expanded` graphs run through `buildEdges` with the pair-only
  dedup key (v1.7 tech+desc key inactive) and 2.0 penwidth on all edges via the
  AllExpanded branch of penWidth assignment.
- Verified inert for the saira fixture: it contains zero duplicate-pair links
  (same unit, same peer twice), so edge counts are unaffected. Full COMPAT-02
  byte-compat guard lands in plan 01-02.
