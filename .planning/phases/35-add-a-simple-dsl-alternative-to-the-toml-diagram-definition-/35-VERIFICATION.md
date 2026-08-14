---
phase: 35-add-a-simple-dsl-alternative-to-the-toml-diagram-definition-
verified: 2026-08-14T00:00:00Z
status: gaps_found
score: 21/24
overrides_applied: 0
gaps:
  - truth: "Converters produce canonical-equivalent round-trips for every legal input (goal: 'bidirectional canonical-equivalent converters'; D-22 contract text)"
    status: failed
    reason: >-
      Five independently reproduced defect classes in which a LEGAL source document
      converts to a twin that is silently corrupted or does not parse, all with exit 0
      and no warning: (CR-01) link technology containing '|' is reshuffled between
      Technology and Description on round-trip; (CR-02) a type-led .c4d unit with a
      display name converts to an unquoted TOML table key ('[My App]') that does not
      parse; (CR-03) TOML unit ids outside [A-Za-z0-9_-]+ (spaces, dots, unicode)
      produce unparseable C4D twins; (CR-05) a multi-line value ending in '"' emits
      an ambiguous triple-quote closer that does not parse. The 35-04-deferred-items
      note claiming "canonical C4D quoting keeps embedded pipes round-trip safe" is
      factually wrong — the D-09 first-pipe split runs after unquoting. The 29-fixture
      parity corpus is ASCII-identifier-safe and pipe-free, so the shipped suite
      never exercises these paths; canonsrc shares the same split logic, so the
      oracle is blind to CR-01.
    artifacts:
      - path: internal/c4d/emit_c4d.go
        issue: "emitEdgeC4D concatenates tech|desc with no representability check; literalFor picks triple quotes for quote-terminated multi-line values (CR-01, CR-05)"
      - path: internal/c4d/emit_toml.go
        issue: "table headers written with raw unquoted key segments (CR-02)"
      - path: internal/c4d/frommodel.go
        issue: "unit ids / peers emitted verbatim with no grammar-charset gate (CR-03)"
      - path: cmd/c4drill/convert.go
        issue: "no fmt-style re-parse/equality gate before writing twins to disk"
    missing:
      - "Loud hard errors (or representable forms) for values the target format cannot express — pipe-containing labels, non-bare TOML keys, non-C4D identifiers, quote-terminated multi-line strings"
      - "A convert write gate mirroring fmt's T-35-08-01 safety gate: re-parse the emitted twin and require model equality before writing"
  - truth: "Full TOML feature parity (phase goal wording; README.adoc:743 claims 'Everything the TOML format expresses ... has a C4D equivalent')"
    status: failed
    reason: >-
      model.Unit carries Width/Height (README 'Styling' documents width = 300), but
      the C4D grammar's FieldKey set rejects width:/height: and neither emitter
      serializes them: convert to-c4d drops them silently (verified 300x200 -> 0x0)
      and the back-converted TOML lacks the keys. canonsrc.go:170-172 deliberately
      excludes them from the canonical form, hiding the loss from the parity oracle.
      The 35-04 PLAN's pinned D-23 field order omitted width/height (a planning-scope
      reduction the executor followed), but the phase GOAL and the README claim full
      parity. Override candidate — see report.
    artifacts:
      - path: internal/c4d/frommodel.go
        issue: "appendUnitBody never emits Width/Height"
      - path: internal/c4d/grammar/c4d.peg
        issue: "FieldKey set has no width/height — twin cannot express them"
    missing:
      - "width/height in the C4D FieldKey set + unitStringField/appendUnitBody/EmitTOML so the fields round-trip, OR an accepted override documenting the deliberate exclusion"
  - truth: "convert --follow-includes -o preserves the graph's relative directory structure (README.adoc:1161 promise)"
    status: failed
    reason: >-
      convertGraph computes entryDir from the raw CLI arg (convert.go:233); for a
      relative entry path, filepath.Rel('.', abs) errors, the err == nil guard fails,
      and every twin lands flat in -o (verified: out/auth.c4d instead of
      out/domains/auth.c4d; the absolute-path invocation produces the documented
      layout). Tests only pass absolute t.TempDir() paths, so the suite cannot catch it.
    artifacts:
      - path: cmd/c4drill/convert.go
        issue: "entryDir := filepath.Dir(entryPath) uses the raw arg; must use filepath.Dir(absEntry)"
    missing:
      - "One-line fix: entryDir from the absolutized entry path; add a relative-entry-path test case"
human_verification:
  - test: "Human review of the README.adoc C4D Format section, the 12 example twins, and the dual-format skill"
    expected: "Syntax documentation correct and readable; side-by-side example clear; twins deliver the promised verbosity win (06-templates: 109 -> 43 lines claimed)"
    why_human: "35-09 Task 4 was a blocking checkpoint:human-verify gate that was AUTO-approved under AUTO-MODE — the five verification steps were executed by the agent, not reviewed by a human; doc readability and the verbosity-win judgment are subjective"
---

# Phase 35: C4D DSL Alternative to TOML — Verification Report

**Phase Goal:** Deliver the C4D format — a `.c4d` brace-block D2-inspired DSL with full TOML feature parity — parseable directly to `*parser.Model` and renderable through the unchanged pipeline, with bidirectional canonical-equivalent converters (`convert to-toml`/`to-c4d`), a gofmt-style comment-preserving formatter (`fmt`) for both formats, nested use and recursive template-instantiating-template expansion, plus full README/skill/example documentation.
**Verified:** 2026-08-14
**Status:** gaps_found
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

Roadmap success_criteria is empty for this phase; truths merged from the phase goal wording and the 9 PLAN frontmatters.

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Core C4D grammar parses brace-block units, nesting, fields, in-block edges, properties into typed comment/position-aware AST (D-01, D-08, D-12) | ✓ VERIFIED | `internal/c4d/grammar/c4d.peg` (787 lines, rules Document/UnitHeader/EdgeStmt/Literal/Comment/Properties), `internal/c4d/ast/ast.go` (211 lines, Pos+Comments on nodes); parse tests green |
| 2 | Unit headers `id: type "Name"` with omittable slots; exact TOML type keywords; external modifier -> *External variants (D-02, D-03, D-04) | ✓ VERIFIED | c4d.peg UnitHeader (3 alternatives); TestParseUnitExternalModifier, TestToModelExternalModifier pass |
| 3 | Arrows `->` `<-` `<->` `--` + desc/tech pipe shorthand with desc-first single value (D-05, D-09) | ✓ VERIFIED | Grammar EdgeStmt; composition/parse tests green |
| 4 | Bareword/double/triple-quoted literals, `#` comments, scheme URLs valid barewords; identifiers `[A-Za-z0-9_-]+`, dotted paths (D-06, D-07) | ✓ VERIFIED | Grammar Literal rule; tests green. Caveat: quote-terminated multiline VALUES break on EMIT (gap 1) |
| 5 | `;` separator + empty blocks (D-18); reserved keyword collision = hard error with Levenshtein suggestion (D-19) | ✓ VERIFIED | `internal/c4d/grammar/reserved.go` (77 lines): 14 isBuiltinField + 5 statement keywords, FormatSuggestion wired; TestParseReservedUnitIdError, TestParseReservedKeywordsTable/List pass |
| 6 | pigeon pinned in go.mod, generated parser committed, go:generate chain (D-20) | ✓ VERIFIED | go.mod `github.com/mna/pigeon v1.3.0` (deviation from planned v1.0.0 — documented: v1.0.0 lacks the -nolint flag the lint gate requires); tools.go build-tag pin; `//go:generate pigeon -o parser_gen.go -nolint c4d.peg`; parser_gen.go committed |
| 7 | Parse failures are `*parser.ParseError` with DSL-native line numbers (D-21) | ✓ VERIFIED | `internal/c4d/errors.go` wraps pigeon errList; tests assert Line == offending line |
| 8 | TOML `[[unit.X.use]]` + `[[template.X.use]]` desugar to Instantiation (D-16, D-17 TOML forms) | ✓ VERIFIED | parser.go extractUnitUses; TestParseUnitUseSugarEquivalentToParentField, TestParseTemplateBodyUseExtracted pass; TemplateDef.Instantiations field present |
| 9 | Recursive template-body expansion: outer-to-inner params, cycle chain "A -> B -> A", depth cap 100, HS-1 deep-copy every level (D-17) | ✓ VERIFIED | expand.go checkRecursion + maxTemplateDepth=100; TestExpandTemplateCycle, TestExpandTemplateSelfCycle, TestExpandTemplateDepthCap, TestExpandThreeLevelNestingHS1 pass |
| 10 | v1.10 STATE.md template-nesting deferral lifted | ✓ VERIFIED | STATE.md Deferred Items: "SHIPPED in Phase 35 (D-17, Plan 35-02) — deferral lifted" |
| 11 | template/use(3 positions)/include-once grammar + both list forms (D-13, D-14, D-15) | ✓ VERIFIED | composition_test.go; ToModel maps UseStmt parents; tests green |
| 12 | EmitTOML deterministic canonical field order, fixture-shaped tables; EmitC4D compact-leaf style; FromModel inverse (D-23, D-33) | ✓ VERIFIED | TestEmitTOMLCanonicalUnitFieldOrder, TestEmitTOMLUnitOrderFollowsUnitOrder, fixpoint tests; emit_toml.go 396 / emit_c4d.go 676 / frommodel.go 357 lines |
| 13 | c4d.Parse -> *parser.Model directly; ParseAST/ParseASTFile exported; inference parity identical Models (D-21, D-02) | ✓ VERIFIED | parse.go:20/32/63/71 signatures confirmed; parser.DefaultTypeForParent/InferGenericType exported (parser.go:954/992); TestToModelInferenceParity passes |
| 14 | Duplicate edge hard error (D-11); peers through unchanged peer.Resolve (D-10); mixed .c4d/.toml include graphs (D-26) | ✓ VERIFIED | TestToModelDuplicateEdgeError; include/resolve.go checkIncludeExtension dispatch; TestResolveTomlIncludesC4d, TestResolveC4dIncludesToml, TestResolveUnknownExtensionHardError, TestResolveMixedFormatCycleFatal pass |
| 15 | D-22 parity over the fixture corpus: both-direction canonical-equivalent round-trips + render equivalence + composed graph | ✓ VERIFIED | parity_test.go: TestRoundTripTOMLToC4DToTOML (29 fixtures, require.Greater > 15 anti-shrinkage), TestRoundTripC4DToTOMLToC4D, TestRenderEquivalence, TestComposedGraphRoundTrip — all green |
| 16 | `c4drill diagram.c4d` renders directly; extension dispatch hard-errors naming .toml/.c4d (D-27, D-29) | ✓ VERIFIED | root.go:218 c4d.ParseFile branch; behavioral: `unsupported input extension ".json" (accepted: .toml, .c4d)` |
| 17 | convert validates first on a discarded pipeline copy, emits from a fresh parse preserving includes/templates/bare peers; output placement swapped-ext + -o (D-24, D-25, D-28, D-30 single-file) | ✓ VERIFIED | convert.go:113-161 runConvert order (direction gate -> validateSourceForConvert -> fresh parse -> emit -> write); convert tests green |
| 18 | convert --follow-includes converts whole graph with rewritten paths, once preserved, cycle-safe (D-25 graph mode) | ✓ VERIFIED | convertGraph + walkIncludeGraph (visited set, maxConvertDepth, ancestor cycle check); retargetExt preserves relative form |
| 19 | fmt rewrites both formats in place; --check exit 1 no write; comments + blank-line grouping preserved; idempotent; semantic safety gate (D-31, D-32, D-33) | ✓ VERIFIED | Behavioral: `fmt --check` exit 1 with zero byte change; in-place rewrite produced canonical `a: system "A" { description: x }`; lead + trailing comments preserved; tomlfmt TestFormatPreservesComments, TestFormatIdempotentOverCorpus, corpus sweeps green |
| 20 | Format named C4D; README.adoc C4D section + convert/fmt CLI reference; 12 render-identical twins; skill extended in place with name kept, plugin copies synced (D-34, D-35) | ✓ VERIFIED | `== C4D Format` section present, 70 c4d/C4D mentions, follow-includes + --check documented; skill/SKILL.md `name: c4drill-toml` kept, byte-identical to packaged copy (diff clean); 12 twins in skill/examples + 12 in plugin examples; TestExampleTwins 12/12 pass |
| 21 | Full repo health: build + all tests + lint | ✓ VERIFIED | `go build ./...` ok; `go test ./... -count=1` 15/15 packages ok; golangci-lint 0 issues; no TBD/FIXME/XXX/TODO/HACK markers in any phase file |
| 22 | Converters produce canonical-equivalent round-trips for every legal input (goal wording, unqualified; D-22 contract text) | ✗ FAILED | Independently reproduced (scratch programs against exported APIs + built CLI): pipe-in-tech silently reshuffles tech/desc; type-led unit with display name -> unparseable `[My App]` TOML twin; space-containing TOML id -> unparseable C4D twin; quote-terminated multiline -> unparseable triple-quote closer. All exit 0 |
| 23 | Full TOML feature parity (goal wording) | ✗ FAILED | width/height silently dropped by convert to-c4d (300x200 -> 0x0 reproduced); C4D grammar cannot express them; README.adoc:743 claims full equivalence. Deliberate exclusion documented in canonsrc.go:170-172 and plan-pinned D-23 order — override candidate |
| 24 | convert --follow-includes -o preserves the graph's relative directory structure (README.adoc:1161) | ✗ FAILED | Reproduced: relative entry path yields flat `out/auth.c4d`; absolute path yields documented `out/domains/auth.c4d`. convert.go:233 uses raw arg |

**Score:** 21/24 truths verified

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `internal/c4d/grammar/c4d.peg` | PEG grammar (plan path `internal/c4d/c4d.peg` — layout deviation, documented) | ✓ VERIFIED | 787 lines, all core + composition rules |
| `internal/c4d/grammar/parser_gen.go` | committed generated parser | ✓ VERIFIED | Present, nolint header, builds |
| `internal/c4d/ast/ast.go` | typed AST with Pos/Comments | ✓ VERIFIED | 211 lines |
| `internal/c4d/errors.go` | ParseError wrapping | ✓ VERIFIED | Wired to parser.ParseError via errors.As |
| `internal/parser/parser.go` | Instantiations + exported inference helpers | ✓ VERIFIED | TemplateDef.Instantiations, DefaultTypeForParent, InferGenericType |
| `internal/template/expand.go` | recursive expansion + cycle detection | ✓ VERIFIED | checkRecursion, maxTemplateDepth, cycle chain message |
| `internal/c4d/tomodel.go` | ToModel with parity hooks | ✓ VERIFIED | 861 lines; calls DefaultTypeForParent/InferGenericType/Humanize |
| `internal/c4d/emit_toml.go` | EmitTOML canonical order | ✓ VERIFIED | 396 lines; canonical-order test pins D-23 |
| `internal/c4d/emit_c4d.go` + `frommodel.go` | compact-leaf printer + Model->AST | ✓ VERIFIED | 676 + 357 lines; defect classes in Gaps |
| `internal/testutil/canonsrc/canonsrc.go` | NormalizeTOML + NormalizeC4D | ✓ VERIFIED | 733 lines; fixpoint + quoting/key-order/newline tests |
| `internal/c4d/parity_test.go` | corpus round-trip + render equivalence | ✓ VERIFIED | 13 test funcs; 29-fixture corpus; 12 twins |
| `testdata/c4d/*.toml` | 6 edge-case fixtures | ✓ VERIFIED | All 6 present, parse + validate green |
| `cmd/c4drill/convert.go` | convert subcommand | ✓ VERIFIED | 404 lines; validate-first + fresh-parse + graph mode |
| `cmd/c4drill/fmt.go` + `internal/tomlfmt/tomlfmt.go` | formatter + TOML formatter | ✓ VERIFIED | 272 + 285 lines; behavioral checks pass |
| `README.adoc`, `skill/SKILL.md`, 12 + 12 twins | docs surface | ✓ VERIFIED | Greps + diff + TestExampleTwins all pass |

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|----|--------|---------|
| cmd/c4drill/root.go runRoot | c4d.ParseFile | extension branch | ✓ WIRED | root.go:218 |
| cmd/c4drill/root.go NewRootCmd | newConvertCmd/newFMTCmd | cmd.AddCommand | ✓ WIRED | root.go:100,103 |
| cmd/c4drill/convert.go | validator.Validate + emitters | D-24 gate then fresh-parse emit | ✓ WIRED | validateSourceForConvert + emitConverted |
| cmd/c4drill/fmt.go | c4d.ParseAST + EmitC4D + tomlfmt.Format | per-extension path | ✓ WIRED | Confirmed in fmt.go + behavior |
| internal/c4d/tomodel.go | parser inference hooks + Model | parity hooks | ✓ WIRED | Signatures + tests |
| internal/include/resolve.go | extension-branched parse | .c4d dispatch | ✓ WIRED | checkIncludeExtension |
| internal/c4d/grammar/reserved.go | validator.FormatSuggestion | suggestion on collision | ✓ WIRED | reserved.go:62 |
| skill/examples twins | render parity | TestExampleTwins walker | ✓ WIRED | 12/12 pass |
| README.adoc CLI Reference | convert/fmt flags | usage examples | ✓ WIRED | follow-includes, --check documented |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Full test suite | `go test ./... -count=1` | 15/15 packages ok | ✓ PASS |
| Corpus round-trip tally | `go test ./internal/c4d/ -run TestRoundTripTOMLToC4DToTOML -v` | 29 fixtures, all pass | ✓ PASS |
| Twins render parity | `go test ./internal/c4d/ -run TestExampleTwins -v` | 12/12 pass | ✓ PASS |
| Unknown extension hard error | `c4drill foo.json` | names .toml and .c4d | ✓ PASS |
| fmt --check CI gate | `c4drill fmt --check bad.c4d` | exit 1, zero byte change | ✓ PASS |
| fmt in-place + comment preservation | `c4drill fmt` on misformatted .c4d | canonical output, comments kept | ✓ PASS |
| CR-01 pipe corruption | scratch round-trip via c4d.Parse/EmitC4D | tech "HTTP \| REST" -> "HTTP", desc "calls" -> "REST \| calls" | ✗ FAIL |
| CR-02 type-led unit to TOML | scratch c4d.ParseAST->ToModel->EmitTOML->parser.Parse | twin `[My App]` -> parse error | ✗ FAIL |
| CR-03 space id to C4D | scratch parser.Parse->EmitC4D->c4d.Parse | twin `my unit: ...` -> parse error | ✗ FAIL |
| CR-04 width/height | scratch round-trip | 300x200 -> 0x0 | ✗ FAIL |
| CR-05 quote-terminated multiline | scratch round-trip | twin `line2""""` -> parse error line 3 | ✗ FAIL |
| CR-06 -o flattening (relative entry) | built CLI, temp include graph | out/auth.c4d instead of out/domains/auth.c4d | ✗ FAIL |
| Lint gate | `golangci-lint run` (phase packages) | 0 issues | ✓ PASS |

### Requirements Coverage

All 35 D-IDs from 35-CONTEXT.md are claimed across the 9 PLAN frontmatters (union = D-01..D-35, no orphans, no gaps in coverage on paper). Codebase status:

| Requirement | Source Plan | Status | Evidence |
|-------------|------------|--------|----------|
| D-01..D-09, D-12, D-18, D-20 | 35-01 | ✓ SATISFIED | Grammar + tests + pigeon chain (v1.3.0 pin documented) |
| D-16, D-17 | 35-02 | ✓ SATISFIED | extractUnitUses, recursive expand, cycle/depth/HS-1 tests, STATE.md updated |
| D-13, D-14, D-15, D-19 | 35-03 | ✓ SATISFIED | Composition grammar + reserved.go (19 keywords) |
| D-23, D-33 | 35-04 | ✓ SATISFIED | Canonical-order + compact-leaf tests. NOTE: pinned field order excludes width/height (see D-35 parity nuance below) |
| D-02, D-10, D-11, D-21, D-26 | 35-05 | ✓ SATISFIED | Signatures, parity test, duplicate-edge error, mixed includes |
| D-22, D-26 | 35-06 | ? PARTIAL | Corpus contract green (29 fixtures both directions + render equivalence). Contract text "must produce canonically-equal text" violated by 4 reproduced legal-input classes outside the corpus (CR-01/02/03/05) |
| D-24, D-25, D-27, D-28, D-29, D-30 | 35-07 | ? PARTIAL | All wired and tested; D-30's -o directory override flattens graph-mode structure for relative entry paths (CR-06); D-24 gate passes inputs whose twins are then silently corrupted (gaps 1-2) |
| D-31, D-32, D-33 | 35-08 | ✓ SATISFIED | fmt behavioral checks + corpus idempotency sweeps |
| D-34, D-35 | 35-09 | ✓ SATISFIED | README/skill/twins verified; PARITY CAVEAT: full-feature-parity claim contradicted by width/height drop (CR-04) |

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| (none — no TBD/FIXME/XXX/TODO/HACK/PLACEHOLDER markers in any phase-modified file) | - | - | - | - |
| 5 pre-existing Go files (not touched by 35-09) | - | gofmt drift (toolchain version) | ℹ️ Info | Logged in 35-09-deferred-items.md; out of phase scope |
| `internal/testutil/canonsrc/canonsrc.go` | 540-554 | False "pipe-safe quoting" comment; oracle shares CR-01 split logic | ⚠️ Warning | Parity suite structurally blind to pipe corruption |
| `internal/testutil/canonsrc/canonsrc.go` | 170-172 | width/height excluded from canonical form | ⚠️ Warning | Parity oracle blind to silent field loss |
| `skill/SKILL.md` | 64-68 | Type-inference table row contradicts DefaultTypeForParent (WR-04) | ⚠️ Warning | Skill users author wrong child types under systemExternal/box |
| `internal/c4d/grammar/c4d.peg` | 223-234 | Duplicate `properties {}` blocks silently last-win (WR-03); TOML twin is a hard parse error | ⚠️ Warning | Silent field loss parity gap |
| `internal/c4d/grammar/c4d.peg` | 159-165 | Quoted edge-label whitespace trimmed (WR-05) | ⚠️ Warning | `description = " padded "` loses padding through C4D |

### Probe Execution

Step 7c: SKIPPED — no `scripts/*/tests/probe-*.sh` probes exist in this repo; the phase's runnable verification surface is the Go test suite (executed above).

### Human Verification Required

#### 1. README C4D section + twins + skill review

**Test:** Read the rendered README.adoc `== C4D Format` section; compare `skill/examples/06-templates.c4d` (43 lines) against `06-templates.toml` (109 lines).
**Expected:** Syntax documentation correct and readable; side-by-side example clear; twins deliver the promised verbosity win.
**Why human:** 35-09 Task 4 was a blocking `checkpoint:human-verify` gate that was AUTO-approved under AUTO-MODE — the five verification steps were executed by the agent, not a human. Doc readability and the verbosity-win judgment are subjective.

#### 2. Optional: WR-04 skill table fix confirmation

**Test:** After any fix pass, verify `skill/SKILL.md` type-inference table matches `internal/parser/parser.go` DefaultTypeForParent.
**Expected:** Rows for systemExternal and box parents reflect the actual C1 fallback behavior.
**Why human:** Requires reading both the table and the Go switch side by side; doc-correctness judgment.

### Gaps Summary

All 9 plans' own must-haves pass: the grammar, AST, Model-hub front-end, emitters, converters, formatter, parity suite (29 fixtures, both directions, render-equivalent), CLI wiring, and documentation are present, substantive, wired, and green — `go test ./...` 15/15, lint 0 issues, no debt markers.

The phase goal is nevertheless not fully achieved, in exactly the dimension the code review flagged and this verification independently reproduced end-to-end (scratch programs against the exported APIs plus the built CLI — none of the six findings are speculative):

1. **Converter correctness outside the corpus (BLOCKER).** Four reproduced classes of legal input produce silently corrupted or unparseable twins with exit 0 (CR-01/02/03/05). D-22's contract text is unqualified ("must produce canonically-equal text"); the fixture corpus happens to avoid all four classes, and canonsrc shares the pipe-split logic so the oracle cannot detect CR-01. This contradicts the goal's "bidirectional canonical-equivalent converters" and the project's hard-error-everywhere stance (D-24). The 35-04-deferred-items claim that quoting keeps pipes round-trip safe is factually wrong.
2. **Full TOML feature parity (BLOCKER unless overridden).** width/height are TOML-expressed, README-documented fields that convert drops silently (CR-04). The exclusion was plan-pinned (D-23 order) and documented in canonsrc — deliberate, but it contradicts the goal wording and README.adoc:743. **This looks intentional at the planning level.** To accept the deviation, add to this file's frontmatter:

```yaml
overrides:
  - must_have: "Full TOML feature parity (width/height fields)"
    reason: "D-23 canonical field set deliberately excludes width/height; C4D grammar cannot express them; documented as authoring constraint"
    accepted_by: "{your name}"
    accepted_at: "{ISO timestamp}"
```

   Otherwise the fix is to add width/height to the C4D FieldKey set and both emitters.
3. **Graph-mode -o structure preservation (BLOCKER).** README.adoc:1161 documents structure preservation; the relative-entry-path invocation (the common case) flattens it (CR-06, one-line fix at convert.go:233).

The parity suite itself is honest over its corpus — the failure is that the corpus does not cover the corrupting input classes (ASCII-identifier-safe, pipe-free, no quote-terminated multiline, absolute test paths only). Recommended single highest-leverage fix (per the review, endorsed here): a fmt-style re-parse + model-equality gate in `runConvert`/`convertGraph` before any twin write, plus loud errors for values C4D cannot represent.

No ROADMAP phases exist after 35, so none of these gaps can be deferred to later milestone work.

---

_Verified: 2026-08-14_
_Verifier: Claude (gsd-verifier)_
