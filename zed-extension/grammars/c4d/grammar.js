/**
 * tree-sitter-c4d — a Tree-sitter grammar for the c4drill C4D DSL.
 *
 * The grammar mirrors the authoritative PEG spec at
 * `internal/c4d/grammar/c4d.peg` (Go/pigeon) in structure and coverage:
 *
 *   - brace-block units in all three header forms (id-led `id: type "Name"`,
 *     type-led `type [external] "Name"`, and bare-id);
 *   - the four ASCII edge arrows (`->` `<-` `<->` `--`) with the
 *     `peer: label` shorthand and trailing `{ key: value }` option blocks;
 *   - field keywords, the properties block, template declarations with
 *     `${param}` tokens, `use` instantiations, `include` directives;
 *   - double-quoted strings with escapes, `"""triple-quoted"""` strings,
 *     inline `[a, b]` lists, barewords, and `#` comments.
 *
 * Design notes — how the PEG's ordered choices map onto tree-sitter:
 *
 *   - There is deliberately NO `word` rule and the type words (person,
 *     system, queue, …) are NOT keyword tokens: models legally use them as
 *     unit ids (`queue: containerQueue "Platform Queue" { … }`), and a
 *     keyword token would win the lexer race and break the id-led header.
 *     Instead the type slot is an aliased identifier and the id-led vs
 *     type-led choice is resolved structurally by the GLR on the `:` —
 *     exactly the PEG's ordered UnitHeader alternatives. The highlighting
 *     query captures the type slot positionally.
 *   - Field/option/property keys are plain identifiers (the PEG restricts
 *     them and errors on unknown keys, D-19; here unknown keys parse so
 *     highlighting survives partially-written code — the LSP still
 *     reports them).
 *   - The true reserved words (properties/template/use/include/once/
 *     external, D-19) stay literal tokens; string literals outrank the
 *     identifier token, so e.g. `use { }` fails to parse as a unit — the
 *     same reserved-word rejection the PEG produces (as a quality error).
 *   - Statement terminators (`;` / newline) are extras, so one-line blocks
 *     and semicolon-separated statements parse without terminator rules.
 *
 * The grammar is intentionally self-contained (no external dependencies
 * beyond tree-sitter itself) so it can be extracted to a standalone
 * tree-sitter-c4d repository unchanged.
 */

module.exports = grammar({
  name: 'c4d',

  extras: $ => [
    /[\s\uFEFF\u2060\u200B]/,
    $.comment,
    ';', // statement separator — tolerated between any statements (D-18)
  ],

  // The bare-id and type-led header forms are structurally ambiguous on
  // `identifier '{'` (the type slot accepts any identifier); the GLR
  // explores both and the dynamic precedences pick the bare-id reading —
  // the PEG's ordered-choice preference for the id-led/bare forms.
  conflicts: $ => [
    [$.unit_header],
  ],

  rules: {
    document: $ => repeat($._top_level_statement),

    _top_level_statement: $ => choice(
      $.properties_block,
      $.template_declaration,
      $.use_statement,
      $.include_statement,
      $.unit_block,
    ),

    // ------------------------------------------------------------------
    // Units
    // ------------------------------------------------------------------

    // `Unit <- UnitHeader '{' BodyPart* '}'` — every statement form of a
    // unit body, including nested units. The dynamic precedence settles
    // `id: type "Name" {` against a field statement whose value would be a
    // bareword (the PEG backtracks out of FieldStmt the same way).
    unit_block: $ => prec.dynamic(2, seq(
      $.unit_header,
      '{',
      repeat($._body_statement),
      '}',
    )),

    _body_statement: $ => choice(
      $.edge_statement,
      $.use_statement,
      $.unit_block,
      $.field_statement,
    ),

    // `UnitHeader`: the id-led form requires the colon; the type-led form
    // omits the id; the bare-id form omits the type. Omitted slots stay
    // empty — inference happens in the model, never at parse time (D-02).
    //
    // The two type slots deliberately use DIFFERENT tokens:
    //   - id-led (`inner: db { }` vs field `inner: db` + stray `{`): after
    //     the colon the field-value branch can only consume a bareword, so
    //     the type slot consumes the same bareword token and the GLR picks
    //     between unit and field STRUCTURALLY (the `{` decides) — no lexer
    //     race. The PEG backtracks out of FieldStmt the same way.
    //   - type-led sits at statement start where only identifiers (and
    //     reserved literals) are valid, so its slot is a plain identifier.
    // Both alias to one `type_keyword` node name for highlighting.
    unit_header: $ => choice(
      // id: type [external] ["Display Name"]
      prec.dynamic(2, seq(
        field('id', $.identifier),
        ':',
        field('type', alias($.bareword, $.type_keyword)),
        optional(field('modifier', $.external_modifier)),
        optional(field('name', $.string)),
      )),
      // type [external] ["Display Name"]
      prec.dynamic(-1, seq(
        field('type', alias($.identifier, $.type_keyword)),
        optional(field('modifier', $.external_modifier)),
        optional(field('name', $.string)),
      )),
      // id (type inferred downstream)
      field('id', $.identifier),
    ),

    external_modifier: _ => 'external',

    // ------------------------------------------------------------------
    // Edges
    // ------------------------------------------------------------------

    // `EdgeStmt`: `-> peer [: label] [{ options }]`. The four ASCII arrows
    // of D-05; `<->` wins over `<-` by maximal munch.
    edge_statement: $ => prec.dynamic(1, seq(
      field('arrow', $.arrow),
      field('peer', $.peer_path),
      optional(seq(':', field('label', $._edge_label))),
      optional($.options_block),
    )),

    arrow: _ => choice('<->', '->', '<-', '--'),

    // Dotted peer path (D-07): bare walk-up names or absolute dotted paths;
    // each segment may be a `${param}` token inside template bodies.
    peer_path: $ => seq(
      $.peer_segment,
      repeat(seq('.', $.peer_segment)),
    ),

    peer_segment: $ => choice($.identifier, $.param_token),

    param_token: $ => seq('${', $.identifier, '}'),

    _edge_label: $ => choice($.triple_string, $.string, $.edge_label_text),

    // D-09 pipe shorthand (`tech | description`) rides one token; the
    // split into technology/description happens server-side. Stops at
    // `{` so trailing option blocks still parse, and the first character
    // can be neither whitespace (extras own it) nor ':' (the label
    // separator — a longer-token match would otherwise swallow it).
    edge_label_text: _ => token(seq(
      /[^{};"#:\s\r\n]/,
      repeat(/[^{};"#\r\n]/),
    )),

    options_block: $ => seq('{', repeat($.edge_option), '}'),

    edge_option: $ => seq(
      field('key', $.identifier),
      ':',
      field('value', $._option_value),
    ),

    _option_value: $ => choice($.string, $.option_bareword),

    option_bareword: _ => token(/[^{};|"#\s]+/),

    // ------------------------------------------------------------------
    // Fields
    // ------------------------------------------------------------------

    // `key: value`. The PEG restricts keys to the FieldKey set and errors
    // on unknown keys (D-19); here unknown keys parse as ordinary fields
    // so highlighting survives partially-written code.
    field_statement: $ => seq(
      field('key', $.identifier),
      ':',
      field('value', $._value),
    ),

    _value: $ => choice(
      $.triple_string,
      $.string,
      $.list,
      $.bareword,
    ),

    // D-15: inline `[a, b]` and one-per-line forms; commas optional.
    list: $ => seq(
      '[',
      optional(seq(
        $._list_item,
        repeat(seq(optional(','), $._list_item)),
      )),
      optional(','),
      ']',
    ),

    _list_item: $ => choice($.triple_string, $.string, $.list_bareword),

    list_bareword: _ => token(/[^{};|"#\[\]\s]+/),

    // Barewords stop at structural characters exactly like the PEG's
    // Bareword: `{};|"#` and (deliberately, so inline lists win) `[`/`]`.
    // Inner spaces are allowed and the token is line-scoped, so
    // `description: Online shopping platform` is ONE value, matching the
    // PEG (which trims). The first character cannot be whitespace.
    // `${param}` tokens ride along via the second alternative (D-13).
    bareword: _ => token(choice(
      seq(
        /[^{};|"#\[\]\s\r\n]/,
        repeat(/[^{};|"#\[\]\r\n]/),
      ),
      seq(
        repeat(/[^\s{};|"#\[\]\r\n]/),
        /\$\{[A-Za-z0-9_-]+\}/,
        repeat(/[^\s{};|"#\[\]\r\n]/),
      ),
    )),

    // ------------------------------------------------------------------
    // Properties
    // ------------------------------------------------------------------

    // D-12 top-level block. The dynamic precedence settles `properties {}`
    // against a bare-id unit named `properties` — the PEG reserves the
    // word, so the block always wins.
    properties_block: $ => prec.dynamic(3, seq(
      'properties',
      '{',
      repeat($.property_field),
      '}',
    )),

    property_field: $ => seq(
      field('key', $.identifier),
      ':',
      field('value', $._value),
    ),

    // ------------------------------------------------------------------
    // Templates
    // ------------------------------------------------------------------

    template_declaration: $ => seq(
      'template',
      field('name', $.identifier),
      '(',
      optional($.parameter_list),
      ')',
      '{',
      repeat($._template_body_statement),
      '}',
    ),

    parameter_list: $ => seq(
      $.identifier,
      repeat(seq(',', $.identifier)),
    ),

    _template_body_statement: $ => choice(
      $.type_statement,
      $.edge_statement,
      $.use_statement,
      $.unit_block,
      $.field_statement,
    ),

    // The template root's type: `type: <type> [external]` — the only
    // place a type statement is legal (D-22). Inside template bodies the
    // literal 'type' outranks the field-key identifier, so this form wins
    // over an equal field statement — the PEG's TemplateBodyStmt order.
    // The type slot consumes the same bareword token the field branch
    // would, so the choice stays structural (the dynamic precedence
    // settles it).
    type_statement: $ => prec.dynamic(1, seq(
      'type',
      ':',
      field('type', alias($.bareword, $.type_keyword)),
      optional(field('modifier', $.external_modifier)),
    )),

    // ------------------------------------------------------------------
    // Use / include
    // ------------------------------------------------------------------

    use_statement: $ => seq(
      'use',
      field('template', $.identifier),
      '(',
      optional($.argument_list),
      ')',
    ),

    argument_list: $ => seq(
      $.argument,
      repeat(seq(',', $.argument)),
    ),

    // Named `key: value` args first-class; positional values pair with
    // declared params later. A positional value containing ':' must be
    // quoted — the named form wins on a bare `key: value` shape.
    argument: $ => choice(
      prec.dynamic(1, seq(
        field('name', $.identifier),
        ':',
        field('value', $._argument_value),
      )),
      field('value', $._argument_value),
    ),

    _argument_value: $ => choice(
      $.triple_string,
      $.string,
      $.list,
      $.param_token,
      $.identifier,
      $.argument_bareword,
    ),

    // Stops at '(', ')', ',', whitespace and ':' — a positional value
    // containing ':' must be quoted (the named `key: value` form owns the
    // bare colon), matching the PEG's ArgBarewordStop semantics. The
    // negative lexical precedence keeps `identifier` winning the lexer
    // race where both match, so the ':' decides named vs positional.
    argument_bareword: _ => token(prec(-1, /[^{};|"#():,\s]+/)),

    include_statement: $ => seq(
      'include',
      field('path', choice($.string, $.include_path)),
      optional($.once_modifier),
    ),

    include_path: _ => token(/[A-Za-z0-9_./-]+/),

    once_modifier: _ => 'once',

    // ------------------------------------------------------------------
    // Trivia
    // ------------------------------------------------------------------

    // `#` to end of line; the newline is not consumed.
    comment: _ => token(seq('#', /[^\r\n]*/)),

    // [A-Za-z0-9_-]+ — unit ids, keys, bare path segments (PEG Ident).
    identifier: _ => /[\w-]+/,

    // Double-quoted with escapes: \" \\ \n \t \r and the pass-through
    // escape of the PEG's unescapeDoubleQuoted.
    string: _ => token(seq(
      '"',
      repeat(choice(/[^"\\\r\n]/, /\\[\s\S]/)),
      '"',
    )),

    // `"""..."""` — ends at the first triple quote (lazy match).
    triple_string: _ => token(seq('"""', /[\s\S]*?/, '"""')),
  },
});
