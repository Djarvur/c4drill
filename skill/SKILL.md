---
name: c4drill-toml
description: Generate valid C4Drill TOML architecture definitions
version: 1.0.0
---

# C4Drill TOML Skill

**Purpose:** Enable AI assistants to generate valid C4Drill TOML architecture definitions without prior training.

**Target:** Generic LLM (model-agnostic). Works with Claude Code, Cursor, OpenCode, or any AI with skill support.

---

## Quick Reference

### Unit Types (16 total)

**C1 Context Level:**
- `person` - Actor/user
- `personExternal` - External actor
- `system` - Software system
- `systemExternal` - External system
- `db` - Database
- `dbExternal` - External database
- `queue` - Message queue
- `queueExternal` - External queue
- `box` - Grouping container

**C2 Container Level:**
- `container` - Container within system
- `containerDb` - Container database
- `containerQueue` - Container queue

**C3 Component Level:**
- `component` - Component within container
- `componentDb` - Component database
- `componentQueue` - Component queue

### Required Fields

**Every unit must have:**
```toml
type = "<unit_type>"  # Required
name = "Display Name"  # Required
```

**Every TOML file must have:**
```toml
[properties]
name = "Architecture Name"
```

---

## Schema Reference

### [properties] Section

Root-level section defining global settings:

```toml
[properties]
name = "My Architecture"        # Required: Diagram name
description = "Description"     # Optional: Project description
color = "#E3F2FD"               # Optional: Default background color
style = "filled"                # Optional: Default visual style
border = "#1565C0"              # Optional: Default border color
edges = "spline"                # Optional: Edge routing (straight|spline|square)
lineLength = 40                 # Optional: Max line length before wrap (0=auto)
expanded = ["payments"]         # Optional: Units expanded by default
```

### Unit Definition

Each unit is a TOML section. Section name becomes the identifier:

```toml
[section_name]
type = "system"                 # Required: Unit type
name = "Display Name"           # Required: Display name
description = "What it does"    # Optional: Brief description
technology = "Go, PostgreSQL"   # Optional: Tech stack (not for person types)
color = "#E3F2FD"               # Optional: Background color override
style = "filled"                # Optional: Visual style override
border = "#1565C0"              # Optional: Border color override
edges = "spline"                # Optional: Edge style (cascades to subunits)
width = 300                     # Optional: Explicit width (0=auto)
height = 200                    # Optional: Explicit height (0=auto)
expanded = ["subunit1"]         # Optional: Subunits expanded by default
link = { ... }                  # Optional: Outgoing links
linkFrom = { ... }              # Optional: Incoming links
```

### Nesting (C2/C3 Diagrams)

Use dotted notation for nested units. Only `system` and `box` types can have subunits:

```toml
[mainapp]                       # C1: System
type = "system"
name = "Main Application"

[mainapp.api]                   # C2: Container
type = "container"
name = "API Service"

[mainapp.api.handlers]          # C3: Component
type = "component"
name = "HTTP Handlers"

[mainapp.db]                    # C2: Container database
type = "containerDb"
name = "Database"
```

### Link Syntax

Define relationships using `link` (outgoing) or `linkFrom` (incoming):

**Outgoing link (defined on source):**
```toml
[user]
type = "person"
name = "User"
link = { "webapp" = { technology = "HTTPS", description = "Browses" } }
```

**Incoming link (defined on target):**
```toml
[api]
type = "system"
name = "API Service"
linkFrom = { "webapp" = { technology = "REST", description = "Calls" } }
```

**Multiple links:**
```toml
link = {
  "webapp" = { technology = "HTTPS", description = "Browses" },
  "mobile" = { technology = "HTTPS", description = "Uses app" }
}
```

### Link Attributes

| Attribute | Values | Description |
|-----------|--------|-------------|
| `arrow` | `forward` (default), `reverse`, `bidirectional`, `none` | Arrow direction |
| `rank` | `forward`, `reverse`, `equal` | Layout ranking hint |
| `color` | `"blue"`, `"#FF5733"` | Edge color |
| `style` | `"solid"`, `"dashed"`, `"dotted"` | Line style |
| `technology` | `"HTTPS"`, `"gRPC"`, `"TCP"` | Protocol/technology label |
| `description` | `"Sends events to"` | Relationship description |
| `labelPosition` | `middle` (default), `tail`, `head` | Where label appears |

---

## Validation Rules

**Rule 1: Referenced units must exist**
```toml
# ❌ INVALID: "api" unit not defined
[user]
link = { "api" = { description = "Calls" } }

# ✅ VALID: Both units defined
[user]
type = "person"
name = "User"
link = { "api" = { description = "Calls" } }

[api]
type = "system"
name = "API"
```

**Rule 2: Units with subunits cannot have links**
```toml
# ❌ INVALID: mainapp has subunits but also has links
[mainapp]
type = "system"
link = { "external" = { description = "Calls" } }

[mainapp.api]
type = "container"

# ✅ VALID: Links on subunits only
[mainapp]
type = "system"

[mainapp.api]
type = "container"
link = { "external" = { description = "Calls" } }
```

**Rule 3: Cannot link to units with subunits**
```toml
# ❌ INVALID: Linking to mainapp which has subunits
[user]
link = { "mainapp" = { description = "Uses" } }

[mainapp]
type = "system"

[mainapp.api]
type = "container"

# ✅ VALID: Link to specific subunit
[user]
link = { "mainapp.api" = { description = "Uses" } }
```

**Rule 4: Subunits only for system and box types**
```toml
# ❌ INVALID: Database cannot have subunits
[postgres]
type = "db"

[postgres.replica]
type = "db"

# ✅ VALID: Use box for grouping
[postgres]
type = "box"
name = "Database Cluster"

[postgres.primary]
type = "db"
name = "Primary"

[postgres.replica]
type = "db"
name = "Replica"
```

---

## Prompt Patterns

Use these patterns when generating C4Drill TOML from natural language:

### Pattern 1: Basic Architecture
```
"Create a C4 diagram with: [list units and relationships]"

Example: "Create a C4 diagram with: user, web application, API, database. User calls webapp, webapp calls API, API queries database."
```

### Pattern 2: Architecture Description
```
"Generate TOML for: [architecture description]"

Example: "Generate TOML for: E-commerce platform with customers browsing products, adding to cart, and checking out. Uses PostgreSQL for storage and Stripe for payments."
```

### Pattern 3: Level Specification
```
"Create a C[C1/C2/C3] diagram for: [description]"

Example: "Create a C2 diagram for: Microservices architecture with API Gateway, User Service, Order Service, and shared database."
```

### Pattern 4: Existing Architecture
```
"Model this architecture in C4Drill TOML: [description]"

Example: "Model this architecture in C4Drill TOML: 3-tier web app with React frontend, Node.js API, and PostgreSQL database. Frontend calls API via HTTPS, API queries database."
```

### Pattern 5: Feature Addition
```
"Add [component] to the architecture: [existing TOML]"

Example: "Add caching layer to the architecture: [paste existing TOML]"
```

---

## Examples

The following examples demonstrate increasing complexity:

1. **[01-minimal.toml](examples/01-minimal.toml)** - Minimal working example (person + system + single link)
2. **[02-nested.toml](examples/02-nested.toml)** - Nested structure (C2/C3 levels)
3. **[03-links.toml](examples/03-links.toml)** - Link syntax (all attributes)
4. **[04-styling.toml](examples/04-styling.toml)** - Visual customization (colors, borders, edges)
5. **[05-ecommerce.toml](examples/05-ecommerce.toml)** - Realistic full architecture

All examples parse and validate successfully with `c4drill validate`.

---

## Generation Guidelines

When generating TOML:

1. **Always include `[properties]` with `name`** - Required field
2. **Always include `type` and `name` for each unit** - Required fields
3. **Use descriptive section names** - They become identifiers (e.g., `user_service` not `svc1`)
4. **Use dotted notation for nesting** - `parent.child` pattern
5. **Prefer `link` over `linkFrom`** - More natural (source → target)
6. **Include `technology` for systems/databases** - Shows tech stack in diagram
7. **Use `external` suffix for external units** - `systemExternal`, `dbExternal`
8. **Validate before committing** - Run `c4drill <file.toml>` to check

---

## Common Mistakes

**❌ Missing required fields:**
```toml
[user]
# Missing type and name
```

**✅ Always include:**
```toml
[user]
type = "person"
name = "User"
```

---

**❌ Linking to parent unit:**
```toml
[user]
link = { "mainapp" = {} }

[mainapp]
type = "system"

[mainapp.api]
type = "container"
```

**✅ Link to specific subunit:**
```toml
[user]
link = { "mainapp.api" = {} }

[mainapp]
type = "system"

[mainapp.api]
type = "container"
```

---

**❌ Invalid unit type for nesting:**
```toml
[postgres]
type = "db"

[postgres.replica]
type = "db"
```

**✅ Use box for grouping:**
```toml
[postgres]
type = "box"
name = "Database Cluster"

[postgres.primary]
type = "db"
name = "Primary"

[postgres.replica]
type = "db"
name = "Replica"
```

---

## Validation Command

Always validate generated TOML:

```bash
c4drill architecture.toml
```

Success = no output (generates SVG by default)
Failure = error message with line number

---

*Skill version: 1.0.0*
*C4Drill version: v1.0+*
