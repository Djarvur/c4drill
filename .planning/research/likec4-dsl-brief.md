# LikeC4 DSL Technical Brief

**Purpose:** Factual reference for scoping the v1.11 LikeC4 → C4Drill converter. All facts sourced from likec4.dev docs and the likec4/likec4 repo. No design recommendations.

**Compiled:** 2026-08-08 (background research for v1.11 milestone).

---

## 1. Grammar Surface

### Source files
- Extensions: `.likec4` or `.c4`.
- All source files in a workspace are **merged into one model** (file/folder boundaries are organizational only).

### Top-level statements (a file must have ≥1)
| Statement | Purpose |
|---|---|
| `specification { ... }` | Defines element kinds, relationship kinds, tags, colors — the notation vocabulary. |
| `model { ... }` | Logical architecture: elements, hierarchy, relationships. |
| `views { ... }` | Visualizations. Can appear multiple times; each block carries its own "local styles". |
| `global { ... }` | Shared global predicate groups and global named styles. |
| `deployment { ... }` | Physical/deployment model (separate layer, see §8). |

Multiple statements of the same type are allowed and concatenated. Order is not significant at the top level.

### Minimal model
```likec4
specification {
  element actor
  element system
  element component
}
model {
  customer = actor 'Customer'
  cloud = system 'Our SaaS' {
    ui = component 'Frontend'
    api = component 'Backend'
    ui -> api 'requests'
  }
  customer -> ui
}
```

### Realistic medium model
See §10 — the bigbank.c4 example (228 lines) is a representative real file: ~6 top-level elements, 3 levels of nesting, ~25 relationships, 8 views.

---

## 2. Element / Kind System

### Built-in kinds: **NONE**
There are no predefined element kinds. Every kind used in the model must be declared in `specification`. Names are free-form identifiers (`actor`, `system`, `container`, `component`, `microservice`, `pgTable`, `graphqlMutation` — all user-chosen). The C4-Canonical names (person/system/container/component) are conventions, not built-ins.

> Note: the bigbank example uses kinds like `enterprise`, `softwaresystem`, `existingsystem`, `staff`, `spa`, `mobileApp`, `container`, `component`, `database`, `person` — every one declared in its spec block.

### `spec.elements` / custom kinds
```likec4
specification {
  element queue {
    title 'Kafka'            // optional default title
    description 'Kafka queue'
    technology 'kafka topic'
    notation 'Kafka Topic'   // short label shown in legend
    style { shape queue }    // default styling for all elements of this kind
    #infra #data-lake        // tags applied to every element of this kind
  }
}
```
A kind declaration is `element <name> { ... }`. The body is optional — `element system` alone is valid. Body may contain: `title`, `description`, `technology`, `notation`, `style { }`, and tags (`#tag`).

### Kind hierarchy: **NONE**
There is **no `kind.of` / `element X extends Y` subtyping**. Element kinds are flat scalar strings. Confirmed by:
- No such syntax in the specification grammar docs.
- The model API exposes `element.kind` as a single string; view predicates test equality (`element.kind = system`, `element.kind != container`) — no ancestor/descendant operators.
- `extends` exists only for views; `extend` (different keyword) is for adding to existing element *instances*, not for kind inheritance.

If you need "subkinds", the convention is naming (`microservice`, `service`) or tags. A converter cannot rely on any hierarchy.

---

## 3. Relationships

### Two syntactic forms

**Infix `->`** (most common):
```likec4
customer -> cloud 'uses to manage data'
frontend -> backend 'requests data' #graphql #team1
```

**Kinded relationships** — two notations:
```likec4
system1 -[async]-> system2     // bracket form
system1 .async system2         // dot-prefix form
system1 -[async]<-> system3    // bidirectional variant exists
```

### Sourceless / `this` / `it` in nested context
Inside an element body, `-> target` implicitly uses the parent as source. `this` and `it` both refer to the parent element:
```likec4
actor customer {
  it -> frontend          // customer is source
  frontend -> this        // customer is target
  -> frontend             // shorthand: customer -> frontend
}
```

### Relationship properties
| Property | Syntax | Notes |
|---|---|---|
| `title` | inline string or nested `title '...'` | Shown as edge label. |
| `description` | inline or nested | Markdown supported with triple-quotes `'''`. |
| `technology` | inline or nested | e.g. `'HTTPS'`, `'Kafka'`. |
| `summary` | (not documented for relationships; only elements) | — |
| `notation` | inherited from kind | Short legend label. |
| tags | `#graphql #team1` inline, or nested `#tag` block | Must precede other nested props. Merged with kind tags. |
| `link` | `link <url> ['label']` | Multiple allowed. |
| `metadata { }` | nested block | Same key-value rules as elements (§7). |
| `navigateTo` | `navigateTo <viewId>` | Links to a dynamic view for drill-down. |
| `style { }` | nested block | color, line, head, tail, multiple (§6). |

Inline positional order: `[title] [description] [technology]`.
```likec4
customer -> frontend 'opens in browser' 'Customer opens...' 'HTTPS'
```

Relationships are uniquely identified by (source, target, kind if any, title if any) — this matters for `extend relationship`.

---

## 4. Model Tree / Nesting

### Lexical nesting (primary mechanism)
Any element can contain children in `{ }`:
```likec4
service service1 {
  component backend {
    component api
  }
  component frontend
}
```
FQNs become: `service1`, `service1.backend`, `service1.backend.api`, `service1.frontend`.

### Two declaration syntaxes
```likec4
component backend { ... }     // kind-first (name required in body? NO — name follows kind)
backend = component { ... }   // name-first with '='
```
Both are equivalent. The bigbank file uses kind-first throughout.

### Explicit cross-file extension via `extend`
To add children/properties to an element defined elsewhere (different file):
```likec4
model {
  extend cloud {
    service1 = service 'Service 1'
    service1 -> service2     // siblings resolve in extended scope
  }
  extend cloud {
    #additional-tag
    metadata { prop1 'value1' }
    link ../src/index.ts
  }
}
```
`extend` merges: tags appended, links appended, metadata merged (dup keys → arrays, dedup).

### There is **no `parent` / `of` property** on elements
Parent-child is established **purely by lexical nesting** or by `extend <element> { children }`. There is no `parent: foo` attribute. The `view X of Y` uses `of`, but that's view-scoping (§5), not parent assignment.

### Dot-path references
Yes — nested elements are referenced by FQN dot-paths:
```likec4
frontend -> service1.backend.api
frontend -> backend1.api        // partial FQN OK if unambiguous
```
Resolution is lexical-scope + hoisting (like JS): short names resolve if unique in scope; otherwise full/partial FQN required. "Bubbling": a unique nested name is reachable from outer scopes.

### Name rules
Letters, digits, hyphens, underscores. Cannot start with a digit, cannot contain `.`. `api`, `Api2`, `_api`, `__Api-1` valid; `1api`, `a.pi` invalid.

---

## 5. Views DSL

### Structure
```likec4
views {
  view <name> [of <element>] [extends <otherView>] {
    #tags                      // optional, must precede other props
    title '...'
    description '...'          // markdown ok
    link <url> ['label']

    include <predicate>
    exclude <predicate>
    style <targets> { ... }
    group 'Label' { include ... }
    autoLayout LeftRight 120 110
    rank same { a, b }

    // dynamic views only:
    // customer -> web 'opens'
  }
}
```
- Views may be **named** (unique, used for export filenames/URLs) or **unnamed**.
- Properties must come **before** predicates.
- `dynamic view <name> { ... }` is a different beast (imperative sequence of interactions defined inline, not predicate-based).

### Scoped views (`view X of Y`)
`of <element>` sets a scope: bare references inside resolve relative to that element (`api` → `cloud.backend.api`). A scoped view also becomes the **default drill-down view** for that element.

### Predicates — what exists (C4Drill drops all of this)
**Element predicates:**
- `include <fqn>` — explicit element
- `include cloud.*` — direct children
- `include cloud.**` — all descendants (that have relationships to visible)
- `include cloud._` — expand children (only those with visible relationships)
- `include *` — wildcard (top-level for unscoped; element+children for scoped)
- `include element.kind = system` / `element.kind != container`
- `include element.tag = #V2` / `element.tag != #next`
- `include x with { title ... color ... shape ... }` — override props per-view
- `exclude <same forms>`

**Relationship predicates:**
- `a -> b` — directed
- `a <-> b` — any direction
- `-> backend` — incoming
- `customer ->` — outgoing
- `-> cloud.* ->` — in/out of children

**Filtering:**
- `where kind is microservice`, `where tag != #deprecated`, `where metadata.environment is "production"`
- Logical: `and`, `or`, `not`, `is`/`==`, `is not`/`!=`
- `where source.tag is #next`, `where target.kind is microservice`, `where source.metadata.env is "prod"`

**Other:**
- `group 'Label' { include ... }` — boundary box, nestable, styleable
- `view v2 extends v1 { ... }` — inherits predicates+styles+scope
- `rank same/min/max/source/sink { a, b }` — layout constraint

### Mapping implication for C4Drill
C4Drill auto-generates views and has no view DSL. A TOML-emitting converter would **drop the entire `views {}` block** unless C4Drill's output needs to produce LikeC4-readable `.c4` files. If only emitting a model, `views` is simply omitted (LikeC4 accepts a model with no views). If a view is required for LikeC4 to render anything, a trivial `view index { include * }` suffices.

---

## 6. Styling

### Three attachment points
1. **Per kind** in `specification`:
   ```likec4
   element actor { style { shape person; color red } }
   ```
2. **Per element** in `model` (overrides kind):
   ```likec4
   customer = actor 'Customer' {
     style { color green }   // inherits shape: person from kind
   }
   ```
3. **Per view** (predicates, §5):
   ```likec4
   style * { color muted; opacity 10% }
   style element.tag = #deprecated { color muted }
   ```

### Element style properties
| Property | Values | Color/Shape/Icon? |
|---|---|---|
| `shape` | `rectangle` (default), `component`, `storage`, `cylinder`, `browser`, `mobile`, `person`, `queue`, `bucket`, `document` | **Shape** (geometry) |
| `color` | `primary` (default), `secondary`, `muted`, `amber`, `gray`, `green`, `indigo`, `red`, `blue`, `slate`, `sky` + custom colors from spec | **Color** |
| `size` | `xsmall`/`xs`, `small`/`sm`, `medium`/`md` (default), `large`/`lg`, `xlarge`/`xl` | **Shape** |
| `textSize` | same enum | **Text** |
| `padding` | same enum | **Shape** |
| `opacity` | `10%` ... `100%` (group/compound nodes) | **Color** |
| `border` | `dashed` (default), `dotted`, `solid`, `none` (compound nodes) | **Shape** |
| `multiple` | `true`/`false` — render as stacked instances | **Shape** |
| `icon` | URL, `@/path`, or bundled `aws:x`, `azure:x`, `gcp:x`, `tech:x`, `bootstrap:x` | **Icon** |
| `iconColor` | theme/custom color (only affects `bootstrap:*`) | **Icon** |
| `iconSize` | same enum as size | **Icon** |
| `iconPosition` | `left` (default), `right`, `top`, `bottom` | **Icon** |

### Relationship style properties
| Property | Values |
|---|---|
| `color` | same palette |
| `line` | `dashed` (default), `solid`, `dotted` |
| `head` / `tail` | `normal`, `onormal`, `diamond`, `odiamond`, `crow`, `vee`, `open`, `none` |
| `multiple` | `true` — render parallel edges separately instead of merging |

### `spec.preset` themes
**Not found in the docs.** The docs reference theme colors (the `primary`/`secondary`/`muted`/`amber`/etc. enum) and project-level style customization via configuration (`/dsl/config/#styles-customization`), but there is **no `preset` keyword in the specification DSL**. Theme switching is a project-config concern, not a DSL statement.

### Shapes-vs-icons split (relevant to `dev/shapes-no-icons`)
- **Shape-only properties** (safe to emit with no icon support): `shape`, `color`, `size`, `textSize`, `padding`, `opacity`, `border`, `multiple`.
- **Icon-only properties**: `icon`, `iconColor`, `iconSize`, `iconPosition`.
- A `shapes-no-icons` converter should simply **omit all `icon*` properties**; shapes and colors work independently.

---

## 7. Tags / Metadata / Links / Title / Description / Technology

All attachable to elements, relationships, and views (with minor variation).

### `title` / `description` / `summary` / `technology`
- **Elements**: all four valid. Inline positional form: `<kind> <name> [title] [summary] [technology] { ... }`. `summary` falls back to `description` if absent; if both set, `summary` shows on diagram, `description` in details dialog.
- **Relationships**: `title`, `description`, `technology` valid (no `summary`). Inline positional: `[title] [description] [technology]`.
- **Views**: `title`, `description` valid (no `technology`/`summary`).
- Markdown supported in `description`/`summary` via triple-quotes `'''` or `"""`.

### `tag` / `#tag`
- Declared in `specification { tag deprecated { color #FF0000 } }`. Color is optional.
- Applied as `#tag` prefix (must come **first** in any block, before other properties). Multiple: `#a, #b` or `#a #b`.
- Attachable to: elements, relationships, views, element kinds (applies to all instances), relationship kinds.
- Tags are merged (not overridden) when inherited from kind.

### `link`
```likec4
link https://example.com
link https://github.com/likec4/likec4 'Repository'
link ssh://bastion.internal 'SSH'
link ../src/index.ts#L1-L10      // relative paths allowed
```
Multiple links per entity. Attachable to elements, relationships, views.

### `metadata { }`
```likec4
metadata {
  prop1 'value1'
  prop2 'value2'
  array_key ['a', 'b', 'c']
  bool_key true
  json_key '{ "type": "object" }'
}
```
- Key-value pairs. Values: strings, string-embedded JSON/YAML, array literals `['x', 'y']`, booleans `true`/`false`.
- **Duplicate keys** collect into arrays preserving order: `version '1.0'` then `version '2.0'` → `version: ['1.0', '2.0']`.
- Display sorted alphabetically (display-only; order in source preserved for merging).
- Attachable to elements and relationships.
- On `extend`, metadata merges: new keys added, same-key different-value → array, dedup.

---

## 8. Deployments

Separate physical layer. **Out of scope for v1 — name it and skip.**

### Top-level: `deployment { }` (distinct from `model`)

### Spec
```likec4
specification {
  deploymentNode environment
  deploymentNode zone
  deploymentNode kubernetes { style { color blue; icon tech:kubernetes } }
  deploymentNode vm { notation 'Virtual Machine'; technology 'VMware' }
}
```
`deploymentNode` is the parallel of `element` for the deployment layer.

### Structure
Hierarchical nodes (same nesting/`=`/`extend` rules as logical model):
```likec4
deployment {
  environment prod {
    zone eu {
      zone zone1 { vm vm1; vm vm2 }
    }
  }
}
```

### Linking logical → deployment
`instanceOf <logicalElement>` places a logical element onto a node:
```likec4
zone zone1 {
  instanceOf frontend.ui
  ui = instanceOf frontend.ui   // '=' form; renamed instance
  api1 = instanceOf backend.api // multiple instances of same element OK
}
```
Instance name defaults to the logical element's name; FQN = path to the node + instance name.

### Deployment relationships
Inherited from logical model; deployment-specific ones definable with same `->` / `-[kind]->` syntax and same properties. Relationships connect **instances**, not logical elements.

Also has its own `views { }` for deployment views (see `/dsl/deployment/views/`).

---

## 9. Imports / Mixins / `#include`

**There is no `import`, `include` (file-level), or `#include` directive.** Composition mechanisms are:

1. **Implicit file merging**: every `.c4`/`.likec4` file in the workspace is parsed and merged into one model. No declaration needed. Top-level statements across files concatenate.

2. **`extend <element> { ... }`**: the only explicit cross-file mechanism — opens an already-defined element's scope to add children/relationships/properties from another file. Same for `extend <node>` and `extend <a -> b>`.

3. **Multi-project workspaces** (config-level, not DSL): separate `likec4.config.json` per project folder; referenced via Vite plugin / API. This is project configuration, not a DSL statement.

So a converter targeting a single output file does not need any import syntax.

---

## 10. Real LikeC4 Source — bigbank.c4

A genuine, complete file from the likec4 repo (228 lines, the canonical showcase). Shows nesting (3 levels), 10 distinct element kinds, ~25 relationships, 8 views with predicates/styles. Quoted in full because it is the single best reference for the actual syntax surface.

**Source:** `apps/playground/src/examples/bigbank/bigbank.c4` (mirror at `apps/docs/src/components/bigbank/bigbank.c4`) in `github.com/likec4/likec4`.

**Model portion (lines 1–117):**
```likec4
model {

  customer = person "Personal Banking Customer" {
    description "A customer of the bank, with personal bank accounts."
  }

  bigbank = enterprise "Big Bank plc" {

    supportStaff = staff "Customer Service Staff" {
      description: "Customer service staff within the bank."
    }

    backoffice = staff "Back Office Staff" {
      description: "Administration and support staff within the bank."
    }

    mainframe = existingsystem "Mainframe Banking System" {
      description: "Stores all of the core banking information about customers, accounts, transactions, etc."
    }

    email = existingsystem "E-mail System" {
      description: "The internal Microsoft Exchange e-mail system."
    }

    atm = existingsystem "ATM" {
      description: "Allows customers to withdraw cash."
    }

    internetBankingSystem = softwaresystem "Internet Banking System" {
      description: "Allows customers to view information about their bank accounts, and make payments."

      singlePageApplication = spa "Single-Page Application" {
        description: "Provides all of the Internet banking functionality to customers via their web browser."
        technology: "JavaScript and Angular"
        style {
          shape browser
        }
      }
      mobileApp = mobileApp "Mobile App" {
        description: "Provides a limited subset of the Internet banking functionality to customers via their mobile device."
        technology: "Xamarin"
      }
      webApplication = container "Web Application" {
        description: "Delivers the static content and the Internet banking single page application."
        technology: "Java and Spring MVC"
      }
      apiApplication = container "API Application" {
        description: "Provides Internet banking functionality via a JSON/HTTPS API."
        technology: "Java and Spring MVC"

        signinController = component "Sign In Controller" {
          description: "Allows users to sign in to the Internet Banking System."
          technology: "Spring MVC Rest Controller"
        }
        accountsSummaryController = component "Accounts Summary Controller" { /* ... */ }
        resetPasswordController = component "Reset Password Controller" { /* ... */ }
        securityComponent = component "Security Component" { /* ... */ }
        mainframeBankingSystemFacade = component "Mainframe Banking System Facade" { /* ... */ }
        emailComponent = component "E-mail Component" { /* ... */ }
      }
      database = database "Database" {
        description: "Stores user registration information, hashed authentication credentials, access logs, etc."
        technology: "Oracle Database Schema"
      }
    }
  }

  // relationships between people and software systems
  customer -> internetBankingSystem "Views account balances, and makes payments using"
  internetBankingSystem -> mainframe "Gets account information from, and makes payments using"
  internetBankingSystem -> email "Sends e-mail using"
  email -> customer "Sends e-mails to"
  customer -> supportStaff "Asks questions to"
  supportStaff -> mainframe "Uses"
  customer -> atm "Withdraws cash using"
  atm -> mainframe "Uses"
  backoffice -> mainframe "Uses"

  // relationships to/from containers
  customer -> webApplication "Visits bigbank.com using HTTPS"
  customer -> singlePageApplication "Views account balances, and makes payments using"
  customer -> mobileApp "Views account balances, and makes payments using"
  webApplication -> singlePageApplication "Delivers to the customer's web browser"

  // relationships to/from components
  singlePageApplication -> signinController "Makes API calls to"
  singlePageApplication -> accountsSummaryController "Makes API calls to"
  singlePageApplication -> resetPasswordController "Makes API calls to"
  mobileApp -> signinController "Makes API calls to"
  mobileApp -> accountsSummaryController "Makes API calls to"
  mobileApp -> resetPasswordController "Makes API calls to"
  signinController -> securityComponent "Uses"
  accountsSummaryController -> mainframeBankingSystemFacade "Uses"
  resetPasswordController -> securityComponent "Uses"
  resetPasswordController -> emailComponent "Uses"
  securityComponent -> database "Reads from and writes to"
  mainframeBankingSystemFacade -> mainframe "Makes API calls to"
  emailComponent -> email "Sends e-mail using"
}
```

**Views portion (lines 119–228):**
```likec4
views {

  view index of bigbank {
    title "Big Bank - Landscape"
    include *
  }

  view context of bigbank {
    title "Internet Banking System / SystemContext"
    include
      bigbank,
      mainframe,
      internetBankingSystem,
      email,
      customer
  }

  view ibsContainers of internetBankingSystem {
    title "Internet Banking System / Containers"
    include
      *,
      -> customer
  }

  view customer of customer {
    include
      *,
      customer -> internetBankingSystem.*,
      customer -> bigbank.*
    exclude webApplication
  }

  view support of supportStaff {
    include
      *,
      bigbank,
      -> backoffice ->
  }

  view apiApp of internetBankingSystem.apiApplication {
    title "Applications / API Application"
    include *
    style * { color muted }
    style singlePageApplication, mobileApp { color secondary }
    style apiApplication, apiApplication.* { color primary }
  }

  view webapp of webApplication {
    title "Applications / Web Application"
    include *, internetBankingSystem, bigbank
    style bigbank { color muted }
    style internetBankingSystem.* { color secondary }
  }

  view mobileApp of mobileApp {
    title "Applications / Mobile Application"
    include
      *,
      internetBankingSystem,
      internetBankingSystem.apiApplication,
      mobileApp -> internetBankingSystem.apiApplication.*
    style * { color secondary }
    style apiApplication, internetBankingSystem { color muted }
    style mobileApp { color green }
  }

  view spa of singlePageApplication {
    title "Applications / Single Page Application"
    include *, apiApplication, internetBankingSystem, -> singlePageApplication ->
    style internetBankingSystem { color muted }
    style singlePageApplication { color green }
  }
}
```

**Key observations from this real file:**
- The spec block (not shown; lives in a separate `specification` snippet in the showcase) declares the 10 kinds used: `person`, `enterprise`, `staff`, `existingsystem`, `softwaresystem`, `spa`, `mobileApp`, `container`, `component`, `database`.
- Note `spa` and `mobileApp` are used both as **kind** and as **element name** (e.g. `mobileApp = mobileApp "Mobile App"`) — legal because kinds and names live in different namespaces.
- Relationships are mostly defined **flat at the end of `model`**, not nested — both styles are used in practice.
- Views reference elements by bare name when unambiguous (e.g. `apiApplication.*`) — lexical scoping + uniqueness resolution at work.
- `:` after `description`/`technology` is optional (`description: "..."` vs `description "..."`).

---

## Sources

Primary documentation scraped from likec4.dev:

- https://likec4.dev/dsl/intro/ — top-level statements, file structure
- https://likec4.dev/dsl/specification/ — element kinds, relationship kinds, tags, colors
- https://likec4.dev/dsl/model/ — elements, properties (title/description/summary/technology/tags/links/metadata), nesting
- https://likec4.dev/dsl/relationships/ — `->` syntax, kinds, properties, `this`/`it`, metadata
- https://likec4.dev/dsl/references/ — scope, hoisting, FQN resolution
- https://likec4.dev/dsl/extend/ — `extend` for elements/relationships/nodes, metadata merging
- https://likec4.dev/dsl/views/ — view definition, scoped views, extends
- https://likec4.dev/dsl/views/predicates/ — include/exclude, selectors (`.*`/`.**`/`._`), `where`, `with`, groups, styles, autoLayout, rank
- https://likec4.dev/dsl/styling/ — shape/color/size/opacity/border/icon, relationship line/head/tail, all enums
- https://likec4.dev/dsl/deployment/model/ — deployment nodes, `instanceOf`, deployment relationships
- https://likec4.dev/showcases/bigbank/ — structured showcase comparing to Structurizr

Real source file:
- https://github.com/likec4/likec4/blob/main/apps/playground/src/examples/bigbank/bigbank.c4 (228-line canonical example, quoted in §10)

Negative findings (features confirmed **not** to exist despite being plausible):
- No built-in element kinds — confirmed across spec docs.
- No element-kind subtyping / `kind.of` hierarchy — confirmed by grammar absence and flat `element.kind` predicate semantics.
- No `spec.preset` DSL keyword — themes are project-config, not DSL.
- No file-level `import` / `#include` — only implicit merge + `extend`.
