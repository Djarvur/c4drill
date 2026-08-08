---
phase: 29-optional-name-humanization
plan: 01
subsystem: testing
tags: [go, toml, parser, humanize, ergonomics, tdd]

# Dependency graph
requires: []
provides:
  - "model.Humanize(segment string) string — dumb camelCase splitter; stable artifact Phase 31 (XC-04) reuses"
  - "parse-time Unit.Name fallback in parseUnitWithOrder (fires when name omitted)"
  - "ERGO-03/04/05 shipped: optional name, dumb split (no acronym allowlist), explicit name wins"
affects:
  - "31-templates (XC-04 relocates the Humanize call to a post-expansion pass)"
  - "29-02 (docs reference model.Humanize's exact outputs)"
  - "33-integration (golden diagrams may now include omitted-name units)"

# Tech tracking
tech-stack:
  added: []  # zero new deps — stdlib unicode/strings only
  patterns:
    - "Hand-rolled rune scan for case-boundary splitting (no regexp, no acronym table)"
    - "Parse-time model mutation hook co-located with type defaulting/inference in parseUnitWithOrder"
    - "TDD RED/GREEN gate: 9-row reference table is the algorithm contract"

key-files:
  created:
    - internal/model/humanize.go
    - internal/model/humanize_test.go
  modified:
    - internal/parser/parser.go
    - internal/parser/parser_test.go

key-decisions:
  - "D-01 dumb split with NO acronym allowlist — gRPC → \"Grpc\"; acronym preservation is an anti-feature (ERGO-04)"
  - "Trailing pure-uppercase runs preserved verbatim (localIDP → \"Local IDP\", sessionAPI → \"Session API\"); leading/mid upper-runs lowercased (IDPToken → \"Idp Token\")"
  - "lower→upper split gated on preceding lowercase run length ≥ 2 so a lone leading lowercase stays glued (gRPC one word → \"Grpc\")"
  - "upper→upper→lower split gated on ≥ 2 trailing lowercase letters so a plural 's' does not tear an acronym (grpcAPIs → \"Grpc Apis\", not \"Grpc Ap Is\")"
  - "D-04 parse-time placement (NOT a runRoot pipeline stage) — XC-04 post-expansion relocation is Phase 31's contract"
  - "D-05 Humanize lives in internal/model — pure string util, no parser/toml dependency, reusable by Phase 31"

patterns-established:
  - "Table-driven pure-function test in internal/model (humanize_test.go is the first model-package test; mirrors parser_test.go testify style)"
  - "Backward-compat regression test pattern: snapshot existing fixtures' explicit name= values and assert byte-identical Unit.Name after the new fallback (TestParseOmittedNameNoRegression)"

requirements-completed:
  - ERGO-03
  - ERGO-04
  - ERGO-05

# Metrics
duration: 6min
completed: 2026-08-08
---

# Phase 29 Plan 01: Optional Name Humanization Summary

**model.Humanize dumb camelCase splitter + parse-time Unit.Name fallback: omit `name` and get "Local IDP" from `localIDP`, with explicit `name =` always winning (zero new deps).**

## Performance

- **Duration:** ~6 min
- **Started:** 2026-08-08T20:19:03+03:00
- **Completed:** 2026-08-08T20:21:59+03:00
- **Tasks:** 3 (RED, GREEN, REFACTOR-gate)
- **Files modified:** 4 (2 created, 2 modified)

## Accomplishments
- Shipped `model.Humanize` — a stdlib-only (unicode/strings) dumb camelCase splitter passing all 9 D-01 reference rows plus 6 edge cases, with no acronym allowlist (ERGO-04 anti-feature confirmed absent from the logic).
- Added a one-line parse-time hook in `parseUnitWithOrder` that populates `unit.Name` from the identifier segment when the author omits `name`; the `name` arg is already the last path segment for both top-level and nested units, so one insertion covers both (ERGO-03, D-02 last-segment-only).
- Proved backward compat (ERGO-05): explicit `name =` is never overwritten, and `TestParseOmittedNameNoRegression` asserts byte-identical `Unit.Name` values for every unit in testdata/valid.toml + nested.toml.
- Full suite green (`go test ./...`, `go vet ./...`, `go build ./...`) — validator/view/render/graph/cmd untouched, confirming Unit.Name flows as opaque data downstream.

## TDD Gate Compliance

| Gate | Commit | Type | Status |
|------|--------|------|--------|
| RED | `346d580` | `test(29-01):` | ✓ Failing (undefined: Humanize; omitted-name tests fail with empty Name) |
| GREEN | `124641d` | `feat(29-01):` | ✓ All 9 reference rows + 6 edge cases + 4 parser tests pass |
| REFACTOR | — | — | No refactor needed (code already well-decomposed into splitWords/titleWord/isLastPureUpper/hasLowerRun helpers); full-suite gate run green |

## Files Created/Modified
- `internal/model/humanize.go` (created) — `Humanize` + helpers `splitWords`, `titleWord`, `isLastPureUpper`, `hasLowerRun`; godoc cites ERGO-04 and Phase 31's XC-04 relocation.
- `internal/model/humanize_test.go` (created) — `TestHumanize` table test, 15 cases (9 mandatory D-01 rows + 6 edge cases), testify assert, `t.Run` subtests, `t.Parallel()`.
- `internal/parser/parser.go` (modified) — 9-line hook block in `parseUnitWithOrder` after `unit.Type = inferGenericType(...)`, before subunit processing.
- `internal/parser/parser_test.go` (modified) — 4 new tests appended: `TestParseOmittedNameTopLevel`, `TestParseOmittedNameNestedSegment`, `TestParseExplicitNameWins`, `TestParseOmittedNameNoRegression` (with valid.toml + nested.toml subtests).

## Boundary Heuristic (final, all 9 rows verified)

1. **lower→upper**: split when the preceding lowercase run has length ≥ 2.
   - `linuxSystem` → linux|System → "Linux System"; but the lone leading `g` in `gRPC` (run length 1) stays glued → one word → "Grpc".
2. **upper→upper→lower**: split when ≥ 2 lowercase letters follow the upper run.
   - `IDPToken` → IDP|Token → "Idp Token"; but `grpcAPIs` → grpc|APIs → "Grpc Apis" (the single trailing `s` does not tear "APIs").
3. **Trailing pure-uppercase word**: preserved verbatim.
   - `localIDP` → "Local IDP"; `sessionAPI` → "Session API"; `fetchURL` → "Fetch URL".
4. **All other words**: lowercased, then first rune ToTitle'd.
   - `webapp` → "Webapp"; `IDP` (non-trailing) → "Idp".

## Deviations from Plan

No deviations. The plan's boundary-heuristic "starting hypothesis" required two refinements discovered during RED→GREEN iteration (both are heuristic refinements within the plan's explicit "THE TABLE IS THE CONTRACT ... adjust the heuristic" allowance, NOT relaxations of the table):

1. **Trailing pure-uppercase preservation** — the plan's hypothesis said "Title-case each word: lowercase the whole word, capitalize the first rune," which would yield "Local Idp" / "Session Api" and FAIL the mandatory `localIDP`→"Local IDP" and `sessionAPI`→"Session API" rows. The contract table overrides the hypothesis: a trailing all-uppercase word is preserved verbatim. This was derived directly from the mandatory table, not invented.
2. **upper→upper→lower needs ≥ 2 trailing lowercase** — without this gate, `grpcAPIs` over-splits to "Grpc Ap Is" (failing the mandatory row). Requiring a 2-letter lowercase run following the upper cluster keeps "APIs" intact while still splitting "IDPToken".

Both refinements keep all 9 mandatory rows passing verbatim. No test row was altered or removed.

## Verification

- `go test ./internal/model/... -run TestHumanize` — all 15 cases PASS.
- `go test ./internal/parser/... -run 'OmittedName|ExplicitName'` — all 4 new tests PASS.
- `go test ./...` — full suite green (model, parser, validator, view, graph, render, output, cmd).
- `go vet ./...` — clean.
- `go build ./...` — clean.
- Zero new dependencies (go.mod unchanged).
