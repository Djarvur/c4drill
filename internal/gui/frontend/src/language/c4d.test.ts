// c4d.test.ts — token-level tests for the C4D Lezer grammar (issue #36):
// the hard cases from c4d.peg (header forms, external, arrows, option
// blocks, `;` one-liners, triple-quoted strings, ${param}, comments) plus
// whole-file parses of the .c4d example fixtures, highlighting spans, and
// the folding/indentation node props wired into the editor.

import { describe, expect, it } from "vitest";
import type { SyntaxNode } from "@lezer/common";
import { highlightTree, tagHighlighter, tags as t } from "@lezer/highlight";
import { foldNodeProp, indentNodeProp } from "@codemirror/language";
import { c4dParser } from "./c4d";
import fixture from "./testdata/01-units.c4d?raw";
import nested from "./testdata/02-nested.c4d?raw";
import links from "./testdata/03-links.c4d?raw";
import styling from "./testdata/04-styling.c4d?raw";
import templates from "./testdata/06-templates.c4d?raw";
import relativePeer from "./testdata/07-relative-peer.c4d?raw";
import edgeKinds from "./testdata/10-edge-kinds.c4d?raw";
import nesting from "./testdata/11-nesting-context.c4d?raw";

/** parse runs the configured parser (highlight/fold/indent props attached). */
function parse(text: string): SyntaxNode {
  return c4dParser.parse(text).topNode;
}

function countNodes(node: SyntaxNode, name: string): number {
  let count = 0;
  const cursor = node.cursor();

  do {
    if (cursor.name === name) count++;
  } while (cursor.next());

  return count;
}

function childNames(node: SyntaxNode): string[] {
  const names: string[] = [];

  for (let child = node.firstChild; child; child = child.nextSibling) {
    names.push(child.name);
  }

  return names;
}

/** expectNoErrors parses and fails with the offending source on ERROR nodes. */
function expectNoErrors(text: string): SyntaxNode {
  const tree = parse(text);
  const spots: string[] = [];
  const cursor = tree.cursor();

  do {
    if (cursor.name === "⚠") {
      spots.push(`@${cursor.from}: ${JSON.stringify(text.slice(cursor.from, Math.min(cursor.to, cursor.from + 60)))}`);
    }
  } while (cursor.next());

  expect(spots, `expected no parse errors in ${JSON.stringify(text.slice(0, 80))}`).toEqual([]);

  return tree;
}

describe("c4d grammar — unit header forms", () => {
  it("parses the id-led header with type, external and display name", () => {
    const tree = expectNoErrors(`payment: system external "Payment Service" {\n  description: External\n}`);

    expect(countNodes(tree, "UnitBlock")).toBe(1);

    const statement = tree.firstChild!;
    expect(statement.name).toBe("TopStatement");

    const unit = statement.firstChild!;
    expect(unit.name).toBe("UnitBlock");

    const header = unit.firstChild!.firstChild!;
    expect(header.name).toBe("IdHeader");

    // The header carries the id, the type word, the modifier and the name.
    expect(childNames(header)).toEqual(["Ident", "Ident", "external", "String"]);
  });

  it("parses the type-led header", () => {
    const tree = expectNoErrors(`system external "Analytics" {\n  technology: Datadog\n}`);

    const unit = tree.firstChild!.firstChild!;
    expect(unit.name).toBe("UnitBlock");

    const header = unit.firstChild!.firstChild!;
    expect(header.name).toBe("TypeHeader");
    expect(childNames(header)).toEqual(["TypeWord", "external", "String"]);
  });

  it("parses the bare-id one-liner form", () => {
    const tree = expectNoErrors(`platform { }`);
    expect(countNodes(tree, "UnitBlock")).toBe(1);
  });

  it("keeps field statements fields when a value has several words", () => {
    const tree = expectNoErrors(`box { description: Used by external systems\n  technology: React }`);

    expect(countNodes(tree, "FieldStmt")).toBe(2);
    expect(countNodes(tree, "UnitBlock")).toBe(1);
  });

  it("treats type words as legal unit ids (id-led headers keep working)", () => {
    const tree = expectNoErrors(`queue: containerQueue "Platform Queue" { }`);
    expect(countNodes(tree, "UnitBlock")).toBe(1);
    expect(countNodes(tree, "FieldStmt")).toBe(0);
  });
});

describe("c4d grammar — edges, one-liners, values", () => {
  it("parses all four arrow forms with labels and option blocks", () => {
    const text = [
      `webapp: system "Web" {`,
      `  -> api: "REST | Queries" { color: blue }`,
      `  <- api: Callbacks`,
      `  <-> cache: "Redis | Caches" { style: dashed }`,
      `  -- analytics: Sends events`,
      `}`,
    ].join("\n");

    const tree = expectNoErrors(text);

    expect(countNodes(tree, "EdgeStmt")).toBe(4);
    expect(countNodes(tree, "Arrow")).toBe(4);
    expect(countNodes(tree, "OptionsBlock")).toBe(2);
    expect(countNodes(tree, "Option")).toBe(2);
  });

  it("parses ; separated one-liner nested blocks", () => {
    const tree = expectNoErrors(`box { api: system { } db: db { } }`);

    expect(countNodes(tree, "UnitBlock")).toBe(3);
  });

  it("parses dotted peer paths and ${param} peers", () => {
    const text = `api: container "API" {\n  -> shop.api.handlers: Routes\n  -> ${"${upstreamBus}"}: Publishes\n}`;
    const tree = expectNoErrors(text);

    expect(countNodes(tree, "PeerPath")).toBe(2);
    expect(countNodes(tree, "ParamToken")).toBe(1);
  });

  it("parses multi-line triple-quoted strings", () => {
    const tree = expectNoErrors(`box { description: """a longer,\nmulti-line value""" }`);
    expect(countNodes(tree, "TripleString")).toBe(1);
  });

  it("parses lists with optional commas and quoted items", () => {
    const tree = expectNoErrors(`box { expanded: [platform, gateway]\n  legendLine: ["Nightly batch|#E65100|dashed"] }`);

    expect(countNodes(tree, "List")).toBe(2);
  });

  it("parses comments anywhere, including trailing", () => {
    const tree = expectNoErrors(`# leading comment\nplatform: system "P" { # trailing\n  technology: Go # another\n}`);
    expect(countNodes(tree, "LineComment")).toBe(3);
  });

  it("parses the composition statements", () => {
    const text = [
      `include shared/auth.c4d once`,
      `template microservice(name, tech) {`,
      `  type: container`,
      `  name: "${"${name}"} Service"`,
      `  -> ${"${upstreamBus}"}: "Publishes events"`,
      `}`,
      `use microservice(name: auth, tech: "Go")`,
    ].join("\n");

    const tree = expectNoErrors(text);

    expect(countNodes(tree, "IncludeStmt")).toBe(1);
    expect(countNodes(tree, "once")).toBe(1);
    expect(countNodes(tree, "TemplateDecl")).toBe(1);
    expect(countNodes(tree, "TypeStmt")).toBe(1);
    expect(countNodes(tree, "ParamToken")).toBe(1);
    expect(countNodes(tree, "UseStmt")).toBe(1);
    expect(countNodes(tree, "Arg")).toBe(2);
  });

  it("parses the properties block", () => {
    const tree = expectNoErrors(`properties {\n  name: Demo\n  edges: spline\n}`);

    expect(countNodes(tree, "PropertiesBlock")).toBe(1);
    expect(countNodes(tree, "FieldStmt")).toBe(2);
  });
});

describe("c4d grammar — example fixtures", () => {
  const fixtures: Record<string, string> = {
    "01-units.c4d": fixture,
    "02-nested.c4d": nested,
    "03-links.c4d": links,
    "04-styling.c4d": styling,
    "06-templates.c4d": templates,
    "07-relative-peer.c4d": relativePeer,
    "10-edge-kinds.c4d": edgeKinds,
    "11-nesting-context.c4d": nesting,
  };

  // Per-file expectations: minimum node inventory that must show up.
  const expectations: Record<string, Partial<Record<string, number>>> = {
    "01-units.c4d": { UnitBlock: 6, IdHeader: 4, external: 1 },
    "02-nested.c4d": { UnitBlock: 7, EdgeStmt: 5, PropertiesBlock: 1 },
    "03-links.c4d": { EdgeStmt: 12, OptionsBlock: 8, external: 2 },
    "04-styling.c4d": { UnitBlock: 6, EdgeStmt: 5, OptionsBlock: 5 },
    "06-templates.c4d": { TemplateDecl: 2, UseStmt: 3, TypeStmt: 2, ParamToken: 1 },
    "07-relative-peer.c4d": { UnitBlock: 7, EdgeStmt: 6 },
    "10-edge-kinds.c4d": { EdgeStmt: 6 },
    "11-nesting-context.c4d": { UnitBlock: 7, List: 1 },
  };

  for (const [name, text] of Object.entries(fixtures)) {
    it(`parses ${name} with no error nodes`, () => {
      const tree = expectNoErrors(text);

      const expected = expectations[name] ?? {};
      for (const [node, min] of Object.entries(expected)) {
        expect(countNodes(tree, node), `${name}: ${node} nodes`).toBeGreaterThanOrEqual(min ?? 1);
      }
    });
  }
});

describe("c4d highlighting", () => {
  interface Highlight {
    from: number;
    to: number;
    cls: string;
  }

  function highlight(text: string): Highlight[] {
    const specs = [
      { tag: t.comment, class: "x-comment" },
      { tag: t.string, class: "x-string" },
      { tag: t.keyword, class: "x-keyword" },
      { tag: t.modifier, class: "x-modifier" },
      { tag: t.operator, class: "x-operator" },
      { tag: t.typeName, class: "x-type" },
      { tag: t.propertyName, class: "x-prop" },
      { tag: t.attributeName, class: "x-peer" },
      { tag: t.variableName, class: "x-var" },
      { tag: t.link, class: "x-link" },
    ];

    const spans: Highlight[] = [];
    const tree = c4dParser.parse(text);

    highlightTree(tree, tagHighlighter(specs), (from, to, cls) => spans.push({ from, to, cls }));

    return spans;
  }

  function classAt(spans: Highlight[], pos: number): string | null {
    for (const s of spans) if (pos >= s.from && pos < s.to) return s.cls;

    return null;
  }

  it("styles comments, strings, keywords, arrows, peers and types", () => {
    const text = [
      `# a comment`,
      `include shared/auth.c4d once`,
      `use microservice(name: auth)`,
      `db "Cache" { }`,
      `payment: system external "Payment Service" {`,
      `  -> api: "REST | Queries" { color: blue }`,
      `}`,
    ].join("\n");

    const spans = highlight(text);
    const posOf = (needle: string) => {
      const idx = text.indexOf(needle);
      expect(idx, needle).toBeGreaterThanOrEqual(0);

      return idx;
    };

    expect(classAt(spans, posOf("# a comment"))).toBe("x-comment");
    expect(classAt(spans, posOf("once"))).toBe("x-modifier");
    expect(classAt(spans, posOf("use "))).toBe("x-keyword");
    expect(classAt(spans, posOf("Payment Service"))).toBe("x-string");
    expect(classAt(spans, posOf("->"))).toBe("x-operator");
    expect(classAt(spans, posOf("shared/auth.c4d"))).toBe("x-link");

    // The type-led statement-start word is styled as a type…
    expect(classAt(spans, posOf("db \"Cache\""))).toBe("x-type");

    // …while the id-led header spine (id + type words) stays plain text —
    // the documented LR trade-off for the shared header spine.

    // …the template name as a function…
    expect(classAt(spans, posOf("microservice"))).toBe("x-var");

    // …and the option key as a property.
    expect(classAt(spans, posOf("color"))).toBe("x-prop");
  });
});

describe("c4d folding and indentation props", () => {
  it("attaches fold ranges to unit blocks", () => {
    const tree = parse(`platform: system "P" {\n  technology: Go\n}`);
    const unit = tree.firstChild!.firstChild;
    expect(unit).not.toBeNull();
    expect(unit!.name).toBe("UnitBlock");
    expect(unit!.type.prop(foldNodeProp)).toBeDefined();
  });

  it("attaches indentation to brace blocks", () => {
    const tree = parse(`properties {\n  name: Demo\n}`);
    const props = tree.firstChild!.firstChild;
    expect(props).not.toBeNull();
    expect(props!.name).toBe("PropertiesBlock");
    expect(props!.type.prop(indentNodeProp)).toBeDefined();
  });
});
