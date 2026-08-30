---
phase: quick-260831-01u
verified: 2026-08-30T22:17:52Z
status: human_needed
score: 5/5 must-haves verified
overrides_applied: 0
human_verification:
  - test: "Re-render the ORIGINAL reporter's model (the one showing 55 vs 260 titles) with the fixed binary and confirm the non-expanded root is compact again (title/canvas count back near the 55-title baseline)"
    expected: "Non-expanded root shows the compact C1 (~55 titles), clearly distinct from --expanded output; removing `expanded:` from properties no longer coincides with expanded"
    why_human: "The reporter's model is not in the repo; the deepcross fixture is only an in-repo proxy (plan Task 1 step 5 explicitly excluded the 93.8 Mpt² re-run for this reason)"
  - test: "Visually inspect the rendered SVG/PNG of the re-baselined goldens (nolabels.dot, nolabels.expanded.dot, deepcross.dot, plain*.dot, multilevel.expanded.dot) and the --no-labels render"
    expected: "Compact root is readable; under --no-labels nodes/clusters/legend are identifiable and only edge label text is gone; no overlapping or degraded layout from the geometry changes"
    why_human: "Diagram layout quality and readability cannot be judged by grep or canonical text comparison"
---

# Quick Task 260831-01u: Fix Three Rendering Bugs — Verification Report

**Phase Goal:** Fix three rendering bugs from todos/pending: root diagram bloated by ancestor wrapping (todo attributed v1.22.0/b2447da; executor bisected actual introduction to v1.21.0 CTX-02/CTX-03), `--no-labels` must suppress only edge labels, edge merge must compare pre-flag attributes.
**Verified:** 2026-08-30T22:17:52Z
**Status:** human_needed (all 5 must-haves verified by code + behavior; 2 items require human eyes)
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
| --- | ----- | ------ | -------- |
| 1 | Non-expanded root (C1) is compact: only top-level units, author-expanded subunits, boundary nodes, CTX-02 chain paths — never whole sibling subtrees (deepcross fixture) | ✓ VERIFIED | Fresh binary render of deepcross.toml shows exactly 9 nodes (4 boundary: webUser/mobileUser/adminCli/metricsScraper; 4 chain leaves: sessionMgr/tokenApi/ingest.api/store.warehouse; 1 collapsed: actors.audit 🔍) in clusters actors⊃identity, yic⊃pipeline⊃(ingest,store). Internal siblings sessionDb, audit.ledger, store.writer ABSENT — pinned by TestC1RootStaysCompactOnDeepCrossFixture (builder_test.go:1489) + TestDeeplinkRootCompactGolden (root_test.go:1611) |
| 2 | C2/C3 drill-down keeps WRAP-01/02 ancestor wrapper clusters; wrap/invariance tests green | ✓ VERIFIED | C2 render of deepcross emits per-system diagrams with container clusters (actors.dot: cluster_actors.identity, cluster_actors.audit). TestBoundaryEntryWrappedInAncestorChain, TestFullyExternalBoundaryStaysTopLevel, TestWrapperClustersHaveNoExploreURL, TestBoundaryViewNodeSetInvariant, TestEdgeEndpointsUnchangedByWrapping all present (builder_test.go:3667-3800) and green in full suite |
| 3 | `--no-labels` suppresses ONLY edge label text; node/cluster/legend labels remain | ✓ VERIFIED | Behavior: --no-labels render carries full HTML node labels ("Session Manager" + [go] + description), legend present, all 6 edges `label=""` (default render: HTML edge labels). Code: NoLabels consulted at exactly one label site per layer — builder.go:1481 (edge Label construction) and converter.go:643 (edge SetLabel("")); all node/cluster guards removed. Flipped tests TestNoLabelsNodesKeepFullLabels / TestNoLabelsClustersKeepLabels green |
| 4 | Formatting flags never change edge count, endpoints, or multiplicity — only styling | ✓ VERIFIED | Behavior: edge endpoint multiset md5 IDENTICAL across no-flags/--no-labels/--no-colors/--no-styles/--plain/--no-labels+--no-colors on root (6 edges) AND expanded (8 edges); penwidth distribution unchanged. Code: graph.Edge.Name (graph.go:112-120), builder-assigned sanitized sequence names (builder.go:1026-1048), converter uses edge.Name verbatim with label-free fallback (converter.go:616-635). `sanitizeForName(edge.Label.Description)` absent from converter.go (plan grep criterion holds). Pinned by TestEdgeTopologyFlagInvariant (converter_test.go:1256) |
| 5 | All golden fixtures pass; nolabels goldens re-baselined; deepcross.dot pins compact root | ✓ VERIFIED | `go test ./... -count=1` = 15/15 packages ok, including TestDeeplinkRootCompactGolden, TestNoLabelsC1Golden, TestNoLabelsExpandedGolden, TestPlainFlagC1Golden. deepcross.dot (240 lines) committed; nolabels goldens show node/cluster HTML labels restored + edge label="" (re-baseline diffs eyeballed per SUMMARY) |

**Score:** 5/5 truths verified

### Required Artifacts

| Artifact | Expected | Status | Details |
| -------- | -------- | ------ | ------- |
| `cmd/c4drill/testdata/deepcross.toml` | Fixture: deep nesting, cross-container links, deep-linking actors (link + linkFrom), properties.expanded | ✓ VERIFIED | 137 lines; expanded=["actors","yic"], linkFrom deep links (lines 18-42), cross-system link actors.audit→yic.pipeline.ingest.api (line 94), internal links that must NOT flood the root (lines 62-63, 76-77, 120-121) |
| `cmd/c4drill/testdata/deepcross.dot` | Compact-root golden | ✓ VERIFIED | 240 lines; content matches fresh binary render (golden test green); key= names use builder `_1` sequence format confirming BUG-3 rename |
| `internal/graph/builder.go` | Edge-label-only NoLabels guard; flag-independent Edge identity source | ✓ VERIFIED | sanitizeEdgeName + per-(source,target) counter (1022-1048); single remaining NoLabels guard at 1481 is the edge label |
| `internal/render/converter.go` | Uses builder edge names; node/cluster labels survive --no-labels | ✓ VERIFIED | createEdge uses edge.Name verbatim, fallback is label-free (608-624); NoLabels only at 643 (edge) |
| `internal/graph/builder_test.go` | Compact-root invariant, NoLabels flips, flag-invariant topology tests | ✓ VERIFIED | TestC1RootStaysCompactOnDeepCrossFixture (1489), TestNoLabelsNodesKeepFullLabels (4344), TestNoLabelsClustersKeepLabels (4395), TestNoLabelsEdgesHaveNoLabel (4379), WRAP suite (3667-3800) |

### Key Link Verification

| From | To | Via | Status | Details |
| ---- | --- | --- | ------ | ------- |
| builder.go createEdge | converter.go createEdge | builder-assigned Edge.Name as cgraph edge name | WIRED | graph.go:120 `Name string` → builder.go:1048 assignment → converter.go:616 `edgeName := edge.Name` → :635 `CreateEdgeByName(edgeName, ...)` |
| root.go:112 --no-labels | converter.go:643 + builder.go:1481 | RenderOpts.NoLabels threading | WIRED | root.go:112-113 flag, :338/:391 `v.NoLabels = noLabels` → view.go:63 → builder.go:23 → single consultation sites at builder.go:1481 / converter.go:643 |
| scope.go ensureDeepLinkChain | builder.go buildC1ViewGraph/buildVisibleCluster | UnfoldChain + VisiblePaths | WIRED | scope.go:445 `depicted.UnfoldChain = true`, :376/:1108/:1152 chain guards → builder.go:111/263/325 VisiblePaths reads → buildVisibleCluster (274, 293) renders visible paths only |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
| -------- | ------------- | ------ | ------------------ | ------ |
| deepcross.dot render | fixture model | cmd/c4drill/testdata/deepcross.toml via parser/view/builder | Yes — unit names, technologies, descriptions from TOML appear in DOT labels | ✓ FLOWING |
| --no-labels render | edge/node/cluster labels | same pipeline with opts.NoLabels | Yes — node labels full, edge labels empty by flag | ✓ FLOWING |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
| -------- | -------- | ------ | ------ |
| Compact C1 root on deepcross | `c4drill deepcross.toml -f dot` → count nodes/clusters | 9 nodes (4 boundary + 4 chain + 1 collapsed), clusters actors⊃identity, yic⊃pipeline⊃(ingest,store); internal siblings absent | ✓ PASS |
| --no-labels edge-only | render + inspect label attributes | node labels = full HTML tables; edges `label=""`; legend node present | ✓ PASS |
| Edge topology flag-invariant (root) | edge multiset md5 across 6 flag sets | identical md5, 6 edges each | ✓ PASS |
| Edge topology flag-invariant (expanded) | same on --expanded | identical md5, 8 edges each | ✓ PASS |
| Penwidth (multiplicity display) unchanged | penwidth extraction across flags | identical distribution (6×penwidth=1) | ✓ PASS |
| C2 drill-down intact | `c4drill deepcross.toml -f dot -o dir` | actors.dot + yic.dot with container clusters | ✓ PASS |
| Full test suite | `go test ./... -count=1` | 15/15 packages ok, 0 failures | ✓ PASS |

### Probe Execution

Not applicable — no probe scripts declared in PLAN/SUMMARY and no `scripts/*/tests/probe-*.sh` convention in this Go repo; the test suite + binary renders above are the runnable checks.

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
| ----------- | ----------- | ----------- | ------ | -------- |
| BUG-1-ROOT-COMPACT | 260831-01u-PLAN | Root stays compact C1; fixture + golden pin; bisection attribution corrected/confirmed | ✓ SATISFIED | Truth 1 + artifacts; bisect verdict (v1.21.0 CTX-02/CTX-03, not b2447da) recorded in SUMMARY with empirical counts. Todo item 4 (93.8 Mpt² re-run) plan-excluded — reporter's model not in repo; routed to human verification |
| BUG-2-NO-LABELS-EDGE-ONLY | 260831-01u-PLAN | --no-labels = edge labels only (todo option 1), everywhere (builder, converter, help, docs ×4, goldens) | ✓ SATISFIED | Truth 3; help text root.go:112-113; README.adoc:1243,1300 + 3 byte-identical SKILL.md copies (line 1006/1017 edge-label-only wording; diff-verified sync) |
| BUG-3-EDGE-MERGE-PRE-FLAG | 260831-01u-PLAN | Merge keys pre-flag & canonical; converter identity flag-independent | ✓ SATISFIED | Truth 4; D-01/D-02 pair keys stay (pre-flag by construction); label-derived converter naming eliminated; validator/view audit recorded in SUMMARY (no changes needed) |

No orphaned requirements — the three BUG-* IDs are plan-local (quick task; REQUIREMENTS.md carries no phase mapping for quick tasks, and no todo was left unclaimed).

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
| ---- | ---- | ------- | -------- | ------ |
| internal/view/scope.go | 211 | "creates a minimal placeholder" in comment | ℹ️ Info | Describes deliberate data-structure construction, not unfinished work; no debt marker |
| internal/render/converter.go | 205,228,251,317,348 | `return nil` | ℹ️ Info | Normal function-terminal returns after real work (e.g., SetLabelHTML legend); not empty implementations |

No TBD/FIXME/XXX/TODO/HACK debt markers in any phase-modified file. Debt-marker gate: not triggered.

### Commit Evidence

All 7 claimed commits verified in git history with correct RED→GREEN ordering and file scope: e0df6cd (RED, tests+fixture only) → bdabffc (GREEN) · d714be5 (RED, builder_test only) → caf0eea (GREEN) · 73a27a1 (RED, converter_test only) → 3eacba3 (GREEN) · 72afbbb (refactor, converter fallback names). Working tree clean except the untracked planning artifacts.

### Human Verification Required

### 1. Reporter's real-model confirmation (todo Bug 1 item 4, plan-excluded)

**Test:** Re-render the original reporter's model (55 vs 260 titles reproduction) with the fixed binary; compare non-expanded title/canvas count against the old 55-title baseline.
**Expected:** Compact C1 restored (~55 titles); non-expanded clearly distinct from --expanded.
**Why human:** The model is not in the repository; the deepcross fixture is the in-repo proxy only. This was explicitly excluded by the plan and recorded in the SUMMARY.

### 2. Visual inspection of re-baselined renders

**Test:** Open rendered SVG/PNG for nolabels.dot, nolabels.expanded.dot, deepcross.dot, plain*.dot, multilevel.expanded.dot; confirm readability of the compact root and that --no-labels keeps nodes/clusters/legend identifiable.
**Expected:** No degraded/overlapping layout from geometry changes; edge-label-only decluttering looks right.
**Why human:** Diagram layout quality and visual readability cannot be judged by grep or canonical text comparison.

### Gaps Summary

None. All three bugs are fixed in code, pinned test-first, and confirmed behaviorally against freshly built binary output. The bisection verdict correcting the todo's attribution (v1.21.0 CTX-02/CTX-03, not v1.22.0 b2447da) is documented with empirical node/cluster counts in the SUMMARY and is consistent with the code: the fix lives in the C1 path (buildVisibleCluster + scope.go chain guards + Mirror skip), not in buildBoundaryViewGraph.

---

_Verified: 2026-08-30T22:17:52Z_
_Verifier: Claude (gsd-verifier)_
