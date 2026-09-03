# Phase 41: Check Command - Research

**Researched:** 2026-09-03
**Domain:** Go CLI subcommand (cobra) reusing an existing validation pipeline — no new libraries, no external services
**Confidence:** HIGH

## Summary

Phase 41 adds `c4drill check <file>`: a render-free validator that runs the exact render front-half (parse → include resolve → template expand → peer resolve → validate) and reports exactly what render reports. The codebase already contains everything needed, and — critically — the front-half check needs **already exists as a working copy**: `validateSourceForConvert` (cmd/c4drill/convert.go:175-201) runs parseInput → include.Resolve → template.Expand → peer.Resolve → validator.Validate with runRoot's exact stage-prefixed error wrapping ("parse: ", "include: ", "expand: ", "resolve peers: "). D-02 asks for a shared helper on root.go so check and render cannot drift; the refactor seam is to promote that stage sequence from runRoot (cmd/c4drill/root.go:176-223) into one shared function both runRoot and the new check command call.

Exit-code behavior is trivially correct with the established pattern: every subcommand surfaces failure via a `RunE` error, and `main()` exits 1 on any error and 0 implicitly (cmd/c4drill/main.go:5-10). Render already reports validation failures exactly this way: `validator.ReportErrors(valErrors, cmd.OutOrStderr())` then `return errValidationFailed` (root.go:218-223), where `errValidationFailed` is the package sentinel (root.go:30). Silent success (CHECK-01/D-04) matches render's own "Success - silent per spec" convention (root.go:246).

Testing is well-precedented: cobra commands are tested in-package by constructing `NewRootCmd()`-style commands, `SetOut`/`SetErr`, `SetArgs`, `Execute()` (fmt_test.go `runFMTTest` helper at fmt_test.go:293-301; root_test.go throughout). Test fixtures live in `cmd/c4drill/testdata/` (valid.toml = valid model, invalid.toml = parse-level failure). One fixture gap exists: no fixture in cmd/c4drill/testdata fails at the *validation* stage (VAL rules) — invalid.toml is a parse error, and reference failures with bare peers are now caught earlier by peer.Resolve (root_test.go:231). The orphan-unit rule (validator.ValidateOrphanUnits, rules.go:125-142) gives the cleanest validation-stage failure: a unit with no links and no subunits.

**Primary recommendation:** Extract runRoot's stages 1→2 into one shared `root.go` helper (parse → include → expand → peers → validate + ReportErrors + errValidationFailed), call it from runRoot and from a new `check.go` cobra subcommand (ExactArgs(1), no render flags of its own), pin behavior TDD-first with valid / parse-invalid / validation-invalid / composed multi-file fixtures, and document `check` in README.adoc's command reference.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- **D-01:** New cobra subcommand `check` (sibling of `fmt`/`convert`/`serve` in cmd/c4drill), taking exactly one input file argument. NOT a root flag — the root command already renders by default; the issue explicitly asks for a command. Extension dispatch comes free via the shared `parseInput` helper (.toml/.c4d, unknown extensions fail closed).
- **D-02:** `check` runs the exact render front-half, stages 1→2, in the same order with the same error prefixes: `parseInput` → `include.Resolve` → `template.Expand` → `peer.Resolve` → `validator.Validate` (cmd/c4drill/root.go runRoot:186-219). Implementation should extract/share this front-half so check and render cannot drift (a shared helper on root.go, not a copy).
- **D-03:** Validation failures print via `validator.ReportErrors` to stderr exactly as render does, then exit with the same `errValidationFailed` path (exit 1). Parse/include/expand/peer failures surface as the same stage-prefixed errors render produces ("parse: ...", "include: ...", etc.). Render-identical output is the contract (CHECK-02/03).
- **D-04:** Silent success — valid model prints nothing, exits 0 (unix convention for CI/`make` gates; matches render's "silent per spec" success). No summary line, no verbosity flag in this phase.
- **D-05:** `check` carries NO render flags (no output dir, no format, no --plain/--edges/--label-ratio). They are meaningless without rendering; keeping the surface minimal prevents implying check produces output.
- **D-06:** README gains a `check` entry beside the existing command docs: usage, both files formats, exit codes (0 valid / 1 invalid), and the no-output guarantee. skill/SKILL.md command table updated if it lists commands (planner verifies).

### Claude's Discretion
- How exactly to factor the shared front-half (extract function vs check calling into runRoot with a dry-run switch) — planner picks the cleanest seam.
- Test fixture placement (cmd/c4drill/testdata precedent exists).

### Deferred Ideas (OUT OF SCOPE)
None — discussion stayed within phase scope. (A future `--json` output for tooling was considered and rejected for this phase: no consumer exists yet.)
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| CHECK-01 | `c4drill check <file>` validates without producing output files or requiring an output directory | check defines no `-o` flag use, never constructs `output.NewWriter`; pipeline helper stops at Validate (root.go:176-223 seam) |
| CHECK-02 | Exit 0 valid / non-zero invalid with render-identical errors (e.g. orphan-unit VAL rules) | Shared helper reuses `validator.ReportErrors` + `errValidationFailed` (root.go:218-223); main() exits 1 on error (main.go:5-10); orphan fixture pins the VAL path |
| CHECK-03 | Same pipeline front-half as render — includes resolved, templates expanded, peers resolved | Shared helper extracted from runRoot:176-223 (precedent: `validateSourceForConvert` convert.go:175-201 already replicates this sequence); composed multi-file fixture proves parity |
| CHECK-04 | README documents `check` | README.adoc command reference has per-command sections (`=== convert` :1417, `=== fmt` :1452, `=== serve` :1474) — `check` gets a sibling section; skill/SKILL.md usage list (:156-165) gains matching entries |
</phase_requirements>

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| CLI arg/flag parsing for `check` | cmd/c4drill (cobra) | — | cobra subcommand, sibling of fmt/convert/serve (D-01) |
| Extension dispatch (.toml/.c4d) | cmd/c4drill `parseInput` | — | Already shared by root + convert (root.go:249-265); fails closed (D-27) |
| Include/template/peer resolution | internal/include, internal/template, internal/peer | — | Pipeline stages 1a/1.5/1.6, called in runRoot order (root.go:186-215) |
| Semantic validation + error formatting | internal/validator | — | `Validate` + `ReportErrors` are the render error surface (validator.go:29-66) |
| Process exit code | cmd/c4drill main() | — | `os.Exit(1)` on any RunE error (main.go:5-10) |
| Docs (usage, exit codes) | README.adoc, skill/SKILL.md | — | CHECK-04 targets |

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| github.com/spf13/cobra | v1.10.2 (go.mod, already vendored via go.sum) | Subcommand registration, arg validation (`cobra.ExactArgs(1)`) | Every existing subcommand uses it [VERIFIED: go.mod] |
| github.com/stretchr/testify | v1.12.1 (go.mod) | Test assertions (require/assert) | Used by every *_test.go in cmd/c4drill [VERIFIED: go.mod] |
| Go standard library | go 1.26.1 toolchain (installed: 1.26.5) | errors sentinel wrapping, filepath | Project-wide convention [VERIFIED: go.mod + `go version`] |

### Supporting
None new. Everything check needs is already an internal package: `internal/parser`, `internal/c4d`, `internal/include`, `internal/template`, `internal/peer`, `internal/validator`.

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Shared helper on root.go (D-02) | check calling `runRoot` with a dry-run switch | Rejected by CONTEXT.md discretion note as messier: runRoot reads render-only globals (format, outputDir, expanded, render.LabelRatio) that check must never touch |
| Reuse `validateSourceForConvert` as-is | Extract new shared helper both call | convert's copy uses `cmd.ErrOrStderr()` while runRoot uses `cmd.OutOrStderr()` (convert.go:195 vs root.go:220) — it is NOT render-identical in writer selection; migrating convert is out of scope and would change convert's tested behavior |

**Installation:** None — no new packages.

**Version verification:** `go version` → go1.26.5 darwin/arm64 (toolchain satisfies go.mod's 1.26.1). Dependency versions pinned in go.mod; no registry lookups needed for an add-no-dependency phase.

## Package Legitimacy Audit

Not applicable — this phase installs zero external packages. All dependencies (cobra, testify) are existing go.mod requirements already in use by the code under modification. [VERIFIED: go.mod, go.sum]

## Architecture Patterns

### System Architecture Diagram

```
                    c4drill check <file>          c4drill <file>  (render)
                            │                             │
                            ▼                             ▼
                  ┌─────────────────────────────────────────────┐
                  │   SHARED FRONT-HALF HELPER (root.go)        │
                  │   1. parseInput        (.toml→TOML parser,  │
                  │    │                    .c4d→C4D parser,    │
                  │    │                    else hard error)    │
                  │   2. include.Resolve   (merge [[include]]   │
                  │    │                    graph)              │
                  │   3. template.Expand   ([[use]] → units)    │
                  │   4. peer.Resolve      (bare peers →        │
                  │    │                    absolute paths)     │
                  │   5. validator.Validate                     │
                  │    ├── errors: ReportErrors → stderr,       │
                  │    │            return errValidationFailed  │
                  │    └── ok: return validated *parser.Model   │
                  └──────────────┬───────────────┬──────────────┘
                                 │               │
                    check: return nil         render: stages 3-6
                    (silent, exit 0)          (views, graph, render, write)
```

### Recommended Project Structure

```
cmd/c4drill/
├── root.go            # shared front-half helper extracted here (D-02: "on root.go, not a copy")
├── check.go           # NEW: newCheckCmd + runCheck (fmt.go single-verb precedent)
├── check_test.go      # NEW: TDD tests (valid / invalid / composed fixtures)
├── testdata/
│   ├── valid.toml     # exists — valid model (happy path)
│   ├── invalid.toml   # exists — parse-stage failure fixture
│   └── check_*.toml   # NEW: validation-stage failure + composed multi-file fixtures
├── README.adoc        # (repo root) gains `=== check` beside `=== fmt` (:1452)
└── skill/SKILL.md     # (repo skill/) usage list gains check entries (:156-165)
```

### Pattern 1: Extract-Don't-Copy Shared Pipeline Helper
**What:** Promote runRoot's stage sequence (root.go:176-223) into one function on root.go; runRoot and runCheck both call it.
**When to use:** Exactly this phase's D-02. The codebase already proved the sequence is duplicable — convert.go's `validateSourceForConvert` is a third live copy; the shared helper makes check/render share ONE implementation so prefixes, order, and writer cannot drift.
**Behavior the helper must preserve (render-identical contract, D-03):**
- Stage order: parse → include → expand → peers → validate (root.go:180-218)
- Error prefixes: `parse: `, `include: `, `expand: `, `resolve peers: ` (root.go:182,194,206,214)
- Validation reporting: `validator.ReportErrors(valErrors, cmd.OutOrStderr())` then `return errValidationFailed` (root.go:218-223) — note runRoot uses `cmd.OutOrStderr()` (resolves to os.Stderr in production since main never sets the out writer); the shared helper keeps runRoot's writer so behavior is byte-identical for both callers.
- `include.Resolve` needs the input's directory: `filepath.Dir(inputPath)` (root.go:193)
- Helper returns the validated `*parser.Model` so runRoot continues to stages 3-6 and runCheck simply returns nil.

### Pattern 2: Single-Purpose Subcommand (fmt.go precedent)
**What:** `newCheckCmd()` returning a `&cobra.Command{Use: "check <file.toml|file.c4d>", Args: cobra.ExactArgs(1), RunE: runCheck, SilenceUsage: true}`, registered in NewRootCmd via `cmd.AddCommand(newCheckCmd())` (root.go:138-146 registration block).
**When to use:** All existing verbs follow this layout; `SilenceUsage: true` prevents usage spam on validation failures (consistent with root.go:112, fmt.go:64).

### Pattern 3: In-Package Cobra Command Tests
**What:** Build the command, `SetOut/SetErr` buffers, `SetArgs`, `Execute()`, assert on error identity and buffer contents (fmt_test.go:293-301 helper; root_test.go:23-45).
**When to use:** Every check behavior: exit-error for invalid, nil for valid, stderr contents match render's output for the same fixture, stdout empty on success.
**Example:**
```go
// Source: cmd/c4drill/fmt_test.go runFMTTest helper pattern
cmd := NewRootCmd() // registers check too — test through the root for argv parity
var out, errBuf bytes.Buffer
cmd.SetOut(&out)
cmd.SetErr(&errBuf)
cmd.SetArgs([]string{"check", "testdata/valid.toml"})
err := cmd.Execute()
require.NoError(t, err)
assert.Empty(t, out.String()) // D-04 silent success
```

### Anti-Patterns to Avoid
- **Copying the pipeline into check.go:** violates D-02; creates the exact drift the phase exists to kill (convert.go already drifted on the error writer — ErrOrStderr vs OutOrStderr).
- **Moving `render.LabelRatio = getLabelRatio()` into the shared helper:** it is render-only global state (root.go:168-169); the helper must start at `parseInput` so check never touches render config.
- **Using a bare-peer reference as the "invalid" fixture:** bare peers are caught by `peer.Resolve` (stage 1.6) BEFORE validation — root_test.go:231 documents this. To pin the validation stage (CHECK-02's orphan-unit example), the fixture must fail a VAL rule: an orphan unit (no Links, no LinksFrom, no Subunits — rules.go:125-142).
- **Adding flags to check "for symmetry":** D-05 forbids it. Root `PersistentFlags` are still *inherited-parseable* by cobra subcommands (same is already true of fmt/convert/serve today), so hide them from check's help output if the planner wants a minimal documented surface; do not redeclare or shadow them.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Extension dispatch | New switch in check.go | `parseInput` (root.go:249-265) | D-27 fail-closed dispatch, already shared, names accepted extensions in the error |
| Include graph merge | Custom file walker | `include.Resolve(m, dir, path)` | Recursive merge with load-bearing ordering guarantees (root.go:186-192) |
| Template expansion | Custom instantiation | `template.Expand(m)` | Produces structurally hand-authored-equivalent models (root.go:197-203) |
| Peer path resolution | Custom rewrite | `peer.Resolve(m)` | Hard-error semantics for missing peers (root.go:209-215) |
| Validation + formatting | Custom message printer | `validator.Validate` + `validator.ReportErrors` | The render-identical error surface, includes "N errors found" summary (validator.go:29-66) |
| Exit codes | Custom os.Exit in check | `RunE` error return + main() exit 1 | Established path; errValidationFailed sentinel exists (root.go:30) |

**Key insight:** check is a *composition* of six existing, individually tested building blocks. The only new logic is the cobra command wrapper and the extraction boundary of the shared helper.

## Runtime State Inventory

Omitted — not a rename/refactor/migration phase. No stored data, service config, OS state, secrets, or build artifacts carry any string this phase changes.

## Common Pitfalls

### Pitfall 1: Validation-Failure Fixture That Never Reaches Validation
**What goes wrong:** Using a bare-peer reference fixture (like root-level `testdata/invalid_references.toml`) to pin CHECK-02 — the failure surfaces as `resolve peers: ...`, not a VAL rule.
**Why it happens:** Pipeline order moved peer misses ahead of validation in Phase 30 (root_test.go:228-232).
**How to avoid:** Orphan-unit fixture: a model where every unit has zero links — `ValidateOrphanUnits` fires with `unit "X" has no incoming or outgoing links` (rules.go:125-142).
**Warning signs:** Test error string starts with `resolve peers:` instead of `error: unit`.

### Pitfall 2: Writer Divergence (OutOrStderr vs ErrOrStderr)
**What goes wrong:** check reports validation errors to a different stream than render, breaking the render-identical contract under `2>/dev/null` CI capture.
**Why it happens:** convert.go:195 uses `cmd.ErrOrStderr()`; runRoot uses `cmd.OutOrStderr()` (root.go:220). In production both land on os.Stderr, but under cobra `SetOut` in tests they diverge.
**How to avoid:** The shared helper hardcodes runRoot's call; check gets identical behavior for free. Do not "fix" convert in this phase (out of scope, would change convert's tests).
**Warning signs:** A test asserting on `SetOut` buffer passes for render but fails for check.

### Pitfall 3: Helper Boundary Too Wide (Dragging Render Globals Along)
**What goes wrong:** Extracting a helper that starts before flag validation or includes `render.LabelRatio = getLabelRatio()` — check would then read/require render flags (violates D-05) or mutate render globals.
**Why it happens:** Copying runRoot wholesale instead of slicing stages 1→2.
**How to avoid:** Helper input is `(cmd *cobra.Command, inputPath string)`, first statement is `parseInput`; label-ratio assignment stays in runRoot (root.go:169).
**Warning signs:** `check` source references `format`, `outputDir`, `labelRatio`, or `render.` symbols.

### Pitfall 4: Forgetting README.adoc Is AsciiDoc, Not Markdown
**What goes wrong:** Writing a `### check` markdown section that doesn't match the file's `=== check` AsciiDoc section style.
**Why it happens:** CONTEXT.md and REQUIREMENTS.md say "README" generically.
**How to avoid:** The file is `README.adoc`; existing command sections are `=== convert` (:1417), `=== fmt` (:1452), `=== serve` (:1474) — follow that style, including the `====` sub-blocks and backtick usage they use.
**Warning signs:** Diff on README.adoc containing `#`/`##` headings.

### Pitfall 5: Cobra Inherited Persistent Flags in `check --help`
**What goes wrong:** `c4drill check --help` lists `-o/--output`, `-f/--format`, `--expanded`, etc. (root PersistentFlags are inherited by all subcommands), contradicting D-05's minimal documented surface.
**Why it happens:** cobra inherits root PersistentFlags onto every subcommand automatically — this already happens to fmt/convert/serve today.
**How to avoid:** Either (a) accept inherited parse but hide from help: iterate `checkCmd.InheritedFlags()` and set `Hidden = true` for the render flags, or (b) document the inherited-but-unused flags as harmless (existing precedent). Planner picks per D-05's intent; hiding is the stronger reading.
**Warning signs:** `check --help` output containing `--label-ratio`.

## Code Examples

Verified from the codebase (all paths verified this session):

### The Exact Stage Sequence To Share (root.go:176-223, condensed)
```go
// Source: cmd/c4drill/root.go runRoot stages 1→2
m, err := parseInput(inputPath)                                    // "parse: "
if err != nil { return fmt.Errorf("parse: %w", err) }
if m, err = include.Resolve(m, filepath.Dir(inputPath), inputPath); err != nil {
    return fmt.Errorf("include: %w", err)                          // "include: "
}
m, err = template.Expand(m)
if err != nil { return fmt.Errorf("expand: %w", err) }             // "expand: "
if err := peer.Resolve(m); err != nil {
    return fmt.Errorf("resolve peers: %w", err)                    // "resolve peers: "
}
valErrors := validator.Validate(m)
if len(valErrors) > 0 {
    validator.ReportErrors(valErrors, cmd.OutOrStderr())
    return errValidationFailed
}
```

### Existing Third Copy Confirming The Sequence (convert.go:175-201)
```go
// Source: cmd/c4drill/convert.go validateSourceForConvert — same stages, same
// prefixes; differs only in writer (ErrOrStderr) and return contract (no model).
```

### Orphan-Unit Validation Rule (the CHECK-02 fixture target)
```go
// Source: internal/validator/rules.go:125-142
// Fires when a unit has no Links, no LinksFrom, and no Subunits:
//   Message: `unit "<path>" has no incoming or outgoing links`, Path: path
```

### Composed Multi-File Fixture Shape (CHECK-03)
```toml
# Source: internal/include/testdata/main.toml + auth.toml pattern (INC-01)
# check_main.toml:
[properties]
name = "Composed"
[user]
type = "person"
name = "User"
[[user.link]]
peer = "authService"        # peer defined in the INCLUDED file — fails peer.Resolve
technology = "HTTPS"        # if includes did not run before validation
[[include]]
path = "check_auth.toml"
# check_auth.toml: [authService] type = "system" ... (links back to user or carries
# its own link so the composed model is VALID; a second variant with the link
# removed pins the invalid case)
```

### Registration Site (root.go:138-146)
```go
// Source: cmd/c4drill/root.go NewRootCmd — append: cmd.AddCommand(newCheckCmd())
```

## State of the Art

Not applicable — internal Go CLI phase; no library landscape shifted. Cobra v1.10.2 and Go 1.26 are current as pinned in go.mod [VERIFIED: go.mod].

## Assumptions Log

| # | Claim | Section | Risk If Wrong |
|---|-------|---------|---------------|
| A1 | GitHub issue #41's contract matches CHECK-01..04 as restated in CONTEXT.md (issue body not fetched this session; CONTEXT.md from discuss-phase is treated as authoritative) | User Constraints | Low — CONTEXT.md is the locked source; if the issue adds constraints, discuss-phase already absorbed them |
| A2 | `cmd.OutOrStderr()` in production resolves to os.Stderr (main.go never calls SetOut) | Pitfall 2 | Low — verified main.go:5-10 constructs and executes without writers; cobra falls back to os.Stderr when no writer is set |

## Open Questions

1. **Hide inherited render flags from `check --help`?**
   - What we know: cobra will inherit root's render PersistentFlags onto check; D-05 says check carries no render flags.
   - What's unclear: whether "carries no flags" means "defines no flags" (weaker) or "documents/exposes none" (stronger, requires hiding inherited flags in help).
   - Recommendation: implement the stronger reading (hide inherited render flags from check's help) — it is a small loop over `InheritedFlags()`, keeps `--help` honest, and needs no cobra hacks beyond `Hidden = true`.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go toolchain | build + tests | ✓ | go1.26.5 (go.mod: 1.26.1) | — |
| graphviz (transitive, go-graphviz) | existing render tests only | ✓ (suite currently green in CI) | — | check tests never render |

**Missing dependencies with no fallback:** None.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go standard testing + testify v1.12.1 |
| Config file | none (go.mod module tests; no testlint config beyond nolint directives) |
| Quick run command | `go test ./cmd/c4drill/ -run 'TestCheck' -count=1` |
| Full suite command | `go test ./...` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| CHECK-01 | check writes nothing, needs no -o | unit (cobra Execute + fs assertion) | `go test ./cmd/c4drill/ -run 'TestCheck' -count=1` | ❌ Wave 1 task creates check_test.go |
| CHECK-02 | exit 0 valid / exit 1 invalid, render-identical VAL errors (orphan unit) | unit | same | ❌ Wave 1 task |
| CHECK-03 | front-half parity: composed multi-file fixture validates as it renders | unit | same | ❌ Wave 1 task |
| CHECK-04 | README.adoc + SKILL.md document check | source assertion | `grep -c '=== check' README.adoc` (verify step) | n/a |

### Sampling Rate
- **Per task commit:** `go test ./cmd/c4drill/ -run 'TestCheck' -count=1`
- **Per wave merge:** `go test ./...`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
None — Go test infrastructure exists and is green; new test files and fixtures are deliverables of the phase's own TDD tasks, not infrastructure gaps.

## Security Domain

### Applicable ASVS Categories (L1, block-on-high)

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | no | CLI reads local files; no auth surface |
| V3 Session Management | no | stateless CLI |
| V4 Access Control | no | no privileged operations |
| V5 Input Validation | yes | extension dispatch fails closed (root.go:249-265, D-27); validation errors report paths, never execute content |
| V6 Cryptography | no | none in scope |

### Known Threat Patterns for a local Go CLI reading repo files

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Malicious [[include]] path traversal | Tampering/Elevation | Unchanged by this phase — include.Resolve already resolves relative to the input's directory; check introduces no new file-write capability (it never writes) |
| Error-message information leak | Information Disclosure | Validator messages name unit paths only (errors.go:18-30) — same surface render already exposes; check adds nothing new |

No new attack surface: check is read-parse-report with zero writes. No high-severity threats introduced; nothing to block on.

## Sources

### Primary (HIGH confidence)
- Codebase (read this session): cmd/c4drill/root.go (runRoot:153-247, parseInput:249-265, sentinels:27-34), cmd/c4drill/convert.go (validateSourceForConvert:175-201), cmd/c4drill/fmt.go (subcommand + --check precedent), cmd/c4drill/main.go (exit codes), internal/validator/validator.go (Validate:29-50, ReportErrors:47-66), internal/validator/rules.go (ValidateOrphanUnits:125-142), internal/validator/errors.go (message formats), internal/include/testdata/ (include fixture syntax), cmd/c4drill/testdata/ (valid.toml, invalid.toml), cmd/c4drill/fmt_test.go + root_test.go (test patterns), README.adoc (command sections :1417-1535), skill/SKILL.md (usage list :156-165), go.mod, go version 1.26.5
- `.planning/REQUIREMENTS.md` CHECK-01..04, `.planning/phases/41-check-command/41-CONTEXT.md` D-01..D-06

### Secondary (MEDIUM confidence)
- None needed — no external-API claims made.

### Tertiary (LOW confidence)
- None.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — no new dependencies; existing pins read from go.mod
- Architecture: HIGH — helper seam verified against three live copies of the same stage sequence (runRoot, validateSourceForConvert, issue-#41 gap); exact line references recorded
- Pitfalls: HIGH — every pitfall traced to a specific line or test comment in the repo

**Research date:** 2026-09-03
**Valid until:** 2026-10-03 (stable internal codebase; 30 days)
