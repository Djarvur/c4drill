---
phase: 37-nesting-context-and-plain-rendering
plan: 06
subsystem: docs
tags: [docs, readme, skill, examples, c4d, plain, nesting]

# Dependency graph
requires:
  - phase: 37-nesting-context-and-plain-rendering
    provides: CTX-01..03 nesting-context rendering (37-01/37-02/37-05) and PLAIN-01..04 --plain rendering (37-03/37-04) — the implemented behavior this plan documents
provides:
  - "DOC-01: README.adoc nesting-context section (ancestor chains, deep-link targets, expanded nested clusters, boundary) + --plain CLI reference with the exact ignored-vs-stays lists and example invocations"
  - "DOC-02: --plain and nesting-context documented in skill/SKILL.md, byte-identical across all three copies (diff -r clean)"
  - "DOC-03: skill/examples/11-nesting-context.toml (+ .c4d twin + rendered SVGs) and 12-plain.toml fixtures, CI-sweep clean, mirrored byte-identical into both plugin copies"
affects: [37-07 (release/tag), verifier]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Docs+fixtures land in one plan behind CI gates (v1.12/v1.13 precedent): diff -r parity on pristine checkout + full example render sweep"

key-files:
  created:
    - skill/examples/11-nesting-context.toml
    - skill/examples/11-nesting-context.c4d
    - skill/examples/11-nesting-context.svg
    - skill/examples/11-nesting-context.expanded.svg
    - skill/examples/11-nesting-context/platform.svg
    - skill/examples/11-nesting-context/platform/gateway.svg
    - skill/examples/11-nesting-context/platform/gateway/pipeline.svg
    - skill/examples/11-nesting-context/platform/data.svg
    - skill/examples/11-nesting-context/platform/data/store.svg
    - skill/examples/12-plain.toml
  modified:
    - README.adoc
    - skill/SKILL.md
    - plugins/c4drill/skills/c4drill-toml/SKILL.md
    - plugins/c4drill/opencode/skills/c4drill-toml/SKILL.md

key-decisions:
  - "Nesting-context README section placed after the TOML Nesting section; --plain documented in the CLI Reference beside --expanded (flag block + bullet with nested ignored/stays lists)"
  - "Example 11 demonstrates all three CTX behaviors in one model: deep-link target at level 3 (admin -> platform.gateway.router), a componentBox grouping nested inside the expanded gateway (nested cluster), and a cross-branch deep link (validator -> data.store.index)"
  - "Example 12 mirrors cmd/c4drill/testdata/plain.toml intent as a user-facing fixture: unit color/style/border, link color/style/rank/labelPosition, properties.edges=ortho — verified stripped under --plain (0 custom hexes in plain SVG vs 5 in default) while the legend stays"
  - "SVG artifacts tracked with git add -f (skill/examples/**/*.svg is gitignored; 06/10 precedent tracks example renders deliberately); 12-plain intentionally ships no SVG (04-styling precedent — styling demos stay source-only, the --plain look is CLI-driven)"

requirements-completed: [DOC-01, DOC-02, DOC-03]

# Metrics
duration: 22min
completed: 2026-08-30
---

# Phase 37 Plan 06: README + Skill Sync + Example Fixtures (DOC-01..03) Summary

**Both v1.21 features are now documented where users and agents read — README nesting-context section and the `--plain` CLI key with the exact ignored-vs-stays boundary, three byte-identical skill copies, and two new CI-clean example fixtures (11-nesting-context with .c4d twin and rendered SVGs, 12-plain) — closing the DOC-01..03 ship-gates.**

## Performance

- **Duration:** 22 min
- **Started:** 2026-08-30T17:20:00Z
- **Completed:** 2026-08-30T17:42:00Z
- **Tasks:** 3/3
- **Files:** 14 created, 4 modified

## Accomplishments
- **DOC-01 (README.adoc):** new "Nesting Context in Rendered Diagrams" section covering full ancestor chains on non-expanded views, deep-link targets terminating inside their ancestor chain, expanded units rendering nested clusters with 🔍 explore links, plus the explicit boundary (sibling/external boundary nodes stay top-level; collapsed subtrees not restructured) and a pointer to example 11. `--plain` added to the CLI Reference flag block and bullets: CLI-only (no model keys), composes with `--expanded`, convert/fmt unaffected, exact ignored list (unit color/style/border incl. expanded clusters; link color/style; length and rank; properties.edges; label formatting) and stays list (kind colours, legend, queue pipes, 🔍/📖 glyphs, collapsed subtrees), with example invocations referencing `skill/examples/12-plain.toml`.
- **DOC-02 (skill sync):** SKILL.md gains a nesting-context note in the Nesting section and a "Plain Rendering (--plain)" section in the CLI area (authoring-focused: the model file gains no new keys), plus examples 10–12 in the example index. Copied byte-identical to both plugin copies; `diff -r skill plugins/...` silent both ways.
- **DOC-03 (fixtures):** `11-nesting-context.toml` — 3-level hierarchy (system → containerBox/container → component) with the deep-link target `platform.gateway.router`, a `componentBox` "Processing Pipeline" nested inside the expanded gateway (nested-cluster demo), and a cross-branch deep link `validator -> platform.data.store.index`; renders cleanly (C1 + expanded + 5 drill-down SVGs committed, twin generated via `convert to-c4d` + `fmt` and render-verified). `12-plain.toml` — custom-formatting fixture rendered both ways: default SVG carries the fixture hexes (5 hits), `--plain` SVG carries zero while the legend remains. CI-style sweep (every `skill/examples/*.toml` and `*.c4d` through `-f dot`) exits clean; new files mirrored into both plugin trees.

## Task Commits

1. **Task 1: README nesting-context + --plain (DOC-01)** - `5a2cbc2` (docs)
2. **Task 2: skill + both plugin copies synced (DOC-02)** - `e27ddc6` (docs)
3. **Task 3: example fixtures + mirrors (DOC-03)** - `cea1467` (feat)

## Files Created/Modified
- `README.adoc` — nesting-context section + --plain CLI reference (modified)
- `skill/SKILL.md`, `plugins/c4drill/skills/c4drill-toml/SKILL.md`, `plugins/c4drill/opencode/skills/c4drill-toml/SKILL.md` — plain/nesting docs + example index, three byte-identical copies
- `skill/examples/11-nesting-context.toml` + `.c4d` twin + `.svg`/`.expanded.svg` + drill-down SVG tree — nesting demo (created)
- `skill/examples/12-plain.toml` — --plain demo fixture (created)
- Both plugin `examples/` trees — all new files mirrored byte-identical

## Decisions Made
- Example 11 combines all three CTX behaviors in one small model rather than three separate fixtures — one render demonstrates chains, deep links, and nested clusters together.
- `12-plain.toml` ships without committed SVGs: styling-focused examples (04-styling precedent) stay source-only, and the plain look is produced by the CLI flag, not the source file.
- Tracked example SVGs re-synced per the v1.19/v1.20 convention: generated from the final fixture with the production CLI and force-added past the `skill/examples/**/*.svg` gitignore rule (06/10 precedent).

## Deviations from Plan

None - plan executed exactly as written. (The plan's Task 3 step 1 SVG convention question resolved to "commit": 06-templates and 10-edge-kinds both track SVG artifacts.)

## Issues Encountered
- First draft of example 11 left `parser` an orphan (no link) — validator caught it; fixed with a `parser -> validator` link.
- First draft of example 12 linked `customer -> shop.web`, but `shop.web` has subunits (validation rule 3) — retargeted to `shop.web.catalog`. Both caught by the render pipeline before commit, as designed.
- `08-include` component fragments (auth.toml, billing.toml) fail when rendered standalone — expected: they are include fragments resolving peers from the entry file, and CI's glob (`skill/examples/*.toml`, non-recursive) never renders them individually.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- DOC-01..03 complete; all Phase 37 requirements (CTX-01..03, PLAIN-01..04, BC-01, DOC-01..03) now landed — 37-07 (release/tag, REL-01) can proceed.
- CI gates verified locally pre-push: both `diff -r` parity checks silent; full example sweep green.

---
## Self-Check: PASSED

- All created files exist on disk (fixtures, twin, SVGs — verified via `find`/`ls`)
- Commits verified in `git log`: `5a2cbc2` (docs DOC-01), `e27ddc6` (docs DOC-02), `cea1467` (feat DOC-03)
- Plan-level verification re-run: README greps `--plain`=4 and nesting/ancestor=8 (both ≥1); `diff -r` both plugin copies silent; CI-style sweep over `skill/examples/*.toml` + `*.c4d` clean; `--plain -f dot` render exits 0; plain-vs-default hex spot-check passes (0 vs 5 custom-hex hits, legend present)
- No deletions in any commit (`git diff --diff-filter=D` empty)

---
*Phase: 37-nesting-context-and-plain-rendering*
*Completed: 2026-08-30*
