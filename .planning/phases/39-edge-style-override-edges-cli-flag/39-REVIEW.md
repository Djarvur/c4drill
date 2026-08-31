---
phase: 39-edge-style-override-edges-cli-flag
reviewed: 2026-08-31T09:20:00Z
depth: standard
files_reviewed: 5
files_reviewed_list:
  - cmd/c4drill/root.go
  - internal/view/view.go
  - internal/graph/builder.go
  - internal/graph/builder_test.go
  - cmd/c4drill/root_test.go
findings:
  critical: 0
  warning: 0
  info: 1
  total: 1
status: issues_found
---

# Phase 39: Code Review Report

**Reviewed:** 2026-08-31T09:20:00Z
**Depth:** standard
**Files Reviewed:** 5
**Status:** issues_found

## Summary

Reviewed the `--edges` implementation surface: flag registration and validation (`cmd/c4drill/root.go`), the override carrier (`internal/view/view.go`), the post-PLAIN-02 application in both graph builders (`internal/graph/builder.go`), and the new test families (`builder_test.go`, `root_test.go`).

The precedence chain is correct and consistent across both builder sites: model-resolved value → `--plain` zeroing → explicit override (D-05). Validation is loud, early (before any file I/O), value-quoting, and enum-complete (D-02/GEDGE-04). Threading reaches both PLAIN-01 choke points including the `--expanded` copy. Tests follow repo conventions (nolint-with-reason, testify split, parallelism rules) and pin every locked decision. No critical or warning findings — no injection surface, no unchecked indexing, no error paths bypassed, no logic errors found.

One info-level consistency note below.

## Info Issues

### CR-39-01: `--edges` is only validated on the render command
- **file:** cmd/c4drill/root.go
- **line:** 132-137 (validateOutputFlags usage in runRoot)
- **issue:** `--edges` is a persistent flag, so it parses on the `convert`/`fmt` subcommands too, but validation runs only in `runRoot` — `c4drill convert to-c4d x.toml --edges diagonal` silently ignores the value instead of erroring.
- **fix:** None required for this phase: this exactly matches the existing `--format`/`--plain`/`--no-*` persistent-flag behavior (validated/used only by the render path), and GEDGE-04's scope is the render invocation. If the project ever wants stricter UX, move enum validation into a `PersistentPreRunE` on the root command.

## Structural Findings (fallow)

Not run — `code_quality.fallow.enabled` is not set (opt-in feature, disabled by default).
