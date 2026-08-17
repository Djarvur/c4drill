# Phase 35 Deferred Items (Plan 35-09 execution)

| Category | Item | Status | Found At |
|----------|------|--------|----------|
| Tooling | Pre-existing gofmt drift in 5 committed files (not touched by 35-09): internal/c4d/composition_test.go, internal/c4d/grammar/reserved.go, internal/graph/shapes.go, internal/parser/inference_test.go, internal/template/expand.go — flagged by the local `gofmt -l` run; likely a gofmt/toolchain version difference. Out of 35-09 scope (scope boundary: pre-existing, unrelated files). | open, low-severity | 2026-08-14, 35-09 Task 1 |
| Docs | Pre-existing "Why TOML?" section argued against bespoke DSLs; 35-09 appended a dual-format pointer sentence rather than rewriting the section (minimal-change stance). A future docs pass may restructure the pitch. | noted | 2026-08-14, 35-09 Task 2 |
