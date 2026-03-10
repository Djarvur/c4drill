# Project Research Summary

**Project:** C4Drill v1.1 AI-Ready
**Domain:** C4 Diagram Generation CLI (Milestone: AI Documentation + All-Expanded Mode)
**Researched:** 2026-03-10
**Confidence:** MEDIUM-HIGH

## Executive Summary

C4Drill v1.1 adds two targeted features to an existing, stable v1.0 codebase (9,624 LOC Go). The TOML Language Manual (CLAUDE.md) enables AI assistants to generate valid C4Drill TOML files through structured documentation with examples and validation rules. The All-Expanded Mode (`--expanded` flag) renders the complete architecture in a single diagram with all nested units expanded and cross-level edges visible, outputting to `{basename}.expanded.{ext}`.

The recommended approach is to implement these as independent parallel tracks with no shared code changes. CLAUDE.md is purely documentation work requiring no code modifications. All-Expanded mode adds a new `GenerateAllExpandedView()` function without modifying existing view generation, ensuring zero regression risk. The existing stack (Go 1.26, go-toml v2, go-graphviz, Cobra) fully supports both features with no new dependencies.

Key risks are: (1) AI documentation drift from code reality - mitigated by CI validation of examples; (2) Cross-level edge visual explosion in all-expanded mode - mitigated by GraphViz native cluster edge handling and documented expectations; (3) Layout quality for deeply nested structures - requires empirical testing with real architectures.

## Key Findings

### Recommended Stack

**No new dependencies required.** The existing stack is at latest stable versions and fully supports v1.1 features:

- **Go 1.26.1**: Runtime - already current
- **go-toml v2.2.4**: TOML parsing - used for CLAUDE.md example validation
- **go-graphviz v0.2.10**: DOT/SVG rendering - handles nested clusters and cross-level edges natively
- **Cobra v1.10.2**: CLI framework - standard flag pattern for `--expanded`

**Implementation footprint:** ~300 lines new code, ~50 lines modified across 4 files. CLAUDE.md adds ~200 lines of documentation.

### Expected Features

**Must have (v1.1 MVP):**

| Feature | Description | Complexity |
|---------|-------------|------------|
| CLAUDE.md | AI-focused TOML manual with schema, examples, validation rules, prompt patterns | LOW (documentation) |
| `--expanded` flag | CLI flag recognized, triggers all-expanded rendering path | LOW |
| All-expanded view | Single view with all units expanded as nested clusters | MEDIUM |
| Cross-level edges | Edges between units at different nesting levels visible | HIGH |
| Output naming | `{basename}.expanded.{ext}` format | LOW |

**Defer (v1.2+):**

- Interactive legend in expanded view
- Partial expansion (`--expand=unit1,unit2`)
- AI validation workflow integration
- Edge filtering/aggregation options for cluttered diagrams

### Architecture Approach

The existing pipeline architecture (Parse -> Validate -> Generate Views -> Build Graphs -> Render -> Write) extends naturally for all-expanded mode. A new view type is added alongside existing C1/C2/C3 generators.

**Major components affected:**

1. **cmd/c4drill/root.go** - Add `--expanded` flag, conditional view generation call
2. **internal/view/scope.go** - Add `GenerateAllExpandedView()` (new function, no modifications to existing)
3. **internal/graph/builder.go** - Extend edge building for cross-level edges
4. **internal/output/writer.go** - Add method for expanded output naming
5. **CLAUDE.md** - Root-level documentation file (no code integration)

**Critical architectural decision:** Implement all-expanded as a completely separate code path. Do NOT modify `GenerateC1View`, `GenerateC2View`, or `GenerateC3View`. This prevents regression and allows independent testing.

### Critical Pitfalls

1. **AI Documentation Drift (Pitfall A2)** - CLAUDE.md examples become out of sync with parser/validator. **Prevention:** CI check that parses all CLAUDE.md TOML blocks with actual parser, fails on mismatch. Add version stamp to CLAUDE.md.

2. **Cross-Level Edge Explosion (Pitfall A3)** - All-expanded diagrams become unreadable "hairballs" with too many edges crossing cluster boundaries. **Prevention:** Document that all-expanded prioritizes completeness over aesthetics. GraphViz handles cluster-crossing edges natively. Consider edge aggregation for v1.2 if needed.

3. **Breaking Existing View Logic (Pitfall A4)** - Changes to view generation inadvertently affect C1/C2/C3 output. **Prevention:** Implement `GenerateAllExpandedView()` as separate function. Add comprehensive regression tests before implementation. Require all existing tests to pass.

4. **Output File Naming Collision (Pitfall A5)** - `{basename}.expanded.{ext}` conflicts with user's system named "expanded". **Prevention:** Use explicit naming convention. Check for existing file before writing. Consider `{basename}.all-expanded.{ext}` for clarity.

5. **GraphViz Layout Issues (Pitfall A6)** - Deep nesting + cross-edges cause poor arrangements. **Prevention:** Test with realistic complex models. Set expectations in documentation. Consider layout algorithm options.

## Implications for Roadmap

Based on research, suggested phase structure for v1.1:

### Phase 1: AI Documentation (CLAUDE.md)

**Rationale:** Documentation task with no code dependencies. Can be completed in parallel with Phase 2. Establishes AI usability foundation.

**Delivers:** CLAUDE.md file in repository root with complete schema reference, validation rules, 3-5 working examples (minimal, medium, complex), prompt patterns, and edge case coverage (external types, deep nesting, bidirectional links).

**Addresses:** Feature 1 (TOML Language Manual)

**Avoids:** Pitfall A1 (context overload), Pitfall A2 (documentation drift), Pitfall A7 (missing edge cases)

**Estimated effort:** LOW (~200 lines documentation)

### Phase 2: All-Expanded Mode

**Rationale:** Code implementation requiring changes to 4 files. Builds on existing architecture patterns. Independent of Phase 1.

**Delivers:** `--expanded` CLI flag, `GenerateAllExpandedView()` function, cross-level edge handling, `{basename}.expanded.{ext}` output.

**Uses:** Cobra for flag parsing, existing view/graph patterns, go-graphviz for nested cluster rendering

**Implements:** Feature 2 (All-Expanded Mode)

**Avoids:** Pitfall A3 (edge explosion - via GraphViz native handling), Pitfall A4 (breaks views - via separate code path), Pitfall A5 (naming collision - via explicit naming), Pitfall A6 (layout - via testing), Pitfall A8 (performance - via benchmarking)

**Estimated effort:** MEDIUM (~300 lines new code, ~50 lines modified)

### Phase Ordering Rationale

- **Phase 1 and Phase 2 are independent** - can run in parallel or either order
- **No shared code changes** between phases - zero merge conflict risk
- **Phase 1 is documentation-only** - no test requirements
- **Phase 2 requires regression testing** - establish baseline tests before changes

### Research Flags

Phases likely needing deeper research during planning:

- **Phase 2 (All-Expanded Mode):** Cross-level edge rendering behavior needs empirical testing with real architectures. GraphViz layout quality for deep nesting is unknown without experimentation.

Phases with standard patterns (skip research-phase):

- **Phase 1 (AI Documentation):** Well-established documentation patterns. Existing model/types define schema. Test examples provide source material.
- **Phase 2 CLI flag:** Standard Cobra pattern, existing code shows exact approach.
- **Phase 2 output naming:** Simple string formatting, existing writer shows pattern.

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | Verified existing stack covers all needs, no new dependencies |
| Features | MEDIUM-HIGH | Clear scope, based on codebase analysis |
| Architecture | HIGH | Existing patterns directly applicable, separate code path reduces risk |
| Pitfalls | MEDIUM | AI documentation pitfalls inferred from patterns, all-expanded rendering needs empirical validation |

**Overall confidence:** MEDIUM-HIGH

### Gaps to Address

- **CLAUDE.md effectiveness:** Test with actual AI assistants (Claude, GPT-4) to validate structure works. Run AI-generated TOML through validator.
- **All-Expanded layout quality:** Test with various architectures (microservices, event-driven, monolith) to assess GraphViz output. May require layout tuning.
- **Performance limits:** Benchmark all-expanded mode with 50+ unit models. Document acceptable limits.
- **Edge routing behavior:** Verify GraphViz handles cross-level edges as expected. Test with deeply nested structures.

## Sources

### Primary (HIGH confidence)

- Project codebase analysis (internal/view, internal/graph, internal/model, cmd/c4drill) - HIGH confidence
- go.mod verified versions - HIGH confidence
- Existing testdata examples - HIGH confidence
- GraphViz documentation on nested clusters - HIGH confidence

### Secondary (MEDIUM confidence)

- Cobra CLI flag patterns - HIGH confidence (well-established)
- C4 Model specification (c4model.com) - HIGH confidence
- go-graphviz library documentation - MEDIUM confidence
- Structurizr reference implementation - MEDIUM confidence

### Tertiary (LOW confidence)

- AI prompt file best practices - LOW confidence (inferred from patterns, not empirical testing)
- Cross-level edge rendering at scale - LOW confidence (needs empirical validation)

---
*Research completed: 2026-03-10*
*Ready for roadmap: yes*
*Focus: v1.1 AI-Ready Milestone (TOML Language Manual + All-Expanded Mode)*
