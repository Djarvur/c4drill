# Phase 10: Link List Format - Context

**Gathered:** 2026-03-11
**Status:** Ready for planning
**Source:** User specification

<domain>
## Phase Boundary

Change Links and LinksFrom from map structures to list structures with explicit `peer` field. This is a breaking TOML syntax change that affects parser, model, validator, graph builder, and all existing TOML examples.

**Current (maps):**
```toml
[api.link.db]
technology = "HTTP"
```
```go
Links map[string]Link     // key = target
LinksFrom map[string]Link // key = source (stored in Target field)
```

**New (lists with peer):**
```toml
[[api.link]]
peer = "db"
technology = "HTTP"

[[api.linkFrom]]
peer = "web"
technology = "HTTP"
```
```go
Links []Link     // each has explicit Peer (the target)
LinksFrom []Link // each has explicit Peer (the source)
```

</domain>

<decisions>
## Implementation Decisions

### Field naming
- **Single unified field name `peer`** for both Links and LinksFrom
- In `link`, `peer` means the target unit
- In `linkFrom`, `peer` means the source unit
- Replaces current `Target` field — rename to `Peer`
- No separate `Source` field needed — same field name, different semantic context

### Link structure
- `link` attribute is a **list** of link objects (TOML array of tables `[[unit.link]]`)
- Each link object has all existing Link attributes (arrow, rank, color, style, technology, description, labelPosition)
- Each link object has **explicit `peer` field** (required) — the target unit
- Multiple links to the same peer are now supported

### LinksFrom structure
- `linkFrom` attribute is a **list** of link objects (TOML array of tables `[[unit.linkFrom]]`)
- Each linkFrom object has all existing Link attributes
- Each linkFrom object has **explicit `peer` field** (required) — the source unit

### TOML syntax
- Use TOML array of tables syntax: `[[unit.link]]` and `[[unit.linkFrom]]`
- Example:
  ```toml
  [[api.link]]
  peer = "db"
  technology = "HTTP"
  description = "queries data"
  
  [[api.link]]
  peer = "db"
  technology = "TCP"
  description = "health checks"
  
  [[db.linkFrom]]
  peer = "api"
  technology = "HTTP"
  ```

### Claude's Discretion
- Order of fields in Link struct definition
- Error message wording for missing peer
- Whether `Peer` field should be `toml:"peer"` explicitly

</decisions>

<specifics>
## Specific Ideas

- "The link attribute will not be a table but list of the link objects with same attributes plus peer (mean the target)"
- "The linkFrom attribute will not be a table but list of the link objects with same attributes plus peer (mean source)"

</specifics>

<code_context>
## Existing Code Insights

### Files to modify
- `internal/model/link.go` — Rename `Target` field to `Peer`, update TOML tag
- `internal/model/unit.go` — Change `Links` and `LinksFrom` from `map[string]Link` to `[]Link`
- `internal/parser/parser.go` — Remove `populateLinkTargets` logic, handle array parsing
- `internal/validator/*.go` — Update all code that iterates Links/LinksFrom, change `Target` to `Peer`
- `internal/graph/*.go` — Update all code that iterates Links/LinksFrom, change `Target` to `Peer`
- `skill/examples/*.toml` — Convert all examples to new `[[link]]`/`[[linkFrom]]` syntax with `peer`
- All `*_test.go` files — Update test data structures, change `Target` to `Peer`

### Established Patterns
- go-toml/v2 handles `[[array]]` syntax automatically for slice fields
- Validation rules iterate over maps — will need to iterate over slices instead
- Graph builder iterates over Links map — will iterate over Links slice
- All references to `link.Target` become `link.Peer`

### Integration Points
- Parser populates Link.Target from map keys — this logic is removed entirely (TOML unmarshals directly)
- Validator checks Link.Target references — will check `link.Peer`
- Graph builder creates edges from Links map — will iterate `unit.Links` slice and use `link.Peer`

</code_context>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 10-link-list-format*
*Context gathered: 2026-03-11*
