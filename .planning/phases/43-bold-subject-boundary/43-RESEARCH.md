# Phase 43: Bold Subject Boundary - Research

**Research date:** 2026-09-03
**Method:** Codebase-verified investigation (read-only scan of builder/converter/view/test paths) — the phase is an in-pattern attribute-emission change; no external services or new dependencies involved, so no web research was required.

## User Constraints (from CONTEXT.md)

- **D-01:** Triple width = `penwidth=3.0` graphviz cluster attribute, applied ONLY to the subject boundary cluster created by `buildBoundaryCluster` (internal/graph/builder.go:373) — the single construction site that comments "the boundary frame IS the unit on its own child diagram".
- **D-02:** `graph.NodeStyle` (graph.go:195) gains a `BorderWidth float64` field — 0 means "renderer default" (no attribute emitted), >0 emits `penwidth`. `applyClusterStyle` (converter.go:656) emits it via `SafeSet("penwidth", ...)`. Edge precedent: `e.SetPenWidth(edge.PenWidth)` with 0→1.0 fallback at converter.go:899-905.
- **D-03:** Semantic navigation aid, NOT author formatting: survives `--plain` and `--no-styles`. No new CLI flag, no model key. `--no-colors` does not interact.
- **D-04:** Everything else byte-identical: `--expanded` copies, collapsed C1 root, non-subject clusters, node shapes, legend.
- **D-05:** TDD RED→GREEN, raw-DOT assertion (subject cluster carries `penwidth=3`, no other cluster does), multi-level fixture (C2 + deep-link C3), goldens re-baselined only for subject-boundary golden, per-golden diff = penwidth attribute only (canonicalDOT keeps attributes).

## Phase Requirements

- **BOLD-01:** Every non-expanded (drill-down) view draws the boundary of the depicted element with a bold triple-width border.
- **BOLD-02:** Semantic — survives `--plain`/`--no-styles`; no other cluster on the same view affected.
- **BOLD-03:** Expanded-mode views, node borders, legend, edge styling unchanged; golden updates limited to the subject-boundary delta.

## Summary

Add `BorderWidth float64` to `graph.NodeStyle` (0 = renderer default → no attribute), set it to `3.0` **inside `buildBoundaryCluster`** (after `applyUnitOverrides`, before the `return` — covering both the unit and nil-unit branches, one construction site per D-01), and emit it in `applyClusterStyle` via `setClusterAttribute(subgraph, "penwidth", "3.0")`. Because `applyClusterStyle` runs for every cluster but only the subject boundary cluster's style carries a non-zero `BorderWidth`, exactly one cluster per drill-down view changes. The emphasis is decided at graph-construction time, outside the author-override path (`applyUnitOverrides` early-returns under `--plain`), so it is semantic and survives `--plain`/`--no-styles` by construction. C1 root, `--expanded`, and unrelated clusters never construct a boundary cluster, so they stay byte-identical.

## Architectural Responsibility Map

| Tier | Component | Responsibility | Change |
|------|-----------|----------------|--------|
| view | `internal/view/scope.go` | View generation (C1/C2/C3/deep-link routing, Plain/NoStyles flags) | none — flags already flow to the builder via `renderOptsFromView` |
| graph | `internal/graph/graph.go` | `NodeStyle` value type | ADD `BorderWidth float64` (0-means-default contract, D-02) |
| graph | `internal/graph/builder.go` | Graph construction; boundary cluster creation | ADD emphasis assignment inside `buildBoundaryCluster` (D-01) |
| render | `internal/render/converter.go` | DOT/SVG/HTML attribute emission | ADD `penwidth` emission in `applyClusterStyle` (D-02) |
| test | `internal/graph/builder_test.go`, `internal/render/converter_test.go`, `cmd/c4drill/root_test.go` | Raw-DOT + byte-equality + canonical-golden assertions | ADD/EXTEND tests (D-05) |
| testdata | `cmd/c4drill/testdata/multilevel.toml` + derived `.dot` golden files | Multi-level public fixture | reuse; re-baseline ONLY goldens whose canonical diff is the sole `penwidth` line (expected: none) |

## Standard Stack

- Go (module `github.com/Djarvur/c4drill`, toolchain 1.26.x), pinned go-graphviz fork (`goccy/go-graphviz@v0.2.10`, WASM engine) — attribute emission via `cgraph.Graph.SafeSet`.
- Tests: stdlib `testing` + `stretchr/testify` (require/assert), canonical comparator `internal/testutil/canonical` (order-insensitive, geometry-stripped, **attributes kept**).
- No new packages, no new external services, no framework installs.

## Architecture Patterns

### Drill-down view routing (the "which views get a boundary cluster" map)

`internal/graph/builder.go:33` `BuildGraph`:

```go
// For C2/C3 views, wrap internal nodes in a boundary cluster
if v.Level != view.LevelC1 && v.ExpandedUnit != "" {
    buildBoundaryViewGraph(v, g)
} else {
    // C1 view: build nodes and clusters in definition order
    buildC1ViewGraph(v, g)
}
```

- C2 (GenerateC2View, scope.go:487), C3 (GenerateC3View, scope.go:594), and deep-link drill-downs (ensureDeepLinkChain, scope.go:397 → normal C2/C3 views) all route here with `Level != C1 && ExpandedUnit != ""` → **one** boundary cluster per diagram, built by `buildBoundaryCluster`.
- Collapsed C1 root → `buildC1ViewGraph` (no boundary cluster).
- `--expanded` → `BuildExpandedGraph` (builder.go:447, no boundary cluster).
- Deep-link drill-down (`BuildGraphWithPath`, builder.go:1622) produces ordinary C2/C3 views — same `BuildGraph` path.

### buildBoundaryCluster (builder.go:373-441) — THE construction site

```go
style = GetStyleForType(unit.Type, false)
applyUnitOverrides(style, unit, renderOptsFromView(v))
...
// Fallback (unit == nil): style = &NodeStyle{BorderColor: ..., FontColor: ...}
return &Cluster{
    ID:       v.ExpandedUnit,
    Label:    label,
    ...
    Style:    style,
    ...
}
```

**Critical pitfall (verified):** `applyUnitOverrides` (builder.go:624) early-returns when `opts.Plain`:

```go
func applyUnitOverrides(style *NodeStyle, unit *model.Unit, opts RenderOpts) {
    if opts.Plain || style == nil || unit == nil {
        return
    }
```

So the bold emphasis MUST be assigned in `buildBoundaryCluster` itself (after the overrides), NOT inside `applyUnitOverrides` — otherwise `--plain` would drop it and violate D-03. Assigning once before the `return` covers both the unit and nil-unit fallback branches and keeps D-01's "single construction site" intact.

### Cluster attribute emission (converter.go)

- `createCluster` (converter.go:573) calls `applyClusterStyle(subgraph, cluster.Style)` (line 586) for **every** cluster — in **all** render modes (`opts.Plain` only switches the label path, lines 619-620; style application is mode-agnostic). This is exactly why the semantic emphasis survives `--plain` at emit time.
- `applyClusterStyle` (converter.go:656-695) emits color/fontcolor/fontname via `setClusterAttribute(subgraph, <attr>, value)` which skips empty values and calls `subgraph.SafeSet(attr, value, "")` (converter.go:696-704).
- **The cluster (`cgraph.Graph` subgraph) has no SetPenWidth method** — the edge setter (`e.SetPenWidth`, go-graphviz attribute.go:1962) is `Edge`-only. Clusters emit penwidth through the `setClusterAttribute`/`SafeSet` path, exactly like `color`.

### Edge penwidth precedent (D-02 mirror contract)

converter.go:899-905:

```go
// Set edge penwidth per D-04: collapsed pairs (2+ links) and --expanded
// edges carry PenWidth 2.0 from the builder; single edges (PenWidth 0)
// render at the default 1.0.
if edge.PenWidth > 0 {
    e.SetPenWidth(edge.PenWidth)
} else {
    e.SetPenWidth(1.0)
}
```

Pinned-fork SafeSet semantics: `SetPenWidth` = `SafeSet("penwidth", fmt.Sprint(v), "1.0")` — an attribute whose value equals the passed default is **omitted** (that's how 0→"renderer default" works). For the cluster we own the emitted string: **`strconv.FormatFloat(style.BorderWidth, 'f', 1, 64)` → `"3.0"`** — deterministic, exactly matches the D-01 literal `penwidth=3.0`, and the `def ""` in `setClusterAttribute` guarantees emission whenever `BorderWidth > 0`.

### Plain semantics of derived styling (D-03 anchor)

`--plain`/`--no-styles` gate only *author* overrides (`applyUnitOverrides`, builder.go:624-651). Kind-derived semantics are decided at build time and emitted unconditionally — same treatment as legend rows and kind-coloured edges (PLAIN-02). The boundary emphasis follows this exact pattern: decided in `buildBoundaryCluster`, not gated by any `opts` flag. `--no-colors` (NoColors) touches only color attributes; penwidth is not a colour — no interaction (D-03).

## Don't Hand-Roll

- **Do not** hand-roll cluster attribute formatting — reuse `setClusterAttribute` (converter.go:696) which already encodes the empty-skip + SafeSet contract.
- **Do not** emit HTML/labels/URLs here — emphasis is a pure attribute; no label or interactive changes.
- **Do not** add a CLI flag or model key (D-03): the value is a constant `3.0` decided at construction.
- **Do not** compare raw DOT bytes in tests — use `canonical.Canonical` and/or attribute multiset extraction (pinned fork map-order, DI-1).

## Runtime State Inventory

| State | Location | Relevance |
|-------|----------|-----------|
| `RenderOpts` (Plain/NoColors/NoStyles/...) | `renderOptsFromView` builder.go:16-25 | flows view flags into builder; NOT consulted for the emphasis |
| `NodeStyle.BorderWidth` | graph.go (NEW) | 0 = default; only the subject boundary cluster's style is non-zero |
| `Cluster.Style *NodeStyle` | graph.go | carried from builder to converter |
| goldens | `cmd/c4drill/testdata/*.dot` | all committed goldens are expanded/plain variants (verified below) |

## Common Pitfalls

1. **Assigning the emphasis inside `applyUnitOverrides`** → silently dropped under `--plain` (early return, builder.go:624). Top trap; the plan's RED test must assert `--plain` output carries the attribute.
2. **Setting penwidth on all clusters** — emission must stay keyed to the per-cluster style (only the subject boundary cluster has `BorderWidth > 0`). Do not touch `createCluster`'s loop or `applyClusterStyle`'s nil-style branch.
3. **Asserting raw byte order or full DOT text** — sibling statement order is map-order dependent in the pinned fork (DI-1). Assert attributes via canonical comparison + targeted regexp/multiset extraction.
4. **"3" vs "3.0" drift** — pin ONE formatting spec (`strconv.FormatFloat(w, 'f', 1, 64)` → `"3.0"`) and assert exactly `penwidth=3.0` in the raw DOT; do not use `fmt.Sprint` (would emit "3").
5. **Golden re-baseline without a diff audit** — every re-baselined golden's canonical diff must be ONLY the added `penwidth` line on the **subject** cluster. Current evidence says no committed golden renders a non-expanded drill-down DOT, so the expected scope is **zero re-baselines**; if the full suite says otherwise, audit before touching.
6. **Clusters are `cgraph.Graph`, not edges** — no `SetPenWidth`; use `setClusterAttribute`. (Edges keep their existing path untouched.)
7. **Tests touching the WASM engine must not run in parallel** — `//nolint:paralleltest` convention (root_test.go:689, etc.).

## Code Examples

### Emission insertion (applyClusterStyle tail, converter.go:689-695 area)

```go
// Set border width (BOLD-01): only the subject boundary cluster carries a
// non-zero BorderWidth; 0 = renderer default (no attribute, D-02).
if style.BorderWidth > 0 {
    if err := setClusterAttribute(subgraph, "penwidth",
        strconv.FormatFloat(style.BorderWidth, 'f', 1, 64)); err != nil {
        return err
    }
}
```

### Construction insertion (buildBoundaryCluster, builder.go — after applyUnitOverrides, before `return &Cluster{...}`)

```go
// Bold subject boundary (BOLD-01): the boundary frame IS the depicted unit,
// so it draws at triple width. Semantic — set at construction, outside the
// author-override path, so it survives --plain/--no-styles (D-03).
style.BorderWidth = 3.0
```

### Raw-DOT assertion pattern (analog: dotEdgeMultiset, converter_test.go:1257-1283)

```go
penRe := regexp.MustCompile(`penwidth=([\d.]+)`)
// extract per-subgraph attribute sets from the DOT; assert exactly ONE
// cluster carries penwidth=3.0 and it is the subgraph "cluster_<subject>"
```

### Byte-equality guards (existing harness)

- `canonical.Canonical(t, dot)` compare against `cmd/c4drill/testdata/multilevel.expanded.dot` (builder_test.go:1233-1241, COMPAT-02; 2846-2851, REF-05) — must stay green unchanged.
- CLI harness `generateMultilevelOutput` (root_test.go:680-800) renders `multilevel.toml` → C1 `multilevel.dot` + C2 `multilevel/mainSystem.dot` + C3 `multilevel/mainSystem/sshAuth.dot` — the natural vehicle for end-to-end raw-DOT assertions on C2 + deep-link C3 drill-downs (BOLD-01) and for `--expanded`/`--plain` variants.

## State of the Art

`penwidth` is a standard graphviz cluster attribute (https://graphviz.gitlab.io/_pages/doc/info/attrs.html#a:penwidth); the pinned go-graphviz fork exposes it for edges via `SetPenWidth` and for graphs/nodes via generic `SafeSet`. Cluster border thickness via `penwidth` on a subgraph is supported by `dot` (default 1.0), so `3.0` renders a distinctly thicker frame while leaving node shapes, legend, and edges untouched.

## Assumptions Log

| ID | Assumption | Confidence |
|----|-----------|------------|
| A1 | Deep-link drill-down views are ordinary C2/C3 views routed through `BuildGraph` → boundary cluster present | HIGH (scope.go:397 ensureDeepLinkChain produces C2/C3 views; BuildGraphWithPath shares BuildGraph) |
| A2 | No committed golden currently renders a non-expanded drill-down DOT (all `testdata/*.dot` are expanded/plain/nolabels variants; canonical golden consumers read `multilevel.expanded.dot`) | MEDIUM — full-suite run in 43-01 REFACTOR + 43-02 audit arbitrates |
| A3 | graphviz `dot` honours cluster `penwidth` (documented attribute; default 1.0) | HIGH (documented; visual confirmation optional in 43-02) |
| A4 | `strconv.FormatFloat(3.0, 'f', 1, 64)` == `"3.0"` and `setClusterAttribute` (def `""`) emits it unconditionally | HIGH (stdlib semantics; SafeSet with def "" is the existing color/fontcolor path) |

## Open Questions (RESOLVED)

- **Q1 — "3" or "3.0"?** RESOLVED: emit `"3.0"` (D-01 literal; `FormatFloat 'f' 1`). Tests assert `penwidth=3.0`; do not use `fmt.Sprint`.
- **Q2 — emphasise the nil-unit fallback style too?** RESOLVED: yes — the single assignment before `return &Cluster{...}` covers both branches and keeps D-01's single-site invariant. The fallback is defensive-only (drill-downs always have an ExpandedUnit model unit in practice).
- **Q3 — `BorderWidth` on `NodeStyle` vs a dedicated cluster field?** RESOLVED per D-02 recommendation: `NodeStyle.BorderWidth` — it rides the existing `Cluster.Style` plumbing (setClusterLabel/applyClusterStyle contract) with zero new wiring.

## Environment Availability

- Go toolchain + pinned go-graphviz WASM fork available in-module; tests run in-process (`go test ./...`).
- No network, no external services, no installs required. Full suite ~10s warm cache (verified precedent: phases 40/41).

## Validation Architecture

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go `testing` + `stretchr/testify` v1.12.1 |
| Config file | none — `go test` conventions (`//nolint:paralleltest` where WASM/cobra package state is shared) |
| Quick run command | `go test ./internal/graph/ ./internal/render/ -run 'SubjectBoundary|Bold|BoundaryCluster' -count=1` |
| Full suite command | `go test ./... -count=1` |
| Estimated runtime | ~10s full suite (warm cache; phases 40/41 precedent) |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| BOLD-01 | Raw DOT of C2 + deep-link C3 drill-down: subject boundary cluster carries `penwidth=3.0`, no other cluster does | unit (raw-DOT attribute multiset) | `go test ./internal/graph/ ./internal/render/ -run 'SubjectBoundary|Bold' -count=1` | ❌ TDD RED (43-01) |
| BOLD-02 | `--plain`/`--no-styles` renders keep `penwidth=3.0` on the subject cluster; no new flag; `--no-colors` unchanged | unit + CLI (flag matrix on multilevel fixture) | `go test ./internal/render/ ./cmd/c4drill/ -run 'SubjectBoundary|Bold|Multilevel' -count=1` | ❌ TDD RED (43-01) |
| BOLD-03 | `--expanded` + C1 root output byte/canonical-identical; goldens re-baselined only for subject-boundary delta | integration + golden audit | `go test ./... -count=1` | ✅ existing + new guards |

### Sampling Rate

- **Per task commit:** `go test ./internal/graph/ ./internal/render/ -count=1`
- **Per wave merge:** `go test ./... -count=1`
- **Phase gate:** full suite green before `/gsd:verify-work`; golden re-baseline only with per-golden canonical-diff audit (sole `penwidth` line on subject cluster).

### Wave 0 Gaps

- None: test fixtures and harnesses exist (`multilevel.toml`, `generateMultilevelOutput`, canonical comparator, dotEdgeMultiset pattern). New tests are TDD RED deliverables of 43-01.

## Security Domain

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | no | N/A — CLI render path, no auth surface |
| V3 Session Management | no | N/A |
| V4 Access Control | no | N/A |
| V5 Input Validation | no (unchanged) | No new input surfaces; attribute value is a hard-coded constant `3.0`, never author-controlled |
| V6 Cryptography | no | N/A |

### Known Threat Patterns for Go CLI + GraphViz WASM render

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Attribute injection via model content | Tampering | Not applicable — `penwidth` value is a compile-time constant; nothing from the model flows into it |
| Golden/byte-drift as a diff-hiding channel | Repudiation | D-04 byte-equality guards + per-golden canonical-diff audit keep the change deliberately visible and bounded |

**No new threat surface introduced by this phase** (attribute-only change, constant value, no I/O, no new inputs).

## Sources

### Primary (HIGH confidence — codebase verification this session)

- `internal/graph/builder.go:33-55` (BuildGraph drill-down routing), `:373-441` (buildBoundaryCluster, single construction site), `:447` (BuildExpandedGraph — no boundary cluster), `:624-651` (applyUnitOverrides **Plain early-return**), `:16-25` (renderOptsFromView), `:1622` (BuildGraphWithPath)
- `internal/graph/graph.go:195-206` (NodeStyle — field addition site)
- `internal/render/converter.go:573-596` (createCluster → applyClusterStyle for EVERY cluster), `:619-620` (Plain label-path only), `:656-704` (applyClusterStyle + setClusterAttribute empty-skip/SafeSet contract), `:899-905` (edge penwidth 0→1.0 precedent)
- go-graphviz pinned fork `cgraph/attribute.go:1962` (`SetPenWidth` = `SafeSet(penwidth, fmt.Sprint(v), "1.0")` — default-omission semantics) and `cgraph/cgraph.go:855-861` (`SafeSet(name, value, def)`)
- `internal/testutil/canonical/canonical.go` (order-insensitive, **attributes kept**, geometry stripped)
- `internal/render/converter_test.go:1257-1283` (dotEdgeMultiset raw-DOT extraction precedent)
- `cmd/c4drill/root_test.go:680-800` (generateMultilevelOutput harness; C2 `multilevel/mainSystem.dot`, C3 `multilevel/mainSystem/sshAuth.dot`)
- `internal/graph/builder_test.go:1233-1241, 2846-2851` (canonical golden consumers — `multilevel.expanded.dot` only)
- `internal/view/scope.go:487, 594, 397` (C2/C3/deep-link view generation)
- `.planning/phases/43-bold-subject-boundary/43-CONTEXT.md` (D-01..D-05 locked decisions)

### Secondary (MEDIUM confidence)

- A2 (no committed drill-down golden) — arbitrated by the full-suite run in 43-01 REFACTOR and the 43-02 golden audit
- graphviz `penwidth` cluster support (documented attribute; A3)

### Tertiary (LOW confidence)

- A1 deep-link edge case (CTX-02 chain to a container with subunits) — inferred from scope.go structure; covered by the multi-level fixture assertion (D-05)

## Metadata

**Confidence breakdown:**
- Change shape: HIGH — mirrors the locked edge-penwidth contract (D-02) exactly; every insertion point read in source
- Semantics-under-plain: HIGH — `applyUnitOverrides` early-return read directly; emphasis placed outside it
- Collateral-zero: HIGH — boundary cluster is the only style holder with the non-zero field; C1/expanded paths never construct one
- Golden impact: MEDIUM (expect zero; audited in-suite)

**Research date:** 2026-09-03
**Valid until:** 2026-10-03 (stable — codebase-anchored, zero external deps)

## RESEARCH COMPLETE

**Phase:** 43 - Bold Subject Boundary
**Confidence:** HIGH