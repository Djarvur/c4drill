# Phase 29: Optional name humanization - Context

**Gathered:** 2026-08-08
**Status:** Ready for planning
**Source:** Autonomous synthesis from REQUIREMENTS.md (ERGO-03/04/05/06), ROADMAP.md Phase 29, research SUMMARY.md §3/§4/§6/§9, and the original feature todo. No discuss-phase fork exists; this phase has a single locked design direction per the research conclusions.

<domain>
## Phase Boundary

Ship optional display-name derivation. When a unit omits its `name` field, the
display name shown in rendered diagrams is derived from the **last path segment**
of the unit's TOML identifier via a dumb camelCase split. Explicit `name =` always
wins (backward-compat hard contract: every existing model renders identically to v1.9).

**In scope (ERGO-03, ERGO-04, ERGO-05):**
- A humanize function (dumb camelCase → space-separated Title Case) in the `model` package.
- Parse-time fallback: when `Unit.Name == ""`, populate `Unit.Name` from the unit's own identifier segment.
- Tests covering the humanize algorithm, the explicit-name-wins rule, and a corpus/no-regression assertion that all existing fixtures render byte-identically.
- README + skill doc update documenting that `name` is optional with the humanize rules and a before/after example.

**At-risk / deferred (ERGO-06):**
- Compact one-liner link shorthand. Research SUMMARY §3 classifies compact-link string shorthand as a v1.10 **anti-feature** (fights structured-TOML; the inline array form + relative peers from Phase 30 are sufficient). The discuss-phase decision on confirm-vs-defer has not happened. Per the original todo's own sequencing ("defer #3; re-evaluate after #1+#2"), and per REQUIREMENTS.md "Future Requirements" which already lists compact-link shorthand variants as deferred, **ERGO-06 is deferred to Future Requirements in this plan**. If a follow-up discuss-phase confirms inclusion, ERGO-06 will need its own phase or be folded into Phase 30 (relative peer, where the inline form gets compact enough). This phase does NOT implement a link shorthand.

**Out of scope for Phase 29 (owned by other phases):**
- XC-04 (humanize runs after template expansion, before validate) — mapped to Phase 31. Phase 29 lands the humanize *function* and a parse-time fallback hook; Phase 31 will relocate/extend the hook into the post-expansion pipeline so templated units humanize from their substituted instantiation key, not `${name}`.
- Relative-peer resolution (ERGO-01/02) — Phase 30.
- Reference field (REF-*) — Phase 28.

</domain>

<decisions>
## Implementation Decisions

### D-01 Humanize algorithm — dumb camelCase split, NO acronym table (ERGO-04)
**Decision:** `Humanize(segment string) string` splits the last identifier segment on camelCase boundaries and Title-cases each word, with **no acronym preservation**. Locked by ERGO-04 and research §3 ("acronym preservation is an anti-feature — Terraform `title()` proves it unsolved").

Rules (deterministic, single source of truth):
- Split at each lower→upper transition (`localIDP` → `["local", "IDP"]`).
- Split at each upper→upper→lower transition (`IDPToken` → `["IDP", "Token"]`) so consecutive capitals don't get glued.
- Join words with single space; Title-case the first letter of the FIRST word's first lowercase letter if it starts lowercase, leave the rest of each word as-authored? **No** — keep it dumb and predictable: each word is lowercased then first-letter-capitalized, so `gRPC` → `Grpc`, `APIs` → `Apis`. This is exactly the ERGO-04 contract (`gRPC` → "Grpc"); authors escape via explicit `name =`.
- Empty input → empty output. Already-title-case or single-word input returns the word capitalized (e.g. `webapp` → "Webapp").

Reference outputs (these become the unit-test table):
| segment | humanized |
|---|---|
| `linuxSystem` | "Linux System" |
| `localIDP` | "Local IDP" |
| `sessionManager` | "Session Manager" |
| `sessionAPI` | "Session API" |
| `gRPC` | "Grpc" |
| `grpcAPIs` | "Grpc Apis" |
| `webapp` | "Webapp" |
| `IDPToken` | "Idp Token" |
| `""` | "" |

**Why no acronym allowlist:** research §3 + §6 HU-1 proposed `golang.org/x/text/cases` + an acronym allowlist, but ERGO-04 and the project brief override this — acronym preservation is an anti-feature, and an allowlist creates an unmaintainable special-case surface (every project wants different acronyms). The dumb split is the entire point of ERGO-04. **No new dependency.**

### D-02 Humanize operates on the LAST path segment only (ERGO-03)
**Decision:** humanization derives from the last segment of the unit's identifier, not the full dotted path. `[linuxSystem.localIDP]` with no `name` → "Local IDP" (from `localIDP`), not "Linux System Local IDP". Locked by ERGO-03 ("derived from the last path segment") and the original todo recommendation ("segment-only is recommended — the full path is structural, the segment is the name"). Top-level units `[linuxSystem]` use the single segment `linuxSystem` → "Linux System".

### D-03 Explicit `name =` always wins (ERGO-05)
**Decision:** humanization ONLY fires when `Unit.Name == ""`. Any explicit `name = "..."` (including `name = ""` if that is ever authored — but empty-name humanize still applies since the field is empty) takes precedence and is rendered verbatim. This is the backward-compat hard contract: every v1.9 model already has explicit `name` on (nearly) every unit, so they render byte-identically. Locked by ERGO-05.

### D-04 Placement — parse-time fallback in the parser (for v1.10's Phase 29 scope)
**Decision:** Phase 29 lands humanization as a **parse-time fallback** in `internal/parser/parser.go`: after a `Unit` is built in `parseUnitWithOrder`, if `unit.Name == ""`, set `unit.Name = model.Humanize(lastPathSegment)`. The segment is the `name`/`subName` argument already threaded through `parseUnitWithOrder` (parser.go:160-210).

**Why parse-time and not a separate `runRoot` pipeline stage for this phase:** XC-04 (humanize-runs-after-template-expansion) is Phase 31's contract. Phase 29 ships with templates not yet existing, so parse-time is correct *today* and is the simplest, lowest-risk integration. Phase 31 will extract the humanize call into a dedicated post-expansion pass; the `model.Humanize` function itself is unchanged by that move. The function is the durable artifact; the call site may move in Phase 31.

This keeps Phase 29 a pure, self-contained, no-pipeline-reorder change — consistent with research §4 "each pass is independently shippable (each is a no-op for models that don't use the feature)" and the ROADMAP "Depends on: Nothing (parallelizable with 28)".

### D-05 Where Humanize lives — `internal/model` package
**Decision:** put `func Humanize(segment string) string` in a new file `internal/model/humanize.go` (package `model`), with its tests in `internal/model/humanize_test.go`. Rationale: it's a pure string utility operating on a unit-identifier segment with no parser/toml dependency; co-locating with `Unit` (unit.go) is the natural home; the parser imports `model` already. Keeps it reusable by Phase 31's pipeline-stage refactor without import cycles.

### D-06 Backward-compat / regression guarantee
**Decision:** the phase MUST include a corpus/regression assertion that every existing test fixture (`testdata/valid.toml`, `testdata/nested.toml`, and any `skill/examples/*.toml` used as golden input) renders identically. Because those fixtures already carry explicit `name =`, the humanize fallback is a no-op for them — the test proves it. New fixtures demonstrating the omitted-name case are added alongside.

### Claude's Discretion
- Exact internal helper signature details (e.g. whether the camelCase split is a single regex or a hand-rolled rune scanner) — planner/executor's choice, but it MUST produce the D-01 reference table exactly and add no dependencies. A hand-rolled rune loop is preferred (no `regexp` dependency, trivially readable, ~25 LOC per research §2).
- Test naming and table-driven structure — follow existing `parser_test.go` conventions (`t.Run` subtests, `require`/`assert` from testify).
- Whether to add a `skill/examples/05-optional-name.toml` fixture number — DOC-03 (example fixtures) is owned by Phase 33, but a small local fixture used by the parser test is in scope here; the skill-doc example is the docs task in this phase (ERGO is ergonomics, and README+skill doc updates are part of every ergonomics feature per the todo's "Validation/docs" section).

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Requirements & roadmap
- `.planning/REQUIREMENTS.md` — ERGO-03, ERGO-04, ERGO-05 (in scope); ERGO-06 (deferred); XC-04 (Phase 31, not this phase). Lines 43-48 for ERGO reqs.
- `.planning/ROADMAP.md` — Phase 29 section (goal, success criteria, ERGO-06 at-risk note).

### Research
- `.planning/research/SUMMARY.md` §3 (Feature Classification — "optional `name`" is a Differentiator; acronym preservation is an Anti-feature; compact-link shorthand is an Anti-feature) — the load-bearing justification for D-01 and the ERGO-06 deferral.
- `.planning/research/SUMMARY.md` §2 (identifier humanization row — "~25-line splitter + small acronym table wins" is OVERRIDDEN by ERGO-04 to drop the acronym table) and §9 HU-2 (humanize ordering — runs after expand in Phase 31; this phase ships parse-time).
- `.planning/research/ARCHITECTURE-v1.10.md` §6 phase table row 29 ("Humanize function in parser/model; `Unit.Name` fallback at parse time; docs").

### Codebase (touch points — read-only references for the planner)
- `internal/model/unit.go` — `Unit` struct, `Name string` field at line 45.
- `internal/parser/parser.go` — `Parse` (line 47), `parseUnitWithOrder` (line 160; the `name`/`subName` arg is the identifier segment), `isBuiltinField` (line 309; `name` is already an allowed builtin field — no change needed).
- `internal/parser/parser_test.go` — existing test patterns (table-driven, `testdata/*.toml`, testify `require`/`assert`).
- `cmd/c4drill/root.go` — `runRoot` pipeline (lines 85-150); parser output flows straight to validator then views. Humanize-at-parse means the pipeline is untouched in this phase.

### Feature intent
- `.planning/todos/pending/2026-08-08-toml-authoring-ergonomic-improvements.md` — original todo, "Optional `name` field" section (#2) and "Compact one-liner link syntax" section (#3, deferred).

</canonical_refs>

<specifics>
## Specific Ideas

- **Reference test table** (D-01) — the humanize unit test MUST encode exactly this table; it is the algorithm's contract. Any implementation that deviates fails the test.
- **Segment source** — in `parseUnitWithOrder(name, value, parentType, subunitOrder, subunitOrders)`, the `name` parameter IS the last path segment for both top-level units (`parseUnitWithOrder(name, ...)` called from the `Parse` loop at parser.go:87) and subunits (`parseUnitWithOrder(subName, ...)` at parser.go:210). No path-splitting logic is needed in the parser — pass the existing arg to `model.Humanize`.
- **Two call sites** to add the fallback: after `toml.Unmarshal(unitData, &unit)` and type inference at parser.go:181-193, insert `if unit.Name == "" { unit.Name = model.Humanize(name) }` (where `name` is the function parameter). This covers both top-level and nested units because both flow through `parseUnitWithOrder`. One insertion point, both cases covered.
- **README/skill doc** — add a short "Optional name" subsection under the TOML format reference, with the `localIDP → "Local IDP"` example and the explicit `name =` escape hatch. Note: ERGO-04 dumb-split behavior (`gRPC → Grpc`) must be documented so authors know to set explicit `name` for acronyms.

</specifics>

<deferred>
## Deferred Ideas

- **ERGO-06 (compact one-liner link shorthand)** — deferred to Future Requirements (already listed there in REQUIREMENTS.md). Research §3 classifies it as a v1.10 anti-feature. The inline array form `link = [{peer="x", technology="y"}]` already exists and, combined with Phase 30's relative peers, is sufficient. Re-open only if a discuss-phase explicitly confirms inclusion; would then likely fold into Phase 30 or get its own phase. ERGO-06 is NOT covered by any plan in this phase.
- **Acronym preservation / acronym allowlist** — explicitly out of scope (research §3 anti-feature; ERGO-04). Never coming back unless ERGO-04 itself is revised.
- **Pipeline-stage relocation (XC-04)** — Phase 31 owns moving the humanize call from parse-time to post-template-expansion. This phase ships the parse-time version; the `model.Humanize` function is the stable artifact Phase 31 reuses.
- **Humanize-from-full-path option** — D-02 locks segment-only; full-path humanization is rejected (structural path is not a name).

</deferred>

---

*Phase: 29-optional-name-humanization*
*Context gathered: 2026-08-08 via autonomous synthesis (no discuss-phase; requirements fully specified in REQUIREMENTS.md + research SUMMARY §3/§6)*
</content>
</invoke>
