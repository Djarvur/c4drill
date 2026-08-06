---
phase: 03-compatibility-validation
reviewed: 2026-08-06T00:00:00Z
depth: standard
files_reviewed: 5
files_reviewed_list:
  - cmd/c4drill/root_test.go
  - cmd/c4drill/testdata/multilevel.toml
  - cmd/c4drill/testdata/multilevel.expanded.dot
  - internal/graph/builder_test.go
  - internal/render/expanded_internal_test.go
findings:
  critical: 0
  warning: 2
  info: 2
  total: 4
status: issues_found
---

# Phase 3: Code Review Report

**Reviewed:** 2026-08-06
**Depth:** standard
**Files Reviewed:** 5
**Status:** issues_found

## Summary

Reviewed the Phase 3 (Compatibility & Validation) test-infrastructure delta: the sanitized public fixture `multilevel.toml`, the CLI-generated golden `multilevel.expanded.dot`, the repointed builder tests (`TestBuildExpandedGraphRealToml`, `TestBuildExpandedGraphBaselineDOT`), the new `canonicalDOT` semantic comparison helper (DI-1), the D-04 properties.expanded-immunity test, the COMPAT-01/COMPAT-02 regressions, and the synthetic-model conversion in `expanded_internal_test.go`. All three package test suites pass (`go test ./internal/graph/ ./internal/render/ ./cmd/c4drill/`), the golden-baseline comparison is stable across repeated runs, and `go vet` is clean.

The repointing work is correct: fixture paths resolve, `mainSystem.storages.externalStorage.client -> externalSys` matches the fixture's `length=3` link and the golden's `minlen=3` edge, and the D-04 count-vs-recursive-walk assertion is meaningful (the poisoned view must still contain all units). The COMPAT-01 assertions correctly avoid asserting `valid/app.dot` absence.

However, the phase's core deliverable — the `canonicalDOT` helper that enforces COMPAT-02 — contains an off-by-one that corrupts the final attribute of every statement and a demonstrated false-pass window (WR-01), plus value-unaware statement terminators that can silently desync parsing (WR-02). Both weaken the "no false passes" property the golden comparison is supposed to guarantee, and the phase summary explicitly designates `canonicalDOT` as the reference pattern for future DOT golden comparisons, so these should be fixed before that pattern is reused.

## Warnings

### WR-01: canonicalDOT off-by-one drops the last character of the last attribute — demonstrated false-pass window

**File:** `internal/graph/builder_test.go:1370` (`parseDOTAttrStatement`)
**Issue:** `text` is sliced with `dot[pos : pos+end]`, where `end` points at the `]` of the `];` terminator — so `text` ends exactly at the last character of the last attribute value. The block is then sliced as `text[open+1 : len(text)-1]`, which drops that final character. Verified empirically against the committed golden file: all 56 edge statements canonicalize with `style=soli` instead of `style=solid`, and node statements drop the last digit of `width=...` (masked only because `width` is geometry-stripped). Because the same corruption is applied to both the golden and the freshly rendered side, the comparison passes today — but the canonical form is not the "sorted attributes" the D-02 contract claims. Worse, the corruption is demonstrable as a false-pass: two documents differing ONLY in the final character of a statement's final attribute compare equal (probe: `penwidth=1` vs `penwidth=2` as a final attribute both canonicalize to `penwidth=`). Today edge statements end with `style=solid`, so a real drift confined to that position (or any future renderer change that moves `penwidth`/`minlen`/`style` to the last attribute slot) would silently pass.
**Fix:** slice the full block and let the existing per-piece `TrimSpace` handle trailing whitespace:
```go
attrs: normalizeDOTAttrs(text[open+1:]),
```

### WR-02: canonicalDOT statement terminators are not value-aware — `];`, `{`, `}` inside attribute values truncate and desync parsing

**File:** `internal/graph/builder_test.go:1354` (`parseDOTAttrStatement`), `:1334` (`parseDOTSubgraph`), `:1297` (`parseDOTBlock`)
**Issue:** The parser terminates an attribute statement at the first `];` and a block at the first `}` (and a subgraph header at the first `{`) via raw `strings.Index`, with no awareness of quoted strings or HTML label content. The HTML-entity comment (dodging `;`) only covers one case. Attribute values are user-authored (descriptions, technologies): a description containing `]` immediately followed by `;` (e.g., `description = "SSH [session]; admin"`) truncates the statement at that point; the remainder of the statement and everything after it is parsed as new, shifted statements. Since a regenerated golden would contain the same values, the truncation is symmetric and silently masks every semantic difference AFTER the truncation point — the exact opposite of what the COMPAT-02 lock is for. The current fixture happens to be free of these byte patterns, so the hazard is latent, but it defeats the "does it normalize what it claims" guarantee on realistic input.
**Fix:** scan for terminators with quote/`<...>` awareness (track whether the cursor is inside a `"`-quoted value or an HTML `<...>` region before accepting `];`, `{`, or `}`), or validate at the end that the parse consumed the document to its final `}` and that heads/attrs have no leftover structural tokens — a cheap invariant that turns silent truncation into a loud failure.

## Info

### IN-01: TestHelpSubcommand is tautological and never exercises the help subcommand

**File:** `cmd/c4drill/root_test.go:39-65` (pre-existing, not part of the Phase 3 delta)
**Issue:** `cmd2.Execute()`'s error is discarded (`_ = err2`), `buf2` is never read, and the only assertion re-checks `helpOutput` from the `--help` invocation — the same assertion `TestHelpText` already makes. If the `help` subcommand broke entirely, this test still passes. The function name claims coverage it does not provide.
**Fix:** assert on the subcommand's own output and error, e.g. `require.NoError(t, err2)` and `assert.Contains(t, buf2.String(), "c4drill <input.toml>")` — or delete the test as a duplicate of `TestHelpText`.

### IN-02: loadCYPAuthInfraModel is now a misleading pass-through wrapper

**File:** `internal/render/expanded_internal_test.go:18-22`
**Issue:** After the fixture-probing loop was removed (good — determinism restored), `loadCYPAuthInfraModel` only calls `createSyntheticCYPModel(t)`. The name implies file loading that no longer happens, and the two-layer indirection adds nothing; the "CYP Auth Infra" naming also survives in both functions while the fixture structure is now generic.
**Fix:** inline the call at the four call sites and drop the wrapper, or rename the pair to something structure-descriptive (e.g., `syntheticAuthInfraModel`).

---

_Reviewed: 2026-08-06_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
