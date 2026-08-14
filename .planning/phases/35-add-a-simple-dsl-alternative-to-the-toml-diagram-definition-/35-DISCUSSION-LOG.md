# Phase 35: Add a simple DSL alternative to the TOML diagram definition - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-08-14
**Phase:** 35-Add a simple DSL alternative to the TOML diagram definition (likec4/d2-style, less verbose syntax) with converters to and from TOML
**Areas discussed:** DSL syntax style, Feature parity mapping, Converter architecture, CLI surface, Template/use placement, Grammar niceties, Docs & examples scope, Emitter style, Formatter (user-added), Parser deps, Reserved keywords, fmt for TOML, Plugin skill update, Comment preservation, use in templates, Field order, Format naming, Duplicate edges, Convert scope, fmt file args, Validate before convert

---

## DSL syntax style

| Option | Description | Selected |
|--------|-------------|----------|
| Indentation blocks | likec4-inspired, nesting via indentation, whitespace-significant | |
| Brace blocks | D2-inspired, explicit { } blocks, arrows anywhere, whitespace-insensitive | ✓ |
| Hybrid | Indentation hierarchy + separate links section | |

**User's choice:** Brace blocks
**Notes:** Same diagram shown in all three styles as previews; user picked the D2 paradigm up front.

| Option | Description | Selected |
|--------|-------------|----------|
| Inline only | Label + [...] options all on the arrow line | |
| Block only | Label on arrow, config always in { } body | |
| Shorthand + block | One-line "tech \| description" for common case, { } when options needed | ✓ |

**User's choice:** Shorthand + block

| Option | Description | Selected |
|--------|-------------|----------|
| id: Type "Name" | Header carries id+type+name; rest inside block; type inferable | ✓ |
| Minimal header | Only id+type on header, name inside block | |
| Rich header | Name, description, technology all on header line | |

**User's choice:** id: Type "Name"

| Option | Description | Selected |
|--------|-------------|----------|
| Barewords + # | Barewords when unambiguous, quotes otherwise, triple-quote multi-line, # comments | ✓ |
| Always quote + // | All strings double-quoted, C-family comments | |

**User's choice:** Barewords + #

| Option | Description | Selected |
|--------|-------------|----------|
| TOML names | Exact type strings: system, person, db, queue, box | ✓ |
| Capitalized words | System, Person, Database, Queue, Box | |

**User's choice:** TOML names

| Option | Description | Selected |
|--------|-------------|----------|
| external modifier | `system external` — scales to any type | ✓ |
| Compound keywords | externalSystem, externalPerson | |
| Star suffix | `system*` | |

**User's choice:** external modifier

| Option | Description | Selected |
|--------|-------------|----------|
| v1.10 semantics | Bare name walks up ancestry; dotted = absolute; reuse internal/peer.Resolve | ✓ |
| Absolute only | Dotted paths from root only | |
| Path-style | Unix-style ../ ./ references | |

**User's choice:** v1.10 semantics

| Option | Description | Selected |
|--------|-------------|----------|
| ASCII | -> <- <-> -- | ✓ |
| Unicode | → ← ↔ · | |

**User's choice:** ASCII

---

## Feature parity mapping

| Option | Description | Selected |
|--------|-------------|----------|
| properties block | `properties { }` top-level, same keys as [properties] | ✓ |
| title + block | `title` line for name, rest in properties block | |

**User's choice:** properties block

| Option | Description | Selected |
|--------|-------------|----------|
| template()/use() | Function-like declaration and instantiation | ✓ |
| TOML-shaped blocks | Declarative blocks mirroring [[use]] shape | |

**User's choice:** template()/use()

| Option | Description | Selected |
|--------|-------------|----------|
| Normalize arrows | No linkFrom concept; converter emits canonical form | |
| Mirror TOML exactly | linkFrom and reverse as distinct constructs | |

**User's choice:** FREE TEXT: "the main idea behind linkFrom is to allow to specify edge on source OR on target. we have to keep this possibility."
**Notes:** Reflected back as: `-> peer` inside a block = link (outgoing), `<- peer` inside a block = linkFrom (incoming) — edges authorable on either endpoint. Confirmed via follow-up.

| Option | Description | Selected |
|--------|-------------|----------|
| Inside blocks only | Every edge attached to its authoring unit | ✓ |
| Also top-level | Freestanding arrows between top-level units | |

**User's choice:** Inside blocks only

| Option | Description | Selected |
|--------|-------------|----------|
| Bare URLs | Scheme-prefixed values are valid barewords (reference) | ✓ |
| Always quoted | No lexer exceptions | |

**User's choice:** Bare URLs

| Option | Description | Selected |
|--------|-------------|----------|
| Single = tech | One value means technology | |
| Single = description | One value means description | ✓ |

**User's choice:** Single = description (user overrode the recommendation)

| Option | Description | Selected |
|--------|-------------|----------|
| Inline [a, b] | Inline bracket lists only | |
| Inline + line form | Both inline and one-per-line accepted | ✓ |

**User's choice:** Inline + line form

| Option | Description | Selected |
|--------|-------------|----------|
| include stmt | `include path [once]` statements | ✓ |
| includes block | `includes { }` block listing files | |

**User's choice:** include stmt

---

## Converter architecture

| Option | Description | Selected |
|--------|-------------|----------|
| Direct to Model | Own lexer/parser → *parser.Model; Model as hub for all four directions | ✓ |
| Transpile to TOML | DSL → TOML text → existing parser | |

**User's choice:** Direct to Model

| Option | Description | Selected |
|--------|-------------|----------|
| Canonical-equivalent | Normalize whitespace/quoting/order (DI-1 precedent); round-trip over fixtures = parity test | ✓ |
| Byte-identical | Exact bytes including comments (impossible: TOML comments) | |
| Semantic only | Model-level equality; text may churn | |

**User's choice:** Canonical-equivalent

| Option | Description | Selected |
|--------|-------------|----------|
| Any mix | DSL includes TOML and vice versa; Model-level merge | ✓ |
| Same-type only | DSL→DSL, TOML→TOML include chains | |

**User's choice:** Any mix

---

## CLI surface

| Option | Description | Selected |
|--------|-------------|----------|
| .c4 extension | Extension-based dispatch | |
| Content sniffing | Inspect file contents | |
| Explicit flag | --lang dsl|toml | |

**User's choice:** FREE TEXT: ".toml is toml, .c4d is dsl"
**Notes:** DSL extension is `.c4d` (not the proposed .c4).

| Option | Description | Selected |
|--------|-------------|----------|
| convert subcommand | `c4drill convert to-toml ...` | |
| --emit flag | Flag on the render command | |
| Both | Subcommand and flag | |

**User's choice:** FREE TEXT: "to-toml and to-c4d"
**Notes:** Follow-up confirmed the shape: `c4drill convert to-toml <file>` / `c4drill convert to-c4d <file>`.

| Option | Description | Selected |
|--------|-------------|----------|
| Render directly | `c4drill diagram.c4d` runs the full pipeline | ✓ |
| Convert-only | DSL feeds converters only | |

**User's choice:** Render directly

| Option | Description | Selected |
|--------|-------------|----------|
| File next to input | Swapped extension; respects -o | ✓ |
| Stdout | Pipe-friendly | |

**User's choice:** File next to input

---

## Template/use placement

| Option | Description | Selected |
|--------|-------------|----------|
| Top-level only | Parity with TOML [[use]] | |
| Nested use too | use inside blocks (exceeds parity) | |

**User's choice:** FREE TEXT: "we need nested use in c4d AND in toml for sure"
**Notes:** Nested use lands in BOTH formats — DSL: use inside blocks; TOML: [[unit.use]] (confirmed in follow-up). Scope expansion accepted by user.

| Option | Description | Selected |
|--------|-------------|----------|
| [[unit.use]] | Array-of-tables nested under unit section | ✓ |
| [[use]] + in field | Top-level use with parent field | |

**User's choice:** [[unit.use]]

| Option | Description | Selected |
|--------|-------------|----------|
| Full unit grammar | Templates contain anything a unit can, incl. edges | ✓ |
| No edges in templates | Structure only | |

**User's choice:** Full unit grammar

---

## Grammar niceties

| Option | Description | Selected |
|--------|-------------|----------|
| Allow `;` | Newline or semicolon separators; one-line nested blocks | ✓ |
| Newline only | One statement per line | |

**User's choice:** Allow `;`

| Option | Description | Selected |
|--------|-------------|----------|
| TOML bare keys | [A-Za-z0-9_-]+ | ✓ |
| Wider ids | Dots, unicode letters | |

**User's choice:** TOML bare keys

---

## Docs & examples scope

| Option | Description | Selected |
|--------|-------------|----------|
| Full docs | README section + .c4d example twins + SKILL.md both formats | ✓ |
| Code only | Docs deferred | |

**User's choice:** Full docs

---

## Emitter style

| Option | Description | Selected |
|--------|-------------|----------|
| Compact leaves | One-line blocks for units without subunits/edges | ✓ |
| Always multi-line | Every unit multi-line | |

**User's choice:** Compact leaves

---

## Formatter (user-added area)

User added via free text: "also we need formatter for c4d files, like the 'gofmt' one"

| Option | Description | Selected |
|--------|-------------|----------|
| fmt --check | In-place rewrite + --check CI mode | ✓ |
| In-place only | No check mode | |

**User's choice:** fmt --check

| Option | Description | Selected |
|--------|-------------|----------|
| Preserve | Comments survive fmt (trivia-aware AST) | ✓ |
| Drop comments | Re-emit from Model | |

**User's choice:** Preserve

| Option | Description | Selected |
|--------|-------------|----------|
| Both formats | fmt handles .c4d and .toml | ✓ |
| .c4d only | As originally asked | |

**User's choice:** Both formats

| Option | Description | Selected |
|--------|-------------|----------|
| Files + dirs | gofmt-style recursion over *.c4d/*.toml | ✓ |
| Single file only | One path per invocation | |

**User's choice:** Files + dirs

---

## Parser deps

| Option | Description | Selected |
|--------|-------------|----------|
| Hand-rolled | stdlib-only recursive descent | |
| Parser generator | participle/PEG library | |

**User's choice:** FREE TEXT: "use PEG, with https://github.com/mna/pigeon"
**Notes:** Specific tool named — mna/pigeon PEG generator, go generate workflow.

---

## Reserved keywords

| Option | Description | Selected |
|--------|-------------|----------|
| Reserve + error | Field names reserved; collision = hard error + suggestion | ✓ |
| Shape-disambiguated | Scalar vs block form decides field vs unit | |

**User's choice:** Reserve + error

---

## Plugin skill update

| Option | Description | Selected |
|--------|-------------|----------|
| Extend in place | c4drill-toml skill keeps name, covers both formats | ✓ |
| New DSL skill | Separate skill alongside | |

**User's choice:** Extend in place

---

## use in templates / field order / naming

| Option | Description | Selected |
|--------|-------------|----------|
| Keep deferred | Template bodies cannot contain use (v1.10 stance) | |
| Enable now | use inside template bodies this phase | ✓ |

**User's choice:** Enable now (lifts the v1.10 template-nesting deferral)

| Option | Description | Selected |
|--------|-------------|----------|
| Fixed canonical | Struct-defined field order in emitted units | ✓ |
| Preserve source order | Parser tracks per-unit field order | |

**User's choice:** Fixed canonical

| Option | Description | Selected |
|--------|-------------|----------|
| C4D | Format named C4D in docs | ✓ |
| Just 'DSL' | No proper name | |

**User's choice:** C4D

---

## Duplicate edges / convert scope / fmt args / validate before convert

| Option | Description | Selected |
|--------|-------------|----------|
| Hard error | Second `-> b` in same block errors (TOML link-map parity) | ✓ |
| Allow list | Multiple edges per pair | |

**User's choice:** Hard error

| Option | Description | Selected |
|--------|-------------|----------|
| Single + --follow | Default single file; --follow-includes converts the graph | ✓ |
| Single only | No graph mode | |

**User's choice:** Single + --follow

| Option | Description | Selected |
|--------|-------------|----------|
| Validate first | Invalid model = hard error, no output | ✓ |
| Best-effort | Convert regardless | |

**User's choice:** Validate first

---

## Claude's Discretion

- Exact type-keyword enumeration and canonical field order
- PEG grammar organization; pigeon error-listener customization
- Package layout for the C4D front-end and emitters
- Emitter spacing rules beyond compact leaves
- Round-trip fixture corpus composition
- fmt flags beyond --check; include-path quoting details

## Deferred Ideas

None. (Template nesting was previously deferred and is now promoted into this phase per user decision.)
