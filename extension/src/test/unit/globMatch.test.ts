import { test } from 'node:test';
import * as assert from 'node:assert/strict';
import { expandBraces, matchGlob } from '../../core/globMatch';

test('plain suffix globs match files at any depth (matchBase)', () => {
    assert.equal(matchGlob('*.architecture.toml', ['service.architecture.toml']), true);
    assert.equal(matchGlob('*.architecture.toml', ['models/service.architecture.toml']), true);
    assert.equal(matchGlob('*.architecture.toml', ['/abs/w/models/service.architecture.toml']), true);
    assert.equal(matchGlob('*.architecture.toml', ['service.toml']), false);
    assert.equal(matchGlob('*.architecture.toml', ['arch.toml']), false);
});

test('** crosses path segments', () => {
    assert.equal(matchGlob('**/*.c4drill.toml', ['a.c4drill.toml']), true);
    assert.equal(matchGlob('**/*.c4drill.toml', ['deep/nested/a.c4drill.toml']), true);
    assert.equal(matchGlob('models/**/*.toml', ['models/a.toml']), true);
    assert.equal(matchGlob('models/**/*.toml', ['models/arch/a.toml']), true);
    assert.equal(matchGlob('models/**/*.toml', ['other/a.toml']), false);
});

test('* stays within one segment', () => {
    assert.equal(matchGlob('arch/*.toml', ['arch/a.toml']), true);
    assert.equal(matchGlob('arch/*.toml', ['arch/sub/a.toml']), false);
});

test('? matches exactly one non-separator character', () => {
    assert.equal(matchGlob('model?.toml', ['model1.toml']), true);
    assert.equal(matchGlob('model?.toml', ['model12.toml']), false);
});

test('brace alternation expands', () => {
    assert.equal(matchGlob('**/*.{c4d,arch.toml}', ['x/y.c4d']), true);
    assert.equal(matchGlob('**/*.{c4d,arch.toml}', ['x/y.arch.toml']), true);
    assert.equal(matchGlob('**/*.{c4d,arch.toml}', ['x/y.toml']), false);

    assert.deepEqual(expandBraces('a{b,c}d{e,f}g'), ['abdeg', 'abdfg', 'acdeg', 'acdfg']);
});

test('character classes and negation', () => {
    assert.equal(matchGlob('model[12].toml', ['model1.toml']), true);
    assert.equal(matchGlob('model[12].toml', ['model3.toml']), false);
    assert.equal(matchGlob('model[!1].toml', ['model2.toml']), true);
    assert.equal(matchGlob('model[!1].toml', ['model1.toml']), false);
});

test('unbalanced or literal patterns degrade to literals', () => {
    assert.equal(matchGlob('weird{toml', ['weird{toml']), true);
    assert.equal(matchGlob('weird{toml', ['weirdtoml']), false);
});

test('multiple candidate paths: first match wins', () => {
    // The extension offers [workspace-relative, absolute] candidates.
    assert.equal(
        matchGlob('arch/*.toml', ['w/arch/a.toml', '/Users/x/w/arch/a.toml']),
        false, // relative pattern does not match the absolute form
    );
    assert.equal(
        matchGlob('arch/*.toml', ['/Users/x/w/arch/a.toml', 'arch/a.toml']),
        true,
    );
});
