# Pitfalls Research — v1.11 LikeC4 Compatibility Layer

**Domain:** Adding a native LikeC4 DSL parser + on-the-fly converter to an existing TOML pipeline (c4drill v1.11)
**Researched:** 2026-08-08
**Confidence:** HIGH (grounded in repo source — `parser.go`, `unit.go`, `link.go`, `peer/resolve.go`, `include/merge.go`, `root.go` — and the LikeC4 DSL brief)
**Scope:** Pitfalls for the new pipeline Stage 0 (LikeC4 parse + convert) feeding the existing `include.Resolve → template.Expand → peer.Resolve → Validate → view → render` stages.

> Supersedes the v1.10 PITFALLS doc. The v1.10 HS-1 mitigation (`Unit.Clone` preserving `Link.Mirror`) is now load-bearing infrastructure the v1.11 converter must NOT regress.

The NORTH STAR ("Render any model — never fatal on unsupported constructs; drop with warnings") is the through-line. Most pitfalls below are ways the converter could accidentally violate it.

---

## Critical Pitfalls

### Pitfall 1: `->` is grammar-ambiguous (relationship vs. view predicate)

**What goes wrong:** A naive parser sees `customer -> cloud` and consumes it as a relationship. But the same `->` inside `views { }` (`include -> customer`, `-> singlePageApplication ->`) is a view predicate the converter must DROP. Worse: sourceless `-> target` means "incoming edges to target" in a view but "implicit-parent-as-source relationship" in a model element body. Same three characters, three semantics.

**Why it happens:** LikeC4 reuses `->` across two layers. Without context-tracking the enclosing block, the parser cannot disambiguate.

**How to avoid:** Carry a parser context stack (`specification` / `model` / `element-body` / `views` / `view-body`). Emit links only when context is `model` or `element-body`. Any `->` inside `views`/`view-body` is dropped.

**Warning signs:** Converter emits spurious links named `customer`/`include`/`*`; C1 diagrams show phantom unlabeled arrows.
**Phase to address:** DSL parser phase (Stage 0) — context stack is foundational.

---

### Pitfall 2: Optional `:` after property keys creates two valid shapes

**What goes wrong:** `description "x"` and `description: "x"` are BOTH valid LikeC4 (the bigbank example mixes both). A parser hard-coding one shape rejects half of real-world files. Triple-quoted strings (`'''...'''`, `"""..."""`) compound this: a `:` followed by `'''multi-line\nmarkdown'''` breaks naive single-line tokenizers.

**Why it happens:** Property-key parsing is easy to write as `IDENT (':')? VALUE` — and equally easy to forget or over-consume the colon.

**How to avoid:** Treat `:` as an optional separator after every property keyword. Tokenize single `'...'`, double `"..."`, AND triple `'''...'''`/`"""..."""` as one string category. Write at least one fixture using BOTH colon styles.

**Warning signs:** Parse fails on `bigbank.c4`; user reports "works when I delete the colon."
**Phase to address:** DSL parser phase — lexer + property-key rule.

---

### Pitfall 3: Comments and strings confuse a brace-counting flattener

**What goes wrong:** LikeC4 has `//` line comments AND `/* */` block comments. A `}` inside a comment (`// end of block }`) or inside a string (`description "use the } character"`) fools a naive brace-depth counter into closing a scope early. Result: orphaned subunits, mis-flattened dotted paths, silently wrong diagrams.

**Why it happens:** The converter must flatten lexical `{}` into C4Drill dotted-path subunits — and only the lexer knows what's a comment/string vs. a real brace.

**How to avoid:** The LEXER strips comments and string contents (returns them as token payloads, never as raw `}`). Brace-depth state lives in the parser and operates on tokens, never raw bytes. Never run a `strings.Count(src, "{")` heuristic.

**Warning signs:** A model with one block comment renders with subunits missing; `/* } */` produces a different structure than the same model uncommented.
**Phase to address:** Lexer + parser phase.

---

### Pitfall 4: `this`/`it`/sourceless `->` resolve to the WRONG scope

**What goes wrong:** Inside an element body, `it -> frontend`, `frontend -> this`, and `-> frontend` all refer to the parent element. If the converter emits them verbatim as `Link.Peer = "this"`, then `peer.Resolve` walks ancestry looking for a sibling named "this" — finds nothing (or worse, finds a user's element literally named `this`) and either hard-errors or silently mis-resolves.

**Why it happens:** The parser records the literal token; resolution requires the enclosing-element path threaded through every nested descent. Deferring to peer.Resolve is wrong — its D-13 contract is "bare name = sibling lookup", a different rule.

**How to avoid:** Resolve `this`/`it`/`->` during conversion using the converter's element-path stack. Emit ONLY absolute dotted paths into `Link.Peer`. Absolute paths short-circuit `peer.Resolve` untouched (D-16 gate: `strings.Contains(peer, ".")`).

**Warning signs:** Links vanish or attach to the wrong source; `peer.Resolve` errors with "cannot resolve peer 'this'".
**Phase to address:** Converter phase.

---

### Pitfall 5: Lexical-scope name resolution (bubbling) is the hardest part

**What goes wrong:** LikeC4 "bubbles" unique nested names to outer scope: `frontend -> service1.backend.api` can be written `frontend -> api` if `api` is unambiguous across the whole workspace. A converter that only resolves bare names against immediate siblings will miss cross-scope references. Conversely, `extend cloud { service1 -> service2 }` re-opens `cloud`'s scope; `service1`/`service2` resolve against `cloud`'s children, NOT the file's top level.

**Why it happens:** LikeC4 resolution is lexical-scope + hoisting + uniqueness (JS-like). C4Drill's `peer.Resolve` is ancestor-walk-up (D-13/14/15). They agree on sibling/root resolution but DIVERGE on cross-scope bubbling.

**How to avoid:** The converter builds a full FQN index and resolves every name to its absolute path BEFORE emitting. Emit only absolute paths. Do NOT rely on `peer.Resolve` to do LikeC4-style resolution — it can't. Hoisting means the converter may need two passes (collect FQNs, then resolve references).

**Warning signs:** "cannot resolve peer 'api'" for a model where `api` exists nested; converter works on flat models but fails on 3-level nesting.
**Phase to address:** Converter phase — name resolution is its own sub-component.

---

### Pitfall 6: TOML byte-identical regression (the hard contract)

**What goes wrong:** v1.11 requires every existing `.toml` fixture to parse byte-identical. The converter shares the `*parser.Model` struct. Any change to `Model`/`Unit`/`Link` (a new field, re-ordered unmarshal, default-value tweak) silently changes TOML output. The dispatch in `cmd/c4drill/root.go:117` calls `parser.ParseFile(inputPath)` unconditionally — wiring the converter THERE makes every TOML input run through converter code paths.

**Why it happens:** The converter produces the same `*parser.Model` type, so it's tempting to route everything through one entry point or add converter-only fields to shared structs.

**How to avoid:**
- Dispatch by extension BEFORE `parser.ParseFile` — `.c4`/`.likec4` → converter, everything else → existing parser unchanged.
- The converter must NOT modify `parser.Model`, `model.Unit`, or `model.Link` struct definitions. New fields go in converter-internal types.
- Keep canonical-DOT golden tests (`internal/testutil/canonical`) green for every TOML fixture.

**Warning signs:** Canonical-DOT golden diff on a TOML fixture; `Unit.SubunitOrder` ordering shifts for TOML input.
**Phase to address:** Converter wiring phase — extension dispatch first; goldens are the gate.

---

### Pitfall 7: `<->` bidirectional — one link or two?

**What goes wrong:** LikeC4 `<->` is bidirectional. Two valid C4Drill encodings exist: (a) one link with `Arrow: bidirectional`, (b) two links (forward + reverse). The validator's `populateIncomingLinks` synthesizes mirror `LinksFrom` from outgoing `Links` (the v1.10 HS-1 concern). If the converter picks (a) but the validator's multiplicity logic (D-05/WR-02) assumed (b), bidirectional edges miscount. The unexported `Link.Mirror` field (the one that bit v1.10's template.Expand) is in play: converter-emitted links must have `Mirror=false`.

**Why it happens:** PROJECT.md says "`<->` becomes two links" — but implementers may reach for `Arrow: bidirectional` because it's one record.

**How to avoid:** Follow PROJECT.md: `<->` → two `Link` entries (forward + reverse), both with `Mirror: false` (the converter is the author, not the validator). Cross-check `cloneLinks` in `unit.go` and `FindLinkByPeer` (returns first match — order matters for duplicates).

**Warning signs:** Multiplicity counts differ; mirror links appear "doubled" after validation; `WR-02` regression fails.
**Phase to address:** Converter phase — gated by an HS-1-style regression test for bidirectional edges.

---

### Pitfall 8: LikeC4 has NO built-in kinds — fuzzy mapping is load-bearing

**What goes wrong:** C4Drill's `UnitType` is a fixed enum. LikeC4 kinds are arbitrary user strings (`actor`, `enterprise`, `softwaresystem`, `spa`, `mobileApp`, `microservice`, `pgTable`). Passing `enterprise` straight through as `UnitType` makes the renderer silently mishandle it. Worse: `inferGenericType` (`parser.go:699`) only special-cases `db`/`queue` — every other unknown kind falls through to whatever `defaultTypeForParent` returns.

**Why it happens:** PROJECT.md says fuzzy-match kinds, but fuzzy matching is easy to get wrong (case sensitivity, substring vs. equality, `softwaresystem` vs. `softwareSystem`). The `box` fallback silently groups everything that doesn't match — a valid-but-wrong diagram (NORTH STAR violation in spirit).

**How to avoid:** Explicit kind→`UnitType` table with case-insensitive lookup; reserve `box` for true unknowns. Emit a warning per dropped/mapped kind (deduped by KIND NAME — see Pitfall 11). Test against bigbank's 10 kinds (`person`, `enterprise`, `staff`, `existingsystem`, `softwaresystem`, `spa`, `mobileApp`, `container`, `component`, `database`).

**Warning signs:** `enterprise` renders as `box`; `database` maps to `db` but `Database` (capitalized) silently falls back.
**Phase to address:** Converter phase — kind-mapping table is discrete.

---

### Pitfall 9: Dotted-path collision when flattening nested `{}`

**What goes wrong:** LikeC4 lexical nesting flattens to C4Drill dotted keys: `cloud { api { controllers } }` → `cloud`, `cloud.api`, `cloud.api.controllers`. C4Drill TOML uses dotted table names with a 2-segment hand-authored limit (`recordHandAuthored` in `parser.go:555` ignores `len(parts) > 2`). The converter bypasses this by building `*model.Unit` directly, BUT `Unit.Subunits` is keyed by the LAST path segment, not the full path — so two siblings with the same name, or a converter bug mis-tracking the parent stack, silently overwrite.

**Why it happens:** Map-keyed children where the key isn't the full path invites collision.

**How to avoid:** Maintain an explicit `(parentFullDottedPath, childName)` tracker. Detect duplicate full paths early and (per NORTH STAR) drop the later one with a warning. Mirror the `produced *pathTracker` pattern from `template/expand.go`.

**Warning signs:** Subunits silently disappear; `len(Subunits)` differs from `len(SubunitOrder)`.
**Phase to address:** Converter phase — flattening logic.

---

### Pitfall 10: `extend` re-opens scope — merge semantics

**What goes wrong:** LikeC4 `extend cloud { ... }` adds children/properties/tags to an element defined elsewhere. C4Drill's `include/merge.go` does similar (`mergeSubunits`), but the converter sees a SINGLE file (per PROJECT.md), so multiple `extend cloud {}` blocks in the same file must be merged by the converter itself. A naive converter treats each `extend` as a fresh declaration and overwrites the previous one.

**Why it happens:** `extend` looks like a declaration syntactically. The merge semantics (tags appended, links appended, metadata merged with dedup) live in LikeC4-land.

**How to avoid:**
- Single-file: the converter accumulates `extend` blocks per element FQN and merges (tags union, links append, properties first-wins).
- Multi-file: out of scope for v1.11 (PROJECT.md says single-file). Document the boundary.
- The `extend relationship` form (`extend a -> b {}`) — drop with warning.

**Warning signs:** Element properties flip-flop based on which `extend` block processed last; tags that should accumulate vanish.
**Phase to address:** Converter phase — `extend` handling is its own sub-component.

---

### Pitfall 11: Warning dedup, stderr routing, and suppression under `-o`

**What goes wrong:** NORTH STAR drops unsupported constructs with warnings. Three sub-pitfalls:
1. **Per-occurrence flood**: 50 `metadata {}` blocks emit 50 identical warnings. Dedup to one per TYPE (`metadata`, `tags`, `views`, `deployments`, `icons`, `extend relationship`).
2. **stdout pollution**: `-f dot` writes DOT to stdout — a stray warning line corrupts the output stream.
3. **Silent under `-o`**: when `-o` redirects output to a file, warnings must still reach the terminal via stderr.

**Why it happens:** It's natural to use `fmt.Println` for warnings. The pipeline is silent-on-success by design (`root.go:186`), so any new stderr traffic is a contract change.

**How to avoid:**
- `seen map[string]bool` of warning TYPES; emit each once.
- Route to `os.Stderr` explicitly — NOT `cmd.OutOrStderr()` (Cobra redirection) or `cmd.OutOrStdout()`.
- Fixture asserting warnings appear on stderr even when `-o` writes to a file.

**Warning signs:** `c4drill model.c4 -f dot -o out/` produces `out/diagram.dot` with warning text prepended; `2>/dev/null` swallows everything.
**Phase to address:** Converter phase + CLI wiring.

---

### Pitfall 12: Parse errors must point to the `.c4` source line

**What goes wrong:** `parser.ParseError` carries a `Line` field; `wrapDecodeError` extracts it from go-toml. If the converter raises errors using a different type or loses the original `.c4` line (reporting the generated IR's line, or `Line: 0`), users can't locate the problem.

**Why it happens:** The converter is multi-stage (lex → parse → flatten → emit). By the time an error surfaces, the original `.c4` line may not be carried through.

**How to avoid:** Every AST node carries its source line (lexer captures per-token). Converter errors embed the original `.c4` line — never the generated model's. Reuse the `*parser.ParseError` shape (or a sibling type with the same `Line`/`Context`/`Message`/`Cause`) so `root.go`'s `fmt.Errorf("parse: %w", err)` is consistent.

**Warning signs:** User reports "error says line 1 but the problem is on line 200"; all converter errors report `Line: 0`.
**Phase to address:** Converter phase — error infrastructure from day one.

---

## Pipeline Interaction Pitfalls

### Pitfall 13: `model.Humanize` double-humanizes LikeC4 titles

**What goes wrong:** LikeC4 elements have a human-readable title (`customer = actor 'Personal Banking Customer'`). C4Drill's parser applies `model.Humanize(name)` ONLY when `unit.Name == ""` (parser.go:614). If the converter sets `Unit.Name` from the title, the hook doesn't fire (good). If it forgets and relies on the identifier (`customer`), the parser humanizes — usually fine, but `gRPC` → "Grpc" (no acronym preservation per ERGO-04) mangles titles containing acronyms.

**Why it happens:** The converter must decide: is `Unit.Name` the title, the identifier, or empty?

**How to avoid:** Always set `Unit.Name` from the LikeC4 title when present. Fall back to identifier only if no title. Never leave `Name` empty for the parser to humanize.
**Warning signs:** "Mobile App" renders correctly but "gRPC Frontend" renders as "Grpc Frontend".
**Phase to address:** Converter phase — name-field mapping.

---

### Pitfall 14: `peer.Resolve` expects relative peers — converter must emit absolute

**What goes wrong:** `peer.Resolve` (D-16) treats `Link.Peer` containing `.` as absolute and bare names as relative (walk-up lookup). If the converter emits BARE LikeC4 names (`api` for `cloud.api`), `peer.Resolve` mis-resolves using C4Drill's ancestor-walk — DIFFERENT from LikeC4's lexical-bubbling (Pitfall 5). Result: silent mis-resolution OR a hard `ResolveError` that fatals — NORTH STAR violation.

**Why it happens:** The converter is the only stage that knows LikeC4 scoping. Deferring resolution is tempting but wrong.

**How to avoid:** Converter emits ONLY absolute dotted paths in `Link.Peer` and `LinksFrom.Peer`. `peer.Resolve` becomes a no-op for `.c4` output. Keep peer.Resolve in the pipeline for TOML backward-compat.
**Warning signs:** `ResolveError: cannot resolve peer "api" from unit "frontend"` where `api` exists nested.
**Phase to address:** Converter phase — link emission.

---

### Pitfall 15: `template.Expand` and `include.Resolve` could misfire on converter output

**What goes wrong:** The pipeline is `converter → include.Resolve → template.Expand → peer.Resolve → Validate`. A `.c4` file produces a `*parser.Model` with empty `Templates`/`Instantiations`/`Includes`. The fast-path in `template.Expand` (`len(m.Templates) == 0 && len(m.Instantiations) == 0` → return unchanged) saves us, and `include.Resolve` is a no-op on empty `Includes`. BUT: `include.merge` uses reflection over `model.Properties` (merge.go:129) and can panic if the converter produces a `Properties` with unexpected zero-values. `mergeSubunits` hard-errors if two files redefine the same parent — not a concern single-file, but a concern if `.c4` + `.toml` composition is ever enabled.

**Why it happens:** Downstream passes assume TOML-originated models. "Structurally identical" ≠ "semantically equivalent" for every edge case.

**How to avoid:**
- Verify `template.Expand` and `include.Resolve` are genuine no-ops on converter output.
- Converter-emitted `model.Properties` must use TOML zero-value conventions (`Color: ""` not `Color: "transparent"`).
- Do NOT let converter output flow through `template.Expand`/`include.Resolve` in ways TOML doesn't.

**Warning signs:** `.c4` model renders fine alone but panics combined with `[[include]]`; `mergeProperties` spurious conflict.
**Phase to address:** Integration phase — pipeline-ordering test.

---

### Pitfall 16: `mobileApp = mobileApp "..."` — name/kind namespace collision

**What goes wrong:** LikeC4 allows an element's NAME to equal its KIND (both `mobileApp`s are the same identifier) — legal because kinds and names live in different namespaces. A converter that resolves the kind token against a name table confuses the two: it treats the kind `mobileApp` as a reference to the element, or fails to find the kind.

**Why it happens:** Kind-first and name-first syntaxes (`component backend {...}` vs `backend = component {...}`) put kind and name in adjacent token positions.

**How to avoid:** Parser distinguishes kind-position from name-position syntactically. Maintain separate symbol tables: `kinds map[string]bool` (from `specification`) and `elements map[string]*Unit` (from `model`). Never cross-resolve.
**Warning signs:** "unknown kind 'mobileApp'" for a file declaring `element mobileApp` in spec; FQN resolution returns the wrong element.
**Phase to address:** Converter phase — symbol-table design.

---

## Technical Debt Patterns

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| Emit bare names, let peer.Resolve handle it | Less converter code | Silent mis-resolution (Pitfall 14); NORTH STAR violation | Never |
| Use `Arrow: bidirectional` for `<->` | One link instead of two | Multiplicity breaks (Pitfall 7) | Never — follow PROJECT.md |
| Drop `extend` with a warning | Faster MVP | Multi-file workspaces break (Pitfall 10) | v1.11 MVP only |
| Reuse TOML struct fields for converter data | Fewer types | TOML byte-identical regression (Pitfall 6) | Never |
| Skip source-line tracking in AST | Simpler parser | Useless errors (Pitfall 12) | Never |
| Fuzzy-match kinds by substring | Trivial code | `spa` matches `space` | Never — use a lookup table |
| Treat `extend` as fresh declaration | Simpler converter | Later overwrites earlier (Pitfall 10) | Never — accumulate per FQN |

## Integration Gotchas

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| `cmd/c4drill/root.go` dispatch | Wiring converter inside `parser.ParseFile` | Dispatch by extension BEFORE `ParseFile` (Pitfall 6) |
| `peer.Resolve` | Expecting LikeC4 lexical resolution | Converter emits absolute paths; peer.Resolve is a no-op for `.c4` (Pitfall 14) |
| `template.Expand` | Assuming converter output has templates | Confirm fast-path no-op; empty Templates (Pitfall 15) |
| `include.merge` | Letting reflection-based mergeProperties see converter Properties | Match TOML zero-value conventions (Pitfall 15) |
| `validator.Validate` | Fatal on unsupported kinds | Converter maps all kinds before validate (Pitfall 8) |
| Cobra output routing | `cmd.OutOrStderr()` for warnings | `os.Stderr` directly; never stdout (Pitfall 11) |

## Performance Traps

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|----------------|
| Re-resolving FQNs per relationship | O(N²) on large models | Build FQN index once; lookup | 500+ relationships |
| Re-parsing spec block per kind lookup | Slow parse on wide specs | Build `kinds` map once | 50+ kinds |
| String-keyed Subunits iteration | Non-deterministic output | Always populate `SubunitOrder` | Any model — v1.7 determinism contract |
| Regex-based comment stripping | Breaks on `//` inside strings | Lexer-based stripping | Any model with `//` in a description |

## Security Mistakes

| Mistake | Risk | Prevention |
|---------|------|------------|
| Trusting `link <url>` for HTML output | XSS via `javascript:` URLs | Reuse v1.10 Phase 28 HTML-shim http(s)-only gate |
| Path traversal via `link ../...` | Local file disclosure | Treat links as opaque; shim no-ops non-http(s) |
| `metadata { json_key '{...}' }` as JSON | Injection if rendered | Drop metadata with warning (NORTH STAR); never render user JSON |

## UX Pitfalls

| Pitfall | User Impact | Better Approach |
|---------|-------------|-----------------|
| Flood of identical warnings | Warning fatigue | Dedup per construct TYPE (Pitfall 11) |
| Fatal on unsupported construct | NORTH STAR violated | Drop with warning; never fatal |
| Error points to generated model line | User can't find problem | Carry `.c4` source line (Pitfall 12) |
| Silent `box` fallback for unknown kinds | Valid-but-wrong diagram | Per-kind warning naming from/to pair (Pitfall 8) |
| `gRPC` title rendered as `Grpc` | Mangled title | Use LikeC4 title verbatim (Pitfall 13) |

## "Looks Done But Isn't" Checklist

- [ ] **Converter:** Drops `views {}` entirely — NO view-predicate `->` leaks as a link (Pitfall 1)
- [ ] **Converter:** Handles BOTH `description "x"` AND `description: "x"` — bigbank.c4 parses (Pitfall 2)
- [ ] **Converter:** Comments/strings don't break brace-depth — `/* } */` and `description "}"` work (Pitfall 3)
- [ ] **Converter:** `this`/`it`/`->` resolve to absolute paths — `peer.Resolve` no-op on output (Pitfall 4)
- [ ] **Converter:** Cross-scope bubbling works — `frontend -> api` resolves to `cloud.backend.api` (Pitfall 5)
- [ ] **TOML path:** Every existing `.toml` fixture byte-identical — canonical-DOT goldens unchanged (Pitfall 6)
- [ ] **Converter:** `<->` emits TWO links with `Mirror: false` — multiplicity test (Pitfall 7)
- [ ] **Converter:** Kind mapping covers bigbank's 10 kinds (Pitfall 8)
- [ ] **Converter:** Duplicate-path detection — `pathTracker`-style check (Pitfall 9)
- [ ] **Converter:** `extend` blocks merge, not overwrite — two `extend cloud {}` test (Pitfall 10)
- [ ] **CLI:** Warnings deduped, on stderr, survive `-o` — `2>/dev/null` and `-o` test (Pitfall 11)
- [ ] **Errors:** Carry `.c4` source line — error-line assertion (Pitfall 12)
- [ ] **Names:** LikeC4 titles used verbatim — `gRPC` stays `gRPC` (Pitfall 13)
- [ ] **Pipeline:** template.Expand + include.Resolve no-op on `.c4` output (Pitfall 15)

## Recovery Strategies

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| TOML regression (6) | LOW | Revert converter wiring; dispatch by extension only |
| Bad kind mapping (8) | LOW | Extend the kind table; re-render |
| Warning flood (11) | LOW | Add dedup set; re-run |
| Mis-resolution of `this`/`it` (4) | MEDIUM | Thread element-path stack; re-test bigbank |
| Dotted-path collision (9) | MEDIUM | Add pathTracker; surface duplicates as warnings |
| Scope bubbling broken (5) | HIGH | Rebuild name resolution with full FQN index + two-pass |
| `extend` overwrite (10) | HIGH | Re-architect converter around per-element accumulation |

## Pitfall-to-Phase Mapping

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| 1 (`->` ambiguity) | DSL parser | View predicates don't appear as links |
| 2 (optional `:`) | Lexer | bigbank.c4 parses with both styles |
| 3 (comments/strings) | Lexer | `/* } */` and `"}"` don't break flattening |
| 4 (`this`/`it`) | Converter | peer.Resolve no-op on `.c4` output |
| 5 (scope bubbling) | Converter | Cross-scope reference resolves |
| 6 (TOML regression) | Wiring | Canonical-DOT goldens unchanged |
| 7 (`<->` two links) | Converter | Multiplicity regression test |
| 8 (kind mapping) | Converter | bigbank's 10 kinds map correctly |
| 9 (path collision) | Converter | Duplicate-path warning test |
| 10 (`extend` merge) | Converter | Two `extend` blocks merge test |
| 11 (warning UX) | CLI wiring | stderr + dedup + `-o` survival test |
| 12 (source lines) | Converter | Error-line assertion test |
| 13 (Humanize) | Converter | `gRPC` title preserved |
| 14 (absolute peers) | Converter | peer.Resolve no-op test |
| 15 (pipeline no-op) | Integration | template.Expand/include.Resolve no-op on `.c4` |
| 16 (name/kind) | Converter | `mobileApp = mobileApp` parses |

## Sources

- `.planning/PROJECT.md` — v1.11 NORTH STAR, kind-mapping decision, `<->` two-link decision, single-file scope
- `.planning/research/likec4-dsl-brief.md` — DSL grammar (sections 1-10), bigbank.c4 canonical example, scope/hoisting rules, `extend` semantics, no-built-in-kinds confirmation
- `internal/parser/parser.go` — `Parse`, `captureDefinitionOrder`, `parseUnitWithOrder`, `inferGenericType`, Humanize hook at line 614, 2-segment hand-authored limit at `recordHandAuthored`
- `internal/parser/errors.go` — `ParseError` shape (Message/Line/Context/Cause), `wrapDecodeError`
- `internal/model/unit.go` — `Unit` struct (SubunitOrder contract, Clone preserving Mirror)
- `internal/model/link.go` — `Link` struct (unexported `Mirror`, `Arrow`/`Rank` enums)
- `internal/model/humanize.go` — `Humanize` rules (no acronym preservation)
- `internal/peer/resolve.go` — D-13/14/15/16 ancestor-walk, absolute-path gate
- `internal/include/merge.go` — `mergeSubunits`, `mergeProperties` (reflection), `mergeTemplates`
- `internal/template/expand.go` — fast-path no-op, `pathTracker` collision pattern, HS-1 Mirror preservation
- `cmd/c4drill/root.go` — `parser.ParseFile` dispatch at line 117, pipeline order, silent-on-success contract

---
*Pitfalls research for: LikeC4 DSL parser/converter integration into c4drill v1.11*
*Researched: 2026-08-08*
