# Phase 2: Validation - Context

**Gathered:** 2026-03-09
**Status:** Ready for planning

<domain>
## Phase Boundary

Validate model integrity after parsing — check references, type rules, subunit constraints. Invalid TOML files produce clear, actionable error messages before any rendering. This phase does NOT render anything, only validates the parsed model.

</domain>

<decisions>
## Implementation Decisions

### Error output format
- Plain single-line format (not multi-line structured, not JSON)
- Prefix style: `error: <message>`
- Example: `error: undefined unit "db1" referenced from "api" at line 15`

### Validation behavior
- Collect all errors before reporting (not fail-fast)
- Report in file order (top-to-bottom traversal)
- Exit code 1 on any error, 0 on success (standard CLI convention)

### Error detail level
- Include suggestions when possible: `did you mean "db"?` for similar names
- Use concise wording: `undefined unit "x"` not `unit "x" is not defined`
- Best-effort line numbers: use go-toml position when available, omit otherwise
- Full path for nested units: `mainapp.api.handler` not just `handler`
- Show error count summary at end: `5 errors found`

### Claude's Discretion
- Exact algorithm for "did you mean" suggestions (Levenshtein distance threshold)
- Maximum errors to collect before stopping (if any limit)
- Exact wording for each error type (follow concise pattern)

</decisions>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/parser/errors.go`: ParseError type with Message, Line, Context, Cause fields — extend or create ValidationError type
- `internal/parser/parser.go`: Model struct with Units map — validator will traverse this
- `internal/model/unit.go`: Unit struct with Type, Subunits, Links, LinksFrom — validation rules apply to these

### Established Patterns
- Error wrapping with position info (go-toml DecodeError.Position())
- TOML struct tags for field mapping
- Type discriminator pattern (UnitType enum)

### Integration Points
- Validator runs after parser.Parse() succeeds
- Called from main CLI flow before any rendering
- Uses model types from internal/model/

</code_context>

<specifics>
## Specific Ideas

- Error format should feel like `rustc` or `gcc` but single-line for CLI friendliness
- Should work in CI pipelines (exit codes, no interactive prompts)

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 02-validation*
*Context gathered: 2026-03-09*
