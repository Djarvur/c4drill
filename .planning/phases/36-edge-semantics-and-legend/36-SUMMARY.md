# Phase 36 Summary — Edge Semantics and Legend

**Shipped:** 2026-08-28 · **Product release:** v1.18.0 · **Plans:** 6 (36-01..36-06) · All 20 requirements satisfied.

## What Shipped

- **COLOR-01/02** — `Unit.Color`/`Style`/`Border` now render on plain nodes AND expanded/boundary clusters (`applyUnitOverrides` in the builder; author overrides beat the box-content heuristic). Dark explicit fills force white labels via a luminance rule; `dotted` border style maps through `buildStyleString`. (36-01)
- **GEDGE-01** — `properties.edges` now falls back into C2/C3 views (`cmp.Or(unit.Edges, properties.Edges)`), so disabling splines disables them everywhere; per-unit values still win. E2E test proves `splines=false` in generated C2 DOT. (36-03)
- **GEDGE-02** — `edges = "square"` now emits `splines=ortho` (documented value no longer a no-op). (36-01)
- **RANK-01/02** — `rank = "reverse"` is a single-knob rank reversal: endpoints swap at emission (`Edge.RankReverse`, logical Source/Target unchanged) with an inverted `dir`, byte-equivalent at canonicalDOT level to the old `"<-"` + `arrow="reverse"` idiom. Full arrow×rank direction matrix pinned in tests; `dir=forward` is never emitted per-edge. Kind/Rank survive all 4 view copiers + the validator mirror. (36-01/03)
- **KIND-01/02/03** — `kind = "read" | "write" | "read-write"` on links: green `#2E7D32` / red `#C62828` / purple `#6A1B9A` (palette in `model/colors.go`, shared with the legend). Precedence: explicit `color` > kind > source-border default. Works in TOML (struct tag) and C4D (grammar `OptionKey` + pigeon regen, `applyEdgeOption`, canonical order arrow→rank→**kind**→color→style→labelPosition→length in both emitters + canonsrc); round-trips through convert/fmt (corpus fixture `kind.toml` pinned); survives template instantiation — **`kind` is NOT substituted by `${param}`** (enum precedent, pinned by test; documented in README/skill). (36-01/02)
- **AGG-01..03** — collapsed edges keep kind identity: `collectPairAggregates` pre-scan beside `countPairMultiplicity`; on collapsed pairs (2+ constituents), colour derives from kinds (all-same → kind colour, mixed → purple), style precedence all-same → any solid → any dashed → dotted (unstyled = solid), and any explicit colour or unset kind suppresses kind colouring to the D-01 default. Single-link pairs and `--expanded` untouched. (36-03)
- **LEG-01/02/03** — default-on legend in the upper-right (top label table, right-aligned rows, `TD BGCOLOR` swatches, text-glyph style samples in `#666666`, all rows in explicit `<FONT POINT-SIZE>`); guard changed so nameless C1/expanded views carry the legend; `properties.legend = false` (`*bool`, nil = on) disables model-wide; custom rows via `[[properties.legendLine]]` (TOML) / `legendLine: ["label|color|style"]` (C4D, pipe-split) render after the defaults, HTML-escaped (T-36-04-01). Palette single-sourced from model consts. (36-04/05)
- **BC-01** — models using no new feature are source-stable (no legend/kind keys emitted) and DOT-identical except the legend label block; `multilevel.expanded.dot` re-baselined in a dedicated commit; no-feature regression test added. (36-05)
- **DOC-01..03** — README (rank/properties/Edge Kinds & Legend/styling/C4D edges) and skill/SKILL.md updated; new `10-edge-kinds.toml` + `.c4d` twin (pinned in the twins manifest) with SVG; 03-links (kind + rank=reverse) and 04-styling (custom legend line) extended in both formats; all plugin copies byte-synced; CI now validates `.c4d` examples and fails on plugin-copy drift. (36-06)
- **REL-01** — tagged **v1.18.0** (release workflow builds 6 platforms + GitHub release).

## Key Decisions

- read-write blend = purple `#6A1B9A` (hue-orthogonal to green/red; a literal blend is muddy brown).
- rank=reverse = emission-side endpoint swap; edge `key` stays logical.
- AGG aggregates only when ALL constituents carry a kind (BC-safe); explicit colour forces the default.
- `kind` NOT substituted by template `${param}` — surfaced in docs as the pinned interpretation (checker warning 3 resolution).
- Unknown C4D edge options are hard parse errors (pre-existing; Levenshtein suggestions remain unit-field scoped).

## Verification

- `go test ./...` green (15 packages with tests) at close.
- fmt `--check` clean on skill/examples; twin parity (13 pairs) green; CI sync check green.
- UAT: rendered SVGs show kind colours, legend swatches, and reverse-dir; E2E test proves global splines=false in C2.
