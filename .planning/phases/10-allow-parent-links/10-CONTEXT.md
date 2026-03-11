# Phase 10: Allow Parent Links - Context

**Gathered:** 2026-03-11
**Status:** Ready for planning

<domain>
## Phase Boundary

Remove the validation restriction that prevents linking to units with subunits. Users must be able to link to parent units (units with subunits) without validation errors.

This is a modification to the existing `ValidateLinkRules` function, not a new phase.

</domain>

<decisions>
## Implementation Decisions

### Validation rule change
- Remove the check that prevents linking to units with subunits (lines 108-115 in rules.go)
- Remove the check that prevents units with subunits from having their own Links/LinksFrom (lines 96-105 in rules.go)
- Keep orphan detection unchanged (units with Links/LinksFrom are not orphans)

### Claude's Discretion
- Exact implementation of the modified `ValidateLinkRules` function
- Whether to split the function or keep backward compatibility

</decisions>

<specifics>
## Specific Ideas

- This enables linking to the parent container in an architecture,- Useful for all-expanded diagrams where you want to show connections between high-level units

</specifics>

<code_context>
## Existing Code Insights

### Reusable Assets
- **ValidateLinkRules**: Current function that needs modification (internal/validator/rules.go)
- **UnitInfo struct**: Provides unit data including Links, LinksFrom, Subunits

### Established Patterns
- **Validation rule pattern**: `ValidateXxx(index map[string]*UnitInfo) ValidationErrors`
- **Error collection**: Not fail-fast, collect all errors

### Integration Points
- **internal/validator/rules.go**: Modify `ValidateLinkRules` function
- **internal/validator/validator.go**: No changes needed (calls all validation rules)

</code_context>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 10-allow-parent-links*
*Context gathered: 2026-03-11*
