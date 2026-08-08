# Roadmap: C4Drill

## Milestones

- ✅ **v1.0 Initial Release** — Phases 1-6 (shipped 2026-03-10)
- ✅ **v1.7 Queue Label Fix & Visual Improvements** — Phases 7-18 (shipped 2026-03-29)
- ✅ **v1.8 Proper C1/C2/C3 View Generation** — Phases 19-26 (shipped 2026-08-06) → [archive](milestones/v1.8-ROADMAP.md)
- ✅ **v1.9 C3 Boundary Node Fix** — Phase 27 (shipped 2026-08-06) → [archive](milestones/v1.9-ROADMAP.md)
- ✅ **v1.10 Model Composition** — Phases 28-33 (shipped 2026-08-08) → [archive](milestones/v1.10-ROADMAP.md)
- 🚧 **v1.11 LikeC4 Compatibility Layer** — Phases 34-37 (in progress)

## Phases

<details>
<summary>✅ v1.9 C3 Boundary Node Fix (Phase 27) — SHIPPED 2026-08-06</summary>

- [x] Phase 27: C3 Boundary Node Fix (1/1 plan) — completed 2026-08-06

Full details: [milestones/v1.9-ROADMAP.md](milestones/v1.9-ROADMAP.md)

</details>

<details>
<summary>✅ v1.10 Model Composition (Phases 28-33) — SHIPPED 2026-08-08</summary>

**Goal:** Expand C4Drill's authoring model from a single static TOML file into a composable, parametrized, multi-file format. Four additive features form a strict runtime pipeline: `include → template-expand → relative-peer-resolve → humanize → validate → generate-views → render`.

- [x] Phase 28: Reference field (📖) (1/1 plan) — completed 2026-08-08
- [x] Phase 29: Optional name humanization (2/2 plans) — completed 2026-08-08
- [x] Phase 30: Relative-peer resolution (2/2 plans) — completed 2026-08-08
- [x] Phase 31: Template expansion (2/2 plans) — completed 2026-08-08
- [x] Phase 32: Include directive (multi-file) (2/2 plans) — completed 2026-08-08
- [x] Phase 33: Docs sweep + end-to-end goldens (4/4 plans) — completed 2026-08-08

**Stats:** 6 phases, 13 plans, 35 tasks, 119 files changed (+17,703/−626). All 39 requirements validated.

Full details: [milestones/v1.10-ROADMAP.md](milestones/v1.10-ROADMAP.md)

</details>

---

### 🚧 v1.11 LikeC4 Compatibility Layer (In Progress)

**Milestone Goal:** Let `c4drill` render any valid LikeC4 (`.c4`) source by on-the-fly converting it into the existing TOML pipeline — breadth over fidelity, never fatal on unsupported constructs. A Pigeon-generated PEG parser produces an AST; a converter flattens it into a structurally-indistinguishable `*parser.Model` inserted as Stage 0 before `parser.ParseFile`. Everything downstream (`include.Resolve → template.Expand → peer.Resolve → Validate → view → render`) is reused verbatim.

**Granularity:** standard (4 phases)
**Coverage:** 16/16 v1.11 requirements mapped ✓ (PEG-01..04, CONV-01..05, INT-01..04, DX-01..03)
**Build/ship order:** de-risk PEG recovery first (Phase 34); converter keystone next (Phase 35) — it is where the "never fatal" contract, the `<->` two-link rule, the Humanize-bypass, and absolute-peer resolution all live; CLI wiring + warnings + single-file `extend` (Phase 36); end-to-end proof + docs last (Phase 37) so the TOML byte-identical contract is enforceable the moment dispatch lands and the `bigbank.c4` "render any model" proof gates ship.

- [ ] **Phase 34: PEG grammar + AST foundation** - `grammar.peg` → `go generate` → `grammar.go`; typed position-carrying AST; explicit recovery rules so `views {}`/`deployments {}` parse-and-discard rather than fatal
- [ ] **Phase 35: Converter core (keystone)** - AST → `*parser.Model`: flatten `{}` nesting to dotted subunits; `->`/`<->` (two links) to absolute-peer Links; three-tier fuzzy kind map; property/shape mapping; Humanize-bypass on `title`
- [ ] **Phase 36: Integration & warnings** - Stage 0 extension routing in `root.go`; `WarnCollector` (dedup by construct type, survives `-o`); single-file `extend` accumulator merge; TOML byte-identical regression lock
- [ ] **Phase 37: End-to-end proof & docs** - `bigbank.c4` canonical render; TOML↔LikeC4 cross-format equivalence golden; CLI regression; README "LikeC4 Compatibility" section + runnable fixture

## Phase Details

### Phase 34: PEG grammar + AST foundation

**Goal**: A LikeC4 source file parses into a typed, position-carrying AST, with unsupported top-level blocks (`views`, `deployment`, `global`) consumed-and-discarded via explicit PEG recovery rules rather than fataling.
**Depends on**: Nothing — de-risk PEG recovery FIRST (research SUMMARY §7 flags it as the new risk surface); the converter (Phase 35) cannot be built until the AST is parseable from real `.c4` files.
**Requirements**: PEG-01, PEG-02, PEG-03, PEG-04
**TDD note**: AST golden tests (via `google/go-cmp` `cmp.Diff`) on `bigbank.c4` and a minimal fixture are written BEFORE `grammar.peg` is considered complete — they lock the typed node shape the converter will consume.
**Build-order steps covered**: 1 (ast.go), 2 (lexer folded into PEG), 3 (lexer gate), 4 (parser = generated grammar.go), 5 (parser golden gate).
**Success Criteria** (what must be TRUE):

  1. `go generate ./internal/likec4/` produces a self-contained `grammar.go` from `grammar.peg` with ZERO runtime deps (Pigeon is codegen-only); `internal/likec4/testdata/bigbank.c4` (the 228-line canonical showcase) parses without fatal error.
  2. Parse failures on a malformed `.c4` file surface a `bigbank.c4:LINE:COL:`-style error reusing the existing `*parser.ParseError` shape (or compatible), NOT an opaque byte offset — Pigeon global state threads position so the user sees the offending source line.
  3. `views {}`, `deployments {}`, `global {}`, and other unrecognized top-level/element blocks parse without error and produce NO AST nodes — the explicit `UnknownBlock`/`UnknownStatement` recovery rules consume-and-discard them. This is the LOAD-BEARING PEG-03 tradeoff: recovery is designed in, not absent.
  4. AST golden tests on `bigbank.c4` + a minimal `.c4` fixture lock the typed node shape (`ast.File`, `ast.Specification`, `ast.Model`, `ast.Element`, `ast.Relationship`, `ast.Property`) including source line:column on every node — the typed handoff to the converter (Phase 35).

---

### Phase 35: Converter core (keystone)

**Goal**: A LikeC4 AST converts into a structurally-indistinguishable `*parser.Model` — flattened lexical `{}` nesting, absolute-peer relationships, mapped kinds, mapped properties. This is the keystone: it is where the "never fatal" contract, the `<->` two-link rule, the Humanize-bypass on `title`, and absolute-peer resolution all live. Do not under-budget it.
**Depends on**: Phase 34 (consumes the typed AST).
**Requirements**: CONV-01, CONV-02, CONV-03, CONV-04, CONV-05
**TDD note**: `convert_test.go` is written BEFORE `convert.go` is considered complete — it asserts the converter output is structurally identical to what `parser.ParseFile` would produce for an equivalent TOML model (the DX-02 cross-format proof is the integration gate, but unit-level fixtures gate this phase).
**Build-order steps covered**: 6 (warnings.go), 7 (kindmap.go), 8 (resolve.go), 9 (convert.go — the keystone), 10 (likec4.go ParseFile/Convert public entry), 11 (convert_test.go gate).
**Pitfalls baked into success criteria**: Pitfall 7 (`<->` two links, NOT `Arrow: bidirectional`), Pitfalls 4/5/14 (lexical-scope resolution diverges from `peer.Resolve` — emit absolute paths only), Pitfall 13 (Humanize trap on `title`).
**Success Criteria** (what must be TRUE):

  1. A nested LikeC4 element tree (`system { component { component } }`) converts to C4Drill's dotted-path subunit tree (`Unit.Subunits` keyed by last segment + `SubunitOrder` in source order), structurally identical to what `parser.ParseFile` produces for the equivalent TOML model.
  2. A LikeC4 `a <-> b` converts to TWO `model.Link` records (forward + reverse, `Mirror: false`) — preserving the HS-1 multiplicity contract (Pitfall 7); `this`/`it`/sourceless `->` resolve to absolute dotted paths at convert time, so downstream `peer.Resolve` is a no-op (Pitfalls 4/5/14); relationship kind (`-[async]->`) is concatenated into `Link.Technology` as preserved text. An unresolved reference is a HARD error (genuine corruption, not a dropped construct).
  3. Every LikeC4 element kind maps to a C4Drill `UnitType` via the three-tier fuzzy table (exact `person`/`db`/`queue`/`system`/`container`/`box` → substring `data`/`service` → fallback `box`); a kind used as a nesting parent promotes to `box` so subunits are legal; an unknown kind is NEVER a hard error (LikeC4 has no built-in kinds and no kind inheritance — this table is load-bearing).
  4. A LikeC4 element with `title`/`description`/`technology`/`link` populates `Unit.Name`/`Description`/`Technology`/`Reference` — with `title` set VERBATIM (bypassing `model.Humanize`, so `gRPC` stays `gRPC` not "Grpc"; Pitfall 13); a per-element `style { shape: cylinder }` override flips the `UnitType` (cylinder→db, queue→queue, browser/mobile→container); icon styles are silently dropped (project is on `dev/shapes-no-icons`).
  5. `Convert(*ast.File, io.Writer) (*parser.Model, error)` public entry chains with `Parse` via `likec4.ParseFile`; a minimal `.c4` round-trips end-to-end through Convert producing a model that passes the EXISTING validator unchanged.

---

### Phase 36: Integration & warnings

**Goal**: `.c4` inputs route to the converter at Stage 0 (one-block switch in `cmd/c4drill/root.go:runRoot`); deduplicated stderr warnings emit one line per construct TYPE; single-file `extend` merges cleanly. The hard TOML byte-identical regression contract is enforceable the moment this lands.
**Depends on**: Phase 35 (converter must produce a valid `*parser.Model`).
**Requirements**: INT-01, INT-02, INT-04
**Build-order steps covered**: 12 (cmd/c4drill/root.go Stage 0 switch).
**Hard contract**: The `.toml` code path is untouched — dispatch by extension happens BEFORE `parser.ParseFile` (Pitfall 6). Every existing `.toml` fixture parses byte-identical.
**Success Criteria** (what must be TRUE):

  1. Running `c4drill model.c4` (or `.likec4`) routes to `likec4.ParseFile` by extension; running `c4drill model.toml` routes to the existing `parser.ParseFile` — ZERO new CLI flags, the default arm of the switch is unchanged.
  2. Every existing `.toml` fixture (testdata/, cmd/c4drill/testdata/, skill/examples/) renders byte-identical (canonical-DOT golden, DI-1) to v1.10 — the INT-01 hard regression contract. The `.toml` code path is provably untouched (Pitfall 6).
  3. A `.c4` file containing `views {}`, `deployments {}`, `#tags`, `metadata {}`, icons, `style {}`, and relationship kinds emits ONE warning per construct TYPE to `os.Stderr` (not per occurrence — Pitfall 11); warnings survive `-o` redirection (stderr is independent of output files, NOT routed through `cmd.OutOrStderr()` which Cobra redirects).
  4. A `.c4` file using `extend <element> {}` (single-file) accumulates children/properties/tags into the element declared earlier in the SAME file (accumulator-merge, not overwrite); cross-file workspace `extend` (LikeC4's implicit multi-file merge, F-01) emits a one-line deferral warning rather than fataling.
  5. Pipeline no-op guards proven: for `.c4` input, `template.Expand` and `include.Resolve` hit their empty-input fast-paths (LikeC4 has no templates or includes) — the converter's emitted `*parser.Model` flows through the v1.10 pipeline unchanged.

---

### Phase 37: End-to-end proof & docs

**Goal**: The canonical 228-line `bigbank.c4` showcase renders end-to-end (the "render any model" proof); the converter's semantic correctness is locked against the native TOML parser via cross-format equivalence; users learn the feature from docs and a runnable fixture.
**Depends on**: Phase 36 (CLI dispatch must be wired).
**Requirements**: INT-03, DX-01, DX-02, DX-03
**Build-order steps covered**: 13 (cmd/c4drill/root_test.go CLI regression) + docs sweep.
**Success Criteria** (what must be TRUE):

  1. `testdata/bigbank.c4` (the 228-line canonical LikeC4 showcase — 10 kinds, 3 nesting levels, ~25 relationships, 8 dropped views) is committed; `c4drill testdata/bigbank.c4` renders without fatal error and produces valid DOT verified via `canonical.Canonical` — the INT-03 "render any model" proof.
  2. A small model authored in BOTH TOML and `.c4` produces canonical-identical DOT (the DX-02 cross-format equivalence proof, the analog of v1.10's XC-05 composed-vs-single-file proof) — locks the converter's semantic correctness against the native parser.
  3. `cmd/c4drill/root_test.go` CLI regression proves `.c4` end-to-end (parse → convert → validate → render → DOT AND SVG on disk) AND that `.toml` inputs render identically to v1.10 (DX-03 backward-compat guard).
  4. `skill/examples/10-likec4-bigbank.c4` runnable fixture committed; README gains a "LikeC4 Compatibility" section documenting extension routing, the supported subset, and the unsupported-with-warning set (matching v1.10's DOC-01/DOC-02/DOC-03 docs pattern).

---

## Coverage Validation

| REQ Group | Phase | Count |
|-----------|-------|-------|
| PEG-01, PEG-02, PEG-03, PEG-04 | Phase 34 | 4 |
| CONV-01, CONV-02, CONV-03, CONV-04, CONV-05 | Phase 35 | 5 |
| INT-01, INT-02, INT-04 | Phase 36 | 3 |
| INT-03, DX-01, DX-02, DX-03 | Phase 37 | 4 |
| **TOTAL** | | **16/16 ✓** |

Every PEG/CONV/INT/DX requirement maps to exactly one phase. No requirement is unmapped; no requirement is double-mapped.

## Progress

| Phase | Milestone | Plans Complete | Status | Completed |
|-------|-----------|----------------|--------|-----------|
| 28. Reference field | v1.10 | 1/1 | Complete | 2026-08-08 |
| 29. Optional name humanization | v1.10 | 2/2 | Complete | 2026-08-08 |
| 30. Relative-peer resolution | v1.10 | 2/2 | Complete | 2026-08-08 |
| 31. Template expansion | v1.10 | 2/2 | Complete | 2026-08-08 |
| 32. Include directive | v1.10 | 2/2 | Complete | 2026-08-08 |
| 33. Docs sweep + goldens | v1.10 | 4/4 | Complete | 2026-08-08 |
| 34. PEG grammar + AST foundation | v1.11 | 0/? | Pending | — |
| 35. Converter core (keystone) | v1.11 | 0/? | Pending | — |
| 36. Integration & warnings | v1.11 | 0/? | Pending | — |
| 37. End-to-end proof & docs | v1.11 | 0/? | Pending | — |
