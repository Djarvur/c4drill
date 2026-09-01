// The live preview panel (issue #27 scope update): a webview that inlines the
// SVG returned by c4drill/renderDiagram for the current document and
// auto-refreshes on edit (debounced, default 200 ms). Click-through
// drill-down navigation intercepts the SVG's internal relative-.svg links
// (resolved per core/renderTarget.ts, mirroring the CLI output tree), external
// http(s) reference links open in the system browser, parse/validate failures
// show the CLI-identical messages instead of a stale diagram, and the toolbar
// exposes the view controls (breadcrumbs, expanded view, collapse-all,
// expanded-set override, legend) plus CLI-convention exports.

import * as cp from 'child_process';
import * as path from 'path';
import * as vscode from 'vscode';
import { LanguageClient } from 'vscode-languageclient/node';

import { debounce } from '../core/debounce';
import {
    buildRenderParams,
    initialViewState,
    PreviewViewState,
} from '../core/renderParams';
import { basenameOfUri, resolveRenderTarget } from '../core/renderTarget';
import { DocumentManager, resolveServerCommand } from '../client';
import {
    DiagnosticItem,
    HostToWebviewMessage,
    WebviewToHostMessage,
} from './messages';
import { RenderDiagramRequest } from './renderRequest';

const exportFormats = ['svg', 'dot', 'html', 'png', 'plantuml'] as const;

export class PreviewPanel implements vscode.Disposable {
    private panel: vscode.WebviewPanel | undefined;
    private document: vscode.TextDocument | undefined;
    private seq = 0;
    private requestSeq = 0;
    private readonly viewStates = new Map<string, PreviewViewState>();
    private scheduleRender: ((doc: vscode.TextDocument) => void) & { cancel(): void };

    private constructor(
        private readonly ready: Promise<void>,
        private readonly client: LanguageClient,
        private readonly manager: DocumentManager,
    ) {
        this.scheduleRender = debounce(
            (doc: vscode.TextDocument) => void this.render(doc),
            this.debounceMs(),
        );
    }

    static create(
        context: vscode.ExtensionContext,
        ready: Promise<void>,
        client: LanguageClient,
        manager: DocumentManager,
    ): PreviewPanel {
        const panel = new PreviewPanel(ready, client, manager);

        context.subscriptions.push(
            vscode.workspace.onDidChangeTextDocument((e) => panel.onDocumentChanged(e)),
            vscode.workspace.onDidCloseTextDocument((doc) => panel.onDocumentClosed(doc)),
            vscode.window.onDidChangeActiveTextEditor((editor) => panel.onActiveEditorChanged(editor)),
            panel,
        );

        return panel;
    }

    dispose(): void {
        this.scheduleRender.cancel();
        this.panel?.dispose();
        this.panel = undefined;
        this.document = undefined;
    }

    // showFor opens (or refocuses) the preview for a document.
    showFor(doc: vscode.TextDocument): void {
        if (this.panel === undefined) {
            this.panel = vscode.window.createWebviewPanel(
                'c4drill.preview',
                'C4Drill Preview',
                { viewColumn: vscode.ViewColumn.Beside, preserveFocus: true },
                {
                    enableScripts: true,
                    enableCommandUris: false,
                    localResourceRoots: [],
                },
            );
            this.panel.webview.html = this.html(this.panel.webview);
            this.panel.webview.onDidReceiveMessage(
                (msg: WebviewToHostMessage) => void this.onMessage(msg),
                undefined,
                [],
            );
            this.panel.onDidDispose(() => {
                this.scheduleRender.cancel();
                this.panel = undefined;
                this.document = undefined;
            });
        } else {
            this.panel.reveal(undefined, true);
        }

        this.document = doc;
        this.renderNow(doc);
    }

    // ---- event wiring ----

    private onDocumentChanged(e: vscode.TextDocumentChangeEvent): void {
        if (this.panel === undefined || this.document === undefined) {
            return;
        }

        if (e.document.uri.toString() !== this.document.uri.toString()) {
            return;
        }

        this.document = e.document;
        this.scheduleRender(e.document);
    }

    private onDocumentClosed(doc: vscode.TextDocument): void {
        if (this.document !== undefined && doc.uri.toString() === this.document.uri.toString()) {
            this.showInfo('The document was closed. Open a c4drill model and press the preview button.');
        }
    }

    // The preview follows the active editor while visible, mirroring a GUI
    // preview area: switching to another managed c4drill document retargets
    // the panel; other editors leave it alone.
    private onActiveEditorChanged(editor: vscode.TextEditor | undefined): void {
        if (
            this.panel === undefined ||
            this.document === undefined ||
            editor === undefined ||
            !this.panel.visible ||
            !this.manager.shouldHandle(editor.document) ||
            editor.document.uri.toString() === this.document.uri.toString()
        ) {
            return;
        }

        this.document = editor.document;
        this.renderNow(editor.document);
    }

    // ---- webview messages ----

    private async onMessage(msg: WebviewToHostMessage): Promise<void> {
        switch (msg.type) {
            case 'ready':
                if (this.document !== undefined) {
                    this.renderNow(this.document);
                }

                break;
            case 'link':
                await this.onLink(msg.href);

                break;
            case 'drillTo':
                this.mutateState((s) => ({ ...s, target: msg.target, allExpanded: false }));
                await this.renderCurrent();

                break;
            case 'toggleExpanded':
                this.mutateState((s) => ({ ...s, allExpanded: !s.allExpanded }));
                await this.renderCurrent();

                break;
            case 'collapseAll':
                this.mutateState((s) => ({ ...s, collapseAll: true }));
                await this.renderCurrent();

                break;
            case 'reload':
                await this.renderCurrent();

                break;
            case 'legend':
                this.mutateState((s) => ({ ...s, legend: msg.value }));
                await this.renderCurrent();

                break;
            case 'expandedText':
                // Editing the input clears the collapse-all override.
                this.mutateState((s) => ({ ...s, expandedText: msg.value, collapseAll: false }));
                await this.renderCurrent();

                break;
            case 'export':
                await this.exportDiagram(msg.format);

                break;
            default:
                break;
        }
    }

    private async onLink(href: string): Promise<void> {
        if (this.document === undefined) {
            return;
        }

        const trimmed = href.trim();

        if (/^https?:\/\//i.test(trimmed)) {
            await vscode.env.openExternal(vscode.Uri.parse(trimmed));

            return;
        }

        if (this.document.uri.scheme !== 'file') {
            return;
        }

        const state = this.currentState();

        // Expanded diagrams carry only external reference links today; treat
        // any relative href there as non-navigable.
        if (state.allExpanded) {
            return;
        }

        const basename = basenameOfUri(this.document.uri.toString());
        const next = resolveRenderTarget(state.target, basename, trimmed);

        if (next === null) {
            return; // not an internal drill-down link
        }

        this.mutateState((s) => ({ ...s, target: next, allExpanded: false }));
        await this.renderCurrent();
    }

    // ---- rendering ----

    private currentState(): PreviewViewState {
        if (this.document === undefined) {
            return initialViewState;
        }

        return this.viewStates.get(this.document.uri.toString()) ?? initialViewState;
    }

    private mutateState(fn: (s: PreviewViewState) => PreviewViewState): void {
        if (this.document === undefined) {
            return;
        }

        const key = this.document.uri.toString();
        this.viewStates.set(key, fn(this.viewStates.get(key) ?? initialViewState));
        this.post({ type: 'state', state: this.viewStates.get(key) as PreviewViewState });
    }

    private renderNow(doc: vscode.TextDocument): void {
        this.scheduleRender.cancel();
        void this.render(doc);
    }

    private async renderCurrent(): Promise<void> {
        if (this.document !== undefined) {
            await this.render(this.document);
        }
    }

    private async render(doc: vscode.TextDocument): Promise<void> {
        if (this.panel === undefined) {
            return;
        }

        const state = this.currentState();
        const seq = ++this.seq;
        const params = buildRenderParams(doc.uri.toString(), state);

        try {
            await this.ready;

            const myRequest = ++this.requestSeq;
            const result = await this.client.sendRequest(RenderDiagramRequest, params);

            if (this.panel === undefined || myRequest !== this.requestSeq) {
                return; // stale response or disposed panel
            }

            if (result === null) {
                this.showError(seq, 'The language server does not have this document open.', [], state);

                return;
            }

            if (result.svg === '' || result.diagnostics.length > 0) {
                // Parse/validate failure: CLI-identical messages, no stale diagram.
                this.showError(
                    seq,
                    result.svg === '' ? 'The model has errors — diagram not rendered.' : 'Diagnostics reported for this model.',
                    result.diagnostics.map(diagnosticToItem),
                    state,
                );

                return;
            }

            const p = this.panel;
            if (p === undefined) {
                return;
            }

            p.title = `C4Drill Preview — ${path.basename(doc.fileName)}${
                state.allExpanded ? ' (expanded)' : viewSuffix(state.target)
            }`;
            this.post({
                type: 'rendered',
                seq,
                svg: result.svg,
                target: state.target,
                allExpanded: state.allExpanded,
                title: p.title,
            });
        } catch (err) {
            this.showError(
                seq,
                `Failed to render: ${err instanceof Error ? err.message : String(err)} — is the c4drill binary available? (c4drill.server.path)`,
                [],
                state,
            );
        }
    }

    private showError(seq: number, reason: string, diagnostics: DiagnosticItem[], state: PreviewViewState): void {
        const p = this.panel;
        if (p !== undefined) {
            p.title = state.allExpanded
                ? 'C4Drill Preview — expanded'
                : `C4Drill Preview — ${viewSuffix(state.target)}`;
        }

        this.post({ type: 'error', seq, reason, diagnostics, target: state.target, allExpanded: state.allExpanded });
    }

    private showInfo(text: string): void {
        this.post({ type: 'info', seq: ++this.seq, text });
    }

    private post(msg: HostToWebviewMessage): void {
        void this.panel?.webview.postMessage(msg);
    }

    private debounceMs(): number {
        return vscode.workspace.getConfiguration('c4drill').get<number>('preview.debounce', 200);
    }

    // ---- exports (CLI conventions: c4drill <file> -f <fmt> -o <dir>) ----

    private async exportDiagram(format: string): Promise<void> {
        const doc = this.document;

        if (doc === undefined || doc.uri.scheme !== 'file') {
            void vscode.window.showErrorMessage('C4Drill: export needs a saved document.');

            return;
        }

        if (!isExportFormat(format)) {
            void vscode.window.showErrorMessage(`C4Drill: unsupported export format "${format}".`);

            return;
        }

        const defaultDir = vscode.Uri.file(path.dirname(doc.uri.fsPath));
        const picked = await vscode.window.showOpenDialog({
            canSelectFolders: true,
            canSelectFiles: false,
            canSelectMany: false,
            defaultUri: defaultDir,
            openLabel: `Export ${format} here`,
            title: `C4Drill: export diagrams (${format})`,
        });

        if (picked === undefined || picked.length === 0) {
            return;
        }

        const outDir = picked[0].fsPath;
        const command = resolveServerCommand();

        cp.execFile(
            command,
            [doc.uri.fsPath, '-f', format, '-o', outDir],
            { timeout: 120_000 },
            (err, _stdout, stderr) => {
                if (err !== null) {
                    const detail = stderr.trim() !== '' ? stderr.trim() : err.message;
                    void vscode.window.showErrorMessage(`C4Drill: export failed — ${detail}`);

                    return;
                }

                void vscode.window.showInformationMessage(
                    `C4Drill: exported ${format} diagram(s) to ${outDir}`,
                );
            },
        );
    }

    // ---- webview HTML ----

    private html(webview: vscode.Webview): string {
        const nonce = Array.from({ length: 24 }, () => Math.floor(Math.random() * 36).toString(36)).join('');
        const cspSource = webview.cspSource;

        return `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta http-equiv="Content-Security-Policy" content="default-src 'none'; img-src data: ${cspSource} https:; style-src 'unsafe-inline'; script-src 'nonce-${nonce}';">
<style>
  :root { --gap: 6px; }
  body { margin: 0; padding: 0; color: var(--vscode-foreground); background: var(--vscode-editor-background); font-family: var(--vscode-font-family); font-size: var(--vscode-font-size, 13px); }
  #toolbar { display: flex; flex-wrap: wrap; align-items: center; gap: var(--gap); padding: 6px 8px; border-bottom: 1px solid var(--vscode-panel-border, #333); position: sticky; top: 0; background: var(--vscode-editor-background); z-index: 2; }
  #toolbar button, #toolbar select, #toolbar input { font-family: inherit; font-size: inherit; color: var(--vscode-input-foreground); background: var(--vscode-button-secondaryBackground, #3a3d41); border: 1px solid var(--vscode-panel-border, #555); border-radius: 3px; padding: 3px 8px; }
  #toolbar button { cursor: pointer; }
  #toolbar button[aria-pressed="true"] { background: var(--vscode-button-background); color: var(--vscode-button-foreground); }
  #toolbar label { display: inline-flex; align-items: center; gap: 4px; opacity: 0.9; }
  #expanded-input { width: 170px; }
  #crumbs { display: inline-flex; align-items: center; gap: 2px; flex-wrap: wrap; }
  #crumbs .crumb { cursor: pointer; text-decoration: underline; }
  #crumbs .crumb.current { cursor: default; text-decoration: none; font-weight: 600; }
  #crumbs .sep { opacity: 0.6; }
  #content { padding: 10px; overflow: auto; }
  #content svg { max-width: 100%; height: auto; display: block; margin: 0 auto; }
  #content svg a { cursor: pointer; }
  .box { margin: 24px auto; max-width: 720px; padding: 12px 16px; border: 1px solid var(--vscode-panel-border, #555); border-radius: 4px; }
  .error .headline { color: var(--vscode-errorForeground); font-weight: 600; margin: 0 0 8px 0; }
  ul.diagnostics { margin: 4px 0 0 0; padding-left: 20px; }
  ul.diagnostics li { margin: 2px 0; word-break: break-word; }
  .line-no { opacity: 0.75; font-family: var(--vscode-editor-font-family); margin-right: 6px; }
  .muted { opacity: 0.75; }
  #status { padding: 2px 8px; font-size: 11px; opacity: 0.7; border-top: 1px solid var(--vscode-panel-border, #333); position: fixed; bottom: 0; width: 100%; box-sizing: border-box; background: var(--vscode-editor-background); }
</style>
</head>
<body>
<div id="toolbar">
  <span id="crumbs" title="View path"></span>
  <button id="btn-expanded" aria-pressed="false" title="Render the single all-nested diagram (c4drill --expanded)">Expanded view</button>
  <button id="btn-collapse" title="Override [properties].expanded with an empty set (collapse all)">Collapse all</button>
  <label for="expanded-input" title="Comma-separated unit paths replacing the [properties].expanded C1 drill-down set; empty = model default">Expanded set</label>
  <input id="expanded-input" type="text" spellcheck="false" placeholder="model default">
  <label for="legend-select">Legend</label>
  <select id="legend-select" title="Override properties.legend">
    <option value="default">default</option>
    <option value="on">on</option>
    <option value="off">off</option>
  </select>
  <button id="btn-reload" title="Re-render now">Reload</button>
  <label for="export-select">Export</label>
  <select id="export-select">${exportFormats.map((f) => `<option value="${f}">${f}</option>`).join('')}</select>
  <button id="btn-export" title="Export via the c4drill CLI (c4drill <file> -f <fmt> -o <dir>)">Export…</button>
</div>
<div id="content"><p class="muted box">Loading preview…</p></div>
<div id="status"></div>
<script nonce="${nonce}">
(function () {
  var vscode = acquireVsCodeApi();
  var lastSeq = -1;
  var state = null;

  function $(id) { return document.getElementById(id); }

  function crumbClick(ev) {
    var el = ev.target;
    if (el && el.dataset && el.dataset.target !== undefined && !el.classList.contains('current')) {
      vscode.postMessage({ type: 'drillTo', target: el.dataset.target });
    }
  }

  function svgClick(ev) {
    var a = ev.target && ev.target.closest ? ev.target.closest('a') : null;
    if (!a) { return; }
    ev.preventDefault();
    ev.stopPropagation();
    var href = a.getAttribute('href') || a.getAttribute('xlink:href') || a.getAttributeNS('http://www.w3.org/1999/xlink', 'href') || '';
    if (href !== '') {
      vscode.postMessage({ type: 'link', href: href });
    }
  }

  document.addEventListener('click', function (ev) {
    var crumb = ev.target.closest ? ev.target.closest('.crumb') : null;
    if (crumb) { crumbClick(ev); return; }
    var inSvg = ev.target.closest && ev.target.closest('svg');
    if (inSvg) { svgClick(ev); }
  });

  $('btn-expanded').addEventListener('click', function () { vscode.postMessage({ type: 'toggleExpanded' }); });
  $('btn-collapse').addEventListener('click', function () { vscode.postMessage({ type: 'collapseAll' }); });
  $('btn-reload').addEventListener('click', function () { vscode.postMessage({ type: 'reload' }); });
  $('btn-export').addEventListener('click', function () {
    vscode.postMessage({ type: 'export', format: $('export-select').value });
  });
  $('legend-select').addEventListener('change', function () {
    vscode.postMessage({ type: 'legend', value: this.value });
  });
  $('expanded-input').addEventListener('change', function () {
    vscode.postMessage({ type: 'expandedText', value: this.value });
  });

  function setState(s) {
    state = s;
    $('btn-expanded').setAttribute('aria-pressed', s.allExpanded ? 'true' : 'false');
    if (document.activeElement !== $('expanded-input')) {
      $('expanded-input').value = s.expandedText || '';
    }
    $('legend-select').value = s.legend;
    var crumbs = [$('<span class="crumb' + (crumbCurrent(s, '') ? ' current' : '') + '" data-target="" title="C1 context diagram">C1</span>')];
    if (s.allExpanded) {
      crumbs.push('<span class="crumb current">expanded</span>');
    } else if (s.target !== '') {
      var parts = s.target.split('.');
      for (var i = 0; i < parts.length; i++) {
        var t = parts.slice(0, i + 1).join('.');
        crumbs.push('<span class="sep">/</span>');
        crumbs.push('<span class="crumb' + (i === parts.length - 1 ? ' current' : '') + '" data-target="' + t + '">' + escapeHtml(parts[i]) + '</span>');
      }
    }
    var host = $('crumbs');
    host.innerHTML = '';
    crumbs.forEach(function (c) { if (typeof c === 'string') { host.insertAdjacentHTML('beforeend', c); } else { host.appendChild(c); } });
  }

  function crumbCurrent(s, target) { return !s.allExpanded && s.target === target; }

  function escapeHtml(text) {
    return text.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
  }

  function renderSvg(msg) {
    $('content').innerHTML = '<div id="diagram">' + msg.svg + '</div>';
    $('status').textContent = msg.title;
  }

  function renderError(msg) {
    var items = (msg.diagnostics || []).map(function (d) {
      var line = d.line >= 0 ? '<span class="line-no">' + (d.line + 1) + ':</span>' : '';
      return '<li>' + line + escapeHtml(d.message) + '</li>';
    }).join('');
    $('content').innerHTML =
      '<div class="box error"><p class="headline">' + escapeHtml(msg.reason) + '</p>' +
      (items !== ''
        ? '<ul class="diagnostics">' + items + '</ul>'
        : '<p class="muted">No diagnostics returned.</p>') +
      '</div>';
    $('status').textContent = 'error';
  }

  function renderInfo(text) {
    $('content').innerHTML = '<p class="muted box">' + escapeHtml(text) + '</p>';
    $('status').textContent = '';
  }

  window.addEventListener('message', function (ev) {
    var msg = ev.data;
    if (msg.type === 'state') { setState(msg.state); return; }
    if (msg.seq !== undefined && msg.seq < lastSeq) { return; }
    if (msg.seq !== undefined) { lastSeq = msg.seq; }
    if (msg.type === 'rendered') { renderSvg(msg); }
    else if (msg.type === 'error') { renderError(msg); }
    else if (msg.type === 'info') { renderInfo(msg.text); }
  });

  vscode.postMessage({ type: 'ready' });
})();
</script>
</body>
</html>`;
    }
}

function isExportFormat(f: string): boolean {
    return (exportFormats as readonly string[]).includes(f);
}

function diagnosticToItem(d: { message: string; range?: { start?: { line?: number } }; source?: string }): DiagnosticItem {
    return {
        message: d.message,
        line: d.range?.start?.line !== undefined ? d.range.start.line : -1,
        source: d.source ?? 'c4drill',
    };
}

function viewSuffix(target: string): string {
    return target === '' ? 'C1' : target.split('.').join(' / ');
}
