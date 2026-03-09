# Stack Research

**Domain:** Go CLI for C4 Architecture Diagram Generation
**Researched:** 2026-03-09
**Confidence:** HIGH

## Recommended Stack

### Core Technologies

| Technology | Version | Purpose | Why Recommended |
|------------|---------|---------|-----------------|
| Go | 1.24+ | Runtime and build system | Latest stable (Feb 2025) with Swiss Tables maps optimization, improved garbage collector, and better type inference. Required for modern Go development patterns. |
| Cobra | v1.9+ | CLI framework | Industry standard used by Kubernetes, Hugo, GitHub CLI, Docker. Provides command structure, flag parsing, help generation, and shell completions out of the box. Single-command tools benefit from its mature ecosystem and documentation. |
| BurntSushi/toml | v1.6.0 | TOML parsing | Most widely adopted TOML library in Go ecosystem (15k+ stars). Supports TOML v1.1.0 spec, handles nested structures cleanly, and provides both strict and lenient parsing modes. Ideal for configuration-heavy CLI tools. |
| goccy/go-graphviz | v0.2.9 | GraphViz DOT/SVG generation | Pure Go implementation with WASM-embedded GraphViz - no external binary dependency. Generates DOT and renders SVG directly. Critical for portable CLI distribution. |

### Supporting Libraries

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| stretchr/testify | v1.10+ | Testing assertions | For all unit tests - provides assertions, mocks, and suite patterns |
| go-playground/validator | v10+ | Struct validation | For validating TOML input against schema rules (reference integrity, type constraints) |
| fatih/color | v1.18+ | Terminal coloring | For CLI output formatting (errors, warnings, success messages) |

### Development Tools

| Tool | Purpose | Notes |
|------|---------|-------|
| golangci-lint | Static analysis | Use with `.golangci.yml` - enables errcheck, staticcheck, govet, ineffassign |
| go test | Unit testing | Built-in, use with `-race -cover` flags |
| goreleaser | Release automation | For building cross-platform binaries (Linux, macOS, Windows) |

## Installation

```bash
# Core
go get github.com/spf13/cobra@latest
go get github.com/BurntSushi/toml@latest
go get github.com/goccy/go-graphviz@latest

# Supporting
go get github.com/stretchr/testify@latest
go get github.com/go-playground/validator/v10@latest
go get github.com/fatih/color@latest

# Development
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/goreleaser/goreleaser/v2@latest
```

## Alternatives Considered

| Recommended | Alternative | When to Use Alternative |
|-------------|-------------|-------------------------|
| Cobra | urfave/cli v3 | When you need ultra-simple single-command CLIs without subcommands. urfave/cli has cleaner API for basic cases but less ecosystem support. |
| BurntSushi/toml | pelletier/go-toml v2 | When you need TOML mutation/writing or strict ordering preservation. pelletier supports round-trip editing; BurntSushi is read-optimized. |
| goccy/go-graphviz | skaldebane/graphviz | When WASM size is critical (skaldebane is smaller). goccy has better documentation and more active maintenance. |
| Go 1.24 | Go 1.21+ | Only if you must support older environments. PROJECT.md specifies 1.21+ minimum; 1.24 recommended for performance gains. |

## What NOT to Use

| Avoid | Why | Use Instead |
|-------|-----|-------------|
| flag stdlib | Limited features, no subcommands, poor help generation | Cobra - production-grade CLI needs more than stdlib offers |
| graphviz binary via exec | Requires external dependency, breaks portability | goccy/go-graphviz - pure Go, no system dependencies |
| XML/JSON for config | Project specifies TOML; TOML is more human-readable for nested config | BurntSushi/toml - as specified in requirements |
| go-dot | Abandoned, no SVG rendering, DOT output only | goccy/go-graphviz - actively maintained, full rendering |

## Stack Patterns by Variant

**If targeting minimal binary size:**
- Use urfave/cli v3 instead of Cobra (saves ~500KB)
- Strip debug symbols with `-ldflags="-s -w"`
- Because Cobra's feature richness adds binary weight

**If needing TOML writing/modification:**
- Consider pelletier/go-toml v2 alongside or instead
- Because pelletier has better mutation support
- BurntSushi is optimized for reading only

**If CI/CD integration is priority:**
- Use goreleaser for automated releases
- Configure `.goreleaser.yml` for multi-platform builds
- Because manual release process is error-prone

## Version Compatibility

| Package A | Compatible With | Notes |
|-----------|-----------------|-------|
| go-graphviz v0.2.9 | Go 1.21+ | Requires CGO for some features; pure Go mode available |
| Cobra v1.9 | Go 1.21+ | No breaking changes from v1.8, safe upgrade |
| BurntSushi/toml v1.6 | Go 1.21+ | API stable since v1.0, backward compatible |
| testify v1.10 | Go 1.21+ | Works with Go modules without issue |

## Confidence Assessment

| Component | Confidence | Reason |
|-----------|------------|--------|
| Go 1.24 | HIGH | Official release, stable since Feb 2025 |
| Cobra | HIGH | Industry standard, verified via pkg.go.dev and GitHub |
| BurntSushi/toml | HIGH | Most adopted library, active maintenance confirmed |
| goccy/go-graphviz | MEDIUM | Best pure-Go option but smaller ecosystem than alternatives requiring binary |
| Development tools | HIGH | Standard Go ecosystem tooling |

## Sources

- pkg.go.dev/github.com/goccy/go-graphviz — Version v0.2.9, documentation, API reference (HIGH confidence)
- github.com/goccy/go-graphviz — Active maintenance, WASM-embedded GraphViz (HIGH confidence)
- pkg.go.dev/github.com/BurntSushi/toml — Version v1.6.0, TOML v1.1.0 support (HIGH confidence)
- github.com/spf13/cobra — Industry adoption, documentation (HIGH confidence)
- pkg.go.dev/github.com/spf13/cobra — Latest v1.9.x versions (HIGH confidence)
- go.dev/doc/go1.24 — Go 1.24 release notes, Swiss Tables optimization (HIGH confidence)
- github.com/urfave/cli — Alternative CLI framework, v3.x development (MEDIUM confidence)

---
*Stack research for: Go C4 Diagram CLI*
*Researched: 2026-03-09*
