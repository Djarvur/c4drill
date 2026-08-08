# Project Research Summary — v1.11 LikeC4 Compatibility Layer

**Project:** C4Drill (Go CLI rendering C4 diagrams from TOML)
**Milestone:** v1.11 LikeC4 Compatibility Layer (native-Go `.c4` parser → `*parser.Model`)
**Researched:** 2026-08-08 (PEG override 2026-08-09)
**Confidence:** HIGH
**Source docs:** [STACK.md](STACK.md), [FEATURES.md](FEATURES.md), [ARCHITECTURE-v1.11.md](ARCHITECTURE-v1.11.md), [PITFALLS.md](PITFALLS.md), [likec4-dsl-brief.md](likec4-dsl-brief.md)

> Supersedes the v1.10 "Model Composition" SUMMARY.md. The single load-bearing document the roadmapper and planner read. The four source docs are NOT re-read downstream. **The Pigeon (PEG) override in STACK.md/ARCHITECTURE-v1.11.md is BINDING and reflected throughout.**

---

## 1. Executive Summary

v1.11 adds a **Stage 0 converter** that accepts LikeC4 (`.c4` / `.likec4`) source and emits a structurally indistinguishable `*parser.Model`, so the entire v1.10 pipeline (`include.Resolve → template.Expand → peer.Resolve → Validate → view.Generate → graph.Build → render`) is reused verbatim. The converter is the ONLY code that knows LikeC4 was the source — no downstream package imports `internal/likec4`, and no `*parser.Model` field gains a "source format" tag. The north-star contract is "render any model, never fatal on unsupported constructs": `views {}`, `deployments {}`, icons, tags, and metadata are dropped with ONE deduplicated stderr warning per construct type.

**Parser technology — LOCKED: Pigeon (PEG) v1.3.0.** The original STACK.md body recommended hand-written recursive-descent; the user overrode this. Pigeon is codegen-only (`go generate ./internal/likec4/`), and the generated `grammar.go` is self-contained Go — so the zero-runtime-dependency story survives. The accepted tradeoff is error recovery: Pigeon's recovery (via `state.Throw()` and labeled failures) is weaker than hand-written panic-mode skip-to-`}`. The milestone's headline "drop unsupported block with warning" contract must therefore be implemented as **explicit recovery rules in the PEG grammar** (e.g. `UnknownBlock ← Keyword "{" BlockBody* "}"` that consumes and discards), NOT merely absent productions. The roadmap budgets time for recovery-rule design.

Key risks: (1) PEG recovery is harder to tune than hand-written — schedule a dedicated recovery-rule design spike; (2) the TOML byte-identical regression contract (every existing `.toml` fixture must render unchanged) is enforced by extension-dispatch BEFORE `parser.ParseFile`; (3) LikeC4 lexical-scope name resolution (`this`/`it`, sourceless `->`, cross-scope bubbling) DIVERGES from C4Drill's `peer.Resolve` ancestor-walk, so the converter must emit only absolute dotted paths and reduce `peer.Resolve` to a no-op for `.c4` input.

---

## 2. Stack Additions

### PEG locked, zero runtime deps, go-cmp test-only

| Component | Version | Purpose | Why |
|---|---|---|---|
| **Pigeon (PEG)** | **v1.3.0** (Sep 2024) | Parser generator — `grammar.peg` → `grammar.go` via `go generate` | LOCKED by user override. Declarative reviewable grammar; mature Go codegen; codegen-only (not imported by production code). Accepted tradeoff: recovery is harder than hand-written. |
| stdlib `slices` / `maps` | Go 1.26.1 | Dedup, ordering, contains | Already a project idiom (`parser.go:733`) |
| `google/go-cmp` | **v0.7.0** (2025-01-14) | `cmp.Diff` + `cmpopts.EquateEmpty` for deep-AST golden diffs | NEW **test-only** dep. Existing `testify/assert.Equal` cannot match readable deep-struct diffs. |

**Net new runtime dependencies: 0.** Pigeon output is self-contained Go — no CGO, no TypeScript/Node, no WASM, no grammar runtime. `testify` v1.11.1 stays as-is. Rejected: ANTLR4 (Java codegen toolchain, overkill), participle v2 (struct-tag model fights optional-`:`, flexible ordering, `this`/`it`), goyacc (legacy, 0 importers), `text/scanner` as lexer (now moot — PEG folds lexing in).

**PEG package layout** (`internal/likec4/`):
```
grammar.peg     # source of truth (declarative, reviewable, fuzzable)
grammar.go      # GENERATED via `go generate` — do not edit
ast.go          # AST node types returned by grammar action blocks
convert.go      # AST → *parser.Model (unchanged keystone)
kindmap.go      # kind string → model.UnitType fuzzy table
resolve.go      # this/it/sourceless -> + dotted-path resolution at convert time
warnings.go     # WarnCollector (dedup by construct type)
likec4.go       # public entry: ParseFile / Parse / Convert
```

---

## 3. Feature Classification

Condensed from FEATURES.md. North star: accept ANY valid LikeC4 source, drop unsupported with deduped warnings, NEVER crash.

**Table stakes (P1 — converter is DOA without these):**
- Top-level block routing — `specification`/`model`/`views`/`deployment`/`global` dispatch; multiple blocks of same type concatenate
- Lexical `{}` nesting → dotted-path subunits (parent-child is purely by braces; both `kind name {}` and `name = kind {}` reduce to same node)
- `->` relationship parsing (infix `a -> b 'title' "desc" "tech"`; single AND double quoted strings)
- `this`/`it`/sourceless `->` resolution to absolute paths at convert time
- Kind name → UnitType fuzzy mapping (exact → substring → `box` fallback; level-adjusted for db/queue generics)
- `title`/`description`/`technology` field mapping (with optional `:` after key)
- Comments (`//`, `/* */`) — folded into PEG grammar
- Graceful-degradation warning emitter (dedup by construct type)
- Extension-based routing (`.c4`/`.likec4` → converter; `.toml` → existing parser) — zero new flags
- Single-file `extend <element> {}` merging (accumulate per FQN, not overwrite)
- Anti-feature drops: `views {}`, `deployments {}`, icons (silent), tags/metadata — each with dedup warning

**Differentiators (P2 — add in v1.x when trigger fires):**
- Relationship kind preservation in `technology` (`-[async]->` → `"async: HTTPS"`) — LOW cost
- `<->` bidirectional as TWO `model.Link` entries with `Mirror: false` (PROJECT.md decision, NOT `Arrow: bidirectional`) — LOW cost
- `link <url>` → `reference` field (first http(s) link only; rides v1.10 📖; non-http dropped) — LOW cost
- Shape mapping override (`style.shape cylinder` → db; warn-once for no-analog shapes) — MEDIUM
- `specification` kind-default metadata inheritance (bigbank case: `technology` on kind) — MEDIUM

**Anti-features (explicitly DROPPED, do not re-litigate):**
- `views {}` DSL + predicates + auto-layout + rank + `navigateTo` — fights auto-generation philosophy; one `views-block` warning
- `deployment {}` / `deploymentNode` — no C4Drill analog; `deployment-block` warning
- Icons (`icon`/`iconColor`/`iconSize`/`iconPosition`) — project is on `dev/shapes-no-icons`; DELIBERATELY removed in v1.6; silent drop
- Tags/metadata visual rendering — no UI surface; `tags`/`metadata` warnings
- Custom kind inheritance/subtyping — CONFIRMED NONEXISTENT in LikeC4; fuzzy mapper handles implicitly
- File-level `import`/`#include` — CONFIRMED NONEXISTENT; multi-file rides v1.10 `[[include]]` if ever added

---

## 4. Architecture Integration

Stage 0 routing by file extension in `cmd/c4drill/root.go` (one-block `switch`): `.c4`/`.likec4` → `likec4.ParseFile(path, stderr)`; everything else → existing `parser.ParseFile` unchanged. The converter emits exactly the `*parser.Model` type — no DTO, no tagged union. Downstream `peer.Resolve` becomes a no-op for `.c4` output (converter pre-resolves all peers to absolute dotted paths); `template.Expand` and `include.Resolve` hit their empty-input fast-paths (LikeC4 has no templates or includes).

**Two-phase public API (Parse + Convert, `ParseFile` chains):**
```go
func ParseFile(path string, w io.Writer) (*parser.Model, error)  // chains Parse+Convert
func Parse(src []byte) (*ast.File, error)                         // Pigeon-generated, pure
func Convert(f *ast.File, w io.Writer) (*parser.Model, error)    // AST → Model, w receives warnings
```
The split is mandatory for testability: PEG/AST golden tests must not require assembling a full `*parser.Model`.

**Warning channel:** `WarnCollector` with `seen map[string]bool` keyed by construct TYPE (`deployment-block`, `views-block`, `metadata`, `tags`, `relationship-kind`, `style-block`, `element-extra-links`, …). Routed to `os.Stderr` directly (NOT `cmd.OutOrStderr()` — Cobra redirects that). One line per type per file; survives `-o` redirection. Required by ALL anti-features.

**Major components:**
1. **`grammar.peg` + generated `grammar.go`** — lexes+parses LikeC4 → `*ast.File`. Recovery rules for unsupported blocks are explicit productions, not absent grammar.
2. **`convert.go`** — keystone: walks AST element tree carrying `parentPath`, flattens braces → dotted `Subunits`/`SubunitOrder`, resolves `this`/`it`/sourceless `->` and bare names to absolute paths, maps kinds, drops anti-features, dedups warnings.
3. **`likec4.go`** — thin public entry; chains Parse+Convert.
4. **`cmd/c4drill/root.go`** — Stage 0 extension-dispatch switch (one new block).

---

## 5. Watch Out For

Top pitfalls condensed from PITFALLS.md (16 total); PEG recovery tradeoff added from STACK.md.

1. **PEG recovery is harder than hand-written** — the "drop unsupported block with warning" contract must be explicit recovery rules (`UnknownBlock ← Keyword "{" BlockBody* "}"`), NOT absent productions. Budget a recovery-rule design spike. Pigeon `state.Throw()` + labeled failures are the mechanism.
2. **`<->` MUST be two links with `Mirror: false`** (Pitfall 7) — NOT `Arrow: bidirectional`. PROJECT.md decision; validator multiplicity logic and unexported `Link.Mirror` (HS-1 concern) assume this. Gate with a bidirectional-edge regression test.
3. **`this`/`it`/sourceless `->` + cross-scope bubbling resolve to the WRONG scope** (Pitfalls 4, 5, 14) — never defer to `peer.Resolve` (its D-13 ancestor-walk diverges from LikeC4's lexical bubbling). Converter builds full FQN index, two-pass, emits ONLY absolute dotted paths; `peer.Resolve` becomes a no-op. A miss is a hard error (genuine corruption).
4. **TOML byte-identical regression** (Pitfall 6) — dispatch by extension BEFORE `parser.ParseFile`; converter must NOT modify `parser.Model`/`model.Unit`/`model.Link` struct defs; canonical-DOT goldens are the gate.
5. **Humanize title trap** (Pitfall 13) — always set `Unit.Name` from LikeC4 title verbatim (`gRPC` stays `gRPC`); never leave `Name` empty for the parser to humanize (no acronym preservation → "Grpc").
6. **`->` is grammar-ambiguous** (Pitfall 1) — same `->` means relationship (in `model`/element-body) vs. view predicate (in `views`/`view-body`, must DROP) vs. incoming-edge wildcard (sourceless). Carry a parser context stack; emit links only in `model`/`element-body`.
7. **Optional `:` and quote styles** (Pitfall 2) — `description "x"` AND `description: "x"` both valid (bigbank mixes them); single `'...'`, double `"..."`, AND triple `'''...'''`/`"""..."""` must all tokenize as one string category.
8. **Absolute-peer contract** (Pitfall 14) — converter emits only absolute dotted paths; `peer.Resolve`'s D-16 gate (`strings.Contains(peer, ".")`) short-circuits them untouched.
9. **Warning UX** (Pitfall 11) — dedup per TYPE not per occurrence; route to `os.Stderr` directly (never stdout — `-f dot` writes DOT to stdout); must survive `-o` redirection.
10. **Source-line tracking** (Pitfall 12) — every AST node carries its `.c4` source line; converter errors embed original line, never generated-model line; reuse `*parser.ParseError` shape.

---

## 6. Single Most Relevant Borrowed Idiom / Decision Per Area

| Area | Borrowed idiom / decision | Source |
|---|---|---|
| Parser tech | **Pigeon (PEG) v1.3.0 — codegen-only, generated parser is self-contained Go.** Recovery via `state.Throw()` + explicit recovery rules, not absent productions. | STACK.md addendum |
| AST-golden testing | `google/go-cmp` `cmp.Diff` + `cmpopts.EquateEmpty` for deep-struct equality | STACK.md (test-only dep) |
| Scope/name resolution | **Two-pass: build FQN index first, resolve references second; emit only absolute paths.** Mirrors JS lexical-scope + hoisting; do NOT rely on `peer.Resolve`. | PITFALLS.md #5, #14 |
| Flattening | `pathTracker` `(parentFullDottedPath, childName)` collision detection; `SubunitOrder` in source order | `template/expand.go` pattern, PITFALLS.md #9 |
| Kind mapping | Three-tier fuzzy table (exact → substring contains → `box` fallback) + `inferGenericType` level-adjustment for db/queue generics | ARCHITECTURE-v1.11 kindmap.go |
| Name field | LikeC4 `title` → `Unit.Name` verbatim; omit only to let `model.Humanize` derive from identifier (v1.10 ERGO-03) — but ONLY when no title | PITFALLS.md #13 |
| Warnings | `WarnCollector` with `seen map[string]bool` keyed by construct TYPE; `os.Stderr` direct; one line per type per file | ARCHITECTURE-v1.11 warnings.go |
| Pipeline convergence | Emit structurally-indistinguishable `*parser.Model`; reuse `include`/`template`/`peer` empty-input fast-paths as no-ops | ARCHITECTURE-v1.11 Pattern 2 |
| TOML safety | Extension-dispatch BEFORE `parser.ParseFile`; converter never touches shared struct defs | PITFALLS.md #6 |
| Test corpus | `testdata/bigbank.c4` (228 lines, 10 kinds, 3 nesting levels, 25 relationships, 8 views) as the single regression gate | FEATURES.md, DSL brief §10 |

---

## 7. Suggested Phase Breakdown

Each phase is independently shippable; converter is a no-op for `.toml` inputs (TOML never enters the new code path).

| Phase | Name | Goal | Depends on |
|---|---|---|---|
| **P1** | **PEG grammar + recovery-rule design spike** | `grammar.peg` covering `specification`/`model`/element bodies/`->`; explicit `UnknownBlock`-style recovery rules; generated `grammar.go`; `ast.go`; `Parse` API; AST golden tests on `bigbank.c4`. | None — de-risk PEG recovery first. |
| **P2** | **Convert keystone + kind map + name resolution** | `convert.go`, `kindmap.go`, `resolve.go`, `warnings.go`; `Convert` API; absolute-peer emission; `<->` → two links; kind fuzzy map; `extend` merge. | P1 (consumes AST). |
| **P3** | **Public entry + CLI wiring + integration regression** | `likec4.go` (`ParseFile`); `root.go` Stage 0 switch; canonical-DOT regression on `bigbank.c4`; TOML byte-identical goldens; warning stderr/`-o` tests; pipeline no-op tests for `template.Expand`/`include.Resolve`. | P2. |

**Phase ordering rationale:** Grammar + recovery first because PEG recovery is the new risk surface; if it can't drop `views {}` cleanly with a warning, the design must change before convert is built. Convert before CLI wiring because convert is the integration keystone emitting `*parser.Model`; it must be gated on `bigbank.c4` before the dispatch switch touches `root.go`. CLI wiring last (one-block diff) so the TOML byte-identical contract is enforceable the moment dispatch lands.

### Research Flags
- **P1 (PEG recovery):** Pigeon recovery semantics need a focused spike — `state.Throw`, labeled failures, `UnknownBlock`-style consuming productions. Schedule extra time.
- **P2 (scope bubbling):** LikeC4 lexical-scope + hoisting resolution is the hardest sub-component (Pitfall 5, HIGH recovery cost if broken). Two-pass design mandatory.

### Standard Patterns (skip research-phase)
- P3 CLI wiring — well-documented Cobra + `os.Stderr` idiom; one-block diff.

---

## 8. Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack (PEG locked) | HIGH | User decision; Pigeon v1.3.0 verified on pkg.go.dev; zero-runtime-dep story confirmed |
| Features | HIGH | Grounded in likec4-dsl-brief.md + bigbank.c4 canonical example + `internal/model/unit.go` |
| Architecture | HIGH | Reuses v1.10 pipeline verbatim; Stage 0 dispatch is one-block diff; converter/kindmap/resolve/warnings design unchanged by PEG override |
| Pitfalls | HIGH | 16 pitfalls grounded in repo source (`parser.go`, `unit.go`, `link.go`, `peer/resolve.go`, `include/merge.go`, `root.go`) |

**Overall confidence:** HIGH

### Gaps to Address
- **PEG recovery-rule design:** Pigeon's `state.Throw`/labeled-failure recovery for the "drop unsupported block" contract needs a P1 spike to validate the pattern works for `views {}` / `deployment {}` / nested unknown blocks. Fallback: grammar productions that consume-and-discard plus a `WarnCollector` call in the action block.
- **Cross-scope bubbling edge cases:** Two-pass FQN resolution handles common cases (bigbank); pathological ambiguities (same short name in two sibling scopes) may need a hard-error policy decision during P2.
- **`extend relationship` form:** `extend a -> b {}` — drop-with-warning policy documented but grammar production needs validation.

---

## 9. Sources

### Primary (HIGH confidence)
- `.planning/research/STACK.md` — Go parser/converter stack; **PEG override addendum is BINDING**
- `.planning/research/FEATURES.md` — LikeC4→C4Drill feature translation landscape
- `.planning/research/ARCHITECTURE-v1.11.md` — Stage 0 integration architecture; **PEG override notice in header**; converter/kindmap/resolve/warnings design unchanged
- `.planning/research/PITFALLS.md` — 16 critical pitfalls grounded in repo source
- `.planning/research/likec4-dsl-brief.md` — DSL grammar reference (§1-10, background)
- `internal/parser/parser.go`, `errors.go`, `parser_test.go` — existing TOML pipeline, `ParseError` shape, `inferGenericType`, Humanize hook
- `internal/model/unit.go`, `link.go`, `humanize.go` — C4Drill model types, `Subunits`/`SubunitOrder`, unexported `Link.Mirror`
- `internal/peer/resolve.go` — D-13/14/15/16 ancestor-walk, absolute-path gate
- `internal/include/merge.go`, `internal/template/expand.go` — empty-input fast-paths, `pathTracker` collision pattern, HS-1 Mirror preservation
- `cmd/c4drill/root.go` — Stage 0 dispatch site (~`:117`), `cmd.OutOrStderr()` pattern
- [pkg.go.dev — mna/pigeon](https://pkg.go.dev/github.com/mna/pigeon) — v1.3.0, codegen-only, failure-label recovery
- [pkg.go.dev — google/go-cmp](https://pkg.go.dev/github.com/google/go-cmp) — v0.7.0 (2025-01-14), test-only

### Secondary (MEDIUM confidence)
- likec4.dev DSL docs (specification/model/relationships/views/styling/deployment/extend)
- `github.com/likec4/likec4/apps/playground/src/examples/bigbank/bigbank.c4` — 228-line canonical real-world example

### Tertiary (LOW confidence)
- Web search: no native-Go LikeC4 or C4-family DSL parser exists (official is TS/Langium; structurizr Go hits are TS/Chevrotain) — high confidence in build-from-scratch

---
*Synthesis for: C4Drill v1.11 LikeC4 Compatibility Layer milestone*
*Researched: 2026-08-08 (PEG override 2026-08-09)*
*Ready for roadmapper + planner: yes*
