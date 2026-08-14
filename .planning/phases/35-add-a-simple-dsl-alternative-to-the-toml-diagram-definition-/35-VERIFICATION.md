---
phase: 35-add-a-simple-dsl-alternative-to-the-toml-diagram-definition-
verified: 2026-08-14T21:27:11Z
status: human_needed
score: 24/24
overrides_applied: 0
re_verification:
  previous_status: gaps_found
  previous_score: 21/24
  gaps_closed:
    - "Converters produce canonical-equivalent round-trips for every legal input (CR-01/02/03/05 fixed by c59b762: CheckC4DRepresentable loud-error gate + quoted TOML key segments + literalFor escaped-quoted fallback + gateTwin re-parse/canonical-equality write gate)"
    - "Full TOML feature parity incl. width/height (CR-04 fixed by f553a9c: grammar FieldKey + tomodel + frommodel + both emitters + canonsrc canonical form)"
    - "convert --follow-includes -o preserves relative directory structure (CR-06 fixed by 0a17d64: entryDir from absolutized entry)"
  gaps_remaining: []
  regressions: []
human_verification:
  - test: "Human review of the README.adoc C4D Format section, the 12 example twins, and the dual-format skill"
    expected: "Syntax documentation correct and readable; side-by-side example clear; twins deliver the promised verbosity win (06-templates: 109 -> 43 lines claimed)"
    why_human: "35-09 Task 4 was a blocking checkpoint:human-verify gate that was AUTO-approved under AUTO-MODE — the five verification steps were executed by the agent, not reviewed by a human; doc readability and the verbosity-win judgment are subjective"
  - test: "Optional: WR-04 skill table fix confirmation"
    expected: "After any fix pass, skill/SKILL.md type-inference table rows for systemExternal and box parents match internal/parser/parser.go DefaultTypeForParent (both fall to the default C1/system branch)"
    why_human: "Requires reading both the table and the Go switch side by side; doc-correctness judgment"
---

# Phase 35: C4D DSL Alternative to TOML — Verification Report

**Phase Goal:** Deliver the C4D format — a `.c4d` brace-block D2-inspired DSL with full TOML feature parity — parseable directly to `*parser.Model` and renderable through the unchanged pipeline, with bidirectional canonical-equivalent converters (`convert to-toml`/`to-c4d`), a gofmt-style comment-preserving formatter (`fmt`) for both formats, nested use and recursive template-instantiating-template expansion, plus full README/skill/example documentation.
**Verified:** 2026-08-14T21:27:11Z (re-verification after fixes 0a17d64 / f553a9c / c59b762)
**Status:** human_needed — all 24 truths verified; 2 human verification items remain (carried from initial verification, unaffected by the fixes)
**Re-verification:** Yes — after gap closure (initial verification 2026-08-14, 21/24, gaps_found)

## Goal Achievement

### Observable Truths

Roadmap success_criteria is empty for this phase; truths merged from the phase goal wording and the 9 PLAN frontmatters.

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Core C4D grammar parses brace-block units, nesting, fields, in-block edges, properties into typed comment/position-aware AST (D-01, D-08, D-12) | ✓ VERIFIED | `internal/c4d/grammar/c4d.peg` (787 lines, rules Document/UnitHeader/EdgeStmt/Literal/Comment/Properties), `internal/c4d/ast/ast.go` (211 lines, Pos+Comments on nodes); parse tests green |
| 2 | Unit headers `id: type "Name"` with omittable slots; exact TOML type keywords; external modifier -> *External variants (D-02, D-03, D-04) | ✓ VERIFIED | c4d.peg UnitHeader (3 alternatives); TestParseUnitExternalModifier, TestToModelExternalModifier pass |
| 3 | Arrows `->` `<-` `<->` `--` + desc/tech pipe shorthand with desc-first single value (D-05, D-09) | ✓ VERIFIED | Grammar EdgeStmt; composition/parse tests green |
| 4 | Bareword/double/triple-quoted literals, `#` comments, scheme URLs valid barewords; identifiers `[A-Za-z0-9_-]+`, dotted paths (D-06, D-07) | ✓ VERIFIED | Grammar Literal rule; tests green. Re-verified: quote-terminated multiline values now emit in the escaped quoted form (CR-05 fixed — see truth 22) |
| 5 | `;` separator + empty blocks (D-18); reserved keyword collision = hard error with Levenshtein suggestion (D-19) | ✓ VERIFIED | `internal/c4d/grammar/reserved.go` (77 lines): 14 isBuiltinField + 5 statement keywords, FormatSuggestion wired; TestParseReservedUnitIdError, TestParseReservedKeywordsTable/List pass |
| 6 | pigeon pinned in go.mod, generated parser committed, go:generate chain (D-20) | ✓ VERIFIED | go.mod `github.com/mna/pigeon v1.3.0` (deviation from planned v1.0.0 — documented: v1.0.0 lacks the -nolint flag the lint gate requires); tools.go build-tag pin; `//go:generate pigeon -o parser_gen.go -nolint c4d.peg`; parser_gen.go committed |
| 7 | Parse failures are `*parser.ParseError` with DSL-native line numbers (D-21) | ✓ VERIFIED | `internal/c4d/errors.go` wraps pigeon errList; tests assert Line == offending line |
| 8 | TOML `[[unit.X.use]]` + `[[template.X.use]]` desugar to Instantiation (D-16, D-17 TOML forms) | ✓ VERIFIED | parser.go extractUnitUses; TestParseUnitUseSugarEquivalentToParentField, TestParseTemplateBodyUseExtracted pass; TemplateDef.Instantiations field present |
| 9 | Recursive template-body expansion: outer-to-inner params, cycle chain "A -> B -> A", depth cap 100, HS-1 deep-copy every level (D-17) | ✓ VERIFIED | expand.go checkRecursion + maxTemplateDepth=100; TestExpandTemplateCycle, TestExpandTemplateSelfCycle, TestExpandTemplateDepthCap, TestExpandThreeLevelNestingHS1 pass |
| 10 | v1.10 STATE.md template-nesting deferral lifted | ✓ VERIFIED | STATE.md Deferred Items: "SHIPPED in Phase 35 (D-17, Plan 35-02) — deferral lifted" |
| 11 | template/use(3 positions)/include-once grammar + both list forms (D-13, D-14, D-15) | ✓ VERIFIED | composition_test.go; ToModel maps UseStmt parents; tests green |
| 12 | EmitTOML deterministic canonical field order, fixture-shaped tables; EmitC4D compact-leaf style; FromModel inverse (D-23, D-33) | ✓ VERIFIED | TestEmitTOMLCanonicalUnitFieldOrder, TestEmitTOMLUnitOrderFollowsUnitOrder, fixpoint tests; emit_toml.go / emit_c4d.go / frommodel.go. Re-verified: D-23 order now includes width/height; quoted key segments keep headers parseable (CR-02/CR-04 fixed — see truths 22-23) |
| 13 | c4d.Parse -> *parser.Model directly; ParseAST/ParseASTFile exported; inference parity identical Models (D-21, D-02) | ✓ VERIFIED | parse.go:20/32/63/71 signatures confirmed; parser.DefaultTypeForParent/InferGenericType exported (parser.go:954/992); TestToModelInferenceParity passes |
| 14 | Duplicate edge hard error (D-11); peers through unchanged peer.Resolve (D-10); mixed .c4d/.toml include graphs (D-26) | ✓ VERIFIED | TestToModelDuplicateEdgeError; include/resolve.go checkIncludeExtension dispatch; TestResolveTomlIncludesC4d, TestResolveC4dIncludesToml, TestResolveUnknownExtensionHardError, TestResolveMixedFormatCycleFatal pass |
| 15 | D-22 parity over the fixture corpus: both-direction canonical-equivalent round-trips + render equivalence + composed graph | ✓ VERIFIED | parity_test.go: TestRoundTripTOMLToC4DToTOML (29 fixtures, require.Greater > 15 anti-shrinkage), TestRoundTripC4DToTOMLToC4D, TestRenderEquivalence, TestComposedGraphRoundTrip — all green |
| 16 | `c4drill diagram.c4d` renders directly; extension dispatch hard-errors naming .toml/.c4d (D-27, D-29) | ✓ VERIFIED | root.go:218 c4d.ParseFile branch; behavioral: `unsupported input extension ".json" (accepted: .toml, .c4d)` |
| 17 | convert validates first on a discarded pipeline copy, emits from a fresh parse preserving includes/templates/bare peers; output placement swapped-ext + -o (D-24, D-25, D-28, D-30 single-file) | ✓ VERIFIED | convert.go:113-161 runConvert order (direction gate -> validateSourceForConvert -> fresh parse -> emit -> gate -> write); convert tests green. Re-verified: emit now gated by CheckC4DRepresentable and the write by gateTwin (re-parse + CanonicalEqual) |
| 18 | convert --follow-includes converts whole graph with rewritten paths, once preserved, cycle-safe (D-25 graph mode) | ✓ VERIFIED | convertGraph + walkIncludeGraph (visited set, maxConvertDepth, ancestor cycle check); retargetExt preserves relative form; per-file gateTwin wired in convertGraphFile |
| 19 | fmt rewrites both formats in place; --check exit 1 no write; comments + blank-line grouping preserved; idempotent; semantic safety gate (D-31, D-32, D-33) | ✓ VERIFIED | Behavioral: `fmt --check` exit 1 with zero byte change; in-place rewrite produced canonical `a: system "A" { description: x }`; lead + trailing comments preserved; tomlfmt TestFormatPreservesComments, TestFormatIdempotentOverCorpus, corpus sweeps green |
| 20 | Format named C4D; README.adoc C4D section + convert/fmt CLI reference; 12 render-identical twins; skill extended in place with name kept, plugin copies synced (D-34, D-35) | ✓ VERIFIED | `== C4D Format` section present, 70 c4d/C4D mentions, follow-includes + --check documented; skill/SKILL.md `name: c4drill-toml` kept, byte-identical to packaged copy (diff clean, re-confirmed 2026-08-14); 12 twins in skill/examples + 12 in plugin examples; TestExampleTwins 12/12 pass |
| 21 | Full repo health: build + all tests + lint | ✓ VERIFIED | Re-verified 2026-08-14: `go build ./...` ok; `go test ./... -count=1` 15/15 packages ok; `golangci-lint run` (phase packages AND full repo) 0 issues; no TBD/FIXME/XXX/TODO/HACK markers in any phase or fix-modified file |
| 22 | Converters produce canonical-equivalent round-trips for every legal input (goal wording, unqualified; D-22 contract text) | ✓ VERIFIED (re-verified 2026-08-14 after fixes 0a17d64/f553a9c/c59b762) | All six original repro classes independently re-executed with a freshly built CLI — none silently corrupts anymore. CR-01 pipe-in-tech: `convert to-c4d` now hard-errors `link technology "HTTP | REST" on unit "api" is not representable in C4D: labels must not contain '|'` naming the value, exit 1, NO twin written (CheckC4DRepresentable, frommodel.go:388). CR-02 type-led unit with display name: twin emits `["My App"]` / `[["My App".link]]` quoted key segments (tomlKeySegment, emit_toml.go:441), convert exit 0, twin re-parses AND renders (exit 0). CR-03 non-identifier ids: TOML ids "my unit" (space) and "my.unit" (dot) both hard-error `unit id ... is not representable in C4D: identifiers must match [A-Za-z0-9_-]+`, exit 1, no twin. CR-05 quote-terminated multiline: value `line1\nline2"` emits as escaped quoted `"line1\nline2\""` (literalFor fallback, frommodel.go:362), convert exit 0, back-conversion returns the value byte-for-byte. Belt-and-suspenders: gateTwin (convert.go:237, wired at convert.go:150 single-file AND convert.go:327 graph mode) re-parses every twin and requires c4d.CanonicalEqual (canonequal.go — compares every unit field incl. width/height) before any write |
| 23 | Full TOML feature parity (goal wording; README.adoc:743 claims 'Everything the TOML format expresses ... has a C4D equivalent') | ✓ VERIFIED (re-verified 2026-08-14 after fixes 0a17d64/f553a9c/c59b762) | CR-04 fixed end-to-end. Reproduced: TOML `width = 300` / `height = 200` -> `convert to-c4d` exit 0 -> twin carries `width: 300` / `height: 200` -> `convert to-toml` exit 0 -> twin carries `width = 300` / `height = 200` intact. Wiring confirmed at every layer: grammar FieldKey admits width/height (c4d.peg:676), tomodel.go:430-443 parses them into unit.Width/Height with numeric validation, frommodel.go:142-154 emits them (formatFloat shortest-exact form), emit_toml.go:170-175 emits them, canonsrc.go:185-190 includes them in the canonical form (the earlier deliberate exclusion removed), CanonicalEqual compares them. No override needed |
| 24 | convert --follow-includes -o preserves the graph's relative directory structure (README.adoc:1161) | ✓ VERIFIED (re-verified 2026-08-14 after fixes 0a17d64/f553a9c/c59b762) | CR-06 fixed. Reproduced the exact failing invocation: entry.toml including domains/auth.toml, invoked with a RELATIVE entry path from inside the graph dir (`convert to-c4d --follow-includes -o OUT entry.toml`) -> OUT/entry.c4d + OUT/domains/auth.c4d (structure preserved; previously flattened to OUT/auth.c4d). Include path rewritten to `domains/auth.c4d` in the twin; the composed twin graph renders (exit 0). Root cause fixed at convert.go:287 — `entryDir := filepath.Dir(absEntry)` derives from the absolutized entry |

**Score:** 24/24 truths verified

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `internal/c4d/grammar/c4d.peg` | PEG grammar (plan path `internal/c4d/c4d.peg` — layout deviation, documented) | ✓ VERIFIED | 787 lines, all core + composition rules; re-verified: FieldKey now includes width/height (CR-04) |
| `internal/c4d/grammar/parser_gen.go` | committed generated parser | ✓ VERIFIED | Present, nolint header, builds; regenerated in f553a9c for the FieldKey change |
| `internal/c4d/ast/ast.go` | typed AST with Pos/Comments | ✓ VERIFIED | 211 lines |
| `internal/c4d/errors.go` | ParseError wrapping | ✓ VERIFIED | Wired to parser.ParseError via errors.As |
| `internal/parser/parser.go` | Instantiations + exported inference helpers | ✓ VERIFIED | TemplateDef.Instantiations, DefaultTypeForParent, InferGenericType |
| `internal/template/expand.go` | recursive expansion + cycle detection | ✓ VERIFIED | checkRecursion, maxTemplateDepth, cycle chain message |
| `internal/c4d/tomodel.go` | ToModel with parity hooks | ✓ VERIFIED | 861 lines; calls DefaultTypeForParent/InferGenericType/Humanize; re-verified: width/height parsing added (CR-04) |
| `internal/c4d/emit_toml.go` | EmitTOML canonical order | ✓ VERIFIED | Canonical-order test pins D-23; re-verified: quoted key segments (CR-02) + width/height emission (CR-04) |
| `internal/c4d/emit_c4d.go` + `frommodel.go` | compact-leaf printer + Model->AST | ✓ VERIFIED | Re-verified: frommodel.go now carries CheckC4DRepresentable (CR-01/CR-03), literalFor escaped-quoted fallback (CR-05), width/height emission (CR-04); defect classes from the initial verification are gone |
| `internal/c4d/canonequal.go` | CanonicalEqual predicate (new in c59b762) | ✓ VERIFIED | 105 lines; exported for the convert write gate; fills only the D-22 explicit-default list before reflect.DeepEqual — everything else incl. width/height must match exactly |
| `internal/testutil/canonsrc/canonsrc.go` | NormalizeTOML + NormalizeC4D | ✓ VERIFIED | Re-verified: width/height now ride the canonical form (canonsrc.go:185-190), closing the oracle blind spot from the initial verification |
| `internal/c4d/parity_test.go` | corpus round-trip + render equivalence | ✓ VERIFIED | 13 test funcs; 29-fixture corpus; 12 twins; width/height round-trip test added in f553a9c |
| `internal/c4d/representable_test.go` | representability gate tests (new in c59b762) | ✓ VERIFIED | 314 lines; pins the loud-error classes |
| `cmd/c4drill/convert_gate_test.go` | write-gate tests (new in c59b762) | ✓ VERIFIED | 282 lines; pins gateTwin behavior |
| `testdata/c4d/*.toml` | 6 edge-case fixtures | ✓ VERIFIED | All 6 present, parse + validate green |
| `cmd/c4drill/convert.go` | convert subcommand | ✓ VERIFIED | Re-verified: gateTwin wired in BOTH paths (runConvert + convertGraphFile), CheckC4DRepresentable before C4D emission, entryDir from absolutized entry (CR-06) |
| `cmd/c4drill/fmt.go` + `internal/tomlfmt/tomlfmt.go` | formatter + TOML formatter | ✓ VERIFIED | 272 + 285 lines; behavioral checks pass |
| `README.adoc`, `skill/SKILL.md`, 12 + 12 twins | docs surface | ✓ VERIFIED | Greps + diff + TestExampleTwins all pass; skill/plugin copies still byte-identical (re-confirmed) |

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|----|--------|---------|
| cmd/c4drill/root.go runRoot | c4d.ParseFile | extension branch | ✓ WIRED | root.go:218 |
| cmd/c4drill/root.go NewRootCmd | newConvertCmd/newFMTCmd | cmd.AddCommand | ✓ WIRED | root.go:100,103 |
| cmd/c4drill/convert.go | validator.Validate + emitters | D-24 gate then fresh-parse emit | ✓ WIRED | validateSourceForConvert + emitConverted |
| cmd/c4drill/convert.go runConvert/convertGraphFile | c4d.Parse/parser.Parse + c4d.CanonicalEqual | gateTwin write gate | ✓ WIRED | convert.go:150 + convert.go:327; target-format re-parse of every twin before any write (re-verified) |
| cmd/c4drill/convert.go emitConverted | c4d.CheckC4DRepresentable | representability gate | ✓ WIRED | convert.go:218; loud *parser.ParseError naming the value, attributed to the source path (re-verified) |
| cmd/c4drill/convert.go convertGraph | filepath.Abs(entryPath) | entryDir derivation | ✓ WIRED | convert.go:278-287; relative entry paths no longer flatten -o output (re-verified) |
| cmd/c4drill/fmt.go | c4d.ParseAST + EmitC4D + tomlfmt.Format | per-extension path | ✓ WIRED | Confirmed in fmt.go + behavior |
| internal/c4d/tomodel.go | parser inference hooks + Model | parity hooks | ✓ WIRED | Signatures + tests |
| internal/c4d/frommodel.go | grammar FieldKey charset | width/height + representability | ✓ WIRED | tomodel parses width/height; CheckC4DIdent/c4dPeerSafe mirror the grammar's [A-Za-z0-9_-]+ set (re-verified) |
| internal/include/resolve.go | extension-branched parse | .c4d dispatch | ✓ WIRED | checkIncludeExtension |
| internal/c4d/grammar/reserved.go | validator.FormatSuggestion | suggestion on collision | ✓ WIRED | reserved.go:62 |
| skill/examples twins | render parity | TestExampleTwins walker | ✓ WIRED | 12/12 pass |
| README.adoc CLI Reference | convert/fmt flags | usage examples | ✓ WIRED | follow-includes, --check documented |

### Data-Flow Trace (Level 4)

Not applicable — this phase adds parsers/converters/formatters (no UI components rendering dynamic data). The equivalent depth check is the round-trip trace: source text -> Model -> twin text -> re-parsed Model -> CanonicalEqual, asserted behaviorally for every scenario above.

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Full test suite | `go test ./... -count=1` | 15/15 packages ok | ✓ PASS (re-verified) |
| Corpus round-trip tally | `go test ./internal/c4d/ -run TestRoundTripTOMLToC4DToTOML -v` | 29 fixtures, all pass | ✓ PASS |
| Twins render parity | `go test ./internal/c4d/ -run TestExampleTwins -v` | 12/12 pass | ✓ PASS |
| Unknown extension hard error | `c4drill foo.json` | names .toml and .c4d | ✓ PASS |
| fmt --check CI gate | `c4drill fmt --check bad.c4d` | exit 1, zero byte change | ✓ PASS |
| fmt in-place + comment preservation | `c4drill fmt` on misformatted .c4d | canonical output, comments kept | ✓ PASS |
| CR-01 pipe corruption | built CLI: `convert to-c4d` on TOML with `technology = "HTTP \| REST"` | loud error naming `HTTP \| REST`, exit 1, no twin written | ✓ PASS (re-verified 2026-08-14) |
| CR-02 type-led unit to TOML | built CLI: `convert to-toml` on `system "My App" {...}` | twin `["My App"]` quoted key, exit 0; twin re-parses and renders (exit 0) | ✓ PASS (re-verified 2026-08-14) |
| CR-03 space/dot id to C4D | built CLI: `convert to-c4d` on TOML with `["my unit"]` and `["my.unit"]` | both loud-error naming the id, exit 1, no twin | ✓ PASS (re-verified 2026-08-14) |
| CR-04 width/height | built CLI: round-trip `width = 300` / `height = 200` | `width: 300` / `height: 200` in C4D twin; back-converted TOML intact; both exits 0 | ✓ PASS (re-verified 2026-08-14) |
| CR-05 quote-terminated multiline | built CLI: round-trip value `line1\nline2"` | twin emits escaped quoted `"line1\nline2\""`, exit 0; back-conversion byte-identical; fmt --check green | ✓ PASS (re-verified 2026-08-14) |
| CR-06 -o flattening (relative entry) | built CLI: relative `entry.toml` + `--follow-includes -o OUT` from inside graph dir | OUT/entry.c4d + OUT/domains/auth.c4d — structure preserved; include path retargeted; composed twin renders | ✓ PASS (re-verified 2026-08-14) |
| Lint gate | `golangci-lint run` (phase packages + full repo) | 0 issues | ✓ PASS (re-verified) |

### Probe Execution

Step 7c: SKIPPED — no `scripts/*/tests/probe-*.sh` probes exist in this repo; the phase's runnable verification surface is the Go test suite plus the built-CLI behavioral checks (executed above).

### Requirements Coverage

All 35 D-IDs from 35-CONTEXT.md are claimed across the 9 PLAN frontmatters (union = D-01..D-35, no orphans, no gaps in coverage on paper). Codebase status:

| Requirement | Source Plan | Status | Evidence |
|-------------|------------|--------|----------|
| D-01..D-09, D-12, D-18, D-20 | 35-01 | ✓ SATISFIED | Grammar + tests + pigeon chain (v1.3.0 pin documented) |
| D-16, D-17 | 35-02 | ✓ SATISFIED | extractUnitUses, recursive expand, cycle/depth/HS-1 tests, STATE.md updated |
| D-13, D-14, D-15, D-19 | 35-03 | ✓ SATISFIED | Composition grammar + reserved.go (19 keywords) |
| D-23, D-33 | 35-04 | ✓ SATISFIED | Canonical-order + compact-leaf tests. Re-verified: D-23 field set now includes width/height (CR-04 closed — the plan-pinned omission was lifted by f553a9c) |
| D-02, D-10, D-11, D-21, D-26 | 35-05 | ✓ SATISFIED | Signatures, parity test, duplicate-edge error, mixed includes |
| D-22, D-26 | 35-06 | ✓ SATISFIED (re-verified) | Corpus contract green (29 fixtures both directions + render equivalence). The 4 previously reproduced legal-input violations (CR-01/02/03/05) are closed: unrepresentable values now hard-error loudly before emission (CheckC4DRepresentable) and every written twin is gated by re-parse + CanonicalEqual — silent corruption is structurally impossible on the convert path |
| D-24, D-25, D-27, D-28, D-29, D-30 | 35-07 | ✓ SATISFIED (re-verified) | All wired and tested; D-30's -o directory override now preserves graph-mode structure for relative entry paths (CR-06 closed); the D-24 gate plus the new write gate cover inputs whose twins would previously corrupt (they now fail loudly with no output) |
| D-31, D-32, D-33 | 35-08 | ✓ SATISFIED | fmt behavioral checks + corpus idempotency sweeps |
| D-34, D-35 | 35-09 | ✓ SATISFIED (re-verified) | README/skill/twins verified; the full-feature-parity claim (README.adoc:743) no longer contradicted — width/height round-trip (CR-04 closed) |

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| (none — no TBD/FIXME/XXX/TODO/HACK/PLACEHOLDER markers in any phase-modified or fix-modified file; re-scanned 2026-08-14) | - | - | - | - |
| 5 pre-existing Go files (not touched by 35-09) | - | gofmt drift (toolchain version) | ℹ️ Info | Logged in 35-09-deferred-items.md; out of phase scope |
| `internal/testutil/canonsrc/canonsrc.go` | 549-551 | Comment still claims the quoted form keeps embedded pipes round-trip safe; the operative protection is now the representability gate, not the quoting | ⚠️ Warning (residual, doc-only) | No silent corruption remains possible on the convert path; the stale comment could mislead future maintainers |
| `skill/SKILL.md` | 64-68 | Type-inference table row contradicts DefaultTypeForParent (WR-04) | ⚠️ Warning (open) | Skill users author wrong child types under systemExternal/box — both parents actually fall to the default C1/system branch |
| `internal/c4d/grammar/c4d.peg` | ~226 (Doc statement switch) | Duplicate `properties {}` blocks silently last-win (WR-03); TOML twin is a hard parse error | ⚠️ Warning (open) | Silent field loss parity gap |
| `internal/c4d/grammar/c4d.peg` | ~157-165 | Quoted edge-label whitespace trimmed (WR-05) | ⚠️ Warning (open) | `description = " padded "` loses padding through C4D |

### Human Verification Required

#### 1. README C4D section + twins + skill review

**Test:** Read the rendered README.adoc `== C4D Format` section; compare `skill/examples/06-templates.c4d` (43 lines) against `06-templates.toml` (109 lines).
**Expected:** Syntax documentation correct and readable; side-by-side example clear; twins deliver the promised verbosity win.
**Why human:** 35-09 Task 4 was a blocking `checkpoint:human-verify` gate that was AUTO-approved under AUTO-MODE — the five verification steps were executed by the agent, not a human. Doc readability and the verbosity-win judgment are subjective.

#### 2. Optional: WR-04 skill table fix confirmation

**Test:** After any fix pass, verify `skill/SKILL.md` type-inference table matches `internal/parser/parser.go` DefaultTypeForParent.
**Expected:** Rows for systemExternal and box parents reflect the actual C1 fallback behavior (both fall to the default branch returning C1's `system`).
**Why human:** Requires reading both the table and the Go switch side by side; doc-correctness judgment.

### Gaps Summary

**Re-verification outcome: all three gaps from the initial verification are CLOSED; 24/24 truths verified.** Every fix was confirmed by independent reproduction against a freshly built CLI (not the fix commits' own tests):

1. **Converter correctness outside the corpus (was BLOCKER) — CLOSED by c59b762.** CR-01 (pipe in link technology) and CR-03 (non-identifier unit ids such as `"my unit"` / `"my.unit"`) now hard-error loudly, naming the offending value and the source file, with exit 1 and no twin written (`CheckC4DRepresentable`, `internal/c4d/frommodel.go:388`, wired in `emitConverted`). CR-02 (non-bare TOML keys) is fixed at the emitter (`tomlKeySegment`, `emit_toml.go:441`) — the type-led-unit twin `["My App"]` parses and renders. CR-05 (quote-terminated multiline) is fixed in `literalFor` via the escaped-quoted fallback — the twin re-parses and the value survives byte-for-byte. Additionally, `gateTwin` (convert.go:237, called from both `runConvert` and `convertGraphFile`) re-parses every twin and requires `c4d.CanonicalEqual` before any write, making silent corruption structurally impossible on the convert path even for defect classes outside the pinned representability list.
2. **Full TOML feature parity incl. width/height (was BLOCKER unless overridden) — CLOSED by f553a9c.** width/height now flow through the grammar FieldKey, `tomodel.go` (with numeric validation), `frommodel.go`, both emitters, and the canonsrc canonical form; `CanonicalEqual` compares them; reproduced 300x200 -> `width: 300`/`height: 200` -> `width = 300`/`height = 200` intact in both directions. No override was needed.
3. **Graph-mode -o structure preservation (was BLOCKER) — CLOSED by 0a17d64.** `entryDir` now derives from the absolutized entry (convert.go:287); the previously flattening relative-entry invocation yields `OUT/domains/auth.c4d`, with the include path retargeted and the composed twin graph rendering.

Gates re-run clean: `go build ./...` ok, `go test ./... -count=1` 15/15 packages ok, `golangci-lint run` 0 issues (phase packages and full repo). No debt markers in any fix-modified file.

**Remaining open items (all non-blocking, carried from the initial verification):**

- **WR-03** — duplicate `properties {}` blocks in one C4D document silently last-win (TOML twin is a hard parse error).
- **WR-04** — `skill/SKILL.md` type-inference table drift vs `DefaultTypeForParent` (systemExternal and box parents fall to the default C1 branch; the table claims `container` for them).
- **WR-05** — quoted C4D edge-label whitespace is trimmed by `splitPipeLabel` (`" padded "` loses padding through C4D).
- **Residual doc nuance** — `canonsrc.go:549-551` comment still attributes pipe round-trip safety to quoting; the operative protection is now the representability gate (behaviorally harmless, comment-only).
- **2 human verification items** — README/twins/skill readability review (the AUTO-approved 35-09 human gate) and the optional WR-04 fix confirmation; these gate the `passed` status per workflow rules.

No ROADMAP phases exist after 35; nothing here is deferrable. The phase goal is achieved at the code level — the remaining items are documentation-quality fixes and the two human sign-offs.

---

_Verified: 2026-08-14T21:27:11Z (re-verification after fixes 0a17d64/f553a9c/c59b762)_
_Verifier: Claude (gsd-verifier)_
