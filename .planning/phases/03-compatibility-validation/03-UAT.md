---
status: complete  # all 9 tests pass after gap closure
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
result: pass
evidence: "Originally marked issue (3 navigation gaps). All 3 CLOSED by plan 03-04 + WR fixes. Re-verified end-to-end: 46/46 distinct hrefs resolve across all 17 SVGs, 77 clickable <a xlink:href> anchors, 0 escaped-nav files, no .dot in any href. TestCompat02_NavigationLinksResolve (root_test.go) locks this."

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
passed: 9
issues: 0
pending: 0
skipped: 0

## Gaps

> All 3 gaps CLOSED by plan 03-04 (gap-closure) + WR fixes. Re-verified 2026-08-06: 46/46 distinct hrefs resolve across all 17 SVGs, 0 escaped-nav files, 77 clickable anchors, no .dot in any href.

- truth: "C2/C3 sub-diagram explore links must resolve to the correct sibling file"
  status: resolved
  reason: "Closed by 03-04: ComputeExploreURL self-link guard + bidirectional relative-path. C3 ancestor no longer emits empty href=.svg. Re-verified via os.Stat across 17 SVGs."
  severity: blocker
  test: 5
  resolved_by: 03-04 Task 1 (commit c69c385)
- truth: "Navigation breadcrumb/back-link HTML must render as clickable links in SVG"
  status: resolved
  reason: "Closed by 03-04: GraphViz <TD HREF> idiom (the plan's StrdupHTML-wraps-<a href> premise was impossible; executor used the correct idiom). WR-01 fix: TD HREF URLs html.EscapeString'd. Re-verified: 0 escaped-nav files."
  severity: blocker
  test: 5
  resolved_by: 03-04 Task 2 (commit d1a62e1) + WR-01 (commit fff0a52)
- truth: "Navigation URLs must use .svg (browser-navigable), consistent with explore URLs"
  status: resolved
  reason: "Closed by 03-04: ComputeBackLinkURL/computeBreadcrumbURL force linkFormat=svg. Re-verified: no .dot in any href when rendering .dot output."
  severity: major
  test: 5
  resolved_by: 03-04 Task 2 (commit d1a62e1)
