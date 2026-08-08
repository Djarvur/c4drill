# Requirements: v1.11 LikeC4 Compatibility Layer

**Status:** Active
**Milestone:** v1.11 LikeC4 Compatibility Layer
**Last updated:** 2026-08-09

c4drill gains the ability to render LikeC4 (`.c4`) source by on-the-fly converting it into the existing TOML pipeline. A native Go PEG parser (Pigeon) produces an AST; a converter flattens it into a `*parser.Model` inserted as Stage 0 before `parser.ParseFile`. Everything downstream (`include.Resolve → template.Expand → peer.Resolve → Validate → view → render`) runs unchanged. North star: **render any valid LikeC4 source, never fatal on unsupported constructs** — drop `deployments {}`, `views {}`, custom-kind styling, icons, tags/metadata with deduplicated stderr warnings.

Source research: [.planning/research/SUMMARY.md](research/SUMMARY.md) | [LikeC4 DSL brief](research/likec4-dsl-brief.md)

---

## v1.11 Requirements

### PEG — Parser Core (Pigeon)

- [ ] **PEG-01**: `internal/likec4/grammar.peg` lexes+parses the LikeC4 subset: line comments (`//`), block comments (`/* */`), single-quoted (`'...'`) and triple-quoted (`'''...'''`) strings, identifiers, keywords (`specification`/`model`/`views`/`deployment`/`extend`), braces (`{ }`), infix arrows (`->`, `<->`, `-[kind]->`, `.kind.`), `#tags`, optional `:` after property keys. Generated `grammar.go` via `go generate`; zero runtime deps (Pigeon is codegen-only).
- [ ] **PEG-02**: Pigeon action blocks return typed AST nodes (`ast.File`, `ast.Specification`, `ast.Model`, `ast.Element`, `ast.Relationship`, `ast.Property`) carrying source position (line:column) for error reporting. The AST is the typed handoff to the converter.
- [ ] **PEG-03**: Explicit PEG recovery rules — `UnknownBlock` consumes `{ ... }` for unrecognized top-level/element keywords (so `views {}`, `deployments {}`, `global {}` parse-and-discard rather than fatal); `UnknownStatement` skips to the next statement boundary. **Load-bearing** for the "never fatal" contract — the accepted tradeoff of PEG over hand-written (recovery is harder, so it must be designed in, not absent).
- [ ] **PEG-04**: Parse errors report the `.c4` source line:column via the existing `parser.ParseError` type (or a compatible error), not an opaque byte offset. Pigeon global state threads position so the user sees `bigbank.c4:42:5: expected '}'`.

### CONV — Converter (AST → `*parser.Model`)

- [ ] **CONV-01**: Recursive `emitElement` flattens LikeC4 lexical `{}` nesting into `Unit.Subunits`/`SubunitOrder` keyed by last path segment, mirroring `parser.parseUnitWithOrder`. FQN construction matches LikeC4's lexical-scope rule (parent path + child name).
- [ ] **CONV-02**: `->` maps to `model.Link` with `Peer` resolved to an absolute dotted path at convert time; `<->` maps to **two** `Link` records (preserving the HS-1 multiplicity contract — a single bidirectional record would re-break multiplicity counting). Sourceless `->`/`this`/`it` resolve to the enclosing parent element; relationship kind (e.g. `-[async]->`) is concatenated into `Link.Technology` as preserved text.
- [ ] **CONV-03**: Three-tier fuzzy kind→`UnitType` mapping: exact (`person`/`actor`→person, `database`/`db`→db, `queue`→queue, `container`→container, `system`/`enterprise`/`softwaresystem`→system, `box`→box) → substring (`data`→db, etc.) → fallback `box`. A kind used as a nesting parent (i.e. an element with children) promotes to `box` so subunits are legal. LikeC4 has NO built-in kinds and NO kind inheritance, so this table is load-bearing — an unknown kind is never a hard error.
- [ ] **CONV-04**: LikeC4 element properties map to `model.Unit` fields: `title`→`Unit.Name` (set explicitly to bypass `model.Humanize`, avoiding `gRPC`→"Grpc" mangling), `description`→`Unit.Description`, `technology`→`Unit.Technology`, `link <url>`→`Unit.Reference` (rides v1.10's 📖 marker).
- [ ] **CONV-05**: Per-element `style { shape ... }` overrides the kind-driven `UnitType` when the shape names a C4Drill type (`person`, `cylinder`→db, `queue`, `browser`/`mobile`→container, `rectangle`/`component`/`storage`→component). Aligns with the `dev/shapes-no-icons` branch — icon styles are silently dropped.

### INT — Integration & UX

- [ ] **INT-01**: `switch` on `filepath.Ext` in `cmd/c4drill/root.go:runRoot` routes `.c4`/`.likec4` inputs to `likec4.ParseFile`; the default case routes to the existing `parser.ParseFile`. **Hard regression contract:** every existing `.toml` fixture (testdata/, cmd/c4drill/testdata/, skill/examples/) parses byte-identical — the `.toml` code path is untouched.
- [ ] **INT-02**: `WarnCollector` with `seen map[string]bool` keyed by construct-type tag (`views-block`, `deployment-block`, `tags`, `metadata`, `icon`, `relationship-kind`, `style-block`) emits one warning per TYPE to `os.Stderr`, never per occurrence. Warnings survive `-o` output redirection (stderr is independent of output files).
- [ ] **INT-03**: `bigbank.c4` (the 228-line canonical LikeC4 showcase) committed as testdata; an end-to-end CLI test renders it without fatal error and produces valid DOT verified via `canonical.Canonical`. This is the "render any model" proof.
- [ ] **INT-04**: Single-file `extend <element> {}` — accumulate children/properties/tags into an element declared earlier in the same `.c4` file (accumulator-merge, not overwrite). Cross-file workspace `extend` (LikeC4's implicit multi-file merge) is deferred to v2.

### DX — Differentiators & Docs

- [ ] **DX-01**: `skill/examples/10-likec4-bigbank.c4` runnable fixture plus a README "LikeC4 Compatibility" section documenting the extension routing, the supported subset, and the unsupported-with-warning set. Matches the v1.10 docs pattern (DOC-01/DOC-02/DOC-03).
- [ ] **DX-02**: TOML↔LikeC4 cross-format equivalence golden — author a small model in both TOML and `.c4`, prove both produce canonical-identical DOT (the analog of v1.10's XC-05 composed-vs-single-file proof). Locks the converter's semantic correctness against the native parser.
- [ ] **DX-03**: `cmd/c4drill/root_test.go` CLI regression proving `.c4` input end-to-end (parse → convert → validate → render → DOT/SVG on disk) and that `.toml` inputs still render identically (backward-compat guard).

---

## Future Requirements (deferred)

- **F-01**: Cross-file workspace `extend` merging — LikeC4's implicit multi-file workspace composition (every `.c4` file in a directory merges into one model). v1.11 ships single-file `extend` only; multi-file would ride v1.10's `[[include]]` pipeline.
- **F-02**: LikeC4 `views {}` predicate translation — include/exclude predicates, `where` filters, `group`, `autoLayout`. Deferred indefinitely; conflicts with C4Drill's auto-generation philosophy.
- **F-03**: `deployments {}` block support — LikeC4's deployment model (separate environment/instance tree). Out of scope; C4Drill has no deployment concept.

---

## Out of Scope

- **TypeScript/Node embedding** of `@likec4/parser` — locked rejection; adds a heavy runtime dep to a pure-Go CLI.
- **Tree-sitter grammar** — locked rejection; adds a CGO/native dependency.
- **WASM-compiled LikeC4 parser** — locked rejection; same reasoning.
- **Icons** (LikeC4 `icon`, `iconColor`, `iconSize`, `iconPosition`) — out of scope; project is on `dev/shapes-no-icons` branch (icons were removed in v1.6).
- **Tags/metadata visual rendering** — out of scope; tags/metadata are dropped with a warning. LikeC4 uses them for view predicates (which we don't translate) and styling (which we don't carry).
- **Full `views {}` DSL** — out of scope; C4Drill auto-generates C1/C2/C3 from the model tree. The entire `views {}` block is dropped with one warning.
- **Custom element-kind styling beyond shape/color** — moot; LikeC4 has NO kind inheritance and NO built-in kinds, so there's no styling taxonomy to carry.
- **`specification` kind-default inheritance** — deferred; kind defaults (`title`/`description`/`technology`/`notation` declared on `element <kind>`) could flow into elements that omit them, but this is polish for a later milestone.

---

## Traceability

(Filled by roadmapper — maps each REQ-ID to its phase.)

| REQ-ID | Phase | Success Criteria |
|--------|-------|------------------|
| PEG-01 | 34 | bigbank.c4 parses via generated grammar.go; AST goldens lock typed node shape; `go generate` produces zero-runtime-dep code |
| PEG-02 | 34 | AST golden tests on bigbank.c4 + minimal fixture lock typed nodes (File/Specification/Model/Element/Relationship/Property) carrying source line:column |
| PEG-03 | 34 | `views {}`/`deployments {}`/`global {}` and unrecognized blocks parse-and-discard via explicit `UnknownBlock`/`UnknownStatement` recovery rules — no fatal |
| PEG-04 | 34 | Parse failures surface `bigbank.c4:LINE:COL:` via the existing `*parser.ParseError` shape, not an opaque byte offset |
| CONV-01 | 35 | Nested `{}` tree converts to dotted-path subunits (`Subunits` keyed by last segment + `SubunitOrder` in source order), structurally identical to `parser.parseUnitWithOrder` output |
| CONV-02 | 35 | `<->` → TWO Links (Mirror:false, HS-1); `this`/`it`/sourceless `->` resolve to absolute paths (peer.Resolve no-op); kind concatenated into Link.Technology; unresolved ref = HARD error |
| CONV-03 | 35 | Three-tier fuzzy kind table (exact → substring → box fallback); nesting-parent promotes to box; unknown kind NEVER a hard error |
| CONV-04 | 35 | title→Name VERBATIM (bypasses Humanize, Pitfall 13); description/technology/link map to Description/Technology/Reference; reference rides v1.10 📖 marker |
| CONV-05 | 35 | Per-element `style { shape: cylinder }` overrides UnitType; icon styles silently dropped (dev/shapes-no-icons) |
| INT-01 | 36 | Stage 0 switch in `cmd/c4drill/root.go:runRoot` routes `.c4`/`.likec4`→converter, default→parser.ParseFile; every existing `.toml` fixture byte-identical (DI-1 canonical-DOT) |
| INT-02 | 36 | `WarnCollector` dedups by construct TYPE to `os.Stderr`; survives `-o` redirection; one warning per type per file regardless of occurrence count |
| INT-03 | 37 | `testdata/bigbank.c4` (228 lines) renders end-to-end without fatal, produces valid DOT via `canonical.Canonical` — the "render any model" proof |
| INT-04 | 36 | Single-file `extend <element> {}` accumulates (not overwrites); cross-file workspace extend emits a one-line deferral warning (F-01) rather than fataling |
| DX-01 | 37 | `skill/examples/10-likec4-bigbank.c4` runnable fixture; README "LikeC4 Compatibility" section documents routing + supported subset + dropped-with-warning set |
| DX-02 | 37 | TOML↔LikeC4 cross-format equivalence golden (canonical-identical DOT); analog of v1.10 XC-05 |
| DX-03 | 37 | `cmd/c4drill/root_test.go` CLI regression: `.c4` end-to-end (DOT+SVG on disk) AND `.toml` renders identically to v1.10 (backward-compat guard) |

**Coverage:** 16/16 requirements mapped (100%). Phases 34-37, four phases total.
