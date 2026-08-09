# Phase 14: Nesting Validation - Research

**Researched:** 2026-03-17
**Domain:** C4 Model Hierarchy Validation in Go
**Confidence:** HIGH

## Summary

This phase implements validation of the C4 model nesting hierarchy. The C4 model defines a strict 3-level hierarchy: C1 (Context) at the top level, C2 (Container) inside systems, and C3 (Component) inside containers. Currently, the `ValidateSubunitRules` function only checks IF a type can have subunits, not WHAT types it can contain.

**Primary recommendation:** Add a new `ValidateNestingHierarchy` rule to `internal/validator/rules.go` that uses the existing `UnitInfo.Parent` field to determine nesting depth and validates that subunits match the expected C4 level for their container.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- Top level: C1 types only (person, system, db, queue, box)
- Inside system/box: C2 types only (container, containerDb, containerQueue)
- Inside container: C3 types only (component, componentDb, componentQueue)
- C3 types: No subunits (leaf nodes) - already validated by ValidateSubunitRules
- External types follow same nesting rules as their base types

### Claude's Discretion
- Error message format and clarity
- Helper function organization (model package vs validator package)
- Test organization and coverage approach

### Deferred Ideas (OUT OF SCOPE)
- None identified
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| NEST-01 | Top-level units must be C1 types | Use `UnitInfo.Parent == ""` to identify top-level; validate against C1 type set |
| NEST-02 | C1 containers (system, box) can only contain C2 types | Use `UnitInfo.Parent` to identify depth; check parent type to determine allowed child types |
| NEST-03 | C2 containers (container) can only contain C3 types | Same pattern as NEST-02; map parent type to allowed child types |
</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go | 1.21+ | Primary language | Project is Go-based CLI tool |
| testing | stdlib | Unit tests | Go's built-in testing framework |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| fmt | stdlib | Error formatting | All validation error messages |
| maps | stdlib | Map operations | Used in BuildIndex for recursive copy |
| slices | stdlib | Slice operations | Used in ValidateReferences for suggestions |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Helper functions in validator | Helper functions in model | Model package has no external deps; validator already imports model |

**Installation:**
No new dependencies required - uses Go standard library only.

## Architecture Patterns

### Current Validator Architecture

The validator follows a clean separation pattern:
- `validator.go` - Entry point `Validate()` orchestrates all rules
- `rules.go` - Individual validation rule functions
- `index.go` - `BuildIndex()` creates flat lookup map with `UnitInfo` metadata
- `errors.go` - `ValidationError` type with human-readable formatting

### Pattern: Validation Rule Function

Each validation rule follows this signature:
```go
func ValidateX(index map[string]*UnitInfo) ValidationErrors
```

**What:** Rules iterate over the index, collect errors, return slice (not fail-fast).
**When to use:** All new validation rules.
**Example:**
```go
// Source: internal/validator/rules.go
func ValidateSubunitRules(index map[string]*UnitInfo) ValidationErrors {
    var errors ValidationErrors
    allowedTypes := map[model.UnitType]bool{
        model.TypeSystem:         true,
        model.TypeSystemExternal: true,
        model.TypeBox:            true,
        model.TypeContainer:      true,
    }
    for path, info := range index {
        if len(info.Unit.Subunits) == 0 {
            continue
        }
        if !allowedTypes[info.Unit.Type] {
            errors = append(errors, &ValidationError{
                Message: fmt.Sprintf(`unit "%s" has type %s which cannot have subunits`, path, info.Unit.Type),
                Path:    path,
            })
        }
    }
    return errors
}
```

### Pattern: UnitInfo for Nesting Context

The `UnitInfo` struct already provides the necessary metadata:
```go
// Source: internal/validator/index.go
type UnitInfo struct {
    Unit     *model.Unit  // Pointer to actual unit
    FullPath string       // Complete dotted path (e.g., "mainapp.api.handler")
    Parent   string       // Parent's full path (empty for top-level)
}
```

### Recommended Project Structure (No changes needed)
```
internal/validator/
    ├── validator.go    # Add ValidateNestingHierarchy call
    ├── rules.go        # Add ValidateNestingHierarchy function
    ├── rules_test.go   # Add tests for new rule
    ├── index.go        # No changes needed
    └── errors.go       # No changes needed
```

### Anti-Patterns to Avoid
- **Duplicating type constants:** Don't re-define C1/C2/C3 type sets; use model.TypeX constants
- **Checking nesting depth via path parsing:** Use `UnitInfo.Parent` field (already computed)
- **Fail-fast validation:** Return all errors, not just first one
- **Inconsistent error messages:** Follow existing format patterns

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Type level classification | Custom switch/if chains | Map-based lookup with model.TypeX | Consistent with existing code, easy to maintain |
| Parent depth calculation | Path splitting logic | UnitInfo.Parent field | Already computed during BuildIndex |
| External type handling | Separate validation path | Strip "External" suffix, use base type rules | Avoids code duplication |

**Key insight:** The `UnitInfo.Parent` field already provides the nesting context. Top-level units have `Parent == ""`, second-level units have `Parent` without dots (or count dots for depth), etc.

## Common Pitfalls

### Pitfall 1: Forgetting External Type Variants
**What goes wrong:** Validation rejects valid `systemExternal` at top level because only `system` is in allowed set.
**Why it happens:** External variants (personExternal, systemExternal, dbExternal, queueExternal) are valid C1 types.
**How to avoid:** Either include all external variants in allowed sets, or normalize types by stripping "External" suffix.
**Warning signs:** Test cases with external types failing.

### Pitfall 2: Incorrect Depth Calculation
**What goes wrong:** `container` at `system.subcontainer` is incorrectly rejected because depth is calculated wrong.
**Why it happens:** Using string operations on path (counting dots) instead of using `UnitInfo.Parent` context.
**How to avoid:** Use the `Parent` field to determine container context, then check parent's type.
**Warning signs:** Nested units at valid depths being rejected.

### Pitfall 3: Missing Integration in Validate()
**What goes wrong:** New rule is implemented but never called.
**Why it happens:** Forgetting to add rule to the `Validate()` function in `validator.go`.
**How to avoid:** Add `errors = append(errors, ValidateNestingHierarchy(index)...)` to `Validate()` immediately after writing the rule.
**Warning signs:** Tests pass but invalid TOML files are not rejected.

### Pitfall 4: Not Handling All Edge Cases
**What goes wrong:** Validation crashes on edge cases like empty models or deeply nested structures.
**Why it happens:** Not checking for nil/empty values before map access.
**How to avoid:** Follow existing patterns - check `if len(info.Unit.Subunits) == 0` before processing, handle empty index gracefully.
**Warning signs:** Panic on empty or minimal test cases.

## Code Examples

### Type Classification Helper (Recommended Addition)

Add to `rules.go`:

```go
// C1 types (Context level) - valid at top level only
var c1Types = map[model.UnitType]bool{
    model.TypePerson:         true,
    model.TypePersonExternal: true,
    model.TypeSystem:         true,
    model.TypeSystemExternal: true,
    model.TypeDb:             true,
    model.TypeDbExternal:     true,
    model.TypeQueue:          true,
    model.TypeQueueExternal:  true,
    model.TypeBox:            true,
}

// C2 types (Container level) - valid inside system/box only
var c2Types = map[model.UnitType]bool{
    model.TypeContainer:      true,
    model.TypeContainerDb:    true,
    model.TypeContainerQueue: true,
}

// C3 types (Component level) - valid inside container only
var c3Types = map[model.UnitType]bool{
    model.TypeComponent:      true,
    model.TypeComponentDb:    true,
    model.TypeComponentQueue: true,
}
```

### Nesting Validation Implementation

```go
// Source: New function for internal/validator/rules.go
func ValidateNestingHierarchy(index map[string]*UnitInfo) ValidationErrors {
    var errors ValidationErrors

    for path, info := range index {
        unitType := info.Unit.Type

        if info.Parent == "" {
            // Top-level: must be C1 type
            if !c1Types[unitType] {
                errors = append(errors, &ValidationError{
                    Message: fmt.Sprintf(`unit "%s" has type %s which is not allowed at top level (C1 types only)`, path, unitType),
                    Path:    path,
                })
            }
            continue
        }

        // Has parent - check parent's type to determine allowed children
        parentInfo, exists := index[info.Parent]
        if !exists {
            continue // Orphan reference - ValidateReferences handles this
        }

        parentType := parentInfo.Unit.Type

        // Check if parent is C1 container (system, systemExternal, box)
        if parentType == model.TypeSystem || parentType == model.TypeSystemExternal || parentType == model.TypeBox {
            if !c2Types[unitType] {
                errors = append(errors, &ValidationError{
                    Message: fmt.Sprintf(`unit "%s" has type %s which must be inside container (C2 types only in %s)`, path, unitType, parentType),
                    Path:    path,
                })
            }
            continue
        }

        // Check if parent is C2 container (container)
        if parentType == model.TypeContainer {
            if !c3Types[unitType] {
                errors = append(errors, &ValidationError{
                    Message: fmt.Sprintf(`unit "%s" has type %s which must be inside component (C3 types only in container)`, path, unitType),
                    Path:    path,
                })
            }
            continue
        }

        // Parent is non-container type - shouldn't have children (ValidateSubunitRules handles)
    }

    return errors
}
```

### Integration in validator.go

```go
// Source: Update to internal/validator/validator.go
func Validate(m *parser.Model) ValidationErrors {
    if m == nil {
        return nil
    }

    index := BuildIndex(m.Units, "")
    populateIncomingLinks(index)

    errors := make(ValidationErrors, 0, typicalErrorCount)

    errors = append(errors, ValidateReferences(index)...)
    errors = append(errors, ValidateSubunitRules(index)...)
    errors = append(errors, ValidateLinkRules(index)...)
    errors = append(errors, ValidateOrphanUnits(index)...)
    errors = append(errors, ValidateNestingHierarchy(index)...)  // ADD THIS LINE

    if len(errors) == 0 {
        return nil
    }

    return errors
}
```

### Test Pattern (Follow Existing)

```go
// Source: Pattern from internal/validator/rules_test.go
func TestValidateNestingHierarchy_RejectsC2AtTopLevel(t *testing.T) {
    t.Parallel()

    units := map[string]*model.Unit{
        "container": {
            Type: model.TypeContainer, // C2 type at top level - INVALID
        },
    }

    index := validator.BuildIndex(units, "")
    errors := validator.ValidateNestingHierarchy(index)

    if len(errors) != 1 {
        t.Fatalf("expected 1 error, got %d", len(errors))
    }

    if !strings.Contains(errors[0].Message, "not allowed at top level") {
        t.Errorf("expected top level error, got %q", errors[0].Message)
    }
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Validate only IF subunits allowed | Need to validate WHAT subunits allowed | Phase 14 | Prevents invalid C4 hierarchies |

**Deprecated/outdated:**
- None for this phase

## Open Questions

1. **Should ValidateNestingHierarchy run before or after ValidateSubunitRules?**
   - What we know: Both are independent, neither depends on the other
   - What's unclear: Whether order affects error message priority
   - Recommendation: Order doesn't matter; add after existing rules for consistency

2. **Should error messages suggest valid types?**
   - What we know: Existing `ValidateReferences` uses suggestions
   - What's unclear: Whether C4 hierarchy violations benefit from suggestions
   - Recommendation: Keep error messages simple; C4 rules are well-defined, suggestions add complexity

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing (stdlib) |
| Config file | None - use `go test` |
| Quick run command | `go test ./internal/validator/... -run TestValidateNesting -v` |
| Full suite command | `go test ./internal/validator/... -v` |

### Phase Requirements -> Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| NEST-01 | Reject C2/C3 at top level | unit | `go test ./internal/validator/... -run TestValidateNestingHierarchy_RejectsC2AtTopLevel -v` | Wave 0 |
| NEST-02 | Reject C3 inside system/box | unit | `go test ./internal/validator/... -run TestValidateNestingHierarchy_RejectsC3InSystem -v` | Wave 0 |
| NEST-03 | Reject C2 inside container | unit | `go test ./internal/validator/... -run TestValidateNestingHierarchy_RejectsC2InContainer -v` | Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./internal/validator/... -v`
- **Per wave merge:** `go test ./... -v`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `internal/validator/rules_test.go` - add TestValidateNestingHierarchy_* functions
- No framework install needed (Go stdlib testing)

## Sources

### Primary (HIGH confidence)
- `internal/validator/rules.go` - existing validation patterns and structure
- `internal/validator/validator.go` - Validate() orchestration pattern
- `internal/validator/index.go` - UnitInfo structure with Parent field
- `internal/model/unit.go` - UnitType constants organized by C1/C2/C3
- `internal/view/view.go` - Level constants and IsExternalType helper

### Secondary (MEDIUM confidence)
- `internal/validator/rules_test.go` - existing test patterns to follow
- `internal/validator/errors.go` - ValidationError structure

### Tertiary (LOW confidence)
- None required - all findings from code inspection

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - Go stdlib only, existing patterns well-established
- Architecture: HIGH - Code inspection shows clear patterns to follow
- Pitfalls: HIGH - Based on analysis of existing code and C4 model rules

**Research date:** 2026-03-17
**Valid until:** 30 days (Go patterns stable, C4 model unchanged)
