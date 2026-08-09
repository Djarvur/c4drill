---
phase: 21-fix-box-labels-dashed-borders-validator-for-mixed-external-non-external-color-by-content
verified: 2026-03-24T19:30:00Z
status: passed
score: 7/7 must-haves verified
re_verification: false
---

# Phase 21: Box Fixes Verification Report

**Phase Goal:** Fix box unit rendering: remove curly brackets from labels, add dashed borders, validate C1 box contents, and color C1 boxes based on contents
**Verified:** 2026-03-24T19:30:00Z
**Status:** passed
**Re-verification:** No - initial verification

## Goal Achievement

### Observable Truths

| #   | Truth                                                       | Status       | Evidence                                                                                     |
| --- | ----------------------------------------------------------- | ------------ | -------------------------------------------------------------------------------------------- |
| 1   | Box labels display as HTML tables without curly brackets   | VERIFIED     | `buildBoxHTMLLabel` in labels.go:439-485 returns HTML table, no `{}` in output              |
| 2   | Box borders appear dashed in both collapsed and expanded views | VERIFIED | `getLevelStyle` and `getExternalStyle` in shapes.go:168-254 set `BorderStyle: "dashed"` for IsBoxType |
| 3   | Box labels show name (bold), [technology] (italic), description rows | VERIFIED | `buildBoxHTMLLabel` generates 3-row table with `<b>name</b>`, `<i>[tech]</i>`, description |
| 4   | C1 box containing both external and non-external units is rejected by validator | VERIFIED | `ValidateBoxMixedContents` in rules.go:305-342, all 6 test cases pass                       |
| 5   | C1 box with only external units has grey border color      | VERIFIED     | `GetBoxStyleByContents` in shapes.go:111-127 returns `BorderColor: model.PersonExternalBorder` (#8A8A8A) |
| 6   | C1 box with only non-external units has dark blue border color | VERIFIED | `GetBoxStyleByContents` returns `BorderColor: model.PersonBorder` (#073B6F)                 |
| 7   | Error message lists the problematic box path               | VERIFIED     | Error message: `box "%s" cannot contain both external and non-external units` with path    |

**Score:** 7/7 truths verified

### Required Artifacts

| Artifact                            | Expected                            | Status    | Details                                                          |
| ----------------------------------- | ----------------------------------- | --------- | ---------------------------------------------------------------- |
| `internal/render/labels.go`         | buildBoxHTMLLabel function          | VERIFIED  | Function exists at line 439, generates HTML table without `{}`  |
| `internal/graph/shapes.go`          | IsBoxType helper, dashed borders    | VERIFIED  | IsBoxType at line 87, dashed logic at lines 172-175, 219-222    |
| `internal/graph/shapes.go`          | HasExternalSubunits, GetBoxStyleByContents | VERIFIED | Functions exist at lines 93-105 and 111-127                     |
| `internal/validator/rules.go`       | ValidateBoxMixedContents rule       | VERIFIED  | Function exists at line 305, externalTypes map at line 294      |
| `internal/validator/validator.go`   | Rule call in Validate function      | VERIFIED  | Call exists at line 35                                           |
| `internal/graph/builder.go`         | GetBoxStyleByContents integration   | VERIFIED  | Called in buildNode (line 154), buildCluster (line 174), buildNestedCluster (line 92) |
| `internal/render/converter.go`      | IsBoxType case in switch            | VERIFIED  | Case exists at line 33-34                                        |

### Key Link Verification

| From                                | To                                  | Via                                     | Status  | Details                                          |
| ----------------------------------- | ----------------------------------- | --------------------------------------- | ------- | ------------------------------------------------ |
| `internal/render/converter.go`      | `buildBoxHTMLLabel`                 | buildHTMLLabelForType switch case       | WIRED   | Line 33-34: `case graph.IsBoxType(t):` returns `buildBoxHTMLLabel(label)` |
| `internal/graph/shapes.go`          | `getLevelStyle`                     | box type check                          | WIRED   | Lines 172-175: `if IsBoxType(t) { borderStyle = "dashed" }` |
| `internal/validator/validator.go`   | `ValidateBoxMixedContents`          | Validate function call                  | WIRED   | Line 35: `errors = append(errors, ValidateBoxMixedContents(index)...)` |
| `internal/graph/builder.go`         | `GetBoxStyleByContents`             | buildNode/buildCluster/buildNestedCluster | WIRED  | Lines 92, 154, 174: `style = GetBoxStyleByContents(entry.Unit)` |

### Requirements Coverage

This phase addresses bug fixes and has no formal requirement IDs in REQUIREMENTS.md. The ROADMAP.md defines the design decisions:

| Design Decision | Description | Status | Evidence |
| --------------- | ----------- | ------ | -------- |
| D-01 | Box labels use HTML table format (no curly brackets) | VERIFIED | buildBoxHTMLLabel generates HTML table |
| D-02 | All box types have dashed borders | VERIFIED | IsBoxType check sets borderStyle="dashed" |
| D-03 | C1 boxes cannot contain both external and non-external units | VERIFIED | ValidateBoxMixedContents rule |
| D-04 | C1 box color based on contents: grey for externals, dark blue for non-externals | VERIFIED | GetBoxStyleByContents returns appropriate colors |

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
| ---- | ---- | ------- | -------- | ------ |
| None found | - | - | - | - |

All modified files scanned for TODO, FIXME, XXX, HACK, PLACEHOLDER patterns. No blocking anti-patterns found.

### Human Verification Required

None required. All verification checks are programmatic and pass:

1. **Box label format** - Verified via code inspection: HTML table output, no `{}` characters
2. **Dashed borders** - Verified via code inspection and tests: IsBoxType check in style functions
3. **Validation rule** - Verified via 6 passing test cases covering all scenarios
4. **Content-based coloring** - Verified via code inspection: correct color constants used

### Verification Summary

**All must-haves verified:**

1. **Plan 21-01 (Box Labels and Dashed Borders):**
   - buildBoxHTMLLabel function generates HTML table without curly brackets
   - IsBoxType helper detects all 3 box variants (TypeBox, TypeContainerBox, TypeComponentBox)
   - Dashed borders applied in both getLevelStyle and getExternalStyle functions
   - Converter dispatches to buildBoxHTMLLabel for box types

2. **Plan 21-02 (Mixed Content Validation and Box Color by Contents):**
   - ValidateBoxMixedContents validation rule rejects C1 boxes with mixed external/non-external units
   - externalTypes map provides O(1) lookup for external type detection
   - HasExternalSubunits helper detects external subunits
   - GetBoxStyleByContents returns grey (#8A8A8A) for external boxes, dark blue (#073B6F) for internal boxes
   - builder.go uses GetBoxStyleByContents for TypeBox in all 3 build functions

**Test Results:**
- All validator tests pass (6/6 for ValidateBoxMixedContents)
- All graph tests pass (4/4 for box styling functions)
- All render tests pass (5/5 for box HTML labels)
- Full test suite passes (no failures)

---

_Verified: 2026-03-24T19:30:00Z_
_Verifier: Claude (gsd-verifier)_
