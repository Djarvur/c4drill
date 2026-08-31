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

## Milestone: v1.15 — Hierarchy Wrapping and Granular Keys

**Shipped:** 2026-08-30 (product releases v1.21.0 & v1.22.0; post-release fixes 2026-08-31)
**Phases:** 2 (37, 38) | **Plans:** 13 + 1 validated quick task (260831-01u, 3 tasks)

### What Was Built
- Phase 37 (v1.21.0): recursive cluster unfolding, deep-link ancestor chains, cluster drill affordance, `--plain` CLI key (CTX/PLAIN/DOC/REL requirements)
- Phase 38 (v1.22.0): full ancestor-chain wrapping incl. boundary/sibling entries (WRAP), granular switches `--no-colors/--no-styles/--no-lengths/--no-ranks` composing with `--plain` (KEY), `--no-labels` (LBL); KEY-03 switch matrix locked E2E; real golden re-baseline
- Quick task 260831-01u (TDD): compact C1 root restored, `--no-labels` narrowed to edge labels only, flag-invariant edge identity via builder-assigned `Edge.Name`
- Repo hardening: 0 golangci-lint issues, branch protection requiring Build/Lint/Test on master, Validate Examples asymmetry (open since 2026-08-14) fixed

### What Worked
- **Single-day two-release cadence**: phases 37+38 planned and executed back-to-back in one calendar day, each shipping its own product release — the single-phase milestone precedent (v1.11–v1.14) held.
- **Capture → validated quick pipeline for user bug reports**: three reports captured with code touch-points and context (including the "93.8 Mpt² measured on a broken basis" insight), then fixed in one `--validate` quick run — plan-checker confirmed the mechanism before execution, verifier confirmed against a freshly built binary.
- **Post-release bisection discipline**: the reporter blamed b2447da/WRAP; empirical version bisection redirected the fix to v1.21.0 CTX-02/03 — fixing the wrong commit would have shipped a second regression.

### What Was Inefficient
- **Golden fixtures under-modeled real usage**: goldens covered C1/expanded on fixtures without cross-container deep links, so a whole-model root flood shipped green through 38-04's "zero-delta" re-baseline. The human's real model caught it immediately.
- **Lint debt accumulated to 60 findings** (some pre-existing from 37/38, some from the quick task) — lint only ran on push, and the push-gate failure was noticed days of commits later. Fixed + branch protection added; new-code lint should be run before push.
- **Plan-ID comment markers (`BUG-1/BUG-2/BUG-3`) tripped godox** — 18 mechanical rewords post-hoc. Avoid TODO/BUG/FIXME tokens in plan-reference comments.
- **Manager dashboard staled after artifact archiving**: `init.manager` recommended re-discussing the shipped-and-archived phase 37; disk status and shipped reality diverged until the milestone close.

### Patterns Established
- **Flags are presentation-only — topology must be flag-invariant**: merge/dedup keys are computed from source-model attributes; `TestEdgeTopologyFlagInvariant` pins identical edge multisets across flag compositions.
- **Compact-root invariant**: non-expanded root shows only the visible neighborhood; unfolding follows visible paths + deep-link chains, never whole subtrees; pinned by `deepcross` fixture + golden.
- **Fixtures as regression contracts**: every behavioral fix ships a fixture that reproduces the reported shape (deepcross), not just a unit test.

### Key Lessons
1. **Bisect empirically before attributing a regression.** The reported culprit commit (b2447da) was provably C2/C3-scoped; static analysis plus version bisection found the real introduction point (v1.21.0). Attribution from a diff read alone is a hypothesis.
2. **Superseding a pinned requirement is a docs+goldens+matrix event.** Narrowing `--no-labels` (LBL-01) required re-baselining goldens, updating the E2E matrix, 4 docs copies, and explicit supersession notes in STATE/archive — plan for the blast radius, not just the code.
3. **Run the linter before pushing, not just CI.** 60 findings across pre-existing and fresh code blocked master's sanity run; local `golangci-lint run ./...` would have caught it pre-push.

### Cost Observations
- Model mix: opus planner, sonnet executor/checker/verifier (quick task); phase 37/38 per-plan mixes per STATE velocity notes.
- Sessions: one orchestrator session covering 3 bug captures → 1 validated quick task → lint hardening → CI green → milestone close.
- Notable: three user-reported bugs were captured, planned, plan-checked, executed (7 TDD commits), verified (5/5 must-haves), and documented within the same session as the milestone close.

---

## Milestone: v1.16 — Edge Style Override

**Shipped:** 2026-08-31 (product release v1.23.0)
**Phases:** 1 | **Plans:** 3

### What Was Built
Invocation-global `--edges <style>` CLI flag (straight|spline|square|ortho) with loud enum validation; beats global AND per-unit model edges on every generated view via a dedicated `View.EdgesOverride` carrier applied post-PLAIN-02 in both builders; explicit flag survives `--plain` (documented KEY-02 delta, pinned by `TestEdgesSurvivesPlain`); GEDGE-07 switch-matrix E2E (~86 cells asserting `splines=` in RAW dot) over a golden-free two-layer precedence fixture; README + 3 byte-identical SKILL.md copies.

### What Worked
- **Design-todo-as-spec:** the folded todo carried flow, file:line list, and open questions — the milestone went from candidate to shipped in one day with zero research dependencies.
- **Pattern replication:** phase 39 is the third member of the PLAIN/KEY flag family; copying the threading + matrix harness made each plan land TDD-clean in ~10–15 min.
- **Verifier-friendly invariance:** the empty-default `EdgesOverride` carrier made flag-off behavior structural (not merely tested) — zero golden churn, zero re-baselining.

### What Was Inefficient
- gocognit (limit 15) tripped twice on first-draft forms (inline validation branch in `runRoot` at 16; nested-loop matrix at 28) — both fixed by extraction/flat-table refactors, but the limit keeps surprising first drafts.
- Plan 39-03's cross-reference pointed at the wrong README section (`=== Edges` documents C4D arrows, not routing) — caught and moved to the Properties Section during execution.

### Patterns Established
- Explicit-CLI-beats-plain precedence: user intent > author-format suppression, expressed as a dedicated post-suppression override field.
- RAW-dot attribute assertions must use the verified graphviz emission form (`splines=true|false|ortho`, unquoted) — verified against actual output before pinning.
- Golden-free fixtures for new-behavior tests keep the backward-compat invariant untouched.

### Key Lessons
1. Settle precedence questions at phase-context time and pin them with a named test — the `--plain` decision (D-05) was the only open design question and it closed cleanly.
2. Free-text "Depends on" fields confuse tooling (init.manager parsed version numbers out of prose as dep IDs) — keep roadmap dependency fields machine-clean.

### Cost Observations
- Model mix: sonnet researchers/synthesizer/roadmapper at milestone setup; the background agent ran the full plan→execute→verify→release chain for phase 39.
- Sessions: one manager session — discuss inline, plan+execute via background agent, UAT and milestone close in the same session.
- Notable: human touches were 3 prompt answers (scope confirm, dispatch Continue, verify choice); all 7 UAT tests were agent-executed CLI checks with grep-able evidence.

---

## Cross-Milestone Trends

### Process Evolution

| Milestone | Sessions | Phases | Key Change |
|-----------|----------|--------|------------|
| v1.10 | 1 + 16 bg agents | 6 | First use of parallel background executors with file-overlap-aware serialization; research-first milestone setup (4 parallel researchers) |
| v1.15 | 1 orchestrator + planner/checker/executor/verifier | 2 (+1 quick task) | Capture→validated-quick pipeline for post-release bug reports; empirical bisection over diff-reading attribution; flags-must-not-change-topology invariant tests; branch protection added after lint debt blocked master |
| v1.16 | 1 manager session (discuss inline; plan+execute via background agent) | 1 | Todo-as-spec single-day milestone; manager-driven dispatch; research skipped deliberately for an in-pattern feature; UAT fully agent-executed for a CLI surface |

### Cumulative Quality

| Milestone | Tests | Coverage | Zero-Dep Additions |
|-----------|-------|----------|-------------------|
| v1.10 | full suite green (12 packages) | canonicalDOT goldens (order-insensitive) | 0 new deps (stdlib + existing go-toml/go-graphviz/testify only) |
| v1.15 | full suite green (15+ packages), 0 lint issues | canonicalDOT goldens + switch-matrix E2E + deepcross root invariant | 0 new deps |
| v1.16 | full suite green, 0 lint issues | canonicalDOT goldens (zero churn) + edges composition matrix (~86 cells) | 0 new deps |

### Top Lessons (Verified Across Milestones)

1. **Hand-roll over library for small, known-shape structs.** v1.10's deep-copy, TOML-merge, humanizer, and param-substitution were all hand-rolled (~95 LOC total) because every candidate library either silently dropped unexported fields (`Link.Mirror`) or was overkill (`text/template`). Verified: zero new deps added in v1.10.
