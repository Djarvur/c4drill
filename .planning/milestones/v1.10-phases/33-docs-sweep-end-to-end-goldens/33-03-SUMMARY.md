---
phase: 33-docs-sweep-end-to-end-goldens
plan: 03
subsystem: docs
tags: [readme, skill, doc-01, doc-02, pipeline-ordering]

# Dependency graph
requires:
  - phase: 28-reference-field
    provides: reference field docs (established the style Phase 33 matches)
  - phase: 29-ergonomics-humanization
    provides: humanization docs (established the style Phase 33 matches)
  - phase: 33-docs-sweep-end-to-end-goldens
    provides: "Plan 33-02 fixtures (skill/examples/06-09) that the doc pointers reference"
provides:
  - "README.md Optional Type + Templates + Multi-File Composition + Relative Peer sections (DOC-01/02)"
  - "skill/SKILL.md expanded Type inference rules + Pipeline Ordering + Templates/Include/Relative Peer schema sections (DOC-01/02)"
affects: []

# Tech tracking
tech-stack:
  added: []
  patterns: ["Function-name citation in doc footnotes (line numbers drift; names are stable)"]

key-files:
  created: []
  modified:
    - README.md
    - skill/SKILL.md

key-decisions:
  - "Cited defaultTypeForParent + inferGenericType by function NAME, not line number. The plan/RESEARCH cited parser.go:250/276 but the functions are now at parser.go:673/699 (code drifted since the plan was written). Names are stable; line numbers aren't."
  - "SKILL.md version bumped 1.1.0 -> 1.2.0 and C4Drill version v1.9+ -> v1.10+ to reflect the v1.10 feature docs now present."

patterns-established:
  - "D-19 fill-gaps discipline: every new section matches the depth/tone of the existing Phase 28/29 sections; no existing section is moved or rewritten"

requirements-completed:
  - DOC-01
  - DOC-02

# Metrics
duration: 14 min
completed: 2026-08-08
---

# Phase 33 Plan 03: Documentation Gap-Fill (D-19) Summary

**README.md and skill/SKILL.md gain DOC-01 (omittable type with full inference tables) and DOC-02 (templates, include, relative-peer sections) matching the established Phase 28/29 style; SKILL.md also gains a Pipeline Ordering section explaining the load-bearing v1.10 composition pipeline.**

## Performance

- **Duration:** ~14 min
- **Started:** 2026-08-08 (inline execution)
- **Completed:** 2026-08-08
- **Tasks:** 4 (2 README + 2 SKILL)
- **Files modified:** 2

## Accomplishments
- README.md gains 4 sections: Optional Type (Inference) with both inference tables + before/after example, and Templates / Multi-File Composition / Relative Peer Resolution with syntax + example + fixture pointers
- skill/SKILL.md's partial "Type defaults" note expanded into the full DOC-01 spec (both tables + example)
- skill/SKILL.md gains a Pipeline Ordering section documenting the fixed runtime order and why each constraint is load-bearing (XC-02/03/04)
- skill/SKILL.md gains 3 feature schema sections (Templates, Include, Relative Peer Resolution) with full field tables + validation rules citing TMPL/INC/ERGO requirement IDs
- Phase 28 (reference) and Phase 29 (humanization) sections byte-identical in both files (git diff confirms additions only)

## Task Commits

Each task was committed atomically (single docs commit for all 4 sections — they form a coherent doc set):

1. **README Optional Type + Templates + Include + Relative Peer; SKILL Type-inference expansion + Pipeline Ordering + Templates/Include/Relative-Peer schema sections** - `f6bfa2b` (docs)

## Files Created/Modified
- `README.md` — +133 lines: 4 new sections (Optional Type after Optional Name; Templates/Multi-File/Relative-Peer before Full Example)
- `skill/SKILL.md` — +129/-6 lines: Type-defaults note expanded to full spec; Pipeline Ordering section; 3 feature schema sections; Examples list 06-09; version bump

## Decisions Made
- Cited parser functions by NAME (defaultTypeForParent, inferGenericType) not line number. The plan/RESEARCH cited lines 250/276 but the functions are now at 673/699 — line numbers drift, function names don't.
- Bumped SKILL.md version 1.1.0 → 1.2.0 and C4Drill version v1.9+ → v1.10+ to reflect the v1.10 feature docs now present.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Plan's parser.go line numbers for the inference functions are stale**
- **Found during:** Task 1 (DOC-01 README section)
- **Issue:** The plan and RESEARCH.md cited `parser.go:250` (defaultTypeForParent) and `parser.go:276` (inferGenericType). Verified read of the current code shows these functions are now at parser.go:673 and parser.go:699 — the line numbers drifted since the plan was written (Phases 31/32 added code above them).
- **Fix:** Cited the functions by NAME in both README and SKILL footnotes ("Source: `defaultTypeForParent` and `inferGenericType` in `internal/parser/parser.go`"). Verified the RESEARCH tables still match the actual code logic before transcribing.
- **Files modified:** README.md, skill/SKILL.md
- **Verification:** `grep -n 'func defaultTypeForParent\|func inferGenericType' internal/parser/parser.go` returns lines 673 and 699; the transcribed tables match the switch cases in those functions.
- **Committed in:** f6bfa2b.

---

**Total deviations:** 1 auto-fixed (1 stale-line-number bug)
**Impact on plan:** Cosmetic — the doc content is correct (function names are more stable than line numbers anyway). Zero scope creep.

## Issues Encountered
None.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- DOC-01 and DOC-02 are now closed in both user-facing (README) and agent-facing (SKILL) docs.
- Plan 04 (Wave 2) can proceed — its E2E tests do not depend on docs, but the SKILL.md Pipeline Ordering section now documents what those tests enforce.

---
*Phase: 33-docs-sweep-end-to-end-goldens*
*Completed: 2026-08-08*

## Self-Check: PASSED

- [x] `grep -c '### Optional Type' README.md` returns 1
- [x] `grep -c 'componentDb\|containerQueue' README.md` returns ≥ 2 (4 found)
- [x] README section order: Optional Name (164) < Optional Type (188) < Nesting (235)
- [x] `grep -c '### Templates' README.md` = 1; `### Multi-File Composition` = 1; `### Relative Peer` = 1
- [x] README fixture pointers present (06-templates, 08-include, 07-relative-peer)
- [x] `grep -c '### Pipeline Ordering' skill/SKILL.md` returns 1
- [x] `grep -c 'include.Resolve|template.Expand|peer.Resolve' skill/SKILL.md` all ≥ 1 (2/3/3)
- [x] `grep -c '### Templates|### Include|### Relative Peer' skill/SKILL.md` = 3
- [x] `grep -c 'TMPL-0' skill/SKILL.md` ≥ 3 (6 found); `INC-0` ≥ 3 (7 found); `ERGO-0` ≥ 2 (2 found)
- [x] Old "Type defaults" note replaced (renamed to "Type inference rules")
- [x] `git diff README.md skill/SKILL.md` shows additions + the planned Type-defaults expansion only; Phase 28/29 sections byte-identical
- [x] Commit message starts with `docs(33-03):`
