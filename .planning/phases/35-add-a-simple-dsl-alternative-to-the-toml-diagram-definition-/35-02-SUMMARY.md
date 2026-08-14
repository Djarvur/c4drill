---
phase: 35-add-a-simple-dsl-alternative-to-the-toml-diagram-definition
plan: 02
subsystem: template
tags: [toml, parser, templates, use, recursion, cycle-detection, dsl]

# Dependency graph
requires:
  - phase: 31-unit-templates
    provides: Instantiation.Parent attachment (XC-03), Unit.Clone HS-1 deep copy, ExpandError, pathTracker
  - phase: 32-include-directive-multi-file
    provides: cycle-detection + maxIncludeDepth pattern (internal/include/resolve.go)
  - phase: 35-add-a-simple-dsl-alternative-to-the-toml-diagram-definition (Plan 01)
    provides: none directly — this plan lands the TOML-side halves of D-16/D-17 the C4D front-end desugars onto later
provides:
  - "[[unit.<name>.use]] TOML sugar desugaring to Instantiation{Parent: enclosing unit path} (D-16)"
  - TemplateDef.Instantiations — [[template.<name>.<path>.use]] body uses with parents relative to the template unit root (D-17)
  - Recursive Expand: outer-to-inner param flow, ancestor-stack cycle detection ("A -> B -> A"), maxTemplateDepth=100 depth cap, HS-1 at every level
  - claimSubtree — every produced subtree's descendant paths claimed (TMPL-07 closes a pre-existing silent-overwrite gap)
affects: [35-03 reserved-word errors, 35-04+ tomodel (C4D use-in-block desugars to these Instantiation forms), 35-07 fmt, README/skill docs for nested use]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "UseSite document-order capture: captureDefinitionOrder admits ArrayTable expressions ONLY for use-terminated keys, recording sites the raw-map extraction pass then pairs with array elements via per-path cursors (Go maps lose authoring order)"
    - "Recursion guard split (checkRecursion): ancestor-stack slices.Contains cycle check naming the chain + depth cap, copied verbatim from the include pattern"
    - "Base-path parent resolution: document-level Parents are model-absolute, template-body-use Parents are clone-root-relative — one joinPath at expandInstantiation entry covers both"

key-files:
  created: []
  modified:
    - internal/parser/parser.go
    - internal/parser/parser_test.go
    - internal/template/expand.go
    - internal/template/expand_test.go

key-decisions:
  - "UseSite + per-path cursor pairing: each [[...use]] header is one unstable-API expression and one array element; sites recorded in document order interleave top-level [[use]], [[unit.X.use]], and [[template.X.use]] correctly where raw-map walking could not"
  - "An explicit `parent` key inside a nested [[unit.X.use]] / [[template.X.use]] resolves RELATIVE to the site path (joinPath), keeping the top-level form's absolute-parent semantics byte-identical"
  - "claimSubtree claims every descendant path of each attached clone: template-declared subunits were never claimed before, so a later instantiation could silently overwrite them — pre-existing TMPL-07 gap, closed for both top-level and nested forms"
  - "Template-declared unit types stay fixed at parse time (C1-level inference in the template namespace): authors of nested templates write final level-specific types (containerBox/containerDb), matching Phase 31 semantics"
  - "\"Unknown key is a hard ParseError\" bullet implemented as non-string-value rejection: template params are free-form string keys, so key-name rejection is unimplementable; the referenced strictness (extractInstantiations') is exactly the non-string-value rule"

patterns-established:
  - "Narrow ArrayTable admission in unstable-API order capture (last-key-segment match) instead of un-filtering [[...]] wholesale — [[a.link]] etc. stay skipped"

requirements-completed: [D-16, D-17]

# Metrics
duration: 21min
completed: 2026-08-14
---

# Phase 35 Plan 02: Nested use + Template-Body use Summary

**[[unit.X.use]] TOML sugar (D-16) and recursive [[template.X.use]] body expansion (D-17) — outer-to-inner param flow, ancestor-stack cycle detection naming the chain, a 100-level depth cap, and subtree-claiming TMPL-07 collision checks, with HS-1 deep-copy isolation holding at every recursion level**

## Performance

- **Duration:** 21 min (15:48–16:10 UTC)
- **Started:** 2026-08-14T15:48:54Z
- **Completed:** 2026-08-14T16:10:12Z
- **Tasks:** 2/2
- **Files modified:** 4 (0 created, 4 modified)

## Accomplishments

- D-16: `[[unit.<name>.use]]` array-of-tables under any unit section (at any nesting depth) desugars to the exact `Instantiation{Template, Parent: enclosing dotted path, Params}` the top-level `[[use]] parent=` form produces — verified by a deep-equal test across both authoring forms, interleaved in document order with top-level entries
- D-17 (parses): `[[template.<name>.<path>.use]]` entries land in `TemplateDef.Instantiations` with parents relative to the template's unit root; an explicit `parent` key resolves relative to the site path
- D-17 (expands): `Expand` recurses through template bodies after the outer clone attaches — params flow outer-to-inner through the outer replacer, produced units attach inside the clone's subtree via the existing attach machinery, and the v1.10 template-nesting deferral is functionally lifted
- Cycle detection copies the include pattern: `slices.Contains(stack, name)` → `ExpandError{Kind: "cycle"}` naming the chain (`A -> B -> A`, self-loops `A -> A`); `maxTemplateDepth = 100` mirrors `maxIncludeDepth` and turns a 150-deep acyclic chain into a hard error (no panic, no hang)
- TMPL-06/TMPL-07 exit gates preserved: residual-`${` scan fires post-recursion on nested bodies; `claimSubtree` claims every descendant path of each attached clone so nested uses cannot silently overwrite template-declared or hand-authored units
- HS-1 at every level: 3-level nesting × 3 instantiations with links yields disjoint validator-synthesized `LinksFrom` mirrors (one per leaf), idempotent on re-expand

## Task Commits

Each task was committed atomically (TDD RED → GREEN):

1. **Task 1: [[unit.X.use]] sugar + [[template.X.use]] body extraction** - `5fec94c` (test, RED) / `ec2a59c` (feat, GREEN)
2. **Task 2: Recursive template-body expansion + cycle detection** - `f08f788` (test, RED) / `20d63b7` (feat, GREEN)

## TDD Gate Compliance

Both behavior-adding tasks followed RED→GREEN: `test(35-02)` commits (5fec94c, f08f788) precede their `feat(35-02)` GREEN commits (ec2a59c, 20d63b7) in git order. RED runs captured: Task 1 compile failure (`TemplateDef has no field Instantiations`) plus behavioral failures; Task 2 all 8 new tests failing (body uses silently ignored) with every pre-existing test green.

## Files Created/Modified

- `internal/parser/parser.go` - TemplateDef.Instantiations field; UseSite type; captureDefinitionOrder records use sites in document order (narrow ArrayTable admission); extractUses/extractSiteUses/parseUseEntries/lookupUseArray helpers; isBuiltinField reserves "use"
- `internal/parser/parser_test.go` - 5 new test functions (12 cases): sugar-equivalence deep-equal, nested paths, document-order interleave, template-body extraction, non-string-value rejection
- `internal/template/expand.go` - checkRecursion (cycle + maxTemplateDepth), expandBodyUses + substituteInstantiation (outer-to-inner flow), claimSubtree + joinPath, basePath-relative parent resolution
- `internal/template/expand_test.go` - 8 new test functions: param flow, 3-level chain, mutual + self cycles, depth cap, residual token, nested collisions (2 variants), 3-level HS-1 with idempotency

## Decisions Made

- UseSite + per-path cursor pairing for document-order interleaving (Go maps lose order; each `[[...use]]` header pairs positionally with its array element)
- Nested explicit `parent` keys resolve relative to the site path; top-level parents stay absolute (Phase 31 semantics byte-identical)
- claimSubtree added (Rule 2): without it, template-declared subunit paths were unclaimed and silently overwritable — a pre-existing gap the nested forms would have amplified
- Nested template authors write final level-specific types (parse-time C1 inference is fixed); documented via the HS-1 fixture using containerBox/containerDb
- "Unknown key" strictness bullet implemented as non-string-value rejection (the literal reading is unimplementable — params ARE free-form keys)

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Per-path cursor pairing in extractUses**
- **Found during:** Task 1 GREEN
- **Issue:** Each site initially re-read the full array at its path, duplicating every entry (2 [[use]] blocks → 4 Instantiations, caught by TestParseUseArrayPreservesOrder)
- **Fix:** The Nth site with a given path pairs with the Nth element of that path's array (every [[...use]] header appends exactly one element); cursors map keyed by site context
- **Files modified:** internal/parser/parser.go
- **Verification:** TestParseUseArrayPreservesOrder + TestParseUnitUseInterleavesDocumentOrder
- **Committed in:** ec2a59c

**2. [Rule 2 - Missing critical] claimSubtree closes a pre-existing TMPL-07 gap**
- **Found during:** Task 2 design (threat T-35-02-02 analysis)
- **Issue:** Only produced ROOT paths were ever claimed — a template-declared subunit (e.g. [template.X.api]) was unclaimed, so a later instantiation (top-level [[use]] parent=... or a template-body use) could silently overwrite it; the plan's collision behavior bullet demands nested uses collide loudly with authored units
- **Fix:** attachProduced claims every descendant path of the freshly attached clone (deterministic SubunitOrder walk)
- **Files modified:** internal/template/expand.go
- **Verification:** TestExpandNestedPathCollision (both variants)
- **Committed in:** 20d63b7

**3. [Rule 1 - Bug] "Unknown key" behavior bullet reinterpreted to the referenced strictness**
- **Found during:** Task 1 test authoring
- **Issue:** The bullet asked for a hard ParseError on "unknown key other than template/parent/params", but template params are free-form string keys — rejecting key names would reject every param; the parenthetical says to mirror extractInstantiations, whose strictness is non-string-value rejection
- **Fix:** Tests + implementation reject non-string VALUES in all three use forms (*parser.ParseError naming the key); any string key is a param
- **Files modified:** internal/parser/parser_test.go, internal/parser/parser.go
- **Committed in:** 5fec94c, ec2a59c

**4. [Rule 3 - Blocking] ArrayTable admission + unstable.Node signature**
- **Found during:** Task 1 implementation
- **Issue:** [[use]]-family headers are unstable.ArrayTable expressions, Kind-filtered by captureDefinitionOrder — sites were invisible; un-filtering ArrayTable wholesale would have recorded [[a.link]] as phantom subunits (catastrophic regression); the helper's expr parameter is *unstable.Node, not an Expression type
- **Fix:** isTableExpression admits both kinds, but the use-site check (last key segment == "use") runs before the Table-only dispatch so every other array-of-tables stays skipped exactly as before
- **Files modified:** internal/parser/parser.go
- **Verification:** full repo suite 13/13 green (links/valid/nested fixtures unchanged)
- **Committed in:** ec2a59c

**5. [Rule 1 - Bug] HS-1 test fixture types corrected (box → containerBox, db → containerDb)**
- **Found during:** Task 2 GREEN
- **Issue:** Template-declared types are fixed at parse time (C1-level inference in the template namespace); a bare `box` clone landing inside a system fails the validator's C2 rule
- **Fix:** Test fixture authors final level-specific types — the semantics nested-template authors must use, now pinned by a validating test
- **Files modified:** internal/template/expand_test.go
- **Verification:** TestExpandThreeLevelNestingHS1 validates cleanly with 3 disjoint mirrors
- **Committed in:** 20d63b7

---

**Total deviations:** 5 auto-fixed (2 bug, 1 missing critical, 1 blocking, 1 test-side bug)
**Impact on plan:** All acceptance criteria met; no scope creep. The only semantic addition beyond the plan text is claimSubtree (strictness hardening consistent with the project's hard-error-everywhere stance).

## Threat Model Disposition

| Threat | Disposition | Where |
|--------|-------------|-------|
| T-35-02-01 (DoS via recursion depth) | mitigated | checkRecursion: cycle stack + maxTemplateDepth=100; TestExpandTemplateCycle/SelfCycle/DepthCap |
| T-35-02-02 (tampering via path overwrite) | mitigated | claim during attach (claimSubtree) + post-loop TMPL-06/TMPL-07 exit gates kept; TestExpandNestedPathCollision, TestExpandNestedResidualToken |
| T-35-02-03 (repudiation) | accepted | ExpandError Kind/Site/Detail names instantiation index + chain |

## Issues Encountered

- Initial extractUses duplicated entries across repeated sites at the same path (cursor pairing fix, deviation 1)
- gocognit (15) and funlen (60) gates forced helper extraction (extractSiteUses/appendTemplateUse, recordUnitTable, checkRecursion) — no nolint needed
- The depth-cap test generates a 150-template TOML document programmatically (fmt.Fprintf builder) rather than a fixture — self-contained, no testdata additions

## Known Stubs

None - both tasks landed complete implementations. (The `ast.UseStmt`/`UnitNode.UseStmts` placeholders from Plan 01 remain intentionally deferred to the C4D-side plans, per the phase sequencing.)

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- TOML authors can nest use under unit sections and inside template bodies; both normalize to the single Instantiation mechanism, so the C4D `use name(args)`-in-block desugar (later plans) targets `Instantiation{Parent: enclosingPath}` with zero further Expand work
- `go test ./...` 13/13 green, golangci-lint 0 issues; TestXC01_PipelineOrdering and all v1.10 template/include/peer tests untouched and passing
- No blockers

## Self-Check: PASSED

All 4 modified files exist on disk; all 4 task commit hashes (5fec94c, ec2a59c, f08f788, 20d63b7) verified in git log.

---
*Phase: 35-add-a-simple-dsl-alternative-to-the-toml-diagram-definition*
*Completed: 2026-08-14*
