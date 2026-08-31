# Phase 39: Edge Style Override (`--edges` CLI flag) - Pattern Map

**Mapped:** 2026-08-31
**Files analyzed:** 9 (all modifications — no new source files)
**Analogs found:** 9 / 9 (every change has an exact in-codebase analog; this phase replicates the v1.21 granular-switch pattern end to end)

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| `cmd/c4drill/root.go` — flag var + registration | config (CLI surface) | request-response | same file: `plain`/`format` vars (:50-62) + registrations (:96-115) | exact |
| `cmd/c4drill/root.go` — `runRoot` enum validation | controller | request-response | same file: `errInvalidFormat` sentinel (:28) + check (:138-141) | exact |
| `cmd/c4drill/root.go` — threading in `processView`/`processExpandedView` | controller | transform | same file: PLAIN-01 blocks (:330-338, :386-391) | exact |
| `internal/view/view.go` — override carrier field on `View` | model | transform | same file: `NoLabels bool` field (:56-63, newest switch field, quick 260831-01u) | exact |
| `internal/graph/builder.go` — override after PLAIN-02 zeroing (both builders) | service (transform) | transform | same file: the PLAIN-02 blocks being extended (:38-49, :411-433) | exact |
| `cmd/c4drill/root_test.go` — `TestEdges*` family | test (e2e CLI) | request-response | `TestFlagValidation` (:114), `TestKeyComposition` matrix (:1998), `generateFixtureOutput` (:1951), `TestNoLabelsC1Golden` (:1968) | exact |
| `internal/graph/builder_test.go` — plain×override unit pin | test (unit) | transform | `TestBuildGraph_PlainSkipsUnitOverrides` (:3320) + EdgeStyle pin (:53-54) | exact |
| `cmd/c4drill/testdata/multilevel.toml` — per-unit `edges` line | config fixture | file-i/o | `cmd/c4drill/testdata/plain.toml:8` (`edges = "straight"` + explanatory comment) | role-match |
| `README.adoc` + `skill/SKILL.md` + 2× `plugins/c4drill/*/skills/c4drill-toml/SKILL.md` | docs | — | `--no-labels` docs (LBL-02, the most recent flag addition): README.adoc:1243,1300-1301; skill/SKILL.md:1006-1007 | role-match |

## Pattern Assignments

### `cmd/c4drill/root.go` — flag var + registration (config, request-response)

**Analog:** same file, `plain` var and `--plain` registration

**Flag var block** (lines 49-62):
```go
//nolint:gochecknoglobals // Cobra flags require package-level variables for PersistentFlags registration
var (
	format     string
	outputDir  string
	expanded   bool
	plain      bool
	noColors   bool
	noStyles   bool
	noLength   bool
	noRank     bool
	noLabels   bool
	labelRatio float64
	version    = "dev"
)
```

**Registration pattern** (lines 100-103):
```go
cmd.PersistentFlags().BoolVar(&expanded, "expanded", false,
	"Generate all-expanded diagram showing all units")
cmd.PersistentFlags().BoolVar(&plain, "plain", false,
	"Ignore author-custom formatting: default unit/edge styling, spacing and ranking, plain-text labels")
```
`--edges` adds one `StringVar` (no shorthand — none of the switches has one) to this block with a help string in the same imperative style, e.g. "Override edge routing style for every diagram (straight|spline|square|ortho)".

### `cmd/c4drill/root.go` — `runRoot` validation (controller, request-response)

**Analog:** same file, format sentinel + early check

**Sentinel declaration** (lines 26-33):
```go
// Static errors for better error handling.
var (
	errInvalidFormat    = errors.New("invalid format: must be dot, svg, or html")
	errValidationFailed = errors.New("validation failed")
	...
)
```

**Early validation before file I/O** (lines 138-141):
```go
// Validate flags early (before file I/O)
if format != formatDot && format != formatSVG && format != formatHTML {
	return fmt.Errorf("%w: %q", errInvalidFormat, format)
}
```
`--edges` gets a sibling sentinel whose message names the full enum (D-02) and a sibling check directly below this one. Error text shape: `invalid edges: must be straight, spline, square, or ortho: "diagonal"`.

### `cmd/c4drill/root.go` — threading (controller, transform)

**Analog:** same file, PLAIN-01 blocks

**processView threading** (lines 330-338):
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

**processExpandedView threading** (lines 383-391):
```go
// PLAIN-01: --plain x --expanded — the expanded view gets the flag too
// (BuildExpandedGraph copies View.Plain into Graph.Opts.Plain).
// KEY-01: the granular switches thread the same way. LBL-02: --no-labels.
v.Plain = plain
...
```
The override joins BOTH blocks (comment updated to name it), exactly per D-03/GEDGE-05.

### `internal/view/view.go` — override carrier field (model, transform)

**Analog:** same file, `NoLabels` field — the newest suppression-switch field

**Field pattern** (lines 56-63):
```go
// NoLabels suppresses EDGE label text only (--no-labels, quick
// 260831-01u BUG-2 — supersedes the phase-38 LBL-01 all-labels pin):
// edges carry no label content while node, cluster (wrapper, boundary,
// expanded, nested) and legend labels all remain visible. ...
NoLabels bool
```
A new string field (e.g. `EdgesOverride string`) follows this documentation style: name the flag, the phase/pin IDs (D-03/D-05/GEDGE-05/06), and the precedence semantics — default empty = flag absent = zero behavior change (D-04).

### `internal/graph/builder.go` — override after PLAIN-02 (service, transform)

**Analog:** same file, the PLAIN-02 blocks being extended

**BuildGraph site** (lines 38-49):
```go
// PLAIN-02: properties.edges is ignored under plain — an empty EdgeStyle
// makes the renderer's configureGraphSettings fall through to the default
// routing.
edgeStyle := v.Edges
if v.Plain {
	edgeStyle = ""
}

g := &Graph{
	Title:     v.Title,
	Direction: "TB",
	EdgeStyle: edgeStyle,
	...
```

**BuildExpandedGraph site** (lines 423-436, same logic):
```go
// PLAIN-02: properties.edges ignored under plain (same as BuildGraph).
edgeStyle := v.Edges
if v.Plain {
	edgeStyle = ""
}
```
Both duplicated blocks get the identical D-05 addition AFTER the zeroing:
```go
// D-05/GEDGE-06: explicit --edges is USER intent — survives plain zeroing.
if v.EdgesOverride != "" {
	edgeStyle = v.EdgesOverride
}
```
(Mechanism per RESEARCH.md Option A; planner may substitute B but must keep: post-zeroing application, both sites in sync, default no-op.)

### `cmd/c4drill/root_test.go` — TestEdges family (test, e2e)

**Analog 1:** `TestFlagValidation` (lines 114-163) — table-driven flag accept/reject with `NewRootCmd()` + `SetOut/SetErr` + `bytes.Buffer`:
```go
for _, tt := range tests {
	t.Run(tt.name, func(t *testing.T) { //nolint:paralleltest // go-graphviz WASM engine has concurrency issues
		cmd := NewRootCmd()
		buf := &bytes.Buffer{}
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		// Set flags, Execute, assert error/empty-output
```
GEDGE-04 test copies this shape: valid values `straight|spline|square|ortho` pass, `diagonal`/`""`/`SPLINE` fail with the sentinel message; assert NO output files written.

**Analog 2:** `generateFixtureOutput` helper (lines 1951-1969) — the e2e harness:
```go
dir := t.TempDir()
args := append([]string{
	filepath.Join("testdata", fixture),
	"--output", dir,
	"--format", format,
}, extraArgs...)
cmd := NewRootCmd()
cmd.SetArgs(args)
require.NoError(t, cmd.Execute(), "%s must render cleanly with %v", fixture, extraArgs)
```
Matrix cells call it with `--edges <style>` / `--plain --edges <style>` extras and read the RAW dot from the returned dir (GEDGE-07: assert `splines="true"|"false"|ortho` substring per combination).

**Analog 3:** `TestKeyComposition` (lines 1998-2035) + `keySwitches()` (2057-2093) — matrix organization: per-flag subtests, per-fixture sub-subtests, per-generation (`C1`/`drilldown`/`expanded`) cells. The `--edges` case needs the extension shown in RESEARCH.md (raw-dot `splines=` assertion instead of absent/present markers).

**Analog 4:** `TestNoLabelsC1Golden` (lines 1968-1981) — golden-invariance idiom (GEDGE-08 is satisfied by NOT touching goldens; existing suite already pins it):
```go
require.Equal(t, canonical.Canonical(t, expected), canonical.Canonical(t, got),
	"--no-labels C1 output must match the committed nolabels.dot golden (canonical, DI-1)")
```
Every render-touching test in this file carries `//nolint:paralleltest // go-graphviz WASM engine has concurrency issues` — mandatory on all new tests.

### `internal/graph/builder_test.go` — plain×override pin (test, unit)

**Analog 1:** EdgeStyle pin (lines 53-54):
```go
// Test 3: BuildGraph sets EdgeStyle from View.Edges
assert.Equal(t, "spline", g.EdgeStyle)
```

**Analog 2:** `TestBuildGraph_PlainSkipsUnitOverrides` (lines 3312-3341) — the plain-guard family:
```go
// Plain-guard tests (PLAIN-01/PLAIN-02): with View.Plain set, author-custom
// ... survive. Plain defaults to false, so the no-flag path is pinned by
...
v.Plain = true
```
New pin: `v.Edges = "straight"` + `v.Plain = true` + `v.EdgesOverride = "spline"` → `g.EdgeStyle == "spline"` (D-05 at builder tier); plus flag-off case: `EdgesOverride == ""` → EdgeStyle exactly as today. These builder tests touch no rendering → `t.Parallel()` allowed per TESTING.md.

### `cmd/c4drill/testdata/multilevel.toml` — per-unit edges line (fixture)

**Analog:** `cmd/c4drill/testdata/plain.toml` line 8:
```toml
edges = "straight" # author edge routing — ignored under --plain
```
Add the same style of annotated `edges = "..."` on ONE subunit (per-unit level) so the D-03 pin (flag beats per-unit override) is observable in a matrix cell without a new fixture file.

### Docs — README.adoc + 3 SKILL.md copies (role-match)

**Analog:** the `--no-labels` documentation (most recent flag, LBL-02)

**README.adoc** — two spots: CLI Reference flags block (line 1243 shows the `--help` transcript entry style) and the narrative bullet (lines 1300-1301). The `--plain` section's exact-union statement (lines 1311-1317, currently: "`properties.edges` (spline routing) is tied to `--plain` only") must gain the D-05 delta sentence.

**SKILL.md** — `skill/SKILL.md` lines 1006-1007 show the flag-example style:
```
c4drill architecture.toml --no-labels           # edge labels only: nodes/clusters/legend keep labels
c4drill architecture.toml --plain --no-labels   # switches compose
```
The three copies (`skill/SKILL.md`, `plugins/c4drill/skills/c4drill-toml/SKILL.md`, `plugins/c4drill/opencode/skills/c4drill-toml/SKILL.md`) must stay byte-identical (37-06 convention) — one edit applied to all three, verified with `cmp`.

## Shared Patterns

### Sentinel errors + `%q` wrap
**Source:** `cmd/c4drill/root.go:26-33,138-141`
**Apply to:** the new `--edges` validation
```go
errInvalidEdges = errors.New("invalid edges: must be straight, spline, square, or ortho")
...
return fmt.Errorf("%w: %q", errInvalidEdges, edges)
```

### nolint with mandatory explanation
**Source:** CONVENTIONS.md; 91 instances e.g. `cmd/c4drill/root_test.go` throughout
**Apply to:** every new render-touching test + the flag var block addition (existing block-level nolint already covers the var)
```go
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
```

### Testify split: require (fatal) / assert (value) + message args
**Source:** `.planning/codebase/TESTING.md`; root_test.go:1968-1981
**Apply to:** all new tests
```go
require.NoError(t, cmd.Execute(), "...")
assert.Equal(t, "spline", g.EdgeStyle, "Graph.EdgeStyle")
```

### Statement-spacing gofmt style
**Source:** CONVENTIONS.md Code Style; visible in every excerpt above
**Apply to:** all edits — blank line after every statement block, `0o` octal prefix (n/a here), grouped imports (no new imports needed in any touched file except possibly none at all)

### Docs 3-copy byte-identical sync
**Source:** 37-06 convention (v1.14); D-07
**Apply to:** SKILL.md edits — one content change, three files, `cmp` proof

## No Analog Found

None — every touched file has an exact or role-match analog in the current tree. This phase is a faithful replication of the v1.21 granular-switch pattern (KEY-01/PLAIN-01/LBL-02) applied to a string-valued flag with a precedence twist (D-05) documented in RESEARCH.md.

## Metadata

**Analog search scope:** `cmd/c4drill/`, `internal/view/`, `internal/graph/`, `internal/render/`, `testdata/`, `README.adoc`, `skill/`, `plugins/c4drill/`
**Files scanned:** 12 (9 targets + 3 supporting)
**Pattern extraction date:** 2026-08-31

## PATTERN MAPPING COMPLETE

**Phase:** 39 - Edge Style Override (`--edges` CLI flag)
**Files classified:** 9
**Analogs found:** 9 / 9

### Key Patterns Identified
- Everything copies the granular-switch family: flag var block → registration → early sentinel validation → PLAIN-01 threading at exactly two sites → builder consumption
- The one new twist (D-05) extends the duplicated PLAIN-02 blocks in BOTH builders, applied after the plain zeroing
- Tests: TestFlagValidation shape for GEDGE-04, generateFixtureOutput/TestKeyComposition harness for the matrix, plain-guard family for builder pins; paralleltest nolint mandatory on render-touching tests
- Docs follow the --no-labels precedent with the 3-copy byte-identical sync

### File Created
`.planning/phases/39-edge-style-override-edges-cli-flag/39-PATTERNS.md`

### Ready for Planning
Pattern mapping complete. Planner can now reference analog patterns in PLAN.md files.
