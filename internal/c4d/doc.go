// Package c4d implements the C4D front-end: parsing of the .c4d brace-block
// DSL (the less verbose alternative to the TOML diagram definition,
// 35-01-PLAN.md) into a typed, comment/position-aware AST, wrapping all
// failures in the repo's standard *parser.ParseError contract.
//
// The package is layered: internal/c4d/grammar holds the pigeon PEG grammar
// and its generated parser; internal/c4d/ast holds the typed syntax tree;
// this package exposes the Parse/ParseFile entry points.
package c4d

//go:generate go generate ./grammar
