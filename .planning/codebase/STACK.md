# Technology Stack

**Analysis Date:** 2026-08-05

## Languages

**Primary:**
- Go 1.26.1 - Entire codebase (`go.mod`, line 3). Local toolchain is go1.26.5 darwin/arm64. 53 `.go` files, 27 test files.

**Secondary:**
- None. No TypeScript/JavaScript, Python, or other languages in the repo.
- TOML is an input data format (user-authored architecture files), not a build language.

## Runtime

**Environment:**
- Native Go binary — compiles to a single static-ish executable CLI (`c4drill` binary at repo root, built from `cmd/c4drill/`)
- Cross-platform by design (darwin/arm64 dev machine; GitHub Actions builds on ubuntu-latest)
- Bundles a WASM-compiled Graphviz engine via `github.com/tetratelabs/wazero` — no system Graphviz installation required at runtime

**Package Manager:**
- Go modules (no external package manager)
- Lockfile: `go.sum` present
- Module path: `github.com/Djarvur/c4drill` (`go.mod`, line 1)

## Frameworks

**Core:**
- `github.com/spf13/cobra` v1.10.2 - CLI command framework (`cmd/c4drill/root.go`)
- `github.com/goccy/go-graphviz` (forked via `replace` to `github.com/onokonem/go-graphviz` v0.0.0-20260321130544-f364b5235161) - Graph/DOT/SVG rendering (`internal/render/`)
- `github.com/tetratelabs/wazero` v1.10.1 - WASM runtime executing the Graphviz WASM engine (indirect dep, used by go-graphviz)

**Testing:**
- `github.com/stretchr/testify` v1.11.1 - assertion library (`require`/`assert`)
- Standard `go test` runner, `-race` and `-cover` flags used in mise task (`go test -v -race -cover ./...` in `.mise.toml`)

**Build/Dev:**
- `golangci-lint` v2 (configured in `.golangci.yml`, pinned in `.mise.toml`)
- `mise` (`.mise.toml`) - tool version manager pinning Go 1.26.1 and golangci-lint 2, also defines `test` and `lint`/`lint-fix` tasks

## Key Dependencies

**Critical:**
- `github.com/goccy/go-graphviz` (onokonem fork) - All diagram rendering; the `replace` directive in `go.mod` (line 28) swaps the upstream library for a fork. Allowed by lint config (`gomoddirectives.replace-allow-list` in `.golangci.yml`). Uses WASM Graphviz engine — thread-safety requires a global `wasmMutex` in `internal/render/render.go` (lines 15-20)
- `github.com/pelletier/go-toml/v2` v2.2.4 - TOML input parsing, including the `unstable` API used to preserve unit definition order (`internal/parser/parser.go`)
- `github.com/spf13/cobra` v1.10.2 - CLI parsing, flags, help/version

**Infrastructure:**
- `github.com/tetratelabs/wazero` v1.10.1 - WASM runtime (indirect, powers graphviz)
- `github.com/agnivade/levenshtein` v1.2.1 - typo suggestions for validator error messages (`internal/validator/suggest.go`, max distance 2)
- Indirect rendering/font deps (from go-graphviz fork): `github.com/disintegration/imaging`, `github.com/flopp/go-findfont`, `github.com/fogleman/gg`, `github.com/golang/freetype`, `golang.org/x/image`, `golang.org/x/text`, `github.com/spf13/pflag`, `github.com/inconshreveable/mousetrap`, `github.com/davecgh/go-spew`, `github.com/pmezard/go-difflib`, `gopkg.in/yaml.v3`

## Configuration

**Environment:**
- Single env var: `C4DRILL_LABEL_RATIO` - overrides the width:height ratio for unit labels (read in `cmd/c4drill/root.go:301`). Precedence: CLI flag `--label-ratio` > env var > default 1.6
- No `.env` files are used or committed (`.env`/`.env.*` gitignored in `.gitignore`)

**Build:**
- `go.mod` / `go.sum` - module definition and checksums
- `.golangci.yml` - golangci-lint v2 config: enables `all` linters with targeted disables (`wsl`, `exhaustruct`, `depguard`, `noinlineerr`, `ireturn`, `varnamelen`, `tagliatelle`, `dupl`, `cyclop`) and settings (`gocognit.min-complexity: 15`, gomoddirectives replace allow-list, errcheck exclusion for deferred `Close()`)
- `.mise.toml` - tool versions and dev tasks

**CLI Configuration (runtime):**
- `-f, --format` (dot|svg, default svg) - `cmd/c4drill/root.go:70`
- `-o, --output` (output directory, default = input file dir) - `cmd/c4drill/root.go:72`
- `--expanded` (generate all-expanded diagram) - `cmd/c4drill/root.go:74`
- `--label-ratio` (label width:height, default 1.6) - `cmd/c4drill/root.go:76`

## Platform Requirements

**Development:**
- Go 1.26.1+ (mise-managed or system Go)
- golangci-lint v2 for linting
- No system Graphviz required (WASM engine bundled)
- macOS arm64 (current dev machine) or any Go-supported OS

**Production:**
- Distributed as source (`go install github.com/Djarvur/c4drill/cmd/c4drill@latest` per `README.md`) or prebuilt binary
- GitHub repository: `github.com/Djarvur/c4drill` (git remote `git@github.com:Djarvur/c4drill.git`)
- Versioned with git tags (`v1.0`, `v1.7`)

---

*Stack analysis: 2026-08-05*
