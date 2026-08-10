# C4Drill plugin

Portable coding-agent plugin for [C4Drill](https://github.com/Djarvur/c4drill) -
C4 architecture diagrams from TOML definitions.

Version: 0.1.0 · Publisher: Djarvur

## Components

| Component | Name | What it does |
| --- | --- | --- |
| Skill | `c4drill-toml` | Authors valid C4Drill TOML architecture definitions (16 unit types, links, templates, includes; 9 example sets) |
| Command | `c4drill-render` | Runs the `c4drill` CLI to validate a TOML file and render SVG/HTML/DOT diagrams |

## Prerequisite

The `c4drill` binary must be installed on PATH:

```bash
go install github.com/Djarvur/c4drill/cmd/c4drill@latest
```

The command component only invokes the locally installed binary; it writes
diagram files to the output directory (default: next to the input file).

## Host support

| Host | Marketplace | Install |
| --- | --- | --- |
| Claude Code | `Djarvur/cc-mplace` | `claude plugin marketplace add Djarvur/cc-mplace` · `claude plugin install c4drill` |
| Codex | `Djarvur/cc-mplace` | `codex plugin marketplace add Djarvur/cc-mplace` · `codex plugin add c4drill` |
| GitHub Copilot CLI | `Djarvur/cc-mplace` | `copilot plugin marketplace add Djarvur/cc-mplace` · `copilot plugin install c4drill` |
| Grok Build | `Djarvur/cc-mplace` | `grok plugin marketplace add ...` · `grok plugin install c4drill` |
| Cursor | - | Import `https://github.com/Djarvur/c4drill` as a team marketplace (Cursor marketplaces require in-repo plugin sources; the cc-mplace Cursor entry is best-effort) |
| OpenCode | - | Copy `opencode/` into a project or global config: `cp -R opencode/. .opencode/` (project) or `cp -R opencode/. ~/.config/opencode/` (global) |

## Layout

- `.codex-plugin/`, `.claude-plugin/`, `.grok-plugin/`, `.plugin/`, `.cursor-plugin/` - per-host manifests (same component set, host-specific format)
- `skills/c4drill-toml/` - skill + examples (shared across all bundled hosts)
- `commands/c4drill-render.md` - command component
- `agents/openai.yaml` - Codex UI metadata
- `opencode/` - OpenCode copy source (`.opencode` is gitignored in this repo)
