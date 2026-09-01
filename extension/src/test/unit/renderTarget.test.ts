import { test } from 'node:test';
import * as assert from 'node:assert/strict';
import { basenameOfUri, resolveRenderTarget, virtualFilePath } from '../../core/renderTarget';

// The expected virtual layout mirrors internal/output/writer.go Write and the
// link computation in internal/graph/path.go ComputeExploreURL. Verified
// against real CLI output for examples/cloud-system:
//
//	cloud-system.svg                 (C1)
//	cloud-system/cloud.svg           (C2 "cloud")
//	cloud-system/amazon/rds.svg      (C3 "amazon.rds")
//
// with hrefs like "cloud-system/amazon/rds.svg" from C1, "amazon.svg" and
// "../cloud-system.svg" from C2, "../amazon.svg" from C3.
test('virtualFilePath mirrors the CLI output layout', () => {
    assert.equal(virtualFilePath('', 'cloud-system'), 'cloud-system.svg');
    assert.equal(virtualFilePath('cloud', 'cloud-system'), 'cloud-system/cloud.svg');
    assert.equal(virtualFilePath('amazon.rds', 'cloud-system'), 'cloud-system/amazon/rds.svg');
    assert.equal(virtualFilePath('a.b.c', 'm'), 'm/a/b/c.svg');
});

test('basenameOfUri strips scheme, directories and the last extension only', () => {
    assert.equal(basenameOfUri('file:///w/models/cloud-system.toml'), 'cloud-system');
    assert.equal(basenameOfUri('file:///w/diagram.c4d'), 'diagram');
    assert.equal(basenameOfUri('file:///w/my%20model.architecture.toml'), 'my model.architecture');
});

test('C1 drill-down links resolve to full dotted targets', () => {
    assert.equal(resolveRenderTarget('', 'cloud-system', 'cloud-system/amazon/rds.svg'), 'amazon.rds');
    assert.equal(resolveRenderTarget('', 'cloud-system', 'cloud-system/cloud.svg'), 'cloud');
});

test('C2/C3 sibling, descendant and back links resolve', () => {
    // From C2 "cloud" (file cloud-system/cloud.svg).
    assert.equal(resolveRenderTarget('cloud', 'cloud-system', 'amazon.svg'), 'amazon');
    assert.equal(resolveRenderTarget('cloud', 'cloud-system', '../cloud-system.svg'), '');

    // From C3 "amazon.rds" (file cloud-system/amazon/rds.svg).
    assert.equal(resolveRenderTarget('amazon.rds', 'cloud-system', '../amazon.svg'), 'amazon');
    assert.equal(resolveRenderTarget('amazon.rds', 'cloud-system', '../cloud.svg'), 'cloud');
    assert.equal(resolveRenderTarget('amazon.rds', 'cloud-system', '../../cloud-system.svg'), '');
});

test('URL-encoded segments are decoded into the dotted path', () => {
    assert.equal(resolveRenderTarget('', 'm', 'm/a%20b/c~d.svg'), 'a b.c~d');
});

test('query and fragment suffixes are tolerated', () => {
    assert.equal(resolveRenderTarget('', 'm', 'm/a/b.svg#frag'), 'a.b');
    assert.equal(resolveRenderTarget('', 'm', 'm/a/b.svg?x=1'), 'a.b');
});

test('external http(s) reference links are not internal targets', () => {
    assert.equal(resolveRenderTarget('', 'm', 'https://example.com/docs'), null);
    assert.equal(resolveRenderTarget('a', 'm', 'http://example.com/x.svg'), null);
    assert.equal(resolveRenderTarget('', 'm', 'mailto:someone@example.com'), null);
});

test('links escaping the diagram tree resolve to null', () => {
    assert.equal(resolveRenderTarget('a', 'm', '../../../../etc/passwd.svg'), null);
    assert.equal(resolveRenderTarget('', 'm', ''), null);
});

test('self-links and empty hrefs are null', () => {
    assert.equal(resolveRenderTarget('a', 'm', ''), null);
});
