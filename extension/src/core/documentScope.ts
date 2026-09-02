// Document scoping for the c4drill TOML dialect (issue #27): the extension
// must NOT hijack other projects' TOML files. A .toml document is handled by
// the c4drill language server only when it opts in — either via a
// `c4drill.toml.patterns` workspace glob or via the explicit
// "C4Drill: Activate for This File" command (persisted per workspace).
//
// .c4d documents are always handled (the language is claimed unconditionally).
//
// Pure logic, no VS Code imports, so it is unit-testable with node:test.

import { matchGlob } from './globMatch';

export const c4dExtension = '.c4d';

// isManagedDocument reports whether the c4drill LSP should attach to a text
// document. activatedUris carries URIs (as returned by Uri.toString()) that
// the user activated explicitly.
export function isManagedDocument(args: {
    uri: string;
    fsPath: string;
    /** Path relative to the workspace folder, '/'-separated; undefined outside a workspace. */
    relativePath?: string;
    patterns: readonly string[];
    activatedUris: ReadonlySet<string>;
}): boolean {
    if (args.activatedUris.has(args.uri)) {
        return true;
    }

    const ext = extensionOf(args.fsPath);

    if (ext === c4dExtension) {
        return true;
    }

    if (ext !== '.toml' || args.patterns.length === 0) {
        return false;
    }

    const candidates = [args.fsPath.split('\\').join('/')];

    if (args.relativePath !== undefined) {
        candidates.unshift(args.relativePath.split('\\').join('/'));
    }

    return args.patterns.some((p) => matchGlob(p, candidates));
}

function extensionOf(path: string): string {
    const file = path.slice(Math.max(path.lastIndexOf('/'), path.lastIndexOf('\\')) + 1);
    const dot = file.lastIndexOf('.');

    return dot > 0 ? file.slice(dot).toLowerCase() : '';
}
