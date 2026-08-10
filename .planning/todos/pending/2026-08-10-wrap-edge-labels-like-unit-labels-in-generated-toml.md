---
created: 2026-08-10T12:16:39.502Z
title: Wrap edge labels like unit labels in generated TOML
area: docs
resolves_phase: 34
files:
  - /Users/nil/DiskD/W/yadro/cyp-mise-repo/docs/mise-architecture.toml
---

## Problem

Inspecting the generated `mise-architecture.toml` in the cyp-mise-repo
(`/Users/nil/DiskD/W/yadro/cyp-mise-repo/docs/mise-architecture.toml`) reveals that
edge labels contain no line breaks. Edge labels must be formatted the same way as
unit labels: text wrapped into a rectangle with a specified aspect ratio, and the
rectangle borders must be invisible.

## Solution

Reuse the existing unit-label formatting logic (wrap text into a rectangle with the
configured aspect ratio, invisible borders) for edge labels in the generator.
