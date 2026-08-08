---
phase: 29-optional-name-humanization
status: passed
verified: 2026-08-08
goal: "Users can omit the `name` field and get a readable display name derived from the unit's identifier, reducing boilerplate."
requirements: [ERGO-03, ERGO-04, ERGO-05]
must_haves_verified: 13
must_haves_total: 13
gaps: 0
human_verification: []
---

# Phase 29 — Verification

**Status: passed** — all 13 must_have truths verified against the shipped codebase; 0 gaps.

## Phase Goal

> Users can omit the `name` field and get a readable display name derived from the unit's identifier, reducing boilerplate.

**Verified.** A unit authored without `name` now parses with `Unit.Name` derived from the last path segment via `model.Humanize` (a dumb camelCase split). Explicit `name =` always wins.

## Requirements Traceability

| Requirement | Status | Evidence |
|-------------|--------|----------|
| ERGO-03 (optional name, derived from last segment) | ✓ Verified | `internal/parser/parser.go:202` hook; `TestParseOmittedNameTopLevel`, `TestParseOmittedNameNestedSegment`; smoke test produced "Linux System" / "Local IDP" |
| ERGO-04 (dumb split, no acronym preservation) | ✓ Verified | `internal/model/humanize.go` — pure case-boundary split, no acronym data; all 9 D-01 rows pass incl. `gRPC`→"Grpc", `IDPToken`→"Idp Token" |
| ERGO-05 (explicit name always wins) | ✓ Verified | `if unit.Name == ""` guard; `TestParseExplicitNameWins`, `TestParseOmittedNameNoRegression` (corpus unchanged) |

ERGO-06 (compact-link) explicitly deferred per CONTEXT.md — out of scope for this phase.

## Must-Haves Verification

### Plan 29-01 truths

| # | Must-have truth | Status | Evidence |
|---|-----------------|--------|----------|
| 1 | `[linuxSystem.localIDP]` with no `name` parses with Name == "Local IDP" | ✓ | Smoke test: `linuxSystem.localIDP.Name = "Local IDP"` |
| 2 | Nested `[parent.child]` with no `name` humanizes from `child` segment-only | ✓ | `TestParseOmittedNameNestedSegment`; D-02 last-segment-only |
| 3 | `model.Humanize` produces the exact D-01 reference table (all 9 rows) | ✓ | `go run` verification: all 9 rows [OK] |
| 4 | Acronym preservation ABSENT — gRPC→"Grpc", no acronym allowlist | ✓ | grep: only comment-line matches; logic is pure case-split |
| 5 | Explicit `name =` always wins; humanize ONLY fires when Name == "" | ✓ | `TestParseExplicitNameWins` ("My gRPC Service" preserved); `if unit.Name == ""` guard |
| 6 | Every existing fixture parses to byte-identical Name values | ✓ | `TestParseOmittedNameNoRegression` on valid.toml + nested.toml — all names unchanged |
| 7 | Humanize is a no-op for any unit with non-empty name | ✓ | Same guard; covered by truth 5 + 6 |

### Plan 29-02 truths

| # | Must-have truth | Status | Evidence |
|---|-----------------|--------|----------|
| 8 | README TOML Format documents `name` as OPTIONAL with humanize rule + before/after example | ✓ | `### Optional Name (Humanization)` subsection; `localIDP`→"Local IDP", `sessionManager`→"Session Manager" examples |
| 9 | README documents dumb-split behavior + acronym escape hatch (gRPC→"Grpc") | ✓ | "Acronyms" paragraph in the subsection |
| 10 | README documents explicit name = always wins | ✓ | "Backward compatibility" + "set `name =` explicitly" in subsection |
| 11 | skill/SKILL.md reflects `name` as optional (not Required) with humanize + escape hatch | ✓ | Required Fields + Unit Definition relabelled; humanize note added |
| 12 | No existing doc content removed/contradicted; change purely additive | ✓ | `git diff`: 25 insertions README, 3 net SKILL; no removals |
| 13 | Doc humanize examples match parser's actual output | ✓ | Cross-checked against 29-01 reference table (same 4 examples) |

## Automated Verification

- `go test ./...` — **PASS** (8 packages: cmd, graph, model, output, parser, render, validator, view)
- `go test ./internal/model/... -run TestHumanize` — **PASS** (15 cases: 9 D-01 + 6 edge)
- `go test ./internal/parser/... -run 'OmittedName|ExplicitName'` — **PASS** (4 new test functions)
- `go vet ./...` — **clean**
- `go build ./...` — **clean**
- `go.mod` unchanged — **zero new dependencies** (stdlib `unicode`/`strings` only)

## Manual / Smoke Verification

- Ran a one-off program calling `model.Humanize` directly on all 9 D-01 inputs → all outputs match the contract table byte-for-byte.
- Ran a one-off program parsing an omitted-name + explicit-name TOML → top-level `linuxSystem`→"Linux System", nested `localIDP`→"Local IDP" (last segment only), explicit `gRPC` name preserved ("Explicit gRPC Wins").
- (Optional visual render smoke from VALIDATION.md skipped — automated parser tests are the real gate and they pass; the manual render would only re-confirm Unit.Name flows to the view, which the full `go test ./...` including `internal/view` already proves.)

## Backward Compatibility

- Every existing fixture (`testdata/valid.toml`, `testdata/nested.toml`) carries explicit `name =` and parses to byte-identical `Unit.Name` values (regression test passes).
- The `if unit.Name == ""` guard means the humanize fallback is strictly a no-op for any unit with a non-empty name — no existing model can be affected.
- validator/view/render/graph/cmd untouched; Unit.Name flows as opaque data downstream (full suite green confirms).

## Conclusion

Phase 29 achieved its goal. All ERGO-03/04/05 requirements verified; zero gaps; zero new dependencies; full suite green; backward compatibility proven. The `model.Humanize` function is the stable artifact Phase 31's XC-04 will reuse when relocating the call to a post-template-expansion pass.
