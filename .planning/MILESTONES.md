# Milestones

## v1.10 Model Composition (Shipped: 2026-08-08)

**Phases completed:** 6 phases, 13 plans, 35 tasks

**Key accomplishments:**

- Optional per-unit `reference` external-docs URL rendered as a clickable 📖 marker, wired through GraphViz's native URL attribute with a Safari-safe HTML shim that routes external links to a new tab.
- model.Humanize dumb camelCase splitter + parse-time Unit.Name fallback: omit `name` and get "Local IDP" from `localIDP`, with explicit `name =` always winning (zero new deps).
- README + skill/SKILL.md now document that `name` is optional with the dumb humanize rule, last-segment-only derivation, and the explicit-name escape hatch for acronyms like `gRPC`.
- Pure post-parse pass rewriting every Link.Peer from relative bare to absolute dotted path, implementing D-13/D-14/D-15/D-16 with a nearest-first ancestor walk-up and a hard error on miss-at-root.
- Connected the Phase 30 resolver to users: peer.Resolve now runs between Parse and Validate so authored bare peers resolve to absolute paths, with integration tests proving the ordering, the CLI error path, and corpus backward-compat.
- template.Expand turns every [[use]] into a concrete parametrized unit subtree via hand-rolled Unit.Clone (HS-1) + strings.NewReplacer substitution, with missing-param/duplicate-path hard errors and pipeline insertion before peer.Resolve
- IncludeDirective type + Model.Includes field landed so `[[include]]` array-of-tables route into the Model in document order with zero phantom units — the foundation Plan 02's resolver consumes.
- Recursive multi-file include resolver + per-field struct merge land as the pipeline's FIRST pre-processing pass (Stage 1a), delivering INC-01 through INC-10 and XC-02 — templates in included files flow through the merge so [[use]] in the entry file can instantiate them.
- Extracted the order-insensitive canonicalDOT comparator (DI-1) from internal/graph/builder_test.go into a reusable internal/testutil/canonical/ package, importable from any _test.go file — unblocking Plan 04's cross-package E2E goldens.
- Nine runnable TOML fixtures across four sets (06-templates, 07-relative-peer, 08-include, 09-composed) demonstrating all four v1.10 features; the composed set ships a hand-expanded single-file equivalent verified to canonicalize identical to its multi-file entry.
- README.md and skill/SKILL.md gain DOC-01 (omittable type with full inference tables) and DOC-02 (templates, include, relative-peer sections) matching the established Phase 28/29 style; SKILL.md also gains a Pipeline Ordering section explaining the load-bearing v1.10 composition pipeline.
- Two end-to-end tests in cmd/c4drill/root_test.go proving the four v1.10 features compose correctly through the full pipeline: XC-05 (composed multi-file ≡ hand-expanded single-file, canonicalDOT) and XC-01 (behavioral proof that include → template.Expand → peer.Resolve ordering is load-bearing, covering XC-02 and XC-03).

---

## v1.9 C3 Boundary Node Fix (Shipped: 2026-08-06)

**Phases completed:** 1 phases, 1 plans, 3 tasks

**Key accomplishments:**

- addResolvedBoundaryNode now stops its peer walk-up at the expanded container's parent, so C3 cross-container links surface the sibling container (e.g. mainSystem.rbac) as the boundary node instead of the parent system (mainSystem)

---

## v1.8 Proper C1/C2/C3 View Generation (Shipped: 2026-08-06)

**Phases completed:** 3 phases, 7 plans, 20 tasks

**Key accomplishments:**

- Pair-only edge dedup for resolved C1/C2/C3 views with first-wins attributes (D-01/D-03/D-06), binary penwidth 1.0/2.0 via new Edge.PenWidth (D-04), mirror-aware multiplicity counting (D-05), and a View.AllExpanded discriminator keeping --expanded byte-compatible with v1.7 (D-02/COMPAT-02)
- D-12 removal of the legacy recursive external-boundary path from GenerateExpandedView, D-02 activation of View.AllExpanded (restoring v1.7 dedup key and 2.0 penwidth in expanded mode), and D-13 minlen gating at all six resolved-link synthesis sites so length applies only to original-pair edges
- D-07..D-11 implemented: C1 links resolve to the deepest VISIBLE ancestor on BOTH sides — visible subunit entries (View.VisiblePaths + v.Units) inside expanded clusters receive subunit-level edges (D-07/D-09), within-cluster edges are recorded instead of dropped (D-10), parent edges are never duplicated (D-08), boxes follow the same rules (D-11), and BuildGraph skips VisiblePaths entries to prevent duplicate node IDs in DOT (Pitfall 5)
- D-07 graph-layer guard (expanded-but-empty units render as plain C1 nodes, no empty cluster box) plus 5 new regression tests locking VIEW-03/04/05 semantics — box and containerBox sub-diagram files at unit-key paths (D-01/D-02/D-03), per-unit expanded container clusters inside C2 boundary clusters (D-04), OR expansion precedence (D-05), silent ignore of unknown expanded entries (D-06), and personExternal actor boundary nodes in C2 (D-08)
- Committed sanitized 499-line multi-level TOML fixture (5 top-level units, 4-level nesting, cross-level length=3 link) and its CLI-generated 1041-line XDOT --expanded golden baseline, replacing the gitignored private saira fixture for CI-enforceable COMPAT-02 — with the pipeline's order-nondeterminism empirically characterized (semantic-stable, byte-unstable).
- Repointed the two private-fixture builder tests onto the committed public fixture, replaced the golden baseline comparison with an order-insensitive semantic canonicalizer (DI-1), locked D-04 (--expanded ignores properties.expanded), locked COMPAT-01 (all-collapsed C1), proved the multilevel 5-node C1 + C2/C3 end-to-end (ROADMAP SC4), and removed every "cyp-auth-infra" reference from .go files — COMPAT-01/COMPAT-02 are now enforceable in CI on fresh checkouts.
- Closed the three diagnosed UAT gaps that broke C2/C3 navigation: bidirectional explore URLs with a self-link guard, a clickable SVG navigation bar via GraphViz `<TD HREF>` HTML labels, and uniform `.svg` nav URLs across render formats — verified by a new end-to-end CLI regression (77/77 hrefs resolve).

---

## v1.7 Queue Label Fix & Visual Improvements (Shipped: 2026-03-29)

**Phases completed:** 8 phases (19–26, excluding skipped 24), 7 plans

**Stats:** 15,163 LOC Go, all tests passing

**Key accomplishments:**

- Queue units render with ASCII art graphic (═╦╩═╦═══) instead of broken rotated cylinder
- Helvetica font for all diagram elements
- Box labels with dashed borders, validation for mixed content, color by content type
- Link length attribute for edge spacing control
- Deterministic node/edge ordering
- Thicker edges (penwidth 2.0) for visual prominence
- TOML definition order preservation for nodes and edges

---

## v1.0 Initial Release (Shipped: 2026-03-10)

**Phases completed:** 6 phases, 16 plans

**Stats:** 9,624 LOC Go, 48 files, 28 feature commits

**Key accomplishments:**

- TOML parser with nested unit definitions and error handling
- C4 model validation with clear error messages and line numbers
- C1/C2/C3 view generation with collapsed/expanded rendering
- GraphViz DOT and SVG rendering via go-graphviz
- Interactive navigation with explore links, back-links, and breadcrumbs
- Production-ready Cobra CLI with help text and error handling

---
