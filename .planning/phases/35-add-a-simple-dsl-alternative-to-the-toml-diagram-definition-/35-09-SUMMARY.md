---
phase: 35-add-a-simple-dsl-alternative-to-the-toml-diagram-definition
plan: 09
subsystem: docs
tags: [docs, readme, c4d, d34, d35, skill, examples, twins, plugin, convert, fmt, parity]
# Dependency graph
requires:
  - phase: 35-add-a-simple-dsl-alternative-to-the-toml-diagram-definition (Plan 07)
    provides: convert to-toml/to-c4d with --follow-includes graph rewriting (twin generator)
  - phase: 35-add-a-simple-dsl-alternative-to-the-toml-diagram-definition (Plan 08)
    provides: c4drill fmt canonical style + the fmt --check CI gate
  - phase: 35-add-a-simple-dsl-alternative-to-the-toml-diagram-definition (Plan 06)
    provides: canonicalModel/canonsrc parity helpers the twins walker reuses
provides:
  - 12 fmt-clean .c4d twins under skill/examples/ (5 single-file + 3 in 08-include + 4 in 09-composed incl. domains/auth.c4d) enforced by TestExampleTwins (model parity all pairs, render parity standalone + graphs, self-contained .c4d include graphs)
  - README.adoc == C4D Format section (D-01..D-19 syntax, 03-links side-by-side) + convert/fmt CLI Reference entries with the real flags
  - c4drill-toml skill extended IN PLACE to dual-format coverage (name kept, D-35); plugin copies + 5 manifests + openai.yaml + plugin README synced; render command accepts .c4d
affects: []
# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Twins-walker classification: entries = parse model has Includes; fragments = file is an include target of some entry (model parity only — standalone render would fail on cross-file peers); everything else standalone (model + render parity). Entry .c4d models differ from .toml pre-resolve BY DESIGN (include paths rewritten), so graph parity compares POST-include.Resolve models"
    - "Pinned-manifest anti-shrinkage for doc fixtures: the walker asserts the found pair set equals a 12-entry expectedExampleTwins list — dropping or adding a twin fails by name"

key-files:
  created:
    - skill/examples/02-nested.c4d
    - skill/examples/03-links.c4d
    - skill/examples/04-styling.c4d
    - skill/examples/06-templates.c4d
    - skill/examples/07-relative-peer.c4d
    - skill/examples/08-include/entry.c4d
    - skill/examples/08-include/auth.c4d
    - skill/examples/08-include/billing.c4d
    - skill/examples/09-composed/entry.c4d
    - skill/examples/09-composed/templates.c4d
    - skill/examples/09-composed/domains/auth.c4d
    - skill/examples/09-composed/single-file-equivalent.c4d
    - .planning/phases/35-add-a-simple-dsl-alternative-to-the-toml-diagram-definition-/35-09-deferred-items.md
  modified:
    - internal/c4d/parity_test.go
    - README.adoc
    - skill/SKILL.md
    - skill/examples/04-styling.toml
    - plugins/c4drill/opencode/skills/c4drill-toml/SKILL.md
    - plugins/c4drill/skills/c4drill-toml/SKILL.md
    - plugins/c4drill/opencode/commands/c4drill-render.md
    - plugins/c4drill/commands/c4drill-render.md
    - plugins/c4drill/README.md
    - plugins/c4drill/.plugin/plugin.json
    - plugins/c4drill/.claude-plugin/plugin.json
    - plugins/c4drill/.cursor-plugin/plugin.json
    - plugins/c4drill/.codex-plugin/plugin.json
    - plugins/c4drill/.grok-plugin/plugin.json
    - plugins/c4drill/agents/openai.yaml
    - (24 packaged .c4d twin copies under plugins/c4drill/{opencode/,}skills/c4drill-toml/examples/)

key-decisions:
  - "Twins are convert-generated then hand-finished with short # header comments (teaching value, mirroring the .toml fixtures' header style) and fmt-canonicalized — comments ride the AST verbatim, so parity is unaffected"
  - "08-include and 09-composed twins generated via convert to-c4d --follow-includes: include paths rewritten to .c4d so each .c4d graph is self-contained; the .toml graphs untouched"
  - "TestExampleTwins classifies pairs structure-driven (has-Includes = entry; include-target = fragment) instead of hardcoding paths — single-file-equivalent.c4d correctly gets full render parity (it is not an include target)"
  - "04-styling.toml formatted in place (Rule 3): the plan's exit-0 fmt gate over skill/examples/ could not hold with the pre-existing non-canonical trailing-comment spacing; the change is formatting-only (18/18 lines, comment text verbatim — the 35-08 documented normalization)"
  - "'Why TOML?' section kept, appended a dual-format pointer sentence instead of rewriting the pitch (minimal-change; full restructure logged as deferred)"
  - "Extra sync beyond the plan's file list: plugins/c4drill/skills/c4drill-toml/ and plugins/c4drill/commands/c4drill-render.md duplicate skill content (referenced by 4 of 5 manifests) — synced to keep the plugin tree coherent; the two render-command copies keep their one pre-existing front-matter difference (name: line)"

requirements-completed: [D-34, D-35]

# Metrics
duration: 15min
completed: 2026-08-14
---

# Phase 35 Plan 09: Docs and Skill Surface Summary

**README.adoc C4D Format section + convert/fmt CLI reference, 12 render-parity-enforced .c4d example twins, and the c4drill-toml skill extended in place to dual-format coverage with all plugin copies synced**

## Performance

- **Duration:** 15 min (19:35–19:50 UTC)
- **Started:** 2026-08-14T19:35:03Z
- **Completed:** 2026-08-14T19:50:06Z
- **Tasks:** 4/4 (3 auto + 1 checkpoint auto-approved under AUTO-MODE)
- **Files modified:** 51 (37 created, 14 modified)

## Accomplishments

- **Task 1 — .c4d twins with render-parity proof (D-35):** TDD RED `TestExampleTwins` (internal/c4d/parity_test.go): a walker over skill/examples/ pinned against a 12-entry manifest — found set must EQUAL the manifest (anti-shrinkage by name). Classification is structure-driven: entries (parse model has Includes) get graph parity (composedPipeline vs composedPipelineC4D, canonicalModel equality post-resolve, canonicalDOT render equality, both graphs validate clean) plus a self-containment assertion (every twin include path has the .c4d extension); fragments (include targets of some entry) get model parity; standalone files get model + render parity. Serial (go-graphviz WASM). GREEN: the 12 twins generated via `convert to-c4d` (include graphs via `--follow-includes`), hand-finished with teaching-value `#` header comments, `fmt`-canonicalized. `fmt --check skill/examples/` exits 0. Verbosity win measured: 39–63 .c4d lines vs 79–118 .toml (≈50% reduction).
- **Task 2 — README.adoc (D-34/D-35):** `== C4D Format` section between TOML Format and Full Example: brace-block form, unit headers (optional type + display name, exact TOML type keywords, `external` modifier), fields/literals (barewords vs quoting rules, triple-quoted strings, URL barewords, `#` comments, inline + one-per-line lists), the four arrows and the desc-first edge shorthand with all three label forms + trailing option blocks, relative-peer semantics pointer, template/use (nested use, template nesting, named/positional args), include (once, mixed-format graphs), semicolons/one-line blocks, reserved words, and a 03-links.toml-vs-.c4d side-by-side (verbatim from the shipped fixtures). CLI Reference gains convert (to-toml/to-c4d, validate-first, --follow-includes, -o, graph-mode semantics) and fmt (--check CI gate) entries with the exact implemented flags; usage line now `<input.toml|input.c4d>`. Intro, Why TOML?, Quick Start mention dual-format authoring. Every snippet render-verified against the real parser before writing it in (braces-required, C1-type, orphan-rule cases probed). Verify greps pass: section present, 70 c4d/C4D lines (>= 20), follow-includes, --check, zero README.md self-refs, zero markdown fences.
- **Task 3 — skill + plugin sync (D-35):** skill/SKILL.md extended IN PLACE: front-matter name kept `c4drill-toml` (compat), description/version updated (2.0.0), new "C4D Format (Compact Alternative)" section (when-to-prefer guidance + syntax cheat-sheet condensed from the README + Converting and Formatting subsection with the real convert/fmt commands), render-path examples show both extensions, examples section documents the twins + regeneration commands. Synced byte-identical to plugins/c4drill/opencode/skills/c4drill-toml/SKILL.md AND plugins/c4drill/skills/c4drill-toml/SKILL.md (the second is the shared copy 4 of 5 manifests point at — plan said "check while syncing", it duplicates skill content). All 12 twins mirrored into both packaged examples dirs (same layout incl. 09-composed/domains/). Both c4drill-render.md copies (opencode + shared, keeping their one pre-existing `name:` line difference) accept `.c4d`: usage, validate/render examples, and a Related-commands block for convert/fmt. All 5 plugin.json descriptions + agents/openai.yaml (shortDescription + default_prompt) + plugins/c4drill/README.md updated from "from TOML" to "from TOML or C4D" — names and identifiers unchanged.
- **Task 4 — checkpoint (AUTO-MODE):** ⚡ Auto-approved checkpoint. All five verification steps executed and passing: (1) README section reads correctly with all snippets render-verified; (2) verbosity win confirmed (06-templates: 109 → 43 lines); (3) `06-templates.c4d` renders dot without error; (4) convert to-c4d + fmt --check succeed (cmd corpus valid.toml in temp — see Deviations 2); (5) `09-composed/entry.c4d` self-contained include graph renders without error. Plus plugin twin presence: 12/12 in both packaged copies.
- `go test ./...` 15/15 green, golangci-lint 0 issues, gofmt clean on all touched files.

## Task Commits

Task 1 followed TDD RED -> GREEN; Tasks 2-3 are docs-only:

1. **Task 1: twins walker (RED)** — `ae82ba9` / **twins (GREEN)** — `7d97712`
2. **Task 2: README.adoc** — `ccbebdf`
3. **Task 3: skill + plugins** — `f96563a`

## TDD Gate Compliance

Task 1's `test(35-09)` commit (ae82ba9) precedes its implementation commit (7d97712). The RED was runtime-level: the walker found 0 pairs against the pinned 12-entry manifest (missing-fixture evidence, not a compile error). Tasks 2-3 are documentation-only (no behavior to drive with a failing test); their acceptance is the plan's grep/ls verifications plus the Task-1 parity suite that now enforces the twins' render equivalence (T-35-09-02 mitigation).

## Files Created/Modified

See frontmatter key-files. Artifact contract met: README.adoc contains "C4D" and the convert/fmt entries; skill/SKILL.md contains "c4d"; skill/examples/09-composed/entry.c4d (14 lines) and domains/auth.c4d (9 lines) exceed min_lines 5.

## Decisions Made

- See key-decisions frontmatter; the load-bearing ones: structure-driven twin classification over hardcoded paths, self-containment assertion on twin include paths, the formatting-only normalization of 04-styling.toml to make the plan's fmt gate honest, and the extra plugin-copy sync

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] 04-styling.toml was not fmt-clean**
- **Found during:** Task 1 — running the plan-mandated `fmt --check skill/examples/` gate
- **Issue:** the fixture shipped with `  #` double-space trailing comments; `fmt --check` exited 1, so the plan's exit-0 acceptance criterion could not hold
- **Fix:** `c4drill fmt skill/examples/04-styling.toml` in place — 18/18 lines changed are the trailing-comment spacing normalization 35-08 documented (comment text verbatim, model unchanged); `fmt --check skill/examples/` now exits 0
- **Files modified:** skill/examples/04-styling.toml
- **Committed in:** 7d97712

**2. [Rule 3 - Plan inaccuracy] Task 4's spot command targets a by-design invalid fixture**
- **Found during:** Task 4 verification
- **Issue:** `convert to-c4d testdata/valid.toml` names the repo-root fixture, which is orphan-invalid BY DESIGN (35-07 already documented this) — the D-24 gate correctly refuses it; additionally converting it in place would drop an untracked twin into the repo
- **Fix:** ran the identical check on cmd/c4drill/testdata/valid.toml copied to a temp dir: convert to-c4d + fmt --check both succeed
- **Files modified:** none (verification-only task)

**3. [Rule 1 - Bug] SKILL.md referenced an invented command**
- **Found during:** Task 3 — editing the examples section
- **Issue:** "All examples parse and validate successfully with `c4drill validate`" — no `validate` subcommand exists (T-35-09-01: docs must not teach agents invented commands)
- **Fix:** replaced with the real invocation semantics: "`c4drill <file>` (rendering runs the full validation pipeline)"
- **Files modified:** skill/SKILL.md (+ synced copies)
- **Committed in:** f96563a

**4. [Rule 3 - Blocking] lint friction in the new walker**
- **Found during:** Task 1 RED
- **Issue:** nilerr on the os.Stat guard inside the WalkDir callback; wrapcheck on the bare filepath.Rel return
- **Fix:** inverted the stat guard (append only when the twin exists); wrapped the Rel error with `fmt.Errorf` (fmt import added)
- **Files modified:** internal/c4d/parity_test.go
- **Committed in:** ae82ba9

**5. [Rule 3 - Plan gap] extra plugin copies needed the same sync**
- **Found during:** Task 3
- **Issue:** the plan's file list covers the opencode copies, but plugins/c4drill/skills/c4drill-toml/ (the shared skills dir 4 of 5 manifests reference) and plugins/c4drill/commands/c4drill-render.md duplicate the same content — leaving them TOML-only would ship a half-updated plugin
- **Fix:** synced both (SKILL.md byte-identical; 12 twins mirrored; render doc updated keeping its one pre-existing front-matter difference)
- **Files modified:** see Task 3 commit f96563a
- **Committed in:** f96563a

---

**Total deviations:** 5 auto-fixed (2 plan inaccuracies, 1 blocking fmt gate, 1 docs bug, 1 plan gap)
**Impact on plan:** All acceptance criteria and must-have truths met.

## Threat Model Disposition

| Threat | Disposition | Where |
|--------|-------------|-------|
| T-35-09-01 (docs teach invented commands) | mitigated | every documented flag/example cross-checked against `--help` output and live runs; the pre-existing `c4drill validate` reference eliminated; no new commands invented |
| T-35-09-02 (broken examples repudiation) | mitigated | TestExampleTwins enforces model parity (12/12), render parity (standalone + both graphs), self-containment, and clean validation — a broken twin fails CI |
| T-35-09-03 (README info disclosure) | accepted (as planned) | public docs, no secrets |

## Issues Encountered

- C4D unit headers REQUIRE the brace block (`x: container "Name"` without `{ }` is a parse error) — documented explicitly ("required even when empty")
- A top-level `use` of a container-typed template is a validation error (C1-types-only at root) — the README template example places `use` inside the parent block, which is also the better teaching form
- fmt canonical style puts no blank line between a file's leading comment block and the first statement — the twins' header comments sit directly above `properties {`
- Pre-existing gofmt drift in 5 committed Go files (not touched by this plan) logged to 35-09-deferred-items.md

## Known Stubs

None. All documentation describes shipped, verified behavior; all 12 twins are generated from the real converter and enforced by the parity suite.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase 35 (all 9 plans) is complete: C4D parses, renders, converts, formats, and is documented end to end with example twins and a dual-format portable skill
- No blockers; `go test ./...` 15/15 green, golangci-lint 0 issues

## Self-Check: PASSED

All 12 skill/examples twins + the 3 parity-test/README/skill files exist on disk; all 4 task commit hashes (ae82ba9, 7d97712, ccbebdf, f96563a) verified in git log; packaged SKILL.md byte-identical to source (diff -q clean); 12 twins present in both packaged examples dirs.

---
*Phase: 35-add-a-simple-dsl-alternative-to-the-toml-diagram-definition*
*Completed: 2026-08-14*
