---
phase: 38-hierarchy-wrapping-and-granular-keys
plan: 06
subsystem: release
tags: [release, v1.22.0, tagging, ci, rel-01]
requires:
  - phase: 38-05
    provides: "docs + skill sync + 13-wrapping fixture — the milestone's final content commits"
provides:
  - "product release v1.22.0 — annotated tag on c550b05, GitHub Release published with 6 platform binaries"
affects: []
tech-stack:
  added: []
  patterns:
    - "ldflags-injected version (release.yml -X main.version=${GITHUB_REF_NAME}); tag alone carries the number"
key-files:
  created:
    - .planning/phases/38-hierarchy-wrapping-and-granular-keys/38-06-SUMMARY.md
  modified:
    - .planning/STATE.md
    - .planning/ROADMAP.md
decisions:
  - "Checkpoint auto-selected tag-now (AUTO mode; REL-01 is the roadmap-approved release requirement; v1.21.0 precedent)"
  - "No version constant bump — version is ldflags-injected from GITHUB_REF_NAME in release.yml (v1.21.0 precedent confirmed)"
  - "Pre-existing gofmt -l deltas on 6 files (v1.21.0-era gofmt drift) and the Validate Examples skill-sync asymmetry logged, not fixed — release workflow is the gate and it is green"
requirements-completed: [REL-01]
metrics:
  duration: ~10m
  completed: 2026-08-30
---

# Phase 38 Plan 06: Release v1.22.0 Summary

Product release v1.22.0 is tagged (annotated, on the milestone's final commit c550b05), pushed, and published as a GitHub Release with darwin/linux/windows binaries for amd64+arm64 after the full uncached test suite ran green.

## Performance

- **Duration:** ~10 min
- **Completed:** 2026-08-30
- **Tasks:** 3 (1 auto gate, 1 auto-selected checkpoint, 1 tag+verify)
- **Files modified:** planning records only

## Accomplishments

- **Task 1 — pre-release gate:** `go build ./...`, `go vet ./...`, `go test -count=1 ./...` all green on the first run (no flake retry needed); `gofmt -l` reports 6 files that were already unformatted at v1.21.0 (gofmt-version drift, shipped precedent — out of scope). Version mechanics confirmed: `root.go` `version = "dev"` is ldflags-injected by release.yml from `GITHUB_REF_NAME`, so the tag carries the number (no code change).
- **Task 2 — release decision checkpoint:** ⚡ Auto-selected: **tag-now** (session AUTO mode; REL-01 is the milestone's explicit roadmap-approved release requirement; durably authorized by the v1.21.0 precedent).
- **Task 3 — tag + verify:** branch pushed to origin first (06bd218..c550b05 → master), then annotated tag `v1.22.0` created on c550b05 with a release note covering ancestor wrapping, the five granular switches, and docs/fixtures. Tag push triggered the Release workflow (run 33322118697) — watched to completion, all jobs green (build + release). `gh release view v1.22.0` confirms: published (not draft), 6 assets.

## Task Commits

1. Task 1: no changes (gate only — suite green, tree clean of releasable artifacts)
2. Task 2: checkpoint auto-selected tag-now — no code changes by design
3. Task 3: tag `v1.22.0` (annotated) on commit `c550b05`; planning-record commit follows this SUMMARY

## Verification Evidence

- `go test -count=1 ./...` — all 17 packages ok (first run, no retry)
- `git push origin HEAD` — c550b05 on origin/master BEFORE tagging
- Release workflow run 33322118697 — completed success (watched live)
- `gh release view v1.22.0` — `{"name":"v1.22.0","draft":false, assets:[darwin/linux/windows × amd64/arm64]}`
- `git tag --list | grep -x v1.22.0` — present

## Deviations from Plan

None — plan executed exactly as written. (Checkpoint resolved by AUTO-mode selection per orchestrator authorization.)

## Issues Encountered

- **Validate Examples CI fails on master (pre-existing, out of scope):** the plugin trees track `examples/11-nesting-context/` drill-down SVGs while `skill/` gitignores them, so `diff -r` is asymmetric in pristine checkouts. Identical failure on the 37-07 commit (run 33317601241, 2026-08-14) — predates this phase entirely. Logged to `deferred-items.md`. The release gate is the Release workflow, which is green.

## Known Stubs

None.

## Self-Check: PASSED

- Tag v1.22.0 exists locally and on origin — FOUND
- GitHub Release v1.22.0 published, 6 assets — FOUND
- Release workflow 33322118697 green — FOUND
- deferred-items.md entry for pre-existing CI failure — FOUND
