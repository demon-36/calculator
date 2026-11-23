import assert from 'node:assert/strict';
import test from 'node:test';
import { displayedExpression, initialState, reducer } from '../.test-build/calculator.js';

function enter(keys, state = initialState) {
  return [...keys].reduce((s, key) => reducer(s, key), state);
}
function complete(state, result) {
  assert.ok(state.request);
  return reducer(state, { type: 'success', request: state.request, result });
}

test('mixed operators submit one whole expression and record the confirmed result', () => {
  const editing = enter('2+3×4');
  assert.equal(editing.request, null);
  const state = enter('=', editing);
  assert.deepEqual(state.request.body, { expression: '2 + 3 × 4' });
  assert.deepEqual(complete(state, 14).history, [{ expression: '2 + 3 × 4', value: '14' }]);
});

test('repeat equals applies the last operation to the result or a freshly entered number', () => {
  let state = complete(enter('8+3='), 11);
  for (const [keys, left, result] of [['=', 11, 14], ['=', 14, 17], ['20=', 20, 23]]) {
    state = enter(keys, state);
    assert.deepEqual(state.request.body, { expression: left + ' + 3' });
    state = complete(state, result);
  }
  const cleared = enter(['AC'], state);
  assert.equal(cleared.value, '0');
  assert.deepEqual(cleared.history.map(row => row.value), ['23', '17', '14']);
  assert.equal(enter('=', cleared).request, null);
});

test('incomplete, invalid and duplicate input can be corrected before submission', () => {
  assert.equal(enter('8+=').request, null);
  assert.equal(enter('1234567890123456').value, '123456789012345');
  assert.deepEqual(enter('4+×1..25⌫=').request.body, { expression: '4 × 1.2' });
  const bad = enter('8÷0=');
  assert.match(bad.error, /zero/i);
  assert.equal(bad.request, null);
  assert.deepEqual(enter('4=', bad).request.body, { expression: '8 ÷ 4' });
});

test('pending requests ignore double submits and stale responses after AC', () => {
  const first = enter('2+2=');
  assert.equal(enter('=', first), first);
  const second = enter(['AC', ...'9−1='], first);
  for (const reply of [{ type: 'success', result: 4 }, { type: 'failure', message: 'Offline' }]) {
    assert.equal(reducer(second, { ...reply, request: first.request }), second);
  }
  assert.equal(complete(second, 8).value, '8');
});

test('failed requests preserve the expression for retry without adding history', () => {
  const pending = enter('24÷4×2=');
  const failed = reducer(pending, { type: 'failure', request: pending.request, message: 'Try again' });
  assert.equal(failed.error, 'Try again');
  assert.deepEqual(failed.history, []);
  const retried = enter('=', failed);
  assert.notEqual(retried.request, pending.request);
  assert.deepEqual(retried.request.body, pending.request.body);
  assert.equal(complete(retried, 12).history[0].expression, '24 ÷ 4 × 2');
});

test('entering a different operator replaces a pending one and backspace removes it', () => {
  const state = enter('8+3×⌫−4=');
  assert.deepEqual(state.request.body, { expression: '8 + 3 − 4' });
});

test('a zero divisor in a chain is correctable before continuing', () => {
  const bad = enter('3÷0+');
  assert.match(bad.error, /zero/i);
  assert.deepEqual(enter('2+4=', bad).request.body, { expression: '3 ÷ 2 + 4' });
});

test('calculator permits up to 128 operators per request', () => {
  const full = enter('1' + '+1'.repeat(128));
  assert.match(enter('+', full).error, /finish/);
  assert.equal(enter('=', full).request.body.expression, '1' + ' + 1'.repeat(128));
});

test('display shows the expression being calculated or retried', () => {
  const editing = enter('8+3');
  assert.equal(displayedExpression(editing), '8 + 3');
  const pending = enter('=', editing);
  assert.equal(displayedExpression(pending), '8 + 3');
  const result = complete(pending, 11);
  assert.equal(displayedExpression(result), '8 + 3 =');
  const repeated = enter('=', result);
  const failed = reducer(repeated, { type: 'failure', request: repeated.request, message: 'Try again' });
  assert.equal(displayedExpression(failed), '11 + 3');
  assert.equal(displayedExpression(enter('7÷0=')), '7 ÷ 0');
});

test('complete nested expressions submit the same text that appears while editing', () => {
  for (const [keys, expression] of [
    ['(2+3)^2', '(2 + 3) ^ 2'],
    ['√9+7)', '√(9 + 7)'],
    ['√(2+7))', '√((2 + 7))'],
    ['2^3^2', '2 ^ 3 ^ 2'],
    ['−2^2', '−2 ^ 2'],
    ['(−2)^2', '(−2) ^ 2'],
    ['2^(−3)', '2 ^ (−3)'],
    ['(100+50)+10%', '(100 + 50) + 10%'],
    ['200+(5+5)%', '200 + (5 + 5)%'],
    ['3÷0^0', '3 ÷ 0 ^ 0'],
  ]) {
    const editing = enter(keys);
    assert.equal(displayedExpression(editing), expression);
    assert.deepEqual(enter('=', editing).request?.body, { expression });
  }
});

test('incomplete groups and operators stay editable and explain what is missing', () => {
  for (const [keys, error, correction, expression] of [
    ['2+(', /operand|number/i, '5)', '2 + (5)'],
    ['√(1+8)', /clos|\)/i, ')', '√((1 + 8))'],
    ['2^', /operand|number/i, '3', '2 ^ 3'],
  ]) {
    const incomplete = enter(keys + '=');
    assert.equal(incomplete.request, null);
    assert.match(incomplete.error, error);
    assert.deepEqual(enter(correction + '=', incomplete).request?.body, { expression });
  }
});

test('backspace edits a group, percent, and root opening one key at a time', () => {
  assert.equal(displayedExpression(enter('2×√⌫(3+45)%⌫⌫⌫6)')), '2 × (3 + 46)');
});

test('adjacent groups, numbers and roots require an explicit operator', () => {
  for (const keys of ['2(', '2√', '(2)3', '2%3']) {
    const editing = enter(keys.slice(0, -1));
    assert.equal(displayedExpression(enter(keys)), displayedExpression(editing));
  }
});

test('repeated equals preserves a percentage on both the result and a fresh number', () => {
  const result = complete(enter('100+15%='), 115);
  assert.deepEqual(enter('=', result).request.body, { expression: '115 + 15%' });
  assert.deepEqual(enter('150=', result).request.body, { expression: '150 + 15%' });
});

test('repeat uses the last outer binary operation and retains its full right operand', () => {
  for (const [keys, value, next] of [
    ['((2×(3+4)))', 14, '14 × (3 + 4)'],
    ['2+3×4', 14, '14 × 4'],
    ['200+(5+5)%', 220, '220 + (5 + 5)%'],
    ['8+√4+5)', 11, '11 + √(4 + 5)'],
  ]) {
    const finished = complete(enter(keys + '='), value);
    assert.deepEqual(enter('=', finished).request.body, { expression: next });
  }
});

test('standalone roots, percentages, and grouped literals have no binary repeat', () => {
  for (const keys of ['√9+7)', '(5+5)%', '10%', '(2)']) {
    const finished = complete(enter(keys + '='), 1);
    assert.equal(enter('=', finished).request, null);
  }
});

test('raising a negative result to a power preserves the sign of its base', () => {
  const finished = complete(enter('1−3='), -2);
  assert.deepEqual(enter('^2=', finished).request.body, { expression: '(−2) ^ 2' });
});

test('group and function nesting is bounded and remains editable at the boundary', () => {
  const full = enter('('.repeat(31) + '√');
  assert.match(enter('(', full).error, /nest|group|parenthes/i);
  assert.equal(enter('9' + ')'.repeat(32) + '=', full).request.body.expression,
    '('.repeat(31) + '√(9' + ')'.repeat(32));
});

test('editing a confirmed scientific result starts a fresh entry after undoing an operator', () => {
  const finished = complete(enter('10^30='), 1e30);
  const edited = enter('+⌫2+1=', finished);
  assert.deepEqual(edited.request.body, { expression: '2 + 1' });
});

test('negative results use one operation for their sign when continued or repeated', () => {
  const negative = complete(enter('1−2='), -1);
  const full = enter('+1'.repeat(127), negative);
  assert.ok(enter('=', full).request);
  assert.match(enter('+', full).error, /finish|limit/);
  const longRight = complete(enter('1−(1' + '+1'.repeat(127) + ')='), -127);
  const repeat = enter('=', longRight);
  assert.equal(repeat.request, null);
  assert.match(repeat.error, /128|limit/);
});
