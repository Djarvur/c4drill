---
phase: 38-hierarchy-wrapping-and-granular-keys
plan: 04
subsystem: cli-e2e-goldens
tags: [composition-matrix, golden-rebaseline, additive-goldens, e2e, visual-checkpoint]
requires:
  - phase: 38-01
    provides: "wrapper clusters (wrap_*) and EMPTY expected-failure census"
  - phase: 38-02
    provides: "granular switches --no-colors/--no-styles/--no-length/--no-rank with per-switch discriminators"
  - phase: 38-03
    provides: "--no-labels graph-layer suppression, census still EMPTY"
provides:
  - "TestKeyComposition E2E matrix: switch × generation × format × composition"
  - "Additive goldens testdata/nolabels.dot + nolabels.expanded.dot, canonically pinned"
  - "BC-01 verified: zero re-baseline needed (census-empty confirmed with evidence)"
affects: [38-05-docs]
tech-stack:
  added: []
  patterns:
    - "matrix table-driven E2E with per-fixture dot discriminators (raw-dot for structural markers, lowercased for hexes)"
    - "additive golden generation via the production CLI, pinned canonically (DI-1)"
key-files:
  created:
    - cmd/c4drill/testdata/nolabels.dot
    - cmd/c4drill/testdata/nolabels.expanded.dot
  modified:
    - cmd/c4drill/root_test.go
decisions:
  - "No re-baseline performed: 38-01/02/03 left every committed golden byte-identical (census EMPTY confirmed by green canonical golden tests + empty git diff on testdata)"
  - "Structural dot markers (<table, dir=back) asserted on RAW dot; only hex colours asserted lowercased — sanctioned graph/legend markup is UPPERCASE (38-03 lesson applied)"
  - "styles fixture (inline in TestKeyComposition) carries the --no-styles discriminators; plain.toml's author styles are not DOT-observable (38-02 finding)"
metrics:
  duration: ~35m
  completed: 2026-08-30
---

# Phase 38 Plan 04: Composition Matrix + Consolidated Re-baseline Summary

TestKeyComposition proves every granular switch and --no-labels across C1/drill-down/--expanded × dot/svg/html including pairwise compositions; BC-01's consolidated re-baseline verified as a NO-OP (zero committed-golden diffs — waves 1–3 changed only opt-in paths), with additive nolabels goldens committed and canonically pinned.

## Performance

- **Duration:** ~35 min
- **Started:** 2026-08-30
- **Completed:** 2026-08-30
- **Tasks:** 3 (Task 3 = visual checkpoint, auto-approved in auto-mode)
- **Files modified:** 3

## Accomplishments

- **KEY-03 (Task 1, commit 11b15e1):** `TestKeyComposition` — {--no-colors, --no-styles, --no-length, --no-rank, --no-labels} × {C1, drill-down, --expanded} × {dot, svg, html}. Dot cells assert each switch's suppression discriminator (hexes lowercased; structural markers on raw dot); svg/html cells assert generation + non-emptiness + one marker each (author hex absent under --no-colors). Pairwise compositions `--plain --no-labels` (all generations) and `--no-colors --no-labels` (plain + multilevel) included. Fixtures: plain.toml (flat styled), multilevel.toml, dedicated inline styles fixture.
- **BC-01 (Task 2, commit 719388e):** consolidated re-baseline executed as verification — result: **no re-baseline needed**. Evidence below. Additive goldens `testdata/nolabels.dot` + `nolabels.expanded.dot` generated from plain.toml via the production CLI and pinned by `TestNoLabelsC1Golden` / `TestNoLabelsExpandedGolden` (canonical, DI-1).
- **Full suite green:** `go test -count=1 ./...` — all packages ok.

## Golden Audit Table (BC-01, T-38-05 mitigation)

| Golden | Consumed by | Fresh-render diff vs committed | Attribute |
|--------|-------------|-------------------------------|-----------|
| testdata/plain.dot | TestPlainFlagC1Golden | none (canonical equal, test PASS) | — no delta |
| testdata/plain.expanded.dot | TestPlainFlagExpandedGolden | none (canonical equal, test PASS) | — no delta |
| testdata/multilevel.expanded.dot | expanded-graph baseline tests | suite green — no delta | — no delta |
| testdata/expanded.dot | expanded-graph baseline tests | suite green — no delta | — no delta |
| git diff -- cmd/c4drill/testdata/ | — | only NEW untracked nolabels.dot / nolabels.expanded.dot | additive only |

**Why the expected WRAP churn never materialized (waves 1–3 root cause, confirmed):** the committed goldens cover C1 and all-expanded generations only; WRAP (38-01) touches C2/C3 boundary-view dispatch, and the cmd E2E asserts those via Contains/NotContains, not byte goldens. The granular switches and --no-labels are opt-in — default-path output unchanged. Zero hunks required attribution; nothing re-baselined; the threat T-38-05 (re-baseline masking regressions) is vacuously mitigated.

## Task Commits

1. **Task 1: KEY-03 composition matrix E2E** - `11b15e1` (test)
2. **Task 2: BC-01 re-baseline verification + additive nolabels goldens** - `719388e` (feat)
3. **Task 3: visual approval checkpoint** - auto-approved, evidence below (no code change)

## Visual Checkpoint Evidence (Task 3 — ⚡ Auto-approved checkpoint (auto-mode))

Renders performed per the plan's how-to-verify, artifacts in /tmp:

1. **Wrapping legible** (`/tmp/p38-visual/multilevel.svg` + dot tree): C3 `mainSystem/storages/localStorage` view shows `cluster_wrap_mainSystem.storages` labelled **"Storages Registry"** (pretty AncestorNames label, no 🔍, no URL attribute) enclosing `cluster_mainSystem.storages.localStorage`; wrapper present in the rendered SVG ("Storages Registry" appears in SVG text). The 2 🔍 hits in that dot belong to legitimately collapsed non-wrapper units.
2. **Labels-off readable** (`/tmp/p38-plain/plain.svg`, --no-labels): 3.3 KB SVG, legend present (3 "legend" hits), zero element label text ("Order API", "order context", "gRPC" all absent); dot carries `label=""` and bare shapes — layout re-flowed without label geometry.
3. **Colour suppression** (`/tmp/p38-plain-nc/plain.svg`, --no-labels --no-colors): zero hits for author/kind hexes (#fff9c4, #f9a825, #1565C0, #2E7D32); monochrome; legend rows still present (kind text rows remain, colours gone).

## Deviations from Plan

**1. [Rule 1 - Test] Matrix discriminator case handling**
- **Found during:** Task 1
- **Issue:** first run failed — `<table` asserted against a lowercased dot also matched the sanctioned UPPERCASE graph/legend markup after ToLower (same pitfall 38-03 documented)
- **Fix:** hex markers compared lowercased; structural markers compared on the raw dot
- **Files modified:** cmd/c4drill/root_test.go
- **Committed in:** 11b15e1

**Total deviations:** 1 auto-fixed (Rule 1, test-side only).
**Impact on plan:** none on behaviour.

## TDD Gate Compliance

Plan 38-04 is an execute-type consolidation plan (matrix + goldens + checkpoint); its tasks carry no tdd="true" frontmatter, so no RED commit gate applies. The matrix test was written first-in-plan and committed before the goldens task.

## Known Stubs

None.

## Threat Flags

None — no new network endpoints, auth paths, or file access patterns. T-38-05 (golden tampering) mitigated: audit table above shows zero re-baselined hunks.

## Self-Check: PASSED

- Files exist: cmd/c4drill/testdata/nolabels.dot, cmd/c4drill/testdata/nolabels.expanded.dot, cmd/c4drill/root_test.go — FOUND
- Commits: 11b15e1, 719388e — FOUND in `git log`
- `go test -count=1 ./...` — all packages ok
- `git diff -- cmd/c4drill/testdata/` — empty (no committed golden modified)
