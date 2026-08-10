---
name: c4drill-render
description: Render C4 architecture diagrams from a C4Drill TOML definition using the c4drill CLI (SVG/HTML/DOT).
---

Run the installed `c4drill` CLI on the user's C4Drill TOML file to validate it and render C4 architecture diagrams.

## Usage

```
c4drill <input.toml> [flags]
```

- **Validate/render**: `c4drill architecture.toml` - silent output + exit 0 = success; writes `architecture.svg` next to the input by default
- **Output directory**: `-o ./docs/diagrams`
- **Formats**: `-f svg|html|dot` (default `svg`; use `-f html` for Safari/WebKit or `file://` viewing; `-f dot` for GraphViz tooling)
- **All-expanded diagram**: `--expanded`
- **Failure**: non-zero exit code with an error message including the TOML line number

## Prerequisite

`c4drill` must be installed: `go install github.com/Djarvur/c4drill/cmd/c4drill@latest`

If `c4drill` is not on PATH, tell the user the prerequisite instead of guessing or fabricating output.
