---
phase: 30-relative-peer-resolution
plan: 02
subsystem: cli
tags: [cli, pipeline, integration-tests, ergonomics, peer-resolution]

# Dependency graph
requires:
  - phase: 30-01
    provides: "internal/peer.Resolve(m *parser.Model) error (the resolver wired in here)"
provides:
  - "cmd/c4drill/root.go Stage 1.6 — peer.Resolve call between Parse and Validate"
  - "Pipeline integration tests proving resolve-before-validate ordering + CLI unresolvable-peer error path + corpus backward-compat"
affects: [31-template-expansion, 33-docs-sweep-end-to-end-goldens]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Pipeline staging comment convention (Stage 1 / Stage 1.6 / Stage 2) with cross-phase forward references"
    - "Cobra-root-command integration testing for error paths via NewRootCmd().SetArgs().Execute()"

key-files:
  created: []
  modified:
    - cmd/c4drill/root.go
    - cmd/c4drill/root_test.go
    - cmd/c4drill/testdata/peer_walkup.toml

key-decisions:
  - "Pre-existing TestFullPipeline_ValidationError fixture adapted: its bare peer=\"nonexistent\" now (correctly per D-16) fails at resolve before validation. Switched to dotted peer=\"no.such.unit\" so the test still exercises the validation undefined-unit path (dotted peers skip the resolver per D-16 step 1). This is a correct behavioral change, not a regression — bare unresolvable peers are now caught earlier with a clearer error."
  - "peer_walkup.toml redesigned to satisfy validator type rules (host/target must be leaves; no orphans; only system/box/container/containerBox/componentBox have subunits) while still exercising all four walk-up cases (sibling/aunt/root/nearest-first). The first draft used components-with-subunits which the validator rejects."

patterns-established:
  - "When adapting pre-existing tests to a new pre-validation pass, route fixtures that intend to test the VALIDATOR through dotted peers (which skip the resolver), reserving bare peers for resolver-specific tests."

requirements-completed: [ERGO-01, ERGO-02]

# Metrics
duration: 7min
completed: 2026-08-08
---

# Plan 30-02: Wire peer.Resolve into CLI pipeline Summary

**Connected the Phase 30 resolver to users: peer.Resolve now runs between Parse and Validate so authored bare peers resolve to absolute paths, with integration tests proving the ordering, the CLI error path, and corpus backward-compat.**

## Performance

- **Duration:** ~7 min
- **Started:** 2026-08-08 (inline background executor)
- **Completed:** 2026-08-08
- **Tasks:** 2 (wiring, integration tests)
- **Files modified:** 3 (root.go, root_test.go, peer_walkup.toml)

## Accomplishments
- Wired `peer.Resolve(m)` into `runRoot` as Stage 1.6, between `parser.ParseFile` (Stage 1) and `validator.Validate` (Stage 2), wrapped with `fmt.Errorf("resolve peers: %w", err)` matching the surrounding error idiom. Exactly one invocation; pipeline order XC-01 preserved.
- Added three integration tests to `cmd/c4drill/root_test.go`: `TestPipelineResolveBeforeValidate` (bare-peer fixture resolves + validates cleanly; sanity-checks >=1 bare peer present), `TestCLIUnresolvablePeerExits` (cobra root command on the unresolvable fixture returns an error naming the resolve stage and the peer, no panic), `TestCLICorpusRendersUnchanged` (valid/expanded/multilevel corpus fixtures parse+resolve+validate cleanly — ERGO-02 end-to-end).
- Adapted the pre-existing `TestFullPipeline_ValidationError` fixture: bare `peer="nonexistent"` now correctly fails at resolve (D-16); switched to dotted `peer="no.such.unit"` so the test still exercises the validator's undefined-unit path.

## Task Commits

1. **Task 1: wire peer.Resolve into CLI pipeline** - `7693b6d` (feat) — includes the validation-test adaptation in the same commit since the wiring caused the test to surface the new (correct) behavior.
2. **Task 2: add pipeline integration tests** - `177b23f` (test) — includes the peer_walkup.toml redesign needed to satisfy validator type rules.

## Files Created/Modified
- `cmd/c4drill/root.go` — added `internal/peer` import (alphabetically between parser and validator) + Stage 1.6 block (4 lines + comment).
- `cmd/c4drill/root_test.go` — added `internal/{model,parser,peer,validator}` imports + 3 test functions + `countBarePeers` helper; adapted `TestFullPipeline_ValidationError` fixture.
- `cmd/c4drill/testdata/peer_walkup.toml` — redesigned to be validator-clean (all hosts/targets are leaves; no orphans; subunits only on allowed types) while still covering sibling/aunt/root/nearest-first.

## Decisions Made
- **Validation-test fixture rerouted through a dotted peer:** the pre-existing `TestFullPipeline_ValidationError` used a bare unresolvable peer to reach the validator's undefined-unit error. With peer.Resolve wired in first, that bare peer is now caught at resolve (correct per D-16). Rather than weaken the resolver or special-case the test, the fixture was changed to a dotted peer (`no.such.unit`) which skips the resolver (D-16 step 1: dots = absolute) and reaches the validator unchanged. This preserves the test's intent (exercising the validation error path) and documents the new stage boundary.
- **peer_walkup.toml redesigned for validator compatibility:** the first draft placed subunits under `component`-typed units (e.g. `frontend.api.handlers`), which the validator rejects (components cannot have subunits). The redesign uses `container`/`componentBox`/`component` typing throughout, keeps every host and target a leaf unit (so VALD-02/VALD-03 hold), and gives every otherwise-orphan unit a link so `ValidateOrphanUnits` passes. The four walk-up cases (sibling, aunt, root, nearest-first) are all still exercised by bare peers.

## Deviations from Plan

None at the task-structure level. Two fixtures needed adjustment beyond a literal reading of the plan, both captured in the relevant task commits:

### Auto-fixed Issues

**1. Pre-existing validation test surfaced the new resolve stage**
- **Found during:** Task 1 (wire peer.Resolve into CLI pipeline) — full cmd test run after wiring.
- **Issue:** `TestFullPipeline_ValidationError` used bare `peer="nonexistent"`; with resolve now running first, the error became "resolve peers: ..." instead of a validation error, failing the `assert.Contains(..., "validation")`.
- **Fix:** Changed the fixture's peer to dotted `no.such.unit` so it skips the resolver (D-16 step 1) and reaches the validator's undefined-unit check. Added a comment documenting why a bare peer would now be caught earlier.
- **Files modified:** cmd/c4drill/root_test.go
- **Verification:** `go test ./cmd/c4drill/ -run TestFullPipeline_ValidationError -v` passes; error message contains "validation".
- **Committed in:** `7693b6d` (part of Task 1 commit — the wiring and the test adaptation are one semantic change).

**2. peer_walkup.toml first draft failed validator type rules**
- **Found during:** Task 2 (add pipeline integration tests) — TestPipelineResolveBeforeValidate reported 4 validation errors.
- **Issue:** The fixture used `component`-typed units with subunits and had an orphan + a C2-in-C2 typing violation, none of which are peer-resolution issues but all of which block the post-resolve Validate step the test asserts on.
- **Fix:** Redesigned the fixture with valid C1→C2→C3 typing (system/container/componentBox/component), made every host and target a leaf, and linked otherwise-orphan units. All four walk-up cases remain exercised by bare peers.
- **Files modified:** cmd/c4drill/testdata/peer_walkup.toml
- **Verification:** `go test ./cmd/c4drill/ -run TestPipelineResolveBeforeValidate -v` passes with zero validation errors.
- **Committed in:** `177b23f` (part of Task 2 commit).

---

**Total deviations:** 2 auto-fixed (both fixture correctness, no scope creep)
**Impact on plan:** All auto-fixes necessary for the integration tests to validate the end-to-end pipeline. No change to the resolver algorithm (Plan 30-01) or the wiring contract.

## Issues Encountered
None beyond the two auto-fixed items above. The Phase 29 humanize hook (parse-time, parser.go:202) was confirmed NOT present in root.go as a separate stage (it runs inside Parse), so there was no ordering conflict to resolve with the new peer.Resolve call.

## User Setup Required
None — no external service configuration required.

## Next Phase Readiness
- Phase 30 is complete: ERGO-01 (relative resolution) and ERGO-02 (absolute fallback + backward-compat) both ship. Users can now author bare peers and they resolve against the enclosing parent's ancestry.
- The pipeline insertion point (Stage 1.6) is correctly positioned for Phase 31: when template.Expand ships, it inserts as Stage 1.5 (after Parse, before peer.Resolve), and peer.Resolve's uniform unit walk then resolves templated-unit peers at the instantiation site automatically (HS-2 satisfied by construction).
- No open blockers. Phase 31 (templates) and Phase 32 (include) can proceed independently; Phase 33 (docs + goldens) will document the bare-peer ergonomics and add end-to-end goldens.

---
*Phase: 30-relative-peer-resolution*
*Completed: 2026-08-08*
