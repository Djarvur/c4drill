---
status: testing
phase: 07-ai-documentation
source: 07-01-SUMMARY.md, 07-02-SUMMARY.md
started: 2026-03-11T04:00:00Z
updated: 2026-03-18T07:30:00Z
---

## Current Test

number: 3
name: Nested structure example validates
expected: Running `c4drill skill/examples/02-nested.toml` produces valid output with nested clusters
awaiting: user response

## Tests

### 1. SKILL.md exists and is complete

expected: skill/SKILL.md file exists with complete TOML schema reference (~392 lines)
result: pass

### 2. Minimal TOML example validates

expected: Running `c4drill skill/examples/01-minimal.toml` produces valid output without errors
result: issue
reported: "Only 1 edge generated but 2 links defined (user.linkFrom and webapp.link). Both should appear - deduplication incorrectly merges distinct user-defined links."
severity: major

### 3. Nested structure example validates

expected: Running `c4drill skill/examples/02-nested.toml` produces valid output with nested clusters
result: [pending]

### 4. Links example validates

expected: Running `c4drill skill/examples/03-links.toml` produces valid output with various link types
result: [pending]

### 5. Styling example validates

expected: Running `c4drill skill/examples/04-styling.toml` produces valid output with custom colors and styles
result: [pending]

### 6. E-commerce example validates

expected: Running `c4drill skill/examples/05-ecommerce.toml` produces valid output for a complex architecture
result: [pending]

### 7. GitHub Actions workflow exists

expected: .github/workflows/validate-skill-examples.yml file exists and has correct structure
result: [pending]

## Summary

total: 7
passed: 1
issues: 1
pending: 5
skipped: 0

## Gaps

- truth: "Both user-defined links (linkFrom and link) appear as edges in output"
  status: failed
  reason: "User reported: Only 1 edge generated but 2 links defined (user.linkFrom and webapp.link). Both should appear - deduplication incorrectly merges distinct user-defined links."
  severity: major
  test: 2
  root_cause: ""
  artifacts: []
  missing: []
  debug_session: ""
