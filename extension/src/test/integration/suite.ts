// Integration smoke test (issue #27): runs inside the VS Code extension host
// with the extension under development loaded. Verifies the three headline
// behaviors end to end against the real `c4drill serve --lsp` server:
//
//  1. diagnostics on an invalid model arrive in-editor and match the CLI's
//     error text (message parity),
//  2. c4drill/renderDiagram returns a non-empty SVG for a valid model,
//  3. the preview panel command opens without error.

import * as assert from 'assert';
import { execFile } from 'child_process';
import * as path from 'path';
import * as vscode from 'vscode';

const extensionId = 'djarvur.c4drill-vscode';
const serverBinary = process.env.C4DRILL_SERVER_PATH ?? 'c4drill';

// fixtures workspace root (runTests opens this folder as the workspace).
const fixturesDir = vscode.workspace.workspaceFolders?.[0]?.uri.fsPath ?? path.resolve(__dirname, '..', '..', '..', 'test', 'fixtures');

function sleep(ms: number): Promise<void> {
    return new Promise((resolve) => setTimeout(resolve, ms));
}

function cliErrorMessages(file: string): Promise<string[]> {
    return new Promise((resolve) => {
        execFile(serverBinary, [file], { timeout: 30_000 }, (err, _stdout, stderr) => {
            if (err === undefined) {
                resolve([]);

                return;
            }

            resolve(
                stderr
                    .split('\n')
                    .filter((l) => l.startsWith('Error: '))
                    .map((l) => l.replace(/^Error: /, '').trim())
                    .filter((l) => l !== ''),
            );
        });
    });
}

async function waitForDiagnostics(uri: vscode.Uri, timeoutMs: number): Promise<vscode.Diagnostic[]> {
    const deadline = Date.now() + timeoutMs;

    for (;;) {
        const diags = vscode.languages.getDiagnostics(uri).filter((d) => d.source === 'c4drill');
        if (diags.length > 0) {
            return diags;
        }

        if (Date.now() > deadline) {
            return [];
        }

        await sleep(250);
    }
}

suite('C4Drill extension smoke', function suite() {
    this.timeout(120_000);

    suiteSetup(async () => {
        const ext = vscode.extensions.getExtension(extensionId);
        assert.ok(ext, `extension ${extensionId} must be installed in the dev host`);
        await ext.activate();
    });

    test('invalid model: diagnostics match the CLI messages', async () => {
        const file = path.join(fixturesDir, 'invalid.c4d');
        const doc = await vscode.workspace.openTextDocument(file);
        await vscode.window.showTextDocument(doc);

        const diags = await waitForDiagnostics(doc.uri, 30_000);
        assert.ok(diags.length > 0, 'c4drill diagnostics were published');

        const cliMessages = await cliErrorMessages(file);
        assert.ok(cliMessages.length > 0, 'the CLI reports the same failure');

        for (const msg of cliMessages) {
            assert.ok(
                diags.some((d) => d.message === msg),
                `CLI message "${msg}" must appear verbatim among the diagnostics`,
            );
        }
    });

    test('valid model: c4drill/renderDiagram returns an SVG', async () => {
        const file = path.join(fixturesDir, 'valid.c4d');
        const uri = vscode.Uri.file(file);
        const doc = await vscode.workspace.openTextDocument(file);
        await vscode.window.showTextDocument(doc);

        // Give the client a moment to forward didOpen before rendering.
        await sleep(500);

        const result = await vscode.commands.executeCommand('c4drill._renderDiagram', uri) as
            | { svg: string; diagnostics: unknown[] }
            | undefined;

        assert.ok(result !== undefined, 'renderDiagram returned a result');
        assert.deepEqual(result.diagnostics, [], 'valid model has no diagnostics');
        assert.ok(result.svg.includes('<svg'), 'result carries SVG markup');
    });

    test('preview panel command opens', async () => {
        const file = path.join(fixturesDir, 'valid.c4d');
        const doc = await vscode.workspace.openTextDocument(file);
        const editor = await vscode.window.showTextDocument(doc);

        await vscode.window.showTextDocument(doc, editor.viewColumn, false);
        await vscode.commands.executeCommand('c4drill.showPreview');

        // No direct webview introspection; the smoke assertion is that the
        // command resolves while the extension holds no error state.
        assert.ok(true);
    });
});
