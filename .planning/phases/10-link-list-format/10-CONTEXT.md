# Phase 10: Link List Format - Context

**Gathered:** 2026-03-11
**Status:** Ready for planning
**Source:** User specification

<domain>
## Phase Boundary

Change Links and LinksFrom from map structures to list structures with explicit `peer` field. This is a breaking TOML syntax change that affects parser, model, validator, graph builder, all existing TOML examples, and documentation (including skill package).

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

### Documentation updates
- **skill/SKILL.md** — Update link syntax documentation to reflect `[[link]]` and `[[linkFrom]]` with `peer` field
- **skill/examples/*.toml** — Convert all example files to new syntax
- **CLAUDE.md** — Update any link syntax examples if present

### Claude's Discretion
- Order of fields in Link struct definition
- Error message wording for missing peer
- Whether `Peer` field should be `toml:"peer"` explicitly

### Execution Learnings (from 10-01 and 10-02)
- **FindLinkByPeer helper**: Added to `internal/model/link.go` for O(n) slice-based lookups. Necessary because slices don't have O(1) map access. Used by tests and downstream code.
- **Orphan validation interaction**: Units with only incoming links failed Phase 9's orphan validation after migration. Examples needed `linkFrom` declarations to pass. All units must have at least one of: Links, LinksFrom, or Subunits.
- **Cascading changes**: Model changes (10-01) required auto-fixes across 7 downstream areas: parser, graph builder, validator, view scope, render tests, CLI tests, and testdata files.

</decisions>

<specifics>
## Specific Ideas

- "The link attribute will not be a table but list of the link objects with same attributes plus peer (mean the target)"
- "The linkFrom attribute will not be a table but list of the link objects with same attributes plus peer (mean source)"
- "Documentation and skill must also be updated"

</specifics>

<code_context>
## Existing Code Insights

### Code files to modify
- `internal/model/link.go` — Rename `Target` field to `Peer`, update TOML tag
- `internal/model/unit.go` — Change `Links` and `LinksFrom` from `map[string]Link` to `[]Link`
- `internal/parser/parser.go` — Remove `populateLinkTargets` logic, handle array parsing
- `internal/validator/*.go` — Update all code that iterates Links/LinksFrom, change `Target` to `Peer`
- `internal/graph/*.go` — Update all code that iterates Links/LinksFrom, change `Target` to `Peer`
- All `*_test.go` files — Update test data structures, change `Target` to `Peer`

### Documentation files to modify
- `skill/SKILL.md` — Update link syntax section with new `[[link]]`/`[[linkFrom]]` format and `peer` field
- `skill/examples/*.toml` — Convert all examples from `[unit.link.target]` to `[[unit.link]]` + `peer = "target"`
- `CLAUDE.md` — Update link examples if any

### Established Patterns
- go-toml/v2 handles `[[array]]` syntax automatically for slice fields
- Validation rules iterate over maps — will need to iterate over slices instead
- Graph builder iterates over Links map — will iterate over Links slice
- All references to `link.Target` become `link.Peer`

### Integration Points
- Parser populates Link.Target from map keys — this logic is removed entirely (TOML unmarshals directly)
- Validator checks Link.Target references — will check `link.Peer`
- Graph builder creates edges from Links map — will iterate `unit.Links` slice and use `link.Peer`
- **NEW: FindLinkByPeer(links []Link, peer string) (Link, bool)** — Helper in internal/model/link.go for finding links by peer name
- **NEW: Orphan validation impact** — All units must have Links, LinksFrom, or Subunits. This affects testdata files.

</code_context>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 10-link-list-format*
*Context gathered: 2026-03-11*
