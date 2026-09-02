// editor.ts sets up the CodeMirror 6 editor area: TOML highlighting via the
// legacy mode and C4D via the dedicated Lezer grammar (issue #36 — folding
// and indentation included), LSP-driven completion, hover, go-to-definition
// and diagnostics squiggles over the backend dispatch.

import { EditorView, keymap, lineNumbers, highlightActiveLine, highlightActiveLineGutter, drawSelection, placeholder, hoverTooltip } from "@codemirror/view";
import { EditorState, Compartment, type Extension } from "@codemirror/state";
import { defaultKeymap, history, historyKeymap, indentWithTab } from "@codemirror/commands";
import { StreamLanguage, syntaxHighlighting, defaultHighlightStyle, foldGutter, foldKeymap } from "@codemirror/language";
import { toml as tomlMode } from "@codemirror/legacy-modes/mode/toml";
import { autocompletion, closeBrackets, closeBracketsKeymap, completionKeymap, type CompletionContext, type CompletionResult } from "@codemirror/autocomplete";
import { linter, lintGutter, setDiagnostics, type Diagnostic as LintDiagnostic } from "@codemirror/lint";
import { searchKeymap, highlightSelectionMatches } from "@codemirror/search";
import type { Tooltip } from "@codemirror/view";
import type { Diag } from "./types";
import { call } from "./rpc";
import { c4d } from "./language/c4d";

export function languageFor(path: string): Extension {
  return path.toLowerCase().endsWith(".c4d") ? c4d() : StreamLanguage.define(tomlMode);
}

export interface EditorHooks {
  onChange(path: string, text: string, version: number): void;
  onSave(path: string, text: string): void;
}

export class EditorArea {
  readonly view: EditorView;

  private langComp = new Compartment();

  private version = 0;

  private path = "";

  private hooks: EditorHooks;

  constructor(parent: HTMLElement, hooks: EditorHooks) {
    this.hooks = hooks;

    this.view = new EditorView({
      parent,
      state: EditorState.create({ doc: "", extensions: this.extensions() }),
    });
  }

  private extensions() {
    const self = this;

    return [
      lineNumbers(),
      highlightActiveLineGutter(),
      foldGutter(),
      highlightActiveLine(),
      drawSelection(),
      history(),
      closeBrackets(),
      autocompletion({ override: [self.completionSource.bind(self)] }),
      highlightSelectionMatches(),
      lintGutter(),
      linter(() => [], { delay: 0 }), // diagnostics are pushed via setDiagnostics
      hoverTooltip((view, pos) => self.hoverRenderer(view, pos), { hoverTime: 300 }),
      this.langComp.of(StreamLanguage.define(tomlMode)),
      syntaxHighlighting(defaultHighlightStyle, { fallback: true }),
      keymap.of([
        ...closeBracketsKeymap,
        ...defaultKeymap,
        ...historyKeymap,
        ...foldKeymap,
        ...searchKeymap,
        ...completionKeymap,
        indentWithTab,
        {
          key: "Mod-s",
          preventDefault: true,
          run: () => {
            self.hooks.onSave(self.path, self.view.state.doc.toString());
            return true;
          },
        },
      ]),
      EditorView.updateListener.of((update) => {
        if (!update.docChanged) return;
        self.version += 1;
        self.hooks.onChange(self.path, update.state.doc.toString(), self.version);
      }),
      placeholder("Open a .toml or .c4d file to start editing"),
    ];
  }

  /** completionSource proxies textDocument/completion through the backend. */
  private async completionSource(ctx: CompletionContext): Promise<CompletionResult | null> {
    if (!this.path) return null;

    const line = ctx.state.doc.lineAt(ctx.pos);
    const character = ctx.pos - line.from;

    try {
      const list = await call<{ items: any[] } | null>("completion", {
        path: this.path,
        line: line.number - 1,
        character,
      });

      if (!list?.items?.length) return null;

      // Replace the current word token (identifiers, dotted paths, keys).
      const word = ctx.matchBefore(/[\w."-]+/);
      const from = word ? word.from : ctx.pos;

      return {
        from,
        options: list.items.map((item) => ({
          label: item.label,
          detail: item.detail,
          type: kindToType(item.kind),
          info: item.documentation,
          apply: item.insertText || item.label,
        })),
        validFor: /^[\w."-]*$/,
      };
    } catch {
      return null;
    }
  }

  /** hoverRenderer proxies textDocument/hover for the tooltip. */
  private async hoverRenderer(view: EditorView, pos: number): Promise<Tooltip | null> {
    if (!this.path) return null;

    const line = view.state.doc.lineAt(pos);
    const character = pos - line.from;

    let res;
    try {
      res = await call<{ hover: { contents: { value: string } } | null }>("hover", {
        path: this.path,
        line: line.number - 1,
        character,
      });
    } catch {
      return null;
    }

    const text = res?.hover?.contents?.value;
    if (!text) return null;

    const el = document.createElement("div");
    el.className = "cm-hover";
    el.textContent = text;
    return { pos, above: true, create: () => ({ dom: el }) };
  }

  /** goDefinition jumps to the definition of the symbol at the cursor. */
  async goDefinition(): Promise<void> {
    const pos = this.view.state.selection.main.head;
    const line = this.view.state.doc.lineAt(pos);
    const character = pos - line.from;

    const res = await call<{ path: string } | null>("definition", {
      path: this.path,
      line: line.number - 1,
      character,
    });

    if (res?.path) {
      // The app wires onOpenRequest in main.ts.
      this.hooks.onChange(this.path, this.view.state.doc.toString(), ++this.version);
      window.dispatchEvent(new CustomEvent("c4drill:open", { detail: { path: res.path } }));
    }
  }

  /** openDoc swaps the editor to a project file. */
  openDoc(path: string, text: string): void {
    this.path = path;
    this.version = 1;
    this.view.dispatch({
      changes: { from: 0, to: this.view.state.doc.length, insert: text },
      effects: this.langComp.reconfigure(languageFor(path)),
    });
  }

  currentText(): string {
    return this.view.state.doc.toString();
  }

  currentPath(): string {
    return this.path;
  }

  /** applyDiagnostics shows the LSP diagnostics for the current document. */
  applyDiagnostics(diags: Diag[]): void {
    const doc = this.view.state.doc;

    const mapped: LintDiagnostic[] = (diags ?? []).map((d) => {
      const from = posToOffset(doc, d.range?.start?.line ?? 0, d.range?.start?.character ?? 0);
      const to = posToOffset(doc, d.range?.end?.line ?? 0, d.range?.end?.character ?? 0);

      return {
        from: Math.min(from, doc.length),
        to: Math.min(Math.max(to, from), doc.length),
        message: d.message,
        severity: d.severity === 2 ? "warning" : d.severity === 3 ? "info" : "error",
        source: d.source,
      };
    });

    // setDiagnostics yields a transaction spec (the LSP-style push path).
    this.view.dispatch(setDiagnostics(this.view.state, mapped));
  }

  clearDiagnostics(): void {
    this.applyDiagnostics([]);
  }
}

function posToOffset(doc: { lines: number; line(n: number): { from: number; to: number } }, line0: number, char: number): number {
  const lineNo = Math.min(Math.max(line0 + 1, 1), doc.lines);
  const line = doc.line(lineNo);
  return Math.min(line.from + Math.max(char, 0), line.to);
}

function kindToType(kind?: number): string {
  // LSP CompletionItemKind → CodeMirror type class (rough map).
  switch (kind) {
    case 2: return "namespace";     // Module
    case 5: return "property";      // Field
    case 7: return "class";         // Class
    case 12: return "constant";     // Value
    case 14: return "keyword";      // Keyword
    case 17: return "file";         // File
    case 19: return "namespace";    // Folder
    case 20: return "enum";         // EnumMember
    default: return "text";
  }
}
