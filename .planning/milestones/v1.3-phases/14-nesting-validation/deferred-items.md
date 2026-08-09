# Deferred Items - Phase 14

## Pre-existing Issues (Out of Scope)

### TestOutputFlag Test Failure
- **File:** `cmd/c4drill/root_test.go:96`
- **Issue:** Test expects output flag default value to be `"."` but actual value is `""` (empty string)
- **Impact:** Low - functionality works correctly, only test expectation is wrong
- **Recommendation:** Either fix test to expect `""` or fix root.go to default to `"."`
- **Discovered during:** Task 3 verification (2026-03-18)
- **Status:** Pre-existing (not caused by 14-01 changes)
