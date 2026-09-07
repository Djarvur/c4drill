# Phase 41: Check Command - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-09-03
**Phase:** 41-Check Command
**Areas discussed:** Command shape, Pipeline reuse, Output behavior, Flag surface, Docs
**Mode:** `--auto` (yolo runtime) — all selections auto-chosen and logged

---

## Command shape

| Option | Description | Selected |
|--------|-------------|----------|
| New `check` subcommand | Sibling of fmt/convert/serve; issue asks for a command | ✓ |
| Root `--check` flag | Ambiguous beside default-render root; collides conceptually with fmt --check | |

**User's choice:** [auto] Q: "Subcommand or flag?" → Selected: "New `check` subcommand" (recommended default).

---

## Pipeline reuse

| Option | Description | Selected |
|--------|-------------|----------|
| Shared front-half helper | Extract stages 1→2 so check and render cannot drift | ✓ |
| Copy stages into check.go | Faster now, drifts later | |

**User's choice:** [auto] Q: "How to guarantee render-identical validation?" → Selected: "Shared front-half helper" (recommended default).

---

## Output behavior

| Option | Description | Selected |
|--------|-------------|----------|
| Silent success, exit 0 | Unix convention; CI-friendly; matches render's silent success | ✓ |
| Print "OK" summary | Friendlier but noisy in CI loops | |

**User's choice:** [auto] Q: "What does check print on a valid model?" → Selected: "Silent success, exit 0" (recommended default).

---

## Flag surface

| Option | Description | Selected |
|--------|-------------|----------|
| Input file only | No render flags — nothing renderable, nothing to configure | ✓ |
| Accept --format/-o for symmetry | Misleading: implies output | |

**User's choice:** [auto] Q: "Which flags does check carry?" → Selected: "Input file only" (recommended default).

---

## Claude's Discretion

- Shared front-half factoring (extracted function vs dry-run switch in runRoot).
- Test fixture placement; whether skill/SKILL.md needs the command table row.

## Deferred Ideas

- `check --json` machine-readable output for tooling — rejected for this phase (no consumer yet); revisit if an editor/CI integration asks for it.
