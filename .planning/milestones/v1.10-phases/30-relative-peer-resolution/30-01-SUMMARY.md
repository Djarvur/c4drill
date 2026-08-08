---
phase: 30-relative-peer-resolution
plan: 01
subsystem: parser
tags: [toml, ergonomics, peer-resolution, tdd, go]

# Dependency graph
requires: []
provides:
  - "internal/peer.Resolve(m *parser.Model) error — pure post-parse pass rewriting Link.Peer from relative bare to absolute dotted path per D-13/D-14/D-15/D-16"
  - "ResolveError{Peer, Host} sentinel error type for unresolvable bare peers"
  - "cmd/c4drill/testdata fixtures peer_walkup.toml (sibling/aunt/root/nearest-first) and peer_unresolvable.toml (miss-at-root error path)"
affects: [30-02, 31-template-expansion, 32-include, 33-docs-sweep-end-to-end-goldens]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Pure pre-validation rewrite pass (mirrors validator.BuildIndex recursion; in-place string mutation; no model-structure change)"
    - "Nearest-first ancestor-scope walk-up for lexical-scoping-style resolution"

key-files:
  created:
    - internal/peer/resolve.go
    - internal/peer/resolve_test.go
    - cmd/c4drill/testdata/peer_walkup.toml
    - cmd/c4drill/testdata/peer_unresolvable.toml
  modified: []

key-decisions:
  - "D-13/D-14/D-15/D-16 implemented verbatim; ROADMAP criterion 3 (same-depth ambiguity) confirmed structurally unreachable (Subunits map keys unique per parent) — defensive branch omitted as dead code per plan"
  - "Fail-fast error policy: first unresolvable bare peer returns *ResolveError (RESEARCH.md Open Question 2 — one clear error beats a wall)"
  - "Stdlib only (fmt, strings); no new dependency"

patterns-established:
  - "Pure function in internal/peer consuming *parser.Model and mutating Link.Peer in place — the template for future pre-validation passes (Phase 31 humanize, Phase 32 include merge)"
  - "Index-assignment in range loops (unit.Links[i].Peer = ...) to mutate slice elements rather than the range copy"

requirements-completed: [ERGO-01, ERGO-02]

# Metrics
duration: 8min
completed: 2026-08-08
---

# Plan 30-01: Relative-peer Resolve (core algorithm) Summary

**Pure post-parse pass rewriting every Link.Peer from relative bare to absolute dotted path, implementing D-13/D-14/D-15/D-16 with a nearest-first ancestor walk-up and a hard error on miss-at-root.**

## Performance

- **Duration:** ~8 min
- **Started:** 2026-08-08 (inline background executor)
- **Completed:** 2026-08-08
- **Tasks:** 2 (RED, GREEN+REFACTOR)
- **Files modified:** 4 (2 new src/test, 2 new testdata)

## Accomplishments
- Implemented `internal/peer.Resolve(m *parser.Model) error` — nil-safe, fail-fast, pure (no I/O, stdlib only). Mirrors `validator.BuildIndex` recursion and rewrites `Link.Peer` on both `Links` and authored `LinksFrom` via index assignment.
- Implemented the D-13/D-14/D-15/D-16 walk-up: a bare peer on host `a.b.c` searches `a.b.Subunits`, then `a.Subunits`, then root `m.Units`, first match wins (nearest-first; cross-depth shadowing silent); root match is an identity rewrite; miss at root yields `*ResolveError`.
- TDD coverage for every resolution branch: sibling, aunt (walk-up), root, nearest-first (cross-depth), dotted-untouched, unresolvable error (with `errors.As` to `*ResolveError`), LinksFrom, and corpus backward-compat (parser-corpus `valid/links/nested.toml` byte-identical `(unitPath, peer)` set).
- Two defensive tests beyond the plan's required 8: `TestResolveNilSafe` (nil model guard) and `TestResolveErrorFormat` (T-30-03 diagnostic wording).

## Task Commits

Each task was committed atomically:

1. **Task 1 (RED): failing tests for peer resolution** - `92d7555` (test)
2. **Task 2 (GREEN+REFACTOR): implement Resolve** - `d5b77fd` (feat)

The GREEN commit already contains the refactored, doc-commented shape (helpers extracted, decisions cited inline); no separate REFACTOR commit was warranted.

## Files Created/Modified
- `internal/peer/resolve.go` — `Resolve`, `resolveUnits`, `resolveUnitLinks`, `resolvePeer`, `ancestorScopes`, `scope` helper struct, `ResolveError` type. ~190 LOC incl. doc comments.
- `internal/peer/resolve_test.go` — 10 test functions (8 required + 2 defensive) + `walkUnits`/`collectPeerSet`/`mustParse`/`mustFindLink` helpers. Black-box `package peer_test`.
- `cmd/c4drill/testdata/peer_walkup.toml` — fixture exercising sibling, aunt, root, and nearest-first cases for Plan 02's pipeline test.
- `cmd/c4drill/testdata/peer_unresolvable.toml` — minimal fixture with a bare peer matching nothing, for Plan 02's CLI error-path test.

## Decisions Made
- **Same-depth ambiguity (ROADMAP criterion 3) treated as dead code:** the plan's CONTEXT flagged that under the walk-up model, a bare name can match at most one child per scope (each scope is a single parent's `map[string]*Unit`, unique keys). The defensive same-depth-multi-match error branch is omitted. A code comment in `resolvePeer` documents this and points at Plan 30 RESEARCH.md Pitfall 3. Cross-depth "ambiguity" is NOT an error (nearest-first handles it silently) — proven by `TestResolveNearestFirst`.
- **`scope` helper struct** introduced to pair each ancestor's children-map with its dotted parent path, so `resolvePeer` can construct the absolute peer path at match time without recomputing prefixes. Cleaner than re-walking segments inside `resolvePeer`.
- **Error format:** `cannot resolve peer %q from unit %q` — quotes both fields for clarity with dotted paths; matches `*parser.ParseError` verbosity (T-30-03).

## Deviations from Plan

None — plan executed exactly as written. The two extra tests (`TestResolveNilSafe`, `TestResolveErrorFormat`) are additive safety-net coverage, not a deviation from the plan's task structure.

## Issues Encountered
None.

## User Setup Required
None — no external service configuration required.

## Next Phase Readiness
- `internal/peer.Resolve` is ready for Plan 30-02 to wire into `cmd/c4drill/root.go` between `parser.ParseFile` and `validator.Validate`.
- The package is NOT yet wired into the pipeline — `go test ./...` is green because the new package is currently unreachable from main. Plan 02 connects it.
- HS-2 (Phase 31 forward-compat) is satisfied by construction: `resolveUnits` treats every unit uniformly (no template special-casing), so once Phase 31 inserts `template.Expand` before this pass in the pipeline, templated-unit peers resolve at the instantiation site automatically.

---
*Phase: 30-relative-peer-resolution*
*Completed: 2026-08-08*
