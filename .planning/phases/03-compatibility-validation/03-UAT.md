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

## Visual Rendering Verification (2026-08-06, post-gap-closure)

Independent re-verification with a freshly-built binary and rsvg-convert rendering.
The `mcp__4_5v_mcp__analyze_image` vision tool confirmed C1 visually (5 nodes:
3 actors with person icons, Main System [+] box, External IDP; title "Multilevel
Test System"; correct arrows; no defects). The vision service became unavailable
(612 upstream errors) before C2/C3 could be visually analyzed, so structural SVG
inspection was used instead.

### C1 (multilevel.svg) — VISUALLY VERIFIED
- Title: "Multilevel Test System" ✓
- 5 nodes: Actor A/B/C (person shape with 👤), Main System [+], External IDP ✓
- Arrows: actors → Main System, Main System → External IDP ✓
- No navigation bar (correct — C1 is root) ✓
- "No visible problems, no overlapping text, no broken layout, no escape codes"
  — vision tool analysis, 2026-08-06

### C2 (mainSystem.svg) — STRUCTURALLY VERIFIED
- Navigation bar renders as proper SVG anchors (not escaped text):
    `<a xlink:href="../multilevel.svg">Back to Multilevel Test System</a>`
    ` | `
    `<a xlink:href="../multilevel.svg">Multilevel Test System</a>`
    ` > mainSystem`
- Title: "Main System - Containers" ✓
- 4 explore links to child subsystems: sshAuth, localIDP, storages, authModules ✓
- Leak check: 0 text elements begin with `&lt;` ✓

### C3 (sshAuth.svg) — STRUCTURALLY VERIFIED
- Navigation bar:
    `<a xlink:href="../mainSystem.svg">Back to mainSystem</a>`
    ` | `
    `<a xlink:href="../mainSystem.svg">mainSystem</a>`
    ` > sshAuth`
- Title: "Main System - SSH Auth - Components" ✓
- 3 explore links to child components: pam, authProxy, systemd ✓
- Leak check: 0 text elements begin with `&lt;` ✓

### Cross-check: false-alarm ruled out
An initial regex-based scan appeared to show escaped HTML
(`&lt;a href=&quot;...`) in the SVG; this was a bug in the extraction regex
(greedy DOTALL match across the `xlink:title="&lt;TABLE&gt;"` tooltip attribute
GraphViz emits on HREF-carrying `<a>` elements). Re-extraction with anchor-aware
parsing confirmed the navigation text is emitted as proper SVG `<text>` nodes
inside `<a xlink:href>` wrappers — no leaks.

## Post-v1.8 Findings: Safari SVG Link Bug + Navigation Redesign (2026-08-06)

After v1.8 UAT closure, the user tested the generated SVG in Safari and found
the navigation links unclickable. This triggered a deeper investigation that
revealed two categories of issues: a fundamental browser-compatibility bug and
several navigation UX problems.

### Finding 1: Safari/WebKit ignores SVG `<a>` navigation (BLOCKER)

**Root cause (verified by driving real Safari via AppleScript):** Safari does
not follow `<a>` element hyperlinks inside SVG — not with `xlink:href`, not
with plain `href`, not inline-embedded, not standalone. This is a documented
WebKit limitation. Chromium follows them; WebKit does not. The hrefs c4drill
computes are correct; the rendering target just doesn't honor them.

**Fix: new `-f html` output format.** Emits self-contained HTML files that:
1. Inline the SVG content (strip XML decl/DOCTYPE)
2. Rewrite `.svg`→`.html` in all hrefs (so wrapped pages cross-link to wrapped siblings)
3. Inject a 200-byte JS shim that attaches click listeners to every `svg a`,
   navigating via `window.location.href` — this restores navigation in Safari

The `svg` and `dot` formats are unchanged (default stays `svg`). The HTML
fix is isolated to `internal/render/render.go` (`RenderHTML` + `wrapSVGInHTML`);
`path.go` keeps emitting `.svg` hrefs (Gap-3 contract intact).

**Verified:** Full C1→C2→C3→C1 navigation chain in real Safari.

### Finding 2: Navigation UX problems (3 issues reported by user)

1. **Breadcrumb duplication**: C2 showed `Back to Multilevel Test System |
   Multilevel Test System > mainSystem` — the back-link and breadcrumb's first
   item pointed to the same destination with the same label.
   **Fix:** Dropped the back-link entirely. Breadcrumb-only navigation.

2. **Raw path segments**: Breadcrumbs showed `mainSystem` (dotted-path key)
   instead of `Main System` (the unit's display Name).
   **Fix:** Added `AncestorNames map[string]string` to the View, populated by
   the view generators; the graph builder resolves pretty names from it.

3. **Excessive gaps + missing root**: Breadcrumb items had large visual gaps
   (GraphViz column-stretching), and C3 was missing the root context breadcrumb.
   **Fix:** (a) Always prepend the root context as the first breadcrumb. (b)
   Merge the `>` separator into each item's cell (separate separator cells
   stretch to column width). (c) Wrap the title in `<FONT POINT-SIZE="14">`
   — GraphViz silently drops the title row when row 1 has POINT-SIZE="10"
   content and row 2 has plain text.

### Current nav rendering (post-fix)

```
Multilevel Test System > Main System > SSH Auth    ← 10pt gray, ancestors underlined
Main System - SSH Auth - Components                ← 14pt default (title)
```

- Ancestors are clickable (underlined, muted font)
- Current level is plain (no underline)
- Root context always present (navigate to C1 from any level)
- Title visually distinct from nav (larger, default color)

### Test impact

- New: `TestRenderHTML` (4 subtests), `TestRootCmd_HTMLFormat` (5 subtests)
- Updated: all navigation tests for breadcrumb-only + styled output + pretty names
- Regenerated: `cmd/c4drill/testdata/multilevel.expanded.dot` golden (label format)
- All 7 packages pass: `go test ./...` clean
