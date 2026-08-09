# Phase 1: Foundation & Model - Research

**Researched:** 2026-03-09
**Domain:** Go development environment, TOML parsing, C4 domain model types
**Confidence:** MEDIUM (web search rate-limited, relied on direct documentation fetches)

## Summary

Phase 1 establishes the development environment and core domain model for C4Drill. The primary technical challenges are: (1) setting up mise tasks for Go development with sandboxed golangci-lint, (2) defining type-discriminated domain types that can represent nested C4 units, (3) implementing TOML parsing for arbitrarily nested structures using pelletier/go-toml v2.

**Primary recommendation:** Use pelletier/go-toml v2's contextualized error handling for rich error messages. Define UnitType as a string-based custom type with a `String()` method. Use flat structs with discriminator field rather than interface-based polymorphism for simpler TOML unmarshaling.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**Project Structure:**
| Decision | Choice | Rationale |
|----------|--------|-----------|
| Layout | `cmd/` + `internal/` | Standard Go CLI pattern |
| Package organization | Pipeline stages | `model/`, `parser/`, `validator/`, `view/`, `graph/`, `render/` |
| Module path | `github.com/Djarvur/c4drill` | User specified |

**Domain Types:**
| Decision | Choice | Rationale |
|----------|--------|-----------|
| Type representation | Type discriminator | `UnitType` enum field on struct |
| Unit struct | Flat struct | All fields at top level, no nested style/location objects |
| Level-specific types | Yes | C1, C2, C3 have distinct type names |

**Unit types by level:**

| Level | Types |
|-------|-------|
| **C1 (Context)** | person, personExternal, system, systemExternal, db, dbExternal, queue, queueExternal, box |
| **C2 (Containers)** | container, containerDb, containerQueue |
| **C3 (Components)** | component, componentDb, componentQueue |

**Containment rules:**

| Container | Can Contain |
|-----------|-------------|
| `system` | C2 types (container, containerDb, containerQueue) |
| `container` | C3 types (component, componentDb, componentQueue) |
| `box` (at C1) | C1 types only |
| `box` (at C2) | C2 types only |
| `box` (at C3) | C3 types only |

**Link Model:**
```go
type Link struct {
    Target string
    Arrow  ArrowDirection  // Forward, Reverse, Bidirectional, None
    Rank   RankDirection   // Forward, Reverse, Equal
    Color  string
    Style  string
}
```

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Arrow vs Rank | Separate fields | Orthogonal concerns (visual vs layout) |
| Storage | Two slices | `Links []Link` and `LinksFrom []Link` -- mirrors TOML schema |
| Target validation | Separate pass | Clean separation, Phase 2 responsibility |

**Style & Defaults:**
| Decision | Choice | Rationale |
|----------|--------|-----------|
| Base defaults | C4-PlantUML | Familiar to C4 users, proven palette |
| Type-specific colors | Exported constants | Configurable if needed |
| Style inheritance | Mixed | `edges` cascades; `color`, `style`, `border` -- no inheritance |

**C4-PlantUML defaults:**
- Font color: `#FFFFFF` (white)
- Arrow color: `#666666` (gray)
- Boundary color: `#444444` (dark gray)

**TOML Parsing:**
| Decision | Choice | Rationale |
|----------|--------|-----------|
| Library | pelletier/go-toml | Recommended, well-maintained |
| Mapping | Direct struct unmarshaling | If possible; intermediate map otherwise |

**Error Handling:**
| Decision | Choice | Rationale |
|----------|--------|-----------|
| Context | Rich context | Include line numbers, surrounding text, suggestions |
| Format | Terminal formatted | Human-readable, not JSON |

**Testing:**
| Decision | Choice | Rationale |
|----------|--------|-----------|
| Test style | Individual test functions | User preference (not table-driven) |
| Coverage | 75% minimum | Per QUAL-04 |

**Phase 1 CLI:**
| Decision | Choice | Rationale |
|----------|--------|-----------|
| Scope | Simple main.go | Calls parser, prints model -- for development testing |
| Full CLI | Phase 6 | Flags, help, error handling deferred |

### Claude's Discretion

(None explicitly stated -- proceed with standard patterns)

### Deferred Ideas (OUT OF SCOPE)

| Idea | Phase | Notes |
|------|-------|-------|
| Person display on C2/C3 diagrams | Phase 3 | View generation shows linked persons |
| Watch mode | v2 | Auto-regenerate on file change |
| Themes | v2 | Predefined color schemes |
| PNG/PDF output | v2 | Additional formats |
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| DEVI-01 | Go version updated to 1.26.1 | Go module setup, go.mod update |
| DEVI-02 | Mise config includes test tasks | Mise task patterns below |
| DEVI-03 | Mise config includes lint tasks | Mise task patterns below |
| DEVI-04 | Mise installs golangci-lint v2 into sandbox | Mise tools configuration |
| DEVI-05 | Modern Go plugin loaded before development | Process requirement, not code |
| INPT-01 | CLI accepts path to TOML input file | cmd/ main.go with os.Args |
| INPT-02 | Parser handles nested unit definitions with arbitrary depth | go-toml nested struct patterns |
| INPT-03 | Parser extracts properties section | Properties struct definition |
| INPT-04 | Parser extracts context-level units | Unit struct with UnitType discriminator |
| INPT-05 | Parser handles external variants | UnitType enum includes all variants |
| INPT-06 | Parser extracts link and linkFrom definitions | Link struct, map-based unmarshaling |
| INPT-07 | Parser handles expanded list | []string field on Properties and Unit |
| TYPE-01 | System defines person type | model.UnitType enum |
| TYPE-02 | System defines personExternal type | model.UnitType enum |
| TYPE-03 | System defines system type | model.UnitType enum |
| TYPE-04 | System defines systemExternal type | model.UnitType enum |
| TYPE-05 | System defines db and dbExternal types | model.UnitType enum |
| TYPE-06 | System defines queue and queueExternal types | model.UnitType enum |
| TYPE-07 | System defines box type | model.UnitType enum |
| TYPE-08 | Link object defines target, reverse, equal, color, style | model.Link struct |
| QUAL-01 | All lint errors fixed before commit | golangci-lint v2 config |
| QUAL-02 | Lint config not adjusted to silence errors | .golangci.yml exists, minimal disables |
| QUAL-03 | nolint directives require explicit confirmation | Process requirement |
| QUAL-04 | Minimum 75% test coverage | go test -cover |
| QUAL-05 | Coverage enforced in CI/quality gate | Process requirement |
</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| pelletier/go-toml | v2.x | TOML parsing | Well-maintained, stdlib-like API, contextualized errors |
| golangci-lint | v2.x | Linting | Industry standard, comprehensive linter set |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| (stdlib only) | - | - | Phase 1 uses no external dependencies beyond go-toml |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| pelletier/go-toml | BurntSushi/toml | BurntSushi is older, less maintained; go-toml v2 has better error context |
| flat struct | interface-based | Interfaces complicate TOML unmarshaling significantly |

**Installation:**
```bash
go get github.com/pelletier/go-toml/v2@latest
```

## Architecture Patterns

### Recommended Project Structure
```
c4drill/
├── cmd/
│   └── c4drill/
│       └── main.go          # CLI entry point (simple for Phase 1)
├── internal/
│   ├── model/
│   │   ├── unit.go          # UnitType, Unit struct definitions
│   │   ├── link.go          # Link struct, ArrowDirection, RankDirection
│   │   ├── properties.go    # Properties struct for root-level config
│   │   └── colors.go        # C4-PlantUML color constants
│   └── parser/
│       ├── parser.go        # Parse function, Model struct
│       └── parser_test.go   # Individual test functions
├── go.mod
├── go.sum
├── .golangci.yml            # Already exists
└── .mise.toml               # To be created
```

### Pattern 1: Type Discriminator with Flat Struct
**What:** Single struct with `UnitType` field to distinguish types, all properties at top level
**When to use:** When parsing heterogeneous types from structured format (TOML/JSON)
**Example:**
```go
// Source: CONTEXT.md decision, Go idiom
package model

// UnitType discriminator for C4 element types
type UnitType string

const (
    // C1 Context level
    TypePerson         UnitType = "person"
    TypePersonExternal UnitType = "personExternal"
    TypeSystem         UnitType = "system"
    TypeSystemExternal UnitType = "systemExternal"
    TypeDb             UnitType = "db"
    TypeDbExternal     UnitType = "dbExternal"
    TypeQueue          UnitType = "queue"
    TypeQueueExternal  UnitType = "queueExternal"
    TypeBox            UnitType = "box"

    // C2 Container level
    TypeContainer       UnitType = "container"
    TypeContainerDb     UnitType = "containerDb"
    TypeContainerQueue  UnitType = "containerQueue"

    // C3 Component level
    TypeComponent       UnitType = "component"
    TypeComponentDb     UnitType = "componentDb"
    TypeComponentQueue  UnitType = "componentQueue"
)

func (t UnitType) String() string {
    return string(t)
}

// Unit represents a C4 model element
type Unit struct {
    Type        UnitType          `toml:"type"`
    Name        string            `toml:"name"`
    Description string            `toml:"description"`
    Color       string            `toml:"color"`
    Style       string            `toml:"style"`
    Border      string            `toml:"border"`
    Edges       string            `toml:"edges"`
    Expanded    []string          `toml:"expanded"`
    Links       map[string]Link   `toml:"link"`
    LinksFrom   map[string]Link   `toml:"linkFrom"`
    Subunits    map[string]*Unit  `toml:",inline"`  // Nested tables
}
```

### Pattern 2: Nested TOML Table Unmarshaling
**What:** go-toml v2 unmarshals nested tables into map or struct fields
**When to use:** When parsing TOML with arbitrary nesting depth
**Example:**
```go
// Source: go-toml v2 README
package parser

import (
    "github.com/pelletier/go-toml/v2"
)

type Model struct {
    Properties Properties           `toml:"properties"`
    Units      map[string]*Unit     `toml:",inline"` // Top-level units
}

func Parse(data []byte) (*Model, error) {
    var model Model
    if err := toml.Unmarshal(data, &model); err != nil {
        // go-toml v2 returns DecodeError with context
        var decodeErr *toml.DecodeError
        if errors.As(err, &decodeErr) {
            // Format with line numbers and context
            return nil, fmt.Errorf("TOML parsing error:\n%s", decodeErr)
        }
        return nil, err
    }
    return &model, nil
}
```

### Pattern 3: Mise Task Configuration
**What:** Define tasks in .mise.toml for test, lint, and tool installation
**When to use:** Go projects needing reproducible dev environment
**Example:**
```toml
# .mise.toml
[tools]
go = "1.26"
golangci-lint = "2"

[tasks.test]
description = "Run Go tests with coverage"
run = "go test -cover ./..."

[tasks.lint]
description = "Run golangci-lint"
run = "golangci-lint run ./..."

[tasks.lint-fix]
description = "Run golangci-lint with auto-fix"
run = "golangci-lint run --fix ./..."
```

### Anti-Patterns to Avoid

- **Interface-based polymorphism for TOML:** TOML unmarshaling into interfaces produces `map[string]interface{}`, not concrete types. Use discriminator pattern instead.
- **Ignoring go-toml DecodeError context:** The library provides rich error context with line numbers -- use it for user-friendly errors.
- **Case-sensitive field matching:** go-toml v2 does case-insensitive matching like encoding/json. Don't rely on casing to differentiate fields.
- **Hand-rolling recursive parsing:** Use map[string]*Unit for arbitrary depth; go-toml handles recursion automatically.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| TOML parsing | Custom parser | pelletier/go-toml v2 | Handles edge cases, provides error context |
| Error formatting with context | Custom line extraction | toml.DecodeError | Built-in with line numbers and caret |
| Color constants | Guess at values | C4-PlantUML source values | Proven, familiar to C4 users |
| Linting rules | Custom linter config | Existing .golangci.yml | Already configured with sensible defaults |

**Key insight:** go-toml v2's DecodeError type provides formatted error output with line numbers and context. Use it directly rather than building custom error formatting.

## Common Pitfalls

### Pitfall 1: Nested Struct Unmarshaling into Wrong Type
**What goes wrong:** Defining nested units as `[]Unit` or fixed struct instead of `map[string]*Unit`
**Why it happens:** TOML tables are key-value maps, not arrays
**How to avoid:** Always use `map[string]*Unit` for nested tables where keys are section names
**Warning signs:** "cannot decode TOML array into map" errors

### Pitfall 2: Missing Pointer for Recursive Types
**What goes wrong:** `map[string]Unit` causes infinite type definition issues
**Why it happens:** Go needs pointer for recursive type definitions
**How to avoid:** Use `map[string]*Unit` for self-referencing structures
**Warning signs:** "invalid recursive type" compiler error

### Pitfall 3: Inline Table vs Regular Table Confusion
**What goes wrong:** TOML `{ key = value }` syntax (inline) vs `[section]` syntax (regular)
**Why it happens:** Both look similar but unmarshal differently
**How to avoid:** Use regular tables for nested units; inline for link definitions (per CONTEXT.md schema)
**Warning signs:** Unmarshal errors on nested definitions

### Pitfall 4: go-toml v1 vs v2 API Differences
**What goes wrong:** Using v1 patterns (Tree, toml.Unmarshaler interface)
**Why it happens:** v2 removed some v1 features
**How to avoid:** Import `github.com/pelletier/go-toml/v2` explicitly; don't use `toml.Tree`
**Warning signs:** "undefined: toml.Tree" errors

### Pitfall 5: Mise Global Install vs Sandbox
**What goes wrong:** Installing golangci-lint globally conflicts with other projects
**Why it happens:** Different projects need different tool versions
**How to avoid:** Use mise's [tools] section to install per-project; mise isolates in sandbox
**Warning signs:** Version mismatch errors, CI failures

## Code Examples

Verified patterns from official sources:

### Basic go-toml v2 Unmarshaling
```go
// Source: https://github.com/pelletier/go-toml
import "github.com/pelletier/go-toml/v2"

type MyConfig struct {
    Version int
    Name    string
    Tags    []string
}

doc := `
version = 2
name = "go-toml"
tags = ["go", "toml"]
`

var cfg MyConfig
err := toml.Unmarshal([]byte(doc), &cfg)
```

### Nested Table Unmarshaling
```go
// Source: https://github.com/pelletier/go-toml
doc := `
age = 45
fruits = ["apple", "pear"]

[my-variables]
first = 1
second = 0.2

[my-variables.b]
bfirst = 123
`

var Document struct {
    Age int
    Fruits []string

    Myvariables struct {
        First  int
        Second float64

        B struct {
            Bfirst int
        }
    } `toml:"my-variables"`
}

err := toml.Unmarshal([]byte(doc), &Document)
```

### DecodeError with Context
```go
// Source: https://github.com/pelletier/go-toml
// go-toml v2 returns DecodeError with human-readable context:
// 1| [server]
// 2| path = 100
//  |        ~~~ cannot decode TOML integer into struct field
// 3| port = 50

import (
    "errors"
    "github.com/pelletier/go-toml/v2"
)

func parseWithErrorContext(data []byte, v interface{}) error {
    err := toml.Unmarshal(data, v)
    if err == nil {
        return nil
    }

    var decodeErr *toml.DecodeError
    if errors.As(err, &decodeErr) {
        // decodeErr.String() returns formatted error with context
        return fmt.Errorf("parse error:\n%s", decodeErr)
    }
    return err
}
```

### Mise Task Configuration
```toml
# Source: https://github.com/jdx/mise README
# .mise.toml

[tools]
terraform = "1"
aws-cli = "2"

[env]
TF_WORKSPACE = "development"

[tasks.plan]
description = "Run terraform plan"
run = "terraform plan"

[tasks.validate]
description = "Validate configuration"
depends = ["plan"]
run = "terraform validate"
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| go-toml v1 Tree API | go-toml v2 stdlib-like API | v2 release | Simpler API, no Tree manipulation |
| toml.Unmarshaler interface | encoding.TextUnmarshaler | v2 release | Standard Go interface |
| Global tool installs | mise sandboxed installs | Modern practice | Per-project version isolation |
| Table-driven tests | Individual test functions | User preference | More explicit test names |

**Deprecated/outdated:**
- `toml.Tree` API: Removed in v2, use struct unmarshaling instead
- `toml.Unmarshaler` interface: Dropped in v2, use `encoding.TextUnmarshaler` instead
- `default` struct tag: Dropped in v2, pre-fill structs with defaults instead

## C4-PlantUML Color Values

Source: https://github.com/plantuml-stdlib/C4-PlantUML (C4.puml, C4_Context.puml, C4_Container.puml, C4_Component.puml)

### Base Colors
| Constant | Hex Value | Usage |
|----------|-----------|-------|
| ELEMENT_FONT_COLOR | `#FFFFFF` | Text color on elements |
| ARROW_COLOR | `#666666` | Default arrow/link color |
| BOUNDARY_COLOR | `#444444` | Boundary/border color |

### C1 Context Level
| Element | Background | Border | Font |
|---------|------------|--------|------|
| Person | `#08427B` | `#073B6F` | `#FFFFFF` |
| External Person | `#686868` | `#8A8A8A` | `#FFFFFF` |
| System | `#1168BD` | `#3C7FC0` | `#FFFFFF` |
| External System | `#999999` | `#8A8A8A` | `#FFFFFF` |

### C2 Container Level
| Element | Background | Border | Font |
|---------|------------|--------|------|
| Container | `#438DD5` | `#3C7FC0` | `#FFFFFF` |
| External Container | `#B3B3B3` | `#A6A6A6` | `#FFFFFF` |

### C3 Component Level
| Element | Background | Border | Font |
|---------|------------|--------|------|
| Component | `#85BBF0` | `#78A8D8` | `#000000` |
| External Component | `#CCCCCC` | `#BFBFBF` | `#000000` |

**Note:** Database (Db) and Queue types use the same colors as their parent level (Container/Component).

## Open Questions

1. **Arbitrary Nesting Depth Implementation**
   - What we know: go-toml can unmarshal nested tables into map[string]*Unit
   - What's unclear: Whether `toml:",inline"` tag correctly captures arbitrary nested tables at root level
   - Recommendation: Test with multi-level nesting early; may need custom unmarshaling if inline tag insufficient

2. **Link Map Unmarshaling**
   - What we know: TOML inline tables like `{ "target" = { ... } }` should unmarshal to map[string]Link
   - What's unclear: Exact struct tag configuration needed
   - Recommendation: Prototype link parsing early in Phase 1

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing package (stdlib) |
| Config file | None -- Go uses *_test.go files |
| Quick run command | `go test ./...` |
| Full suite command | `go test -cover -race ./...` |

### Phase Requirements -> Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| DEVI-01 | Go version 1.26.1 | Manual | Check go.mod | go.mod exists, needs version update |
| DEVI-02 | Mise test task | Integration | `mise test` | No .mise.toml -- Wave 0 |
| DEVI-03 | Mise lint task | Integration | `mise lint` | No .mise.toml -- Wave 0 |
| DEVI-04 | golangci-lint in sandbox | Integration | `mise install` | No .mise.toml -- Wave 0 |
| INPT-01 | CLI accepts TOML path | Unit | `go test ./cmd/...` | No cmd/ -- Wave 0 |
| INPT-02 | Nested unit parsing | Unit | `go test ./internal/parser/...` | No parser/ -- Wave 0 |
| INPT-03 | Properties section | Unit | `go test ./internal/parser/...` | No parser/ -- Wave 0 |
| INPT-04 | Context-level units | Unit | `go test ./internal/parser/...` | No parser/ -- Wave 0 |
| INPT-05 | External variants | Unit | `go test ./internal/parser/...` | No parser/ -- Wave 0 |
| INPT-06 | Link definitions | Unit | `go test ./internal/parser/...` | No parser/ -- Wave 0 |
| INPT-07 | Expanded list | Unit | `go test ./internal/parser/...` | No parser/ -- Wave 0 |
| TYPE-01 to TYPE-08 | Type definitions | Unit | `go test ./internal/model/...` | No model/ -- Wave 0 |
| QUAL-04 | 75% coverage | Quality gate | `go test -cover ./...` | N/A |

### Sampling Rate
- **Per task commit:** `go test ./...`
- **Per wave merge:** `go test -cover -race ./...`
- **Phase gate:** Full suite green with 75%+ coverage before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `.mise.toml` -- mise tasks for test, lint, tool installation
- [ ] `cmd/c4drill/main.go` -- CLI entry point
- [ ] `internal/model/unit.go` -- UnitType and Unit definitions
- [ ] `internal/model/link.go` -- Link struct and direction types
- [ ] `internal/model/properties.go` -- Properties struct
- [ ] `internal/model/colors.go` -- Color constants
- [ ] `internal/parser/parser.go` -- Parse function and Model struct
- [ ] `internal/parser/parser_test.go` -- Individual test functions
- [ ] `go.mod` update -- Change `go 1.25.1` to `go 1.26.1`

## Sources

### Primary (HIGH confidence)
- [go-toml v2 README](https://github.com/pelletier/go-toml) - Unmarshaling patterns, DecodeError usage
- [C4-PlantUML source](https://github.com/plantuml-stdlib/C4-PlantUML) - Color values from C4.puml, C4_Context.puml, C4_Container.puml, C4_Component.puml
- [mise README](https://github.com/jdx/mise) - Task configuration patterns

### Secondary (MEDIUM confidence)
- [CONTEXT.md](file://.planning/1-CONTEXT.md) - User decisions and constraints
- [REQUIREMENTS.md](file://.planning/REQUIREMENTS.md) - Phase 1 requirement IDs
- [.golangci.yml](file://.golangci.yml) - Existing lint configuration

### Tertiary (LOW confidence)
- None -- all critical information from primary sources

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - go-toml v2 is well-documented, mise is current
- Architecture: HIGH - Patterns derived from official docs and user decisions
- Pitfalls: MEDIUM - Based on go-toml migration docs and common Go patterns
- Color values: HIGH - Directly extracted from C4-PlantUML source

**Research date:** 2026-03-09
**Valid until:** 30 days (stable libraries, Go ecosystem)
