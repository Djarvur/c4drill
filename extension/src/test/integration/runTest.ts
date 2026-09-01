// Integration smoke test runner (issue #27): downloads/launches VS Code via
// @vscode/test-electron with this extension in development mode and runs the
// mocha suite in index.ts against test/fixtures.
//
// Usage: npm run compile && node out/test/integration/runTest.js
// The c4drill binary is resolved like in production: C4DRILL_SERVER_PATH env
// (forwarded to the extension host) or `c4drill` on PATH.

import * as path from 'path';
import { runTests } from '@vscode/test-electron';

async function main(): Promise<void> {
    // out/test/integration/runTest.js -> extension/
    const extensionDevelopmentPath = path.resolve(__dirname, '..', '..', '..');
    const extensionTestsPath = path.resolve(__dirname, 'index');
    const workspacePath = path.resolve(extensionDevelopmentPath, 'test', 'fixtures');

    await runTests({
        extensionDevelopmentPath,
        extensionTestsPath,
        // The workspace is passed as the first launch argument (the folder VS
        // Code opens; there is no dedicated workspacePath option).
        launchArgs: [workspacePath, '--disable-extensions'],
        extensionTestsEnv: {
            C4DRILL_SERVER_PATH: process.env.C4DRILL_SERVER_PATH ?? 'c4drill',
        },
    });
}

main().catch((err: unknown) => {
    console.error('Integration tests failed:', err);
    process.exit(1);
});
