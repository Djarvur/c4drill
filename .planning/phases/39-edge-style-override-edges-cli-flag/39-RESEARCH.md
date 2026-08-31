# Phase 39: Edge Style Override (`--edges` CLI flag) - Research

**Researched:** 2026-08-31
**Domain:** Go CLI flag surface + edge-style data flow (cobra flag → view override → graph builder → graphviz `splines`)
**Confidence:** HIGH — every claim below was verified against the current working tree on 2026-08-31 (file:line citations). This is an internal codebase change with no new dependencies and no external-API surface.

## Summary

Phase 39 adds a persistent `--edges <style>` flag that overrides the edge routing style for a whole invocation. The codebase is already fully prepared for it: `Graph.EdgeStyle` (`internal/graph/graph.go:66`) carries the resolved style to the converter, and `configureGraphSettings` (`internal/render/converter.go:264-271`) already maps all four enum values (`spline`→`splines="true"`, `straight`→`splines="false"`, `ortho`/`square`→`splines="ortho"`) with zero render-side changes needed. The work is confined to (1) flag registration + loud enum validation in `cmd/c4drill/root.go`, (2) threading the override onto every generated view at the PLAIN-01 choke points (`processView` root.go:330-338 and `processExpandedView` root.go:386-391), and (3) making the explicit flag survive `--plain`'s suppression, which today lives in the graph builders (`BuildGraph` builder.go:38-49, `BuildExpandedGraph` builder.go:411-433) as `if v.Plain { edgeStyle = "" }`.

The one genuine design subtlety is D-05 (explicit `--edges` beats `--plain`): simply assigning `v.Edges` at the threading point is NOT sufficient, because the builders zero `View.Edges` under `--plain` regardless of provenance. The override must be distinguishable from a model-derived value — a dedicated, default-empty `View` field applied AFTER the plain zeroing implements the precedence chain (explicit CLI > `--plain` suppression > model value) with zero flag-off behavior change (D-04). Two mechanism options are documented below; the planner picks the exact point per CONTEXT.md ("D-03 constrains the semantics, not the mechanism").

Testing rides existing infrastructure: the KEY-03 switch-matrix E2E (`cmd/c4drill/root_test.go:1998-2233`) is the direct template for GEDGE-07, the `generateFixtureOutput` helper (root_test.go:1951) runs the CLI with arbitrary extra flags, and `cmd/c4drill/testdata/plain.toml:8` already carries `edges = "straight"` (commented "author edge routing — ignored under --plain") — the ideal fixture for both the plain-composition pin and the flag-off golden invariance.

**Primary recommendation:** One small TDD plan group: register + validate `--edges` early in `runRoot` (sentinel-error pattern), thread it as a dedicated override onto both view-processing paths, apply it in the graph builders after the PLAIN-02 zeroing, pin everything with unit + matrix E2E tests asserting `splines=` in RAW dot output, keep every existing golden untouched, then docs (README.adoc CLI Reference + `--plain` delta + 3 byte-identical SKILL.md copies) and the v1.23.0 release per the 38-06 precedent.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

- **D-01:** Flag name is `--edges <style>`, mirroring the TOML key (`edges = "spline"`), consistent with `--expanded` mirroring `expanded`. Verified no collision: existing persistent flags are `--format/-f`, `--output/-o`, `--expanded`, `--plain`, `--no-colors`, `--no-styles`, `--no-length`, `--no-rank`, `--no-labels`, `--label-ratio` (`cmd/c4drill/root.go:96-114`).
- **D-02:** Invalid values fail loudly before any output, naming the offending value and the allowed enum — reuse the existing sentinel-wrap pattern (`fmt.Errorf("%w: %q", errInvalidFormat, format)` precedent in `cmd/c4drill/root.go`).
- **D-03 (Claude discretion):** The override is **invocation-global**: `--edges` replaces the *resolved* `View.Edges` for every generated view, so it beats BOTH the global `[properties] edges` AND per-unit `edges` values (the `cmp.Or(unit.Edges, properties.Edges)` chain at `internal/view/scope.go:504`/`:599`). Rationale: (a) the `--plain`/`--no-*` switch family is invocation-global and never respects per-unit opt-outs; (b) the motivating use case is "render this model as a variant" — mixed output (flag wins at C1 but loses to a unit override at C2) would be unpredictable; (c) overriding the resolved view value is the single-choke-point rule. Overriding only `properties.edges` was the alternative and is rejected.
- **D-04:** Flag absent → behavior is exactly today's: per-view resolution `cmp.Or(unit.Edges, properties.Edges)` (C2/C3) and `properties.Edges` (C1/expanded) unchanged. GEDGE-08 backward compat: existing canonicalDOT goldens pass untouched.
- **D-05:** An explicit `--edges <style>` **survives `--plain`** — user intent beats author-format suppression. Precedence order: explicit CLI flag > `--plain` suppression > model-derived value. Implementation note: PLAIN-02 currently empties `Graph.EdgeStyle` under plain (`internal/graph/builder.go:38-41`); the flag must be applied so an explicit value wins over the plain zeroing. This is a deliberate, test-pinned delta to `--plain`'s documented "exact union" contract (KEY-02) — the dedicated GEDGE-06 test and the README `--plain` doc must state the delta explicitly.
- **D-06:** Extend the v1.15 KEY-03-style switch matrix E2E: `--edges` × generation (root / drill-down / `--expanded`) × `--plain`, asserting the graphviz `splines` attribute in RAW dot output per combination. Add a dedicated test proving the flag beats a per-unit `edges` override (D-03 pin) and the GEDGE-06 `--plain` pin. Flag-off default: zero golden changes.
- **D-07 (plan-level note, not a requirement):** README.adoc usage/flags section + the 3 SKILL.md copies document `--edges` and the `--plain` delta, byte-identical sync per the 37-06 convention.

### Claude's Discretion
The user did not respond to the area-selection prompts in this session (autonomous run). D-03 (override scope) and D-01 (naming) were resolved by Claude using codebase evidence and family precedent; all decisions are veto-friendly — the user can amend CONTEXT.md before `/gsd:plan-phase 39`.

### Deferred Ideas (OUT OF SCOPE)
None — discussion stayed within phase scope.

### Folded Todos
- **Add CLI flag to override edge routing style** (`todos/pending/2026-08-30-add-cli-flag-to-override-edge-routing-style.md`, score 0.9, tagged `resolves_phase: 39`). This todo IS the phase's design source: problem statement, flow to extend, file list, naming check, `--plain` open question (resolved as D-05), and the switch-matrix E2E extension. Folded into scope verbatim as design input.

### Additional REQUIREMENTS.md constraints (phase-scoped exclusions)
- No new routing styles beyond `straight|spline|square|ortho` — reuse the existing enum + converter mapping (GEDGE-02).
- `--edges` must not change node/cluster formatting — formatting suppression is the KEY family's job.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| GEDGE-03 | User can override the edge routing style for a whole invocation via `--edges <style>` accepting `straight\|spline\|square\|ortho`, without editing the model file | Flag registration point verified (`root.go:96-115`, no collision); enum + converter mapping already complete (`converter.go:264-271`); threading points verified (`root.go:330-338`, `root.go:386-391`) |
| GEDGE-04 | Invalid `--edges` value fails loudly with an error naming the offending value and the allowed enum (no silent fallback) | Sentinel-wrap precedent verified (`errInvalidFormat` + `fmt.Errorf("%w: %q", ...)` at `root.go:28,140`); validation belongs in `runRoot` before Stage 1 (`root.go:128-155`); converter silently ignores unknown styles — so CLI-side validation is the only loud gate |
| GEDGE-05 | `--edges` overrides the model's `edges` property on every generated view — C1 root, all drill-down views, and the `--expanded` copy (PLAIN-01 threading pattern) | PLAIN-01 threading pattern verified at `root.go:330-338` (processView: C1+all C2/C3 drill-downs) and `root.go:386-391` (processExpandedView); resolved `View.Edges` values verified at `scope.go:25,133,504,599` |
| GEDGE-06 | `--edges` composes with `--plain`: explicit CLI request survives `--plain`'s author-format suppression, pinned by a dedicated test | PLAIN-02 zeroing verified at `builder.go:38-49` and `builder.go:411-433` — assignment to `View.Edges` alone is NOT sufficient; override must be applied after the plain zeroing (mechanism options below); KEY-02 "exact union" delta must be documented |
| GEDGE-07 | Switch-matrix E2E extended to `--edges` × generation (root / drill-down / `--expanded`) × `--plain`, asserting the graphviz `splines` attribute in RAW dot output | KEY-03 matrix verified as the direct template (`root_test.go:1998-2233`: `TestKeyComposition`, `keySwitches`, `runKeySwitchMatrix`, `generateFixtureOutput` helper at :1951); `plain.toml:8` fixture carries `edges = "straight"` |
| GEDGE-08 | Without the flag, output is unchanged — all existing canonicalDOT goldens pass untouched (backward compat) | Goldens verified in `cmd/c4drill/testdata/` (`plain.dot`, `plain.expanded.dot`, `nolabels.dot`, `nolabels.expanded.dot`, `multilevel*.svg/dot`, `expanded.dot`, `deepcross.dot`); default-empty override field preserves D-04 |
</phase_requirements>

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Flag surface + enum validation (GEDGE-03, GEDGE-04) | CLI (`cmd/c4drill/root.go`) | — | Cobra persistent flags and early loud validation are CLI-tier; `runRoot` validates before Stage 1 |
| Invocation-global override threading (GEDGE-05) | CLI (`processView`/`processExpandedView`) | View (`internal/view`) | The PLAIN-01 choke points are CLI-tier; the override carrier field is view-tier |
| Override vs `--plain` precedence (GEDGE-06) | Graph builder (`internal/graph/builder.go`) | — | The PLAIN-02 zeroing lives here; the explicit-flag-wins rule must be applied where the zeroing happens |
| Style → graphviz `splines` mapping | Render converter (`internal/render/converter.go`) | — | Already complete for all four enum values — zero changes |
| Matrix E2E + goldens (GEDGE-07, GEDGE-08) | CLI tests (`cmd/c4drill/root_test.go`) | Render tests | KEY-03 matrix + canonical goldens live at the CLI end-to-end tier |

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go (toolchain) | 1.26.1 (go.mod), 1.26.5 installed | Language | Project standard [VERIFIED: go.mod + `go version`] |
| `github.com/spf13/cobra` | v1.10.2 | CLI flags/commands | Already the flag framework — `PersistentFlags().StringVar` is the registration API [VERIFIED: go.mod, root.go:96-115] |
| `github.com/stretchr/testify` | v1.12.1 | Test assertions | Project standard, `require`/`assert` split [VERIFIED: go.mod, TESTING.md] |
| `github.com/onokonem/go-graphviz` | v0.10.0 | Graphviz rendering (WASM engine) | Already maps `splines` via `cg.SetSplines` [VERIFIED: go.mod, converter.go:264-271] |
| `golangci-lint` | v2.12.2 | Lint gate (`default: all`) | CI-blocking; all `//nolint:` need inline explanations [VERIFIED: `golangci-lint --version`, CONVENTIONS.md] |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `internal/canonical` | in-repo | Order-insensitive golden comparison (DI-1) | Golden comparisons in tests [VERIFIED: root_test.go TestNoLabelsC1Golden] |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| New `View` override field | Overwrite `View.Edges` + provenance bool | Equivalent behavior; separate field is clearer: default-empty = flag-off (D-04), no ambiguity with model-derived values, builder logic reads as precedence chain |
| Validator in flag `PreRunE` | Inline check at top of `runRoot` | Existing precedent (`format` check) is inline at `runRoot:138-141`; follow it — one validation style per file |

**Installation:** None — zero new packages. This phase modifies existing code only.

## Package Legitimacy Audit

> Not applicable — this phase installs no external packages. All libraries (cobra v1.10.2, testify v1.12.1, go-graphviz v0.10.0) are already locked in `go.mod` and vendored through the module proxy. No `go get` will be run.

**Packages removed due to slopcheck [SLOP] verdict:** none
**Packages flagged as suspicious [SUS]:** none

## Architecture Patterns

### System Architecture Diagram

```
                        ┌──────────────────────────────┐
                        │  CLI: c4drill model.toml     │
                        │  --edges spline [--plain]    │
                        └──────────────┬───────────────┘
                                       │
                        ┌──────────────▼───────────────┐
                        │ runRoot (root.go:128)        │
                        │  1. validate --edges enum    │   GEDGE-04: loud error
                        │     (NEW, before Stage 1)    │   errInvalidEdges + %q
                        │  2. parse → include → expand │
                        │     → peers → validate       │
                        └──────────────┬───────────────┘
                                       │
              ┌────────────────────────┼──────────────────────────┐
              │ per unit path          │                          │ --expanded
   ┌──────────▼──────────┐  ┌──────────▼──────────┐  ┌────────────▼───────────┐
   │ processView         │  │ (loop: C1 + C2/C3   │  │ processExpandedView    │
   │ (root.go:313)       │  │  drill-downs)       │  │ (root.go:376)          │
   │ GenerateC1/C2/C3View│  │                     │  │ GenerateExpandedView   │
   │   View.Edges =      │  │                     │  │   View.Edges =         │
   │   props|unit (res.) │  │                     │  │   properties.Edges     │
   │ NEW: thread override│  │                     │  │ NEW: thread override   │
   └──────────┬──────────┘  └──────────┬──────────┘  └────────────┬───────────┘
              └────────────────────────┼──────────────────────────┘
                                       │
                        ┌──────────────▼───────────────┐
                        │ BuildGraph /                 │
                        │ BuildExpandedGraph           │
                        │ (builder.go:38-49 / 411-433) │
                        │  edgeStyle := v.Edges        │
                        │  if v.Plain { "" }           │  ← PLAIN-02 zeroing
                        │  NEW: if override != "" {    │  ← GEDGE-06: explicit
                        │        edgeStyle = override} │    flag wins (after zeroing)
                        │  Graph.EdgeStyle = edgeStyle │
                        └──────────────┬───────────────┘
                                       │
                        ┌──────────────▼───────────────┐
                        │ configureGraphSettings       │
                        │ (converter.go:264-271)       │
                        │  spline→"true" straight→"false"
                        │  ortho|square→"ortho"        │  ← already complete,
                        │  ""→ (attribute unset)       │    ZERO changes
                        └──────────────┬───────────────┘
                                       │
                        ┌──────────────▼───────────────┐
                        │ RAW dot / SVG / HTML output  │
                        │ splines=... observable in dot│  ← GEDGE-07 assertion
                        └──────────────────────────────┘
```

### Recommended Project Structure

No new files or packages. Modifications land in existing files:

```
cmd/c4drill/
├── root.go                  # flag registration, errInvalidEdges sentinel, validation, threading
├── root_test.go             # flag tests, matrix extension, pins
└── testdata/
    └── (existing fixtures; goldens must NOT change)
internal/view/
├── view.go                  # override carrier field on View (if field-based mechanism chosen)
└── scope.go                 # UNCHANGED — resolution stays cmp.Or(unit.Edges, properties.Edges)
internal/graph/
├── builder.go               # apply override after PLAIN-02 zeroing (both BuildGraph + BuildExpandedGraph)
└── builder_test.go          # builder-level unit pins (plain × override)
internal/render/
└── converter.go             # UNCHANGED — SetSplines mapping already handles all four values
docs surface:
├── README.adoc              # CLI Reference flags block (~:1237-1248) + --plain section (~:1256-1317) delta
├── skill/SKILL.md           # byte-identical 3-copy sync (37-06 convention)
├── plugins/c4drill/skills/c4drill-toml/SKILL.md
└── plugins/c4drill/opencode/skills/c4drill-toml/SKILL.md
```

### Pattern 1: Persistent flag + sentinel-error validation (D-01, D-02)

**What:** Register a `PersistentFlags().StringVar` in the package-level `//nolint:gochecknoglobals` var block; validate the enum at the top of `runRoot` with a package sentinel wrapped with `%q`.
**When to use:** Exactly how `--format` works today — copy this pattern.
**Example (verified, cmd/c4drill/root.go:27-33, 50-62, 96-115, 138-141):**

```go
// Static errors for better error handling.
var (
	errInvalidFormat = errors.New("invalid format: must be dot, svg, or html")
	// NEW sentinel follows the same shape; message must name the allowed enum:
	// errInvalidEdges = errors.New("invalid edges: must be straight, spline, square, or ortho")
)

//nolint:gochecknoglobals // Cobra flags require package-level variables for PersistentFlags registration
var (
	expanded   bool
	plain      bool
	// NEW: edges string  — added to this block with the same nolint rationale
)

cmd.PersistentFlags().BoolVar(&plain, "plain", false, "Ignore author-custom formatting: ...")

// runRoot, before Stage 1 (file I/O):
if format != formatDot && format != formatSVG && format != formatHTML {
	return fmt.Errorf("%w: %q", errInvalidFormat, format)
}
```

### Pattern 2: PLAIN-01 flag threading (GEDGE-05)

**What:** Every generated view gets the flag value assigned at exactly two CLI choke points; the builders/readers consume the view fields.
**When to use:** For any invocation-global switch. `--edges` joins this family.
**Example (verified, cmd/c4drill/root.go:330-338 — the same assignment block exists at 386-391 for the expanded copy):**

```go
// PLAIN-01: thread --plain onto every generated view so the graph builder
// suppresses author-custom formatting (PLAIN-02). KEY-01: the granular
// switches thread the same way onto every view. LBL-02: --no-labels too.
v.Plain = plain
v.NoColors = noColors
v.NoStyles = noStyles
v.NoLength = noLength
v.NoRank = noRank
v.NoLabels = noLabels
```

### Pattern 3: PLAIN-02 suppression and the D-05 precedence chain (GEDGE-06)

**What:** The builders zero `View.Edges` under `--plain`; the explicit CLI override must be applied AFTER that zeroing so precedence is: explicit CLI flag > `--plain` suppression > model-derived value.
**When to use:** In BOTH `BuildGraph` (builder.go:38-49) and `BuildExpandedGraph` (builder.go:411-433) — they contain duplicated identical edge-style logic and MUST stay in sync.
**Example (verified, internal/graph/builder.go:38-49):**

```go
// PLAIN-02: properties.edges is ignored under plain — an empty EdgeStyle
// makes the renderer's configureGraphSettings fall through to the default
// routing.
edgeStyle := v.Edges
if v.Plain {
	edgeStyle = ""
}
// D-05/GEDGE-06: an explicit --edges CLI request is USER intent, not author
// formatting — it survives the plain zeroing above (applied here or via a
// dedicated View field; planner picks the exact mechanism).
```

**Mechanism options (planner decides; D-03 constrains semantics only):**
- **Option A (recommended):** New dedicated `View` field (e.g. `EdgesOverride string`, default empty). CLI threads it in the PLAIN-01 blocks; builders do `if v.EdgesOverride != "" { edgeStyle = v.EdgesOverride }` after the plain zeroing. Flag-off → empty field → byte-identical behavior (D-04 is structural, not just tested). Builder-level unit tests can pin plain×override directly.
- **Option B:** Overwrite `v.Edges` at the threading point plus a `EdgesFromCLI bool` marker the builders check to skip the plain zeroing. Same behavior, two coupled fields — no advantage over A.
- **Option C (rejected):** Set `g.EdgeStyle` from `cmd` after `BuildGraph...` returns. Works mechanically but moves the precedence rule out of the layer that owns suppression, and the D-05 pin becomes cmd-only instead of testable at the builder tier.

### Anti-Patterns to Avoid
- **Validating `--edges` in the render converter:** `configureGraphSettings` silently ignores unknown `EdgeStyle` values (bare switch, no default branch) — that is correct for internal use but means CLI-side validation is the ONLY loud gate (GEDGE-04). Never "fix" the converter to validate.
- **Changing `internal/view/scope.go` resolution:** D-03 overrides the *resolved* view value; do not touch the `cmp.Or(unit.Edges, m.Properties.Edges)` chains or add flag plumbing into view generators — the override lands after generation, not inside it.
- **Re-baselining goldens:** GEDGE-08 requires existing canonicalDOT goldens pass untouched. If a golden changes, the implementation is wrong — do not regenerate goldens for this phase.
- **`t.Parallel()` in render-touching tests:** the go-graphviz WASM engine is not concurrency-safe; every test that renders carries `//nolint:paralleltest // go-graphviz WASM engine has concurrency issues` (91 existing annotations).
- **Divergent duplicate logic:** `BuildGraph` and `BuildExpandedGraph` each contain the PLAIN-02 block. The override must be added to BOTH, identically — the KEY-03 matrix covers all three generations precisely to catch this class of drift.
- **Unexplained `//nolint:`:** golangci-lint v2 `default: all` blocks CI; any new nolint directive needs an inline explanation comment (CONVENTIONS.md).

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Enum validation errors | Custom error type/struct | Sentinel + `fmt.Errorf("%w: %q", ...)` wrap | Established repo pattern (root.go:28,140); errors.Is-friendly |
| Style → `splines` mapping | Any new mapping code | Existing `configureGraphSettings` switch | Already correct for all four enum values incl. `square`→ortho alias (GEDGE-02) |
| E2E CLI harness | New exec helpers | `generateFixtureOutput` (root_test.go:1951) + `NewRootCmd`/`SetArgs` | Existing, exercised by the whole KEY-03 matrix |
| Golden comparison | String equality on raw dot | `canonical.Canonical(t, ...)` (order-insensitive, DI-1) | Existing helper keeps goldens stable against cosmetic reordering |

**Key insight:** This phase is almost entirely *wiring* — the enum, the mapping, the threading precedent, the matrix harness, and the fixtures all exist. The only new logic is the precedence rule (D-05) and the loud validation (D-02).

## Common Pitfalls

### Pitfall 1: Assigning `View.Edges` and assuming it survives `--plain`
**What goes wrong:** GEDGE-06 test fails — under `--plain --edges spline` the output still has no `splines` attribute.
**Why it happens:** PLAIN-02 zeroes `View.Edges` unconditionally in the builders; provenance is lost.
**How to avoid:** Dedicated override carrier applied after the zeroing (Pattern 3); builder-level unit pin + matrix cell.
**Warning signs:** Any plan whose only override mechanism is `v.Edges = ...` in `processView`/`processExpandedView`.

### Pitfall 2: Forgetting the `--expanded` copy
**What goes wrong:** `--edges` works on C1/drill-downs but not `{base}.expanded.dot`.
**Why it happens:** `processExpandedView` is a separate code path (root.go:376-421) with its own threading block and its own builder (`BuildExpandedGraph`).
**How to avoid:** The GEDGE-07 matrix explicitly includes the `expanded` generation; both builder functions get the override.
**Warning signs:** Plan touching only `BuildGraph` or only one threading block.

### Pitfall 3: Validation after parse instead of before
**What goes wrong:** `--edges diagonal` parses the model, then fails — or worse, renders with silently-ignored style (converter ignores unknown values).
**Why it happens:** Converter has no default branch; nothing else validates the string.
**How to avoid:** Validate in `runRoot` beside the `format` check (root.go:138-141), before Stage 1 parse. Error message must name the value AND the full enum (D-02: `invalid edges: must be straight, spline, square, or ortho: "diagonal"` shape).
**Warning signs:** Any validation placed after `parseInput` or inside `internal/render`.

### Pitfall 4: Golden churn
**What goes wrong:** Existing `plain.dot`/`nolabels*.dot` goldens fail.
**Why it happens:** Override applied unconditionally, or the flag default leaks into the plain-zeroing path.
**How to avoid:** Default-empty carrier + post-zeroing application means flag-off writes the exact same `EdgeStyle` as today; run `go test ./...` — all goldens must pass with zero re-baselining (D-04/GEDGE-08).
**Warning signs:** Any plan step that regenerates `testdata/*.dot` goldens.

### Pitfall 5: Docs drift across the 3 SKILL.md copies
**What goes wrong:** Byte-inequality between `skill/SKILL.md`, `plugins/c4drill/skills/c4drill-toml/SKILL.md`, and `plugins/c4drill/opencode/skills/c4drill-toml/SKILL.md`.
**Why it happens:** Editing one copy only (37-06 sync precedent exists precisely because this happened).
**How to avoid:** One edit applied to all three + `cmp`/diff assertion in the docs task; also update the README `--plain` section's "exact union" statement (README.adoc:1311-1317 explicitly says `properties.edges` "is tied to `--plain` only") to state the D-05 delta.
**Warning signs:** Docs task listing fewer than 3 SKILL.md paths, or no README `--plain` edit.

### Pitfall 6: Test parallelism violations
**What goes wrong:** CI flake/failure from the WASM graphviz engine.
**Why it happens:** New tests call `t.Parallel()` while touching `render.*` or the full CLI pipeline.
**How to avoid:** All new matrix/pin tests follow root_test.go convention: no `t.Parallel()` + `//nolint:paralleltest // go-graphviz WASM engine has concurrency issues`.
**Warning signs:** New test functions without the nolint annotation in `cmd/c4drill/root_test.go`.

## Code Examples

Verified current-state excerpts (all confirmed in working tree 2026-08-31):

### Where View.Edges is resolved (UNCHANGED by this phase)
```go
// internal/view/scope.go:504 (C2; same shape at :599 for C3, :25 expanded, :133 C1)
v := &view.View{
	// ...
	Edges: cmp.Or(systemUnit.Edges, m.Properties.Edges),
	// ...
}
```

### Where the style becomes graphviz output (UNCHANGED by this phase)
```go
// internal/render/converter.go:264-271
// Edge routing style
switch g.EdgeStyle {
case "spline":
	cg.SetSplines("true")
case "straight":
	cg.SetSplines("false")
case "ortho", "square":
	// "square" is the documented alias for ortho routing (GEDGE-02)
	cg.SetSplines("ortho")
}
```
Empty `EdgeStyle` → no `splines` attribute emitted → graphviz default routing. This is exactly what `--plain` relies on today.

### The fixture that pins everything
```toml
# cmd/c4drill/testdata/plain.toml:8
edges = "straight" # author edge routing — ignored under --plain
```
With `--edges spline`: raw dot must contain `splines=true` (beats author value). With `--plain --edges spline`: must STILL contain `splines=true` (GEDGE-06 pin). Flag-off: `splines=false` (model value) — and the committed `plain.dot` golden is unchanged.

### Matrix cell pattern (to extend for GEDGE-07)
```go
// cmd/c4drill/root_test.go:2095 (shape)
func runKeySwitchMatrix(t *testing.T, sw keySwitchCase) {
	t.Helper()
	for _, fixture := range []string{"plain", "multilevel", "styles"} {
		t.Run(fixture, func(t *testing.T) {
			for _, gen := range []string{"C1", "drilldown", "expanded"} {
				// render cell via generateFixtureOutput(...extraArgs...)
				// read RAW dot, assert markers (here: splines=<value>)
```

## Runtime State Inventory

> Refactor-adjacent (new flag on existing pipeline), included for completeness.

| Category | Items Found | Action Required |
|----------|-------------|------------------|
| Stored data | None — the tool reads model files, stores nothing; `--edges` is not persisted anywhere | none |
| Live service config | None — pure CLI, no services | none |
| OS-registered state | None | none |
| Secrets/env vars | None touched. (`C4DRILL_LABEL_RATIO` env precedent exists for `--label-ratio`; phase 39 needs NO env var — flag only per GEDGE-03) | none |
| Build artifacts | None — regular `go build`; release binaries rebuilt by the existing release workflow (38-06 precedent) | none |

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | No per-unit `edges` value exists in any `cmd/c4drill/testdata/*.toml` fixture today (only global `properties.edges` in plain.toml:8, peer_walkup.toml:31) | Validation Architecture | Low — if the D-03 pin (flag beats per-unit override) needs a per-unit fixture, the plan adds a tiny fixture or extends an existing one; flagged for the planner |
| A2 | Product release for v1.16 is tagged v1.23.0 following the linear patch history (v1.22.0 = phase 38) | Validation Architecture / notes | Low — release number choice is planner-executor facing; convention is unambiguous from ROADMAP milestone table |

All other claims carry `[VERIFIED: file:line]` citations verified in this session.

## Open Questions (RESOLVED)

1. **Exact override mechanism (A/B/C above).**
   - What we know: D-03/D-05 fix semantics and precedence; CONTEXT.md explicitly leaves the mechanism to the planner.
   - RESOLVED (plan 39-01): Option A — dedicated `View.EdgesOverride string` carrier (default empty = flag absent), applied AFTER the PLAIN-02 zeroing in BOTH BuildGraph and BuildExpandedGraph. Gives structural D-04 flag-off invariance and builder-tier testability for the D-05 pin.
2. **Per-unit edges test fixture (D-03 pin).**
   - What we know: no existing cmd-level fixture carries a per-unit `edges` value (A1); `multilevel.expanded.dot` is a compared golden (internal/graph/builder_test.go:1233,2690) so extending multilevel.toml would churn goldens and violate GEDGE-08.
   - RESOLVED (plan 39-02): new dedicated golden-safe fixture `cmd/c4drill/testdata/edges_override.toml` with global `edges = "spline"` plus one per-unit `edges = "ortho"` subunit — zero golden impact.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go toolchain | build + all tests | ✓ | go1.26.5 (go.mod: 1.26.1) | — |
| golangci-lint | lint gate | ✓ | v2.12.2 | — |
| testify / cobra / go-graphviz | tests / CLI / render | ✓ (go.mod, module cache) | v1.12.1 / v1.10.2 / v0.10.0 | — |
| graphviz native `dot` binary | NOT required | — | — | go-graphviz embeds a WASM engine; RAW dot output needs no native binary |

**Missing dependencies with no fallback:** none.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go standard `testing` + testify v1.12.1 |
| Config file | none (go.mod; mise tasks in `.mise.toml`) |
| Quick run command | `go test ./cmd/c4drill/ -run 'TestEdges' -v` (phase-scoped) |
| Full suite command | `go test ./...` (CI: `go test -v -race -cover ./...`; lint: `golangci-lint run ./...`) |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| GEDGE-03 | `--edges <style>` accepted for all four values; renders without model edit | e2e (CLI) | `go test ./cmd/c4drill/ -run TestEdges -v` | ❌ Wave 0/1 (new) |
| GEDGE-04 | Invalid value → loud error naming value + enum, no output | unit (CLI) | `go test ./cmd/c4drill/ -run TestEdges -v` | ❌ new (extends `TestFlagValidation` family, root_test.go) |
| GEDGE-05 | Override beats global AND per-unit `edges` on C1 / drill-down / expanded | e2e matrix | `go test ./cmd/c4drill/ -run TestEdges -v` | ❌ new matrix case |
| GEDGE-06 | `--plain --edges spline` → `splines=true` in raw dot | e2e pin | `go test ./cmd/c4drill/ -run TestEdges -v` | ❌ new dedicated pin |
| GEDGE-07 | Full `--edges` × generation × `--plain` matrix asserting `splines=` in RAW dot | e2e matrix | `go test ./cmd/c4drill/ -run TestEdges -v` | ❌ new (KEY-03 template) |
| GEDGE-08 | Flag-off: all existing canonicalDOT goldens pass untouched | regression suite | `go test ./...` | ✅ existing goldens (must NOT be re-baselined) |

### Sampling Rate
- **Per task commit:** `go test ./...` (fast enough; full suite is the repo's own convention per task)
- **Per wave merge:** `go test -v -race -cover ./...` + `golangci-lint run ./...`
- **Phase gate:** Full suite green (all goldens untouched) + lint clean before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `cmd/c4drill/root_test.go` — new `TestEdges*` family (flag accept/reject, plain pin, matrix) — no new file needed, follows KEY-03 placement
- [ ] `internal/graph/builder_test.go` — builder-level plain×override unit pin (if Option A mechanism chosen)
- [ ] Possibly one fixture line (per-unit `edges`) in `cmd/c4drill/testdata/multilevel.toml` — see Open Question 2
- Framework install: none needed

*(No new test infrastructure required — the repo's test conventions fully cover this phase.)*

## Security Domain

### Applicable ASVS Categories (L1)

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | no | N/A — offline CLI, no auth surface |
| V3 Session Management | no | N/A |
| V4 Access Control | no | N/A — no privileged operations; writes only to user-specified output dir (unchanged) |
| V5 Input Validation | **yes** | Enum allow-list validation of `--edges` in `runRoot` before any file I/O; fail-closed with sentinel error (D-02). Same control as existing `--format` validation |
| V6 Cryptography | no | N/A |

### Known Threat Patterns for Go CLI (cobra) input handling

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Unvalidated enum flag reaching renderer | Tampering | Allow-list check before pipeline start; unknown value = hard error, never silent pass-through (the converter's silent-ignore switch is internal-only and must not become the validation layer) |
| Error message injection via flag value | Information Disclosure | `%q` quoting of the offending value in the error wrap (repo precedent `fmt.Errorf("%w: %q", ...)`) — no raw echo of unquoted input |
| Output-path abuse | — | Unchanged by this phase: `--edges` influences styling only, never paths; existing `--output` handling untouched |

Threat-model verdict for the plan: a single LOW-severity input-validation threat (V5) closed by the GEDGE-04 allow-list; no new attack surface (no network, no exec, no persistence).

## Sources

### Primary (HIGH confidence)
- Working tree, verified 2026-08-31: `cmd/c4drill/root.go` (:27-33 sentinels, :50-62 flag vars, :96-115 registrations, :128-155 runRoot validation, :313-372 processView, :376-421 processExpandedView)
- `internal/view/view.go:20-91` (View struct: `Edges` :33, switch fields), `internal/view/scope.go` (:25, :133, :504, :599 resolution)
- `internal/graph/builder.go` (:38-49 BuildGraph PLAIN-02, :411-433 BuildExpandedGraph PLAIN-02), `internal/graph/graph.go:66` (EdgeStyle)
- `internal/render/converter.go:264-271` (SetSplines mapping), `cmd/c4drill/root_test.go:1944-2233` (KEY-03 matrix + helpers), `cmd/c4drill/testdata/plain.toml:8`
- `.planning/codebase/CONVENTIONS.md`, `.planning/codebase/TESTING.md`, `.planning/codebase/ARCHITECTURE.md`
- `.planning/REQUIREMENTS.md` (GEDGE-03..08, exclusions), `.planning/ROADMAP.md` §Phase 39, `.planning/PROJECT.md` KEY-02, folded todo `todos/pending/2026-08-30-add-cli-flag-to-override-edge-routing-style.md`
- `README.adoc` (:1237-1248 CLI Reference flags, :1256-1317 `--plain` narrative incl. the "properties.edges tied to --plain only" statement)

### Secondary (MEDIUM confidence)
- `.planning/milestones/v1.15-ROADMAP.md` (38-04 matrix plan shape, 38-06 release v1.22.0 precedent), `v1.13-ROADMAP.md` (GEDGE-01/02 heritage)

### Tertiary (LOW confidence)
- None — no external/web sources were needed; everything is verified in-repo.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — zero new dependencies; all versions read from go.mod and installed toolchain
- Architecture: HIGH — every file:line in the data flow opened and read this session; CONTEXT.md's canonical refs all confirmed accurate
- Pitfalls: HIGH — derived from verified code (PLAIN-02 dual sites, silent converter switch, WASM parallelism rule, golden inventory)

**Research date:** 2026-08-31
**Valid until:** 2026-09-30 (stable internal codebase; re-verify only if master moves substantially)

## RESEARCH COMPLETE

**Phase:** 39 - Edge Style Override (`--edges` CLI flag)
**Confidence:** HIGH

### Key Findings
- Zero render-side changes needed: `configureGraphSettings` already maps all four enum values; the phase is wiring (flag → validate → thread → precedence) + tests + docs.
- D-05 is the only real logic: `--plain` zeroes `View.Edges` in BOTH builders, so a bare `v.Edges = flag` fails GEDGE-06; a dedicated default-empty override applied after the zeroing gives explicit-CLI > plain > model with structural flag-off invariance.
- `cmd/c4drill/testdata/plain.toml:8` already carries `edges = "straight"` — the ready-made fixture for the plain pin and the D-03/D-05 matrix cells; no cmd-level fixture currently has a per-unit `edges` value (planner should add one line to multilevel.toml for the D-03 pin).
- KEY-03 matrix (root_test.go:1944-2233) + `generateFixtureOutput` helper are the direct GEDGE-07 template; all render-touching tests need the paralleltest nolint.
- Docs delta touches README.adoc CLI Reference + the `--plain` "exact union" statement (:1311-1317) + 3 byte-identical SKILL.md copies.

### File Created
`.planning/phases/39-edge-style-override-edges-cli-flag/39-RESEARCH.md`

### Confidence Assessment
| Area | Level | Reason |
|------|-------|--------|
| Standard Stack | HIGH | No new packages; go.mod + toolchain verified |
| Architecture | HIGH | Full data flow read at file:line level this session |
| Pitfalls | HIGH | Each pitfall maps to verified code behavior |

### Open Questions
- Override mechanism A/B/C (planner decides; A recommended) — see Open Questions.
- Per-unit `edges` fixture placement for the D-03 pin (extend multilevel.toml recommended).

### Ready for Planning
Research complete. Planner can now create PLAN.md files.
