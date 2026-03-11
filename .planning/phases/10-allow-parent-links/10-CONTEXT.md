# Phase 10: Allow Parent Links - Context

**Gathered:** 2026-03-11
**Status:** Ready for planning
**Source:** User clarification

<domain>
## Phase Boundary

Remove validation restrictions that prevent linking to/from units with subunits.

**Current behavior (to be removed):**
```
error: unit "X" has subunits and cannot be linked to directly
error: unit "X" has subunits and cannot have direct links
```

**Desired behavior:**
```
✓ Links to units with subunits allowed
✓ Units with subunits can have Links/LinksFrom
✓ Orphan detection unchanged (units with Links/LinksFrom are not orphans)
```

This is a code removal task in `ValidateLinkRules` function.

</domain>

<decisions>
## Implementation Decisions

### What to remove
- Lines 107-115 in rules.go: Check preventing linking to units with subunits
- Lines 96-105 in rules.go: Check preventing units with subunits from having Links/LinksFrom

### What stays the same
- `ValidateReferences` - unchanged
- `ValidateSubunitRules` - unchanged
- `ValidateOrphanUnits` - unchanged (units with Links/LinksFrom are not orphans)

### Claude's Discretion
- Exact code cleanup approach

</decisions>

<specifics>
## Use Cases

- Link to a parent container from external systems
- Link from a parent container to external systems
- Bidirectional links between any units regardless of subunit status
- All-expanded diagrams showing connections at all levels

</specifics>

<code_context>
## Existing Code

### File to modify
`internal/validator/rules.go` — `ValidateLinkRules` function (lines 78-119)

### Pattern toValidation rules follow `ValidateXxx(index) ValidationErrors` pattern

</code_context>

<deferred>
## Deferred Ideas

None — straightforward code removal.

</deferred>

---

*Phase: 10-allow-parent-links*
*Context gathered: 2026-03-11 (updated with clarification)*
