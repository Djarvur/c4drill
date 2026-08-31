# Phase 39: Edge Style Override (`--edges` CLI flag) - Context

**Gathered:** 2026-08-31
**Status:** Ready for planning

<domain>
## Phase Boundary

A persistent `--edges <style>` CLI flag that overrides the edge routing style for a whole invocation — every generated diagram (C1 root, all C2/C3 drill-down views, the `--expanded` copy) renders with the requested routing style, beating the model's `edges` settings without editing the model file. No new routing styles; enum stays `straight|spline|square|ortho` (`square` = documented ortho alias per GEDGE-02). Scope anchored by REQUIREMENTS.md GEDGE-03..08 and ROADMAP.md Phase 39.

</domain>

<decisions>
## Implementation Decisions

### Flag Surface
- **D-01:** Flag name is `--edges <style>`, mirroring the TOML key (`edges = "spline"`), consistent with `--expanded` mirroring `expanded`. Verified no collision: existing persistent flags are `--format/-f`, `--output/-o`, `--expanded`, `--plain`, `--no-colors`, `--no-styles`, `--no-length`, `--no-rank`, `--no-labels`, `--label-ratio` (`cmd/c4drill/root.go:96-114`).
- **D-02:** Invalid values fail loudly before any output, naming the offending value and the allowed enum — reuse the existing sentinel-wrap pattern (`fmt.Errorf("%w: %q", errInvalidFormat, format)` precedent in `cmd/c4drill/root.go`).

### Override Semantics
- **D-03 (Claude discretion):** The override is **invocation-global**: `--edges` replaces the *resolved* `View.Edges` for every generated view, so it beats BOTH the global `[properties] edges` AND per-unit `edges` values (the `cmp.Or(unit.Edges, properties.Edges)` chain at `internal/view/scope.go:504`/`:599`). Rationale: (a) the `--plain`/`--no-*` switch family is invocation-global and never respects per-unit opt-outs; (b) the motivating use case is "render this model as a variant" — mixed output (flag wins at C1 but loses to a unit override at C2) would be unpredictable; (c) overriding the resolved view value is the single-choke-point rule. Overriding only `properties.edges` was the alternative and is rejected.
- **D-04:** Flag absent → behavior is exactly today's: per-view resolution `cmp.Or(unit.Edges, properties.Edges)` (C2/C3) and `properties.Edges` (C1/expanded) unchanged. GEDGE-08 backward compat: existing canonicalDOT goldens pass untouched.

### `--plain` Composition
- **D-05:** An explicit `--edges <style>` **survives `--plain`** — user intent beats author-format suppression. Precedence order: explicit CLI flag > `--plain` suppression > model-derived value. Implementation note: PLAIN-02 currently empties `Graph.EdgeStyle` under plain (`internal/graph/builder.go:38-41`); the flag must be applied so an explicit value wins over the plain zeroing. This is a deliberate, test-pinned delta to `--plain`'s documented "exact union" contract (KEY-02) — the dedicated GEDGE-06 test and the README `--plain` doc must state the delta explicitly.

### Testing
- **D-06:** Extend the v1.15 KEY-03-style switch matrix E2E: `--edges` × generation (root / drill-down / `--expanded`) × `--plain`, asserting the graphviz `splines` attribute in RAW dot output per combination. Add a dedicated test proving the flag beats a per-unit `edges` override (D-03 pin) and the GEDGE-06 `--plain` pin. Flag-off default: zero golden changes.

### Docs
- **D-07 (plan-level note, not a requirement):** README.adoc usage/flags section + the 3 SKILL.md copies document `--edges` and the `--plain` delta, byte-identical sync per the 37-06 convention.

### Claude's Discretion
The user did not respond to the area-selection prompts in this session (autonomous run). D-03 (override scope) and D-01 (naming) were resolved by Claude using codebase evidence and family precedent; all decisions are veto-friendly — the user can amend CONTEXT.md before `/gsd:plan-phase 39`.

### Folded Todos
- **Add CLI flag to override edge routing style** (`todos/pending/2026-08-30-add-cli-flag-to-override-edge-routing-style.md`, score 0.9, tagged `resolves_phase: 39`). This todo IS the phase's design source: problem statement (per-invocation variants require editing/duplicating the model), flow to extend (TOML `edges` → `View` → `Graph.EdgeStyle` → `cg.SetSplines`), file list, naming check, `--plain` open question (resolved here as D-05), and the switch-matrix E2E extension. Folded into scope verbatim as design input.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Milestone Planning Docs
- `.planning/REQUIREMENTS.md` — GEDGE-03..08 are the phase's requirements (traceability lives here)
- `.planning/ROADMAP.md` — §Phase 39: goal, success criteria, implementation notes
- `.planning/PROJECT.md` — KEY-02 "exact union" plain contract, validated-requirements history, Key Decisions table
- `.planning/todos/pending/2026-08-30-add-cli-flag-to-override-edge-routing-style.md` — design source (folded todo)

### Prior-Milestone Precedents
- `.planning/milestones/v1.15-ROADMAP.md` — KEY-03 switch-matrix E2E pattern to extend; PLAIN-01 threading precedent
- `.planning/milestones/v1.13-ROADMAP.md` — GEDGE-01 (global edge style everywhere) and GEDGE-02 (`square` → ortho alias)

### Code Touchpoints (verified 2026-08-31)
- `cmd/c4drill/root.go:96-114` — existing flag registrations (collision check evidence)
- `cmd/c4drill/root.go:330-386` — PLAIN-01 root + `--expanded` flag-threading pattern to copy
- `internal/view/scope.go:25,133,504,599` — where `View.Edges` is resolved (`cmp.Or(unit.Edges, properties.Edges)`)
- `internal/graph/builder.go:38-49` — PLAIN-02 plain-suppression of EdgeStyle (override must beat this)
- `internal/graph/builder.go:411-433` — expanded-mode edge style read
- `internal/render/converter.go:264-271` — `SetSplines` mapping (spline→"true", straight→"false", ortho/square→"ortho"); enum lands here unchanged

### Codebase Maps
- `.planning/codebase/ARCHITECTURE.md` — staged pipeline; cmd/c4drill is the only composer
- `.planning/codebase/CONVENTIONS.md` — sentinel-error pattern, nolint conventions, test conventions
- `.planning/codebase/TESTING.md` — testify/table-driven/paralleltest-nolint rules (render tests: no `t.Parallel()`)

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `Graph.EdgeStyle` field (`internal/graph/graph.go:66`) already carries the resolved style to the converter — the flag only needs to influence what lands there.
- `configureGraphSettings` / `SetSplines` mapping (`internal/render/converter.go:264-271`) accepts all four enum values today — zero render-side changes.
- The KEY-03 switch-matrix E2E (v1.15, Phase 38) is the direct template for GEDGE-07's matrix extension.

### Established Patterns
- Flag threading: PLAIN-01 pattern at `cmd/c4drill/root.go:330-386` — root command flag applied per-view AND to the `--expanded` copy. `--edges` follows it exactly.
- Loud CLI validation: `errInvalidFormat`-style sentinel + `%q` value wrap before pipeline start.
- Suppression-vs-explicit layering: PLAIN-02 zeroes `Graph.EdgeStyle` under `--plain` (builder.go:38-41) — D-05 defines explicit-flag-wins on top of it.

### Integration Points
- `runRoot` flag parsing (`cmd/c4drill/root.go:84`) — validate `--edges` value early, before Stage 1.
- `view` constructors (`GenerateC1View`/`GenerateC2View`/`GenerateC3View`/`GenerateExpandedView`) or the `processView`/`processExpandedView` call sites — the single choke point where the resolved `View.Edges` gets overridden (planner chooses exact point; D-03 constrains the semantics, not the mechanism).

</code_context>

<specifics>
## Specific Ideas

- Motivating example (from the folded todo): render the SAME model as expanded-with-straight and non-expanded-with-spline in one workflow — today that requires editing the model or keeping two copies.
- User intent principle (from requirements gate): `--plain` ignores *author* formatting; a typed CLI flag is *user* intent, so explicit `--edges` survives `--plain`.

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 39-Edge Style Override (`--edges` CLI flag)*
*Context gathered: 2026-08-31*
