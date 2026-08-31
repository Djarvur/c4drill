---
phase: 39-edge-style-override-edges-cli-flag
plan: 03
subsystem: docs
tags: [docs, asciidoc, skill-sync, release, tagging]

requires:
  - phase: 39-01/39-02 (--edges implementation + matrix)
    provides: verified behavior to document (flag, precedence chain, --plain delta, RAW dot forms)
  - phase: 37 (37-06 skill sync precedent)
    provides: byte-identical 3-copy SKILL.md convention
  - phase: 38 (REL-01 release precedent)
    provides: tag-and-release workflow shape (6 binaries, gh release)
provides:
  - "--edges documented in README CLI Reference + narrative + --plain delta + properties.edges routing paragraph"
  - "3 byte-identical SKILL.md copies with the Edge Routing Override section"
  - "Release v1.23.0 shipped milestone v1.16 (published, 6 binaries)"
affects: []
tech-stack:
  added: []
  patterns: []
key-files:
  created: []
  modified:
    - README.adoc
    - skill/SKILL.md
    - plugins/c4drill/skills/c4drill-toml/SKILL.md
    - plugins/c4drill/opencode/skills/c4drill-toml/SKILL.md
key-decisions:
  - "Cross-reference for the author-side routing property placed in the Properties Section paragraph (README.adoc ~221) rather than the C4D '=== Edges' section — verified that section documents DSL arrows, not routing"
  - "--plain delta stated in TWO places (ignored-list item + composition notes) so the exact-union amendment cannot be missed"
requirements-completed: [GEDGE-03, GEDGE-06]
duration: 10min
completed: 2026-08-31
---

# Phase 39 Plan 03: Docs + Release v1.23.0 Summary

**--edges documented across README and 3 byte-identical SKILL.md copies with the --plain delta stated explicitly, and milestone v1.16 shipped as published release v1.23.0 with 6 binaries**

## Performance

- **Duration:** ~10 min
- **Started:** 2026-08-31 (wave 3)
- **Completed:** 2026-08-31
- **Tasks:** 3
- **Files modified:** 4 (+ release tag)

## Accomplishments
- README CLI Reference flags block carries `--edges` exactly as `--help` renders it; new narrative bullet covers the whole-invocation scope, global+per-unit precedence, the ortho alias, loud invalid-value errors, and `--plain` survival with copy-paste examples
- The `--plain` exact-union contract delta (D-05) is stated in both the ignored-list and the composition notes: explicit `--edges` is user intent and survives `--plain`; `--plain` alone still suppresses the author value
- The `properties.edges` routing paragraph (TOML Format → Properties Section) cross-references the per-invocation CLI override
- All three SKILL.md copies byte-identical (cmp verified) with a compact `### Edge Routing Override (--edges)` section
- Milestone v1.16 shipped: master pushed (22 commits), tag `v1.23.0` on `1f0d7ca`, Release workflow 33370530140 green, release published (not draft) with all 6 platform binaries

## Task Commits

1. **Task 1: README documentation** - `f513669` (docs)
2. **Task 2: SKILL.md 3-copy sync** - `1f0d7ca` (docs)
3. **Task 3: release v1.23.0** - tag `v1.23.0` pushed (release metadata tracked here)

## Files Created/Modified
- `README.adoc` — CLI Reference entry, `--edges` narrative bullet, two `--plain` delta statements, routing-property cross-reference
- `skill/SKILL.md` + 2 plugin copies — Edge Routing Override section + plain-ignored exception clause

## Decisions Made
- Routing cross-reference placed in the Properties Section (README.adoc ~221) where `edges` routing is actually documented, not the C4D `=== Edges` section named in the plan (that section covers DSL arrow syntax — verified before editing)

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Plan's cross-reference location was the wrong doc section**
- **Found during:** Task 1 (README edits)
- **Issue:** The plan pointed at `=== Edges` (C4D Format section) — that section documents C4D arrow statements, not the routing property
- **Fix:** Cross-reference added to the Properties Section routing paragraph (lines ~221-223), the section that actually documents `properties.edges` and unit-level `edges`
- **Files modified:** README.adoc
- **Verification:** grep confirms the xref sits directly after the routing resolution paragraph
- **Committed in:** f513669

---

**Total deviations:** 1 auto-fixed (line-number drift to the correct section)
**Impact on plan:** Intent honored, better location. No scope change.

## Issues Encountered
- None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Phase 39 scope fully delivered; ready for phase verification (gsd-verifier) and milestone closeout

---
*Phase: 39-edge-style-override-edges-cli-flag*
*Completed: 2026-08-31*
