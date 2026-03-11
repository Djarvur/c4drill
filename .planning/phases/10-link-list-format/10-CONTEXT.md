# Phase 10: Link List Format - Context

**Gathered:** 2026-03-11
**Status:** Ready for planning
**Source:** User specification

<domain>
## Phase Boundary

Change Links and LinksFrom from map structures to list structures with explicit target/source fields. This is a breaking TOML syntax change that affects parser, model, validator, graph builder, and all existing TOML examples.

**Current (maps):**
```toml
[api.link.db]
technology = "HTTP"
```
```go
Links map[string]Link     // key = target
LinksFrom map[string]Link // key = source (stored in Target field)
```

**New (lists):**
```toml
[[api.link]]
target = "db"
technology = "HTTP"
```
```go
Links []Link     // each has explicit Target
LinksFrom []Link // each has explicit Source
```

</domain>

<decisions>
## Implementation Decisions

### Link structure
- `link` attribute is a **list** of link objects (TOML array of tables `[[unit.link]]`)
- Each link object has all existing Link attributes (arrow, rank, color, style, technology, description, labelPosition)
- Each link object has **explicit `target` field** (required)
- Multiple links to the same target are now supported

### LinksFrom structure
- `linkFrom` attribute is a **list** of link objects (TOML array of tables `[[unit.linkFrom]]`)
- Each linkFrom object has all existing Link attributes
- Each linkFrom object has **explicit `source` field** (required)
- The `source` field is a new addition — replaces the confusing reuse of `Target` field

### TOML syntax
- Use TOML array of tables syntax: `[[unit.link]]` and `[[unit.linkFrom]]`
- Example:
  ```toml
  [[api.link]]
  target = "db"
  technology = "HTTP"
  description = "queries data"
  
  [[api.link]]
  target = "db"
  technology = "TCP"
  description = "health checks"
  ```

### Claude's Discretion
- Order of fields in Link struct definition
- Error message wording for missing target/source
- Whether to keep `Target` field name or rename for consistency

</decisions>

<specifics>
## Specific Ideas

- "The link attribute will not be a table but list of the link objects with same attributes plus target"
- "The linkFrom attribute will not be a table but list of the link objects with same attributes plus source (required)"

</specifics>

<code_context>
## Existing Code Insights

### Files to modify
- `internal/model/link.go` — Add `Source` field, consider field ordering
- `internal/model/unit.go` — Change `Links` and `LinksFrom` from `map[string]Link` to `[]Link`
- `internal/parser/parser.go` — Remove `populateLinkTargets` logic, handle array parsing
- `internal/validator/*.go` — Update all code that iterates Links/LinksFrom
- `internal/graph/*.go` — Update all code that iterates Links/LinksFrom
- `skill/examples/*.toml` — Convert all examples to new syntax
- All `*_test.go` files — Update test data structures

### Established Patterns
- go-toml/v2 handles `[[array]]` syntax automatically for slice fields
- Validation rules iterate over maps — will need to iterate over slices instead
- Graph builder iterates over maps — same change needed

### Integration Points
- Parser populates Link.Target from map keys — this logic is removed entirely
- Validator checks Link.Target references — will check explicit Target field
- Graph builder creates edges from Links map — will iterate Links slice

</code_context>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 10-link-list-format*
*Context gathered: 2026-03-11*
