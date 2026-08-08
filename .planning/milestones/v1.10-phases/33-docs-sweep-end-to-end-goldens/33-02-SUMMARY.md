---
phase: 33-docs-sweep-end-to-end-goldens
plan: 02
subsystem: docs
tags: [toml, fixtures, examples, doc-03, xc-05]

# Dependency graph
requires:
  - phase: 31-template-expansion
    provides: "[template.*] + [[use]] grammar (TMPL-01..10)"
  - phase: 32-include-directive-multi-file
    provides: "[[include]] directive + cross-file subunit merge (INC-01..10, D-10)"
  - phase: 30-relative-peer-resolution
    provides: bare-peer walk-up (ERGO-01/02, D-13..D-16)
provides:
  - "Four runnable example-fixture sets (06-templates, 07-relative-peer, 08-include, 09-composed) — DOC-03"
  - "Composed XC-05 golden pair (entry.toml multi-file + single-file-equivalent.toml hand-expanded) — input for Plan 04"
affects: [33-03, 33-04]

# Tech tracking
tech-stack:
  added: []
  patterns: ["Validator-aware fixture design: split templates that need BOTH links and subunits (validator forbids parents-with-subunits from carrying direct links)", "Cross-file subunit re-declaration: included file re-declares the parent so the parser recognizes its subunits (D-10)"]

key-files:
  created:
    - skill/examples/06-templates.toml
    - skill/examples/07-relative-peer.toml
    - skill/examples/08-include/entry.toml
    - skill/examples/08-include/auth.toml
    - skill/examples/08-include/billing.toml
    - skill/examples/09-composed/entry.toml
    - skill/examples/09-composed/templates.toml
    - skill/examples/09-composed/domains/auth.toml
    - skill/examples/09-composed/single-file-equivalent.toml
  modified: []

key-decisions:
  - "Split the 06 microservice template into two: a LEAF template (microservice, with parametrized link + reference) and a PARENT template (dataService, with cache subunit). The validator forbids parents-with-subunits from carrying direct links, so a single template with both would fail validation after expansion. Both TMPL-03/10 (link+reference) and TMPL-04 (subtree) are still demonstrated."
  - "09-composed auditLog subunit declared BEFORE auth in single-file-equivalent.toml. Subunit append order is pipeline-load-bearing: include (pass 1) appends auditLog before template.Expand (pass 2) appends auth, so the single file must match that order for byte-stable canonical equality."
  - "07 Case 4 absolute-fallback uses peer='platform.api.handlers' (a real dotted path that exists) rather than 'externalGateway.gateway' (which would need an external-leaf-with-subunit shape that systemExternal cannot host). Demonstrates the D-16 dot-gate identically."

patterns-established:
  - "skill/examples/ as the home for runnable tutorial fixtures (per-feature standalones + one composed XC-05 set)"
  - "Composed XC-05 fixture ships BOTH multi-file and hand-expanded single-file forms so Plan 04 compares two real committed artifacts, not synthetics"

requirements-completed:
  - DOC-03

# Metrics
duration: 22 min
completed: 2026-08-08
---

# Phase 33 Plan 02: Example Fixtures (D-17) Summary

**Nine runnable TOML fixtures across four sets (06-templates, 07-relative-peer, 08-include, 09-composed) demonstrating all four v1.10 features; the composed set ships a hand-expanded single-file equivalent verified to canonicalize identical to its multi-file entry.**

## Performance

- **Duration:** ~22 min
- **Started:** 2026-08-08 (inline execution)
- **Completed:** 2026-08-08
- **Tasks:** 4 (one per fixture set)
- **Files created:** 9

## Accomplishments
- All four fixture sets from D-17 ship and render cleanly through the full v1.10 pipeline (`ParseFile → include.Resolve → template.Expand → peer.Resolve → validate → views → render`)
- 06-templates demonstrates TMPL-01..10 (define, instantiate, `${param}` substitution into name/description/technology/reference, fixed link count, subunit subtree, reference parametrization)
- 07-relative-peer demonstrates all 4 bare-peer walk-up cases (sibling, aunt, root, absolute-fallback) with inline `Case N` labels
- 08-include demonstrates INC-01..10 (multi-file, once=true dedup, cross-file subunits, relative paths)
- 09-composed uses all four features together AND ships the hand-expanded single-file-equivalent
- **XC-05 pre-verified:** a throwaway test confirmed `canonical.Canonical(render(entry.toml)) == canonical.Canonical(render(single-file-equivalent.toml))` (both 3872 bytes, identical) — the assertion lands permanently in Plan 04

## Task Commits

Each task was committed atomically (one commit for all 9 fixtures — they form a coherent doc set):

1. **All 4 fixture sets** - `122d3e0` (docs) — single commit because the 9 files are interdependent (09 reuses grammar from 06/08; the README/SKILL pointers in Plan 03 reference all of them)

## Files Created
- `skill/examples/06-templates.toml` — standalone templates tutorial (2 templates, 3 instantiations)
- `skill/examples/07-relative-peer.toml` — standalone relative-peer tutorial (4 labeled cases)
- `skill/examples/08-include/entry.toml` + `auth.toml` + `billing.toml` — 3-file include tutorial
- `skill/examples/09-composed/entry.toml` + `templates.toml` + `domains/auth.toml` + `single-file-equivalent.toml` — composed XC-05 golden pair

## Decisions Made
- Split 06's template into two (microservice leaf + dataService parent) to satisfy the validator's "parents with subunits cannot carry direct links" rule while still demonstrating both TMPL-03/10 (link/reference) and TMPL-04 (subtree).
- 09 composed auditLog ordered before auth in single-file-equivalent (matches pipeline append order: include before Expand).
- 07 Case 4 uses a real existing dotted path (`platform.api.handlers`) for the absolute-fallback peer rather than fabricating an external-tree shape incompatible with systemExternal.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Plan's 06 template shape (single template with BOTH link and subunit) violates the validator**
- **Found during:** Task 1 (06-templates render check)
- **Issue:** The plan's sketch had `[template.microservice]` with both a `[[.link]]` AND a `[.cache]` subunit. After expansion, the produced parent unit carries a direct link while having a subunit — the validator rejects this (`unit "X" has subunits and cannot have direct links`).
- **Fix:** Split into two templates: `microservice` (leaf, parametrized link + reference) and `dataService` (parent, cache subunit whose own link keeps the parent non-orphan transitively). Both TMPL-03/10 and TMPL-04 are still covered.
- **Files modified:** skill/examples/06-templates.toml
- **Verification:** `go run ./cmd/c4drill skill/examples/06-templates.toml -f dot` exits 0.
- **Committed in:** fixtures commit.

**2. [Rule 3 - Blocking] Multiple validator type-system constraints surfaced during fixture authoring**
- **Found during:** Tasks 1-4 (render checks for each fixture)
- **Issue:** Initial fixture drafts tripped: (a) systemExternal cannot host subunits; (b) containerBox requires C2 children (not C3 components); (c) every leaf needs ≥1 link (orphan rule); (d) customer cannot link directly to a parent-with-subunits.
- **Fix:** Restructured each fixture to respect the C1/C2/C3 nesting rules and the leaf-connectivity rule. Specifics: 07's auth moved from containerBox-child to a direct C3 component under platform.api; 08's customer link retargeted from `platform.auth` (parent) to `platform.auth.db` (leaf); each fixture's leaf subunits given a link to messageBus.
- **Files modified:** skill/examples/07-relative-peer.toml, 08-include/entry.toml
- **Verification:** all 5 runnable fixtures (9 files) render through the full pipeline with exit 0.
- **Committed in:** fixtures commit.

**3. [Rule 3 - Blocking] 09-composed subunit append order is pipeline-load-bearing**
- **Found during:** Task 4 (single-file-equivalent authoring + equivalence check)
- **Issue:** A naive single-file-equivalent declaring `auth` before `auditLog` under `[platform]` produced a different canonical form than entry.toml (whose pipeline appends auditLog via include pass 1 before auth via Expand pass 2).
- **Fix:** Declared `auditLog` before `auth` in single-file-equivalent.toml, matching the pipeline's append order. Verified byte-stable canonical equality (3872 bytes both sides).
- **Files modified:** skill/examples/09-composed/single-file-equivalent.toml
- **Verification:** throwaway `TestTmpEquivalenceCheck` confirmed `canonical.Canonical(entry) == canonical.Canonical(single)` (test deleted after confirmation; permanent assertion lands in Plan 04).
- **Committed in:** fixtures commit.

---

**Total deviations:** 3 auto-fixed (1 plan-shape bug, 2 blocking validator constraints)
**Impact on plan:** All auto-fixes necessary for the fixtures to render through the real pipeline (the plan's mandate that they be "runnable"). Zero scope creep — every requirement ID (TMPL/INC/ERGO) is still demonstrated. The XC-05 composition proof is preserved.

## Issues Encountered
- The plan's fixture sketches were written before Phases 30-32 shipped, so they predated hands-on validator feedback. Discovering the constraints during authoring was expected (the plan's `read_first` lists the prior-phase CONTEXTs but not the validator rules); resolving them inline was faster than re-planning.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Plan 03 (README + SKILL.md) can point at the new fixtures: 06-templates, 07-relative-peer, 08-include, 09-composed.
- Plan 04 (Wave 2) consumes 09-composed/{entry,single-file-equivalent}.toml as the XC-05 golden pair and reuses entry.toml for the XC-01 behavioral test. XC-05 equivalence already pre-verified.

---
*Phase: 33-docs-sweep-end-to-end-goldens*
*Completed: 2026-08-08*

## Self-Check: PASSED

- [x] All 4 fixture sets exist (`ls skill/examples/06-templates.toml 07-relative-peer.toml 08-include/ 09-composed/`)
- [x] 9 TOML files total (`find ... -name '*.toml' | wc -l` = 9)
- [x] 06: `grep -c '[template\.' ` ≥ 2; `grep -c '[[use]]'` ≥ 2; `${` ≥ 3; `reference = "https` ≥ 1
- [x] 07: 4 `Case N` labels; peer=cache, peer=queue, peer=messageBus, dotted peer all present
- [x] 08: ≥2 `[[include]]`; `once = true`; `[platform.auth]` in auth.toml; `[platform.billing]` in billing.toml
- [x] 09: 2 `[[include]]` in entry; ≥1 `[[use]]`; `once=true`; ≥2 `[template.`; `[platform.auditLog]`; single-file has 0 directives + 0 real `${`
- [x] Every unit has explicit `type =`
- [x] All include paths relative
- [x] All 5 runnable fixtures render through full pipeline (exit 0)
- [x] XC-05 equivalence pre-verified (canonical forms identical, 3872 bytes both sides)
- [x] Commit message starts with `docs(33-02):`
