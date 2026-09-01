// Extension-host test entry (issue #27): @vscode/test-electron requires this
// module to export run(). A Mocha instance with the TDD UI loads the smoke
// suite (suite.ts, compiled next to this file) and fails the run on any
// test failure.

import * as path from 'path';
import Mocha from 'mocha';

export async function run(): Promise<void> {
    const mocha = new Mocha({ ui: 'tdd', color: true, timeout: 120_000 });
    mocha.addFile(path.resolve(__dirname, 'suite.js'));

    const failures = await new Promise<number>((resolve) => {
        mocha.run((f) => resolve(f));
    });

    if (failures > 0) {
        throw new Error(`${failures} integration test(s) failed`);
    }
}
