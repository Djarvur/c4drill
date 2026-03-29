# Phase 7: AI Documentation - Context

**Gathered:** 2026-03-10
**Status:** Ready for planning

<domain>
## Phase Boundary

Create an installable skill that teaches AI assistants how to generate valid C4Drill TOML architecture definitions. The skill includes complete TOML schema reference, working examples, concise prompt patterns, and CI validation for examples. This phase delivers a reusable skill any agent can install — not project-specific documentation.

</domain>

<decisions>
## Implementation Decisions

### Skill Distribution
- Installable via `npx skills add` command
- Lives in `skill/` directory inside c4drill repo
- Users install with: `npx skills add Djarvur/c4drill@c4drill-toml`
- Not a CLAUDE.md file — a proper skill package

### Target Audience
- Generic LLM (not model-specific)
- Any AI assistant that can install skills
- Works across Claude Code, Cursor, OpenCode, etc.

### Content Scope
- Full TOML schema coverage (all unit types, all fields, all link syntax)
- Not a subset — complete reference for AI training
- Include edge cases (external types, nesting rules, link variations)

### Example Selection
- 4-5 examples with varying complexity levels:
  1. Minimal (person + system + single link)
  2. Medium (nested system with containers)
  3. Complex (multi-level nesting, multiple link types)
  4. Styling example (colors, borders, edge styles)
  5. Full architecture (realistic e-commerce or similar)
- Examples must be parseable by c4drill
- Each example demonstrates specific features

### Prompt Pattern Style
- Concise, direct instructions
- Short prompts AI can use: "Generate a C4 diagram with person, system, db. Include links."
- Not step-by-step tutorials
- Focus on what AI needs to know to produce valid TOML

### CI Validation
- Validate examples during skill development only
- NOT in user projects (their TOML files are their responsibility)
- Development CI runs: `c4drill validate skill/examples/*.toml`
- Fail skill PR if examples become invalid

### Claude's Discretion
- Exact directory structure for skill/
- Example TOML file names
- Exact prompt phrasing in skill documentation
- Order of sections in skill README

</decisions>

<specifics>
## Specific Ideas

- "Skill should feel like a reference sheet AI can quickly consult"
- Examples should be copy-pasteable and work immediately
- Prompt patterns should be the kind users actually type to AI assistants

</specifics>

<code_context>
## Existing Code Insights

### Reusable Assets
- `README.md`: Already has good TOML examples and schema documentation — extract and formalize for skill
- `testdata/*.toml`: Existing test files — can inform example selection (valid.toml, nested.toml, links.toml)
- `internal/parser/parser.go`: Model struct defines actual schema — source of truth for field documentation
- `internal/model/unit.go`: Unit struct with all fields — complete field reference

### Established Patterns
- Error format: `error: {message}` (from Phase 2 context)
- CLI flags: `--format/-f`, `--output/-o` (from Phase 6 context)
- Validation rules already documented in PROJECT.md

### Integration Points
- Skill examples should parse with existing c4drill binary
- Schema documentation should match internal/model/ structs exactly
- Examples should exercise all code paths in parser and validator

</code_context>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---
*Phase: 07-ai-documentation*
*Context gathered: 2026-03-10*
