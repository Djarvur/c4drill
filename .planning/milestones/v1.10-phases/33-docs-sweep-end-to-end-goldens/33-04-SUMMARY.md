---
phase: 33-docs-sweep-end-to-end-goldens
plan: 04
subsystem: testing
tags: [gotest, e2e, golden, xc-01, xc-05, pipeline-ordering]

# Dependency graph
requires:
  - phase: 33-docs-sweep-end-to-end-goldens
    provides: "Plan 33-01 canonical.Canonical helper; Plan 33-02 composed XC-05 fixtures"
  - phase: 31-template-expansion
    provides: "internal/template.Expand (the 2nd pipeline pass)"
  - phase: 32-include-directive-multi-file
    provides: "internal/include.Resolve 3-arg signature (entry, entryDir, entryFile)"
  - phase: 30-relative-peer-resolution
    provides: "internal/peer.Resolve (the 3rd pipeline pass)"
provides:
  - "TestXC05_ComposedEquivSingleFile — composed multi-file ≡ single-file canonical equality (XC-05)"
  - "TestXC01_PipelineOrdering — behavioral guard that pipeline order is load-bearing (XC-01, XC-02, XC-03)"
affects: []

# Tech tracking
tech-stack:
  added: []
  patterns: ["Behavioral pipeline-ordering test (D-20): assert order-dependent rendered output, NOT source layout — robust to refactors that move pipeline calls into helper functions"]

key-files:
  created: []
  modified:
    - cmd/c4drill/root_test.go

key-decisions:
  - "Open Question A2 resolved: humanize stays at parse-time (parser.go:614, Phase 29 stopgap). Phase 31's XC-04 relocation to a post-expansion pass was DEFERRED. The composed fixtures carry explicit name= so parse-time humanize does not fire for them; no explicit humanize call needed in renderThroughPipeline. Not a deviation — consistent with the task brief's CONCURRENCY NOTE."
  - "include.Resolve 3-arg signature (Resolve(entry, entryDir, entryFile)) honored per Phase 32. The third arg threads the real entry filename so missing-include errors name the including file (INC-10/D-12)."
  - "XC-01 assertion uses the canonical edge substring '\"platform.auth.cache\" -> messageBus' — the canonical helper's serializeDOTStatement emits edges with the head in that exact form, so the substring is stable."

patterns-established:
  - "renderThroughPipeline(t, path) helper in cmd/c4drill/root_test.go — the canonical way to exercise the full v1.10 pipeline on a fixture and get canonical DOT for assertion. Reusable by future E2E goldens."
  - "Behavioral ordering test over source-scan (D-20) — detects pipeline reordering as a regression without fragility to legitimate refactors"

requirements-completed:
  - XC-01
  - XC-05

# Metrics
duration: 10 min
completed: 2026-08-08
---

# Phase 33 Plan 04: XC-05 Golden + XC-01 Behavioral E2E Tests Summary

**Two end-to-end tests in cmd/c4drill/root_test.go proving the four v1.10 features compose correctly through the full pipeline: XC-05 (composed multi-file ≡ hand-expanded single-file, canonicalDOT) and XC-01 (behavioral proof that include → template.Expand → peer.Resolve ordering is load-bearing, covering XC-02 and XC-03).**

## Performance

- **Duration:** ~10 min
- **Started:** 2026-08-08 (inline execution)
- **Completed:** 2026-08-08
- **Tasks:** 3 (2 implementation + 1 verification-only)
- **Files modified:** 1 (cmd/c4drill/root_test.go, +118 lines pure additions)

## Accomplishments
- `TestXC05_ComposedEquivSingleFile` renders `skill/examples/09-composed/entry.toml` AND `single-file-equivalent.toml` through the complete v1.10 pipeline, canonicalizes both DOT outputs with `canonical.Canonical` (DI-1), and asserts equality — proving include + templates + relative-peer + reference compose to an indistinguishable rendered result
- `TestXC01_PipelineOrdering` is a behavioral test (D-20) asserting both order-dependent properties: XC-02 (template defined in an included file is visible to `[[use]]`) and XC-03 (parametrized peer resolves at the instantiation site) — robust to refactors that move pipeline calls into helpers (no source-scan)
- Both tests share a `renderThroughPipeline(t, path)` helper that mirrors `cmd/c4drill/root.go`'s exact pipeline ordering
- Both tests use `canonical.Canonical` (Plan 33-01) — never byte-exact `require.Equal` on raw DOT (DI-1)
- Both tests use `//nolint:paralleltest` (go-graphviz WASM convention)
- Pre-condition honored: Phases 30/31/32 all shipped before execution (internal/peer, internal/template, internal/include all exist; root.go wires all three passes)

## Task Commits

Each task was committed atomically:

1. **Task 1: TestXC05_ComposedEquivSingleFile** - `c823130` (test) — single commit covers both tests + the shared helper (they are tightly coupled: TestXC01 reuses the same helper)
2. **Task 2: TestXC01_PipelineOrdering** - `c823130` (test) — same commit
3. **Task 3: Full-suite regression gate** - (verification only — suite already green from Task 1; no commit needed per plan)

## Files Created/Modified
- `cmd/c4drill/root_test.go` — +118 lines: 6 new imports (graph, include, render, template, canonical, view); `renderThroughPipeline` helper; `TestXC05_ComposedEquivSingleFile`; `TestXC01_PipelineOrdering`. Pure additions — no existing test modified.

## Decisions Made
- **Open Question A2 (humanize site):** humanize is still at parse-time (`parser.go:614`, the Phase 29 stopgap). Phase 31's XC-04 comment notes the intent to relocate it post-expansion, but that relocation has NOT happened. The composed fixtures carry explicit `name=` on every unit so parse-time humanize does not fire for them; therefore no explicit humanize call is needed in `renderThroughPipeline`. This is consistent with the task brief's CONCURRENCY NOTE and is NOT a deviation.
- **XC-01 assertion form:** used the canonical edge substring `"platform.auth.cache" -> messageBus` (the exact form `serializeDOTStatement` emits for edge heads) rather than a model-level link inspection. Both forms are acceptable per the plan; the DOT-substring form is closer to the "behavioral proof" intent (asserts on the actual rendered output, not intermediate model state).
- **include.Resolve signature:** used the 3-arg form `Resolve(m, filepath.Dir(inputPath), inputPath)` per Phase 32's actual signature (NOT the simplified `Resolve(m)` from the original CONTEXT). The third arg threads the real entry filename so missing-include errors name the including file (INC-10/D-12).

## Deviations from Plan

None - plan executed exactly as written. Both tests passed on first run; no debugging iterations needed.

## Issues Encountered
None.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Phase 33 is now complete (all 4 plans shipped: 33-01 canonical helper, 33-02 fixtures, 33-03 docs, 33-04 E2E tests).
- All v1.10 milestone requirements (DOC-01/02/03, XC-01/05) are covered.
- The `renderThroughPipeline` helper is reusable by any future E2E golden.

---
*Phase: 33-docs-sweep-end-to-end-goldens*
*Completed: 2026-08-08*

## Self-Check: PASSED

- [x] `grep -c 'func TestXC05_ComposedEquivSingleFile' cmd/c4drill/root_test.go` returns 1
- [x] `grep -c 'func TestXC01_PipelineOrdering' cmd/c4drill/root_test.go` returns 1
- [x] `grep -c 'canonical.Canonical' cmd/c4drill/root_test.go` returns ≥ 2 (2 found)
- [x] `grep -c 'include.Resolve|template.Expand|peer.Resolve' cmd/c4drill/root_test.go` returns ≥ 3 (32 found)
- [x] Both tests cite XC-01, XC-02, XC-03 in comment blocks
- [x] No source-scan (no os.ReadFile on .go source files) — D-20 honored
- [x] `go test ./cmd/c4drill/ -run 'TestXC05_ComposedEquivSingleFile|TestXC01_PipelineOrdering'` exits 0
- [x] Both tests reuse the `renderThroughPipeline` helper (no duplicated pipeline logic)
- [x] No existing test modified (git diff shows pure additions)
- [x] Both tests have `//nolint:paralleltest`
- [x] `go test ./...` exits 0 — full suite green, no regression
- [x] `go vet ./...` exits 0
- [x] Cross-phase gate: Phases 30, 31, 32 all shipped before execution (internal/peer, internal/template, internal/include all exist)
- [x] Commit message starts with `test(33-04):`
