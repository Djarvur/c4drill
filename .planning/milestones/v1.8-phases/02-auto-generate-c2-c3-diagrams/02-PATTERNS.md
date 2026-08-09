# Phase 2: Auto-generate C2/C3 Diagrams - Pattern Map

**Mapped:** 2026-08-06
**Files analyzed:** 5 (1 modify + 3 test modifies + 1 read-only constraint)
**Analogs found:** 4 / 4 modification targets (plus 1 confirmed gap per file)

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| `internal/graph/builder.go` (MODIFY — D-07 guard) | builder/service | transform (view→graph) | own C1 branch (builder.go:56-76); expansion decision anchor `isExpandedInC1` (scope.go:182-188) | exact (self-modification) |
| `internal/graph/builder_test.go` (MODIFY — D-07 test) | test | n/a | `TestBuildGraphClusters` (builder_test.go:216-251) | exact |
| `cmd/c4drill/root_test.go` (MODIFY — box sub-diagram file test, D-01) | test | n/a | `TestFullPipeline_NestedWithExpanded` (root_test.go:334-390) | exact |
| `internal/view/scope_test.go` (MODIFY — actor boundary node in C2, D-08) | test | n/a | `TestGenerateC2View_ExternalBoundaryFromSubunitLinks` (scope_test.go:664-691) | exact |
| `internal/view/scope.go` (DO NOT MODIFY — Phase 1 WR-01 constraint) | view | transform | `isExpandedInC1` (scope.go:182-188) | read-only reference only |

**Confirmed coverage gaps (from grep of all three test files):**
- No "box" anywhere in `cmd/c4drill/root_test.go` — box sub-diagram file test does NOT exist (Wave 0 gap).
- No person/personExternal type used in any C2/C3 boundary test in `internal/view/scope_test.go` (TypePerson only at :42 C1, :750 expanded; TypePersonExternal only at :1010 C1 fixture) — D-08 actor-in-C2 test does NOT exist.

## Pattern Assignments

### `internal/graph/builder.go` (builder, transform) — D-07 guard

**Analog:** the file's own existing C1 branch; the guard condition mirrors the `HasSubunits` field already computed per entry.

**C1 branch to modify** (builder.go:56-76) — add `len(entry.Unit.Subunits) > 0` to the cluster decision:

```go
} else {
    // C1 view: build nodes and clusters in definition order (from view.UnitOrder)
    for _, key := range v.UnitOrder {
        // Visible subunits are rendered inside their parent cluster by
        // buildCluster — skipping prevents duplicate node IDs in DOT.
        // Nil-map reads are safe, so views without VisiblePaths (C2/C3,
        // expanded, hand-built) are unaffected.
        if v.VisiblePaths[key] {
            continue
        }

        entry := v.Units[key]
        if entry.IsExpanded {
            cluster := buildCluster(entry)
            g.Clusters = append(g.Clusters, cluster)
        } else {
            node := buildNode(entry)
            g.Nodes = append(g.Nodes, node)
        }
    }
}
```

**D-07 change:** line 68 `if entry.IsExpanded` becomes `if entry.IsExpanded && len(entry.Unit.Subunits) > 0`. A unit with no subunits then falls to `buildNode` (line 71-74) — plain collapsed node. Note `buildNode` already handles the no-subunits case: the `[+]` indicator (builder.go:250-252) only applies when `entry.HasSubunits && !entry.IsExpanded`, so the D-07 node renders without `[+]`.

**DO NOT touch the C2/C3 branch** (builder.go:30-55) — the same `if entry.IsExpanded` cluster pattern exists inside the boundary cluster loop; Phase 1 WR-01 work lives downstream in `buildEdges`/`countPairMultiplicity` (builder.go:337-437) and any edit risks regressing `TestBuildEdgesPenwidthC2C3CollapsedPairs` / `TestBuildEdgesPenwidthLinkFromContributions`.

**Expansion decision anchor (read-only)** — `isExpandedInC1` (scope.go:182-188), OR semantics:

```go
func isExpandedInC1(m *parser.Model, unit *model.Unit, unitPath string) bool {
    if slices.Contains(m.Properties.Expanded, unitPath) {
        return true
    }

    return slices.Contains(unit.Expanded, unitPath)
}
```

D-05 (OR) and D-06 (silent ignore) already hold here — no change. D-07's guard lives downstream in BuildGraph (the C1 branch above), NOT in this function (isExpandedInC1 is C1-only, so an edit would be safe, but the graph-level guard is the agreed single point per RESEARCH Pitfall 1).

---

### `internal/graph/builder_test.go` (test) — new D-07 test

**Analog:** `TestBuildGraphClusters` (builder_test.go:216-251) — exact cluster-vs-node assertion pattern:

```go
func TestBuildGraphClusters(t *testing.T) {
    t.Parallel()

    m := &parser.Model{
        Properties: model.Properties{Name: "Test"},
        Units: map[string]*model.Unit{
            "app": {
                Type:     model.TypeSystem,
                Name:     "App",
                Expanded: []string{"app"}, // Expanded
                Subunits: map[string]*model.Unit{
                    "api": {Type: model.TypeContainer, Name: "API"},
                    "web": {Type: model.TypeContainer, Name: "Web"},
                },
            },
        },
    }

    v := view.GenerateC1View(m)
    g := graph.BuildGraph(v)

    // Test 5: BuildGraph creates Cluster for each expanded ViewUnit
    require.Len(t, g.Clusters, 1)

    cluster := g.Clusters[0]
    assert.Equal(t, "cluster_app", cluster.ID)

    // Test 14: Cluster contains nodes for expanded unit's children
    assert.Len(t, cluster.Nodes, 2)
}
```

**D-07 test shape (mirror the analog, invert the expectation):** same fixture style but `Properties: model.Properties{Name: "Test", Expanded: []string{"app"}}` (properties.expanded is the D-07 wording) with `"app"` having NO `Subunits`. Assert `require.Len(t, g.Clusters, 0)` and `require.Len(t, g.Nodes, 1)` with `assert.Equal(t, "app", g.Nodes[0].ID)`; optionally `assert.NotContains(t, g.Nodes[0].Label.Name, "[+]")`. Follow the existing "// Test N:" numbered-comment convention (CONVENTIONS.md — requirement traceability).

**Conventions to replicate (verified in this file):**
- External package: `package graph_test` (builder_test.go:1)
- Imports grouped stdlib → module → testify (builder_test.go:3-16)
- `t.Parallel()` at test top and in every `t.Run` (unless render is involved)
- `require` for fatal conditions, `assert` for value checks, with field-path message args
- `//nolint:funlen // Test functions with model setup are naturally longer` for long tests (builder_test.go:290)

---

### `cmd/c4drill/root_test.go` (test) — new box sub-diagram file test (D-01)

**Analog:** `TestFullPipeline_NestedWithExpanded` (root_test.go:334-390) — inline-TOML + file-layout assertions:

```go
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestFullPipeline_NestedWithExpanded(t *testing.T) {
    // Test with nested model that has expanded units
    tmpDir := t.TempDir()

    // Create a test file with expanded units
    testPath := filepath.Join(tmpDir, "expanded.toml")
    content := `
[properties]
name = "Expanded Test"

[mainapp]
type = "system"
name = "Main App"
technology = "Go"
expanded = ["mainapp"]

[mainapp.api]
type = "container"
...
`
    err := os.WriteFile(testPath, []byte(content), 0o600)
    require.NoError(t, err)

    outputDir := filepath.Join(tmpDir, "output")
    cmd := NewRootCmd()
    buf := &bytes.Buffer{}
    cmd.SetOut(buf)
    cmd.SetErr(buf)

    if err := cmd.PersistentFlags().Set("output", outputDir); err != nil {
        t.Fatalf("failed to set output flag: %v", err)
    }

    cmd.SetArgs([]string{testPath})

    err = cmd.Execute()
    require.NoError(t, err, "Should succeed for expanded model")

    // Verify C1 diagram was created
    assert.FileExists(t, filepath.Join(outputDir, "expanded.svg"), "C1 diagram should exist")

    // Verify C2 diagram for expanded system was created
    assert.FileExists(t, filepath.Join(outputDir, "expanded", "mainapp.svg"), "C2 diagram for mainapp should exist")
}
```

**Box test shape:** same structure with `[boxname] type = "box"` at top level, `[boxname.child]` inside, then assert `assert.FileExists(t, filepath.Join(outputDir, basename, "boxname.svg"))` — the C2 file layout (no-dot path → `{basename}/{unit}.{format}` per writer.go:37-47). Per-unit `expanded = ["boxname"]` is NOT needed for the sub-diagram file — auto-detect (`collectExpandableUnitPaths`, root.go:170-199) fires on `len(unit.Subunits) > 0` alone; the file is generated whenever the unit has subunits (D-01 uniform rule). A companion `TestExpandedUnits`-style fixture check (root_test.go:483-507, uses `testdata/expanded.toml` + `os.Stat`/`require.NoError`) is the alternative if a fixture is preferred.

**Conventions to replicate (verified in this file):**
- NO `t.Parallel()` — render tests must be annotated `//nolint:paralleltest // go-graphviz WASM engine has concurrency issues` (root_test.go:15, 282, 333)
- Inline TOML written to `t.TempDir()` with `os.WriteFile(path, []byte(content), 0o600)`; package is `package main` (white-box, tests `NewRootCmd`/`cmd.Execute`)
- Fixture-based alternative: `filepath.Join("testdata", "expanded.toml")` relative to the package dir `cmd/c4drill/testdata/`

---

### `internal/view/scope_test.go` (test) — new D-08 actor boundary node in C2 test

**Analog:** `TestGenerateC2View_ExternalBoundaryFromSubunitLinks` (scope_test.go:664-691):

```go
func TestGenerateC2View_ExternalBoundaryFromSubunitLinks(t *testing.T) {
    t.Parallel()

    m := &parser.Model{
        Properties: model.Properties{Name: "Test"},
        Units: map[string]*model.Unit{
            "system": {
                Type: model.TypeSystem,
                Name: "System",
                Subunits: map[string]*model.Unit{
                    "api": {
                        Type: model.TypeContainer,
                        Name: "API",
                        Links: []model.Link{
                            {Peer: "externaldb"},
                        },
                    },
                },
            },
        },
    }

    v := view.GenerateC2View(m, "system")

    require.NotNil(t, v)
    assert.Contains(t, v.Units, "externaldb")
    assert.True(t, v.Units["externaldb"].IsExternal)
}
```

**D-08 test shape:** identical, but the peer is an actor: add top-level `"webUser": {Type: model.TypePersonExternal, Name: "Web User"}` to the model and `{Peer: "webUser"}` on the container's Links; assert `assert.Contains(t, v.Units, "webUser")` and `assert.True(t, v.Units["webUser"].IsExternal)`. The boundary-node machinery is `addExternalBoundaryNodesForSubunits` → `addResolvedBoundaryNode` (scope.go:557-625) — `createExternalBoundaryNode` (scope.go:193-218) preserves the actual unit's data and `IsExternal` from `IsExternalType` (view.go:74-79), which returns true for `TypePersonExternal`. This locks D-08 (actors NOT filtered from deeper views) without touching scope.go.

---

## Shared Patterns

### Definition-order iteration (UnitOrder/SubunitOrder with map-key fallback)
**Sources:** scope.go:27-36 (view-level), scope.go:138-146 + builder.go:292-300 (cluster children), root.go:178-185 (auto-detect recursion), builder.go:199-208 (nested cluster).
**Apply to:** Any new code iterating model units; tests asserting deterministic output.

```go
// scope.go:138-146 — canonical order selection
var subunitOrder []string
if len(entry.Unit.SubunitOrder) > 0 {
    subunitOrder = entry.Unit.SubunitOrder
} else {
    for subName := range entry.Unit.Subunits {
        subunitOrder = append(subunitOrder, subName)
    }
}
```

### buildCluster one-level rendering (D-04 anchor)
**Source:** builder.go:273-319 — expanded unit's DIRECT subunits become nodes inside `cluster_<FullPath>`; no recursion (only `buildNestedCluster` at :180-238 recurses for `--expanded` mode). D-04 already holds — do not change.

### Writer dotted-path layout (D-03 anchor)
**Source:** writer.go:37-47:

```go
// C2/C3: directory hierarchy from dotted path
dirPath := strings.ReplaceAll(unitPath, ".", string(filepath.Separator))
relPath = filepath.Join(basename, dirPath+"."+format)
```

URL derivation must stay in sync via `ComputeExploreURL`/`ComputeBackLinkURL`/`BuildBreadcrumbPath` (path.go:14-135) — D-03's single-source-of-truth. No naming changes.

### Test conventions (project-wide, per .planning/codebase/CONVENTIONS.md and TESTING.md)
- testify: `require` for fatal (nil, error expectations, `require.Len`), `assert` for value checks with field-path message args
- External test packages (`graph_test`, `view_test`) for public APIs; `package main` white-box in root_test.go
- `t.Parallel()` everywhere EXCEPT render-touching tests, which carry `//nolint:paralleltest // go-graphviz WASM engine has concurrency issues`
- Inline model literals `&parser.Model{...}` with `map[string]*model.Unit` fixtures; shared fixtures via `expandedC1Model`-style helpers (scope_test.go:1005-1044) with package-level path constants (scope_test.go:991-998)
- Numbered comments mapping to spec cases ("// Test 5: ..."), and decision-traceability comments ("// D-01: ...") throughout builder_test.go (e.g., :1520-1528)
- `//nolint:funlen` on long test functions with explanation
- No new dependencies — zero (stdlib + testify v1.11.1 only)

### CLI error/silence conventions (context only — no change expected)
- Silent on success: `return nil // Success - silent per spec` (root.go:148); errors to stderr only (D-06 consistent)
- Stage-wrapped errors: `fmt.Errorf("parse: %w", err)` etc. (root.go:112-114); sentinel errors in `var (...)` block (root.go:23-28)

## No Analog Found

None — every modification target has an exact analog. The D-07 guard is a one-condition refinement of code that already exists; all three test additions clone existing test shapes with inverted/changed fixtures.

## Metadata

**Analog search scope:** `internal/graph/`, `internal/view/`, `internal/output/`, `cmd/c4drill/`, `internal/model/`, `.planning/codebase/`
**Files scanned:** 9 (builder.go, builder_test.go, scope.go, scope_test.go, view.go, path.go, writer.go, root.go, root_test.go, CONVENTIONS.md)
**Pattern extraction date:** 2026-08-06
