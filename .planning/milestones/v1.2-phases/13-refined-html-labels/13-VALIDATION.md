---
phase: 13
slug: refined-html-labels
created: 2026-03-13
---

# Phase 13: Refined HTML Labels - Validation Strategy

## Validation Requirements

### BUG-01: Nested Cluster Rendering
**Requirement:** Containers with subunits (nested containers) must render as clusters inside their parent cluster.

**Test Case:** `cyp-auth-infra.toml` with `--expanded` flag
**Expected in DOT output:**
- `cluster_cluster_server_pam` subgraph exists
- `server.pam.unix` node exists inside `cluster_cluster_server_pam`
- `server.pam.cyp` node exists inside `cluster_cluster_server_pam`

**Validation:**
```go
strings.Contains(dotOutput, "cluster_cluster_server_pam")
strings.Contains(dotOutput, "\"server.pam.unix\"")
strings.Contains(dotOutput, "\"server.pam.cyp\"")
```

### BUG-02: Nested Component Nodes
**Requirement:** Components inside nested containers must appear as nodes.

**Test Case:** Same as BUG-01
**Expected:** Both component nodes rendered with correct HTML labels.

**Validation:**
```go
// Verify component labels contain expected names
strings.Contains(dotOutput, "PAM Unix Module")
strings.Contains(dotOutput, "PAM CYP Auth Module")
```

### BUG-03: Links To/From Nested Components
**Requirement:** Links defined in TOML must create edges in DOT, even for nested components.

**Test Case:** `cyp-auth-infra.toml` links
**Expected edges:**
- `server.sshd` → `server.pam.unix`
- `server.sshd` → `server.pam.cyp`
- `server.pam.unix` → `server.etc`
- `server.pam.cyp` → `server.systemd`

**Validation:**
```go
strings.Contains(dotOutput, "\"server.sshd\" -> \"server.pam.unix\"")
strings.Contains(dotOutput, "\"server.sshd\" -> \"server.pam.cyp\"")
strings.Contains(dotOutput, "\"server.pam.unix\" -> \"server.etc\"")
strings.Contains(dotOutput, "\"server.pam.cyp\" -> \"server.systemd\"")
```

### REFINED-01: Shape and Style
**Requirement:** All units render with `shape=box, style=rounded`.

**Validation:**
```go
// Check that nodes have shape=box (not shape=none)
// This is set via go-graphviz API, verify in rendered DOT
```

### REFINED-02: Table Attributes
**Requirement:** All HTML tables include `border="0" cellpadding="0" cellspacing="0"`.

**Validation:**
```go
strings.Contains(dotOutput, `border="0" cellpadding="0" cellspacing="0"`)
```

### REFINED-03: Cluster HTML Labels
**Requirement:** Cluster labels use HTML format with proper coloring.

**Validation:**
```go
// Verify cluster label is HTML (contains table tags)
strings.Contains(dotOutput, "label=<<table")
```

## Test File Location

**Primary test fixture:** `cyp-auth-infra/cyp-auth-infra.toml`
**Test file:** `internal/render/integration_test.go`

## Acceptance Criteria

1. All BUG-* tests pass - nested containers and links render correctly
2. All REFINED-* tests pass - shape, table attributes, cluster labels correct
3. `go test ./internal/render/...` passes
4. `go run ./cmd/c4drill ./cyp-auth-infra/cyp-auth-infra.toml --expanded -f dot` produces complete output
