# Codebase Concerns

**Analysis Date:** 2026-08-05

## Critical Bug: C2/C3 navigation renders as literal HTML text (not clickable links)

**What happens:** The back-link and breadcrumb navigation for C2/C3 diagrams is injected into the graph label as a raw HTML string (`<a href="...">Back to ...</a>`) via `cg.SetLabel(joinLabels(labelParts))` in `internal/render/converter.go:176-195` (label built by `internal/render/navigation.go:13-49`). Because the label is set as a *plain text* label (not via `cg.StrdupHTML`), graphviz HTML-escapes the entire string and renders the raw anchor markup as visible text inside the diagram.

**Verified:** Generating SVG for any C2 diagram produces `&lt;a href=&quot;../inject.svg&quot;&gt;Back to ...&lt;/a&gt;` as a text node. The C2 SVG contains zero clickable `href`/`xlink:href` elements, while the C1 SVG correctly contains `xlink:href="inject/mainapp.svg"` (explore links set via `cn.SetURL` in `internal/render/converter.go:244-246` work fine — only the label-based nav is broken).

**Why it's wrong:** The README headline feature is "Clickable C1, C2, and C3 diagrams with automatic drill-down navigation". Back-navigation and breadcrumbs are visible but not clickable; users see raw HTML source text.

**Files:**
- `internal/render/converter.go:179-195` (nav label joined into plain `SetLabel`)
- `internal/render/navigation.go:13-49` (HTML anchor string construction, no `StrdupHTML` call)
- `internal/render/render.go:98-101` (cgraph built without HTML label registration for graph label)

**Why tests don't catch it:** `TestIntegration_SVG_BackLink` (`internal/render/integration_test.go:424-438`) only asserts the string "Back to" is present, and `TestIntegration_SVG_Breadcrumbs` (lines 441-457) only asserts ">" or "&gt;" — both pass on escaped literal text. Assertions are too weak to detect that the anchor is not an element.

**Fix approach:** Build the navigation HTML string, register it with the cgraph instance via `StrdupHTML` (mirroring `setNodeLabel`/`setClusterLabel` in `internal/render/converter.go:252-270, 351-369`), and `SetLabel` the HTML handle. Add a test asserting the rendered SVG contains an `xlink:href` (not `&lt;a href`) for C2/C3 back-links and breadcrumbs.

## Known Bugs

### `arrow = "bidirectional"` silently renders as a forward arrow
- **Symptoms:** Links with `arrow = "bidirectional"` render with a single forward arrowhead; no `dir` attribute appears in DOT output.
- **Files:** `internal/model/link.go:12-13` (defines `ArrowBidirectional = "bidirectional"`), `internal/graph/builder.go:434` (direct cast `ArrowDirection(link.Arrow)`), `internal/render/converter.go:456-465` (switch only handles `forward`/`reverse`/`both`/`none`).
- **Root cause:** Value mismatch — `model` documents `"bidirectional"`, `graph`/`render` use `"both"`. The parser test `internal/parser/parser_test.go:486` uses `"bidirectional"`; README (`README.md:201`) and skill (`skill/SKILL.md:260`) document different value sets.
- **Trigger:** Any TOML link with `arrow = "bidirectional"`.
- **Workaround:** Use `arrow = "both"` (though `"both"` is undocumented in the model's own constants).
- **Fix approach:** Align the enum: accept `"bidirectional"` and map to `cgraph.BothDir`, or rename `ArrowBidirectional` to `"both"` and keep backward-compat parsing for `"bidirectional"`.

### C3 (depth-3+) subunit order is non-deterministic
- **Symptoms:** Components inside a container (`[system.container.component]` dotted-table syntax) appear in a different order on every run. Verified empirically across runs: `Alpha Mid Zed` vs `Mid Zed Alpha` for definition order `zed, alpha, mid`.
- **Files:** `internal/parser/parser.go:100-157` (`captureDefinitionOrder` only records 1- and 2-part tables; `len(parts) > 2` is explicitly skipped at line 149), `internal/parser/parser.go:196-238` (fallback path iterates Go `map[string]any` in random order when no order was captured).
- **Trigger:** Nested tables at depth 3+, e.g. `[mainapp.api.handlers]`.
- **Impact:** Non-reproducible diagrams; the definition order contract documented for C1/C2 (`TestParseSubunitDefinitionOrder`, `internal/parser/parser_test.go:869-904`) silently breaks one level deeper.
- **Fix approach:** Record order for tables of any depth in `captureDefinitionOrder` (walk the full key path), or switch to go-toml `unstable` `Array`/`Table` iteration that preserves document order.

### Parse errors from the `unstable` API always report line 1
- **Symptoms:** TOML syntax errors detected by the order-capture pass report `at line 1` regardless of the real location.
- **Files:** `internal/parser/errors.go:50-54` (hardcodes `Line: 1` for `*unstable.ParserError`).
- **Fix approach:** Pre-scan the data to map `Highlight` byte offset to a line number, or run the order-capture only after a successful regular unmarshal (fall back to raw-map ordering only on `DecodeError`, which carries a position).

### `edges = "square"` is silently ignored
- **Symptoms:** `[properties] edges = "square"` (documented in `README.md:59` and `skill/SKILL.md:87`) produces default (spline) routing; no `splines` attribute in output.
- **Files:** `internal/render/converter.go:158-166` — switch handles only `"spline"`, `"straight"`, `"ortho"`.
- **Fix approach:** Map `"square"` to `cg.SetSplines("true")` + `cg.SetNodesep(0)` (graphviz has no native square mode; closest is `splines=true` with zero separation) or remove `"square"` from all documentation.

### Validation error order is non-deterministic
- **Symptoms:** The order of reported validation errors changes between runs because `ValidateReferences` and other rules iterate a Go map.
- **Files:** `internal/validator/rules.go:20` (`for path, info := range index`), `internal/validator/validator.go:30-35`.
- **Impact:** Flaky scripted output; snapshot tests would be unstable.
- **Fix approach:** Sort errors by path before returning.

### Explore URLs in `-f dot` output point to nonexistent `.svg` files
- **Symptoms:** With `-f dot`, node `URL` attributes always use the `.svg` extension (`internal/graph/path.go:14-16`, format arg ignored), while back-links use the requested format (`path.go:72-88`). Clicking a `[+]` node in a DOT file opens a missing SVG.
- **Files:** `internal/graph/path.go:14-16`, `internal/graph/builder.go:474-503`.
- **Impact:** Low for dot output (not typically interactive), but the inconsistency is documented in a stale comment only.

## Tech Debt

### Documentation drift: README and skill describe behavior the code no longer has (or never had)
- **Inline link map syntax fails to parse.** `README.md:186,192,213-215,296,340-347` use `link = { "user" = { ... } }` (map form). Verified: this fails with `parse error at line 4: toml: cannot store a table in a slice`. Only `[[unit.link]]`/`[[unit.linkFrom]]` array-of-tables works (`internal/model/link.go:43-62`, `internal/model/unit.go:65-67`).
- **`expanded` no longer controls which diagrams are generated.** README (`README.md:218-254`) says `expanded = [...]` "Generate C2 for mainapp". Since v1.8 the tool auto-detects: any unit with subunits gets a sub-diagram (`cmd/c4drill/root.go:168-199`; confirmed by `.planning/STATE.md`). The `expanded` lists now only control whether subunits render as clusters inside the parent view (`internal/view/scope.go:144-150`, `internal/graph/builder.go:324-326`). README is stale.
- **`c4drill validate` subcommand does not exist.** `skill/SKILL.md:434` says examples "validate successfully with `c4drill validate`" and line 531-541 instructs running it. Running `c4drill validate` treats "validate" as a TOML filename (`cmd/c4drill/root.go:65-66,102`). No subcommand is registered.
- **C2/C3 `box` usage documented in the skill fails validation.** `skill/SKILL.md:170-212` teaches `type = "box"` inside a `system`/`container` with `container`/`component` children. Verified: the validator rejects this (`error: unit "mainapp.services" has type box which must be C2 type (inside system)`). `box` is validated as C1-only (`internal/validator/rules.go:236-246`, `c1Types` at 147-157; `c2Types`/`c3Types` at 162-177 have no `box`).
- **`edges`/styling/`lineLength`/`width`/`height` documented but inert** (see next section).
- **README CLI reference omits `--expanded` and `--label-ratio` flags and documents the default output dir as `.` when the code defaults to the input file's directory** (`README.md:350-360` vs `cmd/c4drill/root.go:104-108`).

### Parsed-but-unused model fields: per-unit styling overrides do nothing
- **What happens:** `model.Unit.Color/Border/Style/Width/Height/Edges` and `model.Properties.Color/Style/Border/LineLength` are parsed from TOML (`internal/model/unit.go:51-61`, `internal/model/properties.go:4-21`) but are never read by the graph builder or renderer. Node/cluster styling is computed solely from unit *type* via `GetStyleForType` (`internal/graph/shapes.go:154-207`) and `GetBoxStyleByContents` (`shapes.go:111-127`); `buildNode` (`internal/graph/builder.go:233-262`) never consults the unit's `Color`, `Border`, `Style`, `Width`, or `Height`.
- **Impact:** The README "Styling" section (`README.md:256-267`) and the entire `skill/examples/04-styling.toml` demonstrate a feature that has no effect. `width`/`height` (auto-layout overrides) and `lineLength` are likewise dead.
- **Files:** `internal/model/unit.go:51-61`, `internal/model/properties.go:4-21`, `internal/graph/builder.go:233-262,265-311`, `internal/graph/shapes.go:111-127,154-207`.
- **Fix approach:** Either wire these fields into `buildNode`/`buildCluster` (override type defaults when set), or remove them from the model and all documentation.

### `skill/examples/04-styling.toml` fails validation — the only CI job is red
- **What happens:** The repo's only workflow, `.github/workflows/validate-skill-examples.yml`, builds the binary and runs it against every `skill/examples/*.toml`. Verified locally that `skill/examples/04-styling.toml` exits with a validation failure (`container must be C1 type (inside C1 box)` for `internal.auth` and `internal.config`).
- **Impact:** Any push/PR touching `skill/**` fails CI. The file also demonstrates the dead styling fields.
- **Fix approach:** Re-express the example with valid nesting (use `containerBox` or hoist `internal.*` to top level), and only reference fields that render.

### Tracked build artifacts and dead assets
- **`cmd/c4drill/c4drill`** — a 12.6 MB Mach-O binary is committed to git (`.gitignore` ignores `/c4drill` but not `cmd/c4drill/c4drill`). `.git` is 33 MB largely due to this.
- **`data/*.svg`** (`data/person.svg`, `data/db.svg`, etc.) — tracked but referenced by no code; the icon system was replaced with emoji/native shapes (`internal/render/render.go:42-49` and `internal/render/labels.go:417-425` note this). Dead assets.
- **`internal/render/labels.go:422-425`** `iconTypeForUnit` is dead code (flagged `unused` by lint), as is `labelMaxChars` (`internal/render/wrap.go:159`).
- **Fix approach:** `git rm cmd/c4drill/c4drill` and `data/`, delete the two unused functions.

### Hardcoded version — `--version` always prints "dev"
- **Files:** `cmd/c4drill/root.go:41` (`version = "dev"`), `cmd/c4drill/root.go:64` (`Version: version`). No `-ldflags "-X ..."` wiring anywhere (no Makefile; `.mise.toml` has only test/lint tasks).
- **Impact:** `c4drill --version` reports `dev` for every build, defeating issue triage.
- **Fix approach:** Inject via ldflags and document in `.mise.toml`/README; default to `dev` only for unreleased builds.

### Lint debt: 122 golangci-lint violations, no CI enforcement
- **What happens:** `golangci-lint run ./...` reports 122 issues with the repo's own config (`golangci-lint 2.12.2`): `goconst` 50, `wsl_v5` 16, `gocognit` 10, `mnd` 7, `nlreturn` 6, `funlen` 4, `lll` 5, `modernize` 4, `exhaustive` 1, `errorlint` 1, `errcheck` 2, `unused` 2, `unparam` 1, plus others.
- **Files:** e.g. `internal/render/converter.go:172-173` (unchecked `cg.Attr`), `internal/parser/errors.go:47` (`errorlint` — type assertion fails on wrapped errors), `internal/graph/builder.go:514` (`exhaustive`), `cmd/c4drill/root.go:170` (`unparam` — unused `m`), `internal/graph/path.go:59,123` (`modernize` string-builder).
- **Impact:** The `.mise.toml` `lint` task exists but no CI job runs it; the `.golangci.yml` `exclusions` only patch one errcheck case, so the baseline is red by default.
- **Fix approach:** Triage in batches; prioritize `unused`, `unparam`, `errcheck`, `errorlint`, `exhaustive`; add a CI job running `golangci-lint run --fix` with a clean gate.

### Complexity hotspots exceed configured thresholds
- **Files:** `internal/view/scope.go:652` `resolveBoundaryNodeLinks` (cognitive complexity 33), `internal/view/scope.go:278` `resolveAndAddBoundary` (24), `internal/parser/parser.go:160` `parseUnitWithOrder` (23), `internal/render/wrap.go:41` `wrapText` (23), `internal/parser/parser.go:100` `captureDefinitionOrder` (20), `internal/validator/rules.go:198` `validateUnitNesting` (19). All exceed the `gocognit.min-complexity: 15` gate in `.golangci.yml`.
- **Impact:** These are exactly the functions where the v1.8 root-cause bugs lived (see `.planning/STATE.md`). Link resolution in `internal/view/scope.go` (5 functions over threshold) is the highest-risk area.

## Security Considerations

**Overall threat model:** CLI tool, user-supplied TOML on the command line — no network surface. Findings below are defense-in-depth notes, not active exploits.

- **HTML/label injection is neutralized by graphviz escaping (with a side effect).** User-controlled `name`/`description`/`properties.name` flow into labels. Node labels are correctly escaped before HTML-label construction via `wrapAndEscape`/`html.EscapeString` (`internal/render/wrap.go:116-130`). The navigation HTML string (`internal/render/navigation.go:22-36`) is NOT escaped, but because it is set as a plain-text graph label, graphviz escapes it in the SVG output — verified that `<script>` appears as `&lt;script&gt;`, not as an executable element. The side effect is the broken navigation (see Critical Bug above). Once the nav is fixed with `StrdupHTML`, the raw text MUST be escaped first, or a malformed unit name could inject arbitrary HTML into the diagram.
- **Output path handling:** Unit names flow into output paths via `strings.ReplaceAll(unitPath, ".", "/")` then `filepath.Join` (`internal/output/writer.go:37-49`). The dot-to-slash replacement plus `filepath.Join` cleaning happens to neutralize `..` segments and leading slashes for typical inputs, so no traversal was demonstrated; however, unit names are never validated as path-safe identifiers, and any change to this transform could reopen traversal. Explore/back-link URLs are built by string concatenation (`internal/graph/path.go:14-66`).
  - **Recommendation:** Validate unit names against a safe identifier pattern (no `/`, `..`, leading `.`) at parse/validate time; add a writer test that a hostile unit name cannot escape `baseDir`.
- **No input size limits.** A very large TOML is unmarshaled to `map[string]any` and each unit re-marshaled/unmarshaled (`internal/parser/parser.go:47-96`) — memory use is several multiples of input size. Acceptable for a CLI; becomes a DoS vector if ever exposed as a service.
- **`gosec` G304 suppressed** on `os.ReadFile` (`internal/parser/parser.go:321-322`) — intentional (path is user-provided via CLI); keep the suppression comment.

## Performance Bottlenecks

- **WASM graphviz instance per diagram, globally serialized.** Every render creates a new `graphviz.New(ctx)` and takes the package-level `wasmMutex` (`internal/render/render.go:81-110`). A 17-sub-diagram architecture pays 17 sequential WASM-engine initializations. Any concurrency (e.g., future parallel view generation) is blocked by the mutex.
  - **Improvement path:** Reuse one graphviz instance per process guarded by the existing mutex (verify fork's thread-safety), or pool instances with per-instance locks; measure with a large model.
- **Parser triple round-trip.** `Parse` runs the unstable-order pass, unmarshals the whole document to `map[string]any`, then for every unit re-marshals and re-unmarshals (`internal/parser/parser.go:47-96,160-241`). Each `map[string]any` unit also recursively re-marshals its whole subtree, so deep nesting multiplies work.
- **Link resolution is O(links × path-depth) with map scans.** `resolveToViewAncestor` (`internal/view/scope.go:716-742`) walks path segments per link; `resolveBoundaryNodeLinks` (complexity 33) iterates the whole view for boundary-node resolution; `addResolvedBoundaryNode` (`scope.go:619-645`) strips path segments per link. Fine for tens of units, degrades on hundreds.

## Fragile Areas

- **`internal/view/scope.go` link/boundary resolution.** Six functions with cognitive complexity 17-33 implement hand-tuned peer-to-ancestor resolution and cross-subunit link hoisting. This is where the two v1.8 root-cause bugs were fixed (`internal/view/scope.go:253-363`, `647-711`, `744-883`; see `.planning/STATE.md`). Edge cases (deep nesting, self-links, multiple links to same peer, links through boundary nodes) are only partially covered (`internal/view/scope_test.go` at 73.1% coverage, the lowest of the well-tested packages).
- **Dependency on go-toml `unstable` API.** `internal/parser/parser.go:8-10,100-157` uses `github.com/pelletier/go-toml/v2/unstable`, an explicitly unstable package not covered by the library's compatibility guarantees. An upstream bump can silently change ordering behavior (which is already broken at depth 3).
- **Forked graphviz dependency.** `go.mod:28` `replace github.com/goccy/go-graphviz => github.com/onokonem/go-graphviz v0.0.0-20260321130544-f364b5235161` pins rendering to a single-maintainer fork. The lint config deliberately allow-lists this replace (`replace-allow-list` in `.golangci.yml`), so it is a conscious decision, but the fork must be maintained for the tool to keep building.
- **Global mutable state.** `render.LabelRatio` (`internal/render/wrap.go:36`) is mutated by the CLI (`cmd/c4drill/root.go:100`) and read during label sizing; `wasmMutex` (`internal/render/render.go:20`) serializes all rendering. Using the packages as a library (or embedding multiple `c4drill` runs in one process) is fragile — 91 test sites need `//nolint:paralleltest` because of it.
- **Stale comment claims.** `internal/parser/parser.go:83` ("Should not happen if captureDefinitionOrder is correct"), `internal/parser/parser.go:218` ("maintains backward compatibility"), `internal/graph/path.go:12-13` (format ignored), `cmd/c4drill/root.go:154-155` ("This can be overridden by properties.expanded") — the last one is flat-out wrong: `collectExpandableUnitPaths` never checks `expanded`.

## Scaling Limits

- **Rendering:** fully serialized by the WASM mutex; wall time scales linearly with diagram count. No parallelism without a render-engine refactor.
- **Model size:** parser re-marshaling and per-link ancestor walks degrade on large models (hundreds of units). No benchmarks exist in the repo to quantify the cliff.
- **Diagram size:** no handling for very wide C1 diagrams beyond label wrapping heuristics (magic constants `pointsPerChar`, `pointsPerRow`, `iconColumnWidth` in `internal/render/wrap.go:9-30` are tuned constants, not derived from fonts).

## Dependencies at Risk

- **`github.com/onokonem/go-graphviz`** (fork replacing `goccy/go-graphviz`, `go.mod:28`) — maintenance and upstream-sync risk; pinned to an immutable pseudo-version.
- **`github.com/pelletier/go-toml/v2` `unstable` API** (`internal/parser/parser.go:9`) — no stability guarantee; the order-capture logic depends on its exact behavior.
- **`github.com/tetratelabs/wazero`** v1.10.1 (indirect, powers the WASM graphviz engine) — the concurrency limitation is inherited from this engine.
- **Go 1.26.1 in `go.mod` with `errors.AsType`** (`internal/parser/errors.go:58`) — requires Go 1.25+; bleeding-edge toolchain requirement narrows the buildable toolchain set. Local is 1.26.5.

## Missing Critical Features

- **`c4drill validate` subcommand** — documented in `skill/SKILL.md` (lines 434, 531-541) and referenced by the skill's "Always validate generated TOML" guidance, but not implemented. Validation only happens implicitly when generating output.
- **Working clickable navigation in SVG** — the flagship feature is broken (see Critical Bug).
- **Real version reporting** — `--version` prints `dev`.
- **Per-unit styling overrides** (`color`, `border`, `style`, `width`, `height`) — parsed, documented, demonstrated, but never applied.
- **CI for tests and lint** — the only workflow validates skill examples (currently failing); there is no `go test` or `golangci-lint` CI gate, so regressions like the navigation bug ship silently.
- **Deterministic output** — needed for reproducible builds of docs sites; currently broken at depth-3 nesting and for validation error ordering.

## Test Coverage Gaps

- **`internal/model` at 0.0%** — no test files for `internal/model` (`unit.go`, `link.go`, `properties.go`, `colors.go`). Link/Unit field mapping, arrow/rank/label-position enum values, and `FindLinkByPeer` are only exercised indirectly.
- **Depth-3 ordering untested** — `TestParseSubunitDefinitionOrder` (`internal/parser/parser_test.go:869-904`) covers only C2; no test asserts C3 (`[a.b.c]`) order, which is why the non-determinism bug shipped.
- **Arrow enum mismatch untested** — no test asserts the rendered `dir` attribute for `arrow = "bidirectional"`.
- **Navigation assertions too weak** — `internal/render/integration_test.go:424-457` assert text presence, not clickable anchors; the broken back-link/breadcrumb feature passes CI.
- **Style override fields untested** — no test exercises `unit.Color/Border/Style/Width/Height` or `properties.LineLength` because they have no behavior; any future implementation must add coverage.
- **View link-resolution edge cases** — `internal/view` at 73.1%; `resolveBoundaryNodeLinks`, `addResolvedBoundaryNode`, `resolveToViewAncestor` boundary conditions (peer at same level, peer under a boundary node, self-references through nesting) lack dedicated tests.
- **Priority: High** for navigation (broken feature, weak assertions), C3 ordering, and arrow mapping; **Medium** for model package and view edge cases; **Low** for style-override fields until the feature is implemented.

---

*Concerns audit: 2026-08-05*
