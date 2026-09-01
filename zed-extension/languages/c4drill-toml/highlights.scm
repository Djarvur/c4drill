; Minimal highlighting for the scoped C4Drill TOML language.
;
; The language reuses Zed's built-in tree-sitter-toml grammar (grammar
; "toml"), but Zed attaches highlight queries per LANGUAGE, so the stock
; TOML extension's queries do not apply to this one. This file mirrors
; their coverage using the actual node names of tree-sitter/tree-sitter-toml
; (pair/bare_key/quoted_key/table/table_array_element/…; note there is no
; named `key` node and no field names on `pair`).

[
  (string)
  (local_date)
  (local_date_time)
  (local_time)
  (offset_date_time)
] @string

(escape_sequence) @string.escape

(comment) @comment
(comment) @spell

(integer) @number
(float) @number
(boolean) @constant

; keys only ever appear in key positions (pair/table/table_array_element)
(bare_key) @property
(quoted_key) @property

["=" "." ","] @punctuation.delimiter
["{" "}" "[" "]"] @punctuation.bracket
