---
status: partial
phase: 16-embed-and-render-level-colored-svg-icons-for-units
source: [16-VERIFICATION.md]
started: "2026-03-20T23:10:00.000Z"
updated: "2026-03-20T23:10:00.000Z"
---

# Phase 16 — Human Verification

## Current Test

[awaiting human testing]

## Tests

### 1. Visual Icon Rendering

**expected:** SVG icons display at 32x32 pixels with colors matching each unit's C4 level border color

**test steps:**
```bash
go build -o c4drill ./cmd/c4drill
./c4drill skill/examples/05-complex.toml -o /tmp/test-icons
open /tmp/test-icons/05-complex.svg
```

**result:** [pending]

### 2. Icon File Extraction

**expected:** Icon files extracted to `.icons/` directory with `type-{hexcolor}.svg` naming

**test steps:**
```bash
ls -la /tmp/test-icons/.icons/
# Should show: person-3C7FC0.svg, db-3C7FC0.svg, etc.
```

**result:** [pending]

## Summary

total: 2
passed: 0
issues: 0
pending: 2
skipped: 0
blocked: 0

## Gaps
