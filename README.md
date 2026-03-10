# C4Drill

Transform simple TOML architecture descriptions into professional, interactive C4 diagrams without manual drawing.

**One TOML file → Clickable C1, C2, and C3 diagrams with automatic drill-down navigation.**

## Install

```bash
go install github.com/Djarvur/c4drill/cmd/c4drill@latest
```

Or build from source:

```bash
git clone https://github.com/Djarvur/c4drill
cd c4drill
go build -o c4drill ./cmd/c4drill
```

## Quick Start

```bash
# Generate SVG diagrams (default)
c4drill architecture.toml

# Generate to specific directory
c4drill architecture.toml -o ./docs/diagrams

# Generate DOT format for customization
c4drill architecture.toml -f dot -o ./output
```

## TOML Format

### Minimal Example

```toml
[properties]
name = "My System"

[user]
type = "person"
name = "User"

[webapp]
type = "system"
name = "Web Application"
```

### Properties Section

The `[properties]` section defines global settings for your architecture:

```toml
[properties]
name = "E-Commerce Platform"        # Required: Architecture name
description = "Online store system" # Optional: Description
edges = "spline"                    # Optional: Edge routing (straight|spline|square)
expanded = ["payments"]             # Optional: Units to expand by default
```

### Unit Types

Each unit is defined as a TOML section. The section name becomes the unit's identifier.

**Type is optional** — defaults based on nesting level:
- Root-level units → `system`
- Units inside system/box → `container`
- Units inside container → `component`

#### Person (Actor)

```toml
[user]
type = "person"                     # Required
name = "Customer"                   # Required: Display name
description = "Online shopper"      # Optional
```

#### System

```toml
[webapp]
type = "system"
name = "Web Application"
description = "Frontend web app"
technology = "React, TypeScript"    # Optional: Shown in diagram
```

#### External System

```toml
[stripe]
type = "systemExternal"
name = "Stripe"
description = "Payment processor"
```

#### Database

```toml
[postgres]
type = "db"
name = "PostgreSQL"
description = "Primary database"
technology = "PostgreSQL 15"
```

#### External Database

```toml
[analytics]
type = "dbExternal"
name = "Analytics DB"
description = "Third-party analytics"
```

#### Queue

```toml
[rabbitmq]
type = "queue"
name = "Message Queue"
description = "Async job processing"
technology = "RabbitMQ"
```

#### External Queue

```toml
[sqs]
type = "queueExternal"
name = "AWS SQS"
description = "External message queue"
```

#### Box (Grouping Container)

```toml
[cloud]
type = "box"
name = "AWS Cloud"
description = "Cloud infrastructure"
```

### Nesting (C2/C3 Diagrams)

Systems and boxes can contain subunits using dotted notation:

```toml
[mainapp]                    # C1 level
type = "system"
name = "Main Application"

[mainapp.api]                # C2 level (container)
type = "container"
name = "API Service"

[mainapp.webapp]             # C2 level (container)
type = "container"
name = "Web App"

[mainapp.api.handlers]       # C3 level (component)
type = "component"
name = "HTTP Handlers"

[mainapp.api.services]       # C3 level (component)
type = "component"
name = "Business Services"
```

### Links (Relationships)

Define relationships between units using `link` or `linkFrom`:

```toml
[user]
type = "person"
name = "User"

[webapp]
type = "system"
name = "Web Application"
# Outgoing link: User → Webapp
link = { "user" = { technology = "HTTPS", description = "Browses" } }

[api]
type = "system"
name = "API Service"
# Incoming link: Webapp → API (defined on target)
linkFrom = { "webapp" = { technology = "REST/JSON", description = "Calls" } }
```

#### Link Attributes

| Attribute | Description | Example |
|-----------|-------------|---------|
| `technology` | Protocol/technology label | `"HTTPS"`, `"gRPC"`, `"TCP"` |
| `description` | Relationship description | `"Sends events to"` |
| `arrow` | Arrow direction | `"forward"`, `"reverse"`, `"both"`, `"none"` |
| `rank` | Layout ranking hint | `"forward"`, `"reverse"` |
| `color` | Edge color | `"blue"`, `"#FF5733"` |
| `style` | Line style | `"solid"`, `"dashed"`, `"dotted"` |

#### Multiple Links

```toml
[api]
type = "system"
name = "API"
link = {
  "user" = { technology = "HTTPS", description = "Authenticates" },
  "webapp" = { technology = "REST", description = "Serves data" }
}
```

### Expanded Units (Drill-Down)

Mark units as "expanded" to generate C2/C3 diagrams for them:

```toml
[properties]
name = "My System"
expanded = ["mainapp"]        # Generate C2 for mainapp

[mainapp]
type = "system"
name = "Main Application"
expanded = ["mainapp.api"]    # Generate C3 for mainapp.api

[mainapp.api]
type = "system"
name = "API Service"

[mainapp.api.handlers]
type = "component"
name = "Handlers"

[mainapp.api.services]
type = "component"
name = "Services"
```

**Output structure:**

```text
output/
├── architecture.svg           # C1 diagram
├── architecture/              # C2 diagrams directory
│   └── mainapp.svg            # C2 for mainapp
│   └── mainapp/               # C3 diagrams directory
│       └── api.svg            # C3 for mainapp.api
```

### Styling

Override default colors and styles per unit:

```toml
[webapp]
type = "system"
name = "Web Application"
color = "#4A90D9"              # Background color
border = "#2E5A8B"             # Border color
style = "solid"                # Border style (solid|dashed|dotted)
```

## Full Example

```toml
[properties]
name = "E-Commerce Platform"
description = "Online shopping system"
edges = "spline"
expanded = ["webapp", "webapp.api"]

# External actors
[customer]
type = "person"
name = "Customer"
description = "Online shopper"

[admin]
type = "person"
name = "Admin"
description = "System administrator"

# Main system with containers
[webapp]
type = "system"
name = "Web Application"
description = "Main e-commerce platform"
technology = "Go, React"
expanded = ["webapp.api"]
link = { "customer" = { technology = "HTTPS", description = "Shops" } }

[webapp.frontend]
type = "system"
name = "Frontend"
description = "React web application"
technology = "React, TypeScript"

[webapp.api]
type = "system"
name = "API Service"
description = "Backend REST API"
technology = "Go"
expanded = ["webapp.api"]

[webapp.api.handlers]
type = "component"
name = "HTTP Handlers"
description = "Request routing"

[webapp.api.services]
type = "component"
name = "Business Logic"
description = "Core services"

[webapp.db]
type = "db"
name = "PostgreSQL"
description = "Primary database"
technology = "PostgreSQL 15"

# External dependencies
[stripe]
type = "systemExternal"
name = "Stripe"
description = "Payment processing"

[redis]
type = "queue"
name = "Redis"
description = "Session cache"
technology = "Redis"

# Relationships
[webapp.api.linkFrom]
"webapp.frontend" = { technology = "REST/JSON", description = "Calls" }

[stripe.linkFrom]
"webapp.api" = { technology = "Stripe API", description = "Processes payments" }

[redis.linkFrom]
"webapp.api" = { technology = "TCP", description = "Caches sessions" }
```

## CLI Reference

```text
c4drill <input.toml> [flags]

Flags:
  -f, --format string   Output format (dot|svg) (default "svg")
  -o, --output string   Output directory (default ".")
      --version         Print version information
  -h, --help            Show help
```

### Output Format

- **svg** (default): Rendered SVG diagrams with clickable navigation links
- **dot**: Raw GraphViz DOT format for customization

### Exit Codes

- **0**: Success (silent output)
- **1**: Error (parse failure, validation error, I/O error)

Errors are written to stderr, making the tool suitable for scripting.

## Validation Rules

1. **Referenced units must exist** - Links can only reference defined units
2. **No links on containers** - Units with subunits cannot have their own links
3. **No linking to containers** - Cannot link to units that have subunits
4. **Subunits only for system/box** - Only systems and boxes can contain subunits

## Architecture

C4Drill implements a compiler-style pipeline:

```text
TOML → Parse → Validate → Generate Views → Build Graphs → Render → Write Files
```

- **Parser**: Reads TOML into structured model
- **Validator**: Enforces C4 rules and reference integrity
- **View Generator**: Creates C1/C2/C3 views from model
- **Graph Builder**: Constructs graphviz-compatible structures
- **Renderer**: Outputs DOT or SVG via go-graphviz
- **Writer**: Creates output directory hierarchy

## License

MIT
