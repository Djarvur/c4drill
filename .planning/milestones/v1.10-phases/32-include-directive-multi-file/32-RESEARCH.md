# Phase 32: Include directive (multi-file) - Research

**Researched:** 2026-08-08
**Domain:** Go CLI TOML multi-file composition (parse-then-merge structs, recursive include, cycle detection)
**Confidence:** HIGH

## Summary

Phase 32 assembles a single `*parser.Model` from multiple TOML files via `[[include]]` directives. The implementation is a pure pre-processing pass (`include.Resolve`) inserted between `parser.ParseFile` and `template.Expand` in `cmd/c4drill/root.go`. It adds exactly one new `parser.Model` field (`Includes []IncludeDirective`), one new extraction block in `parser.Parse`, and one new internal package (`internal/include`). Zero new external dependencies: everything is hand-rolled Go on stdlib `path/filepath` plus the existing `parser.ParseFile` recursion. The validator, view, render, and graph packages are untouched — the merged model is structurally indistinguishable from a hand-authored single-file model.

The four locked decisions (D-09 append ordering, D-10 cross-file subunits, D-11 same-file-diamond-dedup/cross-file-error, D-12 hard-error missing include) resolve every design fork the milestone research left open for this phase (SUMMARY §6 IN-2/IN-3; PITFALLS IN-3/IN-5). The single load-bearing correctness distinction is three separate data structures for three separate concerns: a **stack** for cycle detection (current ancestry chain), a **global visited-set** for `once=true` dedup, and **canonical-path equality** for automatic same-file diamond dedup (D-11). Conflating any two produces either false "cycle detected" errors on diamonds or silent double-inclusion. The merge is a struct-union in Go (Option B per the original todo) — never byte-concatenation — because TOML forbids redefining a table and byte-concat would produce go-toml `DecodeError`s whose line numbers point into the concatenated blob (PITFALLS IN-1).

**Primary recommendation:** Implement `Resolve(entry *parser.Model, entryDir string) (*parser.Model, error)` in a new `internal/include` package. Walk `entry.Includes` in order; for each directive, canonicalize `path` relative to the *including file's* directory via `filepath.Abs(filepath.Join(includingDir, dir.Path))` then `filepath.Clean`, check the stack (cycle) and visited-set (`once`), recurse via `parser.ParseFile` + `Resolve` on the included model with the included file's directory as the new baseDir, then merge the resolved included model into the entry. Merge = union `Units` (cross-file dup path → hard error naming both files), append `UnitOrder`, deep-merge `Subunits`/`SubunitOrder` per parent (D-10), root-wins `Properties` with conflict → hard error, union `Templates`/`Instantiations` (cross-file dup → hard error). Drain `entry.Includes` to empty on the merged model. Total estimated size ~120-160 LOC across `resolve.go` + `merge.go`.

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| `[[include]]` directive parsing (rawMap → `Model.Includes`) | Parser tier (`internal/parser`) | — | Reserved-table extraction mirrors the existing `properties` extraction at `parser.go:68-77`; the BC-1 skip-rule already lands in Phase 31 Plan 1 |
| Recursive include resolution + cycle detection + `once` dedup | Composition tier (NEW `internal/include`) | — | A pure pre-parse pass taking/returning `*parser.Model`; keep it out of `parser` so `parser` stays single-file concern |
| Multi-file model merge (Units/UnitOrder/Properties/Templates/Subunits) | Composition tier (`internal/include/merge.go`) | — | Merge rules (D-09/D-10/D-11/INC-08) are include-specific policy; belong with the resolver |
| Pipeline wiring (call `include.Resolve` first, before `template.Expand`) | CLI orchestration (`cmd/c4drill/root.go`) | — | The single insertion point between `ParseFile` (:112) and `Validate` (:118); include must run first so included templates are visible to `[[use]]` (XC-02) |
| Validation of the merged model | Validator tier (`internal/validator`) — UNCHANGED | — | Validator is the single gatekeeper (STATE.md D-12); `include.Resolve` must produce a model indistinguishable from a single-file model so `validator.Validate` consumes it unchanged |
| Golden-test comparison of multi-file vs single-file output | Test tier (`test*/canonicalDOT`) | — | Multi-file goldens MUST use the order-insensitive canonicalDOT comparator (STATE.md DI-1), never byte-exact `require.Equal` |

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `path/filepath` (Go stdlib) | go1.26.5 | `filepath.Abs`, `filepath.Clean`, `filepath.Join`, `filepath.Dir` — canonicalize include paths for cycle detection, `once` visited-set, same-file diamond dedup | Idiomatic Go; already used in `cmd/c4drill/root.go` (`filepath.Dir`, `filepath.Base`); zero new dep. `[VERIFIED: Go stdlib]` |
| `os` (Go stdlib) | go1.26.5 | `os.ReadFile` for included files (via `parser.ParseFile`, already at `parser.go:328`) | Existing pattern. `[VERIFIED: codebase parser.go:328]` |
| `github.com/pelletier/go-toml/v2` | v2.2.4 (current `go.mod`) | `toml.Unmarshal`/`toml.Marshal` for extracting the `[[include]]` array from rawMap (mirrors `properties` extraction) | Already the project's TOML library; non-strict mode is load-bearing (BC-3). `[VERIFIED: go.mod]` |
| `github.com/stretchr/testify` | current | `require`/`assert` for the test suite | Already used in `internal/parser/parser_test.go`. `[VERIFIED: codebase]` |

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `slices` (Go stdlib) | go1.26.5 | `slices.Contains` for stack/visited-set membership checks; `append` for UnitOrder concatenation | Already a project idiom (`parser.go:6` imports `slices`). `[VERIFIED: codebase parser.go:6]` |
| `fmt` (Go stdlib) | go1.26.5 | `fmt.Errorf` with `%w` for wrapping include errors; building cycle/missing-file error messages | Standard. `[VERIFIED: Go stdlib]` |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Hand-rolled merge in `internal/include` | A generic TOML deep-merge library (koanf, viper) | Rejected (SUMMARY §2): wrong shape — operates on generic maps, throws away the typed `*parser.Model`, and reinvents merge rules. The merge is a small Go function over known structs. |
| Pre-parse byte concatenation (Option A) | Parse-then-merge structs (Option B — CHOSEN) | Option A rejected (PITFALLS IN-1): TOML forbids redefining a table; byte-concat produces go-toml `DecodeError`s whose line numbers point into the blob, defeating `wrapDecodeError`. Struct merge gives type safety + per-file error fidelity. |
| Sidecar `IncludeGraph` struct | Dedicated `Model.Includes` field (CHOSEN) | Sidecar rejected (ARCHITECTURE §"Model Extension"): would force changing signatures of `validator.Validate`, `view.Generate*`, `collectExpandedPaths`, every caller. Dedicated field is invisible to consumers that read only `.Units`/`.UnitOrder`/`.Properties`. |
| `internal/compose/` or function in `internal/parser/` | `internal/include/` package (CHOSEN) | A dedicated package keeps single-file parsing (`parser`) separate from multi-file composition; mirrors the research-recommended `internal/include/Resolve`. Discretion resolved toward the research default for consistency with `internal/template/`, `internal/peer/` siblings. |

**Installation:** None — no new packages. All dependencies are already in `go.mod`.

**Version verification:**
```bash
grep "pelletier/go-toml" go.mod          # v2.2.4 (current; SUMMARY's v2.4.3 bump is NOT required for Phase 32 and is NOT in scope)
go version                                # go1.26.5 darwin/arm64
```

## Package Legitimacy Audit

> This phase installs **zero** external packages. All work is hand-rolled Go on stdlib + existing project deps. No slopcheck/npm-view/pip step is required. `[N/A — no new packages]`

## Architecture Patterns

### System Architecture Diagram

```
CLI (c4drill model.toml)
  │
  ▼
parser.ParseFile(inputPath)          ──▶ *parser.Model (entry; .Includes populated)
  │                                       (BC-1 skip for [[include]] already lands in Phase 31 Plan 1)
  ▼
include.Resolve(entryModel, entryDir) ◀── NEW, Stage 1a (root.go ~:113)
  │
  │  for each IncludeDirective in entry.Includes (order):
  │    1. canonicalize path: filepath.Abs(filepath.Join(includingDir, dir.Path)); filepath.Clean
  │    2. cycle check: if canonical path ON STACK  ──▶ fatal *ParseError naming the cycle (INC-04)
  │    3. once check:  if dir.Once AND path IN VISITED ──▶ skip (INC-06)
  │    4. same-file diamond: if path IN VISITED (regardless of Once) ──▶ skip silently (D-11 automatic dedup)
  │    5. recurse: included, err := parser.ParseFile(canonicalPath)
  │       missing file ──▶ fatal *ParseError naming path + including file (INC-10, D-12 — no optional)
  │    6. included, err = include.Resolve(included, filepath.Dir(canonicalPath))  (transitive, INC-03)
  │    7. push canonicalPath on STACK; add to VISITED
  │    8. entry = merge(entry, included, includingFile, includedFile)
  │    9. pop STACK
  │  drain entry.Includes = nil
  ▼
*parser.Model (merged; .Includes empty; .Units/.UnitOrder/.Properties/.Templates/.Instantiations assembled)
  │
  ▼
template.Expand(merged)              ──▶ XC-02: templates from included files now visible to [[use]]
  │
  ▼
validator.Validate(merged)           ──▶ UNCHANGED gatekeeper (STATE.md D-12)
  │
  ▼
view.Generate* / graph.Build / render.Render   (all UNCHANGED)
```

### Recommended Project Structure

```
internal/
├── include/                  # NEW package — multi-file composition
│   ├── resolve.go            # Resolve(entry, entryDir): recursive walk, cycle stack, once visited-set, same-file dedup
│   ├── merge.go              # merge(dst, src, dstFile, srcFile): per-field merge rules (D-09/D-10/D-11/INC-08)
│   ├── include_test.go       # table-driven unit tests for Resolve (cycle, diamond, once, transitive, missing)
│   └── testdata/             # multi-file fixtures (main.toml + auth.toml + templates.toml + cycle fixtures)
├── parser/
│   └── parser.go             # add IncludeDirective type + Model.Includes field + rawMap["include"] extraction in Parse
└── ... (model, validator, view, graph, render — UNCHANGED)
cmd/c4drill/
└── root.go                   # insert include.Resolve(m, filepath.Dir(inputPath)) between ParseFile and Validate
testdata/                     # existing single-file fixtures (unchanged); new multi-file fixtures live under internal/include/testdata
```

### Pattern 1: Reserved-table extraction mirrors `properties`

**What:** The `[[include]]` array-of-tables is extracted from `rawMap` in `parser.Parse` *before* the unit loop, exactly as `properties` is extracted at `parser.go:68-77`. The BC-1 skip-rule (so `[[include]]` does not become a phantom unit named `include`) lands in Phase 31 Plan 1; Phase 32 only adds the extraction into `Model.Includes`.

**When to use:** Whenever adding a new reserved top-level table.

**Example** (mirror of `parser.go:68-77`):
```go
// Source: internal/parser/parser.go:68-77 (existing properties extraction) + ARCHITECTURE-v1.10.md:230-232
// Extract [[include]] directives if present (mirrors properties extraction above).
if inc, ok := rawMap["include"]; ok {
    incData, err := toml.Marshal(inc)
    if err != nil {
        return nil, &ParseError{Message: "failed to marshal include", Cause: err}
    }
    if err := toml.Unmarshal(incData, &m.Includes); err != nil {
        return nil, wrapDecodeError(err)
    }
    delete(rawMap, "include") // guarantee it never enters the unit loop
}
```

### Pattern 2: Three data structures for three concerns (D-11)

**What:** Cycle detection, `once` dedup, and same-file diamond dedup use distinct mechanisms:

| Concern | Data structure | Semantics |
|---------|----------------|-----------|
| Cycle (INC-04) | **stack** `[]string` of canonical paths in the current ancestry chain | Re-entering a file on the stack = fatal. A diamond (A→B→D, A→C→D) is NOT a cycle because D is not ancestral to itself on either branch. |
| `once=true` dedup (INC-06) | **global visited-set** `map[string]bool` of canonical paths ever included | Opt-in per directive (`Once` field). If a directive has `Once=true` and its canonical path is already in the visited-set, skip it. |
| Same-file diamond dedup (D-11) | **canonical-path equality** against the same visited-set | Automatic (no flag): if the same canonical path is reached via two non-ancestral paths, its units are byte-identical by construction (same file on disk) — silently skip the second inclusion. |

**Implementation note:** The visited-set serves both `once` and same-file-diamond. When a file is about to be included: if it is on the stack → cycle fatal; else if it is in the visited-set → skip (covers both `once=true` and the automatic diamond case — in both, the file's content is already merged). Push onto the stack before recursing into children; add to the visited-set the first time the file is actually included (merged). The distinction between "skip because once/diamond" is invisible to the user (both just skip), which is correct: a same-file diamond's content is identical whether included once or twice, so skipping is always safe; the only thing that would be unsafe is silently *dropping* a *different* file's contribution, which the cross-file dup-path hard-error (D-11) catches at merge time.

**When to use:** This is THE core correctness pattern for Phase 32. Conflating stack and visited-set produces false "cycle detected" on diamonds (PITFALLS IN-3).

### Pattern 3: Merge as struct-union with per-field rules

**What:** `merge(dst, src *parser.Model, dstFile, srcFile string) (*parser.Model, error)` applies one rule per field family:

| Field | Rule | Error condition |
|-------|------|-----------------|
| `Units` (top-level map) | Union; `dst.Units[k] = src.Units[k]` for each new key | Cross-file dup key → `*ParseError` naming both files (INC-07, D-11) |
| `UnitOrder` (slice) | `dst.UnitOrder = append(dst.UnitOrder, src.UnitOrder...)` | None — append preserves authoring order (D-09) |
| `Units[*].Subunits` + `SubunitOrder` (per-unit) | For each top-level unit that exists in BOTH dst and src: `unit.Subunits[k] = srcUnit.Subunits[k]` for new child keys; `unit.SubunitOrder = append(unit.SubunitOrder, srcUnit.SubunitOrder...)` | Cross-file dup subunit key → `*ParseError` naming both files (D-10) |
| `Properties` | Root/first-seen wins: only copy non-zero fields from src if dst's field is zero-value | Any non-zero field conflict → `*ParseError` naming both files (INC-08) |
| `Templates` | Union; `dst.Templates[k] = src.Templates[k]` for new key | Cross-file dup template name → `*ParseError` naming both files (XC-02 requires templates flow through; dup = author error) |
| `Instantiations` | `dst.Instantiations = append(dst.Instantiations, src.Instantiations...)` | None — order preserved, drained by Phase 31's `template.Expand` later |
| `Includes` | Not merged — src's includes were already resolved by the recursive `Resolve` call; dst's are drained to nil at the end of the top-level Resolve | None |

**When to use:** Every `merge` call during the recursive walk.

**Example:**
```go
// Source: ARCHITECTURE-v1.10.md:245-253 (merge table) + CONTEXT.md D-09/D-10/D-11
func merge(dst, src *parser.Model, dstFile, srcFile string) (*parser.Model, error) {
    // Units union (INC-07)
    for k, u := range src.Units {
        if existing, ok := dst.Units[k]; ok {
            // Same top-level unit in both files: merge subunits (D-10), error on dup child
            if err := mergeSubunits(existing, u, k, dstFile, srcFile); err != nil {
                return nil, err
            }
            continue
        }
        dst.Units[k] = u
        dst.UnitOrder = append(dst.UnitOrder, k) // append in include order (D-09)
    }
    // Top-level UnitOrder from src is already handled per-key above (only new keys append).
    // ... Properties (root-wins), Templates (union), Instantiations (append) ...
    return dst, nil
}
```

### Anti-Patterns to Avoid

- **Byte-concatenating files then parsing once** (Option A) — TOML forbids redefining a table; go-toml `DecodeError` line numbers point into the blob (PITFALLS IN-1). Always parse-then-merge structs.
- **Using a global seen-set for cycle detection** — a diamond (A→B→D, A→C→D) would be falsely flagged as a cycle on D's second inclusion (PITFALLS IN-3). Use a stack for cycles, a separate visited-set for `once`/diamond.
- **Resolving include paths relative to the CLI cwd** — "works on my machine, breaks in CI" (PITFALLS IN-5). Always resolve relative to the *including file's* directory.
- **Skipping canonicalization (`filepath.Abs` + `filepath.Clean`)** — `./x.toml`, `x.toml`, and `./a/../x.toml` would dodge cycle detection and visited-set dedup via path-string differences (PITFALLS IN-4). Always canonicalize before stack/set operations.
- **Re-touching `captureDefinitionOrder` or the BC-1 skip** — already lands in Phase 31 Plan 1. Phase 32 CONSUMES the skip; re-touching it creates a merge conflict with the concurrent Phase 31/28 work (CONCURRENCY NOTE in the run prompt).
- **Enabling `toml.DisallowUnknownFields()`** — load-bearing OFF; the inline-subunit trick (`unit.go:71` `toml:",inline"`) depends on non-strict mode (PITFALLS BC-3, STATE.md DI-1).
- **Using byte-exact `require.Equal` for multi-file goldens** — go-graphviz layout is byte-nondeterministic AND multi-file adds another ordering axis. Use the canonicalDOT order-insensitive comparator (STATE.md DI-1).
- **Round-trip-copying `Link`s through marshal/unmarshal** — would silently reset `Link.Mirror` (`link.go:67`, `toml:"-"`), re-breaking multiplicity counting. (Relevant if deep-copy is ever needed; for Phase 32's merge, units are moved by pointer, not copied — no aliasing risk because each included file's units are distinct map entries. But document this constraint.)

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Path canonicalization | Custom path normalization | `filepath.Abs` + `filepath.Clean` (stdlib) | Handles `..`, symlinks (Abs), trailing slashes, OS separators correctly; reinventing it misses edge cases |
| Reading + parsing an included TOML file | A separate file-read+parse in `internal/include` | `parser.ParseFile(canonicalPath)` (existing `parser.go:323`) | Reuses the exact same parsing path as the entry file; guarantees BC-1 skip + properties extraction + (Phase 31) template/use extraction all apply to included files uniformly |
| TOML array-of-tables extraction | Manual map-walking for `[[include]]` | `toml.Marshal` the rawMap value then `toml.Unmarshal` into `[]IncludeDirective` (mirror of `properties` extraction at `parser.go:68-77`) | The established project pattern; handles TOML types correctly; non-strict mode stays load-bearing |
| Error wrapping | Custom error types | `*parser.ParseError{Message, Context, Cause}` (existing `errors.go:13`) | Consistent with all other parser errors; `Context` field carries the filename for multi-file attribution |

**Key insight:** Phase 32 adds almost no new machinery — it composes existing primitives (`ParseFile`, `ParseError`, `filepath.Abs/Clean/Join`, the `properties`-extraction idiom) into a recursive resolver + struct merge. The novelty is the policy (D-09/D-10/D-11/D-12), not the mechanism.

## Runtime State Inventory

> Phase 32 is a **pure code feature** (new package + pipeline insertion), not a rename/refactor/migration. SKIPPED — no stored data, live service config, OS-registered state, secrets/env vars, or build artifacts carry a renamed string. (Verified: no string-replacement, no data migration, no config rename.)

## Common Pitfalls

### Pitfall 1: Diamond falsely flagged as cycle (PITFALLS IN-3)

**What goes wrong:** A diamond include graph (A→B→D, A→C→D) is flagged "cycle detected" when there is no cycle, OR `once=true` silently drops D's content with no explanation.
**Why it happens:** Using a single global seen-set for both cycle detection and dedup, instead of a stack (current ancestry) for cycles and a visited-set for dedup.
**How to avoid:** Stack for cycles, visited-set for `once`/diamond. The acceptance test: a diamond WITHOUT `once` either (a) includes D twice and errors on the resulting cross-file duplicate unit path (if D defines a unit also defined elsewhere), or (b) includes D twice and silently produces the same merged model as including D once (because D's content is byte-identical both times — D-11 automatic same-file dedup). Both outcomes are correct; a false "cycle detected" is the bug.
**Warning signs:** "cycle detected" error on a model the author believes is acyclic; or a definition vanishing after adding `once=true`.

### Pitfall 2: Relative path resolves against CLI cwd (PITFALLS IN-5)

**What goes wrong:** The same model file works or breaks depending on the directory the user invokes `c4drill` from.
**Why it happens:** Resolving include paths relative to `os.Getwd()` instead of the including file's directory.
**How to avoid:** Pass `filepath.Dir(includingFilePath)` as the baseDir for each resolution level; `filepath.Abs(filepath.Join(baseDir, dir.Path))`. The acceptance test: `cd` elsewhere and run the CLI; the multi-file model renders identically.
**Warning signs:** "works on my machine, breaks in CI".

### Pitfall 3: Non-canonical paths dodge cycle/dedup detection (PITFALLS IN-4)

**What goes wrong:** `./x.toml` and `x.toml` and `./a/../x.toml` are treated as different files, so a real cycle (A→`./b.toml`→`../a.toml`) evades the stack check, or `once` fails to dedup.
**Why it happens:** Comparing raw path strings instead of canonicalized paths.
**How to avoid:** `filepath.Clean(filepath.Abs(path))` for every path before pushing on the stack or checking the visited-set. Use the canonical form as the map key everywhere.
**Warning signs:** A cycle that should be caught causes stack overflow / hang; `once=true` includes a file twice.

### Pitfall 4: Properties conflict silently overrides (INC-08)

**What goes wrong:** An included file's `[properties]` overrides the entry file's `name`/`description`, producing a diagram with the wrong title.
**Why it happens:** Last-wins or first-wins merge without conflict detection.
**How to avoid:** Root-file-wins: the entry model's `Properties` is authoritative; if an included model has a non-zero value in a field where the entry also has a non-zero value, hard-error naming both files. (Lean conflict-error for all `Properties` fields per CONTEXT discretion — safer than selective name/description-only.)
**Warning signs:** The rendered diagram's title comes from an included file.

### Pitfall 5: Missing include produces a generic "file not found" (INC-10, D-12)

**What goes wrong:** A missing include file surfaces as a generic `os.ReadFile` error with no indication of which file included it.
**Why it happens:** Letting `parser.ParseFile`'s raw `os.ReadFile` error propagate without wrapping.
**How to avoid:** Wrap every include-file read failure in `*parser.ParseError{Message: "include not found", Context: fmt.Sprintf("%s (included from %s)", canonicalPath, includingFile), Cause: err}`. Hard-error unconditionally — no `optional` flag (D-12).
**Warning signs:** Error message says "failed to read file: foo.toml" with no "included from main.toml".

### Pitfall 6: Deep-merge of Subunits drops or duplicates children (D-10)

**What goes wrong:** An included file's `[linuxSystem.auth]` (subunit of an entry-defined `[linuxSystem]`) is either lost or creates a phantom top-level unit `auth`.
**Why it happens:** Merge only unions top-level `Units` and forgets to recurse into `Subunits` of units present in both files; OR the parser fails to attach the cross-file subunit to its parent.
**How to avoid:** `mergeSubunits(dstUnit, srcUnit, path, dstFile, srcFile)`: for each child key in `srcUnit.Subunits`, if absent from `dstUnit.Subunits`, add it and append to `dstUnit.SubunitOrder`; if present in both, hard-error naming both files. This requires `captureDefinitionOrder`'s subunit capture to have correctly grouped `[linuxSystem.auth]` under parent `linuxSystem` in the included file — which it does (it groups by `parts[0]`, `parser.go:139-148`), so the included file's `Model.Units["linuxSystem"].Subunits["auth"]` exists.
**Warning signs:** A subunit defined in an included file under a parent defined in the entry file vanishes, or appears as a top-level unit.

### Pitfall 7: `go-toml/v2` strict mode regression (BC-3)

**What goes wrong:** Enabling `DisallowUnknownFields()` breaks all subunit parsing.
**Why it happens:** A contributor "tightening up" parsing.
**How to avoid:** Do NOT enable `DisallowUnknownFields()` anywhere in the new extraction or the include resolver. Add a guard comment near the `[[include]]` extraction mirroring the existing `properties` extraction's non-strict reliance.
**Warning signs:** Every multi-level model fails to parse with "unknown field" errors.

## Code Examples

### IncludeDirective type + Model field

```go
// Source: CONTEXT.md "Claude's Discretion" + ARCHITECTURE-v1.10.md:160-168
// internal/parser/parser.go (additions)
package parser

// IncludeDirective represents one [[include]] table.
type IncludeDirective struct {
    Path string `toml:"path"` // relative to the including file's directory
    Once bool   `toml:"once"`  // include_once semantics (PlantUML !include_once)
}

type Model struct {
    Properties model.Properties `toml:"properties"`
    UnitOrder  []string
    Units      map[string]*model.Unit
    // v1.10 Phase 31 fields (Templates, Instantiations) land in Phase 31 Plan 1.
    // Includes is the [[include]] directives, captured in file order; consumed by
    // internal/include.Resolve and drained to nil after resolution.
    Includes []IncludeDirective `toml:"-"`
}
```

### Resolve signature and recursion (ARCHITECTURE signature with baseDir)

```go
// Source: ARCHITECTURE-v1.10.md:108,237 (Resolve(m, baseDir)) — the baseDir param is REQUIRED
// for relative-to-including-file path resolution (INC-02); the CONTEXT's simplified
// Resolve(m *parser.Model) signature cannot resolve relative paths without it.
package include

import (
    "fmt"
    "path/filepath"

    "github.com/Djarvur/c4drill/internal/parser"
)

// maxIncludeDepth is a defense-in-depth cap for pathological (acyclic but huge) graphs.
const maxIncludeDepth = 100

// Resolve recursively resolves all [[include]] directives in m, relative to entryDir,
// merging included files into m. The merged model's .Includes is drained to nil.
// entryDir is filepath.Dir(inputPath) passed from cmd/c4drill/root.go.
func Resolve(entry *parser.Model, entryDir string) (*parser.Model, error) {
    visited := make(map[string]bool) // canonical path -> already merged (once + diamond)
    return resolve(entry, entryDir, nil, visited)
}

func resolve(m *parser.Model, includingDir string, stack []string, visited map[string]bool) (*parser.Model, error) {
    if len(stack) > maxIncludeDepth {
        return nil, &parser.ParseError{Message: fmt.Sprintf("include depth cap (%d) exceeded", maxIncludeDepth), Context: strings.Join(stack, " -> ")}
    }
    for _, dir := range m.Includes {
        absPath := filepath.Clean(filepath.Abs(filepath.Join(includingDir, dir.Path))) // canonicalize (IN-4/IN-5)
        // cycle check (stack)
        for i, p := range stack {
            if p == absPath {
                cycle := append(stack[i:], absPath)
                return nil, &parser.ParseError{Message: "include cycle detected: " + strings.Join(cycle, " -> ")}
            }
        }
        // once + same-file diamond dedup (visited-set)
        if visited[absPath] {
            continue // already merged; content is byte-identical, safe to skip
        }
        // parse the included file
        included, err := parser.ParseFile(absPath)
        if err != nil {
            return nil, &parser.ParseError{Message: "include not found", Context: fmt.Sprintf("%s (included from %s)", dir.Path, includingDir), Cause: err}
        }
        // recurse (transitive includes, INC-03)
        included, err = resolve(included, filepath.Dir(absPath), append(stack, absPath), visited)
        if err != nil {
            return nil, err
        }
        visited[absPath] = true
        // merge (D-09/D-10/D-11/INC-08)
        if m, err = merge(m, included, includingDir, absPath); err != nil {
            return nil, err
        }
    }
    m.Includes = nil // drain
    return m, nil
}
```

### Pipeline insertion (root.go)

```go
// Source: ARCHITECTURE-v1.10.md:104-110, CONTEXT.md "Pipeline ordering"
// cmd/c4drill/root.go, between :115 (after ParseFile) and :118 (before Validate)
m, err := parser.ParseFile(inputPath)
if err != nil {
    return fmt.Errorf("parse: %w", err)
}

// Stage 1a: Resolve includes (recursive, merges *parser.Model structs) — runs FIRST
if m, err = include.Resolve(m, filepath.Dir(inputPath)); err != nil {
    return fmt.Errorf("include: %w", err)
}

// (Stage 1b template.Expand lands in Phase 31; Stage 1c peer.Resolve lands in Phase 30.
//  When those lands, they slot here, AFTER include.Resolve — pipeline order is load-bearing.)

valErrors := validator.Validate(m) // UNCHANGED
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Single-file TOML model only (`parser.ParseFile` → `Validate`) | Multi-file composition via `[[include]]` (ParseFile → include.Resolve → Validate) | Phase 32 (v1.10) | Large models can be split; template libraries isolated (XC-02) |
| `parser.Parse` discards the file path after `os.ReadFile` | `include.Resolve` threads `filepath.Dir(path)` through recursion | Phase 32 | Relative-to-including-file path resolution (INC-02) requires the path; Parse itself stays path-agnostic (single-file concern) |

**Deprecated/outdated:**
- The CONTEXT's simplified `Resolve(m *parser.Model) (*parser.Model, error)` signature (without baseDir) is **incomplete** — it cannot resolve relative paths. Use `Resolve(m *parser.Model, entryDir string) (*parser.Model, error)` per ARCHITECTURE-v1.10.md:108. (Resolved — Discretion item.)

## Assumptions Log

> All claims in this research were verified against the codebase, the milestone research docs (SUMMARY/ARCHITECTURE/PITFALLS, all dated 2026-08-08), or Go stdlib docs. No `[ASSUMED]` tags.

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | `filepath.Abs(filepath.Join(...))` then `filepath.Clean` is the right canonicalization order | Pattern 2, Code Examples | Low — `filepath.Abs` already calls `Clean` on the result per Go stdlib docs; the explicit `Clean` is belt-and-suspenders. `[VERIFIED: Go stdlib `path/filepath` docs]` |
| A2 | The cross-file subunit merge (D-10) works because `captureDefinitionOrder` groups `[linuxSystem.auth]` under parent `linuxSystem` in the included file | Pitfall 6 | Low — verified at `parser.go:139-148` (`parts[0]` grouping). The included file's `Model.Units["linuxSystem"].Subunits["auth"]` will exist. |

**Both items are `[VERIFIED]`, not `[ASSUMED]`.** Table included for transparency; no user confirmation needed.

## Open Questions (RESOLVED)

1. **`Resolve` signature: `(m *parser.Model)` vs `(m *parser.Model, entryDir string)`?**
   - What we know: CONTEXT lists the simplified signature; ARCHITECTURE and the original todo require a baseDir for INC-02 (relative-to-including-file).
   - Resolution: **RESOLVED — use `Resolve(entry *parser.Model, entryDir string) (*parser.Model, error)`** (ARCHITECTURE form). The baseDir is mandatory for INC-02; there is no way to resolve relative paths without knowing the including file's directory. This is a Discretion item (CONTEXT flags the struct shape as planner's call) and is resolved here.
2. **Properties conflict detection scope: name/description only, or all fields?**
   - Resolution: **RESOLVED — all non-zero fields** (CONTEXT discretion leans conflict-error for safety). Root-wins for zero-value fields in entry; any non-zero/non-zero overlap hard-errors.
3. **Where do multi-file test fixtures live?**
   - Resolution: **RESOLVED — `internal/include/testdata/`** (per-package fixtures, matching the `internal/parser` convention of `../../testdata/`). Keeps the multi-file fixtures co-located with the package that tests them.

## Environment Availability

> Phase 32 has **no external dependencies** beyond the existing Go toolchain and project deps. SKIPPED (no external tools/services/runtimes to probe).

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go toolchain | build/test | ✓ | go1.26.5 darwin/arm64 | — |
| `path/filepath` stdlib | canonicalization | ✓ | stdlib | — |
| `pelletier/go-toml/v2` | TOML extraction | ✓ | v2.2.4 (go.mod) | — |
| `testify` | tests | ✓ | current (go.mod) | — |

**Missing dependencies with no fallback:** None.
**Missing dependencies with fallback:** None.

## Validation Architecture

> `workflow.nyquist_validation` is enabled (absent = enabled). This section is included.

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go testing (`testing`) + `github.com/stretchr/testify` (require/assert) |
| Config file | none — Go-native (`go test`) |
| Quick run command | `go test ./internal/include/ -run 'TestResolve' -count=1` |
| Full suite command | `go test ./... -count=1` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| INC-01 | entry + transitively-included files merge into one logical model | unit | `go test ./internal/include/ -run TestResolveTwoFilesMerge -count=1` | ❌ Wave 0 (new `internal/include/include_test.go`) |
| INC-02 | paths resolve relative to including file's dir (cd-proof) | unit | `go test ./internal/include/ -run TestResolveRelativePathIndependentOfCwd -count=1` | ❌ Wave 0 |
| INC-03 | transitive includes resolve recursively | unit | `go test ./internal/include/ -run TestResolveTransitive -count=1` | ❌ Wave 0 |
| INC-04 | cycle (self + mutual) fatal, names the cycle | unit | `go test ./internal/include/ -run TestResolveCycleFatal -count=1` | ❌ Wave 0 |
| INC-05 | diamond not a cycle; no-`once` dup-path hard-errors | unit | `go test ./internal/include/ -run TestResolveDiamondNotCycle -count=1` | ❌ Wave 0 |
| INC-06 | `once=true` skips re-inclusion | unit | `go test ./internal/include/ -run TestResolveOnceDedup -count=1` | ❌ Wave 0 |
| INC-07 | flat merge; cross-file dup unit path hard-errors naming both files | unit | `go test ./internal/include/ -run TestMergeDuplicateUnitPathError -count=1` | ❌ Wave 0 |
| INC-08 | properties root-wins; conflict hard-errors | unit | `go test ./internal/include/ -run TestMergePropertiesConflictError -count=1` | ❌ Wave 0 |
| INC-09 | UnitOrder concatenation preserves authoring order | unit | `go test ./internal/include/ -run TestMergeUnitOrderAppend -count=1` | ❌ Wave 0 |
| INC-10 | missing include names referenced path + including file | unit | `go test ./internal/include/ -run TestResolveMissingIncludeError -count=1` | ❌ Wave 0 |
| XC-02 | templates in included file visible to `[[use]]` (integration w/ Phase 31) | integration | `go test ./cmd/c4drill/ -run TestIncludeTemplateIsolation -count=1` (gated on Phase 31) | ❌ Wave 0 (deferred to post-Phase-31 if needed) |
| (cross-cutting) multi-file model renders same as single-file equivalent (canonicalDOT) | integration/golden | `go test ./... -run TestMultiFileGoldenCanonicalDOT -count=1` | ❌ Wave 0 |

### Sampling Rate

- **Per task commit:** `go test ./internal/include/ ./internal/parser/ -count=1` (the two touched packages; sub-second)
- **Per wave merge:** `go test ./... -count=1` (full suite; ensures no regression in validator/view/render)
- **Phase gate:** full suite green + `golangci-lint run ./...` clean before `/gsd:verify-work`

### Wave 0 Gaps

- [ ] `internal/include/include_test.go` — covers INC-01..10 (table-driven Resolve tests)
- [ ] `internal/include/testdata/` — multi-file fixtures: `main.toml`, `auth.toml`, `templates.toml`, `cycle_a.toml`/`cycle_b.toml`, `diamond_*`, `missing_ref.toml`
- [ ] `internal/parser/parser_test.go` — add `TestParseIncludesExtracted` (the `Model.Includes` extraction; consumes the Phase 31 skip)
- [ ] (If no canonicalDOT helper exists yet) a shared order-insensitive comparator helper for the golden test — check `internal/graph` or test helpers first; STATE.md DI-1 says this is established practice, so it likely exists. Reuse, do not reinvent.

*(If existing canonicalDOT comparator found: "Reuse existing helper — do not reinvent (STATE.md DI-1).")*

## Security Domain

> `security_enforcement` is enabled (absent = enabled), ASVS L1, block-on-high. Phase 32 is a local-file TOML composition feature with **no network, no auth, no untrusted input path** (C4Drill is a single-shot local CLI; input files are author-controlled, not attacker-controlled). The threat surface is minimal but non-zero: include paths are file paths read from TOML.

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|------------------|
| V2 Authentication | no | N/A — local CLI, no auth |
| V3 Session Management | no | N/A — single-shot, no sessions |
| V4 Access Control | no | N/A — runs as the invoking user with the user's filesystem permissions; no privilege boundary |
| V5 Input Validation | yes | Include paths are validated via `filepath.Clean`/`filepath.Abs` (canonicalization); missing files hard-error (INC-10); cycle/depth caps prevent resource exhaustion |
| V6 Cryptography | no | N/A — no crypto |
| V8 Data Protection | no | N/A — no sensitive data handled |
| V12 Files & Resources | yes | Include resolution reads files from disk; bounded by max-depth cap (100) and cycle detection to prevent pathological resource consumption; no file writes |

### Known Threat Patterns for Go CLI file composition

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Path traversal via crafted include path (e.g. `../../../../etc/passwd`) | Information Disclosure | C4Drill is a local rendering tool operating on author-controlled TOML; the author already has full filesystem access as the invoking user, so traversal grants no new capability. `filepath.Clean`+`Abs` canonicalize for cycle/dedup correctness, not as a security boundary. Document this is not a security boundary. |
| Zip-bomb / diamond-explosion (acyclic but exponentially-growing include graph) | Denial of Service | `maxIncludeDepth = 100` cap (defense-in-depth, matches go-metadot precedent); visited-set dedup ensures any single file is parsed at most once regardless of how many paths reach it (D-11 same-file diamond dedup is automatic; `once` is opt-in). The visited-set is the primary DoS mitigation — it bounds total work to O(number of distinct files). |
| Symlink-based cycle evasion | Tampering/DoS | `filepath.Abs` does NOT resolve symlinks (only `filepath.EvalSymlinks` does). A symlink loop could evade the stack/visited check. **Mitigation:** acceptable for v1.10 — C4Drill is author-controlled local tooling; document the limitation. If hardening is desired later, `filepath.EvalSymlinks` is the upgrade path. Note in plan as an accepted limitation, not a blocker. |

**ASVS L1 verdict:** No high-severity threats. The feature adds no new privilege boundary, no network, no untrusted-input path. The DoS mitigations (depth cap + visited-set) are defense-in-depth. No `security_block_on: high` trigger.

## Sources

### Primary (HIGH confidence)
- **Codebase (verified 2026-08-08):** `internal/parser/parser.go` (Model:35-42, Parse:47-96, captureDefinitionOrder:100-157 incl. properties skip:128, ParseFile:323-336), `internal/parser/errors.go` (ParseError:13, wrapDecodeError:45), `internal/model/properties.go` (Properties:4-21), `internal/model/unit.go` (Subunits:71-73, SubunitOrder:70, Expanded:64-65), `cmd/c4drill/root.go` (runRoot pipeline:112-149, collectExpandedPaths:156)
- **Milestone research (2026-08-08):** `.planning/research/SUMMARY.md` (§2 zero-deps, §3 feature classification, §4 pipeline+Model-extension, §6 IN-2/IN-3 forks, §8 phase breakdown, §9 watch-outs), `.planning/research/ARCHITECTURE-v1.10.md` (extended pipeline:86, insertion:90-122, Model extension:147-182, include package:223-265), `.planning/research/PITFALLS.md` (IN-1..IN-5, BC-3, testing implications table)
- **Phase 31 Plan 1 (2026-08-08):** `.planning/phases/31-template-expansion/31-01-PLAN.md` — confirms the BC-1 reserved-table skip for `[template.*]`/`[[use]]`/`[[include]]` + `Model.Templates`/`Instantiations` fields land there; `[[include]]` is skipped but NOT extracted in Phase 31 (reserved for Phase 32)
- **Go stdlib:** `path/filepath` (`Abs`, `Clean`, `Join`, `Dir`) — `https://pkg.go.dev/path/filepath` (Abs calls Clean on result)

### Secondary (MEDIUM confidence)
- **Original feature todo:** `.planning/todos/pending/2026-08-08-include-directive-multi-file-diagrams.md` (Option B recommendation, merge semantics table, cycle/once/depth design, pipeline ordering, verification criteria)
- **STATE.md decision log:** DI-1 (canonicalDOT order-insensitive), D-12 (validator as gatekeeper), BC-3 (non-strict toml load-bearing)

### Tertiary (LOW confidence)
- None — all findings verified against codebase or milestone research.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — all stdlib + existing project deps, verified in `go.mod` and codebase
- Architecture: HIGH — integration points confirmed by reading live `parser.go`/`root.go`/`errors.go`; ARCHITECTURE-v1.10.md file:line citations match current code
- Pitfalls: HIGH — all pitfalls sourced from milestone PITFALLS.md (HIGH confidence per its own assessment) and verified against codebase
- Concurrency: HIGH — confirmed Phase 31 Plan 1 owns the BC-1 skip; Phase 32 must not re-touch `captureDefinitionOrder` (avoids conflict with concurrent Phase 28/31 work)

**Research date:** 2026-08-08
**Valid until:** 2026-09-07 (30 days; stable Go/stdlib, no fast-moving deps)
