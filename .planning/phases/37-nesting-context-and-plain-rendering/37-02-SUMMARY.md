---
phase: 37-nesting-context-and-plain-rendering
plan: 02
subsystem: view
tags: [view, scope, deep-links, go, tdd]

# Dependency graph
requires:
  - phase: 37-nesting-context-and-plain-rendering
    plan: 01
    provides: recursive buildCluster/buildNestedCluster — the unfolding mechanism chain-bearing ancestors render through
provides:
  - "CTX-02: deep-link targets keep their container chain — resolveBoundaryLink (C1) and addResolvedCrossLink/From (C2/C3) resolve such links to the TRUE target via ensureDeepLinkChain"
  - "Entry.UnfoldChain flag + builder dispatch guards in buildC1ViewGraph AND buildBoundaryViewGraph — collapsed ancestors with inserted chains render as recursive clusters"
  - "VisiblePaths skip in buildBoundaryViewGraph — C2/C3 chain entries render only via ancestor-cluster recursion, never as duplicate standalone nodes"
  - "insertAdjacentToAncestor — chain entries land adjacent to their depicted ancestor in UnitOrder (deterministic model order, not link-scan order)"
  - "Tests: TestGenerateC1View_DeepLinkTargetKeepsContainerChain, TestGenerateC1View_ExternalLinkNoChain (boundary lock), TestGenerateC2View_DeepCrossLinkTargetChain, TestBuildGraph_DeepLinkTargetRendersInsideUnfoldedCluster"
affects: [37-05 (golden re-baseline scope + CTX-01 proof), 37-07 (docs: deep-link behavior)]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Triple-registration chain insertion (Units + UnitOrder + VisiblePaths) reusing the visible-subunit contract for a second purpose"
    - "Adjacent-to-ancestor UnitOrder insertion (insertAdjacentToAncestor) — last-descendant scan keeps chains grouped under their depicted ancestor"
    - "In-scope ancestor guard: chains unfold only under depicted ancestors that are neither IsBoundary nor IsExternal"

key-files:
  created: []
  modified:
    - internal/view/view.go
    - internal/view/scope.go
    - internal/view/scope_test.go
    - internal/view/integration_test.go
    - internal/graph/builder.go
    - internal/graph/builder_test.go
    - cmd/c4drill/root_test.go

key-decisions:
  - "Chain walk fires only for peers that findUnitByPath resolves AND whose depicted ancestor is in-scope (not IsBoundary/IsExternal) — junk paths and external/sibling boundary nodes fall through to today's collapsed resolution (T-37-03 mitigation, boundary lock)"
  - "D-13 minlen rule untouched: chain-resolved links keep Length=0 (link.Peer != resolved at comparison time), keeping TestBuildGraphResolvedEdgeMinLen semantics"
  - "ensureDeepLinkChain is called only on links actually recorded (after the resolved==resolvedSource drop) so dropped internal links never create chains"
  - "UnfoldChain kept deliberately separate from IsExpanded: no author-intent semantics, D-07 guard, or 🔍 logic disturbed — unfolded collapsed containers still carry 🔍 + ExploreURL like plan 37-01 nested container clusters"

patterns-established:
  - "CTX-02 chain contract: depicted link target ⇒ ancestor chain (depicted ancestor → target) as VisiblePaths entries + edge to TRUE target, in C1 AND C2/C3"

requirements-completed: [CTX-02]

# Metrics
duration: 32min
completed: 2026-08-30
---

# Phase 37 Plan 02: Nesting Context — Deep-Link Ancestor Chains (CTX-02) Summary

**Deep-link targets keep their container context: `ensureDeepLinkChain` inserts the target's ancestor chain (depicted ancestor → true target) as VisiblePaths entries in C1 and C2/C3, edges terminate at the TRUE target, and `Entry.UnfoldChain` makes both builders render the chain-bearing collapsed ancestor as a recursive cluster — no anonymous top-level collapse, no graphviz implicit nodes.**

## Performance

- **Duration:** 32 min
- **Started:** 2026-08-30T12:48:34Z
- **Completed:** 2026-08-30T13:20:49Z
- **Tasks:** 3 (RED / GREEN / VERIFY)
- **Files modified:** 7

## Accomplishments
- CTX-02: a C1/C2/C3 link whose true target lies under a depicted ancestor but is not itself depicted now adds the target's ancestor chain as visible entries and resolves the edge to the TRUE target (was: silent collapse to the nearest visible ancestor)
- Chain entries follow the triple-registration contract (Units + UnitOrder + VisiblePaths) so BOTH builders skip them as standalone nodes — they render only inside the ancestor-cluster recursion (plan 37-01's recursive buildCluster/buildNestedCluster)
- `Entry.UnfoldChain` (view.go) + dispatch guards `(IsExpanded || UnfoldChain) && len(Subunits) > 0` in `buildC1ViewGraph` and `buildBoundaryViewGraph`; the boundary builder also gained the VisiblePaths skip (C2/C3 had none — chain entries would have double-rendered)
- External/sibling boundary behavior locked unchanged by test (`TestGenerateC1View_ExternalLinkNoChain`): synthesized boundary nodes, no chains, no UnfoldChain
- Ordering: chain entries insert adjacent to their depicted ancestor (`insertAdjacentToAncestor`), not at the UnitOrder tail in link-scan order
- TDD gates: RED commit `6f503cd` (3 failing chain tests + 1 passing behavior lock), GREEN commit `c065c07`

## Verification (Task 3) — observations for plan 37-05

- `go test ./internal/render/` PASS; `go build ./...` OK; CLI run on multilevel.toml clean; **full repo suite green (15/15 packages, uncached), INCLUDING both committed-golden baselines** — zero committed-golden debt after this plan, same as plan 37-01. Expanded-mode goldens are untouched (GenerateExpandedView path does not route through resolveBoundaryLink/buildCluster dispatch). **37-05's re-baseline scope is likely again near-empty; verify flat-model stability (BC-01) instead.**
- Structural assertions on the CLI C1 output (mechanical check, all PASS):
  - (i) all 12 deep-link target identifiers in multilevel.toml appear BOTH as node statements AND as edge endpoints (46 unique nodes, 56 edges, no dangling endpoints — graphviz materialized no implicit nodes)
  - (ii) chain-bearing ancestors render as `subgraph "cluster_…"` (17 clusters total: mainSystem + sshAuth/localIDP/storages/authModules and their nested containers), never as bare top-level nodes
  - (iii) zero duplicate node declarations — the VisiblePaths contract prevents top-level/chain double-rendering
- C2 output (`multilevel/mainSystem.dot`) passes the same assertions (46 nodes, 37 edges, 17 clusters; storages/authModules/localIDP/sshAuth unfold inside the boundary cluster; externalSys stays a plain boundary node)
- Expected C1/C2 delta vs pre-CTX-02 (BY DESIGN, for 37-05's records): deep units' own authored links now render at their true endpoints inside unfolded containers (e.g. `sessionManager → sessionDb`, `grpcAPIs.authAPI → rbac`), and collapsed containerBox/container ancestors reached by deep links unfold with 🔍 labels (`SSH Auth 🔍`, `Storages Registry 🔍`)

## Task Commits

1. **Task 1 (RED): deep-link chain tests** - `6f503cd` (test) — plus inert `Entry.UnfoldChain` field so the tests compile (deviation 1, 37-01 precedent)
2. **Task 2 (GREEN): ancestor-chain insertion + UnfoldChain dispatch** - `c065c07` (feat)
3. **Task 3 (VERIFY): cross-package suite + multilevel fixture** - no commit (verification only, no source edits)

_Note: TDD plan — test → feat commits; no refactor commit needed (implementation landed clean, golangci-lint 0 issues on changed packages)._

## Files Created/Modified
- `internal/view/view.go` — `Entry.UnfoldChain bool` (inert in RED, documented, set by the generator, read by the builder dispatch)
- `internal/view/scope.go` — `ensureDeepLinkChain` (chain registration + UnfoldChain marking + in-scope guard), `insertAdjacentToAncestor`, integration into `resolveBoundaryLink` (after the internal-link drop, before return) and into `addResolvedCrossLink`/`addResolvedCrossLinkFrom` (m threaded through `resolveSubunitCrossLinks`/`resolveDescendantCrossLinks` — signature change, unexported)
- `internal/graph/builder.go` — UnfoldChain dispatch in `buildC1ViewGraph` + `buildBoundaryViewGraph`; VisiblePaths skip in `buildBoundaryViewGraph`
- `internal/view/scope_test.go` — 3 new CTX-02 tests + `deepLinkChainModel` fixture; D-07-era assertions updated to true-target contract (ResolvesToVisibleSubunit, NoRedundantParentEdge, BoxResolutionParity)
- `internal/view/integration_test.go` — pollution test rewritten to the CTX-02 contract (chain-only depiction, explicit definition order); `TestIntegrationC1EdgeResolution` and `TestBuildGraphExpandedC1VisibleSubunitEdges` updated to true-target assertions
- `internal/graph/builder_test.go` — `TestBuildGraph_DeepLinkTargetRendersInsideUnfoldedCluster` (graph-layer unfolded-cluster proof)
- `cmd/c4drill/root_test.go` — `TestCompat01_ValidTomlAllCollapsed` updated: valid.toml's `user → app.api` deep link now unfolds app as a cluster with the edge terminating at `app.api`

## Decisions Made
- Chains unfold only beneath in-scope depicted ancestors — `depicted.IsBoundary || depicted.IsExternal ⇒ fall through` — so links to external top-level peers and sibling boundary nodes keep today's collapsed boundary-node behavior exactly (plan scope guard)
- Adjacent-to-ancestor insertion (plan's option A) implemented via last-descendant scan so successive chains under one ancestor group after previously inserted ones; determinism for parsed models comes from the parser always producing m.UnitOrder + SubunitOrder
- Transitive depiction preserved: once a unit becomes depicted (chain or visible subunit), ITS OWN links resolve at their true endpoints (depicted source ⇒ source entry carries the resolved link). In C1 of multilevel this cascades along link paths — pinned deterministically in the rewritten pollution test (13 entries: 5 top-level + 8 chain, incl. transitive `nss`)

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Added inert `Entry.UnfoldChain` field in the RED commit**
- **Found during:** Task 1 (RED)
- **Issue:** Tests 1 and 3 assert `UnfoldChain == true`, but the field was scheduled for Task 2 — the RED tests could not compile, contradicting Task 1's done criterion ("tests compile and FAIL")
- **Fix:** Added the field (with final doc comment) in the RED commit — inert skeleton, no writers/readers until GREEN; exact precedent is 37-01's deviation #1
- **Files modified:** internal/view/view.go
- **Verification:** RED tests compiled and failed on the collapse assertions, not compilation
- **Committed in:** 6f503cd

**2. [Rule 1 - Bug] Updated 7 existing tests that pinned the old collapse behavior (plan anticipated only golden-baseline debt)**
- **Found during:** Task 2 (GREEN)
- **Issue:** The plan's GREEN expectation was "existing tests stay green EXCEPT the documented committed-golden baselines", but seven content tests pin the pre-CTX-02 collapse for link targets: TestGenerateC1View_ResolvesToVisibleSubunit, TestGenerateC1View_NoRedundantParentEdge, TestGenerateC1View_BoxResolutionParity (scope_test.go), TestIntegrationC1ViewNoNestedBoundaryPollution, TestIntegrationC1EdgeResolution, TestBuildGraphExpandedC1VisibleSubunitEdges (integration_test.go), TestCompat01_ValidTomlAllCollapsed (root_test.go). CTX-02 deliberately reverses exactly the behavior they pin (the objective states: "The collapse was a deliberate anti-pollution choice — CTX-02 reverses it for link TARGETS only")
- **Fix:** Updated each to the CTX-02 contract (true-target peers, chain-entry expectations, unfolded-cluster assertions), following 37-01's precedent for old-structure assertions; no assertion weakened — each now pins the stronger new contract
- **Files modified:** internal/view/scope_test.go, internal/view/integration_test.go, cmd/c4drill/root_test.go
- **Verification:** full internal/view package + cmd package green
- **Committed in:** c065c07

**3. [Rule 1 - Bug] Fixed a map-order flake in the pollution integration fixture (transitive chain depiction)**
- **Found during:** Task 2 (GREEN) — caught by a one-time failure while re-running the suite
- **Issue:** The hand-built fixture lacked UnitOrder/SubunitOrder, so link-scan order was map-random; whether sshd became depicted BEFORE its own links were scanned (making its nss link resolve transitively and add nss to the chain) depended on map iteration — 12 vs 13 entries nondeterministically. Production is deterministic (the parser always captures definition order), so this was a fixture defect
- **Fix:** Gave the fixture explicit UnitOrder + SubunitOrder mirroring TOML definition order and pinned the deterministic result (13 entries incl. transitive nss, sshd → nss recorded at true endpoints); view package re-run 5x + 3 full-suite stress runs clean
- **Files modified:** internal/view/integration_test.go
- **Committed in:** c065c07

---

**Total deviations:** 3 auto-fixed (1 blocking compile gate, 2 behavior-preservation updates). **Impact on plan:** RED→GREEN semantics preserved; no scope creep — all changes are inside the plan's target files.

## Issues Encountered
- One unreproduced full-suite anomaly: a single `go test ./...` run reported a package FAIL between two clean runs; 6 subsequent full/partial runs (incl. 3x stress on view/graph/render/cmd) were fully green, and no failure output was captured. Most likely the pre-fix view flake's last gasp or an environment hiccup (graphviz WASM under load). If it reappears in 37-05, re-check the pollution fixture first.
- Pre-existing, out-of-scope (untouched per scope boundary): gofmt comment-formatting drift in `internal/c4d/composition_test.go`, `internal/c4d/grammar/reserved.go`, `internal/graph/integration_test.go`, `internal/graph/shapes.go`, `internal/parser/inference_test.go`, `internal/template/expand.go` (none of this plan's files); untracked `.planning/.../37-PATTERNS.md` and phase-36 planning-file deletions predate this execution (orchestrator state, left alone).

## TDD Gate Compliance
- RED gate: `6f503cd` `test(37-02)` — 3 chain tests failed for the collapse reason (view AND graph layers), external lock passed
- GREEN gate: `c065c07` `feat(37-02)` — all four Task 1 tests pass; full repo suite green
- REFACTOR gate: not needed (no cleanup beyond lint-driven signature wrapping, included in GREEN)

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- CTX-02 mechanism complete and regression-guarded at view AND graph layers; CTX-01 (every depicted element sits inside its chain) can be proven in 37-05 against this implementation
- Zero committed-golden debt after GREEN (baselines stayed green; see Verification observations) — 37-05's re-baseline task should re-scan and likely reduce to flat-model stability (BC-01) verification
- Full repo suite green (uncached, multiple runs): 15/15 packages; `go build ./...` clean; golangci-lint 0 issues on changed packages; `cmd/c4drill/testdata/` untouched

---

## Self-Check: PASSED

- All 7 modified files exist on disk (verified with `[ -f ]`)
- Commits verified in `git log`: `6f503cd` (test), `c065c07` (feat)
- Plan-level `<verification>` re-run: `go test ./internal/view/ -count=1` PASS; `go test ./internal/render/` PASS; `go build ./...` OK; CLI on multilevel.toml clean; `subgraph cluster_` count = 2 unquoted + 15 quoted nested = 17 total (≥1 required); structural assertions (i)(ii)(iii) all PASS on C1 and C2 outputs; no testdata/golden files modified (`git status` clean of testdata paths)
- No file deletions in either commit (`git diff --diff-filter=D` empty)

---
*Phase: 37-nesting-context-and-plain-rendering*
*Completed: 2026-08-30*
