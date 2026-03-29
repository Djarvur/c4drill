# Phase 9: No Orphan Units - Context

**Gathered:** 2026-03-11
**Status:** Ready for planning

<domain>
## Phase Boundary

Add a validation rule that rejects TOML files containing units with no incoming or outgoing links (orphan units). This is a new validation rule in the existing validator package, not a CLI feature or output change.

</domain>

<decisions>
## Implementation Decisions

### Error reporting format
- One ValidationError per orphan unit (consistent with existing rules like ValidateReferences)
- Standard message format: 'unit "path" has no incoming or outgoing links'
- Report all orphans found (not fail-fast, consistent with existing validation pattern)
- Use existing ValidationError struct with Path field (no new error type needed)

### Orphan definition
- Unit is orphan if it has NO incoming AND NO outgoing links AND NO subunits
- Either direction counts for connectivity (incoming OR outgoing = connected)
- Both Links and LinksFrom fields count toward connectivity
- Only explicitly defined units can be orphans (external boundary nodes are inferred, not checked)

### Unit type exceptions
- No unit type exemptions (all types must have links if they have no subunits)
- Units with subunits are exempt (they have children, not orphans)
- Single unit model is invalid (can't link to anything, will be caught as orphan)
- Empty model is invalid (no architecture defined)

### External boundary nodes
- External boundary nodes are NOT checked for orphan status
- Links to/from undefined units are NOT allowed (caught by ValidateReferences rule first)
- Connectivity checking considers only explicit TOML content
- Only links to defined units count as connectivity

### Claude's Discretion
- Exact implementation of orphan detection algorithm
- How to efficiently check connectivity using the existing index
- Order of validation rules (orphan check should run after reference validation)
- Internal helper function design

</decisions>

<code_context>
## Existing Code Insights

### Reusable Assets
- **BuildIndex()**: Creates flat map of all units for O(1) lookup (internal/validator/index.go)
- **ValidationError struct**: Standard error type with Message, Path, Line fields
- **Validate() function**: Entry point that runs all rules and aggregates errors
- **ValidateReferences rule**: Pattern for checking unit relationships
- **ValidationErrors slice**: Collects all errors (not fail-fast pattern)

### Established Patterns
- **Rule function pattern**: `ValidateXxx(index map[string]*UnitInfo) ValidationErrors`
- **Error collection**: Use `var errors ValidationErrors` and append
- **Path in errors**: Always include full dotted path for context
- **Preallocation**: Use `typicalErrorCount = 4` for error slice capacity

### Integration Points
- **internal/validator/validator.go**: Add new rule call in Validate() function
- **internal/validator/rules.go**: Add new ValidateOrphanUnits() function
- **internal/validator/rules_test.go**: Add unit tests for orphan detection
- **cmd/c4drill/root.go**: No changes needed (existing validation flow)

</code_context>

<specifics>
## Specific Ideas

No specific requirements — follow existing validation patterns for consistency.

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 09-no-orphan-units*
*Context gathered: 2026-03-11*
