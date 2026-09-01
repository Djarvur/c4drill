// LSP client wiring (issue #27): spawns `c4drill serve --lsp` over stdio via
// vscode-languageclient and forwards diagnostics, completion, hover,
// definition, documentSymbol and formatting.
//
// Document scoping lives in the middleware: the client registers broadly
// (all .c4d and .toml files must be selectable so requests can flow after a
// delayed activation), but didOpen/didChange/didSave are only forwarded for
// documents that opt in per core/documentScope.ts. The server never sees a
// plain TOML file, so plain-TOML projects keep their own tooling.

import * as vscode from 'vscode';
import {
    DidChangeTextDocumentNotification,
    DidCloseTextDocumentNotification,
    DidOpenTextDocumentNotification,
    LanguageClient,
    LanguageClientOptions,
    Middleware,
    ServerOptions,
} from 'vscode-languageclient/node';

import { isManagedDocument } from './core/documentScope';

export const CLIENT_ID = 'c4drill';

// DocumentManager tracks which TOML documents opted in to c4drill handling
// ("C4Drill: Activate for This File") and answers the middleware's gating
// queries.
export class DocumentManager {
    private readonly activated: Set<string> = new Set();

    constructor(private readonly context: vscode.ExtensionContext) {
        const persisted = context.workspaceState.get<string[]>(activated_key, []);
        for (const uri of persisted) {
            this.activated.add(uri);
        }
    }

    shouldHandle(doc: vscode.TextDocument): boolean {
        return isManagedDocument({
            uri: doc.uri.toString(),
            fsPath: doc.uri.fsPath,
            relativePath: relativePathOf(doc.uri),
            patterns: tomlPatterns(),
            activatedUris: this.activated,
        });
    }

    handlesUri(uri: vscode.Uri): boolean {
        return isManagedDocument({
            uri: uri.toString(),
            fsPath: uri.fsPath,
            relativePath: relativePathOf(uri),
            patterns: tomlPatterns(),
            activatedUris: this.activated,
        });
    }

    activate(uri: vscode.Uri): boolean {
        const key = uri.toString();
        if (this.activated.has(key)) {
            return false;
        }

        this.activated.add(key);
        this.persist();

        return true;
    }

    deactivate(uri: vscode.Uri): boolean {
        const key = uri.toString();
        if (!this.activated.has(key)) {
            return false;
        }

        this.activated.delete(key);
        this.persist();

        return true;
    }

    private persist(): void {
        void this.context.workspaceState.update(activated_key, [...this.activated]);
    }
}

const activated_key = 'c4drill.activatedTomlUris';

function tomlPatterns(): string[] {
    return vscode.workspace.getConfiguration('c4drill').get<string[]>('toml.patterns', []);
}

function relativePathOf(uri: vscode.Uri): string | undefined {
    const folder = vscode.workspace.getWorkspaceFolder(uri);

    return folder === undefined ? undefined : vscode.workspace.asRelativePath(uri, false);
}

// resolveServerCommand finds the c4drill binary: the c4drill.server.path
// setting wins; when it is untouched the C4DRILL_SERVER_PATH environment
// variable is honored (used by the integration tests); finally PATH is used.
export function resolveServerCommand(): string {
    const config = vscode.workspace.getConfiguration('c4drill');
    const inspected = config.inspect<string>('server.path');
    const explicitlySet =
        inspected?.workspaceValue !== undefined ||
        inspected?.globalValue !== undefined ||
        inspected?.workspaceFolderValue !== undefined;

    if (explicitlySet) {
        const value = config.get<string>('server.path', 'c4drill');
        if (value.trim() !== '') {
            return value;
        }
    }

    const fromEnv = process.env.C4DRILL_SERVER_PATH;

    return fromEnv !== undefined && fromEnv.trim() !== '' ? fromEnv : 'c4drill';
}

// createClient builds (but does not start) the language client.
export function createClient(manager: DocumentManager): LanguageClient {
    const command = resolveServerCommand();

    const serverOptions: ServerOptions = {
        run: { command, args: ['serve', '--lsp'] },
        debug: { command, args: ['serve', '--lsp'] },
    };

    const middleware: Middleware = {
        didOpen: (doc, next) => (manager.shouldHandle(doc) ? next(doc) : Promise.resolve()),
        didChange: (event, next) => (manager.handlesUri(event.document.uri) ? next(event) : Promise.resolve()),
        didSave: (doc, next) => (manager.shouldHandle(doc) ? next(doc) : Promise.resolve()),
        // didClose is always forwarded: the server tolerates closes of
        // unknown documents, and a document deactivated mid-session must
        // still be released server-side.
        didClose: (doc, next) => next(doc),
    };

    const clientOptions: LanguageClientOptions = {
        documentSelector: [
            { language: 'c4drill-c4d' },
            { scheme: 'file', pattern: '**/*.c4d' },
            { scheme: 'file', pattern: '**/*.[tT][oO][mM][lL]' },
        ],
        synchronize: {
            fileEvents: vscode.workspace.createFileSystemWatcher('**/*.{c4d,toml,TOML}'),
        },
        outputChannelName: 'C4Drill',
        middleware,
    };

    return new LanguageClient(CLIENT_ID, 'C4Drill', serverOptions, clientOptions);
}

// nudgeDocument re-sends the document as a didChange so the server reruns
// its validation pipeline (the "C4Drill: Validate File" command). Direct
// sendNotification bypasses the middleware — intended here.
export function nudgeDocument(client: LanguageClient, doc: vscode.TextDocument): void {
    void client.sendNotification(DidChangeTextDocumentNotification.type, {
        textDocument: { uri: doc.uri.toString(), version: doc.version + 1 },
        contentChanges: [{ text: doc.getText() }],
    });
}

// attachDocument sends a synthetic didOpen for a document the client never
// forwarded (used after "C4Drill: Activate for This File").
export function attachDocument(client: LanguageClient, doc: vscode.TextDocument): void {
    void client.sendNotification(DidOpenTextDocumentNotification.type, {
        textDocument: {
            uri: doc.uri.toString(),
            languageId: doc.languageId,
            version: doc.version,
            text: doc.getText(),
        },
    });
}

// detachDocument sends a synthetic didClose after deactivation.
export function detachDocument(client: LanguageClient, doc: vscode.TextDocument): void {
    void client.sendNotification(DidCloseTextDocumentNotification.type, {
        textDocument: { uri: doc.uri.toString() },
    });
}
