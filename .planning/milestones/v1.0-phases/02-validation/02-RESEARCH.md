# Phase 2: Validation - Research

**Researched:** 2026-03-09
**Domain:** Go model validation, error reporting, Levenshtein distance for suggestions
**Confidence:** MEDIUM (web search rate-limited, relied on package cache and training knowledge)

## Summary

Phase 2 implements model validation after TOML parsing succeeds. The validator traverses the parsed `Model` to check reference integrity (all linked units exist), type rules (only system/box can have subunits), and link constraints (units with subunits cannot have direct links). Error messages follow a plain single-line format with optional "did you mean" suggestions for typos. All errors are collected before reporting, allowing users to fix multiple issues in one pass.

**Primary recommendation:** Create a separate `internal/validator` package with a `Validate(*parser.Model) []ValidationError` function. Reuse the existing `ParseError` pattern from `internal/parser/errors.go` but create a distinct `ValidationError` type. Use `github.com/agnivade/levenshtein` for spell-check suggestions.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**Error output format:**
- Plain single-line format (not multi-line structured, not JSON)
- Prefix style: `error: <message>`
- Example: `error: undefined unit "db1" referenced from "api" at line 15`

**Validation behavior:**
- Collect all errors before reporting (not fail-fast)
- Report in file order (top-to-bottom traversal)
- Exit code 1 on any error, 0 on success (standard CLI convention)

**Error detail level:**
- Include suggestions when possible: `did you mean "db"?` for similar names
- Use concise wording: `undefined unit "x"` not `unit "x" is not defined`
- Best-effort line numbers: use go-toml position when available, omit otherwise
- Full path for nested units: `mainapp.api.handler` not just `handler`
- Show error count summary at end: `5 errors found`

### Claude's Discretion

- Exact algorithm for "did you mean" suggestions (Levenshtein distance threshold)
- Maximum errors to collect before stopping (if any limit)
- Exact wording for each error type (follow concise pattern)

### Deferred Ideas (OUT OF SCOPE)

None - discussion stayed within phase scope.

</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| VALD-01 | Validator checks all referenced units are defined | Unit index building, recursive traversal |
| VALD-02 | Validator prevents links on units that have subunits | Subunit detection, link presence check |
| VALD-03 | Validator prevents referencing units that have subunits | Target unit subunit check |
| VALD-04 | Validator prevents subunits on non-system/non-box types | Type rules table below |
| VALD-05 | Error messages include line numbers and context | Reuse ParseError pattern, best-effort |
| VALD-06 | Error messages use human-readable format (not JSON) | Plain single-line format |
| QUAL-01 | All lint errors must be fixed before commit | golangci-lint v2 enforcement |
| QUAL-02 | Lint config MUST NOT be adjusted to silence errors | Existing .golangci.yml |
| QUAL-03 | nolint directives require explicit user confirmation | Process requirement |
| QUAL-04 | Minimum 75% test coverage required | go test -cover |
| QUAL-05 | Coverage enforced in CI/quality gate | Process requirement |

</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| github.com/agnivade/levenshtein | v1.2.1 | Spell-check suggestions | Fast, simple API, pure Go |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| stretchr/testify | v1.11.1 | Testing assertions | Already in project, use assert/require |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| agnivade/levenshtein | texttheater/golang-levenshtein | Both work; agnivade is faster with same API |
| external Levenshtein | Hand-rolled algorithm | Not worth it for 5-line function; library handles edge cases |

**Installation:**
```bash
go get github.com/agnivade/levenshtein@v1.2.1
```

## Architecture Patterns

### Recommended Project Structure
```
internal/
├── validator/
│   ├── validator.go       # Validate function, entry point
│   ├── validator_test.go  # Test cases for each validation rule
│   ├── errors.go          # ValidationError type
│   └── rules.go           # Individual validation rule functions
├── parser/                # (existing) provides Model to validate
└── model/                 # (existing) provides Unit, Link types
```

### Pattern 1: Error Collection Pattern
**What:** Collect all validation errors in a slice, return at end
**When to use:** When users benefit from seeing all issues at once
**Example:**
```go
// Source: Standard Go validation pattern
package validator

type ValidationError struct {
    Message string
    Path    string // Full unit path, e.g., "mainapp.api.handler"
    Line    int    // Best-effort, 0 if unknown
}

func (e *ValidationError) Error() string {
    if e.Line > 0 {
        return fmt.Sprintf("error: %s at line %d", e.Message, e.Line)
    }
    return "error: " + e.Message
}

type ValidationErrors []*ValidationError

func (ve ValidationErrors) Error() string {
    if len(ve) == 0 {
        return "no errors"
    }
    return fmt.Sprintf("%d validation error(s)", len(ve))
}

func Validate(m *parser.Model) ValidationErrors {
    var errors ValidationErrors
    errors = append(errors, validateReferences(m)...)
    errors = append(errors, validateSubunitRules(m)...)
    errors = append(errors, validateLinkRules(m)...)
    return errors
}
```

### Pattern 2: Unit Index Building
**What:** Build a flat map of all units with their full paths for O(1) lookup
**When to use:** When validating references across nested structures
**Example:**
```go
// Source: Common tree traversal pattern
package validator

// UnitInfo holds validation metadata for a unit
type UnitInfo struct {
    Unit     *model.Unit
    FullPath string // e.g., "mainapp.api.handler"
    Parent   string // Parent path, empty for top-level
}

// BuildIndex creates a flat map of all units for O(1) lookup
func BuildIndex(units map[string]*model.Unit, parentPath string) map[string]*UnitInfo {
    index := make(map[string]*UnitInfo)
    for name, unit := range units {
        fullPath := name
        if parentPath != "" {
            fullPath = parentPath + "." + name
        }
        index[fullPath] = &UnitInfo{
            Unit:     unit,
            FullPath: fullPath,
            Parent:   parentPath,
        }
        // Recurse into subunits
        if len(unit.Subunits) > 0 {
            subIndex := BuildIndex(unit.Subunits, fullPath)
            for k, v := range subIndex {
                index[k] = v
            }
        }
    }
    return index
}
```

### Pattern 3: Levenshtein Suggestions
**What:** Use edit distance to suggest similar names for undefined references
**When to use:** When reporting "undefined X" errors
**Example:**
```go
// Source: agnivade/levenshtein README
package validator

import "github.com/agnivade/levenshtein"

const maxSuggestionDistance = 2 // Threshold for "did you mean"

// SuggestSimilar returns the closest matching name from candidates
// Returns empty string if no close match found
func SuggestSimilar(typo string, candidates []string) string {
    bestMatch := ""
    bestDistance := maxSuggestionDistance + 1

    for _, candidate := range candidates {
        dist := levenshtein.ComputeDistance(typo, candidate)
        if dist < bestDistance {
            bestDistance = dist
            bestMatch = candidate
        }
    }

    if bestDistance <= maxSuggestionDistance {
        return bestMatch
    }
    return ""
}

// Format suggestion for error message
func formatSuggestion(typo string, candidates []string) string {
    if suggestion := SuggestSimilar(typo, candidates); suggestion != "" {
        return fmt.Sprintf(` (did you mean "%s"?)`, suggestion)
    }
    return ""
}
```

### Pattern 4: Validation Rule Functions
**What:** Separate functions for each validation rule, returning error slices
**When to use:** When rules are independent and testable in isolation
**Example:**
```go
// Source: Standard Go validator pattern
package validator

// Rule: Only system and box types can have subunits (VALD-04)
func validateSubunitTypes(m *parser.Model) ValidationErrors {
    var errors ValidationErrors
    index := BuildIndex(m.Units, "")

    for path, info := range index {
        if len(info.Unit.Subunits) == 0 {
            continue // No subunits, rule doesn't apply
        }

        allowedTypes := []model.UnitType{
            model.TypeSystem, model.TypeBox,
        }
        if !containsType(allowedTypes, info.Unit.Type) {
            errors = append(errors, &ValidationError{
                Message: fmt.Sprintf(`unit "%s" has type %s which cannot have subunits`, path, info.Unit.Type),
                Path:    path,
            })
        }
    }
    return errors
}

// Rule: Units with subunits cannot have direct links (VALD-02)
func validateNoLinksOnParentUnits(m *parser.Model) ValidationErrors {
    var errors ValidationErrors
    index := BuildIndex(m.Units, "")

    for path, info := range index {
        if len(info.Unit.Subunits) == 0 {
            continue // No subunits, rule doesn't apply
        }
        if len(info.Unit.Links) > 0 || len(info.Unit.LinksFrom) > 0 {
            errors = append(errors, &ValidationError{
                Message: fmt.Sprintf(`unit "%s" has subunits and cannot have direct links`, path),
                Path:    path,
            })
        }
    }
    return errors
}
```

### Anti-Patterns to Avoid

- **Fail-fast validation:** Returning on first error prevents users from fixing all issues in one pass. Use error collection instead.
- **JSON error output:** User explicitly chose plain text format. Do not use structured JSON.
- **Multi-line error format:** User chose single-line format. Keep errors on one line each.
- **Ignoring nested units:** Validation must traverse all subunits recursively.
- **Vague error messages:** "Invalid unit" is not helpful. Use "undefined unit X" or "unit X cannot have subunits".

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Edit distance calculation | Custom Levenshtein | agnivade/levenshtein | Handles edge cases, well-tested |
| Error type hierarchy | Complex error types | Simple ValidationError struct | Flat structure sufficient for this use case |
| Unit traversal | Custom tree walker | Recursive BuildIndex function | Simpler, testable separately |

**Key insight:** Validation is conceptually simple but requires careful handling of nested structures. Build a flat index first, then validate.

## Validation Rules Summary

### Type Rules (VALD-04)
| Type | Can Have Subunits | Can Have Links |
|------|-------------------|----------------|
| person, personExternal | No | Yes |
| system, systemExternal | Yes | No (if has subunits) |
| db, dbExternal | No | Yes |
| queue, queueExternal | No | Yes |
| box | Yes | No (if has subunits) |
| container, containerDb, containerQueue | No | Yes |
| component, componentDb, componentQueue | No | Yes |

### Link Rules (VALD-02, VALD-03)
| Rule | Description |
|------|-------------|
| VALD-02 | A unit with subunits cannot have Links or LinksFrom |
| VALD-03 | A link cannot target a unit that has subunits |

### Reference Rules (VALD-01)
| Link Type | What Must Exist |
|-----------|-----------------|
| Links[target] | Unit named "target" must exist (at any level) |
| LinksFrom[source] | Unit named "source" must exist (at any level) |

## Common Pitfalls

### Pitfall 1: Inconsistent Path Formatting
**What goes wrong:** Using "handler" vs "mainapp.api.handler" inconsistently
**Why it happens:** Easy to forget to build full paths during traversal
**How to avoid:** Always build full paths with parent prefix in BuildIndex
**Warning signs:** "undefined unit" errors for units that exist

### Pitfall 2: Forgetting LinksFrom Validation
**What goes wrong:** Only validating Links map, ignoring LinksFrom
**Why it happens:** LinksFrom is less common in TOML
**How to avoid:** Explicit validation pass for both maps
**Warning signs:** Invalid LinksFrom references not caught

### Pitfall 3: Wrong Suggestion Threshold
**What goes wrong:** Too many false suggestions, or too few helpful ones
**Why it happens:** Levenshtein distance of 2 may be too permissive for short names
**How to avoid:** Use threshold of 2, but consider name length (e.g., skip suggestions for names < 4 chars)
**Warning signs:** "did you mean" suggestions that make no sense

### Pitfall 4: Missing Error Summary
**What goes wrong:** User doesn't know how many errors exist
**Why it happens:** Forgetting to print count at end
**How to avoid:** Always print "N errors found" summary
**Warning signs:** Users re-running to check for more errors

### Pitfall 5: Line Number Unavailable
**What goes wrong:** Parser doesn't preserve line numbers in Model
**Why it happens:** go-toml only provides line numbers during parsing errors
**How to avoid:** Accept best-effort; omit line numbers when unavailable (per CONTEXT.md)
**Warning signs:** All validation errors show "at line 0"

## Code Examples

Verified patterns from existing code and package documentation:

### ValidationError with Formatting
```go
// Source: Pattern from internal/parser/errors.go
package validator

import "fmt"

// ValidationError represents a single validation failure.
type ValidationError struct {
    Message string
    Path    string // Full dotted path, e.g., "mainapp.api.handler"
    Line    int    // Best-effort, 0 if unknown
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
    if e.Line > 0 {
        return fmt.Sprintf("error: %s at line %d", e.Message, e.Line)
    }
    if e.Path != "" {
        return fmt.Sprintf("error: %s in %s", e.Message, e.Path)
    }
    return "error: " + e.Message
}
```

### Main Validate Function
```go
// Source: Standard validation pattern
package validator

import "github.com/Djarvur/c4drill/internal/parser"

// Validate checks a parsed model for semantic errors.
// Returns nil if valid, or a slice of ValidationErrors.
func Validate(m *parser.Model) []*ValidationError {
    if m == nil {
        return nil
    }

    var errors []*ValidationError

    // Build index for reference validation
    index := buildIndex(m.Units, "")

    // Run all validation rules
    errors = append(errors, validateReferences(m, index)...)
    errors = append(errors, validateSubunitRules(m, index)...)
    errors = append(errors, validateLinkRules(m, index)...)

    return errors
}

// ReportErrors prints validation errors and returns exit code.
func ReportErrors(errors []*ValidationError) int {
    if len(errors) == 0 {
        return 0
    }

    for _, err := range errors {
        fmt.Fprintln(os.Stderr, err.Error())
    }

    count := len(errors)
    if count == 1 {
        fmt.Fprintln(os.Stderr, "1 error found")
    } else {
        fmt.Fprintf(os.Stderr, "%d errors found\n", count)
    }

    return 1
}
```

### Reference Validation (VALD-01)
```go
// Source: Standard reference validation pattern
package validator

func validateReferences(m *parser.Model, index map[string]*UnitInfo) []*ValidationError {
    var errors []*ValidationError

    // Collect all known unit names for suggestions
    var allNames []string
    for name := range index {
        allNames = append(allNames, name)
    }

    for path, info := range index {
        // Check Links references
        for target := range info.Unit.Links {
            if _, exists := index[target]; !exists {
                suggestion := formatSuggestion(target, allNames)
                errors = append(errors, &ValidationError{
                    Message: fmt.Sprintf(`undefined unit "%s" referenced from "%s"%s`, target, path, suggestion),
                    Path:    path,
                })
            }
        }

        // Check LinksFrom references
        for source := range info.Unit.LinksFrom {
            if _, exists := index[source]; !exists {
                suggestion := formatSuggestion(source, allNames)
                errors = append(errors, &ValidationError{
                    Message: fmt.Sprintf(`undefined unit "%s" referenced in linkFrom from "%s"%s`, source, path, suggestion),
                    Path:    path,
                })
            }
        }
    }

    return errors
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Fail-fast validation | Error collection | User decision | Users fix all issues in one pass |
| JSON error output | Plain text | User decision | CLI-friendly, script-parsable |
| No suggestions | "Did you mean" hints | User decision | Better developer experience |

**Deprecated/outdated:**
- Multi-line structured errors: User explicitly chose single-line format
- JSON errors: Out of scope for v1

## Open Questions

1. **Line Number Availability**
   - What we know: go-toml provides line numbers via DecodeError.Position() for parsing errors
   - What's unclear: Whether line numbers can be preserved in the parsed Model for validation errors
   - Recommendation: Accept best-effort; omit line numbers for validation errors (per CONTEXT.md decision)

2. **Maximum Error Limit**
   - What we know: Collecting all errors is the decided approach
   - What's unclear: Whether to cap at some large number (e.g., 100) to prevent memory issues
   - Recommendation: No cap needed for reasonable input files; add cap only if issues arise

3. **Suggestion Threshold Tuning**
   - What we know: Levenshtein distance of 2 is typical for typo detection
   - What's unclear: Whether this works well for short unit names like "db" or "api"
   - Recommendation: Start with threshold of 2; skip suggestions for names shorter than 3 characters

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing package (stdlib) |
| Config file | None -- Go uses *_test.go files |
| Quick run command | `go test ./internal/validator/...` |
| Full suite command | `go test -cover -race ./internal/validator/...` |

### Phase Requirements -> Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| VALD-01 | Undefined unit references detected | Unit | `go test ./internal/validator/... -run TestValidateReferences` | No -- Wave 0 |
| VALD-02 | Links on parent units blocked | Unit | `go test ./internal/validator/... -run TestValidateLinkRules` | No -- Wave 0 |
| VALD-03 | References to parent units blocked | Unit | `go test ./internal/validator/... -run TestValidateLinkRules` | No -- Wave 0 |
| VALD-04 | Invalid subunit types blocked | Unit | `go test ./internal/validator/... -run TestValidateSubunitRules` | No -- Wave 0 |
| VALD-05 | Line numbers in errors (best-effort) | Unit | `go test ./internal/validator/... -run TestValidationError` | No -- Wave 0 |
| VALD-06 | Human-readable error format | Unit | `go test ./internal/validator/... -run TestValidationErrorFormat` | No -- Wave 0 |
| QUAL-04 | 75% coverage | Quality gate | `go test -cover ./...` | N/A |

### Sampling Rate
- **Per task commit:** `go test ./internal/validator/...`
- **Per wave merge:** `go test -cover -race ./internal/validator/...`
- **Phase gate:** Full suite green with 75%+ coverage before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `internal/validator/validator.go` -- main Validate function
- [ ] `internal/validator/errors.go` -- ValidationError type
- [ ] `internal/validator/rules.go` -- individual rule functions
- [ ] `internal/validator/suggest.go` -- Levenshtein suggestion helper
- [ ] `internal/validator/validator_test.go` -- test cases for all rules
- [ ] `go.mod` update -- add `github.com/agnivade/levenshtein v1.2.1`

## Sources

### Primary (HIGH confidence)
- [Phase 1 RESEARCH.md](./../01-foundation-model/01-RESEARCH.md) - Existing patterns, model types
- [internal/parser/errors.go](file://internal/parser/errors.go) - Existing ParseError pattern to extend
- [internal/model/unit.go](file://internal/model/unit.go) - UnitType constants and Unit struct
- [internal/model/link.go](file://internal/model/link.go) - Link struct definition

### Secondary (MEDIUM confidence)
- [github.com/agnivade/levenshtein](https://github.com/agnivade/levenshtein) - Package versions from go list
- [CONTEXT.md](file://.planning/phases/02-validation/02-CONTEXT.md) - User decisions on error format

### Tertiary (LOW confidence)
- Training knowledge of Go validation patterns - Standard patterns, may need verification

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - agnivade/levenshtein is well-known, versions verified
- Architecture: HIGH - Patterns derived from existing codebase and standard Go practices
- Pitfalls: MEDIUM - Based on common validation issues and user decisions
- Validation rules: HIGH - Directly from REQUIREMENTS.md and existing model types

**Research date:** 2026-03-09
**Valid until:** 30 days (stable libraries, Go ecosystem)
