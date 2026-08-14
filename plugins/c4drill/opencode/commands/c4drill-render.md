---
description: Render C4 architecture diagrams from a C4Drill TOML or C4D definition using the c4drill CLI (SVG/HTML/DOT).
---

Run the installed `c4drill` CLI on the user's C4Drill file — TOML (`.toml`) or C4D (`.c4d`) — to validate it and render C4 architecture diagrams. Input dispatch is by extension; both formats run the same pipeline.

## Usage

```
c4drill <input.toml|input.c4d> [flags]
```

- **Validate/render**: `c4drill architecture.toml` / `c4drill architecture.c4d` - silent output + exit 0 = success; writes `architecture.svg` next to the input by default
- **Output directory**: `-o ./docs/diagrams`
- **Formats**: `-f svg|html|dot` (default `svg`; use `-f html` for Safari/WebKit or `file://` viewing; `-f dot` for GraphViz tooling)
- **All-expanded diagram**: `--expanded`
- **Failure**: non-zero exit code with an error message including the input line number

## Related commands

- `c4drill convert to-c4d <file.toml>` / `c4drill convert to-toml <file.c4d>` - translate between the two formats (validated first; add `--follow-includes` to convert a whole include graph)
- `c4drill fmt [--check] <file|dir>...` - format both formats in place (gofmt-style); `--check` is the CI gate

## Prerequisite

`c4drill` must be installed: `go install github.com/Djarvur/c4drill/cmd/c4drill@latest`

If `c4drill` is not on PATH, tell the user the prerequisite instead of guessing or fabricating output.
