---
phase: 35-add-a-simple-dsl-alternative-to-the-toml-diagram-definition-
reviewed: 2026-08-14T20:07:42Z
depth: standard
files_reviewed: 30
files_reviewed_list:
  - cmd/c4drill/convert.go
  - cmd/c4drill/convert_test.go
  - cmd/c4drill/fmt.go
  - cmd/c4drill/fmt_test.go
  - cmd/c4drill/root.go
  - cmd/c4drill/root_test.go
  - internal/c4d/ast/ast.go
  - internal/c4d/composition_test.go
  - internal/c4d/doc.go
  - internal/c4d/emit_c4d.go
  - internal/c4d/emit_test.go
  - internal/c4d/emit_toml.go
  - internal/c4d/errors.go
  - internal/c4d/frommodel.go
  - internal/c4d/grammar/c4d.peg
  - internal/c4d/grammar/doc.go
  - internal/c4d/grammar/parser_gen.go
  - internal/c4d/grammar/reserved.go
  - internal/c4d/parity_test.go
  - internal/c4d/parse.go
  - internal/c4d/parse_test.go
  - internal/c4d/tomodel.go
  - internal/c4d/tomodel_test.go
  - internal/include/resolve.go
  - internal/parser/parser.go
  - internal/template/expand.go
  - internal/testutil/canonsrc/canonsrc.go
  - internal/tomlfmt/tomlfmt.go
  - skill/SKILL.md
  - README.adoc
  - tools.go
findings:
  critical: 6
  warning: 7
  info: 3
  total: 16
status: issues_found
---

# Phase 35: Code Review Report

**Reviewed:** 2026-08-14T20:07:42Z
**Depth:** standard
**Files Reviewed:** 30 (+ supporting reads: internal/model/unit.go, internal/model/link.go, internal/include/merge.go, internal/render/render.go)
**Status:** issues_found

## Summary

The Phase 35 surface (C4D grammar, AST, Model-hub converters, convert/fmt
subcommands, mixed-format includes, canonsrc, tomlfmt) is unusually
well-tested on its happy paths: grammar error contracts, comment trivia,
idempotency sweeps and include-graph cycles are all pinned. The safety
engineering inside `fmt` (T-35-08-01 semantic gate) is genuinely good.

The adversarial finding, verified by executing the actual conversion paths:
**`convert` has no safety gate, and every emission boundary has at least one
class of legal input it silently corrupts or breaks.** Six verified defects
were reproduced end-to-end (five via `c4d.EmitC4D(c4d.FromModel(...))` /
`c4d.EmitTOML(...)` round-trips, one via the built CLI):

1. Link technology containing `|` is silently reshuffled between
   Technology and Description (CR-01).
2. A legal type-led `.c4d` unit with a display name converts to TOML with an
   unquoted table key — the twin does not parse (CR-02).
3. TOML unit ids outside `[A-Za-z0-9_-]+` (spaces, dots, unicode) produce
   unparseable C4D twins (CR-03).
4. `width`/`height` are silently dropped by `convert to-c4d` (CR-04).
5. Multi-line values ending in `"` emit an ambiguous triple-quote closer —
   the twin does not parse (CR-05).
6. `convert --follow-includes -o` flattens the graph's directory structure
   when the entry path is relative (CR-06).

`fmt` avoids most of this by construction (it re-derives the AST rather than
the Model and gates the rewrite); `convert` writes whatever the emitters
produce straight to disk. The single highest-leverage fix is a fmt-style
re-parse gate in `convert` (plus explicit loud errors for the values the C4D
format cannot represent, rather than silent mangling).

Test coverage gap: the corpus fixtures are all C4D-identifier-safe ASCII
without pipes in labels or quote-terminated multi-line strings, so the
parity/idempotency suites never exercise the corrupting paths. Tests also
only ever pass absolute `t.TempDir()` paths, which is why CR-06 survived
`TestConvertFollowIncludesOutputDir`.

## Structural Findings (fallow)

No structural pre-pass was provided for this review (no
`<structural_findings>` block). Cross-file observations made during review
are recorded under Narrative Findings instead.

## Narrative Findings (AI reviewer)

All Critical findings below were **reproduced by running the code** (scratch
programs against the module's exported entry points, and the built CLI for
CR-06); none are speculative.

### Critical Issues

### CR-01: Pipe inside link technology silently corrupts Technology/Description on `convert to-c4d`

**File:** `internal/c4d/emit_c4d.go:456-467` (with `internal/c4d/grammar/c4d.peg:159-165`)
**Issue:** `emitEdgeC4D` renders the label as
`quoteC4D(edge.Technology + " | " + edge.Description)`. Quoting does **not**
protect the pipe: on re-parse `splitPipeLabel` runs on the *unescaped* string
and splits on the **first** `|`. Verified round trip of
`technology = "HTTP | REST"`, `description = "calls"`:

```
-> b: "HTTP | REST | calls"      # emitted twin
original tech="HTTP | REST" desc="calls"
twin     tech="HTTP"      desc="REST | calls"   # silent corruption
```

The same defect fires for a description-only value containing a pipe
(`-> b: "queries | writes"` re-parses as tech=`queries`, desc=`writes`), and
`canonsrc.go:543-554` shares the exact logic (its comment even claims quoting
is pipe-safe — see WR-02). The C4D grammar has no escape or option form for a
label containing `|`, so the correct fix is to **fail loudly**, not to mangle.
**Fix:**

```go
// internal/c4d/emit_c4d.go — emitEdgeC4D, before building the label
if strings.Contains(edge.Technology, "|") ||
	(edge.Technology == "" && strings.Contains(edge.Description, "|")) {
	return nil, fmt.Errorf( // or a ParseError with edge.Pos
		"link label cannot be represented in C4D: %q / %q contains '|'",
		edge.Technology, edge.Description)
}
```

(Either that, or extend the grammar with `technology:`/`description:` edge
options so the pipe shorthand is not the only carrier.)

### CR-02: `EmitTOML` writes table headers with unquoted keys — legal `.c4d` input yields an unparseable TOML twin

**File:** `internal/c4d/emit_toml.go:115-116` (also `:242`, `:277`; trigger at `internal/c4d/tomodel.go:235-249`)
**Issue:** `emitUnitTOML` emits `"[" + path + "]"` raw. The C4D type-led form
`system "My App" { ... }` is legal, documented (README "Units and Nesting"),
and `unitKey` intentionally uses the display name as the unit key. Verified:
`c4drill convert to-toml` on such a file writes `[My App]` — the twin fails
with `expected ']' to close table name`. Dotted names would additionally
mis-nest (`["weird.key"]` → `[weird.key]` reads as two key segments).
**Fix:** quote key segments that are not bare TOML keys when building every
header (units, subunits, `template.<name>`, `[[...link]]`):

```go
func tomlKeySeg(seg string) string {
	if seg != "" && !strings.ContainsAny(seg, ". \t\"'=\n") {
		return seg
	}
	return `"` + strings.ReplaceAll(strings.ReplaceAll(seg, `\`, `\\`), `"`, `\"`) + `"`
}
// emitUnitTOML: path := joinDottedToml(prefix, tomlKeySeg(name))
```

### CR-03: `FromModel`/`EmitC4D` emit unit ids verbatim — TOML models with non-C4D identifiers produce unparseable twins

**File:** `internal/c4d/frommodel.go:104-121` + `internal/c4d/emit_c4d.go:384-412`; missing gate in `cmd/c4drill/convert.go:136-160`
**Issue:** TOML table names are arbitrary (`["my unit"]`, `["weird.key"]`,
unicode keys); C4D ids are `[A-Za-z0-9_-]+` (`c4d.peg:357`). `convert` passes
validation (the model is valid TOML) and then writes a twin that cannot
parse. Verified: `["my unit"]` → `my unit: system "My Unit" {` →
`no match found, expected: ":", ...`. The same applies to subunit ids, peer
strings (`emitEdgeC4D` writes `edge.Peer` raw) and template names. `fmt`
protects itself with the T-35-08-01 re-parse gate; `convert` has none.
**Fix:** either validate emitted identifiers against the grammar charset in
`FromModel`/`EmitC4D` and return a hard error, or (better, generalizes to
CR-01/CR-02/CR-05) add a fmt-style gate in `runConvert`/`convertGraph`:
re-parse the emitted text and require model-canonical equality before
writing the twin.

### CR-04: `width`/`height` silently dropped on `convert to-c4d`

**File:** `internal/c4d/frommodel.go:126-159` (`appendUnitBody`), `internal/c4d/emit_toml.go:130-170`
**Issue:** `model.Unit` carries `Width`/`Height` (`internal/model/unit.go:62-65`,
documented in README "Styling": `width = 300`), but `appendUnitBody` never
emits them and the C4D grammar rejects `width:`/`height:` as unknown field
keys. Verified: `width = 300, height = 200` → twin has `width=0 height=0`,
and the back-converted TOML has neither key. The round trip is
**lossy with no warning**, contradicting README's "Everything the TOML format
expresses — units, links, templates, includes, styling — has a C4D
equivalent" and convert's canonical-equivalence promise. (The decision is
acknowledged in `canonsrc.go:170-172`, which makes the parity tests
deliberately blind to it — that hides the loss from users, it does not fix
it.)
**Fix:** at minimum, `convert to-c4d` must warn or hard-error when a unit
carries non-zero Width/Height; better, add `width`/`height` to the C4D
`FieldKey` set and `unitStringField`/`appendUnitBody`/`EmitTOML` so the
fields round-trip.

### CR-05: Multi-line value ending in `"` emits an ambiguous triple-quote closer — twin does not parse

**File:** `internal/c4d/frommodel.go:344-357` (`literalFor`); same bug in `internal/testutil/canonsrc/canonsrc.go:692-698`
**Issue:** `literalFor` selects `KindTriple` for any newline-containing value
that does not contain `"""`. A value like `"line1\nline2\""` emits

```
description: """line1
line2""""
```

The grammar's triple rule (`c4d.peg:731-733`) closes at the **first** `"""`,
leaving a stray `""` — verified parse error at line 3. `convert to-c4d` on a
TOML model with such a description writes a broken twin.
**Fix:** also fall back to `KindQuoted` when the value ends with `"` (or
contains `""`):

```go
case strings.ContainsAny(s, "\n\r"):
	if !strings.Contains(s, `"""`) && !strings.HasSuffix(s, `"`) && !strings.Contains(s, `""`) {
		return ast.Literal{Kind: ast.KindTriple, Str: s}
	}
	return ast.Literal{Kind: ast.KindQuoted, Str: s}
```

Apply the identical guard in `canonicalC4DValue`.

### CR-06: `convert --follow-includes -o` flattens the graph's directory structure for relative entry paths

**File:** `cmd/c4drill/convert.go:233` + `cmd/c4drill/convert.go:391-404` (`graphTwinPath`)
**Issue:** `walkIncludeGraph` absolutizes paths (`absEntry`), but
`convertGraph` computes `entryDir := filepath.Dir(entryPath)` from the raw CLI
arg. When the arg is relative (`entry.toml` → entryDir `.`),
`filepath.Rel(".", "/abs/.../domains")` errors, the `err == nil` guard fails,
and every twin lands **flat** in `-o`. Verified with the built CLI:
`c4drill convert to-c4d --follow-includes -o out entry.toml` yields
`out/auth.c4d` instead of `out/domains/auth.c4d` (absolute entry path yields
the documented layout). README promises "graph mode preserves the graph's
relative directory structure". `TestConvertFollowIncludesOutputDir` only ever
passes absolute `t.TempDir()` paths, so the suite cannot catch this.
**Fix:**

```go
func convertGraph(entryPath, targetExt string) error {
	absEntry, err := filepath.Abs(entryPath)
	if err != nil { ... }
	entryDir := filepath.Dir(absEntry)   // was: filepath.Dir(entryPath)
	...
}
```

## Warnings

### WR-01: canonsrc sorts positional `use` args — order is semantic

**File:** `internal/testutil/canonsrc/canonsrc.go:641-671` (`writeUseLineC4D`)
**Issue:** Args are sorted by `(name, value)`. Positional args (empty `Name`)
carry meaning through their **position** relative to `TemplateDecl.Params`;
sorting re-pairs them with different params. `use t(x, y)` and `use t(y, x)`
canonicalize to the same text — the parity harness can produce false passes
(and a genuine order difference is unobservable).
**Fix:** partition args: named args may sort by key; positional args must
stay in authored order (concatenate: positional prefix in order, then sorted
named).

### WR-02: canonsrc's pipe-safety comment is false and its normalization shares CR-01

**File:** `internal/testutil/canonsrc/canonsrc.go:540-554` (`edgeLabelCanonC4D`)
**Issue:** The comment says "the quoted form keeps embedded pipes round-trip
safe (splitPipeLabel splits on the FIRST pipe only)". Quoting does not
protect anything — the split happens after unquoting (verified in CR-01). As
the canonical-equivalence oracle, this normalizer masks exactly the
corruption CR-01 describes, so the parity suite cannot detect it.
**Fix:** delete the false comment and apply the same loud-failure rule as
CR-01's fix; a tech containing `|` has no canonical form.

### WR-03: Duplicate `properties { }` blocks silently last-win in C4D; the TOML twin is a parse error

**File:** `internal/c4d/grammar/c4d.peg:223-234` (Document action `doc.Properties = n`)
**Issue:** Verified: two `properties` blocks parse cleanly and the second
silently replaces the first (`Properties.Name` = "Second"). TOML rejects a
duplicated `[properties]` table outright — a parity gap with silent field
loss (all fields of the first block vanish).
**Fix:** track whether a PropertiesBlock was already seen in the Document
action and return a `*parser.ParseError` ("duplicate properties block"), or
reject in `ToModel.applyProperties`.

### WR-04: SKILL.md type-inference table contradicts `DefaultTypeForParent`

**File:** `skill/SKILL.md:64-68`
**Issue:** The row `| C2 | system, systemExternal, box | container |` is
wrong on two of three parents: `DefaultTypeForParent`
(`internal/parser/parser.go:954-973`) returns `system` for `systemExternal`
(default/C1 fallback branch) and for `box` (C1 same-level grouping) — not
`container`. README's table (`README.adoc:387-397`) is correct. A skill user
relying on the table authors a child under `systemExternal` expecting C2
`container` and gets a C1 `system` instead (then fails nesting validation).
**Fix:** `| C2 | system | container |` and drop `systemExternal`/`box` from
the row (they belong to the C1 rows, as in README).

### WR-05: Link-label edge whitespace is silently trimmed by TOML→C4D conversion

**File:** `internal/c4d/grammar/c4d.peg:159-165` (`splitPipeLabel`) + `internal/c4d/emit_c4d.go:456-467`
**Issue:** `splitPipeLabel` applies `strings.TrimSpace` to both halves, and
the desc-only path also trims — even for quoted labels, whose documented
contract is "whitespace preserved verbatim" (`ast.go:16-18`). Verified:
TOML `description = " padded desc "` → twin re-parses as `"padded desc"`.
Quoting was explicitly introduced to preserve whitespace; the label grammar
path violates it.
**Fix:** in `EdgeLabel`, skip `TrimSpace` for the quoted/triple forms (keep
trimming only for barewords); or have `emitEdgeC4D` reject values whose edge
whitespace would not survive.

### WR-06: Include merge silently drops an included parent's scalar fields and links

**File:** `internal/include/merge.go:61-77` (adjacent to the listed `resolve.go`, same package; pre-existing Phase 32 behavior)
**Issue:** When an included file re-declares a parent that already exists and
contributes subunits, `mergeUnits` hands off to `mergeSubunits` and then
`continue`s — the src unit's `type`, `name`, `description`, and **its entire
`Links`/`LinksFrom`** are silently discarded. The inline comment even claims
redeclaring non-subunit fields "is a cross-file collision (D-11)" — the code
never raises it. An `[[api.link]]` contributed by an included fragment under
a shared parent disappears without a trace (reachable through `convert`'s
gate and the render pipeline alike).
**Fix:** after `mergeSubunits` succeeds, diff src's scalar fields and links
against dst's and either error (D-11 as documented) or merge the links.

### WR-07: C4D cannot express an id-led unit header with a name but no type

**File:** `internal/c4d/grammar/c4d.peg:315-321` (`UnitHeader`); doc claim `README.adoc:770-784`
**Issue:** README documents `id`, `type` and `"Display Name"` as
independently optional, but the colon form *requires* a `TypeKeyword`
(`x "Name" { }` and `x: "Name" { }` both fail to parse). Only the type-led
form can carry a name without a type. Convert paths never emit that shape
(FromModel always materializes the inferred type), so this is an authoring
expressiveness/doc mismatch rather than a corruption.
**Fix:** either allow `id: _ QuotedString` as a fourth `UnitHeader`
alternative, or document that a name without a type requires the type-led
form.

## Info

### IN-01: `collectExpandedPaths` iterates the `Units` map — nondeterministic processing order

**File:** `cmd/c4drill/root.go:254`
**Issue:** Map iteration order randomizes the order in which sub-diagrams are
generated (and which view's error surfaces first when several fail). Output
content is unaffected (each path writes its own file). Pre-existing v1.10
code (`git blame 2f21325`), only touched by Phase 35 for extension dispatch.
**Fix:** iterate `m.UnitOrder` (and `SubunitOrder`) instead of the map.

### IN-02: canonsrc `writeUnitCanonC4D` recurses at hardcoded depth 1

**File:** `internal/testutil/canonsrc/canonsrc.go:426-455`
**Issue:** Nested units render their inner statements with fixed `"  "`
indent and recurse with `depth = 1` regardless of true nesting, while the
unit's own indent uses `depth`. The canonical text is misindented (still
valid C4D, so the fixpoint property holds). Test-only cosmetic defect.
**Fix:** thread `depth+1` through the recursion and derive the inner indent
from it.

### IN-03: Template-body conversion errors report the unit path as `""`

**File:** `internal/c4d/tomodel.go:719-732` (`templateDefFromAST` → `buildUnit(decl.Name, decl.Body, "", "")`)
**Issue:** Duplicate-field/duplicate-subunit errors inside a template body
render as `duplicate field "x" in unit ""` — the template name is available
but unused. Cosmetic diagnostics only.
**Fix:** pass a synthetic path such as `"template." + decl.Name` for error
context (the Model fields themselves must stay as-is for parity).

---

_Reviewed: 2026-08-14T20:07:42Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
