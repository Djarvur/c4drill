; tree-sitter-c4d highlighting for Zed.
;
; The grammar keeps type words as ordinary identifiers (models legally use
; them as unit ids), so types are captured positionally via the type slots.

; ---- comments ----

(comment) @comment
(comment) @spell

; ---- strings ----

(string) @string
(triple_string) @string

(string) @spell
(triple_string) @spell

; ---- C4 type slots (unit headers, template root type statement) ----

(unit_header
  type: (type_keyword) @type)

(type_statement
  type: (type_keyword) @type)

; the type-led header form has no id, so its type word doubles as the
; unit's identity — keep the type color but let the unit show in outline

; ---- modifiers ----

(external_modifier) @keyword.modifier
(once_modifier) @keyword.modifier

; ---- statement keywords ----

(properties_block "properties" @keyword)
(template_declaration "template" @keyword)
(type_statement "type" @keyword)
(use_statement "use" @keyword)
(include_statement "include" @keyword)

; ---- edge arrows ----

(arrow) @operator

; ---- keys ----

(field_statement
  key: (identifier) @property)

(property_field
  key: (identifier) @property)

(edge_option
  key: (identifier) @property)

; ---- unit ids and peer references ----

(unit_header
  id: (identifier) @variable)

(peer_path
  (peer_segment) @variable)

; ---- template names and ${param} tokens ----

(template_declaration
  name: (identifier) @function)

(use_statement
  template: (identifier) @function)

(parameter_list
  (identifier) @variable.parameter)

(argument
  name: (identifier) @variable.parameter)

(param_token
  (identifier) @variable)
(param_token
  ["${" "}"] @punctuation.special)

; ---- values ----

(bareword) @string.special
(list_bareword) @string.special
(argument_bareword) @string.special
(edge_label_text) @string.special
(include_path) @string.special

; enum-ish option values (arrow: forward, style: dashed, rank: reverse, …)
(option_bareword) @constant

; ---- punctuation ----

["{" "}" "[" "]" "(" ")"] @punctuation.bracket
[":" "," "."] @punctuation.delimiter
