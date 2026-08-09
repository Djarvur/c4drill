# Phase 3: Compatibility & Validation - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-08-06
**Phase:** 3-Compatibility & Validation
**Areas discussed:** Baseline fixture strategy, Compat contract level, COMPAT-01 strictness, Expanded overrides

**Session note:** Final phase of the v1.8 milestone — verification-heavy, locking the compatibility guarantees from Phases 1-2 into enforceable tests.

---

## Baseline fixture strategy

| Option | Description | Selected |
|--------|-------------|----------|
| Public fixture | Sanitized public TOML + committed DOT baseline in testdata/; CI-enforceable COMPAT-02 | ✓ |
| Private + skip-if-missing | cyp-auth-infra/ stays private; tests t.Skip() when absent; CI skips strongest tests | |

**User's choice:** Public fixture
**Notes:** The private fixture (commit 1ae0cb6) stays gitignored — author's real infra data remains out of git; the public copy is sanitized/structural.

---

## Compat contract level

| Option | Description | Selected |
|--------|-------------|----------|
| DOT-level | Node/edge sets, attributes, cluster structure — stable and diffable | ✓ |
| Byte-identical SVG | Strongest, but brittle across environments (fonts, go-graphviz versions) | |

**User's choice:** DOT-level

---

## COMPAT-01 strictness

| Option | Description | Selected |
|--------|-------------|----------|
| Keep OR | Per-unit self-references still expand in C1 (v1.7 mechanism, Phase 2 D-05) | ✓ |
| Strict all-collapsed | Without properties.expanded, self-refs ignored in C1 | |

**User's choice:** Keep OR

---

## Expanded overrides

| Option | Description | Selected |
|--------|-------------|----------|
| Flag overrides all | --expanded expands everything regardless of properties.expanded (v1.7 contract) | ✓ |
| properties constrains | properties.expanded also limits the expanded file | |

**User's choice:** Flag overrides all

---

## Claude's Discretion

None — every area was decided with concrete options.

## Deferred Ideas

None.
