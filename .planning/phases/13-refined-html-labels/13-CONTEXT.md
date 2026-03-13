---
phase: 13
slug: refined-html-labels
created: 2026-03-13
updated: 2026-03-13
status: ready
---

# Phase 13: Refined HTML Labels - Context

**Gathered:** 2026-03-13
**Updated:** 2026-03-13
**Status:** Ready for planning

<domain>
## Phase Boundary

This phase has TWO parts:

**Part A: Bug Fixes (Critical)**
Fix bugs in expanded view graph generation where nested containers are missing:
1. `server.pam` cluster not rendered
2. `server.pam.unix` and `server.pam.cyp` components missing
3. Links to/from nested components not rendered

**Part B: Label Refinements**
1. All units: `shape=box, style=rounded`
2. Table attributes: `border="0" cellpadding="0" cellspacing="0"`
3. Cluster labels use HTML format (same as corresponding unit)
4. Cluster labels use unit type coloring

</domain>

<decisions>
## Implementation Decisions

### Bug Investigation (BUG-01, BUG-02, BUG-03)

**Bug Report:** Running `go run ./cmd/c4drill ./cyp-auth-infra/cyp-auth-infra.toml --expanded -f dot` produces incomplete output:

**Missing in generated DOT:**
1. `server.pam` cluster (container with subunits)
2. `server.pam.unix` component
3. `server.pam.cyp` component
4. Link: `server.sshd` → `server.pam.unix`
5. Link: `server.sshd` → `server.pam.cyp`
6. Link: `server.pam.unix` → `server.etc`
7. Link: `server.pam.cyp` → `server.systemd`

**Root cause hypothesis:** The expanded view graph builder in `internal/graph/builder.go` may not be recursively processing nested containers (containers with subunits).

### Automated Testing (TEST-01)

**Requirement:** Add automated tests that:
1. Generate DOT output from TOML files
2. Verify all expected nodes are present
3. Verify all expected edges/links are present
4. Verify cluster structure is correct

**Test approach:**
- Use `cyp-auth-infra.toml` as test fixture
- Parse generated DOT and assert expected elements
- Test both normal views and `--expanded` mode

### Shape and Style (REFINED-01)

All units MUST render with:
- `shape=box`
- `style=rounded`

This replaces the current `shape=none` approach from Phase 12.

### Table Attributes (REFINED-02)

All HTML tables MUST include:
- `border="0"`
- `cellpadding="0"`
- `cellspacing="0"`

### Cluster Labels (REFINED-03)

**Cluster labels MUST use HTML format** - same as the corresponding unit would have:
- Person cluster → Person HTML label format
- DB cluster → DB HTML label format
- etc.

**Cluster labels MUST be colored** - same font/border color as the unit.

</decisions>

<code_context>
## Existing Code Insights

### Bug Location

The bug is likely in `internal/graph/builder.go`:
- `buildNestedCluster()` - handles recursive cluster building
- `buildCluster()` - handles single-level cluster building
- `buildEdges()` - handles edge creation

**Key observation:** The TOML has nested structure:
```
[server.pam]           # container
[server.pam.unix]      # component inside container
[server.pam.cyp]       # component inside container
```

But the DOT only shows flat structure under `server` cluster.

### Integration Points

**Files to modify:**
1. `internal/graph/builder.go` - Fix nested container processing
2. `internal/graph/graph.go` - Add `Type`, `IsExternal` to Cluster struct
3. `internal/render/labels.go` - Add table attributes
4. `internal/render/converter.go` - Change shape, cluster HTML labels
5. `internal/render/*_test.go` - Add comprehensive tests

### Test Strategy

Create integration tests in `internal/render/integration_test.go`:
```go
func TestExpandedViewGeneratesCompleteGraph(t *testing.T) {
    // Load cyp-auth-infra.toml
    // Generate expanded view
    // Render to DOT
    // Assert all expected nodes present
    // Assert all expected edges present
    // Assert cluster structure correct
}
```

</code_context>

<specifics>
## Specific Ideas

- Use `cyp-auth-infra.toml` as the primary test fixture
- Tests should verify DOT output contains expected substrings (node IDs, edge definitions)
- Consider using a DOT parser library for more robust testing

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

## Files to Modify

| File | Change |
|------|--------|
| `internal/graph/builder.go` | Fix nested container processing in expanded view |
| `internal/graph/graph.go` | Add `Type`, `IsExternal` to Cluster struct |
| `internal/render/labels.go` | Add table attributes to all HTML builders |
| `internal/render/converter.go` | Shape=box, cluster HTML labels |
| `internal/render/integration_test.go` | Add comprehensive DOT verification tests |

---

## Success Criteria

1. **Bug fixed:** `cyp-auth-infra.toml --expanded` generates complete DOT with all clusters and links
2. **Tests pass:** Automated tests verify all expected elements in generated DOT
3. **Labels refined:** All HTML tables have correct attributes, shape=box, style=rounded
4. **Cluster labels:** HTML format with proper coloring

---

*Phase: 13-refined-html-labels*
*Context gathered: 2026-03-13*
*Context updated: 2026-03-13*
