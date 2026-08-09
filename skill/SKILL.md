---
name: c4drill-toml
description: Generate valid C4Drill TOML architecture definitions
version: 1.0.0
---

# C4Drill TOML Skill

**Purpose:** Enable AI assistants to generate valid C4Drill TOML
architecture definitions without prior training.

**Target:** Generic LLM (model-agnostic). Works with Claude Code, Cursor,
OpenCode, or any AI with skill support.

---

## Quick Reference

### Unit Types (16 total)

**C1 Context Level:**

* `person` - Actor/user
* `personExternal` - External actor
* `system` - Software system
* `systemExternal` - External system
* `db` - Database
* `dbExternal` - External database
* `queue` - Message queue
* `queueExternal` - External queue
* `box` - Grouping container (can appear at any level: C1, C2, C3)

**C2 Container Level:**

* `container` - Container within system
* `containerDb` - Container database
* `containerQueue` - Container queue

**C3 Component Level:**

* `component` - Component within container
* `componentDb` - Component database
* `componentQueue` - Component queue

### Required Fields

**Every unit must have:**

```toml
name = "Display Name"  # Optional - humanized from identifier if omitted
type = "<unit_type>"   # Optional - defaults based on nesting level
```

**Default types by nesting level:**

| Level | Parent | Default Type |
|-------|--------|--------------|
| C1 | None (root) | `system` |
| C2 | system, systemExternal, box | `container` |
| C3 | container | `component` |

```toml
[webapp]              # C1: defaults to "system"
name = "Web App"

[webapp.api]          # C2: defaults to "container"
name = "API Service"

[webapp.api.handlers] # C3: defaults to "component"
name = "Handlers"
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
type = "system"                 # Optional: Unit type (defaults based on nesting)
name = "Display Name"           # Optional: defaults to humanized last path segment
description = "What it does"    # Optional: Brief description
technology = "Go, PostgreSQL"   # Optional: Tech stack (not for person types)
color = "#E3F2FD"               # Optional: Background color override
style = "filled"                # Optional: Visual style override
border = "#1565C0"              # Optional: Border color override
edges = "spline"                # Optional: Edge style (cascades to subunits)
width = 300                     # Optional: Explicit width (0=auto)
height = 200                    # Optional: Explicit height (0=auto)
expanded = ["subunit1"]         # Optional: Subunits expanded by default
reference = "https://docs.example.com/runbook" # Optional: External docs URL (📖 marker, clickable)
[[unit.link]]                   # Optional: Outgoing links (array of tables)
[[unit.linkFrom]]               # Optional: Incoming links (array of tables)
```

**Optional name humanization:** when `name` is omitted, the display name
is derived from the last segment of the unit's identifier via a dumb
camelCase split (e.g. `localIDP` → "Local IDP", `sessionManager` →
"Session Manager", `linuxSystem` → "Linux System"). Acronyms are **not**
preserved (`gRPC` → "Grpc") — set `name =` explicitly to override. An
explicit `name =` always wins.

#### reference (External Documentation URL)

Any unit accepts an optional `reference` field — an external
documentation URL. When set, a 📖 marker appears next to the unit name
and the node is clickable.

```toml
[api]
type = "system"
name = "API Service"
reference = "https://wiki.example.com/api-runbook"
```

* Empty string and an omitted field are equivalent (no 📖, not
  clickable).
* The URL is rendered via GraphViz's native `URL` attribute (SVG) and
  routed by the HTML shim; external `http(s)` references open in a new tab
  in `-f html` output (distinct from internal drill-down navigation).

**Type inference rules** (when `type` is omitted, or when a generic
`db`/`queue` type is set):

**Default type by parent** (`defaultTypeForParent`):

| Parent type | Inferred child type | Level |
|---|---|---|
| (none — root) | `system` | C1 |
| `system` | `container` | C2 |
| `box` | `system` | C1 (same-level grouping) |
| `container` | `component` | C3 |
| `containerBox` | `container` | C2 (same-level grouping) |
| `componentBox` | `component` | C3 (same-level grouping) |
| (other: db, queue, etc.) | `system` | C1 fallback |

**Generic `db`/`queue` promotion by nesting level** (`inferGenericType`):

| Parent type | `db` becomes | `queue` becomes | Level |
|---|---|---|---|
| (none) or `box` | `db` | `queue` | C1 (unchanged) |
| `system` or `containerBox` | `containerDb` | `containerQueue` | C2 |
| `container` or `componentBox` | `componentDb` | `componentQueue` | C3 |

```toml
[platform]
# type omitted → inferred "system" (no parent)
[platform.webapp]
# type omitted → inferred "container" (parent is system)
[platform.webapp.cache]
type = "db"
# explicit generic db → promoted to "componentDb" (parent is container)
```

An explicit non-generic `type =` always wins (no inference runs).
Source: `defaultTypeForParent` and `inferGenericType` in
`internal/parser/parser.go`.

### Pipeline Ordering

The v1.10 composition features run in a fixed pipeline order. Ordering
is load-bearing for correctness:

```text
include.Resolve → template.Expand → peer.Resolve → humanize → validate → views → render
```

* **`include.Resolve` runs FIRST** so templates defined in included files
  are visible to `[[use]]` directives in the entry file (XC-02).
* **`template.Expand` runs before `peer.Resolve`** so relative peers
  authored inside templates resolve at the *instantiation site* (the
  `[[use]]` location), not the template's lexical location (XC-03).
* **`humanize` runs after expand** (so it sees substituted names) and
  *before validate* (so error messages show final names) (XC-04).
  Currently humanize runs at parse time; templates carry explicit `name=`
  so parse-time humanize does not fire for them.
* **`validate` sees only absolute paths** (peer.Resolve has rewritten all
  bare peers) and a fully-expanded model (no `${param}` tokens, no
  `[[use]]`/`[[include]]` directives remain).

Reordering any pass breaks the multi-file templated relative-peer case
(enforced as a behavioral regression test — see the XC-01 test in
`cmd/c4drill/root_test.go`).

### Nesting (C2/C3 Diagrams)

Use dotted notation for nested units. Types that can have subunits:

* `system`, `systemExternal` — can contain containers or boxes
* `container` — can contain components or boxes
* `box` — can contain any unit type (grouping container)

**Minimal example** (types inferred from nesting):

```toml
[properties]
name = "My App"

[mainapp]                       # C1: defaults to "system"
name = "Main Application"

[mainapp.api]                   # C2: defaults to "container"
name = "API Service"

[mainapp.api.handlers]          # C3: defaults to "component"
name = "HTTP Handlers"

[mainapp.api.services]          # C3: Another component
name = "Business Services"

[mainapp.db]                    # C2: Container database (type required for non-default)
type = "containerDb"
name = "Database"
```

**Explicit types** (when you need specific variants):

```toml
[mainapp]                       # C1: System
type = "system"
name = "Main Application"

[mainapp.api]                   # C2: Container
type = "container"
name = "API Service"

[mainapp.api.handlers]          # C3: Component (inside container)
type = "component"
name = "HTTP Handlers"

[mainapp.api.services]          # C3: Another component
type = "component"
name = "Business Services"

[mainapp.db]                    # C2: Container database (no subunits)
type = "containerDb"
name = "Database"
```

**`box` can be used at any nesting level** to create logical groupings:

**C2: Box grouping containers (inside system)**

```toml
[mainapp]                       # C1: System
type = "system"
name = "Main App"

[mainapp.services]              # C2: Box grouping containers
type = "box"
name = "Microservices"

[mainapp.services.user]         # C2: Container inside box
type = "container"
name = "User Service"

[mainapp.services.order]        # C2: Another container
type = "container"
name = "Order Service"
```

**C3: Box grouping components (inside container)**

```toml
[mainapp]
type = "system"
name = "Main App"

[mainapp.api]
type = "container"
name = "API Service"

[mainapp.api.domain]            # C3: Box grouping components
type = "box"
name = "Domain Layer"

[mainapp.api.domain.repo]       # C3: Component inside box
type = "component"
name = "Repository"

[mainapp.api.domain.service]    # C3: Another component
type = "component"
name = "Service"
```

### Link Syntax

Define relationships using `[[link]]` (outgoing) or
`[[linkFrom]]` (incoming):

**Outgoing link (defined on source):**

```toml
[user]
type = "person"
name = "User"

[[user.link]]
peer = "webapp"
technology = "HTTPS"
description = "Browses"
```

**Incoming link (defined on target):**

```toml
[api]
type = "system"
name = "API Service"

[[api.linkFrom]]
peer = "webapp"
technology = "REST"
description = "Calls"
```

**Multiple links (including to same peer):**

```toml
[[api.link]]
peer = "webapp"
technology = "HTTPS"
description = "Browses"

[[api.link]]
peer = "webapp"
technology = "WebSockets"
description = "Real-time updates"
```

### Link Attributes

| Attribute | Values | Description |
|---|---|---|
| `peer` | `"unit_name"` | **Required:** Target unit identifier |
| `arrow` | `forward` (default), `reverse`, `bidirectional`, `none` | Arrow direction |
| `rank` | `forward`, `reverse`, `equal` | Layout ranking hint |
| `color` | `"blue"`, `"#FF5733"` | Edge color |
| `style` | `"solid"`, `"dashed"`, `"dotted"` | Line style |
| `technology` | `"HTTPS"`, `"gRPC"`, `"TCP"` | Protocol/technology label |
| `description` | `"Sends events to"` | Relationship description |
| `labelPosition` | `middle` (default), `tail`, `head` | Where label appears |

### Templates ([template.*] + [[use]])

Define a parametrized unit template once and instantiate it N times with
distinct parameter values. A `[template.<name>]` table declares its
parameters and the unit shape (fields, links, subunit subtrees); each
`[[use]]` directive instantiates it under a parent with concrete
parameter values.

**`[template.<name>]` fields:**

| Field | Description |
|---|---|
| `params = ["a", "b", ...]` | Required: declares the named parameters. ALL are required on every `[[use]]` (no defaults). |
| `name`, `description`, `technology`, `reference`, `color` | Standard unit fields; `${param}` substitutes into each. |
| `type` | Standard unit type. Must be valid for the instantiation parent. |
| `[[template.<name>.link]]` | Fixed link set (TMPL-03); peer/description/technology fields accept `${param}`. |
| `[template.<name>.<child>]` | Subunit subtree (TMPL-04); child key verbatim (D-04), field values substituted. |

**`[[use]]` directive fields:**

| Field | Description |
|---|---|
| `template = "<name>"` | Required: the `[template.<name>]` to instantiate. |
| `parent = "<dotted.path>"` | Optional: placement path (empty/omitted = top-level). Produced path = `parent + "." + name-param`. |
| `name = "<value>"` | Required: fills the produced unit's last path segment AND the template's `${name}` token. |
| (other keys) | The template's remaining params; each must match a declared `params` entry. |

**Validation rules:**

* All declared `params` are required on every `[[use]]`
  (TMPL-02/06); a missing param is a hard error naming the template, the
  param, and the instantiation site.
* `${param}` substitutes into
  Name/Description/Technology/Reference/Color and all Link fields
  (TMPL-03, TMPL-10).
* The link set is **fixed** — N instantiations of a template with one link
  produce N links, not a fan-out (no `for_each`).
* Subunit subtrees are deep-copied per instantiation (TMPL-08); the
  subunit key is verbatim, only field values substitute.
* Duplicate unit paths across instantiations (or with hand-authored
  units) are a hard error (TMPL-07).
* Forward references are allowed (a `[[use]]` may precede its
  `[template.*]` definition; an instantiated unit may be
  referenced before the `[[use]]`).

**Interactions:** runs as the 2nd pipeline pass (see Pipeline Ordering).
Relative peers authored inside a template resolve at the instantiation
site (XC-03). Reference-field parametrization composes with the Phase 28
reference feature (TMPL-10). A template with a subunit cannot also carry
direct links (validator rule — split into a leaf template + a
parent template if you need both).

### Include ([[include]])

Assemble a model from multiple TOML files. Each `[[include]]`
pulls in another file relative to the including file and merges its
units.

**`[[include]]` fields:**

| Field | Description |
|---|---|
| `path = "<relative>"` | Required: TOML file to include; relative to the including file's directory (INC-02). |
| `once = true` | Optional: opts into visited-set dedup (INC-06) — a file included again via any path is skipped. |

**Validation rules:**

* Paths resolve relative to the INCLUDING file's directory, not the CLI
  cwd (INC-02).
* Includes are *transitive* (an included file may itself
  `[[include]]` others) (INC-03).
* Include cycles (self or mutual) are a fatal error naming the cycle
  (INC-04).
* A same-file diamond (the same canonical path reached via two
  `[[include]]` paths) is auto-deduped silently (D-11); a
  cross-file unit-path collision (two different files defining the same
  `[unit]`) is a hard error.
* `once=true` adds the file's canonical path to a visited set;
  subsequent `[[include]]` of the same path is skipped (INC-06).
* The merge is *flat* — no namespacing. Included units append to
  `UnitOrder` in include-directive order (D-09).
* Cross-file subunits (D-10): an included file may re-declare a parent
  declared in the entry; its subunits attach onto the entry's existing
  parent, appending to `SubunitOrder` in include-file order.
* Properties follow root-file-wins (the entry's `[properties]`
  takes precedence; included `[properties]` are ignored) (INC-08).
* A missing include file is a hard error naming the including file
  (INC-10/D-12).

**Interactions:** runs as the 1st pipeline pass (see Pipeline Ordering) so
templates defined in included files are visible to `[[use]]` in
the entry (XC-02).

### Relative Peer Resolution

Bare peers (no dot) resolve by walking the host unit's ancestor scopes
nearest-first (ERGO-01); absolute peers (with a dot) are used as-is
(ERGO-02). A `peer` value is either **absolute** (contains a dot) or
**bare** (no dot). The two are gated distinctly (D-16):

* **Absolute peer** (e.g. `"platform.api"`): used as-is. No walk-up.
* **Bare peer** (e.g. `"cache"`): resolved by walking the host unit's
  ancestor scopes nearest-first (D-13/D-14/D-15).

**Walk-up algorithm:** starting at the host unit's immediate parent's
children-map, check for a child matching the peer name. If found,
resolve to that child's absolute path. If not, ascend to the
grandparent's children-map and repeat. Continue until a match is found
or the root scope is exhausted.

| Case | Host | Bare peer | Resolves to |
|---|---|---|---|
| Sibling (D-13) | `platform.api` | `cache` | `platform.cache` (sibling under nearest ancestor) |
| Aunt (D-14) | `platform.api.handlers.auth` | `queue` | `platform.queue` (walk up past empty scopes) |
| Root (D-15) | `platform.api.handlers.auth` | `messageBus` | `messageBus` (top-level, walk reaches root) |
| Absolute (D-16) | any | `platform.cache` | `platform.cache` (dot = no walk-up) |

**Error cases:** a bare peer that matches nothing at any scope up to and
including root is a hard error naming the peer and the host unit.
Multiple matches at the same depth are structurally impossible (sibling
keys are unique per parent).

**Interactions:** runs as the 3rd pipeline pass, after `template.Expand`
(see Pipeline Ordering). Templated units' relative peers resolve at the
instantiation site (XC-03). Absolute peers (with a dot) are always
unchanged (ERGO-02 backward-compat — any pre-v1.10 model using dotted
peers renders identically).

---

## Validation Rules

**Rule 1: Referenced units must exist**

```toml
# ❌ INVALID: "api" unit not defined
[user]
type = "person"
name = "User"

[[user.link]]
peer = "api"
description = "Calls"

# ✅ VALID: Both units defined
[user]
type = "person"
name = "User"

[[user.link]]
peer = "api"
description = "Calls"

[api]
type = "system"
name = "API"
```

**Rule 2: Units with subunits cannot have links**

```toml
# ❌ INVALID: mainapp has subunits but also has links
[mainapp]
type = "system"

[[mainapp.link]]
peer = "external"
description = "Calls"

[mainapp.api]
type = "container"

# ✅ VALID: Links on subunits only
[mainapp]
type = "system"

[mainapp.api]
type = "container"

[[mainapp.api.link]]
peer = "external"
description = "Calls"
```

**Rule 3: Cannot link to units with subunits**

```toml
# ❌ INVALID: Linking to mainapp which has subunits
[user]
type = "person"
name = "User"

[[user.link]]
peer = "mainapp"
description = "Uses"

[mainapp]
type = "system"

[mainapp.api]
type = "container"

# ✅ VALID: Link to specific subunit
[user]
type = "person"
name = "User"

[[user.link]]
peer = "mainapp.api"
description = "Uses"
```

**Rule 4: Subunits only for system, systemExternal, box, and container
types**

```toml
# ❌ INVALID: Database cannot have subunits
[postgres]
type = "db"

[postgres.replica]
type = "db"

# ✅ VALID: Use box for grouping databases
[postgres]
type = "box"
name = "Database Cluster"

[postgres.primary]
type = "db"
name = "Primary"

[postgres.replica]
type = "db"
name = "Replica"

# ✅ VALID: Container can have component subunits
[api]
type = "container"
name = "API"

[api.handlers]
type = "component"
name = "Handlers"
```

---

## Prompt Patterns

Use these patterns when generating C4Drill TOML from natural language:

### Pattern 1: Basic Architecture

```text
"Create a C4 diagram with: [list units and relationships]"

Example: "Create a C4 diagram with: user, web application, API, database. User calls webapp, webapp calls API, API queries database."
```

### Pattern 2: Architecture Description

```text
"Generate TOML for: [architecture description]"

Example: "Generate TOML for: E-commerce platform with customers browsing products, adding to cart, and checking out. Uses PostgreSQL for storage and Stripe for payments."
```

### Pattern 3: Level Specification

```text
"Create a C[C1/C2/C3] diagram for: [description]"

Example: "Create a C2 diagram for: Microservices architecture with API Gateway, User Service, Order Service, and shared database."
```

### Pattern 4: Existing Architecture

```text
"Model this architecture in C4Drill TOML: [description]"

Example: "Model this architecture in C4Drill TOML: 3-tier web app with React frontend, Node.js API, and PostgreSQL database. Frontend calls API via HTTPS, API queries database."
```

### Pattern 5: Feature Addition

```text
"Add [component] to the architecture: [existing TOML]"

Example: "Add caching layer to the architecture: [paste existing TOML]"
```

---

## Examples

The following examples demonstrate increasing complexity:

1. **[01-minimal.toml](examples/01-minimal.toml)** - Minimal working
   example (person + system + single link)
2. **[02-nested.toml](examples/02-nested.toml)** - Nested structure
   (C2/C3 levels)
3. **[03-links.toml](examples/03-links.toml)** - Link syntax (all
   attributes)
4. **[04-styling.toml](examples/04-styling.toml)** - Visual
   customization (colors, borders, edges)
5. **[05-ecommerce.toml](examples/05-ecommerce.toml)** - Realistic full
   architecture
6. **[06-templates.toml](examples/06-templates.toml)** - Templates
   (`[template.*]` define + `[[use]]` instantiate, TMPL-01..10)
7. **[07-relative-peer.toml](examples/07-relative-peer.toml)** -
   Relative peer walk-up (4 cases: sibling, aunt, root, absolute)
8. **[08-include/](examples/08-include/)** - Multi-file composition
   (`[[include]]`, `once`, cross-file subunits)
9. **[09-composed/](examples/09-composed/)** - All four features
   composed (XC-05 golden pair)

All examples parse and validate successfully with `c4drill validate`.

---

## Generation Guidelines

When generating TOML:

1. **Always include `[properties]` with `name`** - Required field
2. **Always include `type` and `name` for each unit** - Required fields
3. **Use descriptive section names** - They become identifiers (e.g.,
   `user_service` not `svc1`)
4. **Use dotted notation for nesting** - `parent.child` pattern
5. **Prefer `[[link]]` over `[[linkFrom]]`** - More natural
   (source → target)
6. **Include `peer` field in every link** - Required field specifies
   target unit
7. **Include `technology` for systems/databases** - Shows tech stack in
   diagram
8. **Use `external` suffix for external units** - `systemExternal`,
   `dbExternal`
9. **Validate before committing** - Run `c4drill <file.toml>` to
   check

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
type = "person"
name = "User"

[[user.link]]
peer = "mainapp"

[mainapp]
type = "system"

[mainapp.api]
type = "container"
```

**✅ Link to specific subunit:**

```toml
[user]
type = "person"
name = "User"

[[user.link]]
peer = "mainapp.api"

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

Success = no output (generates SVG by default) Failure = error message
with line number

### Output Formats

| Flag | Format | Use when |
|---|---|---|
| (default) | `svg` | General use; clickable navigation in Chrome/Firefox |
| `-f html` | `html` | Safari/WebKit (which ignores SVG `<a>` links); also `file://` viewing |
| `-f dot` | `dot` | Customizing layout in GraphViz tools |

```bash
c4drill architecture.toml -f html    # Safari-compatible
```

---

*Skill version: 1.2.0* *C4Drill version: v1.10+*
