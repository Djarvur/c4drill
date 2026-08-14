// Package grammar contains the pigeon-generated parser for the C4D DSL
// grammar defined in c4d.peg (35-01-PLAN.md). The generated parser is
// committed (parser_gen.go) so downstream builds never need pigeon; the
// tool itself is pinned in go.mod via tools.go at the repo root (D-20).
//
// Regenerate after editing c4d.peg:
//
//	go generate ./internal/c4d            (from the repo root)
//	go generate github.com/Djarvur/c4drill/internal/c4d
package grammar

//go:generate pigeon -o parser_gen.go -nolint c4d.peg
