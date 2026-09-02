# C4Drill — JetBrains IDE plugin

Syntax highlighting, autocomplete, diagnostics, formatting, and a live
diagram preview for the [c4drill](https://github.com/Djarvur/c4drill)
architecture-as-code tool inside IntelliJ IDEA, GoLand, PyCharm, WebStorm,
PhpStorm, RubyMine, CLion, DataGrip, Rider, and other IntelliJ-platform IDEs.

The plugin is a **thin client over the `c4drill` language server** (issue
#32): it launches `c4drill serve --lsp` and speaks LSP to it. No c4drill
logic is duplicated here — diagnostics, completion, hover, go-to-definition,
structure view, formatting, and the rendered diagram all come from the same
pipeline the CLI runs, so editor output matches `c4drill <file>` exactly.

## Features

- **C4D highlighting** — the `.c4d` grammar reused verbatim from the VS Code
  extension (#27) (`extension/syntaxes/c4d.tmLanguage.json`), served through
  IntelliJ's TextMate bundles support. `.c4d` files get the C4D file type and
  icon unconditionally.
- **Diagnostics** — parse/validate errors appear on open/save/edit with the
  same messages and line numbers the CLI prints.
- **Autocomplete** — all 17 unit types (nesting-aware, with
  `db`/`queue`/`box` promotion), per-unit and per-link fields, enum values,
  peers (bare + dotted), templates and `${param}` names, include paths.
- **Hover, Go to Definition, Structure view** — via `textDocument/hover`,
  `textDocument/definition`, and `textDocument/documentSymbol`.
- **Format Document** — `textDocument/formatting` for both `.c4d` and the
  c4drill TOML dialect (`Tools | C4Drill | Format Document`).
- **Live diagram preview** (`Tools | C4Drill | Show Preview`, Alt+Shift+P) —
  a "C4Drill Preview" tool window that renders the open model through the
  custom `c4drill/renderDiagram` request and re-renders, debounced (~200 ms),
  as you type:
  - click-through **drill-down navigation** (C1 → C2 → C3) on the SVG's
    internal links, with breadcrumbs in the tool window toolbar;
  - external `http(s)` reference links open in the system browser;
  - **view controls**: all-expanded mode, legend override (model default /
    on / off), expanded-set override, collapse-all — mapping directly onto
    the `renderDiagram` parameters;
  - parse/validate failures show the CLI-identical messages instead of a
    stale diagram;
  - **Export SVG…** saves the currently rendered diagram.
- **Validate File** (`Alt+Shift+V`) re-runs the full validation pipeline on
  demand.

## TOML scoping rule

IntelliJ IDEs ship their own TOML plugin, and c4drill must not hijack
unrelated TOML files. The rule mirrors the VS Code extension:

- `.c4d` files are always handled.
- `.toml` files get c4drill features **only when they opt in**, via either:
  - the `c4drill.toml.patterns` globs (Settings | Tools | C4Drill; one glob
    per line, relative to the project root or absolute, e.g.
    `models/*.toml` or `**/architecture/*.toml`), or
  - the explicit `Tools | C4Drill | Activate for This File` action
    (persisted per project; `Deactivate for This File` reverses it).
- Plain TOML files keep the built-in TOML plugin and are never touched.

## Installing the c4drill binary

The plugin discovers `c4drill` on your `PATH`. Alternatively point
Settings | Tools | C4Drill → *c4drill server path* at the binary
(`c4drill.server.path`); the server is launched as `c4drill serve --lsp`.

Build it from the repository root:

```sh
go build -o /usr/local/bin/c4drill ./cmd/c4drill
```

## Install

**JetBrains Marketplace** (preferred): search for "C4Drill" in
Settings | Plugins | Marketplace. *(Marketplace upload pending vendor
account; until then use the zip.)*

**From a zip**: `./gradlew buildPlugin` produces
`build/distributions/c4drill-idea-plugin-<version>.zip`. Install via
Settings | Plugins | ⚙ | *Install Plugin from Disk…*.

## Supported IDE versions

**2025.3 or newer of the commercial IntelliJ-platform IDEs** (IntelliJ IDEA
2025.3+ is the unified distribution — LSP support is available to all users;
GoLand, PyCharm, WebStorm, PhpStorm, RubyMine, CLion, DataGrip, Rider,
RustRover 2025.3+).

Why 2025.3:

1. The platform's built-in LSP client (`com.intellij.modules.lsp`) is what
   this plugin builds on. It became usable by third-party plugins in
   commercial IDEs with 2025.2, and the unified IDEA 2025.3 made it available
   to everyone (IntelliJ IDEA Community is sunset in favor of the unified
   distribution from 2025.3). Android Studio and older IDEA Community
   versions do not bundle it.
2. `textDocument/documentSymbol` surfaces in the Structure view only from
   2025.3.

The `com.intellij.modules.lsp` and `org.jetbrains.plugins.textmate`
dependencies are declared **optional**: on an IDE without one of them the
file type, settings, and actions still load, with the preview and
highlighting degraded accordingly.

## Building from source

```sh
cd idea-plugin
./gradlew build          # compile + unit tests
./gradlew buildPlugin    # installable zip in build/distributions/
./gradlew verifyPluginStructure
```

Requires JDK 21 (auto-provisioned via the Gradle toolchain). By default the
build downloads IntelliJ IDEA Ultimate 2025.3 from JetBrains; if
`download.jetbrains.com` is unreachable from your network, point
`C4DRILL_IDEA_HOME` at a local 2025.3+ IDEA installation and the build uses
it instead.

The full cross-version Plugin Verifier (`./gradlew verifyPlugin`) additionally
downloads verifier IDE builds; run it where those downloads are reachable. It
is configured to verify against the two ends of the declared compatibility
range — IntelliJ IDEA 2025.3 (oldest supported platform, since-build 253) and
2026.2.1 (newest stable platform at the time of #35); all commercial IDEs of
the same platform build share the API surface the verifier checks. Pass a
wider `ides { … }` list (or `recommended()`) if you want every product family
covered and have the disk space (~5 GB per product).

## Headless validation (issue #35)

What is covered without a human in front of the IDE, and how:

- **Cross-version Plugin Verifier** (`./gradlew verifyPlugin`): the plugin is
  verified **Compatible** against both ends of the compatibility range —
  IU-253.28294.334 (2025.3) and IU-262.9437.185 (2026.2.1). The one fatal
  finding the verifier ever reported (invoking `AnAction.actionPerformed`
  directly in the Format action — an `@ApiStatus.OverrideOnly` API) is fixed;
  Format Document now delegates via `ActionUtil.performActionDumbAware`.
  Remaining non-fatal notes, deliberately kept:
  - *Deprecated/experimental API usages on `C4drillPreviewToolWindowFactory`*
    (`getIcon`, `getAnchor`, `manage`, `isApplicable`, `isDoNotActivateOnStart`):
    the Kotlin compiler generates delegation bridges into every implementor of
    `ToolWindowFactory` (platform Kotlin `jvm-default` configuration); the
    plugin source overrides none of them.
  - *Deprecated API usages on the `com.intellij.platform.lsp.api` family in
    the 2026.2 report*: JetBrains deprecated that API in 262 in favor of a
    newer replacement that does not exist on 253. While the plugin supports
    since-build 253, the deprecated-but-present API is the only way to attach
    a language server; revisit when the minimum moves past 253.
- **Platform tests** (`src/test/kotlin/dev/djarvur/c4drill/platform/`, run by
  `./gradlew test` via the IntelliJ Platform test framework against the real
  2025.3 test distribution): `.c4d` file-type registration, C4D TextMate
  bundle extraction + `source.c4d` scope registration + real tokenization of
  an example file, global/project settings wiring, TOML scoping decisions on
  real VirtualFiles, and the LSP gateway extension wiring.
- **Language server end-to-end probe**: `c4drill serve --lsp` was probed over
  stdio LSP with the real binary — `initialize` (capabilities: completion,
  definition, hover, documentSymbol, formatting, incremental sync),
  `textDocument/didOpen` → `textDocument/publishDiagnostics` with
  CLI-identical messages for a broken model, and the custom
  `c4drill/renderDiagram` request returning an 18 KB SVG for
  `examples/cloud-system/cloud-system.toml`.

## Manual QA script (interactive `runIde` pass)

The items below need a human clicking through a live IDE; everything
statically checkable around them is covered above. Budget: ~10 minutes.

**Setup (once)**

1. Build and install:
   `./gradlew buildPlugin` →
   Settings | Plugins | ⚙ | *Install Plugin from Disk…* →
   `build/distributions/c4drill-idea-plugin-0.1.0.zip` → restart the IDE.
   *Expect:* no startup errors (`Help | Show Log in Files` has no
   `dev.djarvur.c4drill` stack traces).
2. Have a `c4drill` binary on `PATH` (or set *c4drill server path:* in
   Settings | Tools | C4Drill). Open a project containing the repository's
   `examples/` directory.

**A. File type + highlighting**

3. Open `skill/examples/03-links.c4d` in the repository checkout (or any
   `.c4d` file). *Expect:* C4D icon in the tab; colored highlighting
   (keywords `person`/`system`, strings, arrows `->`, `<->`, comments).

**B. LSP features (diagnostics, completion, hover, definition, format)**

4. Open `examples/cloud-system/cloud-system.toml`. *Expect:* no error
   annotations.
5. Break it: delete the `=` on any `key = "value"` line. *Expect:* within a
   second, a red underline whose message matches
   `c4drill <file>` output for the same broken file (e.g.
   `parse: parse error at line N: expected '=' after key`). Undo.
6. Completion: inside a unit block (`[A]`), press Ctrl+Space on a fresh
   line. *Expect:* unit fields (`type`, `name`, `technology`, …) suggested.
7. Hover: hover a unit id. *Expect:* doc popup from the language server.
8. Go to definition: Ctrl+Click (or Ctrl+B) on a `peer = "…"` target name.
   *Expect:* caret jumps to that unit's section.
9. Structure view: open the Structure tool window with the model active.
   *Expect:* units listed via `textDocument/documentSymbol`.
10. Format: scramble whitespace in the model, run
    Tools | C4Drill | Format Document. *Expect:* file reformatted;
    `c4drill <file>` still validates.

**C. Validate / scoping actions**

11. Alt+Shift+V on the valid model. *Expect:* "no problems found"-style
    balloon; on a broken model, error highlights refresh.
12. Open a plain `.toml` that is NOT a c4drill model (e.g. any unrelated
    TOML). *Expect:* no c4drill completions/diagnostics (built-in TOML
    plugin still owns it).
13. Run Tools | C4Drill | Activate for This File on it. *Expect:* c4drill
    features attach (diagnostics from the c4drill pipeline appear);
    Deactivate for This File reverses it.

**D. Preview tool window (JCEF + drill-through)**

14. Focus the model file, run Tools | C4Drill | Show Preview
    (Alt+Shift+P). *Expect:* the "C4Drill Preview" tool window opens on the
    right and renders the C1 diagram of `cloud-system.toml`.
15. Live re-render: type a change in the model (add a unit). *Expect:*
    ~0.2 s after you stop typing, the diagram updates.
16. Drill-through: click a container box that has an internal drill-down
    link. *Expect:* the preview navigates C1 → C2 (diagram of the nested
    system), toolbar shows breadcrumbs; click a breadcrumb to go back.
    *(Requires JCEF — a standard part of 2025.3 desktop IDEs.)*
17. External link: click a node configured with an `http(s)` reference.
    *Expect:* opens in the system browser, not inside the preview.
18. View controls: use the tool window toolbar controls — legend override
    (model default / on / off), all-expanded, collapse-all, expanded-set
    override. *Expect:* each changes the rendered diagram accordingly.
19. Error state: break the model while the preview is open. *Expect:* the
    tool window shows the CLI-identical error message(s) instead of a stale
    diagram; fixing it re-renders.
20. Export SVG…: run it from the tool window toolbar. *Expect:* a file
    chooser; the saved SVG opens in a browser and matches the preview.


## Layout

```
src/main/kotlin/dev/djarvur/c4drill/
  core/      pure logic: TOML scoping globs, document scope, render params,
             drill-down href resolution, debounce (plain-JUnit tested)
  lsp/       platform LSP integration: server provider/descriptor, the
             lsp4j surface for c4drill/renderDiagram, gateway
  preview/   C4Drill Preview tool window (JCEF), view controls, export
  actions/   Validate / Format / Show Preview / (De)Activate for This File
  highlighting/  TextMate bundle provider for the #27 C4D grammar
src/main/resources/
  META-INF/plugin.xml    core: file type, tool window, actions, settings
  META-INF/lsp.xml       optional com.intellij.modules.lsp wiring
  META-INF/textmate.xml  optional TextMate bundles wiring
  textmate/c4d/          the C4D TextMate bundle (grammar from #27)
  icons/                 file type / tool window icon
```
