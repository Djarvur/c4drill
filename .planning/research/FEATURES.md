# Feature Landscape: v1.1 AI-Ready

**Domain:** C4Drill v1.1 Milestone Features
**Researched:** 2026-03-10
**Scope:** NEW features only (TOML Language Manual, All-Expanded Mode)
**Confidence:** MEDIUM (web search rate-limited, based on existing knowledge and codebase analysis)

## Executive Summary

This research covers two specific features for the v1.1 milestone:

1. **TOML Language Manual (AI-focused)** — A CLAUDE.md file that teaches AI assistants how to write valid C4Drill TOML, plus a human reference document
2. **All-Expanded Mode** — A `--expanded` CLI flag that renders the complete architecture in a single view with all units expanded and cross-level edges visible

Both features build on the existing v1.0 foundation (parser, validator, view generator, graph builder, renderer).

---

## Feature 1: TOML Language Manual (AI-Focused)

### What It Is

A documentation file (CLAUDE.md) that provides structured instructions for AI assistants to generate valid C4Drill TOML files. This enables users to describe their architecture in natural language and have an AI produce correct TOML output.

### Table Stakes (Expected Behavior)

| Aspect | Expected Behavior | Complexity | Notes |
|--------|-------------------|------------|-------|
| Schema documentation | Complete TOML schema with all types, attributes, constraints | LOW | Already exists in PROJECT.md and README.md |
| Validation rules | Clear explanation of what makes TOML valid/invalid | LOW | Already enforced by validator |
| Examples | Working examples covering all unit types and link patterns | LOW | Existing testdata/*.toml files |
| Error patterns | Common mistakes and how to avoid them | MEDIUM | Derived from validator errors |

### Differentiators (AI-Specific)

| Aspect | Value Proposition | Complexity | Notes |
|--------|-------------------|------------|-------|
| AI-readable structure | Structured for machine parsing, not just human reading | MEDIUM | Clear sections, explicit rules, examples with explanations |
| Prompt patterns | Example prompts for generating C4Drill TOML | LOW | "Create a C4 diagram for an e-commerce system..." |
| Validation checklist | Step-by-step validation checklist for AI to self-check | LOW | Based on existing validation rules |
| Common patterns | Pre-built patterns for typical architectures | MEDIUM | Microservices, monolith, event-driven, etc. |

### Anti-Features (What NOT to Do)

| Anti-Feature | Why Avoid | What to Do Instead |
|--------------|-----------|-------------------|
| Verbose prose | AI works better with structured, concise rules | Use tables, lists, code blocks |
| Ambiguous language | "Usually," "typically," "often" confuse AI | Use "always," "never," "must," "must not" |
| Missing edge cases | AI will generate edge cases | Document all constraints explicitly |
| Outdated examples | AI will copy outdated patterns | Keep examples in sync with validator |

### Expected Structure

```
# C4Drill TOML Language Manual

## Quick Reference
- Unit types table
- Link syntax
- Common patterns

## Schema Reference
- [properties] section
- Unit sections (all types)
- Link objects
- Inheritance rules

## Validation Rules
- Reference integrity
- Type constraints
- Subunit rules

## Common Patterns
- Microservices architecture
- Event-driven architecture
- Database-per-service

## Examples
- Minimal working example
- Full-featured example

## AI Prompt Patterns
- How to prompt for C4Drill generation
- Self-check checklist
```

### Dependencies on Existing Code

| Dependency | How It's Used |
|------------|---------------|
| internal/model/unit.go | Source of truth for Unit struct fields |
| internal/model/properties.go | Properties schema |
| internal/model/link.go | Link object structure |
| internal/parser/errors.go | Error messages to document |
| internal/validator | Validation rules reference |
| testdata/*.toml | Example source material |

### Complexity Assessment

**Overall: LOW-MEDIUM**

- Documentation task, not implementation
- Source material already exists in codebase
- Primary work: organizing and formatting for AI consumption
- Risk: Keeping documentation in sync with code

---

## Feature 2: All-Expanded Mode

### What It Is

A CLI flag (`--expanded`) that generates a single diagram showing the complete architecture with all nested units expanded inline. This contrasts with the current multi-file drill-down approach where each expanded unit creates a separate diagram file.

### Table Stakes (Expected Behavior)

| Aspect | Expected Behavior | Complexity | Notes |
|--------|-------------------|------------|-------|
| CLI flag | `--expanded` flag recognized by CLI | LOW | Standard Cobra flag pattern |
| Output filename | `{basename}.expanded.{ext}` format | LOW | Simple string manipulation |
| All units expanded | Every unit with subunits rendered as cluster | MEDIUM | Recursive expansion logic |
| Cross-level edges | Edges between units at different nesting levels | HIGH | Key complexity area |

### Differentiators (Value-Add)

| Aspect | Value Proposition | Complexity | Notes |
|--------|-------------------|------------|-------|
| Single-view completeness | See entire architecture at once | LOW | Primary user value |
| Cross-level visibility | Edges from component to external system visible | HIGH | Currently requires drill-down |
| Print/share friendly | One file instead of directory structure | LOW | Easier distribution |
| AI-friendly output | Complete context for AI analysis | MEDIUM | Enables "describe this architecture" |

### Anti-Features (What NOT to Do)

| Anti-Feature | Why Avoid | What to Do Instead |
|--------------|-----------|-------------------|
| Recursive nesting limits | Arbitrarily limiting depth breaks completeness | Expand all levels, document GraphViz limits |
| Edge filtering | "Only show same-level edges" defeats purpose | Show ALL edges, let GraphViz handle routing |
| Separate clusters per level | Creates visual fragmentation | Use nested clusters naturally |
| Replacing drill-down | Some users prefer focused views | Keep as option, not replacement |

### Technical Approach

**Current Behavior (v1.0):**

```
TOML → Parse → Validate → Generate Views (per expanded unit) → Build Graphs → Render → Write Multiple Files
```

**New Behavior (--expanded):**

```
TOML → Parse → Validate → Generate Single Expanded View → Build Graph with Nested Clusters → Render → Write Single File
```

### Implementation Components

| Component | Current State | Changes Needed |
|-----------|---------------|----------------|
| cmd/c4drill/main.go | No `--expanded` flag | Add flag, conditional logic |
| internal/view/scope.go | GenerateC1View, GenerateC2View, GenerateC3View | Add GenerateExpandedView() |
| internal/graph/builder.go | BuildGraph handles single-level expansion | Handle recursive nested clusters |
| internal/graph/builder.go | buildEdges filters by view scope | Include cross-level edges |
| internal/output/writer.go | Multi-file output structure | Single-file output for expanded mode |

### Cross-Level Edge Handling

**The Core Challenge:**

Currently, edges are filtered to only show connections between units in the same view. For all-expanded mode, edges between units at different nesting levels must be rendered.

**Example:**

```toml
[mainapp.api.handlers]
type = "component"
linkFrom = { "external.system" = { description = "Calls" } }
```

In current behavior:
- C1 view shows: `external.system → mainapp` (collapsed)
- C2 view shows: `external.system → mainapp.api` (if mainapp expanded)
- C3 view shows: `external.system → mainapp.api.handlers` (if api expanded)

In all-expanded mode:
- Single view shows: `external.system → mainapp.api.handlers` (edge crosses cluster boundaries)

**GraphViz Handling:**

GraphViz handles edges crossing cluster boundaries natively. The edge will be drawn from the external node to the deeply nested node, crossing through parent cluster boundaries.

### Dependencies on Existing Code

| Dependency | How It's Used |
|------------|---------------|
| internal/view/view.go | Entry type for view representation |
| internal/view/scope.go | Pattern for view generation |
| internal/graph/builder.go | Cluster building pattern |
| internal/graph/graph.go | Graph, Cluster, Node, Edge types |
| internal/model/unit.go | Recursive Subunits map |

### Output Structure

**Current (without --expanded):**
```
output/
├── architecture.svg           # C1 view
└── architecture/              # Expanded units
    ├── mainapp.svg            # C2 for mainapp
    └── mainapp/
        └── api.svg            # C3 for mainapp.api
```

**New (with --expanded):**
```
output/
├── architecture.svg           # C1 view (unchanged)
├── architecture.expanded.svg  # All-expanded view (NEW)
└── architecture/              # Expanded units (unchanged)
    └── ...
```

### Complexity Assessment

**Overall: MEDIUM-HIGH**

| Aspect | Complexity | Reason |
|--------|------------|--------|
| CLI flag | LOW | Standard Cobra pattern |
| Output filename | LOW | String formatting |
| Recursive expansion | MEDIUM | Tree traversal, already done for nested TOML |
| Nested clusters | MEDIUM | GraphViz supports, but need to test layout quality |
| Cross-level edges | HIGH | Must include edges currently filtered out |
| Edge routing | MEDIUM | GraphViz handles, but may produce cluttered diagrams |

### Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Diagram too cluttered | HIGH | MEDIUM | Document that all-expanded works best for small/medium architectures |
| GraphViz layout issues | MEDIUM | LOW | Test with various architectures, tune cluster styling |
| Edge crossings unreadable | MEDIUM | MEDIUM | Use edge routing options (spline, square) |
| Memory limits for large models | LOW | HIGH | Reuse existing go-graphviz limits documentation |

---

## Feature Dependencies

```
[CLAUDE.md Creation]
    └──requires──> [Existing Schema Documentation]
    └──requires──> [Validation Rules (internal/validator)]
    └──requires──> [Example Files (testdata/)]

[All-Expanded Mode]
    └──requires──> [internal/view/scope.go] (new GenerateExpandedView)
    └──requires──> [internal/graph/builder.go] (nested cluster support)
    └──requires──> [internal/output/writer.go] (single-file output)
    └──requires──> [cmd/c4drill/main.go] (flag parsing)
```

---

## MVP Recommendation

### Must Have (v1.1)

1. **CLAUDE.md** — AI-focused TOML manual with:
   - Complete schema reference
   - Validation rules
   - 3-5 working examples
   - Prompt patterns

2. **All-Expanded Mode** — `--expanded` flag with:
   - Single-file output (`{basename}.expanded.{ext}`)
   - All units expanded as nested clusters
   - Cross-level edges visible

### Defer (v1.2+)

- **Interactive legend** in expanded view (too complex for v1.1)
- **Partial expansion** (`--expand=unit1,unit2`) (YAGNI for now)
- **AI validation workflow** (CLAUDE.md instructs AI to run c4drill to validate)

---

## Sources

### Primary (HIGH confidence)
- Project codebase analysis (internal/view, internal/graph, internal/model) — HIGH confidence
- README.md documentation — HIGH confidence
- Existing testdata examples — HIGH confidence

### Secondary (MEDIUM confidence)
- GraphViz documentation on nested clusters and edge routing — MEDIUM confidence
- Cobra CLI flag patterns — HIGH confidence (well-established)

### Tertiary (LOW confidence)
- AI prompt file best practices — LOW confidence (web search rate-limited, based on general knowledge)
- Similar tools' all-expanded modes — LOW confidence (web search rate-limited)

---

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| CLAUDE.md structure | MEDIUM | Based on existing documentation patterns, not empirical AI testing |
| All-Expanded Mode approach | HIGH | Based on codebase analysis, GraphViz capabilities |
| Cross-level edges | MEDIUM | GraphViz supports, but layout quality unknown without testing |
| User value | HIGH | Both features directly address real user needs |

**Overall confidence:** MEDIUM-HIGH

---

## Gaps to Address

- **CLAUDE.md effectiveness:** Test with actual AI assistants to validate structure works
- **All-Expanded layout quality:** Test with various architectures to assess GraphViz output
- **Performance limits:** Test expanded mode with deeply nested, highly connected models

---

*Research for: C4Drill v1.1 AI-Ready Milestone*
*Researched: 2026-03-10*
*Focus: TOML Language Manual + All-Expanded Mode*
