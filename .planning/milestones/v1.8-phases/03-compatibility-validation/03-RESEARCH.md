# Phase 3: Compatibility & Validation - Research

**Researched:** 2026-08-06
**Domain:** Go C4 diagram CLI — backward-compat testing infrastructure
**Confidence:** HIGH

> **Research provenance note:** Same as Phases 1-2 — gsd-phase-researcher cannot spawn (firecrawl/exa
> MCP servers not connected; web research disabled). Orchestrator produced this RESEARCH.md directly
> from codebase analysis. All claims tagged `[VERIFIED: <file>:<line>]`.

## Summary

Phase 3 locks the v1.8 compatibility guarantees (COMPAT-01/COMPAT-02) into enforceable tests. The
compatibility MACHINERY already exists and is verified (Phases 1-2: `View.AllExpanded` exemption keeps
v1.7 dedup key + penwidth 2.0 in `--expanded` mode; OR expansion semantics for COMPAT-01). What's
missing is CI-enforceability: the two strongest compat tests
(`TestBuildExpandedGraphRealToml` builder_test.go:799, `TestBuildExpandedGraphBaselineDOT` :1208)
read the **gitignored** private fixture `cyp-auth-infra/saira-20260320.c2.full.toml` (:803, :1211)
and fail in fresh checkouts. A public synthetic fixture exists (`testdata/expanded.toml`, 51 lines)
but lacks the real-world structure (deep links, `length` attrs, 5 top-level units) the private
fixture exercises.

**Primary recommendation:** Create a sanitized public fixture with the saira STRUCTURE (deep nesting,
`length` attrs, multi-level links — generic unit names), repoint the two tests at it, and commit a
DOT golden baseline. Per D-01..D-04: DOT-level contract, OR semantics kept, `--expanded` ignores
`properties.expanded` (already true — `GenerateExpandedView` never consults it).

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- **D-01:** Public fixture — sanitized public copy of the real-world multi-level TOML (equivalent structure) in `testdata/` + committed `--expanded` DOT baseline. CI-enforceable COMPAT-02; no skip-if-missing paths. Private `cyp-auth-infra/` stays gitignored.
- **D-02:** Compat contract is DOT-level equivalence (node/edge sets, attributes, cluster structure). SVG byte-comparison NOT part of the contract.
- **D-03:** COMPAT-01 keeps OR semantics — per-unit self-references still expand in C1 without `properties.expanded`.
- **D-04:** `--expanded` ignores `properties.expanded` — flag expands everything (v1.7 contract).

### Claude's Discretion
None.

### Deferred Ideas (OUT OF SCOPE)
None.

</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| COMPAT-01 | Existing TOML files without properties.expanded generate correct C1 with all units collapsed | OR semantics verified (isExpandedInC1 scope.go:144); "no hints → collapsed" holds structurally; needs a regression test on `testdata/valid.toml` |
| COMPAT-02 | --expanded flag continues to produce single all-nested diagram unchanged (identical to v1.7) | Machinery verified (View.AllExpanded exemption; penwidth 2.0 everywhere in expanded mode); enforcement blocked on private fixture — D-01 fixes |

</phase_requirements>

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Fixture/baseline assets | `cmd/c4drill/testdata/` | — | Committed test data lives next to the CLI tests |
| COMPAT-02 enforcement | `internal/graph/builder_test.go` | `internal/render` | Baseline DOT built from expanded graph; render for DOT |
| COMPAT-01 enforcement | `cmd/c4drill/root_test.go` | `internal/view` | Full-pipeline C1 rendering of valid.toml |
| --expanded machinery | `internal/view` + `internal/graph` | — | GenerateExpandedView + BuildExpandedGraph (unchanged) |

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go stdlib | 1.26.1 | testdata reading, TOML fixtures | Existing |
| stretchr/testify | v1.11.1 | Assertions | Existing convention |

### Supporting
None — **zero new dependencies** (no golden-file library; plain `os.ReadFile` + `assert.Equal` per existing pattern).

## Architecture Patterns

### Pattern 1: Golden-file DOT comparison
`cmd/c4drill/testdata/expanded.dot` already exists (81 lines) — check whether a test consumes it
(`grep -rn "expanded.dot" cmd/c4drill/root_test.go internal/`); if yes, that's the golden-file
pattern to extend: parse fixture → build expanded graph → RenderDOT → compare against committed
baseline. D-02 contract: compare DOT text (deterministic — order preserved) or structural tokens.

### Pattern 2: Private→public fixture migration
The two tests to repoint:
- `TestBuildExpandedGraphRealToml` (builder_test.go:799): parse → validate → GenerateExpandedView →
  BuildExpandedGraph → assert client→keycloak edge MinLen preserved (:825-832) → RenderDOT contains minlen.
  Requires the fixture to have: a deep nested unit with a link carrying `length` > 0 to a sibling/top-level unit.
- `TestBuildExpandedGraphBaselineDOT` (:1208): all edges PenWidth 2.0 (:1221-1224), DOT contains minlen= and penwidth=2.
Repointing = change the fixture path constant; assertions stay. Add a committed golden DOT for the new fixture.

### Pattern 3: Sanitized structural copy
The saira fixture (88 lines, private): 5 top-level units (webUser, sshUser, adminUser, keycloak,
linuxSystem), linuxSystem with deep nesting (storages → keycloakStorage → client with
`link { peer = "keycloak", length = N }`). Sanitized public version: same STRUCTURE with generic
names (e.g., actorA/actorB/actorC/externalSys/mainSystem) — must preserve: multi-level nesting,
cross-level link with `length`, enough units to exercise the 5-node C1 assertion if reused.

## Common Pitfalls

### Pitfall 1: Deterministic DOT
**What goes wrong:** Golden-file comparison flakes if DOT ordering isn't deterministic.
**How to avoid:** Definition-order preservation is guaranteed (UnitOrder/SubunitOrder); regenerate
the baseline from the SAME pipeline and assert exact match. If any nondeterminism appears (map
iteration), sort-normalize before compare.
**Warning signs:** Golden test passes locally, fails in CI.

### Pitfall 2: Over-sanitizing loses the test's value
**What goes wrong:** A sanitized fixture that flattens the nesting no longer exercises the
deep-link resolution the private fixture was chosen for.
**How to avoid:** Keep the structural skeleton (5 top-level units, 3+ nesting levels, `length`
attrs, cross-level links); only rename identifiers.
**Warning signs:** RealToml test becomes trivial after migration.

### Pitfall 3: `--expanded` + `properties.expanded` (D-04)
**What goes wrong:** A future "helpful" change makes GenerateExpandedView consult
properties.expanded → COMPAT-02 silently breaks.
**How to avoid:** Add an explicit regression test: fixture WITH `properties.expanded` set → expanded
mode still contains ALL units (nothing collapsed).
**Warning signs:** Baseline DOT shrinks.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Golden-file comparison | A new assertion framework | `os.ReadFile` + `require.Equal` on committed baseline | Existing pattern; zero deps |

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | go test + testify |
| Config file | none |
| Quick run command | `go test ./internal/graph/ ./cmd/c4drill/` |
| Full suite command | `go test -v -race -cover ./...` (no t.Parallel on render tests) |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| COMPAT-01 | valid.toml (no properties.expanded) → C1 all collapsed | integration | `go test ./cmd/c4drill/ -run FullPipeline -count=1` | ⚠ verify/add |
| COMPAT-02 | Expanded DOT matches committed golden baseline | unit | `go test ./internal/graph/ -run BaselineDOT -count=1` | ❌ repoint fixture |
| COMPAT-02 | Expanded mode ignores properties.expanded (D-04) | unit | `go test ./internal/graph/ -run Expanded -count=1` | ❌ new |
| COMPAT-02 | minlen preserved via public fixture (RealToml) | unit | `go test ./internal/graph/ -run RealToml -count=1` | ❌ repoint fixture |

### Sampling Rate
- **Per task commit:** quick run command
- **Per wave merge:** `go test -v -race ./...`
- **Phase gate:** full suite green before verify-work

### Wave 0 Gaps
- [ ] `testdata/` sanitized public fixture (saira structure, generic names)
- [ ] Committed golden DOT baseline for the public fixture
- [ ] Repointed RealToml + BaselineDOT tests (no gitignored paths)
- [ ] D-04 regression test (expanded ignores properties.expanded)
- [ ] COMPAT-01 regression test (valid.toml all-collapsed)

## Security Domain

### Applicable ASVS Categories
| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2-V6 | no | — (local CLI, no auth/sessions/crypto) |
| V5 Input Validation | yes | Fixture must pass validator (existing pipeline) — fixtures are test data, not a new surface |

### Known Threat Patterns
| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| None new | — | Test-data only phase; no production code changes expected |

## Sources

### Primary (VERIFIED — codebase reads this session)
- `internal/graph/builder_test.go` — TestBuildExpandedGraphRealToml (:795-840, fixture path :803), TestBuildExpandedGraphBaselineDOT (:1205-1232, fixture path :1211), TestBuildEdgesExpandedExemption (:1912)
- `cmd/c4drill/testdata/` — expanded.toml (51 lines, synthetic), expanded.dot (81 lines), valid.toml, invalid.toml
- `cyp-auth-infra/` (gitignored) — saira-20260320.c2.full.toml (88 lines, real structure), expanded.dot (180 lines)
- `internal/view/scope.go` — GenerateExpandedView (:14, never consults properties.expanded), isExpandedInC1 (:144)
- `internal/graph/builder.go` — BuildExpandedGraph (:125), buildEdges AllExpanded branch (:383/:458)
- `.gitignore` — cyp-auth-infra/ untracked (commit 1ae0cb6)

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | expanded.dot (public) may or may not already be consumed by a golden test | Pattern 1 | If consumed, extend that pattern; if not, create it — planner verifies |
| A2 | The saira fixture structure is reproducible with generic names | Pattern 3 | If sanitization breaks validator rules (orphans etc.), adjust fixture links |

## Open Questions (RESOLVED)

None — all areas decided in discussion.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — zero new dependencies
- Architecture: HIGH — verified against actual code
- Pitfalls: HIGH — grounded in Phase 1/2 execution history

**Research date:** 2026-08-06
**Valid until:** 2026-09-05
