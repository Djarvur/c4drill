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
name = "Customer"                   # Optional: defaults to humanized identifier
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

#### Reference (External Documentation URL)

Any unit accepts an optional `reference` field — an external documentation URL.
When set, a 📖 marker appears next to the unit name and the node becomes
clickable: clicking it opens the URL (via GraphViz's native `URL` attribute in
SVG). In `-f html` output, external `http(s)` references open in a new tab,
distinct from internal drill-down navigation.

```toml
[api]
type = "system"
name = "API Service"
reference = "https://wiki.example.com/api-runbook"   # Optional: 📖 marker, clickable
```

An empty string and an omitted field are equivalent (no 📖, not clickable).

### Optional Name (Humanization)

The `name` field is **optional**. When omitted, the display name is derived from the **last segment** of the unit's identifier via a dumb camelCase split:

```toml
# Explicit name (always wins — use this for acronyms or custom labels)
[linuxSystem.localIDP]
name = "My Custom Name"

# Name omitted — humanized from the last path segment "sessionManager"
[linuxSystem.sessionManager]
# displays as "Session Manager"
```

**Humanization rules:**

- Splits camelCase boundaries and Title-cases each word.
- Operates on the **last path segment only** — `[linuxSystem.localIDP]` becomes "Local IDP", not "Linux System Local IDP".
- Examples: `sessionManager` → "Session Manager", `localIDP` → "Local IDP", `linuxSystem` → "Linux System".

**Acronyms:** acronym preservation is intentionally **not** supported (the split is deliberately dumb). `gRPC` humanizes to "Grpc". To preserve an acronym or set any custom label, set `name =` explicitly — an explicit `name =` always wins.

**Backward compatibility:** existing models that already set `name =` on every unit are completely unaffected — humanization only fires when `name` is omitted.

### Optional Type (Inference)

The `type` field is **optional**. When omitted, the type is inferred from the parent unit's type (and, for the generic `db`/`queue` types, promoted to the level-specific variant based on nesting). Two rules apply:

**1. Default type by parent** — the type assigned when `type` is omitted entirely:

| Parent type | Inferred child type | Level |
|-------------|---------------------|-------|
| (none — root) | `system` | C1 |
| `system` | `container` | C2 |
| `box` | `system` | C1 (same-level grouping) |
| `container` | `component` | C3 |
| `containerBox` | `container` | C2 (same-level grouping) |
| `componentBox` | `component` | C3 (same-level grouping) |
| (other: db, queue, etc.) | `system` | C1 fallback |

**2. Generic `db`/`queue` promotion** — when `type = "db"` or `type = "queue"` is set explicitly, the type is promoted to the level-specific variant based on the parent:

| Parent type | `db` becomes | `queue` becomes | Level |
|-------------|--------------|-----------------|-------|
| (none) or `box` | `db` | `queue` | C1 (unchanged) |
| `system` or `containerBox` | `containerDb` | `containerQueue` | C2 |
| `container` or `componentBox` | `componentDb` | `componentQueue` | C3 |

**Before/after example:**

```toml
# BEFORE — explicit types (verbose)
[platform]
type = "system"
[platform.webapp]
type = "container"
[platform.webapp.cache]
type = "componentDb"     # generic db promoted because parent is container

# AFTER — type omitted, inferred (identical result)
[platform]
# type omitted → inferred "system" (no parent)
[platform.webapp]
# type omitted → inferred "container" (parent is system)
[platform.webapp.cache]
type = "db"
# explicit generic db → promoted to "componentDb" (parent is container)
```

An explicit non-generic `type =` always wins (no inference runs). Source: `defaultTypeForParent` and `inferGenericType` in `internal/parser/parser.go`.

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

### Templates

Define a parametrized unit template once and instantiate it N times with distinct parameter values. A template is a `[template.<name>]` table declaring its parameters and the unit shape (including subunits and links); each `[[use]]` directive instantiates it with concrete values.

```toml
[template.microservice]
params = ["name", "tech", "upstreamBus"]
name = "${name} Service"
type = "container"
technology = "${tech}"
description = "${name} handles its domain"
reference = "https://wiki.example.com/${name}"

[[template.microservice.link]]
peer = "${upstreamBus}"
description = "Publishes ${name} domain events"

[[use]]
template = "microservice"
parent = "platform"
name = "auth"
tech = "Go, gRPC"
upstreamBus = "messageBus"
```

**Rules:**
- All declared params are **required** on every `[[use]]` (no defaults); a missing param is a hard error.
- `${param}` substitutes into every string field — name, description, technology, reference, color, and link fields (peer, description, technology).
- The link set is **fixed**: a template with one `[[template.X.link]]` produces exactly one link per instantiation (no fan-out / `for_each`).
- Subunit subtrees are supported (declare `[template.X.child]`); the subunit key is verbatim, only field values are substituted.
- Duplicate unit paths across instantiations are a hard error.

See `skill/examples/06-templates.toml` for a runnable example.

### Multi-File Composition (Include)

Assemble a diagram from multiple TOML files. Each `[[include]]` directive pulls in another file relative to the including file's directory and merges its units into the model.

```toml
# entry.toml
[platform]
type = "system"
name = "Platform"

[[include]]
path = "auth.toml"

[[include]]
path = "templates.toml"
once = true
```

**Rules:**
- Paths are **relative to the including file's directory** (not the CLI cwd).
- Includes are **transitive** (an included file may itself include others).
- `once = true` deduplicates by canonical path — a file included again (even via a different path) is skipped.
- The merge is **flat** (no namespacing): included units append in include order. Cross-file subunits are supported — an included file may re-declare a parent declared in the entry and contribute subunits under it.
- Include cycles are a fatal error; missing files are a hard error.
- Properties follow root-file-wins (the entry's `[properties]` takes precedence).

See `skill/examples/08-include/` for a runnable multi-file example.

### Relative Peer Resolution

A bare `peer` value (no dot) resolves against the enclosing parent's ancestor scopes — walking up nearest-first until a sibling match is found. A peer **with a dot** is absolute and used as-is.

```toml
[platform.api]
# Bare peer "cache" resolves via walk-up:
[[platform.api.link]]
peer = "cache"                          # sibling: platform.cache (nearest ancestor scope match)

[[platform.api.link]]
peer = "platform.cache"                 # absolute: has a dot, used as-is
```

**The four resolution cases:**
- **Sibling match** — the nearest ancestor scope (the immediate parent) has a child with that name.
- **Aunt/grandparent match** — walk up past the parent; a grandparent's child matches.
- **Root match** — walk all the way to the top-level scope.
- **Absolute fallback** — a peer containing a dot is never walked-up; it is used verbatim.

Multiple matches at the same depth are impossible (sibling keys are unique per parent). A miss at root is a hard error naming the peer and the host unit.

See `skill/examples/07-relative-peer.toml` for a runnable example demonstrating all four cases.

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
  -f, --format string   Output format (dot|svg|html) (default "svg")
  -o, --output string   Output directory (default ".")
      --version         Print version information
  -h, --help            Show help
```

### Output Format

- **svg** (default): Rendered SVG diagrams with clickable navigation links
- **html**: Self-contained HTML files (SVG inlined) with working navigation in
  Safari/WebKit, which silently ignores SVG `<a>` hyperlinks. Use `-f html`
  when diagrams will be opened in Safari or viewed via `file://`.
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
- **Renderer**: Outputs DOT, SVG, or HTML via go-graphviz
- **Writer**: Creates output directory hierarchy

## License

MIT
