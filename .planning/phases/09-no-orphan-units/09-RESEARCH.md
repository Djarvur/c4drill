# Phase 9: No Orphan Units - Research

**Researched:** 2026-03-11
**Domain:** Go validation rule implementation in existing validator package
**Confidence:** HIGH

## Summary

This phase adds a new validation rule `ValidateOrphanUnits` to the existing validator package. The implementation follows established patterns from existing rules (`ValidateReferences`, `ValidateSubunitRules`, `ValidateLinkRules`). The orphan detection algorithm is straightforward: iterate over the index, check each unit for connectivity (Links or LinksFrom) or children (Subunits), and report units with none of these as orphans.

**Primary recommendation:** Follow the existing rule function pattern exactly. Add `ValidateOrphanUnits` to `internal/validator/rules.go`, call it from `Validate()` in `validator.go`, and add unit tests to `rules_test.go`.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**Error reporting format:**
- One ValidationError per orphan unit (consistent with existing rules like ValidateReferences)
- Standard message format: 'unit "path" has no incoming or outgoing links'
- Report all orphans found (not fail-fast, consistent with existing validation pattern)
- Use existing ValidationError struct with Path field (no new error type needed)

**Orphan definition:**
- Unit is orphan if it has NO incoming AND NO outgoing links AND NO subunits
- Either direction counts for connectivity (incoming OR outgoing = connected)
- Both Links and LinksFrom fields count toward connectivity
- Only explicitly defined units can be orphans (external boundary nodes are inferred, not checked)

**Unit type exceptions:**
- No unit type exemptions (all types must have links if they have no subunits)
- Units with subunits are exempt (they have children, not orphans)
- Single unit model is invalid (can't link to anything, will be caught as orphan)
- Empty model is invalid (no architecture defined)

**External boundary nodes:**
- External boundary nodes are NOT checked for orphan status
- Links to/from undefined units are NOT allowed (caught by ValidateReferences rule first)
- Connectivity checking considers only explicit TOML content
- Only links to defined units count as connectivity

### Claude's Discretion
- Exact implementation of orphan detection algorithm
- How to efficiently check connectivity using the existing index
- Order of validation rules (orphan check should run after reference validation)
- Internal helper function design

### Deferred Ideas (OUT OF SCOPE)
None — discussion stayed within phase scope.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| VAL-01 | Validator rejects TOML files with unlinked (orphan) units - all units must have at least one incoming or outgoing link | ValidateOrphanUnits rule checks each unit for Links/LinksFrom/Subunits |
| VAL-02 | Validation error message clearly identifies which units are unlinked | Use ValidationError with Path field; format: 'unit "path" has no incoming or outgoing links' |
</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go | 1.26.1 | Primary language | Project requirement (go.mod) |
| stretchr/testify | v1.11.1 | Test assertions | Already in use in rules_test.go |

### Supporting
| Library | Purpose | When to Use |
|---------|---------|-------------|
| (none) | N/A | All needed infrastructure exists in validator package |

**Installation:** No new dependencies required.

## Architecture Patterns

### Recommended File Changes
```
internal/validator/
├── rules.go           # Add ValidateOrphanUnits() function
├── rules_test.go      # Add unit tests for orphan detection
└── validator.go       # Add call to ValidateOrphanUnits in Validate()
```

### Pattern 1: Rule Function Signature
**What:** All validation rules follow the same signature and error collection pattern.
**When to use:** Always - this is the established convention.
**Example:**
```go
// From rules.go:14-45 - ValidateReferences pattern
func ValidateOrphanUnits(index map[string]*UnitInfo) ValidationErrors {
    var errors ValidationErrors

    for path, info := range index {
        // Check if unit has connectivity
        hasLinks := len(info.Unit.Links) > 0
        hasLinksFrom := len(info.Unit.LinksFrom) > 0
        hasSubunits := len(info.Unit.Subunits) > 0

        if !hasLinks && !hasLinksFrom && !hasSubunits {
            errors = append(errors, &ValidationError{
                Message: fmt.Sprintf(`unit "%s" has no incoming or outgoing links`, path),
                Path:    path,
            })
        }
    }

    return errors
}
```

### Pattern 2: Integration into Validate()
**What:** Add rule call to the validation chain in Validate().
**When to use:** After reference validation (orphans can only be detected after references are validated).
**Example:**
```go
// From validator.go:27-29 - existing pattern
errors = append(errors, ValidateReferences(index)...)
errors = append(errors, ValidateSubunitRules(index)...)
errors = append(errors, ValidateLinkRules(index)...)
errors = append(errors, ValidateOrphanUnits(index)...)  // Add after ValidateLinkRules
```

### Pattern 3: Test Structure
**What:** Table-driven tests with parallel execution, following existing rules_test.go patterns.
**When to use:** All new test cases.
**Example:**
```go
func TestValidateOrphanUnits_DetectsOrphan(t *testing.T) {
    t.Parallel()

    units := map[string]*model.Unit{
        "orphan": {Type: model.TypeSystem},  // No Links, LinksFrom, or Subunits
        "connected": {
            Type: model.TypeSystem,
            Links: map[string]model.Link{"other": {Target: "other"}},
        },
        "other": {Type: model.TypeSystem},
    }

    index := validator.BuildIndex(units, "")
    errors := validator.ValidateOrphanUnits(index)

    if len(errors) != 1 {
        t.Fatalf("expected 1 error, got %d", len(errors))
    }

    expectedMsg := `unit "orphan" has no incoming or outgoing links`
    if errors[0].Message != expectedMsg {
        t.Errorf("expected message %q, got %q", expectedMsg, errors[0].Message)
    }
}
```

### Anti-Patterns to Avoid
- **Don't check external boundary nodes:** Only explicitly defined units in the TOML can be orphans
- **Don't fail-fast:** Collect all orphan errors, consistent with existing rules
- **Don't exempt any unit types:** All unit types must be connected (unless they have subunits)
- **Don't check before ValidateReferences:** Orphan check depends on valid references

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Error formatting | Custom error messages | ValidationError struct with Message/Path fields | Consistency with existing errors |
| Unit lookup | Custom traversal | BuildIndex() returns flat map with O(1) lookup | Already exists, handles nested units |
| Error collection | Custom slice logic | `var errors ValidationErrors` + append | Established pattern |

**Key insight:** All infrastructure exists. Only business logic (orphan detection) is new.

## Common Pitfalls

### Pitfall 1: Checking External Boundary Nodes
**What goes wrong:** Treating inferred external boundary nodes (from links to undefined units) as orphans.
**Why it happens:** Links to undefined units could appear to create orphan-like situations.
**How to avoid:** Only check units in the index (explicitly defined in TOML). ValidateReferences catches undefined references first.
**Warning signs:** Errors mentioning units not in the TOML file.

### Pitfall 2: Wrong Rule Order
**What goes wrong:** Orphan check runs before reference validation, causing confusing errors.
**Why it happens:** All rules are called in sequence; order matters for error clarity.
**How to avoid:** Call ValidateOrphanUnits after ValidateReferences and ValidateLinkRules.
**Warning signs:** Orphan errors for units that reference undefined targets.

### Pitfall 3: Forgetting Subunits Exemption
**What goes wrong:** Units with subunits reported as orphans because they have no direct links.
**Why it happens:** Parent units don't have Links/LinksFrom (that's the subunit's job).
**How to avoid:** Check `len(info.Unit.Subunits) > 0` before reporting orphan.
**Warning signs:** Valid system/container units flagged as orphans.

### Pitfall 4: Inconsistent Error Message Format
**What goes wrong:** Error message doesn't match existing patterns.
**Why it happens:** Ad-hoc message construction.
**How to avoid:** Use exact format: `unit "{path}" has no incoming or outgoing links`
**Warning signs:** Message differs from 'unit "path" has no incoming or outgoing links'

## Code Examples

### Complete ValidateOrphanUnits Implementation
```go
// ValidateOrphanUnits checks that all units have connectivity.
// A unit is an orphan if it has no Links, no LinksFrom, and no Subunits.
// Returns all errors found (not fail-fast).
func ValidateOrphanUnits(index map[string]*UnitInfo) ValidationErrors {
    var errors ValidationErrors

    for path, info := range index {
        hasLinks := len(info.Unit.Links) > 0
        hasLinksFrom := len(info.Unit.LinksFrom) > 0
        hasSubunits := len(info.Unit.Subunits) > 0

        if !hasLinks && !hasLinksFrom && !hasSubunits {
            errors = append(errors, &ValidationError{
                Message: fmt.Sprintf(`unit "%s" has no incoming or outgoing links`, path),
                Path:    path,
            })
        }
    }

    return errors
}
```

### Test Cases Required
```go
// Test cases needed (based on existing test patterns):
// 1. TestValidateOrphanUnits_NoOrphans - connected model passes
// 2. TestValidateOrphanUnits_SingleOrphan - one orphan detected
// 3. TestValidateOrphanUnits_MultipleOrphans - all orphans reported
// 4. TestValidateOrphanUnits_UnitWithSubunits - parent not orphan
// 5. TestValidateOrphanUnits_UnitWithLinksFrom - incoming link counts
// 6. TestValidateOrphanUnits_NestedOrphan - nested units can be orphans
// 7. TestValidateOrphanUnits_AllConnected - fully connected model passes
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| N/A | Rule function pattern | v1.0 | Consistent validation architecture |

**Deprecated/outdated:**
- None relevant to this phase.

## Open Questions

1. **Empty model handling**
   - What we know: Empty model (no units) is mentioned as invalid in CONTEXT.md
   - What's unclear: Should ValidateOrphanUnits handle this or is it caught elsewhere?
   - Recommendation: Empty model passes orphan check (no units to be orphans). If needed, add separate validation in Validate() for empty model.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing + stretchr/testify v1.11.1 |
| Config file | None - standard Go test |
| Quick run command | `go test ./internal/validator/... -v -run TestValidateOrphanUnits` |
| Full suite command | `go test ./internal/validator/... -v` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| VAL-01 | Reject orphan units | unit | `go test ./internal/validator/... -v -run TestValidateOrphanUnits` | ❌ Wave 0 |
| VAL-02 | Error identifies unit | unit | `go test ./internal/validator/... -v -run TestValidateOrphanUnits` | ❌ Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./internal/validator/... -v`
- **Per wave merge:** `go test ./... -v`
- **Phase gate:** Full suite green before `/gsd-verify-work`

### Wave 0 Gaps
- [ ] `internal/validator/rules_test.go` — Add TestValidateOrphanUnits_* test cases
- [x] Test framework installed — Go testing + testify already in go.mod
- [x] Shared fixtures — Use existing pattern (inline test data)

*(If no gaps: "None — existing test infrastructure covers all phase requirements")*

## Sources

### Primary (HIGH confidence)
- `internal/validator/rules.go` - Existing validation rule patterns (ValidateReferences, ValidateSubunitRules, ValidateLinkRules)
- `internal/validator/validator.go` - Validate() entry point and error collection pattern
- `internal/validator/index.go` - UnitInfo struct and BuildIndex() function
- `internal/validator/errors.go` - ValidationError struct and error formatting
- `internal/validator/rules_test.go` - Test patterns and conventions
- `internal/model/unit.go` - Unit struct with Links, LinksFrom, Subunits fields

### Secondary (MEDIUM confidence)
- CONTEXT.md decisions - Locked implementation choices from user discussion

### Tertiary (LOW confidence)
- None required - all information from codebase inspection

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - No new dependencies; all infrastructure exists
- Architecture: HIGH - Existing patterns are clear and well-documented in code
- Pitfalls: HIGH - Context from CONTEXT.md explicitly addresses edge cases

**Research date:** 2026-03-11
**Valid until:** 30 days (stable Go codebase, no external dependencies)
