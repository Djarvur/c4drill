# Phase 30: Relative-peer resolution - Pattern Map

**Mapped:** 2026-08-08
**Source:** CONTEXT.md (D-13/D-14/D-15/D-16) + RESEARCH.md (Architecture Patterns, Code Examples)

Maps every file to be created/modified to its closest existing analog in the codebase, with concrete line references the executor copies from.

## File Classification

| File | Action | Role | Data Flow | Analog |
|------|--------|------|-----------|--------|
| `internal/peer/resolve.go` | NEW | service/transform | transform (Model→Model, in-place rewrite) | `internal/validator/index.go` `BuildIndex` (recursion shape) + `internal/parser/parser.go` `Parse` (transform signature) |
| `internal/peer/resolve_test.go` | NEW | test | unit + integration | `internal/parser/parser_test.go` (testify idiom + fixture-driven) |
| `cmd/c4drill/root.go` | MODIFY | pipeline orchestration | request-response (CLI) | itself — Parse/Validate staging at `:112-118` |
| `cmd/c4drill/root_test.go` | NEW (or extend existing) | integration test | integration | `internal/parser/parser_test.go` + `cmd/c4drill/testdata/*` |
| `cmd/c4drill/testdata/peer_walkup.toml` | NEW | fixture | file-I/O | `cmd/c4drill/testdata/multilevel.toml` (multi-depth unit tree) |
| `cmd/c4drill/testdata/peer_unresolvable.toml` | NEW | fixture | file-I/O | `cmd/c4drill/testdata/invalid.toml` (error-case fixture) |

## Pattern Assignments

### File: `internal/peer/resolve.go` (NEW)

**Primary analog — `validator.BuildIndex` recursion (`internal/validator/index.go:24-43`):**

This is THE closest analog. `BuildIndex` does exactly what `Resolve` must do for its outer walk: recursively visit every unit + subunit, constructing dotted `fullPath`s. `Resolve` mirrors the recursion and the path construction, then rewrites `Link.Peer` at each leaf instead of building an index entry.

Copy-from excerpt (current):
```go
// internal/validator/index.go:24-43
func BuildIndex(units map[string]*model.Unit, parentPath string) map[string]*UnitInfo {
    index := make(map[string]*UnitInfo)
    for name, unit := range units {
        fullPath := name
        if parentPath != "" {
            fullPath = parentPath + "." + name
        }
        index[fullPath] = &UnitInfo{Unit: unit, FullPath: fullPath, Parent: parentPath}
        if len(unit.Subunits) > 0 {
            subIndex := BuildIndex(unit.Subunits, fullPath)
            maps.Copy(index, subIndex)
        }
    }
    return index
}
```

Target shape for `resolve.go` — same recursion, same path math, different leaf action (RESEARCH.md Pattern 2):
```go
func resolveUnits(units map[string]*model.Unit, parentPath string, m *parser.Model) error {
    for name, unit := range units {
        fullPath := name
        if parentPath != "" { fullPath = parentPath + "." + name }
        if err := resolveUnitLinks(unit, fullPath, m); err != nil { return err }
        if len(unit.Subunits) > 0 {
            if err := resolveUnits(unit.Subunits, fullPath, m); err != nil { return err }
        }
    }
    return nil
}
```

The executor reads `internal/validator/index.go:1-43` (full file, ~90 LOC) to copy the recursion shape exactly.

**Secondary analog — `parser.Parse` transform signature (`internal/parser/parser.go:47-96`):**

`Parse` is the project's canonical "transform input → `*parser.Model`" function. `Resolve` follows the same signature family: takes `*parser.Model`, returns `error`. (Resolve mutates in place and returns the same model pointer conceptually; the signature is `func Resolve(m *parser.Model) error`.)

Copy-from excerpt (current):
```go
// internal/parser/parser.go:47
func Parse(data []byte) (*Model, error) {
```

Target:
```go
// internal/peer/resolve.go
package peer

import (
    "errors"
    "strings"

    "github.com/Djarvur/c4drill/internal/parser"
)

// Resolve rewrites every Link.Peer (in every Unit.Links and Unit.LinksFrom,
// including all subunits) from a relative bare name to an absolute dotted path
// per D-13/D-14/D-15/D-16. Peers containing "." are untouched (absolute).
// Returns an error naming the peer and host unit if a bare peer cannot be resolved.
func Resolve(m *parser.Model) error { ... }
```

**Walk-up scope enumeration (the D-13/D-14/D-15 algorithm):**

There is no direct analog in the codebase — the walk-up is new logic. But the mechanism is a simple loop over `strings.Split(hostPath, ".")` progressively dropping the last segment, plus a final root iteration over `m.Units`. The executor reads RESEARCH.md Pattern 1 for the pseudocode. Key invariants:
- `ancestorScopes("a.b.c")` returns scopes in nearest-first order: the children-maps of `a.b`, then `a`, then root (`m.Units`). It does NOT include `a.b.c` itself (D-13: start at immediate parent).
- `ancestorScopes("topLevel")` returns just `[root]` (the children-map is `m.Units`).
- For each scope, check `children[peer]` — map lookup is 0-or-1 (unique keys), so the "same-depth multi-match" error branch (criterion 3) is dead code but should be authored defensively.

**Error type — analog `*parser.ParseError` (`internal/parser/errors.go`):**

The resolver's error should name the peer and the host unit. Two viable idioms:
- Define a small `*ResolveError` struct with `Peer`, `Host` fields and an `Error()` method (cleanest for typed inspection in tests).
- Reuse `fmt.Errorf("cannot resolve peer %q from unit %q", peer, host)` (simplest; tests match on substring).

Recommendation: define `*ResolveError` (mirrors `*parser.ParseError`'s struct-with-fields shape) so tests can assert on the named fields and future code can `errors.As` it. The executor reads `internal/parser/errors.go` (full file, ~80 LOC) for the struct-with-`Error()` idiom.

### File: `internal/peer/resolve_test.go` (NEW)

**Analog — `internal/parser/parser_test.go` (the testify + fixture-driven idiom):**

Copy-from excerpt (current test idiom):
```go
// internal/parser/parser_test.go:1-13
package parser_test

import (
    "os"
    "testing"

    "github.com/Djarvur/c4drill/internal/model"
    "github.com/Djarvur/c4drill/internal/parser"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)
```

Target for `resolve_test.go`:
```go
// internal/peer/resolve_test.go
package peer_test

import (
    "errors"
    "os"
    "testing"

    "github.com/Djarvur/c4drill/internal/parser"
    "github.com/Djarvur/c4drill/internal/peer"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)
```

**Test shape — fixture-driven (`internal/parser/parser_test.go:18` TestParseValidProperties):**

The parser tests read fixtures via `os.ReadFile("../../testdata/<fix>")` (the `../../testdata/` path from `internal/parser/` resolves to the repo-ROOT `/testdata/` dir — NOT `cmd/c4drill/testdata/`). Resolve tests follow the same shape: read fixture → `parser.Parse` → `peer.Resolve` → assert on rewritten `Link.Peer` values via a tree walk (mirroring `BuildIndex` to collect all peers).

For small cases (sibling, aunt, root, nearest-first), prefer **inline TOML literals** parsed via `parser.Parse([]byte(\`...\`))` — sharper than fixtures, no file I/O. For the corpus backward-compat test (ERGO-02 acceptance), use the ROOT `testdata/` corpus (`valid.toml`, `links.toml`, `nested.toml` — do NOT include `invalid_*.toml`); the `cmd/c4drill/testdata/` corpus is covered end-to-end by Plan 02. NOTE: there are TWO `valid.toml` files (root vs cmd/c4drill) with DIFFERENT content — the unit test in `internal/peer/` reads the ROOT one via `../../testdata/valid.toml`.

The HS-2 forward-compatibility contract (resolve treats templated units uniformly) is NOT testable in Phase 30 (templates don't exist yet). The plan should note this: the corpus test + the uniform tree-walk (visits all units including any future templated ones) together satisfy HS-2 by construction. No dedicated HS-2 test in Phase 30.

### File: `cmd/c4drill/root.go` (MODIFY)

**Analog — itself: the Parse → Validate staging at `root.go:111-123`.**

Copy-from excerpt (current):
```go
// cmd/c4drill/root.go:111-123
// Stage 1: Parse
m, err := parser.ParseFile(inputPath)
if err != nil {
    return fmt.Errorf("parse: %w", err)
}

// Stage 2: Validate
valErrors := validator.Validate(m)
```

Target: insert between these two stages:
```go
// Stage 1: Parse
m, err := parser.ParseFile(inputPath)
if err != nil {
    return fmt.Errorf("parse: %w", err)
}

// Stage 1.6: Resolve relative peers (Phase 30; runs after future template.Expand, before humanize + Validate)
if err := peer.Resolve(m); err != nil {
    return fmt.Errorf("resolve peers: %w", err)
}

// Stage 2: Validate
valErrors := validator.Validate(m)
```

**Concurrency note (from the invoking prompt):** A background agent is concurrently executing Phase 29 (humanize), which edits `internal/parser/parser.go` and `internal/model/unit.go` (adds Humanize). Phase 29's humanize call will ALSO insert into this same gap (after peer.Resolve, per pipeline order XC-04). This is a future insert point — Phase 30 only adds the `peer.Resolve` call shown above. The executor reads `cmd/c4drill/root.go:95-150` (full staging block) to place the call correctly. Phase 30 does NOT touch parser.go or unit.go (read-only for pattern-mapping), so there is no merge conflict with Phase 29.

**Import addition:** add `"github.com/Djarvur/c4drill/internal/peer"` to the import block. The executor reads the existing import block at `cmd/c4drill/root.go:1-20` to place it alphabetically/conventionally.

### File: `cmd/c4drill/root_test.go` (NEW or extend)

**Analog — `internal/parser/parser_test.go` integration-test shape + `cmd/c4drill/testdata/*` fixture consumption.**

The pipeline test (30-02-01) must: (1) parse a fixture with a bare peer, (2) call `peer.Resolve`, (3) call `validator.Validate`, (4) assert validation passes (proving the validator saw the rewritten absolute peer). The CLI error test (30-02-02) must invoke the `runRoot` path (or `Execute`) on an unresolvable-peer fixture and assert non-zero exit + error message. The executor checks whether `cmd/c4drill/root_test.go` exists; if not, create it with `package c4drill_test` (or the existing cobra-based test idiom if present).

### File: `cmd/c4drill/testdata/peer_walkup.toml` (NEW)

**Analog — `cmd/c4drill/testdata/multilevel.toml` (multi-depth unit tree).**

`multilevel.toml` is the canonical multi-depth fixture (units nested 3-4 levels deep with absolute peers). The new fixture mirrors its structure but uses BARE peers to exercise the walk-up. The executor reads `cmd/c4drill/testdata/multilevel.toml` (full file) for the `[properties]` / `[unit]` / `[[unit.link]]` / `[unit.subunit.subsubunit]` TOML shape.

Required cases (one fixture, multiple units covering all):
- Sibling match: `[scope.parent]` with children `host` and `target`; `host` declares `peer = "target"` → resolves to `scope.parent.target`.
- Aunt match: `[scope.parent.host]` declares `peer = "aunt"` where `scope.aunt` exists → walk-up resolves to `scope.aunt`.
- Root match: `[deeply.nested.unit]` declares `peer = "topService"` where top-level `[topService]` exists → resolves to `topService`.
- Nearest-first: `[a.b.c]` declares `peer = "x"` where BOTH `a.b.x` and `a.x` exist → resolves to `a.b.x` (nearer), no error.

### File: `cmd/c4drill/testdata/peer_unresolvable.toml` (NEW)

**Analog — `cmd/c4drill/testdata/invalid.toml` (error-case fixture).**

`invalid.toml` is the existing fixture for validator-error cases. The new fixture mirrors its shape: a minimal model with a bare peer that matches no sibling, aunt, or top-level unit. Used by `TestResolveUnresolvableError` (30-01-06) and the CLI error test (30-02-02).

## Shared Patterns

### Error handling / reporting
- Parser errors: `*parser.ParseError` (`internal/parser/errors.go`) — struct with `Message`/`Line`/`Context`/`Cause` + `Error()` method. Resolve's `*ResolveError` follows this shape (struct with `Peer`/`Host` + `Error()`).
- CLI error wrapping: `fmt.Errorf("stage: %w", err)` (`cmd/c4drill/root.go:114,117`). Resolve's error wraps identically: `fmt.Errorf("resolve peers: %w", err)`.

### Non-strict TOML (load-bearing)
go-toml/v2 stays non-strict (STATE.md DI-1 / BC-3). Phase 30 does NOT touch TOML decoding, so this is a non-issue here — but the resolver must not introduce any re-marshaling that would enable strict mode. It reads the parsed `*parser.Model` directly.

### testify test idiom (all test files)
`t.Parallel()` + `require.NoError` for preconditions + `assert.Equal`/`require.Len`/`require.True` for assertions. Fixtures via `os.ReadFile("../../testdata/...")`. All new test files follow this (parser_test.go is the reference).

### Order preservation (load-bearing)
`captureDefinitionOrder` (`parser.go:100`) exists to preserve authoring order. Resolve does NOT touch `UnitOrder` or `SubunitOrder` — it only rewrites `Link.Peer` strings in place. The executor must NOT reorder units, links, or any other structure.

### Pipeline ordering (load-bearing, XC-01)
Resolve MUST run after `parser.ParseFile` and after future `template.Expand` (Phase 31), and before `humanize` (Phase 29) and `validator.Validate`. The `cmd/c4drill/root.go` insertion point (between Parse and Validate) is correct for Phase 30's single-call scope. Phase 31 will insert `template.Expand` ABOVE the `peer.Resolve` line; Phase 29 will insert humanize BELOW it.
