import { test } from 'node:test';
import * as assert from 'node:assert/strict';
import { debounce } from '../../core/debounce';

const sleep = (ms: number): Promise<void> => new Promise((resolve) => setTimeout(resolve, ms));

test('debounce coalesces bursts into a single trailing call', async () => {
    let calls = 0;
    let lastArg = '';

    const fn = debounce((tag: string) => {
        calls++;
        lastArg = tag;
    }, 20);

    fn('a');
    await sleep(5);
    fn('b');
    await sleep(5);
    fn('c');

    assert.equal(calls, 0, 'no call while the burst continues');

    await sleep(40);

    assert.equal(calls, 1, 'exactly one trailing call');
    assert.equal(lastArg, 'c', 'the trailing call carries the last arguments');
});

test('debounce fires once after the quiet period elapses', async () => {
    let calls = 0;

    const fn = debounce(() => {
        calls++;
    }, 15);

    fn();
    await sleep(30);
    fn();
    await sleep(30);

    assert.equal(calls, 2);
});

test('debounce.cancel drops the pending call', async () => {
    let calls = 0;

    const fn = debounce(() => {
        calls++;
    }, 15);

    fn();
    fn.cancel();
    await sleep(30);

    assert.equal(calls, 0);
});

test('debounce.flush runs the pending call immediately', async () => {
    let calls = 0;
    let value = '';

    const fn = debounce((v: string) => {
        calls++;
        value = v;
    }, 500);

    fn('now');
    fn.flush();

    assert.equal(calls, 1);
    assert.equal(value, 'now');

    await sleep(20);

    assert.equal(calls, 1, 'the trailing timer was cleared by flush');
});

test('debounce.flush with nothing pending is a no-op', () => {
    let calls = 0;

    const fn = debounce(() => {
        calls++;
    }, 10);

    fn.flush();

    assert.equal(calls, 0);
});
