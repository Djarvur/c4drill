; Outline: units (all header forms), templates and the properties block.
; The display name prefers the id; type-led headers fall back to the type.
(unit_block) @item

(unit_header
  id: (identifier) @name)

; type-led headers (no id): the type word names the item
(unit_header
  type: (type_keyword) @name) @item

(template_declaration
  name: (identifier) @name) @item

(properties_block) @item
