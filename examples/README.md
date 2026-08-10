# C4Drill Examples

This directory contains examples **borrowed from the
[likec4](https://github.com/likec4/likec4) project** — specifically its
[`examples/`](https://github.com/likec4/likec4/tree/main/examples)
directory — translated into C4Drill TOML and rendered with `c4drill`.

likec4 is MIT-licensed (Copyright (c) 2023-2026 Denis Davydkov); this
project is MIT-licensed as well (see the root
[LICENSE](../LICENSE)). Each example documents its adaptations — which
likec4 features were mapped onto C4Drill equivalents and which were
dropped — in the comments of its TOML file.

| Directory | likec4 original | Notes |
|---|---|---|
| `cloud-system/` | [likec4 `examples/cloud-system`](https://github.com/likec4/likec4/tree/main/examples/cloud-system), a copy of [likec4/example-cloud-system](https://github.com/likec4/example-cloud-system) | Flagship: actors, system + containers, external provider; C1/C2/C3 drill-down + `--expanded` diagram |
| `overflow-test/` | [likec4 `examples/overflow-test`](https://github.com/likec4/likec4/tree/main/examples/overflow-test) | Long-name / long-label stress test |
| `rank-for-better-layout/` | [likec4 `examples/rank-for-better-layout`](https://github.com/likec4/likec4/tree/main/examples/rank-for-better-layout) | Rank hints (`rank = "equal"`) |

The generated SVG diagrams are committed next to each TOML file, so the
directory is self-contained.

## Regenerating

```bash
c4drill examples/cloud-system/cloud-system.toml
c4drill examples/cloud-system/cloud-system.toml --expanded
c4drill examples/overflow-test/overflow-test.toml
c4drill examples/rank-for-better-layout/rank-for-better-layout.toml
```

See the [Examples section of the main
README](../README.adoc#examples) for details and screenshots.
