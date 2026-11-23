import assert from 'node:assert/strict';
import test from 'node:test';
import { calculate } from '../.test-build/api.js';

const expression = { expression: '0 ÷ (2 + 3)' };
const signal = new AbortController().signal;

test('client posts JSON to the same-origin API and accepts a zero result', async t => {
  const fetch = t.mock.method(globalThis, 'fetch', async () => Response.json({ result: 0 }));
  assert.equal(await calculate(expression, signal), 0);
  const [path, init] = fetch.mock.calls[0].arguments;
  assert.equal(path, '/api/calculate');
  assert.equal(init.method, 'POST');
  assert.deepEqual(JSON.parse(init.body), expression);
  assert.equal(init.headers['Content-Type'], 'application/json');
  assert.equal(init.signal, signal);
});

for (const [name, respond, message] of [
  ['domain error', () => Response.json({ error: { code: 'division_by_zero', message: 'Cannot divide by zero.' } }, { status: 422 }), /divide by zero/],
  ['unavailable API', () => new Response('unavailable', { status: 503 }), /retry/],
  ['malformed success', () => new Response('not JSON'), /Invalid response/],
  ['wrong result type', () => Response.json({ result: '0' }), /Invalid response/],
  ['connection lost', () => { throw new TypeError('offline'); }, /Cannot reach calculator/],
]) {
  test(name, async t => {
    t.mock.method(globalThis, 'fetch', async () => respond());
    await assert.rejects(calculate(expression, signal), message);
  });
}
