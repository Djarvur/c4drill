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
downloads verifier IDE builds; run it where those downloads are reachable.

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
