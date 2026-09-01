import { test } from 'node:test';
import * as assert from 'node:assert/strict';
import { isManagedDocument } from '../../core/documentScope';

const base = {
    fsPath: '/w/service.architecture.toml',
    relativePath: 'service.architecture.toml',
    patterns: ['**/*.architecture.toml'] as string[],
    activatedUris: new Set<string>(),
};

test('.c4d documents are always managed, globs or no globs', () => {
    assert.equal(isManagedDocument({ ...base, uri: 'file:///w/model.c4d', fsPath: '/w/model.c4d', relativePath: 'model.c4d', patterns: [] }), true);
    assert.equal(isManagedDocument({ ...base, uri: 'file:///w/model.c4d', fsPath: '/w/model.c4d', relativePath: 'model.c4d' }), true);
});

test('.toml managed when a configured glob matches', () => {
    assert.equal(isManagedDocument({ ...base, uri: 'file:///w/service.architecture.toml' }), true);
    assert.equal(
        isManagedDocument({
            ...base,
            uri: 'file:///w/arch/service.toml',
            fsPath: '/w/arch/service.toml',
            relativePath: 'arch/service.toml',
            patterns: ['arch/*.toml'],
        }),
        true,
    );
});

test('.toml NOT managed for plain TOML files', () => {
    assert.equal(
        isManagedDocument({
            ...base,
            uri: 'file:///w/plain.toml',
            fsPath: '/w/plain.toml',
            relativePath: 'plain.toml',
        }),
        false,
    );
});

test('.toml NOT managed when no patterns configured (opt-in only)', () => {
    assert.equal(
        isManagedDocument({
            ...base,
            uri: 'file:///w/service.architecture.toml',
            patterns: [],
        }),
        false,
    );
});

test('explicit activation manages any TOML regardless of globs', () => {
    const activated = new Set(['file:///w/plain.toml']);

    assert.equal(
        isManagedDocument({
            ...base,
            uri: 'file:///w/plain.toml',
            fsPath: '/w/plain.toml',
            relativePath: 'plain.toml',
            patterns: [],
            activatedUris: activated,
        }),
        true,
    );
});

test('activation is keyed on the URI, not the file type', () => {
    const activated = new Set(['untitled:Untitled-1']);

    assert.equal(
        isManagedDocument({
            ...base,
            uri: 'untitled:Untitled-1',
            fsPath: 'Untitled-1',
            relativePath: undefined,
            patterns: [],
            activatedUris: activated,
        }),
        true,
    );
});

test('other file types are never managed', () => {
    assert.equal(
        isManagedDocument({
            ...base,
            uri: 'file:///w/notes.txt',
            fsPath: '/w/notes.txt',
            relativePath: 'notes.txt',
        }),
        false,
    );
});

test('extension match is case-insensitive (uppercase .TOML opts in via its glob)', () => {
    assert.equal(
        isManagedDocument({
            ...base,
            uri: 'file:///w/MODEL.TOML',
            fsPath: '/w/MODEL.TOML',
            relativePath: 'MODEL.TOML',
            patterns: ['**/*.TOML'],
        }),
        true,
    );
    assert.equal(
        isManagedDocument({
            ...base,
            uri: 'file:///w/model.C4D',
            fsPath: '/w/model.C4D',
            relativePath: 'model.C4D',
            patterns: [],
        }),
        true,
    );
});
