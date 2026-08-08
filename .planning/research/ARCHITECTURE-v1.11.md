# Architecture Research

**Domain:** LikeC4 → C4Drill converter (Stage 0 of the parse pipeline)
**Researched:** 2026-08-08
**Confidence:** HIGH

> **⚠ DECISION OVERRIDE (2026-08-09, BINDING):** The user has chosen **Pigeon (PEG)** as the parser technology. The parser/lexer/package-layout sections below were originally written for a hand-written recursive-descent parser. The binding layout is: `grammar.peg` (source) + `grammar.go` (generated via `go generate`) replace `parser.go` + `lexer.go`; the converter (`convert.go`), kind mapper (`kindmap.go`), resolver (`resolve.go`), warning collector (`warnings.go`), and public entry (`likec4.go`) are UNCHANGED. The two-phase `Parse`/`Convert` API still holds — Pigeon's generated `Parse()` produces the AST that `Convert()` consumes. See the STACK.md "Decision Override: PEG Parser" addendum for the full rationale and the accepted recovery-rule tradeoff.

## Standard Architecture

### System Overview

```
┌─────────────────────────────────────────────────────────────────┐
│  cmd/c4drill/root.go  runRoot — Stage 0 routing by extension    │
├─────────────────────────────────────────────────────────────────┤
│   .c4 / .likec4               │            .toml                 │
│         ▼                      │              ▼                  │
│  ┌─────────────────────┐       │   ┌──────────────────────┐      │
│  │ internal/likec4     │       │   │ internal/parser      │      │
│  │  ParseFile(path)    │       │   │  ParseFile(path)     │      │
│  │   └─ lexer → parser │       │   │   └─ go-toml v2      │      │
│  │      → ast.File     │       │   └──────────┬───────────┘      │
│  │   └─ Convert(ast)   │       │              │                  │
│  │      → *parser.Model│       │              │                  │
│  └──────────┬──────────┘       │              │                  │
├─────────────┴──────────────────┴──────────────┴──────────────────┤
│            *parser.Model  (single converged type)                │
├──────────────────────────────────────────────────────────────────┤
│  include.Resolve → template.Expand → peer.Resolve → Validate     │
│  → view.Generate → graph.Build → render → output    (UNCHANGED)  │
└──────────────────────────────────────────────────────────────────┘
```

The converter is the **only** code that knows LikeC4 was the source. Everything
downstream of the `*parser.Model` convergence point is reused verbatim — the
v1.10 composition pipeline, validator, view generator, graph builder, and
renderer never see the difference. This is the load-bearing architectural
invariant: no `*parser.Model` field gains a "source format" tag, no downstream
package imports `internal/likec4`.

### Component Responsibilities

| Component | Responsibility | Typical Implementation |
|---|---|---|
| `internal/likec4` (root) | Public entry: `ParseFile`, `Convert`, types | Thin wrappers + conversion |
| `internal/likec4/lexer` (lexer.go) | Tokenize LikeC4 source (idents, strings, `{`, `}`, `->`, `-[x]->`, `<->`, `#tag`, `//`) | Hand-written scanner, `[]Token` stream |
| `internal/likec4/parser` (parser.go) | Recursive-descent over tokens → `*ast.File` | One function per grammar rule |
| `internal/likec4/ast` (ast.go) | AST types: `File`, `Specification`, `Model`, `Element`, `Relationship`, etc. | Plain structs with `Pos` |
| `internal/likec4/convert` (convert.go) | `*ast.File` → `*parser.Model`; warning collection; kind mapping | Single struct `conv` with context |
| `cmd/c4drill/root.go` | Extension routing: dispatch to converter or TOML parser | One `switch` block (Stage 0) |

## Recommended Project Structure

```
internal/
└── likec4/
    ├── likec4.go          # package doc + ParseFile/Convert public entry points
    ├── lexer.go           # tokenizer
    ├── parser.go          # recursive-descent: parseFile → *ast.File
    ├── ast.go             # AST node types
    ├── convert.go         # Convert(*ast.File) (*parser.Model, error)
    ├── kindmap.go         # LikeC4 kind string → model.UnitType fuzzy table
    ├── resolve.go         # this/it/sourceless `->` + dot-path name resolution
    ├── warnings.go        # WarnCollector type
    ├── lexer_test.go
    ├── parser_test.go
    ├── convert_test.go
    └── testdata/
        ├── bigbank.c4     # the canonical 228-line showcase
        ├── minimal.c4
        └── golden/        # expected *parser.Model dumps (canonical form)
```

### Structure Rationale

- **One package, not split:** `internal/likec4/` holds lexer+parser+ast+converter
  together. The proposal to split into `internal/likec4/` (parse) + `internal/likec4/convert/`
  is **rejected** — `convert.go` needs access to AST internals and would create a
  circular-feeling split for no isolation benefit. A single package with clear
  file boundaries (lexer.go, parser.go, ast.go, convert.go) mirrors how the
  existing `internal/parser` keeps its multi-pass logic in one package.
- **`testdata/bigbank.c4`:** the showcase file from §10 of the DSL brief is the
  single best regression corpus — real syntax, 3 levels of nesting, 10 kinds,
  ~25 relationships, 8 views (all dropped).

## Architectural Patterns

### Pattern 1: Parse/Convert split (two-phase, one public surface)

**What:** The public API is split into two functions so each is independently
testable, but `ParseFile` chains them for the CLI's convenience.

**When to use:** Always — the split is mandatory for testability (lex/parse
golden tests must not require assembling a full `*parser.Model`).

**Trade-offs:** Slightly more public surface (2 entry points vs 1); worth it.

**Example:**
```go
// likec4.go

// ParseFile reads a LikeC4 (.c4/.likec4) file and converts it to a
// *parser.Model suitable for the existing c4drill pipeline. It is the
// Stage 0 entry point routed by file extension in cmd/c4drill/root.go.
// Warnings about dropped constructs are emitted to stderr via w; pass
// io.Discard to silence. Never returns an error for unsupported
// constructs — only for syntactic/lexical failure.
func ParseFile(path string, w io.Writer) (*parser.Model, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("likec4 read %q: %w", path, err)
    }
    f, err := Parse(data)
    if err != nil {
        return nil, fmt.Errorf("likec4 parse %q: %w", path, err)
    }
    return Convert(f, w)
}

// Parse lexes+parses LikeC4 source bytes into an AST. Pure; no I/O.
func Parse(src []byte) (*ast.File, error)

// Convert turns an AST into a *parser.Model. w receives one warning per
// dropped construct TYPE (deduped), never per occurrence.
func Convert(f *ast.File, w io.Writer) (*parser.Model, error)
```

### Pattern 2: Single converged type (`*parser.Model`)

**What:** The converter emits exactly the type `parser.ParseFile` would. No
intermediate DTO, no tagged union on the model.

**When to use:** Always — this is the milestone's core invariant.

**Trade-offs:** Some LikeC4 semantics (relationship kinds, view predicates,
metadata) have no representation and are dropped. Acceptable per the milestone
brief (breadth over fidelity).

The converter constructs the model using the same invariants the TOML parser
establishes:
- `Units map[string]*model.Unit` keyed by **dotted absolute path** at the root
  level only contains **top-level** elements (e.g. `customer`, `bigbank`); nested
  elements live under their parent's `Subunits`.
- `UnitOrder []string` preserves **source declaration order** of top-level units.
- Each `Unit.Subunits map[string]*model.Unit` is keyed by the **last path
  segment** (not the full path), and `SubunitOrder []string` preserves the order
  children appeared in the LikeC4 `{}` body.
- `Links []model.Link` carries outgoing relationships; `Link.Peer` is the
  **absolute dotted path** of the target (resolved at conversion time — see
  Pattern 4).
- `Templates`/`Instantiations`/`Includes` are left **nil/empty** — LikeC4 has
  no template or include directives, so the downstream `template.Expand` and
  `include.Resolve` stages become no-ops (their existing "empty ⇒ return m
  unchanged" guards already handle this).
- `Properties` is populated from `specification {}` only if a project name is
  derivable; otherwise left zero-value (C4Drill treats empty properties as the
  default transparent/straight style).

### Pattern 3: Lexical `{}` nesting → dotted subunit tree

**What:** LikeC4 nesting is purely lexical (braces). The converter walks the
AST element tree and flattens it into C4Drill's dotted-path subunit model.

**When to use:** Always — there is no `parent` attribute in LikeC4.

**Trade-offs:** None — this is a 1:1 structural map.

**Example:**
```likec4
model {
  cloud = system 'Our SaaS' {     // → Units["cloud"], Type=system
    ui = component 'Frontend'     //   → Subunits["ui"], Type=component
    api = component 'Backend' {   //   → Subunits["api"], Type=component
      auth = component 'Auth'     //     → Subunits["auth"]
    }
    ui -> api 'requests'          //   → cloud.Links: {Peer:"cloud.api", ...}
  }                               //      (resolved to absolute at convert time)
}
```

The walk is a single recursive function that carries `parentPath string`:

```go
func (c *conv) emitElement(e *ast.Element, parentPath string, parentType model.UnitType) {
    path := e.Name
    if parentPath != "" {
        path = parentPath + "." + e.Name
    }
    unit := &model.Unit{
        Type:        c.kindMap(e.Kind, parentType),
        Name:        orDefault(e.Title, model.Humanize(e.Name)),
        Description: e.Description,
        Technology:  e.Technology,
        Reference:   firstLink(e.Links),  // first link becomes Reference; rest dropped+warned
    }
    // ... attach to parent's Subunits or to m.Units[root] ...
    for _, child := range e.Body.Elements   { c.emitElement(child, path, unit.Type) }
    for _, rel  := range e.Body.Relationships { c.emitRel(rel, path) }
}
```

`SubunitOrder` is appended in iteration order of `e.Body.Elements`, mirroring
how `parser.parseUnitWithOrder` builds it from `subunitOrders`.

### Pattern 4: Relationship resolution at convert time

**What:** `this`/`it`/sourceless `->` and bare-name targets are resolved to
**absolute dotted paths** before being written into `Link.Peer`. The downstream
`peer.Resolve` stage then sees only absolute paths and is a no-op (its
"absolute peers are untouched" branch).

**When to use:** Always — C4Drill links require absolute peers.

**Trade-offs:** Requires building a name index during the walk; manageable.

The resolution rules (from DSL brief §3, §4):

| LikeC4 form | Source | Target resolution |
|---|---|---|
| `a -> b` (top-level in `model`) | `a` (must be top-level FQN) | `b` resolved via scope |
| `-> b` (inside element `X {}`) | `X` (implicit) | `b` resolved in `X`'s scope |
| `this -> b` / `it -> b` (inside `X`) | `X` | `b` resolved in `X`'s scope |
| `a -> this` / `a -> it` (inside `X`) | `a` resolved | `X` |
| `frontend -> service1.backend.api` | `frontend` | already-absolute dotted path |

Resolution is lexical-scope + hoisting (like JS): short names resolve if unique
in scope; otherwise full/partial FQN required. The converter builds a scope map
per element body and a global FQN index, then resolves targets. **A miss is a
hard error** (returned from `Convert`), matching the v1.10 hard-error stance —
unlike dropping unsupported constructs (soft warn), an unresolved reference is
genuine model corruption.

`<->` emits **two** `model.Link` entries (forward + reverse). Relationship kind
(`-[async]->`) is **dropped** (C4Drill links carry only `Technology`, no kind
taxonomy); the `title`/`technology` inline strings populate `Link.Description`
and `Link.Technology` respectively.

### Pattern 5: Warning channel via `io.Writer` + dedup-by-type

**What:** The converter takes an `io.Writer` (the CLI passes `cmd.OutOrStderr()`).
A `WarnCollector` struct dedups warnings by **construct type**, not occurrence.

**When to use:** Always — the milestone contract is "one warning per construct
type, never fatal".

**Trade-offs:** Hides the count of dropped constructs; acceptable (the goal is
"tell me something was dropped", not a full audit).

```go
// warnings.go
type WarnCollector struct {
    w   io.Writer
    seen map[string]bool  // keyed by construct type tag
}

func (c *WarnCollector) Warn(construct, detail string) {
    if c.seen[construct] { return }
    c.seen[construct] = true
    fmt.Fprintf(c.w, "warning: likec4: %s dropped (%s)\n", construct, detail)
}
```

Construct type tags (the dedup keys), emitted once each per file:

| Tag | When |
|---|---|
| `deployment-block` | `deployment {}` top-level statement encountered |
| `views-block` | `views {}` top-level statement encountered |
| `global-block` | `global {}` top-level statement encountered |
| `metadata` | any `metadata {}` block on element/relationship |
| `tags` | any `#tag` on element/relationship |
| `icon` | any `icon`/`iconColor`/`iconSize`/`iconPosition` style property |
| `relationship-kind` | any `-[kind]->` or `.kind` relationship |
| `style-block` | any per-element or per-relationship `style {}` block |
| `element-extra-links` | an element with >1 `link` (only first becomes `Reference`) |
| `extend` | any `extend <element> {}` block (merged at AST level, but warned once if used) |

The CLI wiring in `runRoot`:

```go
// Stage 0: Route by extension. .c4/.likec4 → converter; else → TOML parser.
var m *parser.Model
switch ext := strings.ToLower(filepath.Ext(inputPath)); ext {
case ".c4", ".likec4":
    m, err = likec4.ParseFile(inputPath, cmd.OutOrStderr())
    if err != nil {
        return fmt.Errorf("likec4: %w", err)
    }
default:
    m, err = parser.ParseFile(inputPath)
    if err != nil {
        return fmt.Errorf("parse: %w", err)
    }
}
```

## Kind Mapping Table

LikeC4 has **no built-in kinds** — every kind is user-declared in `specification`.
The converter maps each kind string to a C4Drill `model.UnitType` via
three-tier fuzzy match in `kindmap.go`:

```go
// kindmap.go
var kindExact = map[string]model.UnitType{
    "person":      model.TypePerson,
    "actor":       model.TypePerson,        // LikeC4 convention for person
    "system":      model.TypeSystem,
    "softwaresystem": model.TypeSystem,
    "enterprise":  model.TypeSystem,        // grouping at top level
    "container":   model.TypeContainer,
    "component":   model.TypeComponent,
    "db":          model.TypeDb,
    "database":    model.TypeDb,
    "queue":       model.TypeQueue,
    "box":         model.TypeBox,
    "fallback":    model.TypeBox,           // not a real kind; see resolver
}

func (c *conv) kindMap(kind string, parentType model.UnitType) model.UnitType {
    k := strings.ToLower(strings.TrimSpace(kind))
    // Tier 1: exact match.
    if t, ok := kindExact[k]; ok {
        return c.adjustForLevel(t, parentType)  // db→containerDb at C2, etc.
    }
    // Tier 2: substring match against known hints.
    for hint, t := range kindSubstringHints {   // "data"→db, "queue"→queue, ...
        if strings.Contains(k, hint) {
            return c.adjustForLevel(t, parentType)
        }
    }
    // Tier 3: fallback.
    return c.adjustForLevel(model.TypeBox, parentType)
}
```

| LikeC4 kind (case-insensitive) | Tier | C4Drill UnitType (C1 / C2 / C3) |
|---|---|---|
| `person`, `actor` | exact | person / person / person |
| `system`, `softwaresystem` | exact | system / system / system |
| `enterprise` | exact | system / system / system |
| `container` | exact | system / container / container |
| `component` | exact | system / container / component |
| `db`, `database` | exact | db / containerDb / componentDb |
| `queue` | exact | queue / containerQueue / componentQueue |
| `box` | exact | box / box / box |
| contains `data` | substring | db / containerDb / componentDb |
| contains `queue` | substring | queue / containerQueue / componentQueue |
| contains `service`, `app`, `api` | substring | system / container / component |
| contains `storage`, `cache` | substring | db / containerDb / componentDb |
| _(anything else)_ | fallback | box / containerBox / componentBox |

The level adjustment reuses the **same generic-type inference logic** the TOML
parser applies (`inferGenericType` in `internal/parser/parser.go:699`) — db/queue
promote to containerDb/containerQueue at C2 and componentDb/componentQueue at C3.
Non-generic types (person, system, component, box) are not level-adjusted except
that boxes nest as same-level groupings (`box`/`containerBox`/`componentBox`).

A kind that hits Tier 2 or Tier 3 does **not** emit a warning — the milestone
brief explicitly says fuzzy kind matching is not a hard error and the warning
budget is reserved for genuinely dropped constructs (views, metadata, icons).

## Data Flow

### Stage 0 → Stage 1 Convergence

```
.c4 file
   ▼
[lexer] bytes → []Token
   ▼
[parser] tokens → *ast.File (specification + model + views + ...)
   ▼
[Convert] ast → *parser.Model
   │          ├─ build FQN index + scope map (pass 1 over model)
   │          ├─ emit elements → Units/Subunits/SubunitOrder (pass 2)
   │          ├─ emit relationships → Links (pass 2, resolved absolute)
   │          ├─ WarnCollector dedups dropped-construct warnings
   │          └─ drops: deployments, views, global, metadata, tags, icons,
   │                    relationship kinds, style blocks
   ▼
*parser.Model  ──── converges with parser.ParseFile output ────┐
   ▼                                                          │
include.Resolve → template.Expand → peer.Resolve → Validate   │
   (all no-ops for LikeC4 input: no [[include]], no templates,│
      all peers already absolute)                              │
   ▼                                                          │
view.Generate → graph.Build → render → output (UNCHANGED) ◀──┘
```

### Key Data Flows

1. **Element tree flow:** AST `*Element` with nested `Body.Elements` →
   recursive `emitElement` carrying `parentPath` → dotted-path-keyed
   `Subunits` map + ordered `SubunitOrder` slice.
2. **Relationship flow:** AST `*Relationship` (source expr, target expr,
   title, technology) → `resolveRef` against scope/FQN index → absolute
   `Link.Peer` → appended to source unit's `Links`. `<->` becomes two links.
3. **Warning flow:** dropped construct encountered → `WarnCollector.Warn(tag, detail)`
   → first occurrence only prints `warning: likec4: <tag> dropped (<detail>)`
   to stderr → subsequent occurrences of same tag silently absorbed.

## Build Order

Strict dependency chain — each layer depends only on the layers above it:

```
1. internal/likec4/ast.go              (no deps; pure struct defs)
2. internal/likec4/lexer.go            (deps: ast for Pos)
3. internal/likec4/lexer_test.go       (deps: lexer) — gate before parser
4. internal/likec4/parser.go           (deps: ast, lexer)
5. internal/likec4/parser_test.go      (deps: parser) — gate before convert
6. internal/likec4/warnings.go         (no deps; io.Writer + map)
7. internal/likec4/kindmap.go          (deps: internal/model)
8. internal/likec4/resolve.go          (deps: ast, internal/model)
9. internal/likec4/convert.go          (deps: ast, model, parser, warnings,
                                        kindmap, resolve) — the keystone
10. internal/likec4/likec4.go          (deps: convert, parser) — ParseFile/Parse
11. internal/likec4/convert_test.go    (deps: likec4 full pkg, testdata/bigbank.c4)
12. cmd/c4drill/root.go edit           (deps: internal/likec4) — Stage 0 switch
13. cmd/c4drill/root_test.go           (new — extension routing regression)
```

**Rationale for the order:**
- **Lexer before parser:** parser cannot be tested until tokens are stable.
- **Parser before convert:** convert consumes `*ast.File`; AST must be parseable
  from real `.c4` files (use `testdata/bigbank.c4` as the gate) before model
  emission can be debugged.
- **AST, warnings, kindmap, resolve before convert:** convert is the integration
  point that pulls everything together; it must land last among the package files.
- **`likec4.go` after convert:** the public `ParseFile` chains Parse+Convert.
- **root.go last:** the CLI wiring is a one-block diff; only touches Stage 0 and
  leaves Stages 1–6 byte-identical. The hard regression contract (every existing
  `.toml` fixture parses byte-identical) is enforced because `.toml` inputs never
  enter the new code path.

## Anti-Patterns

### Anti-Pattern 1: Tagging the model with its source format

**What people do:** Add a `SourceFormat string` or `IsLikeC4 bool` to
`*parser.Model` so downstream stages can branch.
**Why it's wrong:** Violates the milestone invariant ("the converter is the ONLY
code that knows LikeC4 was the source"). Every downstream package gains a
dependency on the source format, destroying the clean Stage 0 boundary.
**Do this instead:** Emit a structurally indistinguishable `*parser.Model`. If
the converter needs to defer peer resolution, it writes absolute paths and lets
the existing `peer.Resolve` no-op them.

### Anti-Pattern 2: Emitting warnings per occurrence

**What people do:** Print `warning: tags dropped` for every `#tag` encountered.
**Why it's wrong:** A model with 50 tags floods stderr; users learn to ignore
all warnings, including the meaningful ones.
**Do this instead:** `WarnCollector` with a `seen map[string]bool` keyed by
construct type — one `tags dropped` line per file no matter how many tags.

### Anti-Pattern 3: Generating TOML text and re-parsing it

**What people do:** Convert LikeC4 → TOML string → `parser.Parse([]byte(toml))`.
**Why it's wrong:** Doubles the parse cost, introduces string-escaping bugs,
and loses source-position information for error messages.
**Do this instead:** Construct `*parser.Model` structurally — the TOML parser's
own `parseUnitWithOrder` shows the target shape; the converter builds it directly.

### Anti-Pattern 4: Hard erroring on unsupported constructs

**What people do:** Return an error when encountering `views {}` or an unknown
kind.
**Why it's wrong:** The milestone contract is "breadth over fidelity, never
fatal on unsupported constructs". A `views {}` block is present in essentially
every real LikeC4 file; erroring would make the converter useless on the
canonical `bigbank.c4` showcase.
**Do this instead:** Warn-and-drop for unsupported constructs (views, metadata,
icons, tags, relationship kinds, style blocks, deployment/global blocks). Hard
error only for genuine model corruption (unresolved references, malformed syntax).

## Integration Points

### Internal Boundaries

| Boundary | Communication | Notes |
|---|---|---|
| `cmd/c4drill` ↔ `internal/likec4` | `likec4.ParseFile(path, io.Writer)` call | New import; sole entry. `io.Writer` = `cmd.OutOrStderr()`. |
| `cmd/c4drill` ↔ `internal/parser` | `parser.ParseFile(path)` call | Unchanged; same call site, now in the `default` arm of the switch. |
| `internal/likec4` ↔ `internal/parser` | Imports `parser.Model` type only | Does NOT call `parser.Parse`; constructs the model structurally. |
| `internal/likec4` ↔ `internal/model` | Imports `Unit`, `Link`, `UnitType`, `Humanize` | Reuses `model.Humanize` for default names (matches v1.10 ERGO-03). |
| Downstream stages ↔ source format | None | `include.Resolve`, `template.Expand`, `peer.Resolve`, `Validate`, `view`, `graph`, `render`, `output` all unchanged. |

### External Services

None. The converter is pure Go — no TypeScript/Node runtime, no CGO, no
external grammar dependency, no network. This is an explicit milestone
requirement (matches the "Hand-written LikeC4 parser" target feature).

## Sources

- `.planning/PROJECT.md` — v1.11 milestone scope, hard contracts, kind-mapping intent
- `.planning/research/likec4-dsl-brief.md` — LikeC4 grammar facts (§2 kinds, §3 relationships, §4 nesting, §5 views dropped, §9 no imports)
- `cmd/c4drill/root.go` — pipeline at line 117 (`parser.ParseFile`), stderr pattern (`cmd.OutOrStderr()` at line 157)
- `internal/parser/parser.go` — `Model` struct (line 36), `ParseFile` (line 746), `parseUnitWithOrder` (line 572), `inferGenericType` (line 699)
- `internal/model/unit.go` — `Unit` struct (line 43), `Subunits`/`SubunitOrder`, `Clone` (line 93)
- `internal/model/link.go` — `Link` struct (line 43), `Peer`/`Mirror`
- `internal/peer/resolve.go` — `Resolve(m *parser.Model) error` signature (line 45), absolute-peer passthrough contract

---
*Architecture research for: LikeC4 → C4Drill Stage 0 converter (v1.11)*
*Researched: 2026-08-08*
