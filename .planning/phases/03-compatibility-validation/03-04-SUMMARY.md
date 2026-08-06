---
phase: 03-compatibility-validation
plan: 04
subsystem: testing
tags: [graphviz, html-labels, svg-navigation, relative-paths, tdd, gap-closure]

# Dependency graph
requires:
  - phase: 03-compatibility-validation (plans 01, 02)
    provides: C1/C2/C3 view generation, auto-detected sub-diagrams, canonicalDOT golden baseline
provides:
  - Resolved C2/C3 explore hrefs (ancestor, sibling, descendant) with self-link guard
  - Clickable SVG/DOT navigation bar rendered via GraphViz HTML <TD HREF> cells
  - Uniform .svg navigation URLs across render formats
  - End-to-end CLI regression enforcing the COMPAT-02 navigation contract at the SVG level
affects: [compatibility-validation, future render-format work, browser navigation]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "GraphViz clickable HTML labels via <TD HREF=url> (NOT <a href>, which GraphViz drops)"
    - "Bidirectional relative-path computation for full dotted-path targets"
    - "Force .svg in navigation URLs regardless of render format (parameter kept, value ignored)"

key-files:
  created:
    - cmd/c4drill/root_test.go (TestCompat02_NavigationLinksResolve + helpers)
  modified:
    - internal/graph/path.go
    - internal/graph/path_test.go
    - internal/render/converter.go
    - internal/render/navigation.go
    - internal/render/navigation_test.go
    - internal/render/integration_test.go
    - cmd/c4drill/testdata/multilevel.expanded.dot (regenerated golden)

key-decisions:
  - "GraphViz HTML-like labels do NOT support <a href> tags (verified empirically: labels containing them are silently dropped at render time). Clickable links in HTML labels are expressed via the HREF attribute on a <TD> element, which GraphViz renders as <a xlink:href> in SVG. The plan's <a href> + StrdupHTML premise was unachievable; implemented the TD HREF idiom instead."
  - "Regenerated the canonicalDOT golden (multilevel.expanded.dot): the graph label is now always HTML-wrapped (label=<...> instead of label=\"...\"), a benign format change. URLs/edges/clusters unaffected; the expanded diagram carries no navigation."
  - "Fixed a pre-existing C3 breadcrumb resolution bug (Rule 2) in computeBreadcrumbURL: the index-0 special case assumed the first segment was the C1 root, so the 'mainSystem' crumb in a C3 diagram 404'd on ../multilevel.svg. Now delegates to ComputeExploreURL's bidirectional logic so it resolves to ../mainSystem.svg."

patterns-established:
  - "Clickable HTML labels: use <TABLE><TR><TD HREF=url>text</TD></TR></TABLE> through cg.StrdupHTML; never embed <a href>."
  - "HTML-escape all plain text before embedding in a GraphViz HTML label (titles, names) to prevent label breakage/injection."

requirements-completed: [COMPAT-02]

# Metrics
duration: 19min
completed: 2026-08-06
---

# Phase 03 Plan 04: C2/C3 Navigation Gap Closure Summary

**Closed the three diagnosed UAT gaps that broke C2/C3 navigation: bidirectional explore URLs with a self-link guard, a clickable SVG navigation bar via GraphViz `<TD HREF>` HTML labels, and uniform `.svg` nav URLs across render formats — verified by a new end-to-end CLI regression (77/77 hrefs resolve).**

## Performance

- **Duration:** ~19 min
- **Started:** 2026-08-06T13:09:12Z
- **Completed:** 2026-08-06T13:28:18Z
- **Tasks:** 3 (TDD: Task 1 and Task 2 RED→GREEN; Task 3 end-to-end)
- **Files modified:** 8

## Accomplishments
- **Gap 1 (blocker) closed**: `ComputeExploreURL` rewritten with a self-link guard (`targetPath == currentPath` → `""`) and a bidirectional relative-path algorithm; ancestor targets now yield `../mainSystem.svg` (and `../../` for deeper ancestors), descendant/sibling resolution unchanged. The C3 collapsed-ancestor node no longer emits the empty `href=".svg"`.
- **Gap 2 (blocker) closed**: navigation bar renders as a visible, clickable bar in SVG via `<TABLE><TD HREF=url>` HTML labels through `cg.StrdupHTML`; GraphViz emits `<a xlink:href>` anchors. Pre-fix output showed escaped `&lt;a href=&quot;...&quot;&gt;` literal text.
- **Gap 3 (major) closed**: `ComputeBackLinkURL` and `computeBreadcrumbURL` ignore the `format` parameter and always emit `.svg`; `.dot` diagrams now link to browser-navigable `.svg` siblings.
- **End-to-end regression**: `TestCompat02_NavigationLinksResolve` runs the multilevel fixture through the CLI and asserts all three gaps at the SVG level (the canonicalDOT golden normalizes URLs away, so it cannot enforce this contract). 77/77 hrefs across all 17 SVGs resolve; zero empty `href=".svg"`.

## Task Commits

Each task was committed atomically (TDD: test RED → feat GREEN):

1. **Task 1 RED** — `42afc1f` (test): failing ComputeExploreURL ancestor/self-link subtests
2. **Task 1 GREEN** — `c69c385` (feat): bidirectional ComputeExploreURL + self-link guard
3. **Task 2 RED** — `d96348d` (test): failing nav-.svg and clickable-SVG-label tests
4. **Task 2 GREEN** — `d1a62e1` (feat): clickable nav via TD HREF + force .svg + breadcrumb fix + golden regen
5. **Task 3** — `1bc6bcd` (test): end-to-end CLI regression TestCompat02_NavigationLinksResolve

**Plan metadata:** (pending final docs commit below)

## Files Created/Modified
- `internal/graph/path.go` — ComputeExploreURL bidirectional rewrite + self-link guard; ComputeBackLinkURL/computeBreadcrumbURL force .svg and delegate to ComputeExploreURL; added encodePathSegments + commonDirectoryPrefixLength helpers
- `internal/graph/path_test.go` — RED subtests for ancestor/self-link/2-levels-up + file-resolution regressions + Gap 3 .svg-forcing subtests
- `internal/render/navigation.go` — BuildNavigationLabel emits `<TABLE><TD HREF>` HTML; extracted navigationTDs/breadcrumbTDs/breadcrumbItemTD/plainTD helpers; names HTML-escaped
- `internal/render/converter.go` — configureGraphSettings builds an HTML graph label (nav TABLE row + HTML-escaped title row) via cg.StrdupHTML; removed unused joinLabels; added `html` import
- `internal/render/navigation_test.go` — updated TestBuildNavigationLabel assertions to the TD HREF contract; TestNavigationInGraphOutput stays green
- `internal/render/integration_test.go` — new TestIntegration_Navigation_SVGRendersClickableLinks (SVG `<a xlink:href>` + DOT `<TD HREF>`)
- `cmd/c4drill/root_test.go` — new TestCompat02_NavigationLinksResolve + generateMultilevelOutput/readOutputFile/collectNavAndExploreHrefs/assertHrefsResolve helpers
- `cmd/c4drill/testdata/multilevel.expanded.dot` — regenerated golden (graph label now HTML form; semantic content unchanged)

## Decisions Made
- **GraphViz HTML labels reject `<a href>`** — verified empirically that any HTML label containing an `<a href>` tag is silently dropped at render time (SVG shrinks to the no-label form). The correct idiom for clickable HTML labels is `<TD HREF="url">text</TD>`, rendered as `<a xlink:href="...">` in SVG. This contradicts the plan's literal `<a href>` + StrdupHTML premise; implemented the achievable idiom and adjusted test assertions accordingly (anchored on the plan's <behavior> intent: visible + clickable + unescaped).
- **Golden regeneration is benign** — the expanded diagram carries no navigation; only its title label form changed (`label="X"` → `label=<TABLE>...X...</TABLE>`). URLs, edges, clusters, and penwidths are unaffected. Regenerated via the documented `--expanded` command rather than letting the golden test fail.
- **Pre-existing C3 breadcrumb bug fixed under Rule 2** — `computeBreadcrumbURL`'s index-0 special case assumed the first path segment was the C1 root, so the "mainSystem" crumb in a C3 `mainSystem.sshAuth` diagram pointed at `../multilevel.svg` (404). It is the same navigation contract this plan closes, so it was fixed by delegating to the new bidirectional ComputeExploreURL logic.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] GraphViz drops HTML labels containing `<a href>` tags**
- **Found during:** Task 2 (GREEN implementation of the StrdupHTML wrap)
- **Issue:** The plan's prescribed fix (wrap navigation HTML containing `<a href>` via `cg.StrdupHTML`, mirroring node labels) caused GraphViz to silently drop the entire graph label during SVG rendering — the navigation bar disappeared. Verified that GraphViz HTML-like labels do not support `<a>` tags at all; node labels work only because they use `<TABLE>` (no anchors). Clickability for nodes comes from the `URL` attribute, not HTML anchors.
- **Fix:** Re-expressed the navigation bar as a GraphViz HTML `<TABLE BORDER="0">` with clickable `<TD HREF="url">` cells (rendered as `<a xlink:href>` in SVG) and plain `<TD>` cells for separators/current level. HTML-escaped all plain text (names, title) before embedding.
- **Files modified:** internal/render/navigation.go, internal/render/converter.go, internal/render/navigation_test.go, internal/render/integration_test.go
- **Verification:** `TestIntegration_Navigation_SVGRendersClickableLinks` asserts `<a xlink:href=` present and `&lt;a href=` absent; `TestNavigationInGraphOutput` ("Back to Main System") stays green; CLI output shows 6 clickable anchors per C2/C3 SVG.
- **Committed in:** d1a62e1

**2. [Rule 2 - Missing Critical] Pre-existing C3 breadcrumb resolution 404**
- **Found during:** Task 2 (end-to-end href-resolution check during GREEN verification)
- **Issue:** `computeBreadcrumbURL` assumed the first path segment was always the C1 root (`ancestorIndex == 0` → basename URL). For a C3 path like `mainSystem.sshAuth`, the "mainSystem" crumb pointed at `../multilevel.svg`, which resolves to `multilevel/multilevel.svg` (non-existent). This is the same navigation contract the plan closes.
- **Fix:** `computeBreadcrumbURL` now reconstructs the ancestor's dotted path and delegates to `ComputeExploreURL`'s bidirectional relative-path logic, so the crumb resolves to `../mainSystem.svg`.
- **Files modified:** internal/graph/path.go
- **Verification:** C3 `sshAuth.svg` "mainSystem" crumb now resolves; all 77 hrefs across the generated tree resolve via os.Stat join.
- **Committed in:** d1a62e1

**3. [Rule 3 - Blocking] canonicalDOT golden needed regeneration**
- **Found during:** Task 2 (after the graph-label HTML wrap)
- **Issue:** `TestBuildExpandedGraphBaselineDOT` failed because the graph label form changed from plain text (`label="Multilevel Test System"`) to HTML (`label=<TABLE>...`), which is an expected consequence of the Gap 2 fix. URLs/edges/clusters were unaffected (the expanded diagram has no navigation).
- **Fix:** Regenerated `cmd/c4drill/testdata/multilevel.expanded.dot` via the documented `go run ./cmd/c4drill cmd/c4drill/testdata/multilevel.toml --format dot --expanded --output <dir>` command. Confirmed the canonical diff is label-form-only.
- **Files modified:** cmd/c4drill/testdata/multilevel.expanded.dot
- **Verification:** `TestBuildExpandedGraphBaselineDOT` green; canonical comparison of URLs/edges/clusters unchanged.
- **Committed in:** d1a62e1

---

**Total deviations:** 3 auto-fixed (1 bug, 1 missing critical, 1 blocking)
**Impact on plan:** All deviations directly serve the plan's stated goal ("Make C2/C3 sub-diagrams actually navigable"). The GraphViz-idiom correction was forced by a hard engine constraint discovered during implementation, not scope creep. No user decision needed — there is no alternative way to render clickable HTML labels in GraphViz.

## Issues Encountered
- Initial GraphViz HTML-label experiments: confirmed via isolated `StrdupHTML` + `Render` probes that `<a href>` (bare, in `<FONT>`, or in `<TABLE><TD>`) is always dropped, while `<TD HREF="url">` renders correctly. This isolated-test step prevented a wrong implementation from shipping.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- All 9 UAT scenarios now pass (Gap 1, 2, 3 closed; the lone `issue` in 03-UAT.md is resolved).
- COMPAT-02 navigation contract is now enforced semantically (SVG-level) by `TestCompat02_NavigationLinksResolve`, complementing the order-insensitive canonicalDOT golden.
- `go test -race ./...` clean; `golangci-lint run --new-from-rev=511c700 ./...` 0 issues.
- No blockers. Phase 03 gap-closure complete.

## TDD Gate Compliance
- Task 1: `test(03-04)` RED (`42afc1f`) → `feat(03-04)` GREEN (`c69c385`). RED failed as expected (4 new subtests failed; 2 file-resolution regressions pre-passed, correctly reflecting today's descendant/sibling behavior).
- Task 2: `test(03-04)` RED (`d96348d`) → `feat(03-04)` GREEN (`d1a62e1`). RED failed as expected (3 nav-URL and SVG-label subtests failed). The GREEN commit's assertion-contract updates to the `<TD HREF>`/`xlink:href` idiom are part of the deviation documented above (GraphViz rejects `<a href>` in HTML labels).
- Task 3: end-to-end regression (no TDD gate; `test` commit `1bc6bcd`).
- All three gates (test → feat) present in git log.

---
*Phase: 03-compatibility-validation*
*Completed: 2026-08-06*

## Self-Check: PASSED

All 9 created/modified files exist; all 5 task commits (42afc1f, c69c385, d96348d, d1a62e1, 1bc6bcd) present in git log.
