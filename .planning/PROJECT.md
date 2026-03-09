# C4Drill

## What This Is

A Go CLI tool for generating C4 architecture diagrams from TOML definitions. Users describe their system architecture in a structured TOML file, and C4Drill renders it as GraphViz DOT or SVG diagrams. Supports C1 (Context), C2 (Containers), and C3 (Components) layers with collapsed/expanded views and interactive explore links.

## Core Value

Transform simple TOML architecture descriptions into professional C4 diagrams without manual drawing.

## Requirements

### Validated

(None yet — ship to validate)

### Active

- [ ] Parse TOML input file with C4 model definition
- [ ] Validate model integrity (references, type rules, subunit constraints)
- [ ] Generate GraphViz DOT output for C1-C3 layers
- [ ] Render SVG output via go-graphviz
- [ ] Support collapsed/expanded unit views
- [ ] Generate explore links for drilling into nested structures
- [ ] Support all unit types: person, personExternal, system, systemExternal, db, dbExternal, queue, queueExternal, box
- [ ] Apply styling: colors, borders, edge routing styles
- [ ] Single CLI command interface

### Out of Scope

- **C4 Layer 4 (Code)** — Class/function level diagrams not needed for v1
- **Go library/module** — Pure CLI tool, no library interface
- **Manual positioning** — Rely on GraphViz auto-layout
- **JSON error output** — CLI errors only for v1
- **Live editing/watch mode** — Single-shot rendering
- **Multiple commands** — Single command does everything

## Context

C4 model is a lean approach to software architecture documentation created by Simon Brown. It uses four levels of abstraction:
- **C1 (Context)**: System context showing users and external systems
- **C2 (Containers)**: Deployable units within a system (apps, databases, etc.)
- **C3 (Components)**: Logical components within a container
- **C4 (Code)**: Class/function level (out of scope)

The tool uses nested TOML objects where each level contains strictly typed subunits. Systems and boxes can contain subunits; other types cannot. Links define relationships between units with styling options.

## Constraints

- **Tech Stack**: Go 1.21+ — User specified
- **Input Format**: TOML — User specified, single file
- **Output**: GraphViz DOT and SVG — User specified, via go-graphviz
- **Diagram Scope**: C1-C3 layers only — User specified

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Single CLI command | Simplicity for documentation workflow | — Pending |
| Auto-layout only | Let GraphViz handle positioning | — Pending |
| TOML input | Human-readable, supports nested structures | — Pending |
| go-graphviz library | Native Go, no external graphviz binary needed | — Pending |

## TOML Schema (Reference)

### Root Level
```toml
[properties]
name = "Project Name"
description = "Project description"
color = "transparent"        # optional, default transparent
style = "none"               # optional: none, solid, dotted, dashed
border = "transparent"       # optional, default transparent
edges = "straight"           # optional: straight, spline, square
expanded = ["system1", "box1"]  # optional, default empty

[section_name]  # Context-level unit
type = "system"              # optional, default system
name = "Display Name"
description = "Description"
color = "transparent"
style = "none"
border = "transparent"
link = { target = { reverse = false, equal = false, color = "black", style = "solid" } }
linkFrom = { source = { ... } }

# For system and box types only:
edges = "straight"           # optional, inherits from parent
expanded = ["subunit1"]      # optional

[section_name.subunit]       # Container-level unit (inside system)
# Same attributes as context-level units
```

### Unit Types
- `person`, `personExternal` — Actors using the system
- `system`, `systemExternal` — Software systems
- `db`, `dbExternal` — Databases
- `queue`, `queueExternal` — Message queues
- `box` — Grouping container

### Link Object
```toml
link = { "target_unit" = { reverse = false, equal = false, color = "black", style = "solid" } }
```

## Validation Rules

1. Referenced units must be defined
2. Units with subunits cannot be referenced by links
3. Units with subunits cannot have their own links
4. Subunits only allowed for `system` and `box` types

## Rendering Behavior

- **Collapsed**: Single record shape with "explore" link to drill-down file
- **Expanded**: Cluster with subunits rendered inside
- **File structure**: `{basename}.svg` for context, `{basename}/` directory for expanded units
- **Shapes**: Person, DB, Queue, System each have distinct record shapes

---
*Last updated: 2026-03-09 after initialization*
