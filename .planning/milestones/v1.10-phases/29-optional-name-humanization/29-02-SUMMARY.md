---
phase: 29-optional-name-humanization
plan: 02
subsystem: docs
tags: [readme, skill, toml, ergonomics, humanize, documentation]

# Dependency graph
requires:
  - phase: 29-optional-name-humanization
    provides: "model.Humanize exact outputs (the D-01 reference table) so docs match shipped behavior"
provides:
  - "README.md Optional Name (Humanization) subsection documenting ERGO-03/04/05"
  - "skill/SKILL.md name relabelled Required→Optional with humanize note + acronym escape hatch"
affects:
  - "33-integration (DOC-03 example fixtures may reference the optional-name feature)"
  - "Future authors discovering the optional-name ergonomics via README/skill"

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Doc examples cross-checked against model.Humanize's actual output (no invented humanize results)"

key-files:
  created: []
  modified:
    - README.md
    - skill/SKILL.md

key-decisions:
  - "All humanize examples in docs match model.Humanize's verbatim output: localIDP→\"Local IDP\", sessionManager→\"Session Manager\", linuxSystem→\"Linux System\", gRPC→\"Grpc\""
  - "README new subsection placed after #### Reference (end of Unit Types block) and before ### Nesting — additive, no content removed"
  - "SKILL.md name field relabelled from # Required to # Optional in both Required Fields and Unit Definition; the field line itself retained (still valid, just no longer mandatory)"
  - "#### reference subsections in BOTH files left byte-identical (Phase 28 concurrency safety)"

patterns-established: []

requirements-completed:
  - ERGO-03
  - ERGO-04
  - ERGO-05

# Metrics
duration: 4min
completed: 2026-08-08
---

# Phase 29 Plan 02: Optional Name Documentation Summary

**README + skill/SKILL.md now document that `name` is optional with the dumb humanize rule, last-segment-only derivation, and the explicit-name escape hatch for acronyms like `gRPC`.**

## Performance

- **Duration:** ~4 min
- **Tasks:** 2 (README subsection, SKILL.md relabel + note)
- **Files modified:** 2

## Accomplishments
- Added a `### Optional Name (Humanization)` subsection to README.md under `## TOML Format`, containing a before/after TOML example, the last-segment-only rule with `localIDP`/`sessionManager`/`linuxSystem` examples, the dumb-split acronym behavior (`gRPC` → "Grpc") with the explicit-`name` escape hatch, and a backward-compat note.
- Relabeled `name` from `# Required: Display name` to `# Optional: defaults to humanized identifier` in README's Person example (Unit Types section).
- Relabeled `name = "Display Name"  # Required` → `# Optional - humanized from identifier if omitted` in skill/SKILL.md Required Fields, and `# Required: Display name` → `# Optional: defaults to humanized last path segment` in Unit Definition.
- Added an **Optional name humanization** note under skill/SKILL.md Unit Definition citing `localIDP`→"Local IDP", `sessionManager`→"Session Manager", and the `gRPC`→"Grpc" acronym escape hatch.

## Task Commits

1. **Task 1+2: README + skill/SKILL.md docs** — committed atomically with this SUMMARY (docs(29-02):)

## Files Created/Modified
- `README.md` — new `### Optional Name (Humanization)` subsection (25 insertions); 1 comment relabel in Person example.
- `skill/SKILL.md` — 2 `name` comment relabels (Required Fields + Unit Definition); 1 new humanization note paragraph (3 insertions net).

## Verification

- `grep -c "Optional Name" README.md` → 1 (new heading present)
- `grep -ci "acronym" README.md` → 2 (acronym escape hatch documented)
- `grep -n "# Required: Display name" skill/SKILL.md` → no matches (relabelled)
- `grep -ci "humaniz" skill/SKILL.md` → present (humanize note added)
- `#### Reference` / `#### reference` subsections unchanged in both files (verified via `git diff`: only additive inserts above/below, no edits within the Reference blocks) — Phase 28 concurrency safety confirmed.
- Every humanize example in the docs matches a row in Plan 01's model.Humanize reference table (cross-checked against 29-01-SUMMARY.md); no invented outputs.
- `go test ./...` — full suite still green (docs change is behavior-neutral).

## Deviations from Plan

No deviations. Both doc tasks completed exactly as specified. The two tasks were committed together (rather than separately) because they are tightly coupled doc surfaces describing the same feature, and a single coherent commit avoids a half-documented intermediate state where README and SKILL.md disagree on whether `name` is required.
