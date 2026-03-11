# Phase 10: Allow Parent Links - Context

**Gathered:** 2026-03-11
**Status:** Ready for planning

<domain>
## Phase Boundary

Remove the validation restriction that prevents linking to units with subunits. The rule must be: **Links and LinksFrom must NOT be considered as subunits**. Units having Links and LinksFrom sections can be linked from other units without validation errors.

This is a modification to the existing `ValidateLinkRules` function.

</domain>

<decisions>
## Implementation Decisions

### Validation rule change (REQUIRED)
- **Remove the check that prevents linking to units with subunits** (lines 107-115 in rules.go)
- **Remove the check that prevents units with subunits from having Links/LinksFrom** (lines 96-105 in rules.go)
- Keep orphan detection unchanged (units with Links/LinksFrom are not orphans)

### What changes
Current validation rejects:
```
error: unit "mainapp" has subunits and cannot be linked to directly
error: unit "mainapp" has subunits and cannot have direct links
```

After Phase 10:
- These errors are removed
- Units with subunits CAN have Links and LinksFrom
- Units with subunits CAN be linked to by other units

</decisions>

<specifics>
## Specific Ideas

- This enables linking to parent containers in architectures
- Useful for all-expanded diagrams showing connections between high-level units
- Allows bidirectional linking between any units regardless of subunit status

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
