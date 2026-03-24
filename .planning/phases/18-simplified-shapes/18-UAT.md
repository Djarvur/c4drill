---
status: testing
phase: 18-simplified-shapes
source: 18-01-SUMMARY.md
started: 2026-03-24T08:10:00Z
updated: 2026-03-24T08:10:00Z
---

## Current Test

number: 1
name: DB Cylinder Shape
expected: |
  Generate a diagram with a DB unit. The DB should appear as a native GraphViz cylinder shape (3D cylinder icon, like a classic database symbol). No custom SVG icons.
awaiting: user response

## Tests

### 1. DB Cylinder Shape
expected: Generate a diagram with a DB unit. The DB should appear as a native GraphViz cylinder shape (3D cylinder, like a classic database). No custom SVG.
result: pass

### 2. Queue Horizontal Cylinder
expected: Generate a diagram with a Queue unit. The Queue should appear as a horizontal cylinder (cylinder rotated 90°), like a pipe/tube. No custom SVG.
result: pending

### 3. Person Emoji Label
expected: Generate a diagram with a Person unit. The label should show 👤 emoji on the left, name/description on the right. No SVG icons.
result: pending

### 4. Simplified System/Box Labels
expected: Generate diagrams with System, Box, Container, Component units. Labels should be simple 3-row tables (name, technology, description) with no icon column.
result: pending

### 5. No Icon Directory Generated
expected: Generate any diagram. No `.icons/` directory should be created in the output location.
result: pending

### 6. Word-Wrap Still Works
expected: Run with `--label-ratio=1.6` flag. Labels should be word-wrapped to maintain proportions. Flag should work without errors.
result: pending

## Summary

total: 6
passed: 0
issues: 0
pending: 6
skipped: 0

## Gaps

(none yet)
