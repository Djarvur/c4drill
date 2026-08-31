# Requirements: C4Drill — Milestone v1.16 Edge Style Override

**Defined:** 2026-08-31
**Core Value:** Transform simple TOML architecture descriptions into professional C4 diagrams without manual drawing.

## v1.16 Requirements

Requirements for this milestone. Each maps to roadmap phases.

### Edge Routing (GEDGE — continues v1.13 numbering)

- [ ] **GEDGE-03**: User can override the edge routing style for a whole invocation via `--edges <style>` accepting `straight|spline|square|ortho`, without editing the model file
- [ ] **GEDGE-04**: An invalid `--edges` value fails loudly with an error naming the offending value and the allowed enum (no silent fallback)
- [ ] **GEDGE-05**: `--edges` overrides the model's `edges` property on every generated view — C1 root, all drill-down views, and the `--expanded` copy (PLAIN-01 threading pattern)
- [ ] **GEDGE-06**: `--edges` composes with `--plain`: an explicit CLI request survives `--plain`'s author-format suppression (user intent wins over model-derived formatting), with the decision pinned by a dedicated test
- [ ] **GEDGE-07**: The switch-matrix E2E is extended to `--edges` × generation (root / drill-down / `--expanded`) × `--plain`, asserting the graphviz `splines` attribute in RAW dot output
- [ ] **GEDGE-08**: Without the flag, output is unchanged — all existing canonicalDOT goldens pass untouched (backward compat)

## Future Requirements

Tracked in PROJECT.md "Next Milestone Goals"; not in this roadmap.

- Template multi-output / `for_each` fan-out
- Compact-link shorthand variants beyond baseline
- C4D polish warnings: WR-03 duplicate `properties {}` last-win, WR-04 skill type-inference drift, WR-05 quoted-label whitespace trim

## Out of Scope

Explicitly excluded. Documented to prevent scope creep.

| Feature | Reason |
|---------|--------|
| New routing styles beyond `straight\|spline\|square\|ortho` | Reuse existing enum + converter mapping (GEDGE-02); no new styles requested |
| Per-unit CLI style overrides | Flag is invocation-global, mirroring the TOML `edges` property scope |
| Config-file mechanism for CLI overrides | CLI-only per established KEY/PLAIN pattern |
| `--edges` changing node/cluster formatting | Out of GEDGE scope — formatting suppression is the KEY family's job |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| GEDGE-03 | — | Pending |
| GEDGE-04 | — | Pending |
| GEDGE-05 | — | Pending |
| GEDGE-06 | — | Pending |
| GEDGE-07 | — | Pending |
| GEDGE-08 | — | Pending |

**Coverage:**
- v1.16 requirements: 6 total
- Mapped to phases: 0
- Unmapped: 6 ⚠️ (roadmap pending)

---
*Requirements defined: 2026-08-31*
*Last updated: 2026-08-31 after initial definition*
