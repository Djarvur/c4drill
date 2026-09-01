// C4Drill VS Code extension (issue #27): LSP client over `c4drill serve
// --lsp`, C4D grammar/language contributions, TOML-dialect scoping, and the
// live diagram preview panel.

import * as vscode from 'vscode';
import { LanguageClient } from 'vscode-languageclient/node';

import {
    attachDocument,
    createClient,
    detachDocument,
    DocumentManager,
    nudgeDocument,
} from './client';
import { buildRenderParams, initialViewState } from './core/renderParams';
import { RenderDiagramRequest } from './preview/renderRequest';
import { PreviewPanel } from './preview/previewPanel';

let client: LanguageClient | undefined;
// Resolves when the server finished initializing (the v9 replacement for the
// removed onReady()); rejects with the startup failure.
let ready: Promise<void> = Promise.resolve();

function startClient(c: LanguageClient): Promise<void> {
    return c.start().catch((err: unknown) => {
        const detail = err instanceof Error ? err.message : String(err);
        void vscode.window.showErrorMessage(
            `C4Drill: failed to start the language server (c4drill serve --lsp): ${detail}. ` +
                'Install the c4drill binary or point c4drill.server.path at it.',
        );
    });
}

export async function activate(context: vscode.ExtensionContext): Promise<void> {
    const manager = new DocumentManager(context);
    client = createClient(manager);
    ready = startClient(client);

    // C4Drill: Validate File — re-sends the document so the server reruns the
    // CLI-identical validation pipeline and republishes diagnostics.
    context.subscriptions.push(
        vscode.commands.registerTextEditorCommand(
            'c4drill.validateFile',
            (editor) => {
                if (!manager.shouldHandle(editor.document)) {
                    void vscode.window.showInformationMessage(notHandledMessage());

                    return;
                }

                nudgeDocument(client as LanguageClient, editor.document);
            },
        ),
    );

    // C4Drill: Format Document — delegates to the LSP formatter through the
    // built-in format action (the server formats both .c4d and .toml).
    context.subscriptions.push(
        vscode.commands.registerTextEditorCommand('c4drill.formatDocument', () => {
            void vscode.commands.executeCommand('editor.action.formatDocument');
        }),
    );

    // C4Drill: Activate for This File — explicit TOML opt-in, persisted per
    // workspace, independent of c4drill.toml.patterns.
    context.subscriptions.push(
        vscode.commands.registerTextEditorCommand(
            'c4drill.activateForFile',
            async (editor) => {
                if (editor.document.uri.scheme !== 'file' && editor.document.uri.scheme !== 'untitled') {
                    void vscode.window.showInformationMessage('C4Drill: this document cannot be activated.');

                    return;
                }

                manager.activate(editor.document.uri);

                try {
                    await ready;
                    attachDocument(client as LanguageClient, editor.document);
                } catch {
                    // startup failure already surfaced by startClient
                }
            },
        ),
    );

    context.subscriptions.push(
        vscode.commands.registerTextEditorCommand(
            'c4drill.deactivateForFile',
            (editor) => {
                const changed = manager.deactivate(editor.document.uri);
                const c = client;

                if (changed && c !== undefined && c.isRunning()) {
                    detachDocument(c, editor.document);
                }
            },
        ),
    );

    // C4Drill: Show Preview — the live diagram panel.
    const preview = PreviewPanel.create(context, ready, client, manager);

    context.subscriptions.push(
        vscode.commands.registerTextEditorCommand(
            'c4drill.showPreview',
            (editor) => {
                if (!manager.shouldHandle(editor.document)) {
                    void vscode.window.showInformationMessage(notHandledMessage());

                    return;
                }

                preview.showFor(editor.document);
            },
        ),
    );

    // Internal seam (not in the command palette): renders a document through
    // the language server and resolves with the renderDiagram result. Used by
    // the integration tests and available to tooling.
    context.subscriptions.push(
        vscode.commands.registerCommand(
            'c4drill._renderDiagram',
            async (uri: vscode.Uri) => {
                const c = client;

                if (c === undefined) {
                    return undefined;
                }

                await ready;

                const doc = await vscode.workspace.openTextDocument(uri);

                if (!manager.shouldHandle(doc)) {
                    throw new Error(notHandledMessage());
                }

                return c.sendRequest(
                    RenderDiagramRequest,
                    buildRenderParams(uri.toString(), initialViewState),
                );
            },
        ),
    );

    // A changed server path restarts the client; pattern changes apply to
    // newly opened documents.
    context.subscriptions.push(
        vscode.workspace.onDidChangeConfiguration((e) => {
            const c = client;

            if (c !== undefined && e.affectsConfiguration('c4drill.server.path') && c.isRunning()) {
                void c.stop().then(() => {
                    ready = startClient(c);
                });
            }
        }),
    );

    context.subscriptions.push(
        vscode.Disposable.from({
            dispose: () => {
                const c = client;
                client = undefined;

                if (c !== undefined) {
                    void c.stop();
                }
            },
        }),
    );
}

export function deactivate(): Thenable<void> | undefined {
    const c = client;
    client = undefined;

    return c?.stop();
}

function notHandledMessage(): string {
    return 'C4Drill: this file is not handled by c4drill. For .toml models use "C4Drill: Activate for This File" or c4drill.toml.patterns.';
}
