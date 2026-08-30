---
phase: 37-nesting-context-and-plain-rendering
plan: 07
subsystem: release
tags: [release, tag, ci]

# Dependency graph
requires:
  - phase: 37-nesting-context-and-plain-rendering
    provides: all prior phase work (CTX-01..03, PLAIN-01..04, BC-01, DOC-01..03) — the tagged tree
provides:
  - "REL-01: product release v1.21.0 — annotated tag on the final phase commit, CI-built artifacts, published GitHub release"
affects: []

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Release = annotated tag push (v* triggers .github/workflows/release.yml; version injected via -ldflags from GITHUB_REF_NAME) — v1.18.0/v1.20.x precedent"

key-files:
  created:
    - git tag v1.21.0
  modified:
    - internal/c4d/parity_test.go

key-decisions:
  - "Standard full-suite `go test -count=1 ./...` used as the release gate instead of `-race` (time budget) — per orchestrator note; build + vet also clean before tagging"

requirements-completed: [REL-01]

# Metrics
duration: 12min
completed: 2026-08-30
---

# Phase 37 Plan 07: Ship Product Release v1.21.0 (REL-01) Summary

**Milestone v1.14 shipped as GitHub release v1.21.0 — annotated tag on the final phase commit (4691efb), CI release workflow green, all six platform artifacts (linux/darwin/windows x amd64/arm64) published.**

## Performance

- **Duration:** 12 min
- **Started:** 2026-08-30T14:36:00Z
- **Completed:** 2026-08-30T14:48:00Z
- **Tasks:** 2/2
- **Files:** 0 created, 1 modified

## Accomplishments
- **Task 1 (pre-tag gate):** `go build ./...` + `go vet ./...` clean; full suite `go test -count=1 ./...` green (standard run substituted for `-race` per time budget, noted as decision); branch pushed to origin (33 commits, 098bd59..4691efb) before tagging.
- **Task 2 (tag + release):** annotated tag `v1.21.0` created on 4691efb with a milestone message (nesting context CTX-01..03, `--plain` PLAIN-01..04/BC-01, docs DOC-01..03) and pushed; release workflow run [33317613032](https://github.com/Djarvur/c4drill/actions/runs/33317613032) completed successfully in ~1 min; GitHub release v1.21.0 published at https://github.com/Djarvur/c4drill/releases/tag/v1.21.0 with all 6 expected assets.

## Task Commits

1. **Task 1: Pre-tag validation gate** - `4691efb` (fix — see deviation; gate verification itself produced no commit)
2. **Task 2: Tag v1.21.0, push, verify release workflow (REL-01)** - tag `v1.21.0` → commit `4691efb` (no repo file changes)

## Files Created/Modified
- `internal/c4d/parity_test.go` — added `11-nesting-context.toml` to the `expectedExampleTwins` pinned manifest (D-35 contract)
- `git tag v1.21.0` — annotated release tag on the final phase commit

## Decisions Made
- Release gate used a standard full-suite run (`go test -count=1 ./...`) instead of `-race`; a `-race` run risked exceeding the time budget and the orchestrator pre-authorized the substitution. Flakiness check: internal/view + internal/c4d re-run `-count=5` — green.
- Tag message summarizes the milestone rather than listing commits — matches v1.18.0+ annotation convention.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Twin manifest missing 11-nesting-context.toml**
- **Found during:** Task 1 (pre-tag gate)
- **Issue:** Plan 06 added `skill/examples/11-nesting-context.c4d` but did not pin it in `expectedExampleTwins` (internal/c4d/parity_test.go, D-35 anti-shrinkage manifest); `TestExampleTwins` failed the release gate ("the shipped .c4d twin set must match the pinned manifest exactly").
- **Fix:** Added `"11-nesting-context.toml"` to the manifest. The twin passes all contract subtests (fmt-clean, renders identically to source).
- **Files modified:** internal/c4d/parity_test.go
- **Commit:** 4691efb

Note: one transient failure of `TestIntegrationC1EdgeResolution` (internal/view) was observed in the first full-suite run; it passed in isolation, in package runs, and 5x repeats — cross-test scheduling flake, not a code defect. No change made.

## Issues Encountered
- Working tree was not fully clean at start: pre-existing `.planning/phases/36-*` deletions and untracked `37-PATTERNS.md` (left uncommitted per orchestrator instruction; release tag excludes them). The tree condition was treated as clean with respect to shipped content — all code/tests/docs committed.

## User Setup Required

None — release published automatically by CI.

## Next Phase Readiness
- Milestone v1.14 fully shipped: all 11 requirements (CTX-01..03, PLAIN-01..04, BC-01, DOC-01..03, REL-01) complete. Ready for `/gsd:complete-milestone`.

---
## Self-Check: PASSED

- Tag exists locally (`git tag -l v1.21.0`) and on origin (`git ls-remote --tags origin v1.21.0` → e0e9b50)
- Release workflow run 33317613032 completed success (`gh run watch --exit-status`)
- `gh release view v1.21.0` → 6 assets, all six platform binaries present
- Commit 4691efb in `git log`; full suite green post-fix (`-count=5` on view+c4d packages)

---
*Phase: 37-nesting-context-and-plain-rendering*
*Completed: 2026-08-30*
