---
phase: 41-check-command
plan: 02
subsystem: docs
tags: [asciidoc, readme, skill, documentation, issue-41]

requires:
  - phase: 41-check-command (plan 01)
    provides: the shipped check command whose Use line and exit codes the docs pin
provides:
  - README.adoc `=== check` command section (usage, formats, exit codes, no-output guarantee)
  - skill/SKILL.md command examples covering check + the CI gate pair
affects: [users, CI adopters, skill consumers]

tech-stack:
  added: []
  patterns: []

key-files:
  created: []
  modified:
    - README.adoc
    - skill/SKILL.md

key-decisions:
  - "Docs pin the exact cobra Use line (check <file.toml|file.c4d>) and the errValidationFailed exit semantics — no invented flags or outputs (D-05/D-06)"

requirements-completed: [CHECK-04]

duration: ~8 min
completed: 2026-09-03
---

# Phase 41 Plan 02: Document Check Command Summary

`c4drill check` documented across both CLI surfaces (CHECK-04, D-06): an AsciiDoc `=== check` section in README.adoc beside `=== fmt`/`=== serve`, and check entries in skill/SKILL.md's command examples — usage, both authoring formats, exit codes, no-output guarantee, and the fmt --check + check CI gate pair.

## What Was Built

- **README.adoc** — `=== check` section (line ~1474) between `=== fmt` and `=== serve`, mirroring sibling AsciiDoc style (`[source,text]` usage block, `*` bullets, `[source,bash]` examples): usage `c4drill check <file.toml|file.c4d>`; same-front-half-as-render bullet; both formats `.toml`/`.c4d`; exit codes 0 valid (silent) / 1 invalid with render-identical errors; no-output guarantee; examples including `c4drill fmt --check . && c4drill check architecture.c4d`. No flags documented (check defines none — D-05).
- **skill/SKILL.md** — section retitled to `### Converting, Formatting, Validating (convert / fmt / check)`; bash block gains `c4drill check architecture.c4d` (render-free validation, nothing written, exit 0/1) and `c4drill check architecture.toml && c4drill fmt --check .` (CI gate pair, both exit 1 on failure).

## Tasks

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | README.adoc `=== check` section | 9623027 | README.adoc |
| 2 | skill/SKILL.md usage entry | 7c39752 | skill/SKILL.md |

## Verification

- `grep -c '^=== check$' README.adoc` → 1, positioned after `=== fmt` (1452) and before `=== serve` (1499)
- `grep -c 'c4drill check <file\.toml|file\.c4d>' README.adoc` → 1 (matches the cobra Use line exactly)
- `grep -c 'c4drill check' skill/SKILL.md` → 2; heading `convert / fmt / check` present
- No markdown headings added to README.adoc (diff inspected — AsciiDoc-consistent)
- `go build ./...` clean (docs-only plan)

## Deviations from Plan

None - plan executed exactly as written.

## Self-Check: PASSED

- Modified files exist with the pinned strings (greps above)
- git log --grep 41-02: 2 commits
- All task acceptance criteria re-run: PASS
