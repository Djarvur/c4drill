# Phase 26: Preserve TOML Definition Order - Research

**Researched:** 2026-03-25
**Domain:** TOML parsing with order preservation, Go data structures for ordered maps
**Confidence:** HIGH

## Summary

Phase 26 aims to preserve the exact order units and links are defined in the TOML file, replacing Phase 23's alphabetical sorting with definition-order preservation. This requires fundamental changes to the parser and model data structures.

**Current state (Phase 23):** Units and edges appear in deterministic alphabetical order via `slices.Sorted(maps.Keys())`. This is reliable but not intuitive - users expect units to appear in the order they defined them.

**Target state:** Units and edges appear in the exact order they are defined in the TOML file. If a user defines `[zulu]` before `[alpha]`, `zulu` appears first in the diagram.

**Primary recommendation:** Use go-toml's `unstable` parser API to capture definition order during parsing, then propagate this order through the pipeline by replacing `map[string]*Unit` with an ordered data structure (slice with name field, or ordered map library).

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| github.com/pelletier/go-toml/v2 | v2.2.4 | TOML parsing | Already in use, has unstable AST API |
| github.com/pelletier/go-toml/v2/unstable | v2.2.4 | Low-level parsing with order | Built into go-toml, provides iterative AST |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| slices (stdlib) | Go 1.21+ | Sorted iteration | Already used, pattern continues |
| maps (stdlib) | Go 1.21+ | Key extraction | Already used |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Custom slice-based ordered structure | github.com/elliotchance/orderedmap | Custom is simpler, no new dependency |
| go-toml unstable API | Custom parser | Unstable API is well-maintained and purpose-built |
| Replacing all maps | Ordered map wrapper | Wrapper adds complexity; targeted changes are cleaner |

**Installation:**
No new packages required - go-toml v2.2.4 already includes the `unstable` subpackage.

## Architecture Patterns

### Current Data Structure (Problem)

```go
// model/unit.go - Unit.Subunits uses map (order lost)
type Unit struct {
    Subunits map[string]*Unit `toml:",inline"`  // Go map = random iteration
}

// parser/parser.go - Model.Units uses map (order lost)
type Model struct {
    Units map[string]*model.Unit  // Go map = random iteration
}
```

**Root cause:** Go maps do not preserve insertion order. Even if go-toml parses in definition order, unmarshaling into `map[string]*T` loses that order immediately.

### Recommended Pattern: Slice-Based Ordered Structure

**Option A: Slice with inline name (simpler, recommended)**

```go
// model/unit.go - New ordered unit type
type Unit struct {
    // ... existing fields ...
    Subunits []NamedUnit `toml:",inline"`  // Ordered slice
}

type NamedUnit struct {
    Name string  // Not a TOML field - the section key
    Unit *Unit   // The actual unit
}
```

**Option B: Separate order tracking (less invasive)**

```go
// parser/parser.go - Track order separately
type Model struct {
    UnitOrder []string            // Definition order
    Units     map[string]*model.Unit
}
```

### Pattern: Using go-toml unstable API

The `unstable` package provides iterative AST parsing that preserves definition order:

```go
import "github.com/pelletier/go-toml/v2/unstable"

// Source: https://pkg.go.dev/github.com/pelletier/go-toml/v2/unstable
doc := `
hello = "world"
value = 42
`
p := unstable.Parser{}
p.Reset([]byte(doc))
for p.NextExpression() {
    e := p.Expression()
    // e.Kind tells us: KeyValue, Table, ArrayTable
    // Process in the order they appear in the document
}
```

### Anti-Patterns to Avoid

- **Using `map[string]*T` for ordered data:** Go maps iterate randomly by design
- **Sorting after parsing:** Defeats the purpose - use slices from the start
- **Multiple order tracking mechanisms:** Keep it simple - one source of truth for order

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| TOML parsing with order | Custom TOML tokenizer | go-toml unstable API | Handles edge cases, errors, encoding |
| Ordered map | Custom linked-list map | Slice with order index | Simpler, no pointer chasing |
| Iterating ordered data | Sort on every access | Store in order once | Performance and correctness |

**Key insight:** The unstable API exists specifically for cases like this where you need document-order access. Don't fight the library.

## Common Pitfalls

### Pitfall 1: Assuming maps preserve order
**What goes wrong:** Code iterates `for k, v := range m` expecting definition order
**Why it happens:** Some languages (Python 3.7+, Ruby) preserve map order; Go explicitly does not
**How to avoid:** Use slices for ordered data, maps only for lookup
**Warning signs:** Code that "sometimes works" depending on map size/hash

### Pitfall 2: Breaking existing map-based lookups
**What goes wrong:** Changing `Unit.Subunits` from map to slice breaks all `unit.Subunits[name]` access patterns
**Why it happens:** Significant API change across codebase
**How to avoid:**
  - Option A: Add helper method `GetSubunit(name string) *Unit`
  - Option B: Keep both map and slice (map for lookup, slice for order)
  - Option C: Use ordered map library with `Get()` method

### Pitfall 3: Nested subunit order
**What goes wrong:** Top-level units are ordered but nested `[parent.child]` units are not
**Why it happens:** Parser recursively processes units; each level needs order tracking
**How to avoid:** Ensure order tracking is recursive through all nesting levels

### Pitfall 4: Forgetting link order
**What goes wrong:** Units appear in definition order but `[[link]]` arrays don't
**Why it happens:** Links are already slices (`[]Link`) - they preserve order
**How to avoid:** Verify link order is preserved (it should be - already slices)

## Code Examples

### Current Parser Approach (loses order)

```go
// internal/parser/parser.go (current)
func Parse(data []byte) (*Model, error) {
    var rawMap map[string]any
    if err := toml.Unmarshal(data, &rawMap); err != nil {
        return nil, wrapDecodeError(err)
    }

    m := &Model{
        Units: make(map[string]*model.Unit),  // ORDER LOST HERE
    }

    for name, value := range rawMap {  // Random iteration
        if name == "properties" {
            continue
        }
        unit, err := parseUnit(name, value, "")
        m.Units[name] = unit
    }
    return m, nil
}
```

### Recommended: Using unstable Parser API

```go
// internal/parser/parser.go (new approach)
import "github.com/pelletier/go-toml/v2/unstable"

type Model struct {
    Properties model.Properties
    UnitOrder  []string           // Definition order
    Units      map[string]*model.Unit
}

func Parse(data []byte) (*Model, error) {
    m := &Model{
        UnitOrder: make([]string, 0),
        Units:     make(map[string]*model.Unit),
    }

    p := unstable.Parser{}
    p.Reset(data)

    for p.NextExpression() {
        expr := p.Expression()
        if expr.Kind == unstable.Table {
            // Extract table name and process
            keyIter := expr.Key()
            var keyParts []string
            for keyIter.Next() {
                keyParts = append(keyParts, string(keyIter.Node().Data))
            }

            // Skip [properties]
            if len(keyParts) == 1 && keyParts[0] == "properties" {
                continue
            }

            // Process unit in definition order
            name := keyParts[0]
            m.UnitOrder = append(m.UnitOrder, name)
            // ... parse unit content ...
        }
    }

    if p.Error() != nil {
        return nil, wrapDecodeError(p.Error())
    }

    return m, nil
}
```

### Propagating Order Through View Generation

```go
// internal/view/scope.go (modified)
func GenerateC1View(m *parser.Model) *View {
    v := &View{
        Units: make(map[string]*Entry),
    }

    // Use definition order from parser
    for _, name := range m.UnitOrder {
        unit := m.Units[name]
        v.Units[name] = &Entry{
            Unit:     unit,
            FullPath: name,
            // ...
        }
    }

    return v
}
```

### Modified builder.go (already uses sorted keys)

```go
// internal/graph/builder.go (current Phase 23 approach)
keys := slices.Sorted(maps.Keys(v.Units))  // Alphabetical

// Would change to:
// Use view's internal order tracking instead of sorting
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Random map order | Alphabetical sort (Phase 23) | 2026-03-25 | Deterministic but not intuitive |
| Alphabetical sort | Definition order (Phase 26) | This phase | Intuitive, matches user expectation |

**Deprecated/outdated:**
- `toml.Unmarshal` into map for order-sensitive data: Loses definition order immediately

## Open Questions

1. **Should Subunits also use definition order?**
   - What we know: Top-level units are defined as `[name]`, subunits as `[parent.child]`
   - What's unclear: Does the unstable API process `[parent.child]` as nested or flat?
   - Recommendation: Test with nested TOML; if flat, track parent-child relationships during parsing

2. **Should we add an OrderedUnit type or modify Unit?**
   - What we know: Unit.Subunits is `map[string]*Unit`
   - What's unclear: Breaking change scope vs. adding complexity
   - Recommendation: Track order in parser.Model only; keep Unit.Subunits as map for backward compatibility

3. **How does this interact with external boundary nodes?**
   - What we know: External nodes are auto-generated for missing link targets
   - What's unclear: Where should they appear in order?
   - Recommendation: Append external nodes at the end of the order list

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | testing + testify v1.11.1 |
| Config file | None - standard Go test pattern |
| Quick run command | `go test ./internal/parser/... -v` |
| Full suite command | `go test ./...` |

### Phase Requirements -> Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| N/A | Parse preserves unit definition order | unit | `go test ./internal/parser/... -run TestParseDefinitionOrder -v` | No - Wave 0 |
| N/A | View generates units in definition order | unit | `go test ./internal/view/... -run TestViewDefinitionOrder -v` | No - Wave 0 |
| N/A | Graph builds nodes in definition order | unit | `go test ./internal/graph/... -run TestBuildGraphDefinitionOrder -v` | No - Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./internal/parser/... ./internal/view/... ./internal/graph/... -v`
- **Per wave merge:** `go test ./...`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `internal/parser/parser_test.go` - TestParseDefinitionOrder (units appear in TOML order)
- [ ] `internal/view/scope_test.go` - TestViewDefinitionOrder (C1 view preserves order)
- [ ] `internal/graph/builder_test.go` - TestBuildGraphDefinitionOrder (nodes in definition order, not alphabetical)

## Sources

### Primary (HIGH confidence)
- [go-toml unstable package](https://pkg.go.dev/github.com/pelletier/go-toml/v2/unstable) - Iterative AST parsing with order preservation
- [go-toml GitHub](https://github.com/pelletier/go-toml) - Library documentation and examples

### Secondary (MEDIUM confidence)
- Project codebase analysis: internal/parser/parser.go, internal/model/unit.go, internal/graph/builder.go

### Tertiary (LOW confidence)
- None - all findings verified against primary sources

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - Using existing go-toml library with unstable API
- Architecture: HIGH - Clear pattern from unstable API docs
- Pitfalls: HIGH - Common Go map behavior, well-documented

**Research date:** 2026-03-25
**Valid until:** 30 days - go-toml API is stable, but unstable API could change (hence the name)
