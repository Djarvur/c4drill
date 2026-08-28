# Requirements: v1.13 Edge Semantics and Legend

**Status:** Active
**Milestone:** v1.13 Edge Semantics and Legend (product release tag: v1.18.0)
**Last updated:** 2026-08-28

Generated diagrams must have trustworthy colour semantics. Six user-reported problems define this milestone: custom unit colours are silently dropped at render; the global edge style (`properties.edges`) is ignored by C2/C3 drill-down diagrams; reversing an edge's layout ranking requires an unclear `"<-"` + `arrow = "reverse"` combination; edges cannot be coloured by data-flow kind (read/write); collapsed edges lose kind identity; and diagrams carry no legend explaining the colour conventions.

Source: user request 2026-08-28 + codebase scan (root causes pre-confirmed):
- `Unit.Color/Style/Border` and `Properties.Color/Style/Border` are parsed but have zero render-side reads — `buildNode`/`buildCluster` style exclusively via type palettes (`internal/graph/shapes.go`).
- `properties.edges` reaches C1/expanded views; C2 (`view/scope.go:377`) and C3 (`view/scope.go:470`) copy only the drilled-into unit's own `Unit.Edges` — no global fallback. `"square"` is documented but unimplemented in `configureGraphSettings`.
- `rank = "forward"/"reverse"` parse and round-trip but are consumed nowhere; only `rank = "equal"` → `constraint=false` works.
- `graph.Graph.Legend` placeholder struct exists, never populated or rendered.

---

## v1.13 Requirements

### COLOR — Unit styling fix

- [ ] **COLOR-01**: A unit's explicit `color` actually renders — node background/font/border colours reflect author-specified `Unit.Color` for plain nodes AND expanded-unit clusters (via `FillColor`/`FontColor`/`Color` overrides at `buildNode`/`buildCluster`, emitted through `applyNodeStyle`/`applyClusterStyle`). Unset fields keep the type-palette defaults.
- [ ] **COLOR-02**: The sibling parsed-but-dead unit fields `style` and `border` render too (same override mechanism), so the full documented unit styling triple (`color`/`style`/`border`) works in one fix.

### GEDGE — Global edge style

- [ ] **GEDGE-01**: Every diagram generation respects `properties.edges`: C2 and C3 views fall back to the global setting when the drilled-into unit does not set its own `edges`; per-unit values keep winning where present. Disabling splines globally disables them on every generated diagram.
- [ ] **GEDGE-02**: The documented value set and implementation agree — `square` either maps to a real GraphViz splines mode (`ortho`) or is removed from the docs; no documented value silently no-ops.

### RANK — Convenient rank reversal

- [ ] **RANK-01**: `rank = "reverse"` on a link reverses the edge's layout ranking with that single option — same visual arrow direction, opposite vertical ordering (emit swapped endpoints with `dir=back`). It replaces the `"<-"` + `arrow = "reverse"` idiom; `rank = "forward"` is the explicit default; `rank = "equal"` keeps its existing `constraint=false` behavior.
- [ ] **RANK-02**: `rank = "reverse"` works in collapsed and expanded views and survives edge resolution (view copiers carry Rank; collapsed copies keep it).

### KIND — Edge kinds

- [ ] **KIND-01**: Links accept `kind = "read" | "write" | "read-write"` with kind-derived colours: read = green, write = red, read-write = a blend colour distinct from both. The kind colour applies when the link sets no explicit `color`.
- [ ] **KIND-02**: An explicit `color` overrides the kind colour (kind is recorded and round-trips, but does not colour the edge).
- [ ] **KIND-03**: `kind` works in both formats — TOML and C4D (grammar `OptionKey` + `applyEdgeOption` + emitters + reserved-word suggestions) — and round-trips through `convert`/`fmt` canonically, including `${param}` substitution inside templates.

### AGG — Collapsed-edge kind semantics

- [ ] **AGG-01**: When edges collapse to a visible ancestor, the collapsed edge's colour derives from the constituent kinds: all read → read colour, all write → write colour, mixed → read-write colour.
- [ ] **AGG-02**: The collapsed edge's line style follows precedence: all constituents share one style → that style; otherwise any solid → solid; otherwise any dashed → dashed; otherwise dotted.
- [ ] **AGG-03**: If the constituent edges carry explicit custom colours (not kind-derived), kind colouring is suppressed on the collapsed edge and the default edge colour is used.

### LEG — Legend

- [ ] **LEG-01**: Every generated diagram includes a legend in the upper-right corner, controlled by a single global setting (properties-level) that defaults to **enabled**; authors can disable it for the whole model.
- [ ] **LEG-02**: The legend explains the default colour conventions — the edge kind colours (read/write/read-write) and the default edge/line-style variants — matching the palette actually used by the renderer.
- [ ] **LEG-03**: Authors can add custom legend lines (label + colour + optional style) in both formats; custom lines render after the defaults.

### BC — Backward compatibility

- [ ] **BC-01**: Models that use none of the new features render unchanged **except** for the default-on legend (accepted, user-mandated default) — canonicalDOT goldens re-baselined only for the legend block and any documented deltas; full suite stays green.

### DOC — Documentation and skills

- [ ] **DOC-01**: README.adoc documents unit styling (now actually working), the global edge style fallback, the `rank = "reverse"` idiom (replacing the old dance), link `kind` colours, and the legend setting with custom lines.
- [ ] **DOC-02**: skill/SKILL.md and all plugin copies are synced with the same surface.
- [ ] **DOC-03**: Skill/example fixtures demonstrate the new features (both formats where applicable) and render cleanly through the full pipeline.

### REL — Release

- [ ] **REL-01**: Milestone ships as product release **v1.18.0** (git tag; CI release workflow builds artifacts and creates the GitHub release).

---

## Out of Scope

| Feature | Reason |
|---------|--------|
| Per-edge custom kind palettes / themeable kind colours | One default palette keeps the legend honest; custom `color` already covers one-off needs |
| Legend positioning controls (corner choice, floating) | User asked for upper-right, global, default-on — no configurability beyond on/off |
| Manual positioning / rank=same subgraph authoring | Violates the auto-layout design decision (PROJECT.md) |
| `rank = "forward"` doing anything beyond the default | Forward is the natural direction; the option exists for symmetry/explicitness only |
| Custom unit iconography/fonts | Not requested |
| `properties.color/style/border` (page-level styling) | Parsed-but-dead today; not part of the reported defect — deferred until asked for |

## Traceability

*Filled by roadmap (phase ↔ requirement mapping).*
