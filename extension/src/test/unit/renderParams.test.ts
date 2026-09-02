import { test } from 'node:test';
import * as assert from 'node:assert/strict';
import {
    breadcrumbTrail,
    buildRenderParams,
    initialViewState,
    parseExpandedSet,
} from '../../core/renderParams';

test('default view state renders the C1 context diagram', () => {
    const params = buildRenderParams('file:///w/model.toml', initialViewState);

    assert.deepEqual(params, {
        textDocument: { uri: 'file:///w/model.toml' },
        target: '',
        format: 'svg',
    });
});

test('a deep target is passed through for C3 rendering', () => {
    const params = buildRenderParams('file:///w/model.toml', {
        ...initialViewState,
        target: 'amazon.rds.pg',
    });

    assert.equal(params.target, 'amazon.rds.pg');
    assert.equal(params.allExpanded, undefined);
});

test('allExpanded overrides the target', () => {
    const params = buildRenderParams('file:///w/model.toml', {
        ...initialViewState,
        target: 'cloud',
        allExpanded: true,
    });

    assert.equal(params.allExpanded, true);
    assert.equal(params.target, undefined);
});

test('expanded text maps to the expanded array override', () => {
    const params = buildRenderParams('file:///w/model.toml', {
        ...initialViewState,
        expandedText: 'cloud, amazon.rds  next',
    });

    assert.deepEqual(params.expanded, ['cloud', 'amazon.rds', 'next']);
});

test('blank expanded text omits the field (model default)', () => {
    const params = buildRenderParams('file:///w/model.toml', {
        ...initialViewState,
        expandedText: '  ,  ',
    });

    assert.equal(params.expanded, undefined);
});

test('collapseAll sends the explicit empty-array override', () => {
    const params = buildRenderParams('file:///w/model.toml', {
        ...initialViewState,
        collapseAll: true,
        expandedText: 'cloud',
    });

    assert.deepEqual(params.expanded, []);
});

test('legend choices map to the boolean override or omission', () => {
    assert.equal(buildRenderParams('u', initialViewState).legend, undefined);
    assert.equal(buildRenderParams('u', { ...initialViewState, legend: 'on' }).legend, true);
    assert.equal(buildRenderParams('u', { ...initialViewState, legend: 'off' }).legend, false);
});

test('breadcrumbs chain from the C1 root', () => {
    const trail = breadcrumbTrail({ ...initialViewState, target: 'amazon.rds' });

    assert.deepEqual(
        trail.map((e) => [e.label, e.target, e.current]),
        [
            ['C1', '', false],
            ['amazon', 'amazon', false],
            ['rds', 'amazon.rds', true],
        ],
    );
});

test('C1 trail is a single current entry', () => {
    const trail = breadcrumbTrail(initialViewState);

    assert.equal(trail.length, 1);
    assert.equal(trail[0].current, true);
});

test('allExpanded mode appends the expanded marker', () => {
    const trail = breadcrumbTrail({ ...initialViewState, allExpanded: true });

    assert.deepEqual(
        trail.map((e) => e.label),
        ['C1', 'expanded'],
    );
    assert.equal(trail[1].current, true);
    assert.equal(trail[0].current, false);
});

test('parseExpandedSet is nil-for-blank and splits on mixed separators', () => {
    assert.equal(parseExpandedSet(''), undefined);
    assert.equal(parseExpandedSet('   '), undefined);
    assert.deepEqual(parseExpandedSet('a,b'), ['a', 'b']);
    assert.deepEqual(parseExpandedSet('a.b\n c ,\td'), ['a.b', 'c', 'd']);
});
