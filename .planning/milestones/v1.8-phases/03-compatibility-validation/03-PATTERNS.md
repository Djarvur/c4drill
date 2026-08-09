# Phase 3: Compatibility & Validation - Pattern Map

**Mapped:** 2026-08-06
**Files analyzed:** 4 (2 CREATE testdata assets, 2 MODIFY test files)
**Analogs found:** 4 / 4 (1 exact, 2 role-match, 1 partial + verified structural source)

## Verification Notes (corrections to RESEARCH.md)

These were verified against the actual files and MUST be respected by the planner:

1. **The saira fixture is 500 lines, not 88** (`cyp-auth-infra/saira-20260320.c2.full.toml`). RESEARCH.md's "88 lines" is wrong. Structure is intact: 5 top-level units, 4-level nesting, cross-level `length` links.
2. **No test consumes `cmd/c4drill/testdata/expanded.dot`** — verified via `grep -rn "expanded.dot" cmd/c4drill/ internal/`: only hits are inside committed generated DOT files under `cmd/c4drill/testdata/expanded/` (href references in labels), no `os.ReadFile` in any test. It is also NOT an expanded-mode output (it is a C1 `--format dot` output with `[+]` collapsed labels and no per-edge `penwidth=2`/`minlen`). The committed golden baseline for COMPAT-02 must be a **new file**.
3. **The private fixture is untracked** — `git ls-files cyp-auth-infra/` returns 0; `git check-ignore` confirms `.gitignore:71-72` (`cyp-auth-infra/`). Repointing the two builder tests is required for fresh-checkout CI (D-01).
4. **`isExpandedInC1` lives at `internal/view/scope.go:182`, not :144** (CONTEXT/RESEARCH line drift). `GenerateExpandedView` at `scope.go:13`.
5. **Test-body hardcoded names**: `TestBuildExpandedGraphRealToml` contains hardcoded paths `linuxSystem.storages.keycloakStorage.client` and `keycloak` (builder_test.go:811, :814, :825). Repointing requires renaming these to the sanitized fixture's names, not just the path constant.

---

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| `cmd/c4drill/testdata/multilevel.toml` (CREATE) | fixture (static test data) | static config | `cyp-auth-infra/saira-20260320.c2.full.toml` (structural source, 500 lines) + `cmd/c4drill/testdata/expanded.toml` (committed placement analog) | exact (structure) / role-match (placement) |
| `cmd/c4drill/testdata/multilevel.expanded.dot` (CREATE) | golden baseline (static test data) | static | `cyp-auth-infra/saira-20260320.c2.full.expanded.dot` (private generated baseline, 1041 lines) | exact (generation source); NO consumer analog exists (nothing consumes a committed DOT today) |
| `internal/graph/builder_test.go` (MODIFY) | test (unit/integration) | transform: parse → validate → view → graph → render | itself — repoint `TestBuildExpandedGraphRealToml` (:795-840) and `TestBuildExpandedGraphBaselineDOT` (:1205-1232); D-04 modeled on `TestBuildEdgesExpandedExemption` (:1912-1949) | exact (self-analog) |
| `cmd/c4drill/root_test.go` (MODIFY) | test (CLI integration) | request-response (Cobra pipeline) | `TestFormatFlag_Dot` (:583-599) + `TestExpandedUnits` (:554-578) | exact |

---

## Pattern Assignments

### CREATE `cmd/c4drill/testdata/multilevel.toml` (sanitized public fixture)

**Structural source:** `cyp-auth-infra/saira-20260320.c2.full.toml` (500 lines) — copy structure, rename identifiers.

**Required skeleton (what the two repointed tests depend on):**

| Source element (file:line) | Sanitized equivalent (suggested) | Why it must be preserved |
|---|---|---|
| `[properties] name/description` (:1-3) | generic title | validator + DOT graph label |
| 5 top-level units `webUser`/`sshUser`/`adminUser` (:7-30) | `actorA`/`actorB`/`actorC` (personExternal) | 5-node C1 shape (ROADMAP success criterion) |
| `[keycloak]` (:31-35) | `[externalSys]` (systemExternal) | target of the `length=3` cross-level link |
| `[linuxSystem]` (:39-42) | `[mainSystem]` (system) | the deep-nesting root |
| 4-level nesting `linuxSystem.sshAuth.systemd.logind` (:70-74) | `mainSystem.sshAuth.systemd.logind` | exercises deep-link resolution (Pitfall 2) |
| `linuxSystem.storages.keycloakStorage.client` with `link = [{peer = "keycloak", length=3}]` (:392-399) | `mainSystem.storages.externalStorage.client` → `{peer = "externalSys", length=3}` | THE edge `TestBuildExpandedGraphRealToml` asserts (builder_test.go:822-832) — cross-level link with `length` |
| `length=2` links (:175, :259-260, :270) | keep 2-3 of them | minlen variety in baseline DOT |

**TOML syntax pattern to mirror** (inline array-of-tables, per saira fixture :10-13):
```toml
[webUser]
name = "Web User"
type = "personExternal"
link = [
  {peer = "linuxSystem.localIDP.grpcAPIs.authAPI"},
  {peer = "linuxSystem.localIDP.grpcAPIs.sessionAPI"},
]
```
Note: saira uses inline `link = [{peer = ...}]` arrays (not `[[unit.link]]`). The validator runs in both tests (`require.Empty(t, validator.Validate(m))` at builder_test.go:808, :1215), so the fixture must pass validation — no orphan peers, all link targets resolvable. The fixture must contain NO `expanded` hints (neither `[properties] expanded` nor per-unit), matching saira — that is exactly the COMPAT-01/COMPAT-02 baseline state.

**Committed-placement analog:** `cmd/c4drill/testdata/expanded.toml` (:1-51) and `valid.toml` (:1-26) — committed TOML fixtures live in `cmd/c4drill/testdata/` and are read via relative paths (`filepath.Join("testdata", ...)` from `cmd/c4drill` cwd; `../../cmd/c4drill/testdata/<name>.toml` from `internal/graph` — replaces the current `../../cyp-auth-infra/...` at builder_test.go:803, :1211).

---

### CREATE `cmd/c4drill/testdata/multilevel.expanded.dot` (committed golden baseline)

**Generation source:** `cyp-auth-infra/saira-20260320.c2.full.expanded.dot` (1041 lines, private). It is the output of the exact pipeline the baseline test runs:
`parser.ParseFile` → `validator.Validate` → `view.GenerateExpandedView` → `graph.BuildExpandedGraph` → `render.RenderDOT` (builder_test.go:1211-1231 pattern).

**What the DOT looks like** (from the private baseline): XDOT format `digraph "" { ... }` with graph-level `rankdir=TB`, per-edge blocks like:
```
"linuxSystem.storages.keycloakStorage.client" -> keycloak	[key="linuxSystem.storages.keycloakStorage.client_to_keycloak",
    color="#78A8D8",
    fontsize=10,
    label="",
    minlen=3,
    penwidth=2,
    pos="e,420.07,58.792 ...",
    style=solid];
```
(private baseline :1032-1040). Note the `key=` attribute is the dedup key (Phase 1 D-02) and `minlen`/`penwidth=2` are the D-02 contract attributes. The baseline contains layout geometry (`pos`, `bb`, `rects`, `lp`) — deterministic for the go-graphviz version pinned in go.mod; the graph package's own tests already assert on this output via `assert.Contains` (builder_test.go:1230-1231).

**Golden comparison pattern** (RESEARCH "Don't Hand-Roll" — no golden library): `os.ReadFile` + `require.Equal`:
```go
expected, err := os.ReadFile("../../cmd/c4drill/testdata/multilevel.expanded.dot")
require.NoError(t, err)
dotData, err := render.RenderDOT(g)
require.NoError(t, err)
require.Equal(t, string(expected), string(dotData))
```
**Filename convention:** `{basename}.expanded.{format}` per writer.go:66 (CONTEXT code_context) — `multilevel.expanded.dot` — so it can be regenerated by hand with `c4drill testdata/multilevel.toml --format dot --expanded`. D-02 contract = node/edge sets + attributes + cluster structure; exact byte match is achievable via the pinned pipeline (regenerate from the SAME pipeline; see Pitfall 1).

---

### MODIFY `internal/graph/builder_test.go` (repoint 2 tests + add D-04 test)

**File conventions:** `package graph_test` (external, black-box), builder_test.go:1-16 imports:
```go
import (
	"fmt"
	"strings"
	"testing"

	"github.com/Djarvur/c4drill/internal/graph"
	"github.com/Djarvur/c4drill/internal/model"
	"github.com/Djarvur/c4drill/internal/parser"
	"github.com/Djarvur/c4drill/internal/render"
	"github.com/Djarvur/c4drill/internal/validator"
	"github.com/Djarvur/c4drill/internal/view"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)
```
Group order: stdlib, then module packages, then third-party (testify) — CONVENTIONS.md import rules.

**Repoint target 1 — `TestBuildExpandedGraphRealToml` (:795-840).** Current body to migrate (only the fixture path and the hardcoded unit paths change):
```go
m, err := parser.ParseFile("../../cyp-auth-infra/saira-20260320.c2.full.toml")  // :803 → ../../cmd/c4drill/testdata/multilevel.toml
require.NoError(t, err)

valErrors := validator.Validate(m)                                              // :807
require.Empty(t, valErrors, "model should be valid")

client := m.Units["linuxSystem"].Subunits["storages"].Subunits["keycloakStorage"].Subunits["client"]  // :811 → sanitized names
require.NotNil(t, client, "client unit should exist")
require.Len(t, client.Links, 1, "client should have 1 link")
assert.Equal(t, "keycloak", client.Links[0].Peer)                               // :814 → "externalSys"
expectedLength := client.Links[0].Length
require.Greater(t, expectedLength, 0, "link should have Length > 0")

v := view.GenerateExpandedView(m)                                               // :819
g := graph.BuildExpandedGraph(v)                                                // :820

for _, edge := range g.Edges {                                                  // :823-828
	if edge.Source == "linuxSystem.storages.keycloakStorage.client" && edge.Target == "keycloak" {
		foundEdge = edge
	}
}
require.NotNil(t, foundEdge, "Edge from client to keycloak should exist")       // :831
assert.Equal(t, expectedLength, foundEdge.MinLen, "Edge MinLen should match TOML length attribute")  // :832

dotData, err := render.RenderDOT(g)                                             // :835
require.NoError(t, err)
expectedMinlen := fmt.Sprintf("minlen=%d", expectedLength)
assert.Contains(t, string(dotData), expectedMinlen, "DOT output should contain minlen for the edge")  // :838-839
```
Assertion-relevant struct fields: `graph.Edge{Source, Target, MinLen, PenWidth}` (graph.go:74-92).

**Repoint target 2 — `TestBuildExpandedGraphBaselineDOT` (:1205-1232).** Body to migrate: same ParseFile/Validate/GenerateExpandedView/BuildExpandedGraph preamble (:1211-1218); penwidth-2.0 loop over `g.Edges` (:1221-1224, `assert.InDelta(t, 2.0, edge.PenWidth, 0.001, ...)`); RenderDOT contains checks (`"minlen="`, `"penwidth=2"`) (:1226-1231). Per D-01 this becomes the golden-baseline test: replace the two `assert.Contains` calls with `os.ReadFile("../../cmd/c4drill/testdata/multilevel.expanded.dot")` + `require.Equal` against the full DOT (pattern above). Keep the penwidth loop and the `t.Parallel()` — safe because `render.RenderDOT` serializes via `wasmMutex` (render.go:15-20, :86-88).

**New test — D-04: `--expanded` ignores `properties.expanded`.** Pattern to model on: `TestBuildEdgesExpandedExemption` (:1912-1949) — constructs a view literally, but for D-04 use the fixture pipeline so the deep structure is exercised:
1. Parse `../../cmd/c4drill/testdata/multilevel.toml`, validate (`require.Empty`).
2. **Set the poison**: `m.Properties.Expanded = []string{"mainSystem"}` — field declared at `internal/model/properties.go:19-20` (`Expanded []string \`toml:"expanded"\``); per-unit `Expanded` at `internal/model/unit.go:62-63`.
3. `v := view.GenerateExpandedView(m)` — structural guarantee: `GenerateExpandedView` NEVER consults `m.Properties.Expanded` (scope.go:13-49; `addUnitRecursive` sets `IsExpanded: len(unit.Subunits) > 0` at scope.go:56).
4. Assert ALL units present — nothing collapsed:
   - `require.Len(t, v.Units, <total unit count>)` (count by walking the fixture, or assert no `Entry` has `IsExpanded == false` while `HasSubunits == true`; Entry fields at scope.go:53-59).
   - `g := graph.BuildExpandedGraph(v)` then walk `g.Clusters` recursively (Cluster struct graph.go:105-120: `ID`, `Nodes`, `Clusters`) and assert deep leaves exist as node IDs, e.g. `mainSystem.sshAuth.systemd.logind`, and that no `Node.Label.Name` contains `"[+]"` (collapsed indicator assertion pattern: `TestBuildGraphCollapsedIndicator` builder_test.go:116-143, `assert.Contains(t, g.Nodes[0].Label.Name, "[+]")`).
5. `//nolint:funlen` not needed; the test may exceed nothing — keep `t.Parallel()`.

---

### MODIFY `cmd/c4drill/root_test.go` (COMPAT-01 regression)

**File conventions:** `package main` (internal), imports (root_test.go:3-11): `bytes`, `os`, `path/filepath`, `testing` + testify `assert`/`require`. **No `t.Parallel()` anywhere** — every test carries `//nolint:paralleltest // go-graphviz WASM engine has concurrency issues` (e.g. :15, :67, :165; file-level note :162-163). New tests MUST follow this.

**COMPAT-01 regression skeleton** — model on `TestFormatFlag_Dot` (:583-599, the dot-output fixture test):
```go
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestCompat01_ValidTomlAllCollapsed(t *testing.T) {
	tmpDir := t.TempDir()

	cmd := NewRootCmd()
	cmd.SetArgs([]string{
		filepath.Join("testdata", "valid.toml"),  // no properties.expanded, no per-unit expanded (valid.toml:1-26)
		"--output", tmpDir,
		"--format", "dot",
	})

	err := cmd.Execute()
	require.NoError(t, err)

	// C1 DOT must contain the collapsed app node ("[+]" indicator) ...
	dotData, err := os.ReadFile(filepath.Join(tmpDir, "valid.dot"))
	require.NoError(t, err)
	dot := string(dotData)
	assert.Contains(t, dot, "[+]", "app has subunits and no expansion hint -> collapsed in C1")
	// ... and must NOT contain expanded content (no app.api node).
	assert.NotContains(t, dot, "app.api", "subunits must not appear in C1 when collapsed")
	// No C2 sub-diagram directory must be generated (no expansion anywhere).
	_, err = os.Stat(filepath.Join(tmpDir, "valid", "app.svg"))
	assert.True(t, os.IsNotExist(err), "C2 diagram must not exist for collapsed unit")
}
```
- `[+]` collapsed-indicator appears in C1 dot labels (`cmd/c4drill/testdata/expanded.dot:41`: `label="{Web Application [+]|Frontend}"` — note: that file is a stale C1 dot output, use it only as label-format reference).
- File-existence assertions: `assert.FileExists` (root_test.go:276, :634) and `os.Stat` + `os.IsNotExist` (:658-663, :689-690).
- OR-semantics (D-03) is NOT part of this test (valid.toml has no hints); if desired, a second table entry can reuse the `TestFullPipeline_NestedWithExpanded` inline-TOML pattern (:334-390) with a self-referencing `expanded = ["app"]` on the unit — but that is optional per CONTEXT.

---

## Shared Patterns

### Test-data access paths
- From `cmd/c4drill` (cwd = package dir): `filepath.Join("testdata", "valid.toml")` — root_test.go:475, :542, :559.
- From `internal/graph`: relative from `internal/graph/` — current `../../cyp-auth-infra/...` (builder_test.go:803, :1211) becomes `../../cmd/c4drill/testdata/multilevel.toml` and `../../cmd/c4drill/testdata/multilevel.expanded.dot`.

### Inline TOML fixtures (when not using committed fixtures)
`t.TempDir()` + `os.WriteFile(path, []byte(content), 0o600)` — root_test.go:181-183, :201-213, :256; CONVENTIONS.md:45 (0o prefix), :146-147. Valid for COMPAT-01 table entries / D-04 variants.

### Assertion split (CONVENTIONS.md:142-143)
- `require` for fatal: parse/validate errors (`require.NoError`, `require.Empty`), nil checks (`require.NotNil`).
- `assert` for value checks with a trailing message argument naming the expectation (`assert.Equal(t, expectedLength, foundEdge.MinLen, "Edge MinLen should match TOML length attribute")`).

### Parallelism rules
- `internal/graph` tests: `t.Parallel()` at top of every test INCLUDING those calling `render.RenderDOT` — safe because render.go serializes WASM via `wasmMutex` (render.go:15-20, :86-88). The two repointed tests and the D-04 test keep `t.Parallel()`.
- `cmd/c4drill` tests: NO `t.Parallel()`; `//nolint:paralleltest // go-graphviz WASM engine has concurrency issues` on every test (root_test.go:15, :39, :67, ..., :669; rationale :162-163).

### Lint compliance (CONVENTIONS.md:48-59, `.golangci.yml` all-linters + `tests: true`)
- Every `//nolint:` directive needs a trailing explanation comment.
- Doc comments on new exported helpers; unexported helpers (e.g., a `countUnits` walker) need none but stay camelCase.
- Comments reference the decision IDs: `// D-04: --expanded ignores properties.expanded` style (existing precedent: `// D-05)` at builder_test.go:1907, `// D-02: expanded mode keeps...` :1943).

### Error handling in tests
No custom error paths — tests assert error expectations via `require.NoError(t, err)` and `require.Empty(t, valErrors, "model should be valid")` (builder_test.go:808, :1215). CLI-level error assertions: `assert.Contains(t, err.Error(), "parse", ...)` (root_test.go:175, :194).

---

## No Analog Found

| File | Role | Data Flow | Reason / What to use instead |
|------|------|-----------|------------------------------|
| Committed golden-DOT consumer test (new baseline comparison in `TestBuildExpandedGraphBaselineDOT`) | test | transform | No test today reads a committed DOT (`expanded.dot` verified unconsumed). Use RESEARCH "Don't Hand-Roll": `os.ReadFile` + `require.Equal` on the committed `multilevel.expanded.dot`; regenerate baseline from the SAME pipeline |
| D-04 regression (expanded ignores `properties.expanded`) | test | transform | No existing test sets `properties.expanded` against expanded mode (all existing expanded tests use hint-free or per-unit-hint models). Model structure on `TestBuildEdgesExpandedExemption` (builder_test.go:1912-1949) + fixture pipeline; poison input via `m.Properties.Expanded` (model/properties.go:19-20) |

---

## Metadata

**Analog search scope:** `internal/graph/`, `cmd/c4drill/`, `internal/render/`, `internal/view/`, `internal/model/`, `cmd/c4drill/testdata/`, `cyp-auth-infra/` (private, read-only), `.planning/codebase/CONVENTIONS.md`, `.gitignore`
**Files scanned:** ~14
**Pattern extraction date:** 2026-08-06

**Key evidence lines for the planner:**
- Fixture structure source: `cyp-auth-infra/saira-20260320.c2.full.toml` (500 lines; top-level units :7-42, deep nesting :70-74, `length=3` cross-level link :392-399, `length=2` links :175/:259-260/:270)
- Tests to repoint: `internal/graph/builder_test.go:795-840` (RealToml), `:1205-1232` (BaselineDOT); fixture paths at :803, :1211; hardcoded names at :811/:814/:825
- D-04 machinery evidence: `internal/view/scope.go:13-49` (GenerateExpandedView never consults `Properties.Expanded`), `:182-188` (isExpandedInC1 OR semantics, D-03), `internal/model/properties.go:19-20`
- Golden DOT generation: `internal/render/render.go:31-33` (RenderDOT), wasmMutex :15-20/:86-88; private baseline shape `cyp-auth-infra/saira-20260320.c2.full.expanded.dot:1032-1040` (minlen=3/penwidth=2 edge block)
- COMPAT-01 anchor: `cmd/c4drill/testdata/valid.toml:1-26` (no expansion hints); assertion pattern `internal/graph/builder_test.go:116-143` (`[+]` indicator)
