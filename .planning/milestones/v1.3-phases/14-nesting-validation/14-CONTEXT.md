# Phase 14: Nesting Validation

## Problem Statement

C4Drill currently allows invalid nesting combinations that violate the C4 model hierarchy:
- C2 types (containers) could be placed at top level (should only be inside system/box)
- C3 types (components) could be placed inside system/box (should only be inside container)
- C1 types could be nested incorrectly

## C4 Nesting Hierarchy

```
Top Level → C1 types only (person, system, db, queue, box)

C1 containers (system, box) → C2 types only (container, containerDb, containerQueue)

C2 containers (container) → C3 types only (component, componentDb, componentQueue)

C3 types → No subunits allowed (leaf nodes)
```

## Current State

- `ValidateSubunitRules` only checks IF a type can have subunits, not WHAT types it can contain
- No validation exists for proper C4 level nesting

## Goal

Add a new validation rule `ValidateNestingHierarchy` that enforces:
1. Top-level units must be C1 types
2. C1 containers (system, box) can only contain C2 types
3. C2 containers (container) can only contain C3 types
4. External types follow same nesting rules as their base types

## Success Criteria

1. Top-level C2/C3 types are rejected with clear error message
2. C3 types inside system/box (skipping C2) are rejected
3. C2 types inside container are rejected
4. Existing valid TOML files continue to pass validation
5. All tests pass with new validation

## Files to Modify

- `internal/validator/rules.go` — add ValidateNestingHierarchy function
- `internal/validator/validator.go` — integrate new rule
- `internal/validator/rules_test.go` — add tests for new rule

## Dependencies

- None (standalone validation enhancement)
