# Domain Pitfalls

**Domain:** C4 Diagram Generation (Go CLI)
**Researched:** 2026-03-09
**Confidence:** MEDIUM (based on web research and go-graphviz issue analysis; no direct production experience with this specific stack)

## Critical Pitfalls

### Pitfall 1: GraphViz Layout Non-Determinism

**What goes wrong:**
Same input TOML produces different visual layouts across runs. Diagrams look different on different machines or even on the same machine between invocations. Node positions shift, edge routing changes, making version control diffing impossible.

**Why it happens:**
GraphViz uses non-deterministic algorithms by default. Without explicit seed or deterministic settings, layout varies based on memory addresses, hash ordering, and other environmental factors. go-graphviz inherits this behavior from the underlying WASM-embedded GraphViz.

**Consequences:**
- Documentation diffs become noise (SVG changes even when architecture unchanged)
- Users cannot reproduce identical diagrams
- CI/CD pipelines generate inconsistent artifacts
- Collaboration breaks (team members see different outputs)

**Prevention:**
- Investigate go-graphviz for deterministic layout options
- If not available, document that outputs may vary
- Consider pinning go-graphviz version tightly
- Generate DOT as primary artifact (deterministic), SVG as secondary

**Detection:**
- Run generation 3 times on same input, compare outputs
- Check if go-graphviz exposes any seed/determinism settings
- Monitor SVG output in version control for spurious changes

**Phase to address:** Phase 1 (Core Generation) - establish deterministic baseline early

---

### Pitfall 2: TOML Reference Integrity Violations

**What goes wrong:**
Links reference non-existent units, circular references create infinite loops, or units with subunits are incorrectly referenced directly. Validation passes but rendering produces broken diagrams or crashes.

**Why it happens:**
TOML parsing is separate from semantic validation. Developers implement basic syntax checking but miss edge cases:
- Forward references (unit defined after link to it)
- Circular dependencies between units
- Referencing a container unit directly when it has subunits (violates C4 model)
- Typos in unit names that pass syntax but fail semantics

**Consequences:**
- Silent failures: diagram renders but is incomplete
- Confusing error messages point to wrong location
- Users frustrated debugging their TOML files
- Invalid C4 model output (violates C4 principles)

**Prevention:**
- Implement two-pass parsing: syntax then semantic validation
- Build reference graph and detect cycles before rendering
- Validate that referenced units exist and are leaf nodes (no subunits)
- Provide clear error messages with line numbers and unit names
- Create comprehensive validation test suite with edge cases

**Detection:**
- Unit tests for each validation rule
- Fuzz testing with malformed TOML
- Integration tests with intentionally broken references

**Phase to address:** Phase 1 (Core Generation) - validation is foundational

---

### Pitfall 3: go-graphviz Memory/Segfault Issues

**What goes wrong:**
Large or complex diagrams cause segfaults, panics, or out-of-memory errors. The WASM-embedded GraphViz has limits that native GraphViz doesn't encounter.

**Why it happens:**
go-graphviz embeds GraphViz compiled to WASM, which has constrained memory compared to native binaries. Complex graphs with many nodes, edges, or deeply nested clusters can exhaust WASM memory or trigger bugs in the WASM runtime.

**Consequences:**
- CLI crashes on legitimate inputs
- No graceful degradation for large diagrams
- Users hit invisible limits with no guidance
- Unreliable tool drives users to alternatives

**Prevention:**
- Document known limits (max nodes, max edges, max nesting depth)
- Implement graceful error handling with actionable messages
- Consider splitting very large diagrams automatically
- Test with progressively larger inputs to find breaking points
- Provide fallback to native GraphViz DOT generation (skip SVG rendering)

**Detection:**
- Stress testing with large generated TOML files
- Memory profiling during rendering
- Monitor crash reports from users

**Phase to address:** Phase 1 (Core Generation) - establish error handling patterns early

---

### Pitfall 4: Collapsed/Expanded State Inconsistency

**What goes wrong:**
Expanded units list doesn't match actual rendering. Some units show as expanded when they shouldn't, or explore links point to non-existent drill-down files. The `expanded` property in TOML doesn't propagate correctly through nesting levels.

**Why it happens:**
The `expanded` property has complex inheritance rules:
- Global default in `[properties]`
- Per-unit overrides
- Nested units can have their own expanded settings
- Edge routing (`edges`) also inherits and affects cluster rendering

Developers miss the interaction between these settings and create inconsistent state.

**Consequences:**
- Diagrams don't match user intent
- Explore links broken (404s)
- Confusion about what "expanded" means at each level
- Output file structure doesn't match expectations

**Prevention:**
- Implement explicit inheritance resolution before rendering
- Validate that expanded units actually have subunits
- Generate drill-down files only for units that are expanded
- Test all combinations of global/local expanded settings
- Document inheritance rules clearly

**Detection:**
- Test matrix: global expanded + local expanded for various nesting levels
- Verify file structure matches expanded state
- Check that all explore links resolve

**Phase to address:** Phase 2 (Views & Styling) - after basic rendering works

---

### Pitfall 5: SVG Interactive Link Generation Failures

**What goes wrong:**
Explore links in SVG don't work, point to wrong files, or break when diagrams are moved. Relative vs absolute path handling is inconsistent. Links work locally but break when hosted on web servers.

**Why it happens:**
SVG links (`<a>` elements with `xlink:href`) have subtle path resolution rules. The tool generates links at render time but doesn't know where the SVG will be hosted. Relative paths work in some contexts but not others.

**Consequences:**
- "Drill-down" functionality broken in practice
- Diagrams only work in specific hosting contexts
- User frustration when links don't work as expected
- Tool feels incomplete despite core functionality working

**Prevention:**
- Use relative paths consistently
- Generate links relative to output directory structure
- Document expected hosting setup
- Consider making link path style configurable (relative/absolute/root-relative)
- Test links in multiple contexts (file://, http://localhost, hosted)

**Detection:**
- Test generated SVGs in browser
- Verify links work from both root and nested diagrams
- Test with different output directory structures

**Phase to address:** Phase 2 (Views & Styling) - links are core to drill-down UX

---

## Moderate Pitfalls

### Pitfall 6: Font Rendering Inconsistency

**What goes wrong:**
Text in diagrams renders with wrong fonts, missing characters, or inconsistent sizing across platforms. Special characters (Unicode) don't display correctly.

**Prevention:**
- Specify font explicitly in generated DOT
- Use widely available fonts (Arial, Helvetica, sans-serif)
- Test with Unicode characters early
- Document font requirements

---

### Pitfall 7: Edge Routing Style Ignorance

**What goes wrong:**
`edges` property (straight, spline, square) doesn't produce expected results. Splines look messy, straight edges overlap, square edges don't respect clusters.

**Prevention:**
- Map `edges` values to correct GraphViz `splines` attribute
- Test each edge style with various graph topologies
- Document what each style actually does (vs what users expect)

---

### Pitfall 8: Color Value Validation Gaps

**What goes wrong:**
Invalid color values cause rendering failures or are silently ignored. Named colors, hex codes, and transparency handling inconsistent.

**Prevention:**
- Validate color values before rendering
- Support common formats (hex, RGB, named colors)
- Provide clear error for unsupported formats
- Default gracefully when colors invalid

---

## Minor Pitfalls

### Pitfall 9: Missing Input File Error UX

**What goes wrong:**
When input TOML file doesn't exist, error message is generic Go file-not-found rather than helpful context.

**Prevention:**
- Wrap file operations with context-aware error messages
- Include file path in error output
- Suggest common fixes (typo in filename, wrong directory)

---

### Pitfall 10: Output Directory Creation Failure

**What goes wrong:**
Tool fails when output directory doesn't exist, rather than creating it.

**Prevention:**
- Create output directory structure if it doesn't exist
- Handle permission errors gracefully
- Document required permissions

---

## Technical Debt Patterns

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| Skip validation for simple diagrams | Faster initial development | Broken edge cases accumulate, hard to add validation later | Never - validation is core |
| Use string matching for TOML parsing | Avoid dependency on TOML library | Fragile, breaks on edge cases, hard to extend | Never - use proper TOML library |
| Generate SVG without DOT intermediate | Simpler code path | Harder to debug, no fallback, can't inspect intermediate | Never - always generate DOT |
| Hardcode graphviz layout settings | Faster to implement | Users can't tune layouts, tool inflexible | MVP only - make configurable in Phase 2 |
| Ignore cluster depth limits | Simpler rendering code | Crashes on deeply nested structures | MVP only - add limits and validation |

---

## Integration Gotchas

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| go-graphviz | Assuming it behaves like native GraphViz | Test thoroughly, document differences, expect WASM limitations |
| TOML library | Using default struct tag mapping | Implement custom unmarshalling for nested unit types and link objects |
| File system | Assuming UTF-8 everywhere | Handle encoding explicitly, test with non-ASCII paths |
| SVG browser rendering | Assuming consistent viewBox handling | Test in multiple browsers, validate SVG structure |

---

## Performance Traps

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|----------------|
| Unbounded graph size | OOM, crashes, infinite render time | Enforce max nodes/edges, fail fast with clear message | 500+ nodes (estimate, test to confirm) |
| Deep nesting | Stack overflow in traversal, cluster rendering failures | Limit nesting depth, validate before rendering | 5+ levels (estimate) |
| Many edges between same nodes | Overlapping edges unreadable, rendering slow | Aggregate edges, use edge labels, limit edge count | 10+ edges same pair |
| Large labels | Text overflow, broken layout | Truncate long labels, validate label length | Labels > 100 chars |

---

## Security Mistakes

| Mistake | Risk | Prevention |
|---------|------|------------|
| Arbitrary file paths in TOML | Path traversal, overwriting sensitive files | Validate paths stay within output directory, sanitize inputs |
| Unbounded input file size | DoS via memory exhaustion | Enforce max file size, stream parsing if possible |
| Executing embedded code | Arbitrary code execution | Never execute content from TOML, treat as pure data |

---

## UX Pitfalls

| Pitfall | User Impact | Better Approach |
|---------|-------------|-----------------|
| Cryptic validation errors | Users can't fix their TOML files | Error messages with line numbers, examples, and suggestions |
| Silent output overwriting | Users lose previous work | Warn before overwriting, require --force flag |
| No preview of what will be generated | Users iterate blindly | Dry-run mode showing what files will be created |
| Inconsistent CLI flags | Users can't predict behavior | Follow standard CLI conventions, consistent flag naming |

---

## "Looks Done But Isn't" Checklist

- [ ] **Validation:** Often missing forward reference detection - verify links resolve after full parse
- [ ] **Validation:** Often missing circular reference detection - verify no cycles in reference graph
- [ ] **Validation:** Often missing subunit reference rule - verify units with subunits aren't directly linked
- [ ] **Collapsed/Expanded:** Often missing inheritance validation - verify global + local expanded settings interact correctly
- [ ] **File Output:** Often missing directory creation - verify output directories created if missing
- [ ] **Error Messages:** Often missing line numbers - verify errors point to specific TOML locations
- [ ] **Edge Cases:** Often missing empty input handling - verify graceful handling of empty/minimal TOML
- [ ] **Links:** Often missing broken link detection - verify all explore links resolve to actual files

---

## Recovery Strategies

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| Non-deterministic layout | LOW | Regenerate with same input; if persistent, investigate go-graphviz settings |
| Reference integrity violations | MEDIUM | Fix TOML file; tool should provide clear error location |
| go-graphviz crashes | HIGH | Reduce diagram complexity; use DOT-only output; consider native GraphViz |
| Broken explore links | LOW | Fix expanded state; regenerate; verify file structure |
| Font rendering issues | LOW | Specify font explicitly in TOML or tool defaults |

---

## Pitfall-to-Phase Mapping

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| GraphViz Layout Non-Determinism | Phase 1 (Core Generation) | Run generation multiple times, compare outputs |
| TOML Reference Integrity | Phase 1 (Core Generation) | Unit tests for all validation rules, edge case tests |
| go-graphviz Memory Issues | Phase 1 (Core Generation) | Stress testing, memory profiling, documented limits |
| Collapsed/Expanded Inconsistency | Phase 2 (Views & Styling) | Test matrix for all expanded combinations |
| SVG Link Generation | Phase 2 (Views & Styling) | Test links in browser, multiple hosting contexts |
| Font Rendering | Phase 2 (Views & Styling) | Test on multiple platforms, Unicode characters |
| Edge Routing | Phase 2 (Views & Styling) | Visual inspection of each edge style |
| Color Validation | Phase 1 or 2 | Tests for each supported format |
| Error Message UX | Phase 1 (Core Generation) | User testing with intentional errors |
| Output Directory Handling | Phase 1 (Core Generation) | Test with missing directories |

---

## Sources

- C4 Model Official Site (c4model.com) - Core diagram principles
- go-graphviz GitHub Repository and Issues - Common rendering problems, segfaults, font issues
- Structurizr Documentation (docs.structurizr.com) - C4 model validation patterns
- GraphViz Documentation - Layout algorithms, edge routing options
- TOML Specification - Parsing and validation considerations

**Confidence Assessment:**
- go-graphviz issues: HIGH (direct from repository issues)
- C4 model pitfalls: MEDIUM (inferred from model principles, no production experience)
- TOML validation pitfalls: MEDIUM (general parsing experience, not specific to this schema)
- Performance limits: LOW (estimated, need empirical testing)

---

*Pitfalls research for: C4 Diagram Generation CLI*
*Researched: 2026-03-09*
