# Phase 41: Check Command - Context

**Gathered:** 2026-09-03
**Status:** Ready for planning

<domain>
## Phase Boundary

A new `c4drill check <file>` subcommand that runs the full model validation without rendering or writing anything — exit 0 when valid, non-zero with render-identical errors when not (issue #41). No new validation rules, no changes to the render path's behavior.

</domain>

<decisions>
## Implementation Decisions

### Command shape
- **D-01:** New cobra subcommand `check` (sibling of `fmt`/`convert`/`serve` in cmd/c4drill), taking exactly one input file argument. NOT a root flag — the root command already renders by default; the issue explicitly asks for a command. Extension dispatch comes free via the shared `parseInput` helper (.toml/.c4d, unknown extensions fail closed).

### Pipeline reuse
- **D-02:** `check` runs the exact render front-half, stages 1→2, in the same order with the same error prefixes: `parseInput` → `include.Resolve` → `template.Expand` → `peer.Resolve` → `validator.Validate` (cmd/c4drill/root.go runRoot:186-219). Implementation should extract/share this front-half so check and render cannot drift (a shared helper on root.go, not a copy).
- **D-03:** Validation failures print via `validator.ReportErrors` to stderr exactly as render does, then exit with the same `errValidationFailed` path (exit 1). Parse/include/expand/peer failures surface as the same stage-prefixed errors render produces ("parse: ...", "include: ...", etc.). Render-identical output is the contract (CHECK-02/03).

### Output behavior
- **D-04:** Silent success — valid model prints nothing, exits 0 (unix convention for CI/`make` gates; matches render's "silent per spec" success). No summary line, no verbosity flag in this phase.

### Flag surface
- **D-05:** `check` carries NO render flags (no output dir, no format, no --plain/--edges/--label-ratio). They are meaningless without rendering; keeping the surface minimal prevents implying check produces output.

### Docs
- **D-06:** README gains a `check` entry beside the existing command docs: usage, both files formats, exit codes (0 valid / 1 invalid), and the no-output guarantee. skill/SKILL.md command table updated if it lists commands (planner verifies).

### Claude's Discretion
- How exactly to factor the shared front-half (extract function vs check calling into runRoot with a dry-run switch) — planner picks the cleanest seam.
- Test fixture placement (cmd/c4drill/testdata precedent exists).

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Issue
- GitHub issue #41 (Djarvur/c4drill) — motivation, `fmt --check` gap demo (formats-clean-but-invalid model), CI use case

### Code
- `cmd/c4drill/root.go` §runRoot (lines ~150-245) — the stage pipeline check must mirror; `parseInput` extension dispatch; `errValidationFailed`
- `cmd/c4drill/fmt.go` — existing single-purpose subcommand precedent (incl. its own --check flag semantics)
- `internal/validator/` — `Validate(m)` + `ReportErrors` (the render-identical error surface)

### Project docs
- `.planning/REQUIREMENTS.md` — CHECK-01..04 definitions
- `README.md` — CLI surface section where `check` will be documented

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `parseInput` helper — extension dispatch already shared by root; check reuses it directly
- `validator.Validate` + `validator.ReportErrors` — the exact error surface render uses
- `errValidationFailed` sentinel — established exit-1 path

### Established Patterns
- Cobra subcommand-per-verb layout (fmt.go, convert.go, serve.go) with `RunE` returning errors
- Stage-prefixed error wrapping in runRoot ("parse:", "include:", "expand:", "resolve peers:")

### Integration Points
- `cmd/c4drill/root.go` — shared front-half extraction point
- `README.md` + `skill/SKILL.md` — documentation targets (CHECK-04)

</code_context>

<specifics>
## Specific Ideas

From the issue: the driving use case is CI/edit-loop validation of repo-checked `.c4d` sources without handing over an output directory — keep `check` invocable as `c4drill check file.c4d` with nothing else required.

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope. (A future `--json` output for tooling was considered and rejected for this phase: no consumer exists yet.)

</deferred>

---

*Phase: 41-Check Command*
*Context gathered: 2026-09-03*
