# 35-04 Deferred Items

## C4D template root type is not expressible in the grammar (found during 35-04 Task 2)

The 35-03 grammar's template bodies reuse the unit-body FieldKey set, which
has no `type` key (and `type:` errors as an unknown field). A TOML
`[template.X]` with `type = "container"` therefore cannot carry its root type
into C4D text: `FromModel` records `TemplateDecl.Body.Type/External` on the
AST, but `EmitC4D` does not render them (the output would not re-parse).

Impact: TOML -> C4D -> TOML round-trips lose a non-default template root
type. The 35-06 round-trip fixtures use `template_basic.toml` whose root type
is `system` (the root default), so D-22's "explicit defaults may normalize
away" covers the shipped fixtures.

Owner: 35-05 (tomodel owns the AST->Model inverse and may extend the grammar)
or 35-06 (round-trip contract). Options: add `type` to FieldKey for template
bodies only, or a template-header type slot.

## C4D edge labels cannot carry literal pipes (grammar, pre-existing)

`splitPipeLabel` splits an edge label on the FIRST pipe (D-09), so a link
description or technology containing `|` does not survive C4D text: emitting
`"a | b"` re-parses as tech `a`, description `b`. The emitters quote such
values correctly, but the value changes meaning. Pipe-free descriptions
round-trip exactly; all fixtures are pipe-free.

Owner: 35-06 (round-trip normalizer) if it matters; otherwise document as a
C4D authoring constraint.

## Update (35-05, 2026-08-14): Model-level half CLOSED

ToModel now honors `TemplateDecl.Body.Type/External` when set (Task 2),
so `ToModel(FromModel(m))` preserves a non-default template root type —
pinned by TestToModelTemplateRootTypeFromModel. The TEXT-level half
remains: the grammar still cannot parse a template root type, so
TOML -> C4D(text) -> TOML round-trips still lose it unless D-22's
explicit-defaults normalization covers the fixture. Remaining owner: 35-06
(round-trip contract) or a grammar extension in a later plan.

## Update (35-06, 2026-08-14): CLOSED (text level too)

The grammar now admits a root type statement in template bodies only —
`template X(p) { type: container external ... }` (TemplateBodyStmt adds
TemplateTypeStmt ahead of the shared BodyStmt forms; unit bodies still
reject `type:` as an unknown field, and a duplicate type statement is a
hard parse error). EmitC4D renders Body.Type/External as that statement,
so TOML -> C4D -> TOML round-trips preserve non-default template root
types — enforced by the 35-06 corpus round-trip suite (06-templates.toml,
template_reserved.toml carry `type = "container"` roots) and pinned by
TestToModelTemplateTypeStatement. The same plan extended PeerRef with
${param} segments so parametrized template link peers (XC-03) round-trip.

The pipe-in-edge-label constraint below remains an authoring constraint
(canonical C4D quoting keeps embedded pipes round-trip safe in the
tech+desc label; a desc-ONLY value containing a pipe cannot survive the
D-09 first-pipe split — all corpus fixtures are pipe-free).

