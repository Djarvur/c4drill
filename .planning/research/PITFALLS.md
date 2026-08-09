# Domain Pitfalls — v1.10 Model Composition

**Domain:** C4 Diagram Generation (Go CLI) — v1.10 "Model Composition"
**Researched:** 2026-08-08
**Scope:** Pitfalls for the milestone pipeline `include → template-expand → relative-peer-resolve → validate → generate-views → render`, where backward compatibility for existing single-file models is non-negotiable.
**Confidence:** HIGH for code-cited items (verified against code at time of writing); MEDIUM for go-toml/Go-ecosystem behavior (verified against docs/spec, not all empirically run).

> Supersedes the v1.1 "AI-Ready" pitfalls doc (all-expanded + AI docs shipped in v1.7–v1.9). The two High-Severity items below are the ones most likely to cause real, hard-to-debug bugs.

---

## High-Severity Pitfalls

These two are flagged as high-severity because (a) they cause *silent* corruption — the model parses and renders without error but is wrong, and (b) the failure surfaces far from its cause, making it very expensive to debug after the fact.

### HS-1: Deep-copy aliasing corrupts the template on the Nth instantiation

**What goes wrong.** Under templates Option B (structured post-parse expansion — recommended in `2026-08-08-unit-templates-parametrized-definitions.md`), the expansion pass deep-copies a template `Unit` and substitutes params. If the copy is *shallow*, the copy shares the template's underlying slices and maps. Mutating the instantiation then mutates the template, so the *second* instantiation starts from an already-corrupted template. Worse, the validator itself mutates units in place (see below), so even read-looking passes are dangerous.

**Why this codebase is specifically exposed.**

1. The `Unit` struct (`internal/model/unit.go:41-72`) has three reference-typed fields that alias under a shallow copy:
   - `Subunits map[string]*Unit` (`unit.go:71`, `toml:",inline"`) — a map of *pointers*;
   - `Links []Link` (`unit.go:65`) and `LinksFrom []Link` (`unit.go:67`) — slices (slice header copied, backing array shared);
   - `Expanded []string` (`unit.go:63`) — slice.
   A plain `*target = *template` (or `copier.Copy`) copies these by value-of-pointer/slice-header: `target.Subunits` and `template.Subunits` point at the same map; `target.Links` and `template.Links` share a backing array.

2. There is **no existing deep-copy helper anywhere in the codebase** (grep for `deepCopy`/`clone`/`Copy` over `internal/` returns nothing). The template feature would introduce the first one.

3. The validator **mutates units in place** during a normal run: `populateIncomingLinks` (`internal/validator/index.go:53-84`) does `targetInfo.Unit.LinksFrom = append(targetInfo.Unit.LinksFrom, model.Link{... Mirror: true})` (`index.go:70-81`). Because `BuildIndex` (`index.go:23-46`) stores *pointers* into the same `Model.Units` graph, an aliased `LinksFrom` means appending a mirror link to instantiation #1 visibly grows the backing array that instantiation #2 (and the template) also see. With Go slice semantics, if the shared backing array has spare capacity, the append *mutates in place without reallocating* — instantiation #2 silently inherits instantiation #1's mirror links.

4. The graph builder reads these same shared structures directly: `entry.Unit.Subunits[childName]` (`internal/graph/builder.go:220,312`), `entry.Unit.Links` / `entry.Unit.LinksFrom` (`builder.go:360-367,410-434`), `parent.Expanded` (`builder.go:342`). So corruption propagates into rendered output.

5. `Link.Mirror` (`internal/model/link.go:67`, `toml:"-"`) is an **unexported-capability flag** set only by the validator. `Mirror` is `false` on an authored link. A naive deep-copy that reconstructs `Link` literals field-by-field (instead of copying the struct) would drop nothing critical here because `Mirror` survives struct copy — but an encoder/round-trip-based copy (marshal→unmarshal to dodge the deep-copy work) would **silently reset `Mirror=false` on every link**, re-breaking multiplicity counting D-05 (`STATE.md`). Do not round-trip-copy Links.

**Concrete failure scenario.**

```toml
[template.svc]
params = { name = "x" }
name = "${name} Service"
type = "container"
[[template.svc.link]]
peer = "bus"
description = "${name} publishes"

[[use]]
template = "svc"
name = "auth"        # instantiation #1 -> top-level path "auth"

[[use]]
template = "svc"
name = "billing"     # instantiation #2 -> top-level path "billing"
```

Shallow copy: both `auth` and `billing` share `template.svc.Links`'s backing array `[{peer:"bus", ...}]` (len 1, cap ≥1). After expansion, validator runs `populateIncomingLinks`: it appends a `Mirror=true` incoming link onto `bus.LinksFrom` (fine — `bus` is its own unit), but it also walks `auth.Links` and `billing.Links`. The aliasing bug shows up differently depending on what the expansion pass wrote: if the pass substituted `${name}` *into the shared array* (writing `auth` into element 0 of the shared backing array), then `billing`'s link description reads "auth publishes" — wrong. If the pass did `links = append([], template.Links...)` per instantiation (good) but copied `Subunits` shallowly, two instantiations of a template that *contains subunits* share the subunit map, and substituting a param into `auth.child.Name` mutates the entry `billing.child` reads.

The symptom: the first instantiation renders correctly; the Nth renders with the first's name/links/description, or with phantom mirror links, or — if the backing array reallocates differently per append — symptoms that change between runs depending on slice growth, i.e. **non-deterministic output that is not the go-graphviz kind already documented**.

**Prevention (the right deep-copy for THIS struct).**

- Write a hand-rolled recursive `deepCopyUnit(*model.Unit) *model.Unit` in the `model` (or a new `model/clone` / parser-internal) package. Do not use reflection-based copiers without an explicit test that covers `Subunits` (nested, pointer-valued) and `Expanded`.
- The recursion must: allocate a fresh `Subunits` map and deep-copy each `*Unit` value; allocate fresh `Links` and `LinksFrom` slices by copying each `Link` **struct-by-struct** (not pointer-to-array) so the `Mirror` field survives; allocate a fresh `Expanded` slice.
- Because `Unit` has no unexported fields (all fields are exported, `Mirror` lives on `Link`), a same-package copy is straightforward. Keep it in `package model` so it can touch fields without accessors.
- **Critical ordering rule:** deep-copy the template *once per instantiation, then substitute, then discard the copy*. Never mutate the template unit itself. The template registry entry should be treated as immutable after parse.
- Add a regression test: instantiate the same template 3× with distinct params; assert (a) each instantiation's `Links[0].Description` differs, (b) re-running expansion a second time (idempotency) yields identical output — this catches the "second run reads a corrupted template" class.
- The `Mirror` field specifically: expansion happens *before* validation, so template links start with `Mirror=false`. After `populateIncomingLinks` runs (post-expansion), mirror links are appended only to the *instantiations' own* `LinksFrom` — safe **only if** each instantiation owns its `LinksFrom` slice. The deep-copy test must verify `auth.LinksFrom` and `billing.LinksFrom` are disjoint slices after a full validate pass.

**Warning signs during development.**
- Output changes when you add a throwaway first instantiation.
- Mirror links (`Mirror=true`) appear on authored linkFrom pairs in multiplicity counting.
- Template expansion "works for one, breaks for many."
- `go test` passes individually but a randomly-ordered test run (`-shuffle=on`) fails.

**Phase to address:** the Templates implementation phase — this is THE core correctness concern of that feature; do not land template expansion without the deep-copy helper and the 3-instantiation regression test.

---

### HS-2: Relative-peer resolution inside a template — template-parent vs instantiation-parent ambiguity

**What goes wrong.** A relative peer is a short name that resolves against the *structural parent* of the unit that authored the link (sibling lookup). When the link is authored *inside a template*, "structural parent" is ambiguous: is it the template's lexical parent (where `[template.x]` is defined, possibly none/top-level), or the instantiation site's parent (where `[[use]]` places the concrete unit)? These give different siblings, different resolved peers, and in a model with sibling name collisions, **silently wrong edges**.

**Why this codebase is specifically exposed.**

The current peer-resolution model is *absolute*: `Link.Peer` (`internal/model/link.go:44`) is a full dotted path, and every resolution site (`resolveToViewAncestor` / `resolveToTopLevel` in `internal/view/scope.go:367-378`, `addResolvedBoundaryNode` at `scope.go:723-757`, the validator's `index[link.Peer]` lookup at `internal/validator/rules.go:23,34`) treats it as an exact key into the path index built by `BuildIndex` (`validator/index.go:23-46`). There is **no relative-peer resolution today** — `peer` is always absolute. So the relative-peer feature introduces a brand-new resolution step, and its interaction with templates is undefined by current code.

The pipeline ordering in the include todo (`2026-08-08-include-directive-multi-file-diagrams.md:87-96`) explicitly says relative-peer resolution runs **after** template expansion: `include → template-expand → relative-peer-resolve → validate`. So by the time relative-peer resolution runs, the template has been instantiated and the link sits on a concrete unit at a concrete path. The ambiguity is: when the resolver sees `link.Peer = "cache"` on a unit whose path is `mainSystem.auth`, does it look for `mainSystem.cache` (instantiation-site sibling) or for whatever sibling the *template* implied (e.g. a `cache` defined next to `[template.svc]`)?

**Concrete failure scenario.**

```toml
# templates.toml — defines template with a RELATIVE peer
[template.svc]
params = { name = "x" }
name = "${name}"
type = "container"
[[template.svc.link]]
peer = "cache"              # RELATIVE — author meant "a sibling named cache"
description = "uses cache"
```

```toml
# model.toml
[mainSystem]
type = "system"
[[mainSystem.subunit]]      # or inline — mainSystem contains auth, billing, cache
...
[mainSystem.auth]           # (via [[use]] template=svc name=auth)
[mainSystem.billing]        # (via [[use]] template=svc name=billing)
[mainSystem.cache]
type = "containerDb"

[otherSystem]
type = "system"
[otherSystem.cache]         # COLLISION — different cache, different sibling set
type = "containerDb"

[[use]]
template = "svc"
name = "billing"
parent = "otherSystem"      # instantiation placed inside otherSystem
```

- If resolution is **instantiation-site-relative** (recommended): `billing`'s `peer="cache"` resolves to `otherSystem.cache` — almost certainly what the author wants. `auth`'s resolves to `mainSystem.cache`.
- If resolution is **template-site-relative**: there is no template parent (templates are top-level data), so `"cache"` either fails to resolve (validator error) or resolves globally/ambiguously — and `billing` and `auth` resolve to the *same* `cache`, which is wrong for one of them.

The silent-wrong-edges variant is worse than the error variant: if a global "first `cache` found" fallback exists, both links point at `mainSystem.cache`; the diagram renders, the validator passes, and `billing → mainSystem.cache` is a fabricated cross-system edge.

A second, subtler failure: even with instantiation-site resolution, if relative resolution is *nearest-ancestor-first* and the instantiation is nested (e.g. `[[use]] parent="mainSystem.auth"` placing a component inside a container), the resolver must walk up from the instantiation's actual path, not from the template's lexical position. Getting "walk up from the concrete unit's parent" right requires the resolver to see the post-expansion tree, which is why the pipeline ordering pins relative-peer-resolve after expansion.

**Prevention.**

- **Decide explicitly (this is a Design Fork — see Decision Required):** relative peers authored in a template resolve against the **instantiation site's** structural parent, never the template's lexical location. Document this as the contract. Rationale: templates are data, not structural location; the whole point of instantiation is to place a unit somewhere, and "sibling" only has meaning relative to that somewhere.
- **Fallback order for safety:** relative-first (search the unit's own parent's children), then absolute-fallback (if `peer` contains a `.` OR exactly matches a top-level path, treat as absolute). This preserves backward compat: every existing model uses absolute peers, and they all either contain `.` or match a top-level key, so they skip the relative search and resolve exactly as today.
- **Ambiguity within a parent is a hard error**, not first-match: if `mainSystem` had two children literally named `cache` (impossible by map key, but possible across `link`/`linkFrom` resolution scopes) — more realistically, if two siblings share a short name in different branches and a relative peer is ambiguous at the chosen level, emit a validator error naming both candidates rather than silently picking nearest. The "nearest wins" rule is for *nesting depth*, not for *ties at the same depth*.
- Implement relative resolution as a **separate pass that rewrites `Link.Peer` in place on the (deep-copied, post-expansion) model** before `BuildIndex`/validation, so the validator's existing absolute-path logic is untouched. This keeps the validator the single gatekeeper (consistent with STATE.md D-12).
- Test: template with relative peer, instantiated under two different parents; assert each instantiation's link resolves to its own sibling, not a shared global. Test: existing absolute-peer model produces byte-identical `Link.Peer` set after the relative pass (backward-compat guard).

**Warning signs during development.**
- Two instantiations of the same template both link to the same target despite living in different parents.
- A template link "can't find" a peer that clearly exists as a sibling of the instantiation.
- Existing models start emitting "ambiguous peer" errors (means the relative search is firing on absolute peers — fix the absolute-fallback gate).

**Phase to address:** Relative-peer implementation phase AND the discuss phase (the resolution-site decision must be settled before implementation, see Decision Required). The template-relative-peer interaction test belongs in the Templates phase but depends on relative-peer existing.

---

## Per-Feature Pitfalls

### (1) INCLUDE pitfalls

#### IN-1: go-toml/v2 silently merges duplicate tables — but only the right kinds
**Pitfall.** The include todo (`2026-08-08-include-directive-multi-file-diagrams.md:41`) warns that "TOML does not naively concatenate." The concrete rules (verified against the TOML v1.0.0 spec):
- Two files each with `[properties]` → **hard TOML error** ("you cannot define a table more than once") if byte-concatenated. This is why Option A (pre-parse byte concat) is fragile.
- Dotted-path sub-tables `[a.b]` in two files merge correctly (TOML allows defining `[a]` then `[a.b]`), *provided* `[a]` wasn't already created via dotted keys inside an `[a]` block.
- go-toml/v2 `Unmarshal` is **non-strict by default** (it does NOT call `DisallowUnknownFields()`). C4Drill uses bare `toml.Unmarshal` everywhere (`internal/parser/parser.go:57,69,74,183`). Unknown keys are silently ignored at the struct level — which is *why* the `Subunits` trick (`toml:",inline"`, `unit.go:71`) works. This is load-bearing: turning strict mode on globally would break subunit parsing.

**Warning sign.** An include implementation that byte-concatenates and then "it works for my test case" — it will work right up until two files both define `[properties]` or both define `[a]`, then produce a confusing go-toml `DecodeError` whose line number points into the concatenated blob (not the original file), defeating `wrapDecodeError`'s line extraction (`errors.go:78-82`).

**Prevention.** Use Option B (parse-then-merge structs) as the todo recommends. Merge `parser.Model` structs in Go, where you control the conflict rules per-field. Never byte-concatenate. Preserve file-of-origin for error messages (the merged `Model`/unit should carry a source-file annotation, or the merge function should wrap conflicts with the originating filename).

**Phase:** Include implementation phase — the merge strategy IS the feature.

#### IN-2: UnitOrder semantics across merged files affect rendering
**Pitfall.** `UnitOrder` (`parser.go:39`) is "definition order" captured by `captureDefinitionOrder` (`parser.go:100-157`), and it is **load-bearing for rendering** — it drives the order units appear in views (e.g. `GenerateC2View` iterates `systemUnit.SubunitOrder`, `scope.go:408-432`), which affects GraphViz layout determinism (the pinned go-graphviz fork uses map iteration order internally, but node *insertion* order still biases layout per STATE.md's order-nondeterminism notes).

If file1 defines `[a][b]` and file2 (included) defines `[c][d]`, the merged `UnitOrder` could be `[a,b,c,d]` (append-in-include-order, the todo's recommendation at line 52) or interleaved at the include-site position. For C4Drill, **append-in-include-order is correct and least-surprising**: the include directive is a single logical line in the root file; its position within the root's unit list is not meaningful (the root's units are defined by their own `[unit]` headers, not by where the `[include]` table sits). But `SubunitOrder` (per-unit, `unit.go:69`) is a different question: if an included file adds subunits to a unit defined in the root file, do they append to that unit's `SubunitOrder`? That requires a per-unit merge, and the order is meaningful at C2/C3 (sibling layout).

**Warning sign.** Golden SVGs for multi-file models differ from the equivalent single-file model beyond the already-documented go-graphviz sibling-order nondeterminism (STATE.md: blockers DI-1). I.e., a *semantic* ordering difference on top of the cosmetic one.

**Prevention.** Pin the rule: top-level `UnitOrder` = root file's units in definition order, then each included file's units appended in include directive order. Per-unit `SubunitOrder`: included files may *append* subunits to a root-defined unit (merge), with included subunits after root-authored subunits. Reject same-named subunit across files (same hard-error rule as top-level units). Test the multi-file vs single-file equivalence with the **order-insensitive canonicalDOT** comparator (STATE.md decision log), NOT byte-exact `require.Equal`.

**Phase:** Include implementation + its golden tests.

#### IN-3: Cycle detection edge cases — diamond must NOT be a cycle
**Pitfall.** The todo (line 100-103) specifies an include stack for cycle detection, citing go-metadot's `@incChain`. The edge cases:
- **Self-include** (A includes A): trivially a cycle → fatal error. Easy.
- **Mutual** (A↔B): A includes B, B includes A → cycle → fatal. Easy with a stack.
- **Diamond** (A→B, A→C, B→D, C→D): D is included by two paths but neither path is ancestral to the other. **This is NOT a cycle** — but a naive "already seen" set (as opposed to "on current stack") would treat D's second inclusion as a cycle and either error or skip. The right data structure is a **stack** (path from root to current file), not a global seen-set. A file may legitimately be included twice via non-ancestral paths *if* the user wants its definitions twice (rare, usually an error) — or once if `once=true`.

**Warning sign.** A diamond include graph fails with "cycle detected" when there is no cycle; or conversely, `once=true` dedup silently drops D's second inclusion and the user can't tell why a definition vanished.

**Prevention.** Use a **stack** (list of absolute paths in the current include chain) for cycle detection — re-entering a file on the stack = fatal. Separately, implement `once=true` (include_once) via a **global visited-set** that is *opt-in per directive* (PlantUML `!include_once` semantics, todo line 101). Document the distinction: stack = "is this an infinite loop," visited-set = "has this file's content already been spliced." The diamond with no `once` flags includes D twice (possibly producing duplicate-definition errors downstream, which is the correct signal to the author to add `once=true`). Max depth cap (todo line 103, e.g. 100) catches pathological graphs that are technically acyclic but huge.

**Phase:** Include implementation phase.

#### IN-4: Nested includes and resolution context
**Pitfall.** An included file has its own `[include]` directive. Path resolution for the nested include must be **relative to the included file's directory**, not the root's (go-metadot precedent, todo line 27). But the *include stack* and *visited-set* must be global across the whole tree, not per-file.

**Warning sign.** A template library in `templates/lib/common.toml` includes `templates/lib/helpers.toml` — if nested includes resolve relative to the root file's cwd, the path breaks when the root is in a different directory.

**Prevention.** Pass the including file's absolute directory into the recursive parse call. Canonicalize all paths to absolute before pushing onto the stack/set (so `./x.toml` and `x.toml` and `./a/../x.toml` don't dodge cycle detection via path-string differences). Use `filepath.Abs` + `filepath.Clean`.

**Phase:** Include implementation phase.

#### IN-5: Path resolution surprise — relative-to-file vs relative-to-cwd
**Pitfall.** Related to IN-4 but user-facing: users invoke `c4drill model.toml` from various cwds. If includes resolved relative to cwd, the same model file would work or break depending on where the user stands. Relative-to-including-file is the only sane default (both reference tools do this).

**Warning sign.** "Works on my machine, breaks in CI" — classic cwd-relative path bug.

**Prevention.** Resolve every include path relative to the directory of the file that contains the directive. Support absolute paths as an escape hatch. Document prominently. Add a test that `cd`s elsewhere and runs the CLI.

**Phase:** Include implementation phase.

---

### (2) TEMPLATE pitfalls

#### TM-1: Duplicate unit path from parameter collision is a silent shadow
**Pitfall.** Two instantiations with params that produce the same unit path (e.g. two `[[use]] template=svc name=auth`, or a `[[use]] name=auth` whose path collides with a hand-authored `[auth]`) — the merge into `Model.Units` (`parser.go:92` does `m.Units[name] = unit`) silently overwrites. The overwritten unit's links/subunits vanish, but its path may still be referenced by other links → broken-peer errors that point at the *surviving* unit, confusing the author.

**Warning sign.** A model with N instantiations renders N-1 units; or a "undefined unit" error for a path the author definitely wrote.

**Prevention.** The expansion pass must **error on duplicate unit path** before insertion. Because the instantiation key (e.g. `[[use]] name=auth` with optional `parent`) determines the path, validate uniqueness across both `[[use]]` blocks AND against hand-authored `[unit]` tables at expansion time, with an error naming both definitions (file:line if available, else the param values that collided). This is the templates analog of IN-1's "same path in two files = hard error."

**Phase:** Templates implementation phase.

#### TM-2: Forward references to templates used before definition
**Pitfall.** The todo (line 131) says "templates must be defined before use (consistent with file-order semantics)." go-metadot fails forward references (todo line 30). Under Option B, go-toml parses the whole file into structs first, so `[[use]] template="svc"` *syntactically* resolves even if `[template.svc]` appears later in the file — the question is whether the expansion pass should *honor* that or *reject* it.

**Warning sign.** A user moves their `[template]` block below their `[[use]]` blocks for readability; depending on implementation it either works (because structs are fully parsed before expansion) or fails with a confusing "template not found."

**Prevention.** Two acceptable answers, but **pick one and document it**: (a) expansion is order-independent within a file (since parse completes before expansion) — allow forward references; or (b) require definition-before-use to match go-metadot/file-order intuition. Recommendation: **(a) order-independent** — it costs nothing under Option B (the template registry is built from the fully-parsed Model before any `[[use]]` is processed), is friendlier to authors, and avoids a fake restriction. The "before use" intuition comes from textual-preprocessor references (Option A / go-metadot) and does not apply to structured post-parse expansion. Note this as a decision so it is not "implemented" as a restriction by accident.

**Phase:** Templates discuss phase (it is a documented semantics choice) → implementation honors the choice.

#### TM-3: Recursion / expansion explosion
**Pitfall.** Template A instantiates template B which instantiates A. The todo (line 130) says nested templates are disallowed, which makes recursion impossible — *if the ban is enforced*. Without enforcement, a self-referential template causes unbounded recursion / stack overflow at expansion time.

**Warning sign.** Stack overflow / OOM during parse of a model with templates; or a hang.

**Prevention.** Even with nesting disallowed, add an expansion-depth/visit cap (e.g. 100, matching go-metadot's recursion cap, todo line 132) as defense-in-depth. Explicitly reject `[[use]] template=X` appearing inside `[template.X]` (or any template cycle) with a clear "template recursion not allowed" error. The cap catches the case where the ban has a bug.

**Phase:** Templates implementation phase.

#### TM-4: Links inside templates referencing peers — relative resolution MUST run after expansion
**Pitfall.** A template link's `Peer` may itself be a `${param}` substitution (todo line 139: "substitution must also apply to `Link.Peer`") OR a relative name. If relative-peer resolution ran *before* template expansion, it would try to resolve the pre-substitution peer against the wrong tree (the template skeleton, which is not in the real model). The pipeline ordering (include todo line 92-93) already pins relative-peer-resolve after template-expand; the pitfall is *implementing it in the wrong order by accident* (e.g. reusing the existing validator's peer-existence check too early).

**Warning sign.** Template links fail to resolve; or worse, resolve against a leftover skeleton path.

**Prevention.** The relative-peer-resolution pass must be a distinct stage that runs strictly after expansion and strictly before `validator.Validate` (which does the final absolute-path `index[link.Peer]` existence check at `rules.go:23,34`). The validator stays the gatekeeper (STATE.md D-12). Add a pipeline-order assertion test: a template with a relative peer, expanded and resolved, yields the same `Model.Units[path].Links[].Peer` as the hand-expanded hand-resolved equivalent.

**Phase:** Templates + relative-peer integration; cross-feature test.

#### TM-5: Parameter substitution leaving unresolved `${name}` literals
**Pitfall.** go-metadot leaves unresolved `${N}` literally in place (todo line 27). Under Option B, an unresolved `${name}` (typo in param name, or missing required param) silently becomes a literal string in a field — e.g. a unit named "${anem} Service" because `anem` was misspelled. The model parses, validates, and renders with a broken name.

**Warning sign.** Rendered names/descriptions containing literal `${...}`; or "works" but with weird text.

**Prevention.** Two rules: (a) a required param (no default) that is not supplied at instantiation = **hard error** naming the missing param (todo line 125 already specifies this); (b) after substitution, scan string fields for any remaining `${...}` pattern and **error** — any leftover means either a typo or a param referenced in the template that wasn't declared in `params`. Do not silently leave literals. (PlantUML leaves literals; go-metadot leaves literals; C4Drill should NOT, because structured post-parse gives us the affordance to be strict.)

**Phase:** Templates implementation phase.

---

### (3) RELATIVE-PEER pitfalls

#### RP-1: Nearest-ancestor resolution surprise with cross-branch short-name collisions
**Pitfall.** Relative-first picks the nearest sibling. If two siblings in *different branches* share a short name (e.g. `mainSystem.cache` and `otherSystem.cache`), a relative `peer="cache"` from `mainSystem.auth` resolves to `mainSystem.cache` (nearest) — which is correct. But if `mainSystem.auth` had *no* sibling `cache`, a naive resolver might fall through to the *first* `cache` found globally (otherSystem.cache), producing a fabricated cross-system edge. This is the "nearest wins is fine; first-found globally is a bug" distinction.

**Warning sign.** A relative peer resolves to a target in a different parent than the source, with no absolute path having been authored.

**Prevention.** Relative resolution searches *only* the source unit's parent's direct children. If not found there, fall back to **absolute** (if the peer matches an absolute path or contains `.`), else emit a validator "undefined peer" error (reusing the existing `ValidateReferences` machinery, `rules.go:14-45`, including the Levenshtein suggestion via `FormatSuggestion`). Never fall through to a global first-match. The "relative-first-then-absolute-fallback" order (HS-2) is the backward-compat-safe contract.

**Phase:** Relative-peer implementation phase.

#### RP-2: Relative peer resolution must be a no-op for existing models (backward compat)
**Pitfall.** Every existing C4Drill model uses absolute peers (there is no relative syntax today). The relative-resolution pass must not change a single resolved peer for any existing model, or it breaks the non-negotiable backward-compat guarantee.

**Warning sign.** Any existing golden test (`internal/view/scope_test.go`, integration tests) fails after enabling relative resolution.

**Prevention.** Gate relative resolution: a peer is treated as relative *only if* it is a bare name (no `.`) AND it does not exactly match any top-level unit path AND it does not already resolve as absolute in the index. In practice, run the existing absolute resolution first; only if that fails does the relative search kick in. Equivalent formulation: try absolute, fall back to relative, **never both for the same peer**. Add a corpus test: run every example/fixture TOML in the repo through the new pipeline and assert the set of `(source, resolved-peer)` pairs is identical to today. (The HS-2 "relative-first-then-absolute" phrasing refers to *search order within resolution*, not "rewrite all peers"; for backward compat the absolute path must still resolve identically — make this explicit in the design to avoid a contributor reading "relative-first" as "rewrite absolute peers.")

**Phase:** Relative-peer implementation phase — the backward-compat test is the acceptance criterion.

#### RP-3: Relative peer inside template (HS-2) — restated as a checklist item
See HS-2. The non-negotiable sub-rule: resolve against **instantiation-site parent**, not template lexical parent. This must be in the relative-peer phase's tests even though it is exercised via templates.

---

### (4) OPTIONAL-NAME / HUMANIZATION pitfalls

#### HU-1: Acronym handling — gRPC vs Grpc, IDP vs Idp, APIs vs Apis
**Pitfall.** Humanizing a unit key into a display name (when `name` is omitted) requires title-casing. Go's `strings.Title` is **deprecated since Go 1.18** (it does not handle Unicode word boundaries and is not language-aware). A naive replacement (capitalize first letter of each word) turns `grpcApi` → `Grpcapi` or `gRPC` → `GRPC`/`Grpc`, `idp` → `Idp`, `apis` → `Apis`. None of these match author expectations.

**Warning sign.** Rendered unit names look unprofessional ("Idp Service", "Apis Gateway") and authors add explicit `name` fields purely to dodge the humanizer — defeating the feature's purpose.

**Prevention.** This needs an explicit acronym policy because no library gets it right out of the box:
- Recommended library: `golang.org/x/text/cases` with `cases.Title(language.English)` (the standard successor to `strings.Title`). But it does NOT know `gRPC`/`IDP`/`API` are special.
- For acronym preservation, maintain a **small C4-relevant acronym allowlist** (`gRPC`, `gRPC`, `API`, `APIs`, `IDP`, `OAuth`, `SAML`, `REST`, `GraphQL`, `SQL`, `NoSQL`, `UI`, `URL`, `URI`, `SSH`, `TLS`, `HTTP`, `HTTPS`, `gRPC`, `CI`, `CD`, `SaaS`, `PaaS`, `IaaS`) and treat matches as atomic tokens (do not re-case them). Hand-rolling a 30-line tokenizer (split on non-alnum, lookup token case-insensitively, emit either the canonical acronym or title-case the token) is cheaper and more predictable than pulling `gobuffalo/flect` or `serenize/snaker`, which have their own (different) acronym lists and behaviors.
- Document the allowlist and how to extend it. Accept that it will never be complete; the escape hatch is an explicit `name` field.
- This is genuinely a **decision point** (see Decision Required): library + allowlist vs hand-roll vs "don't humanize, require explicit name."

**Phase:** Optional-name discuss phase (pick the approach) → implementation.

#### HU-2: Humanization runs AFTER template param substitution, not before
**Pitfall.** A templated unit may omit `name` and rely on humanization of its instantiation key (the `[[use]] name=...` value, possibly after `${param}` substitution). If humanization runs before substitution, it title-cases the literal `${name}` token. If it runs on the instantiation key before substitution, it humanizes the wrong string.

**Warning sign.** Templated units with omitted names render as "${name} Service" or with the wrong casing; non-templated omitted-name units render fine.

**Prevention.** Pin humanization to run **after template expansion** (so it sees the final, substituted name/key) and **only when the `name` field is empty** (explicit `name` always wins — it is never re-humanized). Pipeline position: `include → template-expand → humanize-omitted-names → relative-peer-resolve → validate`. (Humanize before relative-peer-resolve so that any name-derived logic in resolution sees final names; though resolution is path-based, not name-based, so order is mostly cosmetic — but keep humanize before validation so the validator sees final names in error messages.)

**Phase:** Optional-name implementation phase; the ordering rule is a discuss-phase capture.

#### HU-3: Humanizing an already-human-readable identifier vs a non-readable one
**Pitfall.** A unit key like `userProfile` → `User Profile` (good). A key like `auth_service` → `Auth_service` (underscore preserved badly) or `AuthService` (if underscores are word boundaries). A key like `2fa` → `2fa` (leading digit). The humanizer's word-boundary rules determine quality.

**Warning sign.** Inconsistent name quality depending on key style (camelCase vs snake_case vs kebab-case).

**Prevention.** Define the word-boundary rule explicitly: treat camelCase boundaries, `_`, `-`, and non-alphanumeric as word separators; reconstruct with spaces; apply the acronym allowlist; title-case remaining words. Test the four common key styles. Document that explicit `name` is the escape hatch for anything the humanizer mangles.

**Phase:** Optional-name implementation phase.

---

### (5) BACKWARD-COMPAT / FIELD-NAME COLLISION pitfalls

#### BC-1: Adding `reference` to the allowlist is safe; adding top-level `[template]`/`[include]`/`[use]` tables is NOT — they'd be parsed as units
**Pitfall.** The parser treats unknown keys as subunits. Two locations enforce this:
- `captureDefinitionOrder` (`internal/parser/parser.go:100-157`): it skips only `[properties]` (line 128). A `[template.microservice]` table has `len(parts)==2`, so it falls into the subunit branch (line 139-148) and registers `microservice` under parent `template` in `subunitOrders["template"]` — i.e., the parser thinks there is a top-level unit named `template` with a subunit `microservice`. A `[include]` or `[[include]]` table (`len(parts)==1`) registers a top-level unit named `include`. A `[[use]]` array-of-tables similarly registers `use`.
- `isBuiltinField` (`parser.go:309-316`): the allowlist that distinguishes struct fields from subunits. `reference`, `template`, `include`, `use`, `params`, `parent` are all absent. Adding `reference` is a one-line safe addition (it is a leaf field). Adding `template`/`include`/`use` to this list would make the parser *ignore* them as subunits but would NOT stop `captureDefinitionOrder` from registering them as top-level units — the two functions have independent detection logic.

**Warning sign.** After adding templates, the validator (`ValidateNestingHierarchy`, `rules.go:187`) emits "unit 'template' has type system which cannot have subunits" or "unit 'include' has no incoming or outgoing links" (orphan, `rules.go:125`) — because the parser dutifully created phantom units named `template`/`include`/`use`.

**Prevention.** This is the single most important backward-compat implementation detail. The parser must be changed in **two coordinated places**:
1. `captureDefinitionOrder` (`parser.go:128`): extend the skip rule beyond `properties` to also skip `[include]`, `[[include]]`, `[template.*]`, `[[use]]` (and any other reserved top-level tables). These tables are not units and must not appear in `UnitOrder` or `subunitOrders`.
2. `Parse` (`parser.go:47-96`): after extracting `properties` (line 68-77), extract the new reserved tables (`include`, `template`, `use`) from `rawMap` into dedicated structs *before* the unit-processing loop, and ensure they are not also processed as units. Add a guard in the loop (line 80-93) to skip reserved names.
3. `isBuiltinField` (`parser.go:309`): add `reference` (safe, leaf field). Do NOT add `template`/`include`/`use` here as a substitute for steps 1-2 — those need top-level-table handling, not field handling.
4. Reject (or namespace) any user unit literally named `template`, `include`, or `use` at parse time with a clear "reserved name" error, OR reserve them only at top-level (a nested `[foo.template]` is a legitimate subunit named `template`). Decide scope (see BC-2).

**Phase:** Whichever feature lands first (likely include or templates, per the pipeline) — the parser change is a prerequisite for both, and should be its own small, well-tested change before the features are built on top.

#### BC-2: Reserved-word collision with existing user units named `template`/`include`/`use`
**Pitfall.** An existing user model might already have a top-level unit named `[include]` or `[template]` or `[use]` (unlikely but possible — `use` is a plausible short name). Reserving these names at top-level is a **breaking change** for that model. The non-negotiable backward-compat guarantee means this must be handled.

**Warning sign.** A user upgrades and their model that defined `[use]` as a unit now fails with "reserved name."

**Prevention.** Options (decision required — see Decision Required):
- (a) **Reserve only the table forms C4Drill itself interprets** (`[include]`, `[[include]]`, `[template.*]`, `[[use]]`) and detect the collision: if a model uses one of these names as a plain unit AND as a directive table, error with a clear message. If used only as a plain unit (legacy), treat it as a unit (but then the feature is unavailable for that model). This is messier but maximally backward-compatible.
- (b) **Pick directive-table names unlikely to collide** — e.g. `[c4.include]` / `[c4drill.template]` (namespaced) so they never collide with user units. Costs a little ergonomics, buys zero collision risk. Strongly worth considering given backward-compat is non-negotiable.
- (c) Reserve the bare names and accept the (tiny) break, documented in the changelog. Simplest implementation, violates the "existing models work unchanged" guarantee for the rare colliding model.

Recommendation: (b) namespaced tables, OR (a) collision-detection. Avoid (c). The directive syntax in the include todo (line 56-77) leans toward bare `[include]`/`[[include]]`; revisit in discuss phase against this collision risk.

**Phase:** Discuss phase for both include and templates — the chosen directive syntax determines the collision risk.

#### BC-3: go-toml/v2 strict mode is OFF and must stay OFF (load-bearing for subunits)
**Pitfall.** A contributor "tightening up" parsing might enable `toml.NewDecoder(...).DisallowUnknownFields()` to get strict validation. This would **immediately break all subunit parsing**, because subunits work precisely *because* unknown keys are silently accepted and then reinterpreted as subunit maps via `toml:",inline"` on `Subunits` (`unit.go:71`). With strict mode on, every `[unit.child]` becomes an "unknown field `child`" error.

**Warning sign.** Enabling strict mode makes every multi-level model fail to parse.

**Prevention.** Document (code comment near `parser.go:57` and in the parser package doc) that non-strict unmarshal is **intentional and load-bearing** for subunit parsing, with a pointer to `unit.go:71`. Do not enable `DisallowUnknownFields()` at the Model/Unit level. (If strict checking is ever wanted for specific reserved tables, it would have to be per-decoder on the extracted sub-document, not globally — and even then it fights the inline-subunit design.) When adding the new reserved tables (`include`/`template`/`use`), parse them with their own dedicated struct types via the same non-strict unmarshal + manual extraction pattern already used for `properties` (`parser.go:68-77`).

**Phase:** Parser-change phase for include/templates — add the guard comment as part of that change so future contributors don't regress it.

---

## Testing Implications

The following test cases are *required* by specific pitfalls above — they are not optional nice-to-haves:

| Test | Required by | Why it catches |
|---|---|---|
| Template instantiated **3×** with distinct params; assert each instantiation's `Links[0].Description`, `Subunits`, `LinksFrom` are independent and correct | HS-1 | Shallow-copy aliasing (the corruption is invisible at N=1, appears at N≥2 or on re-run) |
| Re-run expansion **twice** on the same parsed model (idempotency); assert identical output | HS-1 | "Second run reads a corrupted template" — catches mutation of the template registry |
| After full validate pass, two instantiations' `LinksFrom` slices are disjoint (no shared mirror links) | HS-1 | `populateIncomingLinks` append-aliasing via shared backing array |
| Template with **relative** peer, instantiated under **two different parents**; assert each link resolves to its own parent's sibling, not a shared global | HS-2, RP-1, RP-3 | Template-site vs instantiation-site resolution ambiguity |
| **Existing** absolute-peer model: run through new pipeline; assert the `(source, resolved-peer)` set is byte-identical to today | RP-2 | Backward-compat regression — the non-negotiable guarantee |
| **Diamond** include graph (A→B, A→C, B→D, C→D) with no `once`: either includes D twice (and errors on dup) or is explicitly allowed — assert NOT a false "cycle detected" | IN-3 | Stack-vs-seen-set confusion in cycle detection |
| **Self-include** and **mutual** (A↔B) include: assert fatal cycle error | IN-3 | Cycle detection actually works |
| `once=true` diamond: D included once; assert no duplicate-definition error and no missing definition | IN-3 | include_once semantics |
| Multi-file model produces the same SVG (order-insensitive canonicalDOT, per STATE.md) as the equivalent single-file model | IN-1, IN-2 | Merge correctness + UnitOrder semantics |
| Two `[[use]]` blocks producing the **same unit path**: assert hard error naming both | TM-1 | Silent shadowing |
| Template `[[use]]` with a **missing required param**: assert hard error naming the param; assert no literal `${name}` survives in any field | TM-5 | Silent literal-substitution |
| Template referenced **before** its `[template.*]` definition: assert behavior matches the documented choice (allow or reject), not an accident | TM-2 | Forward-reference semantics |
| After enabling relative resolution, every existing fixture/example TOML resolves peers identically (corpus test) | RP-2 | Backward-compat across the whole example set |
| Humanizer on `grpcApi`, `idp`, `userProfile`, `auth_service`, `2fa`: assert expected output per the chosen acronym policy | HU-1, HU-3 | Acronym/title-case quality |
| Templated unit with **omitted** `name`: assert humanization runs post-substitution on the instantiation key, not on `${name}` | HU-2 | Humanize-vs-substitute ordering |
| A model with a top-level `[include]` / `[template.x]` / `[[use]]` table: assert NO phantom unit named `include`/`template`/`use` appears in `UnitOrder` or triggers a validator error | BC-1 | Reserved-table parsing |
| A model with a top-level unit literally named `use` (legacy): assert behavior matches the chosen collision policy (BC-2) | BC-2 | Reserved-word break |
| Temporarily enable `DisallowUnknownFields()`: assert a multi-level model still parses (or document why this test is skipped) — better, a code comment test that the flag is NOT set | BC-3 | Strict-mode regression |

All golden comparisons must use the **order-insensitive canonicalDOT** comparator (STATE.md decision log: sort-normalize, strip layout geometry) — NOT byte-exact `require.Equal`. This is already-established practice for this codebase and is doubly important now that multi-file and template models add another axis of ordering variance.

---

## Decision Required (design forks for the discuss phase)

These pitfalls surface genuine design forks that cannot be resolved by "implementation care" — they need an explicit decision before/during the discuss phase, because the wrong default is silent corruption or a broken backward-compat guarantee.

1. **Relative-peer resolution site for template-authored links (HS-2).** Decision: instantiation-site parent (recommended). The alternative (template-site) is wrong but plausible-sounding; left undecided, an implementer could pick either. This is the single most consequential design decision for the milestone.

2. **Relative resolution search order vs backward compat (RP-2 / HS-2 wording).** "Relative-first" must be clarified: it means relative-search-first *for peers that are not already valid absolute paths* — NOT "rewrite all peers as relative." Existing absolute peers must resolve identically. Pin the exact gate (bare name + no dot + not a top-level key + not already in index).

3. **Forward-reference policy for templates (TM-2).** Allow (recommended, free under Option B) or reject (go-metadot parity). Document the choice; do not let it emerge by accident.

4. **Unresolved `${param}` policy (TM-5).** Error (recommended) vs leave-literal (go-metadot/PlantUML parity). Strictness is a feature here.

5. **Humanization approach and acronym allowlist (HU-1).** Library (`x/text/cases`) + hand-rolled allowlist (recommended) vs pure hand-roll vs `flect`/`snaker` vs "require explicit `name`, no humanize." The allowlist contents are themselves a mini-decision.

6. **Directive-table naming and reserved-word collision policy (BC-1, BC-2).** Bare `[include]`/`[template]`/`[use]` (ergonomic, collides with rare legacy units) vs namespaced `[c4.include]`/`[c4drill.template]` (collision-proof, uglier). This determines whether backward compat is truly non-negotiable or has a tiny carve-out. Decide before fixing the directive syntax.

7. **`UnitOrder`/`SubunitOrder` merge semantics across included files (IN-2).** Append-in-include-order (recommended for top-level) and append-to-existing-unit (for subunits) — pin both, with the same-name-conflict rule (hard error). The alternative (interleave at include-site position, or last-wins) is worse and must be explicitly rejected.

8. **Diamond-include behavior without `once` (IN-3).** Include-twice-and-error-on-dup (recommended, signals the author to add `once`) vs silent-last-wins vs silent-first-wins. The diamond is legal; the second inclusion's effect on duplicate definitions is the decision.

Each of these should be captured as an explicit decision in STATE.md's decision log when the discuss phase resolves it, with a one-line rationale, so the implementer has a single source of truth and does not re-litigate.

---

## Sources

- C4Drill codebase (verified at writing): `internal/model/unit.go`, `internal/model/link.go`, `internal/model/properties.go`, `internal/parser/parser.go`, `internal/parser/errors.go`, `internal/validator/{index,rules,validator}.go`, `internal/view/{view,scope}.go`, `internal/graph/builder.go`, `cmd/c4drill/root.go`, `go.mod`
- `.planning/STATE.md` — decision log (go-graphviz order-nondeterminism, canonicalDOT, HTML-label quirks, D-05 multiplicity, D-12 validator-as-gatekeeper)
- `.planning/todos/pending/2026-08-08-include-directive-multi-file-diagrams.md`
- `.planning/todos/pending/2026-08-08-unit-templates-parametrized-definitions.md`
- TOML v1.0.0 spec (toml.io) — "you cannot define a table more than once"; dotted-key table creation rules
- go-toml/v2 (pelletier) — `Unmarshal` is non-strict by default; `Decoder.DisallowUnknownFields()` is opt-in
- Go stdlib — `strings.Title` deprecated since 1.18; `golang.org/x/text/cases` is the successor

*Pitfalls research for: C4Drill v1.10 Model Composition (include / templates / relative-peer / humanization).*
*Researched: 2026-08-08.*
