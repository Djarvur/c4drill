# Phase 30: Relative-peer resolution - Context

**Gathered:** 2026-08-08
**Status:** Ready for planning

## Phase Boundary

Users can write short `peer` values that resolve against the enclosing parent block's ancestry, eliminating repetitive absolute paths. A post-parse pass rewrites `Link.Peer` from relative to absolute on the assembled (post-include, post-template-expansion) model before validation. The pass is a **no-op for every existing model** — backward-compat is a hard contract.

This phase delivers ERGO-01 (relative resolution) and ERGO-02 (absolute-fallback). It depends on nothing at runtime (ships as a no-op for absolute-only models) but composes with Phase 31 (template expansion): per D-01/D-02 from Phase 31's CONTEXT, relative peers authored inside templates resolve at the **instantiation site**, and the pass runs AFTER expansion so templated-unit links resolve correctly.

## Implementation Decisions

### Enclosing-parent definition (foundational)

- **D-13:** For a unit at path `a.b.c` declaring a link, the "enclosing parent" whose children the resolver searches first is **`a.b`** — `c`'s *immediate parent*. Relative resolution tries `a.b.x` (c's sibling). NOT `a`'s children (grandparent — that's the walk-up, D-14), and NOT `c`'s own children (subunits; a unit doesn't link to its own subunits anyway per validation rule VALD-02). Natural lexical-scoping reading.

### Resolution depth

- **D-14:** **Walk-up ancestry, nearest-first.** When a bare peer doesn't match a direct sibling, the resolver walks UP: immediate parent's children → grandparent's children → ... → top-level (root, per D-15). The first depth with exactly-one-match wins (closest ancestor's matching child). **Multiple matches at the SAME depth = hard error** (ROADMAP criterion 3); first depth with exactly one match resolves. Mirrors lexical scoping in programming languages; makes relative peers genuinely useful in deep models (e.g. a handler at `frontend.api.handlers.auth` can say `peer = "cache"` and find `frontend.cache` without spelling the full path).
- **D-15:** The walk-up **reaches root** — top-level units are the outermost scope. So `peer = "messageBus"` from any depth resolves to a top-level `messageBus` if one exists, even with no sibling/aunt match. This makes top-level shared services (message bus, auth service) reachable by bare name from any depth — the common case in real models. Miss at root → absolute-fallback (D-16) → hard error.

### Bare-vs-absolute gate (precedence)

- **D-16:** **Unified algorithm, no special top-level short-circuit.** For a peer value:
  1. **If it contains `.`** → treat as absolute. Resolve as-is; miss = existing peer-existence error (validator `rules.go:14`). No relative resolution attempted.
  2. **If bare (no `.`)** → run the walk-up (D-13/D-14/D-15): nearest ancestor's children first, up to top-level as outermost scope. First depth with exactly-one-match wins; multiple matches at same depth = hard error; miss at root = hard error ("cannot resolve peer `X` from unit `Y`").
  - **NO separate "exact top-level match" check** — top-level is just the final step of the walk-up. A bare peer that happens to match a top-level unit resolves via the walk-up's root step, giving the identical result to today's absolute resolution. This preserves the backward-compat hard contract: every existing model resolves identically (absolute-with-dot via step 1; bare-matching-top-level via step 2's root step).
  - Rejected the three-step gate (with explicit top-level short-circuit) because it introduces a shadowing ambiguity: a sibling named the same as a top-level unit would resolve to top-level (short-circuit) instead of the sibling, which is silently wrong. The unified rule has no special cases and handles cross-depth shadowing via nearest-first.

### Already-locked decisions (carried from REQUIREMENTS / Phase 31, NOT re-discussed)

- **ERGO-01:** Bare peer (no `.`) resolves against enclosing parent's children (now precisely defined: walk-up per D-14).
- **ERGO-02:** Absolute fallback when peer contains `.` OR does not resolve as a relative sibling (now precisely defined: D-16 unified gate).
- **HS-2 (from Phase 31 D-01/D-02):** Relative peers authored inside a template resolve at the **instantiation site's** parent, NOT the template's lexical location. A parentless (top-level) templated unit resolves against top-level siblings with absolute-fallback — uniform with hand-authored top-level units. The relative-peer pass runs AFTER template expansion (per the pipeline `include → template-expand → relative-peer-resolve → humanize → validate`) and treats every unit identically — NO template-special-case logic.
- **Implementation site:** Separate post-parse pass that rewrites `Link.Peer` in place on the assembled model before `BuildIndex`/validation, so the validator's existing absolute-path logic is untouched. Stdlib only (`strings.Contains` for the `.` gate; the model's parent/child relationships for the walk-up). No new dependency.

### Claude's Discretion

- Internal package/location for the resolver (research suggests `internal/peer/Resolve` or a function in `internal/parser/`; planner's call). Likely signature: `func Resolve(m *parser.Model) error` (in-place rewrite of `Link.Peer`; error for unresolvable/ambiguous).
- Error message wording for the two hard-error cases (same-depth ambiguity; miss-at-root) — lean toward naming the peer, the link-host unit, and (for ambiguity) the competing matches.
- Whether to surface a warning (not error) on cross-depth shadowing (e.g. `frontend.cache` shadowed by `frontend.api.cache`) — the nearest-first rule silently picks the nearer; a non-blocking warning would aid debugging. Optional, planner's call.

## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Project planning (load-bearing context)
- `.planning/REQUIREMENTS.md` — ERGO-01, ERGO-02 (Phase 30 requirements). §Traceability maps both to Phase 30.
- `.planning/ROADMAP.md` — Phase 30 section: goal, requirements, 3 success criteria (incl. criterion 3: single-depth ambiguity = hard error; criterion 2: corpus byte-identical backward-compat test is the acceptance criterion).
- `.planning/phases/31-template-expansion/31-CONTEXT.md` — D-01/D-02 (HS-2 resolution site: instantiation-site; parentless = top-level siblings). Phase 30's pass must work uniformly on post-expansion units per this decision.
- `.planning/todos/pending/2026-08-08-toml-authoring-ergonomic-improvements.md` — original feature intent (tagged resolves_phase: [29,30]); Part 1 (relative peer resolution) is this phase.

### Research (load-bearing conclusions)
- `.planning/research/SUMMARY.md` — §3 (ergonomics: relative-peer is a differentiator — no surveyed tool does sibling-scope-walk; walk-up precedence); §4 (pipeline — peer.Resolve runs AFTER template.Expand, before humanize, before validate); §6 RP-2 (walk-up semantics — now settled by D-14/D-15/D-16); §9 RP-2 (backward-compat: the pass must be a no-op for every existing model; corpus test is acceptance).
- `.planning/research/PITFALLS.md` — HS-2 (resolution-site ambiguity, settled by Phase 31 D-01); RP-2 (backward-compat gate: bare + no `.` + not top-level key + not in index — now refined by D-16's unified algorithm); BC-3 (non-strict toml load-bearing).

### Code (integration points — confirmed live by the Phase 31 planner)
- `internal/parser/parser.go:35` (`Model` struct) — the resolver reads `m.Units` (the map of path→Unit) to walk parent/child relationships.
- `internal/model/link.go:43` (`Link` struct, `Peer` field) — the resolver rewrites `Link.Peer` in place (relative → absolute).
- `internal/model/unit.go:71` (`Subunits map[string]*Unit`) — used to enumerate a parent's children at each walk-up depth.
- `cmd/c4drill/root.go:112-118` — the resolver inserts here, AFTER `template.Expand` (Phase 31) and BEFORE `Validate` (`:118`). Per the Phase 31 planner's finding, the pipeline currently goes straight Parse→Validate; this pass and template.Expand both insert in this gap.
- `internal/validator/rules.go:14` — existing peer-existence check; runs AFTER the resolver, so it sees absolute paths only. Unchanged.

## Existing Code Insights

### Reusable Assets
- `strings.Contains(peer, ".")` — the bare-vs-absolute gate (D-16 step 1). Trivial stdlib.
- The model's parent/child relationships are already implicit in the dotted-path keys + `Unit.Subunits` map — the walk-up (D-14) is a loop over `strings.Split(path, ".")` progressively dropping the last segment, checking each ancestor's `Subunits` for the bare name. No new data structure needed.

### Established Patterns
- **Validator is the single gatekeeper** (STATE.md D-12) — the resolver is a pure pre-processing pass that produces a model whose `Link.Peer` values are all absolute; `validator.Validate` consumes it unchanged.
- **Order preservation is load-bearing** — the resolver does NOT touch `UnitOrder`; it only rewrites `Link.Peer` strings in place.
- **Non-strict TOML unmarshaling is load-bearing** (PITFALLS BC-3) — do NOT enable `DisallowUnknownFields()`.
- **CanonicalDOT for corpus tests** (STATE.md DI-1) — the backward-compat corpus test (ERGO-02 acceptance criterion) should assert the `(source, resolved-peer)` set is byte-identical pre/post, OR use order-insensitive canonicalDOT on rendered output if it goes end-to-end. Lean toward the peer-set comparison — it's a sharper, faster test than rendering.

### Integration Points
- **Pipeline insertion** (`cmd/c4drill/root.go`): `peer.Resolve` is a new call in the gap between `ParseFile` (`:112`) / `template.Expand` (Phase 31) and `Validate` (`:118`). Signature: `func Resolve(m *parser.Model) error`.
- **Rewrite target** — every `Link.Peer` on every `Link` in every unit's `Links` and `LinksFrom` (including templated/subunit links, post-expansion). In-place rewrite; no model-structure change.

## Specific Ideas

- Example resolution traces (for the planner's test design):
  - `[linuxSystem.localIDP.sessionAPI]` with `peer = "sessionManager"` → sibling match → `linuxSystem.localIDP.sessionManager`.
  - `[frontend.api.handlers.auth]` with `peer = "cache"` → no sibling, walk up → `frontend.api.cache`? no → `frontend.cache`? yes → resolves to `frontend.cache`.
  - `[linuxSystem.sshAuth.sshd]` with `peer = "messageBus"` → walk up to root → top-level `messageBus` → resolves.
  - `[a.b.c]` with `peer = "x"` where both `a.b.x` and `a.x` exist → nearest-first → `a.b.x` (the nearer). No error (different depths).
  - `[a.b.c]` with `peer = "x"` where both `a.b.x` and `a.b.y` exist but two siblings named `x`? impossible (map keys unique). Same-depth ambiguity requires the bare name to match two DIFFERENT-typed children at the same ancestor — actually impossible given map-key uniqueness within one parent. So same-depth ambiguity (ROADMAP criterion 3) can only arise if... hmm, actually it can't arise within a single parent (Subunits is a map). It could only arise across the walk if the same name exists at multiple depths, which nearest-first handles silently. Worth the planner verifying whether criterion 3's "single-depth ambiguity" is reachable at all under the walk-up model — if not, that error case may be dead code. **Flag this for the planner to analyze.**

## Deferred Ideas

- **Cross-depth shadowing warning** — when `frontend.cache` is shadowed by a nearer `frontend.api.cache`, nearest-first silently picks the nearer. A non-blocking warning ("peer cache resolved to frontend.api.cache, shadowing frontend.cache") would aid debugging but adds noise. Deferred — planner may include as opt-in/non-blocking if it's cheap.
- **Compact one-liner link shorthand (ERGO-06)** — assigned to Phase 29 (at-risk per research §3); not Phase 30's concern.

### Reviewed Todos (not folded)
- `2026-08-08-toml-authoring-ergonomic-improvements.md` — tagged resolves_phase: [29,30]. Part 1 (relative peer) is this phase; Part 2 (optional name) is Phase 29; Part 3 (compact link) is Phase 29 (at-risk). Already accounted for in REQUIREMENTS/ROADMAP.
- Other pending todos (reference field → 28; templates → 31; include → 32; docs → 33) — not Phase 30's scope.

---

*Phase: 30-Relative-peer resolution*
*Context gathered: 2026-08-08*
