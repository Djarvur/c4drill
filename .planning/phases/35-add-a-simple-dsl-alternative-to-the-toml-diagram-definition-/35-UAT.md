---
status: complete
phase: 35-add-a-simple-dsl-alternative-to-the-toml-diagram-definition-
source: [35-01-SUMMARY.md, 35-02-SUMMARY.md, 35-03-SUMMARY.md, 35-04-SUMMARY.md, 35-05-SUMMARY.md, 35-06-SUMMARY.md, 35-07-SUMMARY.md, 35-08-SUMMARY.md, 35-09-SUMMARY.md]
started: 2026-08-14T21:30:07Z
updated: 2026-08-17T00:00:00Z
---

## Current Test

[testing complete]

## Tests

### 1. Render a .c4d diagram directly
expected: `c4drill skill/examples/06-templates.c4d` renders the full C1/C2/C3 view set through the unchanged pipeline — same outputs as the equivalent .toml twin. Exit 0.
result: pass

### 2. Unknown extension fails closed
expected: `c4drill foo.json` exits non-zero with a parse error naming both accepted extensions (.toml, .c4d).
result: pass

### 3. Convert TOML to C4D (single file)
expected: `c4drill convert to-c4d <valid.toml>` writes the swapped-extension twin next to the input, exit 0; twin parses, validates, and renders.
result: pass

### 4. Convert refuses invalid input
expected: `c4drill convert to-c4d` on an orphan-invalid model (e.g. repo-root testdata/valid.toml) reports validation errors and writes NO output file.
result: pass

### 5. Convert C4D back to TOML
expected: `c4drill convert to-toml <file.c4d>` writes a canonical-order TOML twin that re-parses; source and twin are model-equivalent.
result: pass

### 2. Unknown extension fails closed
expected: `c4drill foo.json` exits non-zero with a parse error naming both accepted extensions (.toml, .c4d).
result: [pending]

### 3. Convert TOML to C4D (single file)
expected: `c4drill convert to-c4d <valid.toml>` writes the swapped-extension twin next to the input, exit 0; twin parses, validates, and renders.
result: [pending]

### 4. Convert refuses invalid input
expected: `c4drill convert to-c4d` on an orphan-invalid model (e.g. repo-root testdata/valid.toml) reports validation errors and writes NO output file.
result: [pending]

### 5. Convert C4D back to TOML
expected: `c4drill convert to-toml <file.c4d>` writes a canonical-order TOML twin that re-parses; source and twin are model-equivalent.
result: [pending]

### 6. Graph conversion preserves directory structure
expected: `c4drill convert to-c4d entry.toml --follow-includes -o out/` converts the whole include graph; relative layout preserved (e.g. out/domains/auth.c4d, not flat out/auth.c4d); include paths rewritten to .c4d, `once` preserved. Works with a RELATIVE entry path too (F-03 fix).
result: pass

### 7. Unrepresentable values hard-error (F-01 fix)
expected: Converting to C4D a model whose link technology contains a pipe (`|`), or whose unit id has spaces/dots, exits 1 with a loud error naming the offending value and file — no corrupt twin written. (Values the format CAN express — e.g. display names needing quoted TOML keys, quote-terminated multiline — convert correctly instead.)
result: pass

### 8. width/height round-trip parity (F-02 fix)
expected: A unit with `width = 300`/`height = 200` converts to C4D carrying `width: 300`/`height: 200`, and converts back to TOML with the values intact.
result: pass

### 9. fmt formats in place, comments preserved
expected: `c4drill fmt <file>` (both .toml and .c4d) rewrites canonical style in place; comments and blank-line grouping preserved; running fmt again changes nothing (idempotent).
result: pass

### 10. fmt --check CI gate
expected: `c4drill fmt --check` on a misformatted file exits 1, lists the offender, writes zero bytes; `c4drill fmt --check skill/examples/` exits 0.
result: pass

### 11. Reserved-word collision gets a suggestion
expected: A .c4d file using a reserved word as unit id (e.g. `description: system { }`) fails with `*parser.ParseError` at the DSL-native line number and a "did you mean" suggestion.
result: pass

### 12. Documentation and examples shipped
expected: README.adoc has a C4D Format section (syntax + 03-links side-by-side + convert/fmt CLI reference); skill/examples/ carries 12 .c4d twins (~50% fewer lines than their .toml twins); c4drill-toml skill documents both formats.
result: pass

## Summary

total: 12
passed: 12
issues: 0
pending: 0
skipped: 0

## Gaps

[none yet]
