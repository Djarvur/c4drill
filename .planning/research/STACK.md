# Technology Stack — v1.1 AI-Ready Milestone

**Project:** C4Drill
**Milestone:** v1.1 AI-Ready (AI documentation + All-Expanded mode)
**Researched:** 2026-03-10
**Confidence:** HIGH

## Executive Summary

For the v1.1 milestone features (TOML Language Manual and All-Expanded rendering mode), **NO new external dependencies are required**. The existing stack fully supports both features through internal code additions only.

## Existing Stack (Verified Current)

These dependencies are already at their latest stable versions:

| Technology | Version | Purpose | Status |
|------------|---------|---------|--------|
| Go | 1.26.1 | Runtime | Latest |
| go-toml v2 | v2.2.4 | TOML parsing | Latest (verified: `go list -m -versions`) |
| go-graphviz | v0.2.10 | DOT/SVG rendering | Latest (verified: `go list -m -versions`) |
| cobra | v1.10.2 | CLI framework | Latest (verified: `go list -m -versions`) |
| testify | v1.11.1 | Testing | Latest |

## New Stack Requirements

### Feature 1: TOML Language Manual (CLAUDE.md + Human Reference)

**Dependencies Required:** NONE

The CLAUDE.md file is a static documentation file placed in the repository root. It contains:
- TOML schema reference
- Example snippets
- Validation rules
- AI prompt guidance

**Implementation approach:**
- Manual authoring in Markdown
- No tooling required
- Existing model/types already define the schema

**Integration points:**
- Reference `internal/model/unit.go` for type definitions
- Reference `internal/model/properties.go` for root properties
- Reference `internal/parser/parser.go` for parsing behavior
- Reference `internal/validator/rules.go` for validation rules

### Feature 2: All-Expanded Rendering Mode (`--expanded` flag)

**Dependencies Required:** NONE

The `--expanded` flag is a CLI addition that triggers a different rendering path. All required capabilities exist:

| Capability | Existing Module | How to Use |
|------------|-----------------|------------|
| CLI flag parsing | `spf13/cobra` | Add `PersistentFlags().Bool` in `cmd/c4drill/root.go` |
| View generation | `internal/view/scope.go` | Create new `GenerateAllExpandedView()` function |
| Graph building | `internal/graph/builder.go` | Extend to handle recursive expansion |
| Edge rendering | `internal/graph/builder.go` | Extend `buildEdges()` for cross-level edges |
| Output naming | `internal/output/writer.go` | Add method for `{basename}.expanded.{ext}` |

**Integration points:**
1. `cmd/c4drill/root.go`:
   - Add `expanded bool` flag variable
   - Modify `runRoot()` to check flag and call different view generation

2. `internal/view/scope.go`:
   - Add `GenerateAllExpandedView(m *parser.Model) *View`
   - Recursively expand all systems/boxes
   - Collect all units at all levels into single flat view

3. `internal/graph/builder.go`:
   - Modify edge building to include cross-level edges
   - Handle edges between units at different hierarchy depths

4. `internal/output/writer.go`:
   - Add `WriteExpanded(basename, format string, data []byte)` method
   - Output to `{basename}.expanded.{ext}` instead of hierarchy

## What NOT to Add

| Avoid | Why | Reasoning |
|-------|-----|-----------|
| Additional CLI frameworks | Cobra already installed | No need for urfave/cli or others |
| Template engines (text/template, etc.) | CLAUDE.md is static | Hand-authored documentation, not generated |
| Schema validation libraries | Already have custom validator | `internal/validator/` handles all rules |
| Markdown generators | CLAUDE.md written manually | Direct authoring gives better control |
| Graph layout libraries | go-graphviz handles layout | No need for dagre, d3-hierarchy, etc. |

## Implementation Complexity Assessment

| Feature | New Code | Modified Code | Complexity |
|---------|----------|---------------|------------|
| CLAUDE.md documentation | ~200 lines | None | LOW (documentation only) |
| `--expanded` CLI flag | ~5 lines | `root.go` | LOW |
| All-expanded view generation | ~50 lines | `scope.go` | MEDIUM (recursive traversal) |
| Cross-level edge handling | ~30 lines | `builder.go` | MEDIUM (edge path resolution) |
| Expanded output naming | ~10 lines | `writer.go` | LOW |

**Total estimated new code:** ~300 lines
**Total estimated modified code:** ~50 lines across 4 files

## Testing Strategy

No new testing libraries required. Use existing `stretchr/testify`:

1. **CLAUDE.md**: Manual review, no automated tests needed
2. **All-Expanded mode**:
   - Unit tests in `view/scope_test.go` for `GenerateAllExpandedView()`
   - Integration test verifying `{basename}.expanded.svg` output
   - Test case: TOML with nested systems/boxes, verify all appear in single diagram

## Confidence Assessment

| Area | Confidence | Reason |
|------|------------|--------|
| No new dependencies | HIGH | Verified existing stack covers all needs |
| CLI flag integration | HIGH | Cobra well-documented, existing pattern in root.go |
| View generation | HIGH | Existing `GenerateC1View/C2View/C3View` provide clear patterns |
| Cross-level edges | MEDIUM | Requires careful path resolution, but model supports it |
| Output naming | HIGH | Simple string manipulation in existing writer |

## Sources

- `go.mod` — Verified actual dependency versions (HIGH confidence)
- `go list -m -versions` — Verified latest available versions (HIGH confidence)
- `cmd/c4drill/root.go` — CLI flag pattern analysis (HIGH confidence)
- `internal/view/scope.go` — View generation patterns (HIGH confidence)
- `internal/graph/builder.go` — Graph building patterns (HIGH confidence)
- `internal/output/writer.go` — Output file naming patterns (HIGH confidence)

---
*Stack research for: C4Drill v1.1 AI-Ready milestone*
*Researched: 2026-03-10*
