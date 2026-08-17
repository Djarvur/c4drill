---
phase: 35
slug: add-a-simple-dsl-alternative-to-the-toml-diagram-definition
status: verified
threats_open: 0
asvs_level: 1
created: 2026-08-17
---

# Phase 35 — Security

> Per-phase security contract: threat register, accepted risks, and audit trail.

---

## Trust Boundaries

| Boundary | Description | Data Crossing |
|----------|-------------|---------------|
| .c4d file → parser | untrusted author input crosses into the pigeon-generated parser; malformed/hostile input must fail closed with bounded work | DSL source text |
| authored TOML → Expand recursion | untrusted author input drives recursion depth; unbounded recursion = DoS | template use graph |
| .c4d/.toml include graph → parser dispatch | file paths from include directives drive file reads; extension dispatch must fail closed on unknown types | include paths |
| Model → emitted text | converter/fmt output written to disk; emitters are pure functions (no I/O) | generated twin documents |
| CLI args (-o dir, input path) → file writes | user-controlled output location; output must not clobber unrelated files or escape expectations | output paths, file contents |
| fmt args → in-place file rewrites | formatter output REPLACES author files — corruption or injection here is destructive | author files |
| docs/skill content → AI agents and users | the skill instructs agents running c4drill; examples are executed as-is | instruction text |

---

## Threat Register

| Threat ID | Category | Component | Disposition | Mitigation | Status |
|-----------|----------|-----------|-------------|------------|--------|
| T-35-01-SC | Tampering | go.mod / pigeon install | mitigate | pigeon exact-pinned `github.com/mna/pigeon v1.3.0` (go.mod:7, deviation from planned v1.0.0 documented — v1.0.0 lacks `-nolint`); go.sum h1 checksums (go.sum:24-25); tools.go build-tag blank-import pin; generated parser_gen.go committed so downstream builds never invoke pigeon | closed |
| T-35-01-01 | DoS | internal/c4d/parse.go | mitigate | pigeon Memoize(true) + MaxExpressions(1_000_000) on every parse call (parse.go:82-89) — pathological inputs (catastrophic backtracking) terminate with an error instead of hanging | closed |
| T-35-01-02 | Information disclosure | internal/c4d/errors.go | mitigate | ParseError carries Message/Line/Context only (wrapPigeonError, errors.go:28-54) — never dumps full file contents; Context is the caller-supplied path | closed |
| T-35-01-03 | Tampering | parser_gen.go | accept | Committed generated code matches the pinned pigeon + committed grammar; regeneration (`go generate ./internal/c4d`) diff reviewable in CI | closed |
| T-35-02-01 | DoS | internal/template/expand.go recursion | mitigate | ancestor-stack cycle detection (`slices.Contains(stack, name)`, expand.go:200-218) naming the chain + maxTemplateDepth=100 depth cap (expand.go:40); every level allocates via Clone so memory is bounded by depth × body size | closed |
| T-35-02-02 | Tampering | pathTracker claim ordering | mitigate | claimSubtree (expand.go:383-412) claims every descendant path of each attached clone; seedAuthored pre-seeds hand-authored units; nested uses cannot silently overwrite authored units (TMPL-07 closed) | closed |
| T-35-02-03 | Repudiation | error reporting | accept | ExpandError carries Kind/Site/Detail naming the instantiation index and chain — sufficient for diagnosis; no security logging surface here | closed |
| T-35-03-01 | Tampering | reserved-keyword disambiguation | mitigate | exact-match 19-word table (reserved.go:17-27); near-miss ids remain legal (TestParseNearMissUnitIdLegal); brace-lookahead false-positive guard (TestParseReservedLedFieldStatementsStillParse) | closed |
| T-35-03-02 | DoS | composition grammar combinatorics | mitigate | Memoize(true) + MaxExpressions(1M) apply unchanged (parse.go:88-89); list/arg forms bounded by input size | closed |
| T-35-03-03 | Information disclosure | suggestion output | accept | FormatSuggestion echoes only the author's own token against a fixed public keyword list — no disclosure surface | closed |
| T-35-04-01 | Tampering | emitter determinism | mitigate | Emitters walk UnitOrder/SubunitOrder only — never map iteration (sortedTemplateNames/sortedParamKeys where maps are unavoidable); idempotency fixpoint tests pin emit(emit(x)) == emit(x) | closed |
| T-35-04-02 | Information disclosure | quoted string emission | mitigate | quoteTOML (emit_toml.go:332-369) / quoteC4D (emit_c4d.go:618-641) escape control characters and quotes — no raw injection into the output format; c4dBarewordSafe excludes every grammar stop char; post-fix hardening: CheckC4DRepresentable loud-errors values C4D cannot express (frommodel.go:388, CR-01/CR-03), literalFor escaped-quoted fallback for quote-terminated multiline (CR-05), tomlKeySegment quoted TOML keys (emit_toml.go:441-449, CR-02) | closed |
| T-35-04-03 | Repudiation | canonical order | accept | D-23 fixed order (now including width/height after f553a9c) makes diffs reviewable — a security-relevant field change always shows in the same position | closed |
| T-35-05-01 | Tampering | include extension dispatch | mitigate | Unknown extension = hard `*parser.ParseError` naming accepted extensions (resolve.go:205-222); no fallback parsing, no content-sniffing (TestResolveUnknownExtensionHardError) | closed |
| T-35-05-02 | DoS | mixed-graph cycles | mitigate | maxIncludeDepth=100 (resolve.go:37) + visited-set cycle detection (resolve.go:56, 123) — dispatch change does not touch traversal | closed |
| T-35-05-03 | Elevation | Model construction | accept | ToModel builds in-memory structs only; no I/O, no exec; unknown types/edges fail closed | closed |
| T-35-06-01 | Tampering | corpus walker | mitigate | filepath.WalkDir does not follow symlinked directories; *.toml/*.c4d extension filter only; repo-root anchored (parity_test.go:386) | closed |
| T-35-06-02 | DoS | render-equivalence tests | accept | WASM engine serial execution (repo rule, no t.Parallel) bounds concurrency; corpus size is fixed and small | closed |
| T-35-06-03 | Repudiation | parity failures | mitigate | Round-trip failures are hard per-fixture test failures naming the fixture path (t.Run(rel, ...)) — no silent skips | closed |
| T-35-07-01 | Tampering | convert output writes | mitigate | Output filename = input basename + swapped extension ONLY (deriveBasename, convert.go:260-266 — no user-controlled filename portion); gateTwin re-parse + c4d.CanonicalEqual before ANY write (convert.go:237-256, wired at :150 single-file and :327 graph mode — c59b762); -o accepts a user-named directory (standard CLI trust); overwrite is the documented idempotent contract | closed |
| T-35-07-02 | DoS | --follow-includes traversal | mitigate | maxConvertDepth=100 (convert.go:58) + walkIncludeGraph visited-set/ancestor-stack cycle detection (convert.go:357-371); D-24 gate's include.Resolve cycle rejection runs first (TestConvertGraphCycle — no partial outputs) | closed |
| T-35-07-03 | Tampering | symlinked include targets | mitigate | Traversal keys on canonicalized paths (filepath.Clean + Abs, convert.go:431); a symlink loop collapses into the visited-set cycle error; CR-06 fix: entryDir from the absolutized entry path (convert.go:278-287, 0a17d64) — relative entry paths preserve -o graph structure | closed |
| T-35-07-04 | Information disclosure | error output | accept | Stage-prefixed errors echo author-supplied paths/messages only — same surface as the existing root command | closed |
| T-35-08-01 | Tampering | in-place rewrite | mitigate | Semantic safety gate (applyFormatted, fmt.go:243-255): candidate output must re-parse to a Model reflect.DeepEqual to the original BEFORE any file is replaced; malformed outputs are hard errors and the file is left untouched; both failure legs unit-tested (TestFMTGateBlocksBrokenRewrite, TestFMTRefusesModelBrokenC4D) | closed |
| T-35-08-02 | Tampering | recursive walk | mitigate | filepath.WalkDir (no symlinked-dir following, fmt.go:160), extension filter *.c4d/*.toml only (fmt.go:168) — fmt cannot be pointed at arbitrary file types; direct file args gated on ext (fmt.go:139) | closed |
| T-35-08-03 | DoS | huge/pathological files | accept | Local developer tool over own files; parse costs bounded by MaxExpressions (C4D) and go-toml (TOML) | closed |
| T-35-08-04 | Repudiation | --check CI gate | mitigate | Differing paths print one per line to stdout + exit 1 (fmt.go:94-105); zero byte writes in check mode (TestFMTCheckMisformattedExitsOne) | closed |
| T-35-09-01 | Tampering | skill/plugin instruction files | mitigate | Content documents only the real implemented flags (verified by acceptance greps — zero hits for the invented `c4drill validate` across skill/README/plugins); no invented commands an agent could execute blindly | closed |
| T-35-09-02 | Repudiation | example twins | mitigate | TestExampleTwins enforces model + render parity for all 12/12 twins against a pinned manifest with set-equality anti-shrinkage — a broken example fails CI, not just review | closed |
| T-35-09-03 | Information disclosure | README | accept | Public docs for a public CLI; no secrets involved | closed |

*Status: open · closed*
*Disposition: mitigate (implementation required) · accept (documented risk) · transfer (third-party)*

---

## Accepted Risks Log

| Risk ID | Threat Ref | Rationale | Accepted By | Date |
|---------|------------|-----------|-------------|------|
| AR-01 | T-35-01-03 | Committed generated code matches the pinned pigeon + committed grammar; regeneration diff is reviewable in CI | user | 2026-08-14 |
| AR-02 | T-35-02-03 | ExpandError Kind/Site/Detail is sufficient for diagnosis; no security logging surface in a CLI tool | user | 2026-08-14 |
| AR-03 | T-35-03-03 | FormatSuggestion echoes only the author's own token and fixed public keywords — no disclosure surface | user | 2026-08-14 |
| AR-04 | T-35-04-03 | D-23 fixed order makes diffs reviewable; canonical order is a repudiation aid, not a security control | user | 2026-08-14 |
| AR-05 | T-35-05-03 | ToModel builds in-memory structs only; no I/O, no exec; unknown types/edges fail closed | user | 2026-08-14 |
| AR-06 | T-35-06-02 | WASM engine serial execution (repo rule) bounds concurrency; corpus is fixed and small | user | 2026-08-14 |
| AR-07 | T-35-07-04 | Stage-prefixed errors echo author-supplied paths/messages only — same surface as the existing root command | user | 2026-08-14 |
| AR-08 | T-35-08-03 | Local developer tool over own files; parse costs bounded by MaxExpressions (C4D) and go-toml (TOML) | user | 2026-08-14 |
| AR-09 | T-35-09-03 | Public docs for a public CLI; no secrets involved | user | 2026-08-14 |

*Accepted risks do not resurface in future audit runs.*

---

## Security Audit Trail

| Audit Date | Threats Total | Closed | Open | Run By |
|------------|---------------|--------|------|--------|
| 2026-08-17 | 30 | 30 | 0 | gsd-security-auditor (post-gap-fix verification, commits 0a17d64/f553a9c/c59b762) |

---

## Sign-Off

- [x] All threats have a disposition (mitigate / accept / transfer)
- [x] Accepted risks documented in Accepted Risks Log
- [x] `threats_open: 0` confirmed
- [x] `status: verified` set in frontmatter

**Approval:** verified 2026-08-17
