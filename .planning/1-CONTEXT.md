# Phase 1 Context: Foundation & Model

**Created:** 2026-03-09
**Phase Goal:** Development environment and core domain model ready for parsing TOML architecture definitions

---

## Domain

**Boundary:** This phase delivers:
- Go 1.26.1 development environment with mise tasks
- Domain types for all C4 unit types (C1, C2, C3 levels)
- TOML parser that populates model structs
- Simple CLI entry point for testing (main.go → parser → print model)

**Out of scope:**
- Validation logic (Phase 2)
- View generation (Phase 3)
- Graph construction (Phase 3)
- Rendering (Phase 4)
- Navigation/drill-down (Phase 5)
- Polished CLI with flags (Phase 6)

---

## Decisions

### Project Structure

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Layout | `cmd/` + `internal/` | Standard Go CLI pattern |
| Package organization | Pipeline stages | `model/`, `parser/`, `validator/`, `view/`, `graph/`, `render/` |
| Module path | `github.com/Djarvur/c4drill` | User specified |

### Domain Types

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Type representation | Type discriminator | `UnitType` enum field on struct |
| Unit struct | Flat struct | All fields at top level, no nested style/location objects |
| Level-specific types | Yes | C1, C2, C3 have distinct type names |

**Unit struct fields:**

```go
type Unit struct {
    Type         UnitType  // Discriminator
    Name         string
    Description  string
    Technology   string    // C4: technology used (NOT on person types)
    Color        string
    Style        string
    Border       string
    Edges        string    // Cascades from parent
    Width        float64   // 0 = auto (for future use)
    Height       float64   // 0 = auto (for future use)
    Expanded     []string
    Links        []Link
    LinksFrom    []Link
    Subunits     map[string]*Unit  // Recursive, needs pointer
}
```

**Properties struct fields:**

```go
type Properties struct {
    Name        string
    Description string
    Color       string
    Style       string
    Border      string
    Edges       string
    LineLength  int       // Max chars before wrap, 0 = auto
    Expanded    []string
}
```

**Unit types by level:**

| Level | Types |
|-------|-------|
| **C1 (Context)** | person, personExternal, system, systemExternal, db, dbExternal, queue, queueExternal, box |
| **C2 (Containers)** | container, containerDb, containerQueue, containerBox |
| **C3 (Components)** | component, componentDb, componentQueue, componentBox |

**Containment rules:**

| Container | Can Contain |
|-----------|-------------|
| `system`, `systemExternal` | C2 types (container, containerDb, containerQueue, containerBox) |
| `container`, `containerBox` | C3 types (component, componentDb, componentQueue, componentBox) |
| `box` (at C1) | C2 types |
| `containerBox` (at C2) | C3 types |

- Persons defined at C1 only (but displayed on deeper diagrams when linked — Phase 3 concern)

### Link Model

```go
type LabelPosition string  // "tail", "head", "middle"

type Link struct {
    Target        string
    Arrow         ArrowDirection  // Forward, Reverse, Bidirectional, None
    Rank          RankDirection   // Forward, Reverse, Equal
    Color         string
    Style         string
    Technology    string          // C4: protocol/technology used
    Description   string          // C4: relationship description
    LabelPosition LabelPosition   // Where to show label: tail, head, middle (default)
}
```

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Arrow vs Rank | Separate fields | Orthogonal concerns (visual vs layout) |
| Storage | Two slices | `Links []Link` and `LinksFrom []Link` — mirrors TOML schema |
| Target validation | Separate pass | Clean separation, Phase 2 responsibility |
| Label position | Default "middle" | Most common for relationship labels |

### Style & Defaults

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Base defaults | C4-PlantUML | Familiar to C4 users, proven palette |
| Type-specific colors | Exported constants | Configurable if needed |
| Style inheritance | Mixed | `edges` cascades; `color`, `style`, `border` — no inheritance |

**C4-PlantUML defaults:**
- Font color: `#FFFFFF` (white)
- Arrow color: `#666666` (gray)
- Boundary color: `#444444` (dark gray)

**Color scheme by C4 level (all unit types at each level use the same color):**

| Level     | Internal   | External   | Description               |
|-----------|------------|------------|---------------------------|
| **C1**    | `#073B6F`  | `#8A8A8A`  | Dark blue / Dark gray     |
| **C2**    | `#3C7FC0`  | `#A6A6A6`  | Blue / Medium gray        |
| **C3**    | `#78A8D8`  | `#BFBFBF`  | Light blue / Light gray   |

- Border color = Font color for each unit
- Fill color = transparent for all units
- All unit types at the same C4 level share the same color (person, system, db, queue, box at C1; container, containerDb, containerQueue, containerBox at C2; component, componentDb, componentQueue, componentBox at C3)

### TOML Parsing

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Library | pelletier/go-toml | Recommended, well-maintained |
| Mapping | Direct struct unmarshaling | If possible; intermediate map otherwise |

### Text Wrapping (Phase 3 implementation)

| Decision | Choice | Rationale |
| -------- | ------- | --------- |
| Library | mitchellh/go-wordwrap | Same as go-metadot, proven |
| Justification | Aspect ratio optimization | Target 15:5 ratio, adapt from go-metadot |
| Reference | go-metadot/internal/text/justify.go | Port logic in Phase 3 |

Full reference path: `/Users/nil/DiskD/W/Djarvur/go-metadot/internal/text/justify.go`

Text wrapping is applied during graph construction (Phase 3), not parsing. Model stores raw text; wrapping happens at render time based on `LineLength` configuration.

### Error Handling

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Context | Rich context | Include line numbers, surrounding text, suggestions |
| Format | Terminal formatted | Human-readable, not JSON |

### Testing

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Test style | Individual test functions | User preference (not table-driven) |
| Coverage | 75% minimum | Per QUAL-04 |

### Phase 1 CLI

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Scope | Simple main.go | Calls parser, prints model — for development testing |
| Full CLI | Phase 6 | Flags, help, error handling deferred |

---

## Specifics

### C4-PlantUML Color Reference

Base defaults from C4-PlantUML:
- `$ELEMENT_FONT_COLOR ?= "#FFFFFF"`
- `$ARROW_COLOR ?= "#666666"`
- `$BOUNDARY_COLOR ?= "#444444"`

Type-specific colors to be defined as constants (research C4-PlantUML for exact values):
- Person
- System / System_Ext
- Container / Container_Ext
- Component / Component_Ext
- Db / Db_Ext
- Queue / Queue_Ext

### TOML Schema Reference

```toml
[properties]
name = "Project Name"
description = "Project description"
color = "transparent"
style = "none"
border = "transparent"
edges = "straight"
lineLength = 0               # 0 = auto wrap
expanded = ["system1", "box1"]

[section_name]
type = "system"              # UnitType discriminator
name = "Display Name"
description = "Description"
technology = "Go, PostgreSQL" # NOT for person types
color = "transparent"
style = "none"
border = "transparent"
edges = "straight"           # inherits if not set
width = 0                    # 0 = auto
height = 0                   # 0 = auto
expanded = ["subunit1"]
link = { "target" = { reverse = false, equal = false, color = "black", style = "solid", technology = "HTTP", description = "API calls", labelPosition = "middle" } }
linkFrom = { "source" = { ... } }

[section_name.subunit]
# Same structure, nested
```

---

## Code Context

### Existing Files

| File | Status | Notes |
| ---- | ------ | ----- |
| `go.mod` | Exists | Currently `go 1.25.1` — must update to `1.26.1` |
| `.golangci.yml` | Configured | v2 format, sensible disables |
| `.mise.toml` | Not present | Must create with test/lint tasks |

### Dependencies (Phase 1)

| Package | Version | Purpose |
| ------- | ------- | ------- |
| pelletier/go-toml | v2 (latest) | TOML parsing |
| mitchellh/go-wordwrap | latest | Text wrapping (used in Phase 3, add now) |

### Integration Points

- Parser output feeds into validator (Phase 2)
- Model types used by all downstream phases
- Error formatting pattern establishes precedent for whole project

### Patterns to Follow

- `UnitType` as custom type with `String()` method
- Exported style constants in `model/` package
- Parser returns `(*Model, error)` — errors wrap position info

---

## Deferred

Ideas captured but not for Phase 1:

| Idea | Phase | Notes |
| ---- | ----- | ----- |
| Person display on C2/C3 diagrams | Phase 3 | View generation shows linked persons |
| Expanded cluster labeling | Phase 3 | Cluster shows parent unit info (name, description, technology) |
| Width/height rendering | Phase 3 | Unit.Width/Height used for node sizing when > 0 |
| Watch mode | v2 | Auto-regenerate on file change |
| Themes | v2 | Predefined color schemes |
| PNG/PDF output | v2 | Additional formats |

---

*Context created: 2026-03-09*
*Last updated: 2026-03-24 — Added containerBox/componentBox types, unified colors per C4 level*
