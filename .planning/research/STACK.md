# Technology Stack — v1.11 LikeC4 Compatibility Layer

**Project:** C4Drill
**Milestone:** v1.11 LikeC4 Compatibility Layer (native-Go `.c4` parser → `*parser.Model`)
**Researched:** 2026-08-08
**Confidence:** HIGH

## Executive Summary

> **⚠ DECISION OVERRIDE (2026-08-09):** The user has chosen **Pigeon (PEG)** over the hand-written approach recommended below. PEG is now the LOCKED parser technology. The hand-written analysis is preserved verbatim for rationale/history; see the **"Decision Override: PEG Parser"** addendum at the bottom of this file for the binding decision. Every downstream artifact (ARCHITECTURE-v1.11.md, SUMMARY.md, requirements, roadmap) must reflect PEG, not hand-written.

**Hand-written recursive-descent parser with a hand-rolled byte lexer.** *(Original recommendation — OVERRIDDEN; see addendum.)* Add **one** test-only dependency (`google/go-cmp` for AST-golden diffs) and **zero** runtime dependencies. No parser generator, no `text/scanner`, no codegen step.

Three forces lock this decision:
1. **No reusable Go parser exists.** The official LikeC4 is TypeScript/Langium; a "structurizr-parser" hit is also TypeScript/Chevrotain. No Go port of LikeC4 or any C4-family DSL was found. We are building from scratch regardless.
2. **Error recovery is the v1.11 headline.** The milestone contract is "breadth over fidelity, never fatal on unsupported constructs" — drop `deployments {}`, `views {}`, tags, icons, metadata with deduplicated warnings. Hand-written panic-mode synchronization (skip to `}` / next statement) gives precise control over this; every generator alternative offers weaker or more awkward recovery.
3. **Grammar shape fights declarative generators.** Optional `:` after keys, flexible property ordering, `this`/`it` scoping, sourceless `-> target`, two declaration syntaxes (`kind name {}` vs `name = kind {}`), and lexical-scope FQN resolution are all natural in recursive descent but awkward in PEG/struct-tag/LL\* grammars.

The hand-rolled lexer is preferred over `text/scanner` because LikeC4's lexicon — single-quoted `'...'` strings, triple-quoted `'''...'''`, `#tags`, multi-char arrows (`->`, `<->`, `-[kind]->`, `.kind.`) — is not Go-like; `text/scanner`'s token model would be fought more than used.

---

## Recommended Stack

### Core Technologies

| Technology | Version | Purpose | Why Recommended |
|---|---|---|---|
| Hand-written recursive-descent parser | — (stdlib) | Parse LikeC4 DSL subset into a converter-internal AST | Best error recovery; matches existing `internal/parser` hand-rolled style; full control over brace/scope handling and `this`/`it` resolution |
| Hand-rolled byte lexer | — (stdlib) | Tokenize `.c4` source into `(kind, value, line, col)` tokens | LikeC4's lexicon (single/triple quotes, `#tag`, arrow ops) is non-Go; a ~150-line lexer is cleaner than bending `text/scanner` |
| `slices` / `maps` stdlib | Go 1.26.1 | Dedup, ordering, contains checks (already a project idiom) | Already used at `parser.go:733` (`slices.Contains`); no new dep |

### Supporting Libraries

| Library | Version | Purpose | When to Use |
|---|---|---|---|
| `github.com/google/go-cmp` | **v0.7.0** (2025-01-14) | Structural AST-equality diffs in golden tests | New **test-only** dep. The converter emits a nested AST; `cmp.Diff` with `cmpopts.EquateEmpty` gives readable failure output that `testify/assert.Equal` cannot match for deep structs |
| `github.com/stretchr/testify` | v1.11.1 (current) | Assertions, fixtures, `require` — keep as-is | Existing precedent (307 uses in `parser_test.go`); do not displace, only supplement for AST goldens |

### Development Tools

| Tool | Purpose | Notes |
|---|---|---|
| `testdata/*.c4` fixtures | Golden input sources | Mirror the existing `testdata/*.toml` pattern; pair each with a `.golden` AST or expected-`*parser.Model` fixture |
| `internal/testutil/canonical` | DOT-output normalization (existing) | Reuse for end-to-end LikeC4 → TOML-pipeline → DOT regression tests (COMPAT-style, already v1.8/1.10 idiom) |
| `go:generate` | **NOT used** | Deliberate: no codegen step keeps `go build` self-contained |

---

## Architecture Sketch

```
.c4 source
   │
   ▼
[lexer]  ── tokens (Ident/String/Arrow/LBrace/RBrace/Eq/Colon/Comma/Hash/Dot/…)
   │         line+col on every token; skips // and /* */ comments
   ▼
[parser]  ── recursive descent over tokens
   │         top-level: specification | model | views | global | deployment
   │         model:     elementDecl | relationship | extend
   │         elementDecl: (kind name | name '=' kind) [strings] '{' body '}'
   │         body:      nested elementDecl | relationship | property | metadata | style
   │         recovery:  on unexpected token, advance to '}' or stmt boundary, emit warning
   ▼
[converter AST]  ── converter-internal types (Element{Kind,Name,Props,Children}, Rel{Src,Dst,…})
   │
   ▼
[converter] ── flatten braces → dotted paths; resolve this/it/sourceless->; fuzzy kind map;
   │            drop deployments/views/tags/icons/metadata (dedup warnings); map <-> → 2 links
   ▼
*parser.Model  ── feeds existing pipeline (include → template → peer → validate → render)
```

The parser must produce `*parser.ParseError` (`internal/parser/errors.go:13`) for hard errors — same `Message`/`Line`/`Context`/`Cause` shape so the CLI error path is unchanged. The lexer tracks line/col per token to populate `ParseError.Line`.

### Error-recovery pattern

```go
// Inside parseElementBody: unknown construct encountered.
p.warnDedup(constructKind)           // e.g. "metadata", one stderr line per kind
p.skipToSync(syncSet)                 // advance past tokens until '}' or ';' or newline-at-brace-zero
continue                              // resume parsing siblings
```
`warnDedup` is a `map[string]bool` on the parser struct — directly implements PROJECT.md's "deduplicated stderr warnings (one per construct type)".

---

## Alternatives Considered

| Recommended | Alternative | When to Use Alternative |
|---|---|---|
| Hand-written RD | **Pigeon (PEG)** v1.3.0 | If the grammar stabilizes and you want a single-source-of-truth `.peg` grammar with failure-label recovery. Pigeon *does* support recovery. Cost: codegen step, PEG error messages need tuning, packrat memory. Justifiable if a second DSL is ever added |
| Hand-written RD | **ANTLR4 go-runtime** v4.13.1 | If you want best-in-class builtin error recovery and are OK with a Java codegen toolchain (Java only at build time; runtime is pure Go, no CGO). Overkill for one bounded DSL subset |
| Hand-written RD | **participle v2** v2.1.4 | If the grammar were simple and regular. Rejected here: struct-tag model fights optional-`:`, flexible ordering, `this`/`it`, sourceless `->`; backtracking gives weak recovery; infix needs Pratt workaround |
| Hand-written RD | **goyacc** v0.48.0 | Never for new work. LALR(1) conflict pain, 0 importers, legacy Plan-9 lineage |
| Hand-rolled lexer | **`text/scanner`** (stdlib) | Viable but not recommended. It natively handles `//`/`/* */` and strings, but its token kinds are Go-oriented; single-quoted strings, `#tags`, `-[async]->` all need post-processing. A purpose lexer is ~150 lines and zero friction |
| go-cmp (test-only) | testify `assert.Equal` only | If the AST is kept shallow (≤2 levels) and failures stay readable. Unlikely given LikeC4 nesting; go-cmp earns its keep on deep structs |

---

## What NOT to Use

| Avoid | Why | Use Instead |
|---|---|---|
| TypeScript/Langium embedding, WASM, tree-sitter, any CGO | **LOCKED OUT** by user decision (native Go only) | Hand-written Go |
| `goyacc` | Legacy, 0 importers on pkg.go.dev, LALR conflict resolution is the wrong trade for a clean-slate grammar | Hand-written RD |
| `participle` for this grammar | No left recursion, backtracking-only (no real recovery), struct-tag model is a poor fit for LikeC4's flexible block bodies | Hand-written RD |
| `text/scanner` as the lexer | Token model oriented to Go syntax; single-quote strings, `#tag`, arrow operators all need custom handling on top | Hand-rolled lexer |
| Reusing the TOML parser for `.c4` | Different grammar family (braced/infix vs key=value tables); the unstable-API order-capture trick in `parser.go:430` does not transfer | New lexer + RD parser |
| A grammar in a `.g4`/`.peg` file checked in | Adds a codegen build step and a second source-of-truth; the v1.11 scope is a *subset*, so the "grammar" is intentionally fluid during development | Hand-written, single `.go` source |

---

## Stack Patterns by Variant

**If a construct is unsupported (`deployments {}`, `views {}`, custom kind, tags, icons, metadata):**
- Parse it structurally (recognize the block boundary), drop it semantically, emit one deduped warning via `warnDedup`. Never fatal. This is the hand-written-recovery superpower.

**If the lexer hits an unterminated string or unmatched `}`:**
- Hard error: return `*ParseError{Line, Message}` — same type as TOML path. The CLI surfaces it unchanged.

**If `this`/`it`/sourceless `->` appears inside a body:**
- Resolve at converter time (after the AST is built), not in the parser — the parser records a symbolic `Parent` reference; the converter walks scopes. Keeps the parser context-free and testable.

---

## Version Compatibility

| Dependency | Current (go.mod) | Latest verified | Action | Source |
|---|---|---|---|---|
| Go | 1.26.1 | 1.26.x | Keep | [go.mod:3](../../go.mod) |
| `github.com/google/go-cmp` | — (absent) | **v0.7.0** (2025-01-14) | **ADD** as test-only dep | [pkg.go.dev/google/go-cmp](https://pkg.go.dev/github.com/google/go-cmp) |
| `github.com/stretchr/testify` | v1.11.1 | current line | Keep | [go.mod:10](../../go.mod) |
| `github.com/pelletier/go-toml/v2` | v2.2.4 | v2.4.3 (2026-07-05, per v1.10 STACK) | Bump (already recommended v1.10; orthogonal to v1.11) | [pkg.go.dev/go-toml/v2](https://pkg.go.dev/github.com/pelletier/go-toml/v2) |
| Pigeon (rejected) | n/a | v1.3.0 (2024-09-12) | — | [pkg.go.dev/mna/pigeon](https://pkg.go.dev/github.com/mna/pigeon) |
| ANTLR4-go (rejected) | n/a | v4.13.1 (2024-05-15) | — | [pkg.go.dev/antlr4-go/antlr/v4](https://pkg.go.dev/github.com/antlr4-go/antlr/v4) |
| participle (rejected) | n/a | v2.1.4 (2025-03-24) | — | [pkg.go.dev/alecthomas/participle/v2](https://pkg.go.dev/github.com/alecthomas/participle/v2) |
| goyacc (rejected) | n/a | v0.48.0 (legacy) | — | [pkg.go.dev/golang.org/x/tools/cmd/goyacc](https://pkg.go.dev/golang.org/x/tools/cmd/goyacc) |

**Net new runtime dependencies: 0.** Net new test dependencies: 1 (`go-cmp`). Net new codegen/build steps: 0.

---

## Integration Summary

| Component | New file(s) | Touches | Approach |
|---|---|---|---|
| Lexer | `internal/likec4/lexer.go` | none | Hand-rolled byte scanner; tokens carry line/col |
| Parser | `internal/likec4/parser.go` | none | Recursive descent; builds converter-internal AST; panic-mode recovery + `warnDedup` |
| Converter | `internal/likec4/convert.go` | none | AST → `*parser.Model`; flatten braces, resolve `this`/`it`, fuzzy kind map, drop unsupported |
| Extension routing | `cmd/c4drill/root.go` | `ParseFile` call site (~`:112`) | Branch on `.c4`/`.likec4` ext → `likec4.Parse`; `.toml` → existing `parser.ParseFile` |
| Errors | — | reuse `internal/parser/errors.go:13` (`ParseError`) | Converter returns `*parser.ParseError` so CLI path is unchanged |
| Tests | `internal/likec4/*_test.go`, `testdata/*.c4` | go-cmp for AST goldens | Mirror `testdata/*.toml` fixture pattern; COMPAT-style DOT regression reuses `testutil/canonical` |

**Estimated size:** ~600–900 lines across `internal/likec4/` (lexer ~150, parser ~300, convert ~200, tests ~300). Comparable to the existing `internal/parser/parser.go` (~758 lines).

---

## Sources

- Codebase: [PROJECT.md](../PROJECT.md), [likec4-dsl-brief.md](likec4-dsl-brief.md), [parser.go](../../internal/parser/parser.go), [errors.go](../../internal/parser/errors.go), [parser_test.go](../../internal/parser/parser_test.go), [go.mod](../../go.mod), prior [STACK.md](STACK.md) (v1.10)
- [pkg.go.dev — mna/pigeon](https://pkg.go.dev/github.com/mna/pigeon) — v1.3.0, failure-label recovery confirmed
- [pkg.go.dev — antlr4-go/antlr/v4](https://pkg.go.dev/github.com/antlr4-go/antlr/v4) — v4.13.1
- [pkg.go.dev — alecthomas/participle/v2](https://pkg.go.dev/github.com/alecthomas/participle/v2) and [/v2/lexer](https://pkg.go.dev/github.com/alecthomas/participle/v2/lexer) — v2.1.4, no left recursion, stateful lexer
- [pkg.go.dev — golang.org/x/tools/cmd/goyacc](https://pkg.go.dev/golang.org/x/tools/cmd/goyacc) — v0.48.0, 0 importers, legacy
- [pkg.go.dev — google/go-cmp](https://pkg.go.dev/github.com/google/go-cmp) — v0.7.0 (2025-01-14)
- Web search: no native Go LikeC4 or C4-family DSL parser exists (official is TS/Langium; structurizr Go hits are TS/Chevrotain) — low/none confidence in reuse, HIGH confidence in build-from-scratch

---
*Stack research for: C4Drill v1.11 LikeC4 Compatibility Layer*
*Researched: 2026-08-08*

---

## Addendum: Decision Override — PEG Parser (2026-08-09, BINDING)

**Decision:** The user has overridden the hand-written recommendation. **Pigeon (PEG)** is the locked parser technology for v1.11.

### What changed

| Aspect | Hand-written (rejected) | PEG / Pigeon (LOCKED) |
|---|---|---|
| Lexer | Hand-rolled byte scanner (~150 lines) | Folded into the grammar — PEG lexes+parses in one pass |
| Parser source | Hand-written `parser.go` + `lexer.go` (~600–900 lines) | `grammar.peg` (source) → `go generate` → `grammar.go` (generated) |
| Build step | `go build` | `go generate ./internal/likec4/` then `go build` |
| Error recovery | Full manual control (panic-mode skip-to-`}`) | Pigeon `state` with `Throw`/recovery rules — harder but workable |
| AST construction | Returned from hand-written code | Returned from Pigeon action blocks (`{ return node }`) |
| Grammar maintenance | Code reads top-to-bottom but grammar is implicit | Grammar is explicit, declarative, reviewable; easier to fuzz |

### Pigeon specifics (verified versions)

- **Pigeon v1.3.0** (Sep 2024) — mature PEG generator emitting pure Go via `//go:generate pigeon ...`.
- Zero runtime deps (generated parser is self-contained Go).
- Recovery via `state.Throw()` and labeled failures; supports `Results` for custom state threading.
- Action blocks `{ ... }` receive `*ov"` (global state), `pt` (current), `pos` (position); return AST nodes.

### Accepted tradeoff (load-bearing)

The original recommendation's strongest argument against PEG was **error recovery** — the v1.11 headline is "render any model, never fatal," and Pigeon's recovery is weaker than hand-written panic-mode. This is now an **accepted tradeoff**: the "drop unsupported block with a warning" contract must be implemented via **explicit recovery rules in the PEG grammar** (e.g. `UnknownBlock ← Keyword "{" BlockBody* "}"` consuming and discarding), NOT merely absent grammar productions. The roadmap must budget time for recovery-rule design.

### Revised dependency summary

- **Runtime:** zero new deps (Pigeon output is self-contained Go).
- **Codegen-only:** `github.com/mna/pigeon@v1.3.0` — invoked by `go generate`, not imported by production code.
- **Test-only:** `google/go-cmp@v0.7.0` for AST golden diffs (unchanged from original recommendation).

### Revised package layout (replaces the hand-written layout)

```
internal/likec4/
├── grammar.peg       # PEG grammar source (declarative, reviewable)
├── grammar.go        # GENERATED by `go generate` — do not edit
├── ast.go            # AST node types returned by grammar actions
├── convert.go        # AST → *parser.Model (unchanged from arch design)
├── kindmap.go        # kind → UnitType fuzzy mapping
├── resolve.go        # this/it/sourceless -> target resolution at convert time
├── warnings.go       # WarnCollector (dedup by type)
└── likec4.go         # public entry: ParseFile(path, w) (*parser.Model, error)
```

The two-phase public API (`Parse` + `Convert`, with `ParseFile` chaining) from ARCHITECTURE-v1.11.md still holds — Pigeon produces the AST that `Convert` consumes.
