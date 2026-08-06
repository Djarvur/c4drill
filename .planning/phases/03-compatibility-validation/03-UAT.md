---
status: complete
phase: milestone-wide (v1.8, session in 03-compatibility-validation)
source: [03-01-SUMMARY.md, 03-02-SUMMARY.md, 02-01-SUMMARY.md, 01-01-SUMMARY.md, 01-02-SUMMARY.md, 01-03-SUMMARY.md]
started: 2026-08-06T10:45:00Z
updated: 2026-08-06T11:05:00Z
---

## Current Test

[testing complete]

## Tests

### 1. C1 shows only top-level units
expected: C1 graph has exactly 5 nodes (actorA, actorB, actorC, externalSys, mainSystem); no nested units as nodes.
result: pass
evidence: CLI-generated C1 SVG contains exactly Actor A, Actor B, Actor C, External IDP, mainSystem (Linux server). No nested units appear. (VIEW-01)

### 2. Deep link resolves to visible ancestor
expected: Links from deeply nested units render as edges between top-level units.
result: pass
evidence: C1 DOT shows 4 edges all between top-level units: actorA→mainSystem, actorB→mainSystem, actorC→mainSystem, mainSystem→externalSys. The deep client→externalSys link resolved to mainSystem→externalSys. (VIEW-02, EDGE-01)

### 3. Duplicate edges collapse with penwidth
expected: Collapsed edges render penwidth=2, single edges penwidth=1.
result: pass
evidence: C1 DOT — 2 edges penwidth=2 (collapsed: the 3 actor→mainSystem links collapsed to one thickened edge, plus mainSystem→externalSys), 2 edges penwidth=1. (EDGE-02, D-04)

### 4. Expanded units show one level with direct edges
expected: properties.expanded unit shows as cluster with direct subunits; edges attach to visible subunits.
result: skipped
reason: multilevel.toml has no properties.expanded (COMPAT-01 fixture); not exercised in this run. Phase 1 verification (9/9) and Phase 2 D-04 tests cover this.

### 5. C2/C3 sub-diagram files auto-generated
expected: C2/C3 files created with working explore links from C1.
result: issue
reported: "C2 explore URLs are WRONG: C2 diagram (at multilevel/mainSystem.svg) emits href='mainSystem/sshAuth.svg' but should emit 'sshAuth.svg' (C3 files are siblings). Root cause: BuildGraphWithPath passes node.ID (full dotted path 'mainSystem.sshAuth') to ComputeExploreURL, which expects a leaf name. Same bug produces URL='.svg' (empty) for parent node in C3 diagrams. Clicking any C2 explore link 404s."
severity: blocker

### 6. Box sub-diagrams
expected: Box with subunits gets its own sub-diagram.
result: pass
evidence: Phase 2 TestFullPipeline_BoxWithSubunitsGeneratesSubDiagram verified boxtest/boxname.svg generation. (No box in multilevel fixture, but covered by Phase 2 regression.)

### 7. Expanded-but-empty renders as plain node
expected: properties.expanded unit with no subunits renders as plain node.
result: skipped
reason: No such case in fixtures; Phase 2 TestBuildGraphExpandedEmptyUnitRendersPlainNode + WR-01 C2/C3 variant cover it.

### 8. --expanded output unchanged
expected: --expanded produces one all-nested diagram, minlen + penwidth=2 preserved.
result: pass
evidence: multilevel.expanded.dot has 56 edges all penwidth=2, single all-nested diagram, matches golden baseline (canonicalDOT semantic comparison). (COMPAT-02)

### 9. No-properties file renders collapsed C1
expected: TOML without properties.expanded renders C1 all-collapsed.
result: pass
evidence: valid.toml C1 DOT — 0 clusters, all units collapsed ([+] labels). C2 sub-diagram still generated. (COMPAT-01)

## Summary

total: 9
passed: 6
issues: 1
pending: 0
skipped: 2

## Gaps

- truth: "C2/C3 sub-diagram explore links must resolve to the correct sibling file (e.g. C2 at {basename}/{system}.svg links to {system}/{container}.svg via relative path '{container}.svg', not '{system}/{container}.svg')"
  status: failed
  reason: "BuildGraphWithPath (internal/graph/builder.go:596,605) passes node.ID (full dotted path like 'mainSystem.sshAuth') to ComputeExploreURL, but ComputeExploreURL expects a leaf target name. Result: C2 emits 'mainSystem/sshAuth.svg' (wrong — should be 'sshAuth.svg'); C3 parent node emits '.svg' (empty). Reproduced with cmd/c4drill/testdata/multilevel.toml — all C2 hrefs broken."
  severity: blocker
  test: 5
  artifacts:
    - internal/graph/builder.go:587-616 (BuildGraphWithPath)
    - internal/graph/path.go:14-66 (ComputeExploreURL — correct in isolation, wrong input)
  missing:
    - Either pass the leaf name (last path segment) to ComputeExploreURL, or make ComputeExploreURL accept full dotted paths and strip the current-path prefix
    - Regression test: C2 diagram hrefs resolve to sibling files, not C1-style paths
- truth: "Navigation breadcrumb/back-link HTML must render as clickable <a> tags in SVG, not escaped literal text"
  status: failed
  reason: "internal/render/converter.go:193 calls cg.SetLabel(joinLabels(labelParts)) for the graph label (where navigation goes) WITHOUT cg.StrdupHTML(), so GraphViz treats the <a href> tags as literal text and escapes them to &lt;a href=&quot;...&quot;&gt;. Node labels work because they use StrdupHTML (converter.go:264). Reproduced: C2/C3 SVGs show the breadcrumb bar as raw escaped text."
  severity: blocker
  test: 5
  artifacts:
    - internal/render/converter.go:175-195 (graph label construction — missing StrdupHTML)
    - internal/render/navigation.go (BuildNavigationLabel — produces correct HTML, but it gets escaped downstream)
  missing:
    - Wrap the graph label (or at least the navigation portion) as an HTML-like label via StrdupHTML
    - Regression test: SVG output contains literal <a href=...> in the navigation bar, not &lt;a href=
- truth: "Navigation breadcrumb/back-link URLs must use .svg (browser-navigable), consistent with explore URLs"
  status: failed
  reason: "ComputeExploreURL hardcodes '.svg' (path.go:16) for browser navigation, but BuildNavigationLabel/converter passes the render format (e.g. '.dot') to ComputeBackLinkURL/buildBreadcrumbs. Result: in a .dot render, breadcrumb links point to .dot files while explore links point to .svg — inconsistent. The C3 navigation label shows href=\"../mainSystem.dot\" but explore URLs are .svg."
  severity: major
  test: 5
  artifacts:
    - internal/render/converter.go:180 (passes `format` to BuildNavigationLabel indirectly via graph construction)
    - internal/graph/builder.go:612,647,651 (buildNavigation/buildBreadcrumbs receive render format)
    - internal/graph/path.go:72-135 (ComputeBackLinkURL, computeBreadcrumbURL use the passed format)
  missing:
    - Force .svg in navigation URLs (matching ComputeExploreURL), OR document the inconsistency
    - Regression test: all clickable URLs in a diagram use the same format
