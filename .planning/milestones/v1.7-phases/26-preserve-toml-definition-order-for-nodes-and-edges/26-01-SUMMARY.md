---
plan: 26-01
status: complete
completed: 2026-03-25
---

# Phase 26-01: Preserve TOML Definition Order for Nodes and Edges

## Summary

Implemented TOML definition order preservation for nodes and edges in C4 diagrams. Instead of alphabetical sorting, the system now respects the order units are defined in the TOML file.

## Changes

### Model Layer
- Added `UnitOrder []string` to `parser.Model` struct
- Added `SubunitOrder []string` to `model.Unit` struct

### Parser Layer
- Implemented `captureDefinitionOrder()` using go-toml's unstable API
- Parser now makes two passes:
  1. First pass: captures definition order using unstable.Parser
  2. Second pass: unmarshals and builds model using captured order
- Updated error handling to support `unstable.ParserError`

### View Layer
- Updated `GenerateC1View`, `GenerateC2View`, `GenerateC3View`, `GenerateExpandedView`
- All functions now use explicit `UnitOrder` when available
- Fallback to map keys for test models without explicit order

### Graph Layer
- Updated `BuildGraph` and `BuildExpandedGraph` to use `view.UnitOrder`
- Updated `buildCluster` and `buildNestedCluster` to use `unit.SubunitOrder`
- Updated `buildEdges` to iterate in definition order

## Tests

- Updated `TestBuildGraphDeterministicOrder` tests to expect definition order
- All tests pass with new behavior

## Commits

1. `feat(26-01): add order tracking fields to model structs`
2. `test(26-01): add failing tests for definition order preservation`
3. `feat(26-01): implement order-preserving parser using unstable API`
4. `fix(26-01): propagate order through view and graph layers`
5. `fix(26-01): preserve TOML definition order for nodes and edges`
