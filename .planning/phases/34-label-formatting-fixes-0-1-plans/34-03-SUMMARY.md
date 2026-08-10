# Phase 34 Plan 03: punctuation-aware tokenization Summary

**wrapText now breaks lines at word boundaries AND punctuation/symbol boundaries: `Multi-Consumer`, `YUV420->EXTERNAL:...`, `IMAGE_NATIVE_PROCESSED` and `lock-state` all wrap at their separators; pure letter/digit runs still stay unsplit (D-05); in-budget output byte-identical except where a punctuation break now triggers across a line boundary**

## Performance

- **Duration:** ~25 min
- **Started:** 2026-08-10T17:30:00Z
- **Completed:** 2026-08-10T17:55:00Z
- **Tasks:** 3 (TDD: RED/GREEN/REFACTOR-verify)
- **Files modified:** 3 (+1 golden re-baselined)

## Root Cause Closed (UAT tests 1 & 3, severity major)

`wrapText` tokenized with `strings.Fields` — whitespace-only. Space-free punctuation tokens (`IMAGE_NATIVE_PROCESSED_CUDA_YUV420->EXTERNAL:...:dwImageHandle_t`, `(open/close)`, `lock-state`) received no break opportunities and overflowed unsplit. User requirement (verbatim): "any punctuation must be considered word boundary, not just spaces".

## Implementation

- **`tokenizeWrapText`** (internal/render/wrap.go): splits into `wrapToken{text, spaceBefore}` units. Classification is three-way — wordChar (`unicode.IsLetter`/`IsDigit`), whitespace, separator (everything else). Deliberately NOT `unicode.IsPunct`: `>` in `->` is Unicode category Sm (symbol), which `IsPunct` misses; "not wordChar and not space" covers punctuation AND symbols.
- **Attachment semantics:** a separator run attaches to the preceding word ("Multi-", "YUV420->", "IMAGE_"); a leading separator run attaches to the following word ("[CGF"); a lone separator string emits as its own token.
- **Rejoin without space:** `spaceBefore` tracks whitespace separation only — punctuation-separated tokens rejoin with NO space on the same line ("foo-bar" stays "foo-bar", byte-identical to v1.10).
- **Line-fitting:** existing `fitsOnLine`/`flushIfPending` machinery reused; the over-budget branch (D-05) unchanged — a token with no internal break opportunity (pure letter/digit run) still emits unsplit on its own line.

## Task Commits (TDD)

1. **RED** `02b4c50` — `test(34-03): add punctuation-break cases to TestWrapText` (6 new subtests; 3 failed for the right reason: hyphen, arrow+colon, underscore)
2. **GREEN** `c15ebd2` — `feat(34-03): wrap at punctuation boundaries too (tokenizer)` (all 15 subtests pass)
3. **REFACTOR** `acacb45` — `test(34-03): re-baseline expanded DOT golden for punctuation-break semantics (COMPAT-02)`

## COMPAT-01 / Golden Impact

`TestBuildExpandedGraphBaselineDOT` initially failed: the multilevel fixture's unit labels contain `(open/close)` and `lock-state`, which now break at `/` and `-`:

- `API for managing SSH<BR/>sessions<BR/>(open/close)` → `...sessions (open/<BR/>close)` — `/` break (intended)
- `API for user lockout<BR/>checks and<BR/>lock-state updates` → `...checks and lock-<BR/>state updates` — `-` break (intended)

Both are the exact punctuation-break semantics requested; all other changes in the golden are layout geometry (bb/pos/lp), which canonicalDOT strips. Re-baselined via the documented regeneration procedure (`go run ./cmd/c4drill cmd/c4drill/testdata/multilevel.toml --format dot --expanded`), never hand-edited. Full suite: 12/12 packages green (DI-1 canonical comparisons).

## Decisions

- Separator class = "not wordChar, not whitespace" (covers symbols like `>`, `/`; deliberately not `unicode.IsPunct`)
- Trailing attachment for separator runs; leading attachment at string start
- `spaceBefore` only from whitespace — punctuation rejoin inserts no space
- D-05 preserved: pure letter/digit runs never char-split

## Requirements

- LABEL-02 (re-asserted): break opportunities = whitespace ∪ punctuation/symbols ✓
- COMPAT-01: in-budget output byte-identical; goldens green after documented re-baseline ✓
