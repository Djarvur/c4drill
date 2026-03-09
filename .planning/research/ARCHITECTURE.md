# Architecture Research: C4 Diagram Generation Tools

**Domain:** CLI tool for generating C4 architecture diagrams from structured input
**Researched:** 2026-03-09
**Confidence:** HIGH (based on Structurizr reference implementation, go-graphviz library docs, and C4 model specification)

## Standard Architecture

### System Overview

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              CLI Interface                                   │
│                         (command-line parsing)                               │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                              Input Layer                                     │
├─────────────────────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐          │
│  │  TOML Parser    │───▶│  Model Builder  │───▶│   Validator     │          │
│  │  (go-toml)      │    │  (AST → Model)  │    │  (integrity)    │          │
│  └─────────────────┘    └─────────────────┘    └─────────────────┘          │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                              Core Model                                      │
├─────────────────────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐          │
│  │    Workspace    │    │    Elements     │    │  Relationships  │          │
│  │   (root scope)  │    │  (Person,       │    │    (Links)      │          │
│  │                 │    │   System, etc)  │    │                 │          │
│  └─────────────────┘    └─────────────────┘    └─────────────────┘          │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                              View Layer                                      │
├─────────────────────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐          │
│  │  View Generator │───▶│  Graph Builder  │───▶│  DOT Renderer   │          │
│  │  (scope filter) │    │  (node/edge     │    │  (GraphViz      │          │
│  │                 │    │   construction) │    │   DOT output)   │          │
│  └─────────────────┘    └─────────────────┘    └─────────────────┘          │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                              Export Layer                                    │
├─────────────────────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐          │
│  │   DOT Export    │    │   SVG Render    │    │  File Writer    │          │
│  │   (.dot files)  │    │  (go-graphviz)  │    │  (output to     │          │
│  │                 │    │                 │    │   disk)         │          │
│  └─────────────────┘    └─────────────────┘    └─────────────────┘          │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Component Responsibilities

| Component | Responsibility | Implementation |
|-----------|----------------|----------------|
| CLI Interface | Parse command-line arguments, handle help/version flags, invoke pipeline | Go `flag` or `cobra` package |
| TOML Parser | Read input file, parse TOML syntax into structured data | `github.com/pelletier/go-toml/v2` |
| Model Builder | Transform parsed TOML into domain model (units, links, properties) | Custom Go structs with Unmarshal hooks |
| Validator | Enforce C4 rules: reference integrity, subunit constraints, type rules | Validation functions on model structs |
| Core Model | Domain types: Workspace, Unit (person/system/db/queue/box), Link | Go structs with JSON/TOML tags |
| View Generator | Filter model by scope (C1/C2/C3), handle collapsed/expanded state | View scope logic, expansion tracking |
| Graph Builder | Convert filtered view to graph structure (nodes, edges, clusters) | Internal graph representation |
| DOT Renderer | Generate GraphViz DOT language from graph structure | String builder / template |
| SVG Render | Invoke go-graphviz to render DOT to SVG | `github.com/goccy/go-graphviz` |
| File Writer | Write output files to correct locations, create directories | Go `os` package |

## Recommended Project Structure

```
c4drill/
├── cmd/
│   └── c4drill/
│       └── main.go           # CLI entry point
├── internal/
│   ├── model/                # Core domain types
│   │   ├── workspace.go      # Workspace container
│   │   ├── unit.go           # Unit types (person, system, db, etc.)
│   │   ├── link.go           # Relationship/link types
│   │   └── properties.go     # Styling properties
│   ├── parser/               # Input handling
│   │   ├── toml.go           # TOML parsing logic
│   │   └── unmarshal.go      # Custom unmarshaling
│   ├── validate/             # Model validation
│   │   ├── validator.go      # Validation entry point
│   │   ├── references.go     # Reference integrity checks
│   │   └── constraints.go    # Subunit/type constraints
│   ├── view/                 # View generation
│   │   ├── generator.go      # View scope filtering
│   │   ├── context.go        # C1 (Context) view
│   │   ├── container.go      # C2 (Container) view
│   │   └── component.go      # C3 (Component) view
│   ├── graph/                # Graph construction
│   │   ├── builder.go        # Graph from model
│   │   ├── node.go           # Node creation
│   │   ├── edge.go           # Edge creation
│   │   └── cluster.go        # Subgraph/cluster handling
│   ├── render/               # Output rendering
│   │   ├── dot.go            # DOT format generation
│   │   ├── svg.go            # SVG via go-graphviz
│   │   └── shapes.go         # C4 shape definitions
│   └── output/               # File output
│       ├── writer.go         # File writing logic
│       └── paths.go          # Output path resolution
├── pkg/                      # (empty for v1 - no public API)
├── go.mod
├── go.sum
└── README.md
```

### Structure Rationale

- **cmd/c4drill/:** Standard Go convention for CLI entry point; keeps main.go minimal
- **internal/model/:** Core domain types first; these have no dependencies on other internal packages
- **internal/parser/:** Isolated input handling; can be replaced (e.g., add JSON support) without touching model
- **internal/validate/:** Validation is a separate concern from parsing; enables standalone validation commands
- **internal/view/:** View generation is distinct from model; handles scope filtering and expansion logic
- **internal/graph/:** Graph representation is an intermediate format between model and DOT output
- **internal/render/:** Rendering is output-format specific; DOT and SVG are separate concerns
- **internal/output/:** File system operations isolated for testability

## Architectural Patterns

### Pattern 1: Pipeline Architecture

**What:** Sequential processing stages where each stage transforms data and passes to the next.

**When to use:** CLI tools with clear input → processing → output flow (exactly C4Drill's case).

**Trade-offs:**
- Pros: Simple to understand, easy to test each stage independently, clear data flow
- Cons: Less flexible for complex workflows, all-or-nothing processing

**Example:**
```go
func Run(inputPath string, outputPath string) error {
    // Stage 1: Parse
    raw, err := parser.ParseFile(inputPath)
    if err != nil {
        return fmt.Errorf("parse: %w", err)
    }

    // Stage 2: Build Model
    workspace, err := model.Build(raw)
    if err != nil {
        return fmt.Errorf("model: %w", err)
    }

    // Stage 3: Validate
    if err := validate.Check(workspace); err != nil {
        return fmt.Errorf("validate: %w", err)
    }

    // Stage 4: Generate Views
    views := view.GenerateAll(workspace)

    // Stage 5: Render & Output
    for _, v := range views {
        dot := render.ToDOT(v)
        svg, err := render.ToSVG(dot)
        if err != nil {
            return fmt.Errorf("render: %w", err)
        }
        output.Write(outputPath, v.Name, dot, svg)
    }

    return nil
}
```

### Pattern 2: Domain Model with Separate Views

**What:** Single source of truth (the model) with multiple view projections that filter/transform for specific diagram scopes.

**When to use:** When the same data needs to be presented in different ways (C1 vs C2 vs C3 views, collapsed vs expanded).

**Trade-offs:**
- Pros: Consistency across views, single validation point, easy to add new view types
- Cons: View generation logic can become complex, memory overhead for multiple views

**Example:**
```go
// Model is the source of truth
type Workspace struct {
    Properties Properties
    Units      map[string]*Unit
}

// View is a projection of the model for a specific scope
type View struct {
    Scope     Scope       // C1, C2, or C3
    Expanded  []string    // Which units are expanded
    RootUnits []*Unit     // Filtered units for this scope
    Links     []*Link     // Filtered links for this scope
}

// Generate creates a view by filtering the workspace
func (w *Workspace) GenerateView(scope Scope, expanded []string) *View {
    v := &View{Scope: scope, Expanded: expanded}

    // Filter units based on scope and expansion
    for _, unit := range w.Units {
        if v.shouldInclude(unit) {
            v.RootUnits = append(v.RootUnits, unit)
        }
    }

    // Filter links to only those between included units
    // ...

    return v
}
```

### Pattern 3: Double-Dispatch for Shape Rendering

**What:** Each unit type knows how to render itself, called via interface dispatch.

**When to use:** When different entity types need different visual representations (person vs system vs database shapes).

**Trade-offs:**
- Pros: Extensible (add new types easily), logic co-located with types
- Cons: Rendering logic spread across model types, harder to see all rendering in one place

**Example:**
```go
type Renderer interface {
    RenderDOT() string
    Shape() string
}

func (p *Person) RenderDOT() string {
    return fmt.Sprintf(`"%s" [label="%s\n%s", shape=record, ...]`,
        p.ID, p.Name, p.Description)
}

func (p *Person) Shape() string { return "person" }

func (s *System) RenderDOT() string {
    if len(s.Subunits) > 0 && s.IsExpanded() {
        return s.renderAsCluster()
    }
    return s.renderAsNode()
}

func (s *System) Shape() string { return "box" }
```

## Data Flow

### Request Flow (End-to-End)

```
[TOML File]
     │
     ▼
[TOML Parser] ──parse──▶ [map[string]interface{} / RawModel]
     │
     ▼
[Model Builder] ──unmarshal──▶ [Workspace{Units, Links, Properties}]
     │
     ▼
[Validator] ──check──▶ [Workspace (validated)] or [ValidationError]
     │
     ▼
[View Generator] ──filter──▶ [View{RootUnits, Links, Scope}]
     │
     ▼
[Graph Builder] ──construct──▶ [Graph{Nodes, Edges, Clusters}]
     │
     ▼
[DOT Renderer] ──serialize──▶ [string (DOT format)]
     │
     ├──────────────────────────────┐
     ▼                              ▼
[DOT File Writer]           [SVG Renderer] ──render──▶ [SVG bytes]
     │                              │
     ▼                              ▼
[output.dot]                [SVG File Writer] ──▶ [output.svg]
```

### Key Data Flows

1. **Input → Model:** TOML file is parsed into a raw map, then unmarshaled into strongly-typed Workspace with Units and Links. The model captures the hierarchical structure (units with subunits) and all relationships.

2. **Model → View:** Workspace is filtered based on scope (C1/C2/C3) and expansion state. View contains only the units and links relevant to that specific diagram. Expanded units include their subunits as separate renderable elements.

3. **View → Graph:** View elements are converted to a graph representation with nodes (for atomic units), edges (for links), and clusters (for expanded containers). This is the intermediate format that DOT rendering consumes.

4. **Graph → Output:** Graph is serialized to DOT string format, then optionally rendered to SVG via go-graphviz. Both formats are written to the output location with appropriate file structure.

## Scaling Considerations

| Scale | Architecture Adjustments |
|-------|--------------------------|
| Small (< 50 units) | Single-pass rendering is fine; no optimization needed |
| Medium (50-200 units) | Consider lazy evaluation for views; only generate requested scope |
| Large (200+ units) | May need incremental rendering; split into multiple output files |

### Scaling Priorities

1. **First bottleneck:** Memory for large models. Mitigation: stream parsing for very large TOML files (unlikely for architecture docs, but possible).

2. **Second bottleneck:** SVG rendering time. Mitigation: generate DOT only by default, make SVG optional or parallel.

**Note:** C4Drill is a documentation tool, not a real-time system. Performance is secondary to correctness and output quality. Most architecture models will be small (< 100 units).

## Anti-Patterns

### Anti-Pattern 1: Conflating Model and View

**What people do:** Put rendering logic (DOT generation, styling) directly into model structs.

**Why it's wrong:** Violates separation of concerns. Model becomes coupled to output format. Cannot support multiple output formats cleanly.

**Do this instead:** Keep model as pure data. Rendering logic lives in a separate `render` package that accepts model types as input.

### Anti-Pattern 2: Skipping Validation

**What people do:** Assume TOML input is always valid; proceed directly to rendering.

**Why it's wrong:** Invalid references, circular dependencies, or constraint violations produce confusing errors later in the pipeline or incorrect diagrams.

**Do this instead:** Always validate after parsing. Fail fast with clear error messages pointing to the problematic TOML section.

### Anti-Pattern 3: Global State for Expansion

**What people do:** Use a global variable or singleton to track which units are expanded.

**Why it's wrong:** Makes testing difficult, prevents parallel view generation, creates hidden dependencies.

**Do this instead:** Pass expansion state explicitly to view generation functions. Make it part of the View configuration.

### Anti-Pattern 4: Direct DOT String Concatenation

**What people do:** Build DOT output by concatenating strings with `+` or `fmt.Sprintf`.

**Why it's wrong:** Error-prone with escaping (quotes in labels), hard to maintain, difficult to test individual elements.

**Do this instead:** Use `strings.Builder` with proper escaping functions, or template-based generation with clear node/edge formatting.

## Integration Points

### External Libraries

| Library | Integration Pattern | Notes |
|---------|---------------------|-------|
| go-toml | Unmarshal TOML file to struct | Use `toml.Unmarshal` with custom struct tags |
| go-graphviz | Render DOT to SVG | Create graphviz instance, parse DOT, render to SVG bytes |

### Internal Boundaries

| Boundary | Communication | Notes |
|----------|---------------|-------|
| parser → model | Structured data (map or intermediate struct) | Parser produces raw data; model package defines types |
| model → validate | Workspace struct | Validator reads model, returns nil or error |
| validate → view | Workspace struct (validated) | View generator receives validated model |
| view → graph | View struct | Graph builder converts filtered view to graph |
| graph → render | Graph struct | Renderer accepts internal graph representation |
| render → output | string (DOT) + []byte (SVG) | Output writer handles file system |

## Build Order Implications

Based on the architecture, the recommended build order is:

1. **Phase 1: Core Model** (`internal/model/`)
   - Define Workspace, Unit, Link, Properties structs
   - No external dependencies beyond standard library
   - Enables all downstream work

2. **Phase 2: TOML Parser** (`internal/parser/`)
   - Integrate go-toml library
   - Unmarshal TOML to model structs
   - Depends on: model

3. **Phase 3: Validator** (`internal/validate/`)
   - Reference integrity checks
   - Type constraint validation
   - Depends on: model

4. **Phase 4: View Generator** (`internal/view/`)
   - Scope filtering (C1/C2/C3)
   - Expansion handling
   - Depends on: model

5. **Phase 5: Graph Builder** (`internal/graph/`)
   - Node/edge/cluster construction
   - Depends on: model, view

6. **Phase 6: DOT Renderer** (`internal/render/dot.go`)
   - DOT format generation
   - Depends on: graph

7. **Phase 7: SVG Renderer** (`internal/render/svg.go`)
   - Integrate go-graphviz
   - Depends on: graph, DOT renderer (for intermediate)

8. **Phase 8: Output Writer** (`internal/output/`)
   - File writing logic
   - Path resolution for expanded structures
   - Depends on: render

9. **Phase 9: CLI** (`cmd/c4drill/`)
   - Command-line interface
   - Pipeline orchestration
   - Depends on: all above

## Sources

- **Structurizr Architecture:** https://structurizr.com/help/model (building blocks: Person, Software System, Container, Component)
- **Structurizr Views:** https://structurizr.com/help/views (view types and scoping)
- **Structurizr Java Modules:** https://github.com/structurizr/java (reference module structure: core, client, export, dsl)
- **go-graphviz Library:** https://github.com/goccy/go-graphviz (pure Go Graphviz via WASM, supports DOT/SVG output)
- **go-toml Library:** https://github.com/pelletier/go-toml (TOML v1.0.0 parser for Go)
- **C4 Model:** https://c4model.com/ (C1-C4 abstraction levels, core concepts)

---
*Architecture research for: C4 Diagram Generation CLI Tool*
*Researched: 2026-03-09*
