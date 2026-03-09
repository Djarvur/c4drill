# Feature Research

**Domain:** C4 Architecture Diagram Generation Tools
**Researched:** 2026-03-09
**Confidence:** MEDIUM

## Feature Landscape

### Table Stakes (Users Expect These)

Features users assume exist. Missing these = product feels incomplete.

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| DSL/Config file parsing | All C4 tools use "diagrams as code" approach | MEDIUM | TOML parsing in Go is straightforward with libraries like pelletier/go-toml |
| C1-C3 layer support | Core C4 model layers are the minimum viable scope | MEDIUM | Context, Containers, Components - each has distinct visualization |
| Basic unit types (person, system, db, queue) | Standard C4 notation elements | LOW | Record shapes with distinct styling per type |
| Relationships/links between units | Diagrams without connections are useless | MEDIUM | Directed edges with labels, styling options |
| Auto-layout (GraphViz) | Manual positioning is tedious and brittle | LOW | Let GraphViz handle layout via dot algorithm |
| SVG output | Vector format is standard for diagrams | MEDIUM | Via go-graphviz library |
| DOT output | Debugging and pipeline compatibility | LOW | Native GraphViz format, intermediate representation |
| Validation with error messages | Broken references silently fail without it | MEDIUM | Must validate unit references, type rules, subunit constraints |

### Differentiators (Competitive Advantage)

Features that set the product apart. Not required, but valuable.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| Collapsed/expanded views | Drill-down navigation without separate files | HIGH | Toggle between summary and detailed views |
| Interactive explore links | Click-through navigation between diagram levels | MEDIUM | Generates clickable links to subunit diagrams |
| Themes/color schemes | Consistent branding, visual differentiation | LOW | Inheritable styling from root properties |
| Tags/stereotypes | Categorize and filter elements | MEDIUM | Apply metadata to units for grouping |
| Legend generation | Auto-document notation and color meanings | MEDIUM | Useful for complex diagrams |
| Sprites/icons | Visual distinction beyond shapes | HIGH | Custom icons for specific tech/roles |
| External/internal variants | Clear visual distinction for boundary elements | LOW | person vs personExternal, system vs systemExternal |
| Box grouping | Logical clustering without C4 hierarchy | MEDIUM | Generic container for visual organization |
| Single-file definition | No complex project structures | LOW | All architecture in one TOML file |
| Native Go binary | Zero runtime dependencies, easy deployment | LOW | Static binary distribution |

### Anti-Features (Commonly Requested, Often Problematic)

Features that seem good but create problems.

| Feature | Why Requested | Why Problematic | Alternative |
|---------|---------------|-----------------|-------------|
| Manual node positioning | "I want it to look exactly like my sketch" | Fights GraphViz, breaks when model changes, maintenance burden | Tune GraphViz parameters, use rank/groups |
| JSON structured errors | "I want to parse errors programmatically" | Over-engineering for v1 CLI tool | Clear text error messages with context |
| Watch/live mode | "Re-render on file save" | Adds complexity, file watching edge cases | Run command explicitly, shell integration |
| Multiple CLI commands | "I want separate validate/generate/render" | Fragmented UX, more to learn, more to break | Single command with flags for control |
| C4/Code (level 4) | "I want class diagrams too" | Scope explosion, different visualization needs | Stop at C3, use UML tools for code level |
| Go library interface | "I want to embed this in my app" | API design overhead, different use case | Stay focused on CLI tool first |
| Web UI | "Visual editing would be nice" | Massive scope shift, defeats "diagrams as code" | Keep it CLI + text editor workflow |

## Feature Dependencies

```
[TOML Parsing]
    └──requires──> [Unit Type Definitions]
                       └──requires──> [Validation Rules]

[DOT Generation]
    └──requires──> [TOML Parsing]
    └──requires──> [Validation Rules]

[SVG Rendering]
    └──requires──> [DOT Generation]
    └──requires──> [go-graphviz library]

[Collapsed/Expanded Views]
    └──requires──> [DOT Generation]
    └──requires──> [Subunit Constraint Logic]

[Explore Links]
    └──requires──> [Collapsed/Expanded Views]
    └──requires──> [File Structure Convention]

[Themes/Styling]
    └──enhances──> [DOT Generation]
    └──requires──> [Property Inheritance]

[Tags/Stereotypes]
    └──enhances──> [Validation Rules]
    └──conflicts──> [Simplicity Goal] (adds complexity)

[Legend Generation]
    └──requires──> [Themes/Styling]
    └──requires──> [Tags/Stereotypes]

[Sprites/Icons]
    └──conflicts──> [Single-file Goal] (requires external assets)
    └──conflicts──> [Simplicity Goal] (complex asset management)
```

### Dependency Notes

- **SVG Rendering requires go-graphviz library:** Native Go implementation, no external graphviz binary dependency
- **Explore Links requires Collapsed/Expanded Views:** Interactive drill-down only makes sense when there's something to expand
- **Themes/Styling enhances DOT Generation:** Visual polish without core functionality change
- **Tags/Stereotypes conflicts with Simplicity Goal:** Adds metadata complexity; defer if not essential
- **Sprites/Icons conflicts with Single-file Goal:** Requires external image assets, breaks self-contained definition

## MVP Definition

### Launch With (v1)

Minimum viable product — what's needed to validate the concept.

- [x] TOML parsing with nested unit definitions — Core input mechanism
- [x] C1-C3 layer generation — The fundamental C4 model scope
- [x] All unit types (person, system, db, queue, external variants, box) — Complete type coverage
- [x] Link/relationship definitions — Diagrams need connections
- [x] Validation with clear errors — Prevent silent failures
- [x] DOT output — Intermediate format, debugging
- [x] SVG output — Final deliverable format
- [x] Basic styling (color, border, style inheritance) — Visual distinction
- [x] Collapsed/expanded views — Core value proposition
- [x] Explore links — Navigation between levels
- [x] Single CLI command — Simple workflow

### Add After Validation (v1.x)

Features to add once core is working.

- [ ] Tags/stereotypes — When users ask for filtering/grouping
- [ ] Legend generation — When diagrams get complex
- [ ] Edge routing options (spline, square, straight) — When layout needs tuning
- [ ] Property inheritance optimization — DRY improvements

### Future Consideration (v2+)

Features to defer until product-market fit is established.

- [ ] Themes/color schemes — When users want branding
- [ ] Sprites/icons — When visual distinction becomes critical
- [ ] Multiple output formats (PNG, PDF) — When distribution needs vary
- [ ] Watch mode — When iterative workflow becomes painful
- [ ] JSON error output — When tooling integration is requested

## Feature Prioritization Matrix

| Feature | User Value | Implementation Cost | Priority |
|---------|------------|---------------------|----------|
| TOML parsing | HIGH | MEDIUM | P1 |
| C1-C3 generation | HIGH | MEDIUM | P1 |
| Unit types (all) | HIGH | LOW | P1 |
| Relationships/links | HIGH | MEDIUM | P1 |
| Validation | HIGH | MEDIUM | P1 |
| DOT output | MEDIUM | LOW | P1 |
| SVG output | HIGH | MEDIUM | P1 |
| Basic styling | MEDIUM | LOW | P1 |
| Collapsed/expanded | HIGH | HIGH | P1 |
| Explore links | HIGH | MEDIUM | P1 |
| Single command | MEDIUM | LOW | P1 |
| Tags/stereotypes | MEDIUM | MEDIUM | P2 |
| Legend generation | LOW | MEDIUM | P2 |
| Edge routing options | MEDIUM | LOW | P2 |
| Themes | MEDIUM | LOW | P3 |
| Sprites/icons | MEDIUM | HIGH | P3 |
| Watch mode | LOW | MEDIUM | P3 |

**Priority key:**
- P1: Must have for launch
- P2: Should have, add when possible
- P3: Nice to have, future consideration

## Competitor Feature Analysis

| Feature | Structurizr | C4-PlantUML | LikeC4 | C4Drill (Our Approach) |
|---------|-------------|-------------|--------|------------------------|
| Definition format | DSL + JSON | PlantUML DSL | YAML-like DSL | TOML |
| Output formats | Multiple (via exporters) | Multiple (via PlantUML) | SVG, PNG | DOT, SVG |
| Auto-layout | Yes | Yes (PlantUML) | Yes | Yes (GraphViz) |
| Interactive nav | Yes (web workspace) | Limited | Yes (live preview) | Yes (explore links) |
| Themes | Yes (cloud themes) | Yes (includes) | Yes | Planned (v2) |
| Sprites/icons | Limited | Extensive | Customizable | Not planned |
| Tags/stereotypes | Yes | Yes | Yes | Planned (v1.x) |
| Legend | Auto-generated | Manual macros | Auto-generated | Planned (v1.x) |
| VSCode extension | Yes | Yes | Yes | Not planned |
| Deployment | SaaS + self-host | CLI + server | CLI | CLI (single binary) |
| Learning curve | Medium | Low (if know PlantUML) | Medium | Low (TOML is simple) |

## Sources

- Structurizr DSL documentation (structurizr.com/help/dsl) — HIGH confidence
- Structurizr themes documentation (structurizr.com/help/themes) — HIGH confidence
- C4-PlantUML GitHub repository (github.com/plantuml-stdlib/C4-PlantUML) — HIGH confidence
- LikeC4 documentation (docs.likec4.dev) — HIGH confidence
- Mermaid C4 documentation (mermaid.js.org/syntax/c4.html) — MEDIUM confidence
- Project requirements from `.planning/PROJECT.md` — HIGH confidence

---
*Feature research for: C4 Architecture Diagram Generation Tools*
*Researched: 2026-03-09*
