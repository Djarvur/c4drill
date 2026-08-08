# Project Retrospective

*A living document updated after each milestone. Lessons feed forward into future planning.*

## Milestone: v1.10 — Model Composition

**Shipped:** 2026-08-08
**Phases:** 6 | **Plans:** 13 | **Tasks:** 35

### What Was Built
- Four-feature composition pipeline (`include → template-expand → relative-peer-resolve → humanize → validate → render`) with zero changes to validator/view/render consumers
- `[template.*]` parametrized unit definitions with `[[use]]` instantiation; hand-rolled recursive `Unit.Clone()` preserving the unexported `Link.Mirror` field
- `[[include]]` multi-file assembly with cycle detection, `once` dedup, same-file-diamond dedup, cross-file-subunit merge, 3-arg `Resolve(entry, entryDir, entryFile)`
- `internal/peer/Resolve` walk-up ancestry (D-13..D-16) + reusable `internal/testutil/canonical` order-insensitive DOT comparator extracted from builder_test.go prior art
- 9 runnable example fixtures + E2E composition proof (`TestXC05_ComposedEquivSingleFile` — multi-file ≡ single-file, byte-identical canonicalDOT)

### What Worked
- **Research-first milestone setup**: the 4-parallel-researcher phase (STACK/FEATURES/ARCHITECTURE/PITFALLS → SUMMARY) settled stack decisions (zero deps, hand-rolled clone) and surfaced the two HS risks (deep-copy aliasing, relative-peer-in-template ambiguity) before any code was written. Both were addressed correctly first-try.
- **Discuss-phase on the two design-heavy phases (31, 32)**: locked 8 + 4 decisions before planning, eliminating rework. The HS-2 resolution-site decision (instantiation-parent, not template-parent) was decisive — it would have caused silent cross-system edge corruption if left to the planner.
- **Parallelization with file-overlap awareness**: executing 28 first (smallest, independent) then 29 in parallel with Plan 30, then serializing 31→32→33, maximized throughput without git conflicts. The one narrow root.go overlap (30 + 31 each add one pipeline line) was correctly identified as mergeable.
- **Strictness stance throughout**: hard-error on missing param, missing include, duplicate path, unresolved peer, residual `${param}` — caught real issues early. No silent corruption shipped.
- **Plan 33's canonicalDOT extraction**: unblocked cross-package E2E goldens (Go's `_test.go` import boundary forced `internal/testutil/canonical/` — a sound autonomous call by the planner).

### What Was Inefficient
- **Pipeline-call-counter service noise**: the `agent-skills` SDK queries returned empty strings, forcing every spawned planner/executor to run "solo mode" (no Agent() tool in background context). Each phase's plan-phase ran the researcher/pattern-mapper/planner/checker inline rather than spawning subagents — higher per-phase token cost, slower wall-clock. Not a v1.10-specific issue but worth noting.
- **Two testdata corpora discovery (Phase 30)**: the planner caught mid-plan that root `/testdata/` and `cmd/c4drill/testdata/` both contain bare peers. Caught during planning (not execution), but the two-tree split is a latent footgun for future phases that touch testdata.
- **XC-04 humanize relocation deferred**: Phase 29 shipped humanize at parse-time as a stopgap; Phase 31's full relocation was deferred because templates carry explicit `name=`. This is a latent gap — a templated unit omitting `name` won't humanize from its instantiation key until the relocation happens. Not exercised by any fixture, so not a shipped bug, but a known limitation.
- **golangci-lint pre-existing findings in humanize.go** flagged out-of-scope by the Phase 31 executor (lll/mnd/godoclint). Carried forward as tech debt.

### Patterns Established
- **Composition as pure pre-processing passes on `*parser.Model`**: each v1.10 feature is a `func(m *parser.Model) (*parser.Model, error)` that runs in the Parse→Validate gap. Validator/view/render consume the assembled model unchanged. This pattern scales to future model-level features without touching downstream consumers.
- **Order-insensitive canonicalDOT for all multi-file/template goldens** (DI-1, now `internal/testutil/canonical.Canonical`): never byte-exact `require.Equal` on rendered DOT — go-graphviz nondeterminism + composition ordering variance makes it spuriously fail.
- **Hard-error-everywhere for composition features**: strictness over convenience. Missing-include, missing-param, duplicate-path, unresolved-peer, residual-token all error loudly with named sources.
- **Discuss-phase on design-heavy phases, planner-autonomous on clear-cut ones**: phases with open forks (31, 32) got full discuss; clear-cut phases (28, 29, 30, 33) let the planner synthesize CONTEXT autonomously. Right calibration.

### Key Lessons
1. **Verify the "reference implementation" before trusting it.** go-metadot's Go port silently drops macros at `graph.go:535`; the working spec is the Perl `metadot.pl`. Using the Go code as the spec would have replicated a broken feature. Always read the actual working implementation.
2. **Reserved-keyword collision policy dissolves when there's no legacy.** C4Drill isn't public yet, so the entire BC-2 reserved-word question (bare `[include]` vs namespaced `[c4drill.include]`) was moot — no migration machinery needed. Decide collision policy based on actual deployment state, not hypothetical future users.
3. **Map-key uniqueness makes some "ambiguity" rules dead code.** ROADMAP criterion 3 ("single-depth ambiguity = hard error") was structurally unreachable because `Subunits` is `map[string]*Unit` (unique keys per parent). Same-name matches can only arise at different depths, handled by nearest-first. Worth verifying whether error branches are actually reachable before authoring tests for them.
4. **Two parallel executors editing `root.go`'s pipeline gap is a real but narrow conflict.** Each adds one line calling its own package's `Resolve`/`Expand`. Mergeable, but not zero-risk. Serialize executors when they share even a one-line edit site.
5. **The `_test.go` import boundary forces shared test helpers into non-test packages.** `internal/testutil/canonical/` exists as a regular package (not `_test.go`) specifically so `cmd/c4drill/` tests can import it. Go's test-file import exclusion is a real architectural constraint on where reusable test utilities live.

### Cost Observations
- Model mix: ~95% sonnet (researchers, planners, executors, synthesizer), ~5% inherited (orchestrator). Zero opus dispatches.
- Sessions: 1 long orchestrator session managing the full milestone; 6 executor + 4 planner + 4 researcher + 1 synthesizer + 1 triage background agents.
- Notable: the entire 6-phase milestone (39 reqs, 13 plans, 35 tasks, +17.7k LOC) shipped in a single calendar day (2026-08-08) via parallel background agents. The research-first setup + discuss-on-design-heavy-phases pattern kept rework near zero.

---

## Cross-Milestone Trends

### Process Evolution

| Milestone | Sessions | Phases | Key Change |
|-----------|----------|--------|------------|
| v1.10 | 1 + 16 bg agents | 6 | First use of parallel background executors with file-overlap-aware serialization; research-first milestone setup (4 parallel researchers) |

### Cumulative Quality

| Milestone | Tests | Coverage | Zero-Dep Additions |
|-----------|-------|----------|-------------------|
| v1.10 | full suite green (12 packages) | canonicalDOT goldens (order-insensitive) | 0 new deps (stdlib + existing go-toml/go-graphviz/testify only) |

### Top Lessons (Verified Across Milestones)

1. **Hand-roll over library for small, known-shape structs.** v1.10's deep-copy, TOML-merge, humanizer, and param-substitution were all hand-rolled (~95 LOC total) because every candidate library either silently dropped unexported fields (`Link.Mirror`) or was overkill (`text/template`). Verified: zero new deps added in v1.10.
