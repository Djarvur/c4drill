---
created: 2026-08-10T12:20:48.105Z
title: Wrap labels at word boundaries only
area: docs
resolves_phase: 34
files:
  - /Users/nil/DiskD/W/yadro/cyp-mise-repo/docs/mise-architecture.toml
---

## Problem

While inspecting the generated `mise-architecture.toml`
(`/Users/nil/DiskD/W/yadro/cyp-mise-repo/docs/mise-architecture.toml`), long words
turn out to be split mid-word by the label alignment/wrapping procedure. Line
division must happen only at word boundaries — never inside a word.

## Solution

Change the wrapping/alignment procedure so lines break exclusively at word
boundaries. No mid-word (character-level) breaking fallback: if the resulting unit
shape is unacceptable for a given label, the document author can simply choose a
different word — do not try to force the text into the rectangle by splitting words.
