// c4d.ts — the CodeMirror 6 language package for the C4D DSL (issue #36):
// the Lezer grammar in c4d.grammar (generated parser: c4d_parser.ts) wrapped
// with syntax highlighting, code folding and indentation. Replaces the P0
// StreamLanguage fallback.

import { parser } from "./c4d_parser";
import {
  LRLanguage,
  LanguageSupport,
  foldNodeProp,
  foldInside,
  indentNodeProp,
  delimitedIndent,
} from "@codemirror/language";
import { styleTags, tags as t } from "@lezer/highlight";

// Highlighting maps node positions to tags. Path selectors scope tokens to
// their parent: the bare Ident spines shared by unit headers and field
// statements (see c4d.grammar's design notes) stay unstyled, while the
// unambiguous positions get their roles. The reserved statement keywords
// are specialized Ident content, so their node names are the bare words.
const c4dParser = parser.configure({
  props: [
    styleTags({
      // Field keys (the only direct Ident child of FieldStmt).
      "FieldStmt/Ident": t.propertyName,
      // Edge-option keys.
      "OptionName!": t.propertyName,
      // Statement-start type word (`system "Payment" { … }`) and the
      // template root type statement (`type: container`).
      "TypeWord!": t.typeName,
      "TypeStmt/Ident": t.typeName,
      // Dotted peer paths after an arrow.
      "PeerPath!": t.attributeName,
      // Template names in declarations and instantiations.
      "TemplateName!": t.function(t.variableName),
      // include paths.
      "IncludePath!": t.link,
      // Template parameter tokens, quoted strings, arrows and label pipes.
      ParamToken: t.variableName,
      String: t.string,
      TripleString: t.string,
      Arrow: t.operator,
      Pipe: t.operator,
      // Reserved statement keywords (D-19) and modifiers.
      "use include template properties type": t.keyword,
      "once external": t.modifier,
      // Skipped trivia still lands in the tree, comments included.
      LineComment: t.lineComment,
    }),
    // Brace blocks fold: units (incl. nested one-liners), templates, the
    // properties block and inline edge-option blocks.
    foldNodeProp.add({
      UnitBlock: foldInside,
      TemplateDecl: foldInside,
      PropertiesBlock: foldInside,
      OptionsBlock: foldInside,
    }),
    // Indentation follows the brace structure.
    indentNodeProp.add({
      "UnitBlock TemplateDecl PropertiesBlock OptionsBlock": delimitedIndent({ closing: "}" }),
    }),
  ],
});

const c4dLanguage = LRLanguage.define({
  name: "c4d",
  parser: c4dParser,
  languageData: {
    commentTokens: { line: "#" },
  },
});

/** c4d returns the LanguageSupport for .c4d model files. */
export function c4d(): LanguageSupport {
  return new LanguageSupport(c4dLanguage);
}

export { c4dLanguage, c4dParser };
