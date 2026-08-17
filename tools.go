//go:build tools

// Package tools pins build-time code-generator dependencies in go.mod so a
// plain `go install`/`go generate` reproduces the committed generated code
// (the standard Go tools.go pattern). The `tools` build tag keeps this file
// out of regular builds — nothing here is linked into c4drill.
//
// Per D-20 (35-CONTEXT.md), github.com/mna/pigeon is the PEG parser
// generator used to produce internal/c4d/grammar/parser_gen.go from
// internal/c4d/grammar/c4d.peg. It is a build-time-only dependency with
// zero runtime imports; the generated parser is committed so consumers
// never need pigeon installed.
package tools

import (
	// Blank import pins the pigeon version in go.mod/go.sum (D-20, T-35-01-SC).
	_ "github.com/mna/pigeon"
)
