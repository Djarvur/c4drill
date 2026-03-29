---
phase: 10-link-list-format
plan: "01"
subsystem: model
tags: [toml, parser, links, slices]

# Dependency graph
requires:
  - phase: 09-no-orphan-units
    provides: validation rules that expect proper link structures
provides:
  - Slice-based Link structures (Links/LinksFrom as []Link instead of map[string]Link)
  - Explicit Peer field on Link struct
  - TOML array syntax support [[link]]
  - Multiple links to same peer capability
affects: [10-add-parent-link, 10-expand-defaults, future features requiring multiple links per peer]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - Slice iteration pattern for links (for _, link := range unit.Links)
    - FindLinkByPeer helper function for slice-based link lookup
    - TOML array syntax [[unit.link]] with peer field

key-files:
  created: []
  modified:
    - internal/model/link.go - Added FindLinkByPeer helper function
    - internal/model/unit.go - Links and LinksFrom changed to []Link slices
    - internal/parser/parser.go - Removed populateLinkTargets function
    - internal/parser/parser_test.go - Updated for slice-based links
    - internal/graph/builder.go - Updated to iterate over slices
    - internal/validator/rules.go - Updated to iterate over slices
    - internal/view/scope.go - Updated to iterate over slices
    - testdata/links.toml - Updated to array syntax
    - testdata/invalid_links.toml - Updated to array syntax
    - testdata/invalid_references.toml - Updated to array syntax
    - cmd/c4drill/testdata/expanded.toml - Added links for validation
    - cmd/c4drill/testdata/valid.toml - Added links for validation

key-decisions:
  - Use slice []Link instead of map[string]Link to enable multiple links to same peer
  - Add explicit Peer field instead of deriving from map key
  - TOML syntax changes to [[unit.link]] array format with peer field
  - Add FindLinkByPeer helper for slice-based lookups

patterns-established:
  - Pattern 1: Always iterate over Links/LinksFrom slices using "for _, link := range"
  - Pattern 2: Use FindLinkByPeer() helper to find links by peer name
  - Pattern 3: TOML test fixtures use [[unit.link]] syntax with peer field
  - Pattern 4: Link.Peer field accessed directly, not derived from map key

requirements-completed: [LLIST-01, LLIST-02, LLIST-03]

# Metrics
duration: 4 min
completed: 2026-03-11T19:57:17Z
---

# Phase 10 Plan 01: Link List Format Summary

**Changed Link struct from map-based Target field to explicit Peer field, and Links/LinksFrom from maps to slices, enabling multiple links to the same peer with simpler TOML [[link]] array syntax.**

## Performance

- **Duration:** 4 min
- **Started:** 2026-03-11T19:53:14Z
- **Completed:** 2026-03-11T19:57:17Z
- **Tasks:** 4
- **Files modified:** 20

## Accomplishments

- Updated Link struct to use explicit Peer field with `toml:"peer"` tag instead of derived Target
- Changed Unit.Links and Unit.LinksFrom from `map[string]Link` to `[]Link` slices
- Removed populateLinkTargets function from parser (go-toml handles array unmarshaling)
- Updated all testdata files to use new [[link]] array syntax
- Updated all test files to iterate over slices and use link.Peer
- Added FindLinkByPeer helper function for slice-based lookups
- Cascaded slice-based changes through graph builder, validator, and view scope

## Task Commits

Each task was committed atomically:

1. **Task 1: Update Link and Unit structs** - `552ed4c` (feat) - Previously committed
2. **Task 2: Update parser and remove populateLinkTargets** - `08c8cd7` (feat) - Previously committed
3. **Task 3: Update testdata TOML files** - `0611eb9` (feat) - Add FindLinkByPeer helper and update testdata files
4. **Task 4: Update parser tests** - `2e5a95b` (test) - Update parser tests for slice-based links
5. **Additional: Graph builder updates** - `533088a` (feat) - Update graph builder for slice-based links
6. **Additional: Validator updates** - `00271d2` (feat) - Update validator for slice-based links
7. **Additional: View scope updates** - `3d309dc` (feat) - Update view scope for slice-based links
8. **Additional: Render test updates** - `f758cef` (test) - Update render tests for slice-based links
9. **Additional: CLI test updates** - `dcb5782` (test) - Update CLI tests for slice-based links

**Plan metadata:** (will be created after SUMMARY)

_Note: Tasks 1 and 2 were already committed from a previous execution. Tasks 3-4 and all cascading changes were committed during this execution._

## Files Created/Modified

- `internal/model/link.go` - Added FindLinkByPeer helper function for finding links by peer name
- `internal/model/unit.go` - Links and LinksFrom changed from map[string]Link to []Link slices
- `internal/parser/parser.go` - Removed populateLinkTargets function (no longer needed)
- `internal/parser/parser_test.go` - Updated all assertions to use link.Peer and slice iteration
- `internal/graph/builder.go` - Updated processOutgoingLinks/processIncomingLinks to iterate over slices
- `internal/graph/integration_test.go` - Fixed peer reference in multiple links test
- `internal/validator/rules.go` - Updated ValidateReferences/ValidateLinkRules to iterate over slices
- `internal/validator/rules_test.go` - Updated all test fixtures to use []Link syntax
- `internal/validator/validator_test.go` - Updated all test fixtures to use []Link syntax
- `internal/view/scope.go` - Updated addExternalBoundaryNodes* to iterate over slices
- `internal/view/scope_test.go` - Updated test fixtures
- `internal/view/integration_test.go` - Updated test fixtures
- `internal/render/integration_test.go` - Updated buildTestModelWithLinks to use slice syntax
- `cmd/c4drill/root_test.go` - Updated inline TOML test fixtures to use [[link]] syntax
- `cmd/c4drill/testdata/expanded.toml` - Added links to fix orphan unit validation
- `cmd/c4drill/testdata/valid.toml` - Added links to fix orphan unit validation
- `testdata/links.toml` - Updated to [[webapp.link]] syntax with peer field
- `testdata/invalid_links.toml` - Updated to [[parent.link]] syntax with peer field
- `testdata/invalid_references.toml` - Updated to [[user.link]] syntax with peer field

## Decisions Made

- Use slice []Link instead of map[string]Link to enable multiple links to same peer
- Add explicit Peer field instead of deriving from map key for clarity
- TOML syntax changes to [[unit.link]] array format for consistency
- Add FindLinkByPeer helper function to make slice-based lookups idiomatic

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Added FindLinkByPeer helper function**
- **Found during:** Task 4 (Update parser tests)
- **Issue:** Tests needed to find links by peer name in slice-based structure, no helper existed
- **Fix:** Added FindLinkByPeer function to internal/model/link.go that searches a slice of links by peer name and returns the link and a boolean
- **Files modified:** internal/model/link.go
- **Verification:** All parser tests now use FindLinkByPeer to find links and pass
- **Committed in:** 0611eb9 (Task 3 commit)

**2. [Rule 3 - Blocking] Updated internal/graph/builder.go for slice iteration**
- **Issue:** Graph builder still used map iteration pattern for links, causing compilation errors
- **Fix:** Updated processOutgoingLinks and processIncomingLinks to iterate over []Link slices and use link.Peer instead of map keys
- **Files modified:** internal/graph/builder.go, internal/graph/builder_test.go, internal/graph/integration_test.go
- **Verification:** All graph tests pass
- **Committed in:** 533088a (Graph builder commit)

**3. [Rule 3 - Blocking] Updated internal/validator for slice iteration**
- **Issue:** Validator rules still used map iteration pattern for links, causing compilation errors
- **Fix:** Updated ValidateReferences and ValidateLinkRules to iterate over []Link slices and use link.Peer
- **Files modified:** internal/validator/rules.go, internal/validator/rules_test.go, internal/validator/validator_test.go
- **Verification:** All validator tests pass
- **Committed in:** 00271d2 (Validator commit)

**4. [Rule 3 - Blocking] Updated internal/view for slice iteration**
- **Issue:** View scope still used map iteration pattern for links, causing compilation errors
- **Fix:** Updated addExternalBoundaryNodesRecursive and addExternalBoundaryNodesForSubunits to iterate over []Link slices and use link.Peer
- **Files modified:** internal/view/scope.go, internal/view/scope_test.go, internal/view/integration_test.go
- **Verification:** All view tests pass
- **Committed in:** 3d309dc (View scope commit)

**5. [Rule 3 - Blocking] Updated internal/render tests**
- **Issue:** Render integration tests still used map syntax for Links in test fixtures
- **Fix:** Updated buildTestModelWithLinks to use []Link slice syntax
- **Files modified:** internal/render/integration_test.go
- **Verification:** All render tests pass
- **Committed in:** f758cef (Render test commit)

**6. [Rule 3 - Blocking] Updated cmd/c4drill tests**
- **Issue:** CLI tests had orphan units failing validation and used old map syntax
- **Fix:** Updated all inline TOML test fixtures and testdata files to use [[link]] syntax with proper links
- **Files modified:** cmd/c4drill/root_test.go, cmd/c4drill/testdata/expanded.toml, cmd/c4drill/testdata/valid.toml
- **Verification:** All CLI tests pass
- **Committed in:** dcb5782 (CLI test commit)

**7. [Rule 3 - Blocking] Fixed TestIntegrationMultipleLinksBetweenSameUnits peer reference**
- **Found during:** Task 4 (Update parser tests)
- **Issue:** Test had wrong peer reference (app instead of db) in LinksFrom, causing assertion failure
- **Fix:** Changed LinksFrom peer from "app" to "db" to match expected db->app edge
- **Files modified:** internal/graph/integration_test.go
- **Verification:** TestIntegrationMultipleLinksBetweenSameUnits now passes
- **Committed in:** 533088a (Graph builder commit)

---

**Total deviations:** 7 auto-fixed (7 blocking issues)
**Impact on plan:** All auto-fixes were necessary cascading changes from the model restructuring. The changes to graph builder, validator, view scope, and test files were required for the codebase to compile and tests to pass after changing Link/Unit structs. No scope creep - all changes directly related to slice-based link migration.

## Issues Encountered

- Build errors in multiple packages due to map-to-slice migration affecting all code that accessed Links/LinksFrom
- Test failures due to orphan units in test fixtures (units with no Links, LinksFrom, or Subunits) - fixed by adding proper links to testdata files
- Test failures due to validation rule changes (orphan unit detection) - fixed by updating test fixtures

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Model struct changes complete and tested
- Parser updated to handle array syntax
- All tests passing with new structure
- Ready for plan 10-02 (add parent link field) which will build on the new slice-based link structure

---
*Phase: 10-link-list-format*
*Completed: 2026-03-11*

## Self-Check: PASSED

- [x] Key files created/modified exist on disk
  - [x] internal/model/link.go - Found
  - [x] internal/parser/parser.go - Found
  - [x] testdata/links.toml - Found
  - [x] .planning/phases/10-link-list-format/10-01-SUMMARY.md - Found
- [x] Commits exist (12 commits with "10-01" in message)
- [x] All tests pass (go test ./...)
