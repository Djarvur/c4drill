# Deferred Items — Phase 03 (compatibility-validation)

Out-of-scope discoveries from plan 03-01 execution. Per executor scope boundary: NOT fixed here; logged for the manager/planner.

## DI-1: Render pipeline output is order-nondeterministic (byte-unstable XDOT)

- **Found during:** 03-01 Task 2 (golden baseline determinism check)
- **Evidence:** 5 pipeline runs (`go run ./cmd/c4drill <fixture> --format dot --expanded`) — every run pairwise differs. Sibling cluster order under a parent flips (e.g. `systemd` vs `pam` under `sshAuth`), node order within a cluster flips, and layout geometry (`bb`, `pos`, `lp`) is order-dependent. Verified identically with the private `cyp-auth-infra/saira-20260320.c2.full.toml` (2/2 runs differ) — pre-existing, NOT caused by the new fixture. The private 1041-line baseline was generated once; its "determinism" claim was never empirically verified.
- **Root cause:** map-order iteration inside the pinned external fork `github.com/onokonem/go-graphviz v0.0.0-20260321130544-f364b5235161` (cgraph subgraph/node storage + gvc WASM layout engine), i.e. `go.mod` `replace` target. c4drill's own code (builder/converter) is slice/definition-order deterministic.
- **What IS stable:** semantic content — node/edge sets, `key=` dedup attributes, labels, `minlen`/`penwidth` values, cluster structure (the D-02 contract). Verified: minlen=3 x1, penwidth=2 x56, minlen=2 x4, deep node present, zero `[+]` in every run.
- **Impact on 03-02 (BLOCKING as designed):** 03-02-PLAN.md lines 103-105 plan `require.Equal(t, string(expected), string(dotData))` byte-equality against `cmd/c4drill/testdata/multilevel.expanded.dot`. This WILL fail. 03-02 must use an order-insensitive comparison instead (per RESEARCH.md Pitfall 1: "sort-normalize before compare"): e.g. sort-normalize DOT block-level lines on both sides, or strip layout geometry (`pos=`, `bb=`, `lp=`, `lheight=`, `lwidth=`, `height=`, `width=`) and compare semantic tokens, or assert structurally (node set, edge set with attributes, cluster containment) instead of string equality.
- **Deferred fix options (Rule 4 — user decision required in a future phase, if byte-equality is ever needed):** (a) fork+fix the pinned go-graphviz fork to preserve cgraph insertion order; (b) repin to a different go-graphviz version/fork; (c) drop byte-equality entirely and rely on semantic comparison (recommended, matches D-02).

## DEFERRED-04: C3 boundary node resolves to parent instead of sibling container

**Discovered:** 2026-08-06 (post-v1.8, during nav-redesign visual review)
**Severity:** minor (cosmetic — diagram shows a confusing boundary node)
**Package:** `internal/view/scope.go`

### Symptom
C3 diagrams (e.g., `mainSystem.localIDP`) show **`Main System [+]`** as a
boundary node when a component links to a sibling container (e.g.,
`localIDP.grpcAPIs.authAPI` → `mainSystem.rbac`). The boundary node should be
the sibling container (`RBAC`), not the parent system (`Main System`).

### Root cause
`addResolvedBoundaryNode` (scope.go:625) walks the peer path up to find the
nearest visible ancestor. For peer `mainSystem.rbac` in a C3 view of
`mainSystem.localIDP`:
1. `mainSystem.rbac` — not in view → strip to `mainSystem`
2. `mainSystem` — not in view (only `mainSystem.localIDP.*` entries exist) →
   idx=0, break
3. Creates boundary node for `mainSystem` (the parent) — **wrong**

The walk-up should stop at the **sibling level** (path sharing the same parent
as the expanded container). `mainSystem.rbac` shares parent `mainSystem` with
`mainSystem.localIDP`, so `rbac` is the correct boundary node.

### Fix direction
`addResolvedBoundaryNode` needs to know the **container scope** (the C3 view's
expanded container path, e.g., `mainSystem.localIDP`). When walking up the peer
path, stop before crossing above the container's parent — i.e., a peer at the
sibling level (`mainSystem.rbac`) is the boundary, not its parent
(`mainSystem`).

### Test coverage gap
No C3 boundary-resolution test exists. The existing boundary tests cover C1
only (TestC1DeepestVisibleAncestor). When fixing, add a C3 test asserting that
a link to a sibling container produces the sibling as the boundary node, not
the parent system.

### Not blocking
This is cosmetic and pre-existing (predates the nav redesign and html format
work). The nav/html work is correct and should not be blocked on this fix.
