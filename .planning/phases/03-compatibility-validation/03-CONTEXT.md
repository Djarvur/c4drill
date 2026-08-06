# Phase 3: Compatibility & Validation - Context

**Gathered:** 2026-08-06
**Status:** Ready for planning

<domain>
## Phase Boundary

Verify and lock backward compatibility of the v1.8 view-generation refinements (Phases 1-2). COMPAT-01: existing TOML files without `properties.expanded` generate correct C1 with all units collapsed. COMPAT-02: the `--expanded` flag continues to produce the single all-nested diagram unchanged (identical to v1.7). The Phase 1/2 implementation already carries the compatibility machinery (`View.AllExpanded` exemption); this phase makes the compatibility guarantees testable and enforceable (public fixture + DOT-level baseline) and validates with real TOML files.

</domain>

<decisions>
## Implementation Decisions

### Fixture & Baseline Strategy
- **D-01:** Public fixture — add a sanitized public copy of the real-world multi-level TOML (equivalent structure: top-level units with deep nested links) to `testdata/`, plus a committed `--expanded` DOT baseline. COMPAT-02 becomes enforceable in CI and fresh checkouts; no skip-if-missing paths. The private `cyp-auth-infra/` fixture stays gitignored (author's real infra data remains out of git).
- **D-02:** Compat contract is DOT-level equivalence — node/edge sets, attributes (penwidth, minlen, labels), cluster structure. SVG byte-comparison is NOT part of the contract (font rendering / go-graphviz versions differ across environments).

### Compatibility Semantics
- **D-03:** COMPAT-01 keeps OR semantics — per-unit `expanded` self-references still expand units in C1 even without `properties.expanded` (v1.7 mechanism, Phase 2 D-05). "All units collapsed" means no expansion hints by either mechanism, which is the common case.
- **D-04:** `--expanded` mode ignores `properties.expanded` entirely — the flag expands EVERYTHING in one file (unchanged v1.7 contract). `properties.expanded` only affects normal C1/C2/C3 mode.

### Claude's Discretion
None — every area was decided with concrete options.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Requirements & Scope
- `.planning/REQUIREMENTS.md` — COMPAT-01, COMPAT-02 (phase 3); VIEW-01/02, EDGE-01/02 (phase 1), VIEW-03/04/05 (phase 2) — the behaviors being protected
- `.planning/ROADMAP.md` — Phase 3 goal, success criteria (4 items incl. saira 5-node C1 + C2/C3 sub-diagrams)
- `.planning/phases/01-fix-c1-view-scoping/01-CONTEXT.md` — D-02 (--expanded exempt from pair-only collapse), D-04/D-05 (penwidth: expanded keeps 2.0) — the machinery COMPAT-02 relies on
- `.planning/phases/02-auto-generate-c2-c3-diagrams/02-CONTEXT.md` — D-05/D-06 (OR expansion + silent ignore) — COMPAT-01 interplay
- `.planning/PROJECT.md` — TOML schema, rendering behavior

### Implementation Files
- `cmd/c4drill/root.go` — `--expanded` flag handling (:133-136), `processExpandedView`
- `internal/view/scope.go` — `GenerateExpandedView` (:14), `isExpandedInC1` (:144)
- `internal/graph/builder.go` — `BuildExpandedGraph` (:125), `buildNestedCluster` (:172), `View.AllExpanded` consumption in `buildEdges` (:383/:458)
- `internal/graph/builder_test.go` — `TestBuildExpandedGraphBaselineDOT`, `TestBuildExpandedGraphRealToml` (currently depend on the gitignored `cyp-auth-infra/` fixture)
- `internal/render/converter.go` — penwidth conditional (:481-485)
- `cmd/c4drill/root_test.go` — integration test patterns (`TestFullPipeline_NestedWithExpanded`, `TestExpandedUnits`)
- `cmd/c4drill/testdata/valid.toml` — COMPAT-01 target fixture

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `View.AllExpanded` (`internal/view/view.go`): the mode discriminator — `GenerateExpandedView` sets it; `buildEdges` branches dedup key + penwidth on it (Phase 1 D-02/D-04/D-05). COMPAT-02's enforcement point.
- `GenerateExpandedView` (`scope.go:14`): adds ALL units recursively (no properties.expanded consultation) — D-04 already holds structurally.
- `isExpandedInC1` (`scope.go:144`): OR semantics — D-03 already holds.
- Test patterns: inline TOML via `t.TempDir()` + `os.WriteFile(..., 0o600)`, external packages, `//nolint:paralleltest` on render tests.

### Established Patterns
- Definition-order preservation — baseline DOT must be deterministic
- Dotted paths → dirs → URLs (path.go) — baselines reference the same layout
- CLI silent-on-success — compat validation is test-side, not CLI output
- `--expanded` writes `{basename}.expanded.{format}` (writer.go:66)

### Integration Points
- `TestBuildExpandedGraphBaselineDOT` — currently reads gitignored fixture; D-01 repoints it at `testdata/` public fixture
- `TestBuildExpandedGraphRealToml` — same treatment (or skip→remove in favor of public fixture)
- `cmd/c4drill/testdata/valid.toml` — COMPAT-01 regression anchor
- Full-suite gate: `go test ./...` must stay green with the fixture migration

</code_context>

<specifics>
## Specific Ideas

No specific external references — decisions captured from discussion with code grounding.

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 3-Compatibility & Validation*
*Context gathered: 2026-08-06*
