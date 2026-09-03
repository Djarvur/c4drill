# Requirements: C4Drill — Milestone v1.17 Issue Sweep

**Defined:** 2026-09-03
**Core Value:** Transform simple TOML architecture descriptions into professional C4 diagrams without manual drawing.
**Source:** Open GitHub issues #42, #41, #38.

## v1.17 Requirements

Requirements for this milestone. Each maps to roadmap phases.

### Deterministic Rendering (issue #42)

- [ ] **REPRO-01**: Rendering the same model file twice with no changes produces byte-identical SVG output — including the generated `id="edge<N>"` group ids
- [ ] **REPRO-02**: Edge ids are assigned in a deterministic order derived from model content (sorted/insertion order), never from Go map iteration order
- [ ] **REPRO-03**: The determinism guarantee holds for every supported output format (`dot`, `svg`, `html`) across repeated runs

### Check Command (issue #41)

- [x] **CHECK-01**: User can run `c4drill check <file>` to validate a model without producing any output files or requiring an output directory
- [x] **CHECK-02**: `check` exits 0 when the model is valid and exits non-zero when invalid, reporting the same validation errors the render path reports (e.g. orphan-unit VAL rules)
- [x] **CHECK-03**: `check` runs the same pipeline front-half as render — includes resolved, templates expanded, relative peers resolved — so composed multi-file sources validate exactly as they render
- [ ] **CHECK-04**: README documents the `check` command alongside the existing CLI surface

### Desktop GUI Fix (issue #38)

- [x] **GUI-01**: In desktop-window mode the frontend RPC layer calls the Wails-generated binding namespace that matches the actually bound struct (`main.desktop`), so desktop-window RPC works again (broken since the #31/#37 restructure)
- [x] **GUI-02**: The `--serve` HTTP fallback path is unchanged and the existing e2e suite stays green

## Future Requirements

Deferred to future milestones. Existing backlog, unchanged by this sweep.

- Template multi-output / `for_each` fan-out
- Compact-link shorthand variants beyond baseline
- C4D polish warnings: WR-03 duplicate `properties {}` last-win, WR-04 skill type-inference table drift, WR-05 quoted-label whitespace trim

## Out of Scope

Explicitly excluded. Documented to prevent scope creep.

| Feature | Reason |
|---------|--------|
| #34 Zed preview panel via visual extension API | Blocked upstream: `zed_extension_api` 0.7.0 has no WebView/panel API; open RFC zed-industries/zed#53403. Documented fallback already shipped in #30. |
| #35 JetBrains plugin live-IDE validation (runIde + Plugin Verifier) | Environmental, not codeable from a CLI agent: requires `download.jetbrains.com` reachable (HTTP 451 at dev site) and an interactive IDE session. Revisit when on an unrestricted network. |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| REPRO-01 | Phase 40 | Pending |
| REPRO-02 | Phase 40 | Pending |
| REPRO-03 | Phase 40 | Pending |
| CHECK-01 | Phase 41 | Complete |
| CHECK-02 | Phase 41 | Complete |
| CHECK-03 | Phase 41 | Complete |
| CHECK-04 | Phase 41 | Pending |
| GUI-01 | Phase 42 | Complete |
| GUI-02 | Phase 42 | Complete |
