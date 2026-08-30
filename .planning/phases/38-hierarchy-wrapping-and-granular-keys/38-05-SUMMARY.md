---
phase: 38-hierarchy-wrapping-and-granular-keys
plan: 05
subsystem: docs
tags: [readme, skill, plugin-sync, examples, wrapping, granular-keys, c4d-twins]
requires:
  - phase: 38-01
    provides: "wrapper clusters (wrap_*) with pretty AncestorNames labels, no explore URL"
  - phase: 38-02
    provides: "--no-colors/--no-styles/--no-length/--no-rank semantics and the pinned NoColors && !Plain interaction"
  - phase: 38-03
    provides: "--no-labels graph-layer suppression with legend-stays pin"
  - phase: 38-04
    provides: "composition matrix + goldens proving the documented behaviour"
provides:
  - "README.adoc v1.22 docs: ancestor wrapping (reversed v1.21 boundary) + all five granular keys with pinned boundaries"
  - "skill/SKILL.md + both plugin copies byte-identical (CI diff -r parity)"
  - "skill/examples/13-wrapping.toml fixture rendering clean through the pipeline"
affects: []
tech-stack:
  added: []
  patterns:
    - "single-commit 3-copy skill sync (edit skill/, propagate to both plugin trees)"
key-files:
  created:
    - skill/examples/13-wrapping.toml
  modified:
    - README.adoc
    - skill/SKILL.md
    - plugins/c4drill/skills/c4drill-toml/SKILL.md
    - plugins/c4drill/opencode/skills/c4drill-toml/SKILL.md
decisions:
  - "Docs use the real release tag v1.22 (git tag convention: v1.21.0 documented phase 37); the plan's 'v1.15' is GSD milestone numbering, not a user-facing version"
  - "13-wrapping.toml ships WITHOUT a .c4d twin, following the 12-plain.toml precedent — it is a CLI-flag demo (a twin adds file-format equivalence noise, and the expectedExampleTwins manifest in internal/c4d/parity_test.go stays untouched)"
  - "Fixture uses a container (core.audit) rather than a box so the sibling boundary component passes C2/C3 validation"
requirements-completed: [DOC-01, DOC-02, DOC-03]
metrics:
  duration: ~30m
  completed: 2026-08-30
---

# Phase 38 Plan 05: Documentation — Wrapping + Granular Keys Summary

README.adoc and the skill (synced byte-identical across its three copies) now document the v1.22 ancestor-wrapping reversal and every granular key with the pinned boundaries, and a new `13-wrapping.toml` fixture proves wrapping plus all five keys render cleanly through the production CLI.

## Performance

- **Duration:** ~30 min
- **Completed:** 2026-08-30
- **Tasks:** 2
- **Files modified:** 5 (plus 2 pre-existing 38-03 test files landed first)

## Accomplishments

- **Task 1 — README (DOC-01, commit 5d5580d):** the "Nesting Context" section's v1.21 boundary paragraph replaced by an "Ancestor Wrapping (v1.22)" subsection — every depicted node (regular, boundary, expanded) renders inside its full ancestor-container chain via pretty-named wrapper clusters (no 🔍, no explore URL), only fully-external units stay top-level, collapsed subtrees untouched. CLI flags table gained all five flags with verbatim help text from `cmd/c4drill/root.go`. New granular-switches bullet block: per-flag semantics, the three pinned boundaries (`--no-colors` takes kind-derived colours too with the structural D-01 default border surviving and legend rows losing swatches; legend STAYS under `--no-labels` as `properties.legend` metadata; `properties.edges` tied to `--plain` only), the `NoColors && !Plain` interaction (`--plain` alone keeps semantic kind colours, so `--plain` ≡ union and `--plain --no-colors` removes them), and an `--expanded` note for `--no-labels`.
- **Task 2 — skill sync + fixture (DOC-02/03, commit a3dd132):** SKILL.md nesting paragraph extended with the v1.22 wrapping rule; new "Granular Render Flags (--no-*)" section with the same pinned boundaries; example 13 added to the fixture list. `skill/examples/13-wrapping.toml`: external person (stays top-level), expanded system with container/box chains, cross-branch sibling boundary + custom-formatted kind/rank/length/color/style links. Rendered through the production CLI — C2 view of `core.app` shows `cluster_wrap_core` labelled "Core Platform" (pretty name, no URL) wrapping the expanded cluster and the `core.audit` boundary entry; `--no-colors --no-rank` drops the kind hex and rank markers; `--no-labels` emits `label=""` with the legend intact. Both plugin copies synced in the SAME commit; `diff -r` deltas are exclusively gitignored local render artifacts (tracked-content parity exact; CI diffs pristine checkouts).
- **Pre-task hygiene:** landed the residual 38-03 test-precision fixes that its SUMMARY described as committed but were still uncommitted (commit 4238dc8), so the docs plan starts from a clean tree.

## Task Commits

1. **Pre-task: residual 38-03 test fixes** - `4238dc8` (test)
2. **Task 1: README wrapping + keys** - `5d5580d` (docs)
3. **Task 2: skill 3-copy sync + 13-wrapping fixture** - `a3dd132` (docs)

## Verification Evidence

- `diff -r skill plugins/c4drill/skills/c4drill-toml` / opencode copy — all remaining deltas are `skill/examples/**/*.dot|svg|drill-down-dir` entries matching `.gitignore:71`; zero tracked-content differences.
- `go run ./cmd/c4drill skill/examples/13-wrapping.toml -f dot` — clean (CI's exact validation loop covers it).
- `go test -count=1 ./...` — green (2 consecutive clean full-suite runs; see Issues for one flake occurrence).
- README flag count: all five `--no-*` flags present ≥2 occurrences each (flags table + semantics + examples).

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] core.audit as box failed C2/C3 validation**
- **Found during:** Task 2
- **Issue:** `type = "box"` directly under a system requires C2-type children; the planned sibling-boundary component `core.audit.trail` failed validation
- **Fix:** `core.audit` is a `container`; trail is its C3 component — same wrapping story
- **Files modified:** skill/examples/13-wrapping.toml
- **Committed in:** a3dd132

**2. [Rule 1 - Docs] Version numbering**
- **Found during:** Task 1
- **Issue:** plan says "v1.15"; README/skill document user-facing release tags (phase 37 = v1.21.0)
- **Fix:** documented as v1.22 throughout
- **Files modified:** README.adoc, skill/SKILL.md (+ both plugin copies)
- **Committed in:** 5d5580d, a3dd132

**Total deviations:** 2 auto-fixed (1 Rule 3, 1 Rule 1).
**Impact on plan:** none on behaviour or scope.

## Issues Encountered

- `TestIntegrationC1EdgeResolution` (internal/view) failed once in four full-suite runs — the exact parallel-scheduling flake 38-03 documented; passes alone and in all repeat runs. Docs-only plan adds no Go source; treated as out-of-scope pre-existing flake.
- rsync of skill/ into plugin trees briefly copied gitignored render artifacts into the plugin examples; the rsync-introduced files were removed by name before commit (plugin trees deliberately do not track them).

## Known Stubs

None.

## Threat Flags

None — docs-only; T-38-06 mitigated: the new fixture was rendered locally through the production CLI before commit (RENDER-OK evidence above).

## Self-Check: PASSED

- Files exist: README.adoc, skill/SKILL.md, skill/examples/13-wrapping.toml, both plugin SKILL.md copies — FOUND
- Commits: 4238dc8, 5d5580d, a3dd132 — FOUND in `git log`
- Both `diff -r` checks: zero tracked-content deltas
- `go test -count=1 ./...` — green (×2 consecutive clean runs)
