# Domain Pitfalls

**Domain:** C4 Diagram Generation (Go CLI)
**Researched:** 2026-03-10
**Updated for:** v1.1 AI-Ready Milestone
**Confidence:** MEDIUM (based on codebase analysis, GraphViz patterns, and AI documentation best practices; limited direct experience with cross-level edge rendering at scale)

---

## v1.1-Specific Pitfalls (AI-Ready Milestone)

These pitfalls are specific to adding AI documentation and all-expanded rendering mode to the existing C4Drill v1.0 codebase.

### Pitfall A1: AI Prompt File Context Overload

**What goes wrong:**
CLAUDE.md (or similar AI prompt file) contains too much or too little information, making AI assistants either hallucinate features or miss critical constraints.

**Why it happens:**
Developers either dump entire documentation (overwhelming context window) or write vague high-level descriptions (insufficient specificity). The AI cannot infer implicit rules from incomplete documentation.

**Consequences:**
- AI generates invalid TOML that fails validation
- AI suggests patterns that violate C4 model constraints (e.g., linking to units with subunits)
- AI outputs syntax that doesn't match the actual schema
- Users frustrated when "AI knows the tool" but produces broken diagrams

**Prevention:**
- Include complete TOML schema with all unit types and their constraints
- Provide 2-3 complete, valid example TOML files (minimal, medium, complex)
- Document validation rules explicitly (reference integrity, subunit constraints)
- Include edge cases: empty files, single-unit systems, deeply nested structures
- Keep examples in sync with actual parser behavior (test AI outputs against validator)

**Detection:**
- Have AI generate TOML files, run through validator
- Track validation errors from AI-generated inputs
- Compare AI understanding against actual schema quarterly

**Phase to address:** Phase 1 (AI Documentation) - must be correct from the start

---

### Pitfall A2: AI Documentation Drift from Code Reality

**What goes wrong:**
CLAUDE.md describes behavior that doesn't match actual implementation. Schema changes, new features, or bug fixes aren't reflected in the AI documentation.

**Why it happens:**
Documentation is maintained separately from code. No automated check ensures AI docs match actual parser/validator behavior. Developers forget to update CLAUDE.md when modifying schema.

**Consequences:**
- AI generates syntactically correct but semantically wrong TOML
- Users trust AI suggestions that don't work
- Support burden increases as users report "AI said this should work"
- Documentation credibility erodes

**Prevention:**
- Add CI check: parse CLAUDE.md examples with actual parser, validate they pass
- Include version stamp in CLAUDE.md matching code version
- Treat CLAUDE.md as code: require update in same PR as schema changes
- Add "Last verified against version X.X" header

**Detection:**
- Automated: CI pipeline parses all CLAUDE.md examples
- Manual: Periodic review against actual behavior
- User reports of AI-generated TOML failing validation

**Phase to address:** Phase 1 (AI Documentation) - establish CI pattern early

---

### Pitfall A3: All-Expanded Mode Cross-Level Edge Explosion

**What goes wrong:**
When rendering all units expanded simultaneously, cross-level edges (e.g., C3 component to C1 external system) create visual chaos. Too many edges crossing cluster boundaries makes diagrams unreadable.

**Why it happens:**
Normal C4 model assumes one level of detail at a time. All-expanded mode violates this assumption by showing C1, C2, and C3 simultaneously. Every C3 component that links to an external system now has an edge crossing two cluster levels.

**Consequences:**
- Diagrams become "hairball" - visually incomprehensible
- GraphViz layout algorithm struggles, produces poor arrangements
- Edges overlap and become untraceable
- SVG files become huge (performance impact)
- Feature perceived as useless despite being technically correct

**Prevention:**
- Implement edge filtering: option to hide cross-level edges beyond N levels
- Aggregate edges: show single edge from cluster boundary instead of each internal edge
- Add visual indicators for "edges go to external" without drawing all
- Consider hybrid approach: expand one branch at a time, not all simultaneously
- Document that all-expanded is for export/review, not interactive viewing

**Detection:**
- Test with real-world TOML files (10+ systems, each with 5+ containers)
- Count edges in all-expanded vs per-level views
- Visual inspection: can humans trace edge paths?
- Performance: measure SVG file size growth

**Phase to address:** Phase 2 (All-Expanded Mode) - requires iteration on edge rendering

---

### Pitfall A4: All-Expanded Mode Breaks Existing View Logic

**What goes wrong:**
Adding `--expanded` flag requires changes to view generation that inadvertently break existing collapsed/expanded behavior. C1, C2, C3 views start producing incorrect output.

**Why it happens:**
Current view generation (`GenerateC1View`, `GenerateC2View`, `GenerateC3View`) assumes specific scope boundaries. All-expanded mode requires different scope handling. Changes to shared code paths affect existing functionality.

**Consequences:**
- Regression: existing diagrams change unexpectedly
- Validation errors on previously valid TOML files
- User trust erodes as "stable" features break
- Difficult to isolate bug to all-expanded changes

**Prevention:**
- Implement all-expanded as separate view type (`GenerateAllExpandedView`)
- Do NOT modify existing view generation functions
- Add comprehensive regression tests before implementing new mode
- Use feature flag or separate code path initially
- Require all existing tests to pass before merging all-expanded changes

**Detection:**
- Run full test suite before/after each change
- Visual diff: compare SVG output for existing examples
- Integration tests: all-expanded should not affect non-expanded output

**Phase to address:** Phase 2 (All-Expanded Mode) - establish regression tests first

---

### Pitfall A5: All-Expanded Output File Naming Collision

**What goes wrong:**
All-expanded output file (`{basename}.expanded.{ext}`) overwrites or conflicts with existing output files. User runs `--expanded` and loses their C1 context diagram.

**Why it happens:**
Current output structure uses `{basename}.{ext}` for C1 and `{basename}/{unit}.{ext}` for C2/C3. Adding `{basename}.expanded.{ext}` looks safe but edge cases exist:
- User has a system named "expanded" - creates collision
- Overwrite protection not triggered because different flag

**Consequences:**
- Data loss: user's existing diagrams overwritten
- Confusion: which diagram shows what?
- File management complexity for users

**Prevention:**
- Use clear naming: `{basename}.all-expanded.{ext}` (more explicit)
- Check for existing file before writing, warn or error
- Consider subdirectory: `{basename}/_all-expanded.{ext}`
- Document naming convention prominently
- Add `--list-output` flag to show what files will be created

**Detection:**
- Test with system named "expanded" in TOML
- Test with existing `{basename}.expanded.svg` file present
- Verify overwrite behavior matches user expectations

**Phase to address:** Phase 2 (All-Expanded Mode) - determine naming early

---

### Pitfall A6: GraphViz Layout Algorithm Unprepared for Deep Nesting + Cross-Edges

**What goes wrong:**
All-expanded mode with deeply nested structures (C3 components) and cross-level edges causes GraphViz `dot` layout to produce poor arrangements: nodes stacked oddly, edges routing through nodes, excessive whitespace.

**Why it happens:**
GraphViz `dot` algorithm optimizes for hierarchical directed graphs. Deeply nested clusters with cross-level edges violate its assumptions. The algorithm makes local decisions that compound into global poor layout.

**Consequences:**
- Professional-looking diagrams become amateurish in all-expanded mode
- Users blame tool quality rather than understanding the inherent complexity
- Support requests for "fix the layout" that have no good solution

**Prevention:**
- Test with maximum realistic nesting (C3 with 20+ components per container)
- Consider alternative layouts for all-expanded: `fdp`, `neato`, or custom
- Allow user to specify layout algorithm for all-expanded mode
- Set expectations in documentation: "all-expanded prioritizes completeness over aesthetics"
- Implement layout hints: increased `ranksep`, `nodesep` for all-expanded

**Detection:**
- Visual inspection of complex all-expanded diagrams
- User feedback on layout quality
- Compare against professional C4 tools' all-expanded output

**Phase to address:** Phase 2 (All-Expanded Mode) - may require experimentation

---

### Pitfall A7: AI Prompt File Missing Edge Case Examples

**What goes wrong:**
CLAUDE.md includes only "happy path" examples. AI generates TOML that works for simple cases but fails on edge cases: external units, deeply nested structures, bidirectional links, self-referential units.

**Why it happens:**
Documentation writers focus on common cases. Edge cases feel like "advanced topics" to add later. AI has no training on how to handle unusual but valid TOML structures.

**Consequences:**
- AI-generated TOML works for demos but fails on real architectures
- Users discover limitations only after investing time
- Perception that "AI support" is incomplete

**Prevention:**
- Include explicit examples for:
  - External units (`personExternal`, `systemExternal`, `dbExternal`, `queueExternal`)
  - Bidirectional links (`link` + `linkFrom` patterns)
  - Three-level nesting (system -> container -> component)
  - Multiple systems with cross-links
  - Single-unit TOML (minimal valid file)
  - Styling overrides (colors, borders, edge styles)
- Add "Anti-patterns" section showing what NOT to do with explanations
- Test AI with edge case prompts specifically

**Detection:**
- Prompt AI to generate each edge case type
- Run generated TOML through validator
- Track which edge cases AI handles correctly

**Phase to address:** Phase 1 (AI Documentation) - completeness from start

---

### Pitfall A8: All-Expanded Mode Memory/Performance Regression

**What goes wrong:**
All-expanded mode requires rendering the entire model as a single graph, consuming significantly more memory and time than per-level rendering. Large models cause crashes or unacceptable delays.

**Why it happens:**
Per-level rendering creates small graphs (C1, then C2 per system, then C3 per container). All-expanded creates one massive graph with all units simultaneously. GraphViz memory and time scale non-linearly with graph complexity.

**Consequences:**
- CLI hangs on large models (bad UX)
- OOM crashes (worse UX)
- Users avoid using all-expanded mode
- Feature perceived as broken

**Prevention:**
- Benchmark memory/time for progressively larger models
- Set and document maximum model size for all-expanded mode
- Implement graceful degradation: error message suggesting per-level rendering
- Consider progress indicator for long-running all-expanded generation
- Profile and optimize hot paths before release

**Detection:**
- Performance tests with generated large TOML files
- Memory profiling during all-expanded rendering
- User reports of slowness or crashes

**Phase to address:** Phase 2 (All-Expanded Mode) - establish limits early

---

## Existing v1.0 Pitfalls (Still Relevant)

These pitfalls from v1.0 remain relevant for v1.1 development.

### Pitfall 1: GraphViz Layout Non-Determinism

**What goes wrong:**
Same input TOML produces different visual layouts across runs. Diagrams look different on different machines or even on the same machine between invocations.

**Why it happens:**
GraphViz uses non-deterministic algorithms by default. Without explicit seed or deterministic settings, layout varies based on memory addresses, hash ordering, and other environmental factors.

**Consequences:**
- Documentation diffs become noise (SVG changes even when architecture unchanged)
- Users cannot reproduce identical diagrams
- CI/CD pipelines generate inconsistent artifacts

**Prevention:**
- Generate DOT as primary artifact (deterministic), SVG as secondary
- Document that outputs may vary
- Pin go-graphviz version tightly

**Detection:**
- Run generation 3 times on same input, compare outputs
- Monitor SVG output in version control for spurious changes

**Status:** Mitigated in v1.0 by treating DOT as primary artifact

---

### Pitfall 2: TOML Reference Integrity Violations

**What goes wrong:**
Links reference non-existent units, circular references create infinite loops, or units with subunits are incorrectly referenced directly.

**Why it happens:**
TOML parsing is separate from semantic validation. Forward references, circular dependencies, and constraint violations pass syntax but fail semantics.

**Consequences:**
- Silent failures: diagram renders but is incomplete
- Confusing error messages point to wrong location
- Invalid C4 model output

**Prevention:**
- Two-pass parsing: syntax then semantic validation
- Build reference graph and detect cycles before rendering
- Validate that referenced units exist and are leaf nodes
- Clear error messages with line numbers and suggestions

**Status:** Implemented in v1.0 validator

---

### Pitfall 3: go-graphviz Memory/Segfault Issues

**What goes wrong:**
Large or complex diagrams cause segfaults, panics, or out-of-memory errors. WASM-embedded GraphViz has limits that native GraphViz doesn't encounter.

**Why it happens:**
go-graphviz embeds GraphViz compiled to WASM, which has constrained memory. Complex graphs can exhaust WASM memory or trigger runtime bugs.

**Consequences:**
- CLI crashes on legitimate inputs
- No graceful degradation for large diagrams
- Unreliable tool drives users to alternatives

**Prevention:**
- Document known limits (max nodes, max edges, max nesting depth)
- Implement graceful error handling with actionable messages
- Provide fallback to DOT-only output

**Status:** Documented limits in v1.0; more critical for all-expanded mode

---

### Pitfall 4: Collapsed/Expanded State Inconsistency

**What goes wrong:**
Expanded units list doesn't match actual rendering. Some units show as expanded when they shouldn't, or explore links point to non-existent drill-down files.

**Why it happens:**
The `expanded` property has complex inheritance rules across global defaults, per-unit overrides, and nested units.

**Consequences:**
- Diagrams don't match user intent
- Explore links broken (404s)
- Confusion about what "expanded" means at each level

**Prevention:**
- Explicit inheritance resolution before rendering
- Validate that expanded units actually have subunits
- Test all combinations of global/local expanded settings

**Status:** Implemented in v1.0; all-expanded mode adds new complexity

---

## Moderate Pitfalls

### Pitfall M1: Font Rendering Inconsistency

**What goes wrong:**
Text in diagrams renders with wrong fonts, missing characters, or inconsistent sizing across platforms. Unicode characters don't display correctly.

**Prevention:**
- Specify font explicitly in generated DOT
- Use widely available fonts (Arial, Helvetica, sans-serif)
- Test with Unicode characters early

---

### Pitfall M2: Edge Routing Style Ignorance

**What goes wrong:**
`edges` property (straight, spline, square) doesn't produce expected results. Splines look messy, straight edges overlap, square edges don't respect clusters.

**Prevention:**
- Map `edges` values to correct GraphViz `splines` attribute
- Test each edge style with various graph topologies
- Document what each style actually does

---

### Pitfall M3: Color Value Validation Gaps

**What goes wrong:**
Invalid color values cause rendering failures or are silently ignored. Named colors, hex codes, and transparency handling inconsistent.

**Prevention:**
- Validate color values before rendering
- Support common formats (hex, RGB, named colors)
- Default gracefully when colors invalid

---

## Phase-Specific Warnings for v1.1

| Phase Topic | Likely Pitfall | Mitigation |
|-------------|---------------|------------|
| AI Documentation (CLAUDE.md) | Context overload, missing edge cases | Start with minimal viable examples, expand based on validation failures |
| AI Documentation | Drift from code reality | CI check that parses CLAUDE.md examples |
| All-Expanded Rendering | Cross-level edge explosion | Edge aggregation, filtering, visual simplification |
| All-Expanded Rendering | Breaks existing view logic | Separate code path, comprehensive regression tests |
| All-Expanded Output | File naming collision | Use explicit naming, collision detection |
| All-Expanded Layout | Poor GraphViz arrangements | Layout algorithm options, increased spacing |
| All-Expanded Performance | Memory/time regression | Benchmark, document limits, graceful degradation |

---

## Integration Pitfalls (v1.1 Specific)

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| CLAUDE.md <-> Parser | Examples that don't match actual schema | CI parses all examples, fails on mismatch |
| All-Expanded <-> View | Modifying existing view functions | Separate `GenerateAllExpandedView` function |
| All-Expanded <-> Graph | Using same edge building logic | Custom edge handling for cross-level edges |
| All-Expanded <-> Output | Assuming file naming is trivial | Explicit naming convention, collision checks |

---

## "Looks Done But Isn't" Checklist (v1.1)

- [ ] **AI Documentation:** Examples parse without errors - verify all CLAUDE.md TOML blocks with actual parser
- [ ] **AI Documentation:** Examples validate without errors - verify all examples pass validator
- [ ] **AI Documentation:** Edge cases covered - verify external types, deep nesting, bidirectional links
- [ ] **All-Expanded:** Existing output unchanged - verify C1/C2/C3 output identical before/after
- [ ] **All-Expanded:** Cross-level edges visible - verify edges between C3 components and C1 systems render
- [ ] **All-Expanded:** File naming collision-safe - verify with unit named "expanded"
- [ ] **All-Expanded:** Performance acceptable - verify memory/time for 50+ unit model
- [ ] **All-Expanded:** Layout readable - visual inspection of complex all-expanded diagram

---

## Recovery Strategies (v1.1)

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| AI documentation drift | LOW | Update CLAUDE.md, verify examples |
| AI missing edge cases | LOW | Add examples, verify AI generates correctly |
| All-expanded breaks views | MEDIUM | Revert, implement separate code path |
| All-expanded edge explosion | MEDIUM | Implement edge filtering/aggregation |
| All-expanded layout issues | HIGH | May require layout algorithm experimentation |
| All-expanded performance | HIGH | May require model size limits or optimization |

---

## Sources

- C4 Model Official Site (c4model.com) - Core diagram principles
- go-graphviz GitHub Repository and Issues - Rendering problems, performance limits
- GraphViz Documentation - Layout algorithms, cluster handling, cross-level edges
- TOML Specification - Parsing and validation considerations
- Claude Code Documentation patterns - AI prompt file best practices
- Existing C4Drill codebase analysis - View generation, graph building, edge handling

**Confidence Assessment:**
- Existing v1.0 pitfalls: HIGH (from codebase analysis and existing documentation)
- AI documentation pitfalls: MEDIUM (inferred from AI prompt patterns, not direct experience with this codebase)
- All-expanded rendering pitfalls: MEDIUM (inferred from GraphViz behavior and C4 model principles, empirical testing needed for cross-level edge handling)
- Performance limits: LOW (estimated, need empirical testing with large models)

---

*Pitfalls research for: C4 Diagram Generation CLI (v1.1 AI-Ready Milestone)*
*Researched: 2026-03-10*
