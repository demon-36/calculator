// Each number and each keypad action occupies one token. A root key opens its group.
const binaryOperators = new Set(['+', '−', '×', '÷', '^']);
const maxOperations = 128;
const maxDepth = 32;
const maxDigits = 15;

export function isOperator(token = ''): boolean {
  return binaryOperators.has(token);
}

export function isNumber(token = ''): boolean {
  return /^-?\d/.test(token);
}

function isOpen(token: string): boolean {
  return token === '(' || token === '√(';
}

function depthChange(token: string): number {
  if (isOpen(token)) return 1;
  return token === ')' ? -1 : 0;
}

function isOperandEnd(token = ''): boolean {
  return isNumber(token) || token === ')' || token === '%';
}

export function format(tokens: string[]): string {
  return tokens.map(token => {
    if (isOperator(token)) return ' ' + token + ' ';
    if (token === 'negate') return '−';
    // A confirmed negative result is one operand, even when raised to a power.
    if (token.startsWith('-')) return '(−' + token.slice(1) + ')';
    return token;
  }).join('').trim();
}

export function currentEntry(tokens: string[]): string {
  for (let i = tokens.length - 1; i >= 0; i--) {
    if (isOpen(tokens[i])) break;
    if (isNumber(tokens[i])) return tokens[i - 1] === 'negate' ? '−' + tokens[i] : tokens[i];
  }
  return '0';
}

function groupDepth(tokens: string[]): number {
  return tokens.reduce((depth, token) => depth + depthChange(token), 0);
}

function endsInZeroDivisor(tokens: string[]): boolean {
  return tokens.at(-2) === '÷' && Number(tokens.at(-1)) === 0;
}

function operationCount(tokens: string[]): number {
  return tokens.filter(token =>
    isOperator(token) || token === 'negate' || token === '%' || token === '√(' || token.startsWith('-')
  ).length;
}

export function submissionError(tokens: string[]): string {
  if (!isOperandEnd(tokens.at(-1))) return 'Enter a number or operand.';
  if (groupDepth(tokens) > 0) return 'Close the open parenthesis.';
  if (operationCount(tokens) > maxOperations) return 'The limit is 128 operations.';
  return endsInZeroDivisor(tokens) ? 'Cannot divide by zero.' : '';
}

// Keep the last operator entered outside a group and everything to its right.
// Whole-expression parentheses can be peeled away after this level has no operator.
export function repeatOperation(tokens: string[]): string[] | null {
  let depth = 0;
  for (let i = tokens.length - 1; i >= 0; i--) {
    const token = tokens[i];
    if (depth === 0 && isOperator(token)) return tokens.slice(i);
    depth -= depthChange(token);
  }
  if (tokens[0] === '(' && tokens.at(-1) === ')') return repeatOperation(tokens.slice(1, -1));
  return null;
}

function append(tokens: string[], token: string): string[] | string {
  if (operationCount(tokens) >= maxOperations) return 'Press = to finish this expression.';
  return [...tokens, token];
}

export function edit(tokens: string[], key: string): string[] | string {
  const last = tokens.at(-1) ?? '';
  if (key === '⌫') {
    const shorter = tokens.slice(0, -1);
    if (isNumber(last) && last.length > 1 && !last.includes('e') && last.slice(0, -1) !== '-') {
      shorter.push(last.slice(0, -1));
    }
    return shorter;
  }
  if (/^[0-9.]$/.test(key)) {
    if (last === ')' || last === '%') return tokens;
    const previous = isNumber(last) && !last.includes('e') ? last : '0';
    if (key === '.' && previous.includes('.')) return tokens;
    const number = previous === '0' && key !== '.' ? key : previous + key;
    if (number.replace(/\D/g, '').length > maxDigits) return tokens;
    return [...(isNumber(last) ? tokens.slice(0, -1) : tokens), number];
  }
  if (key === '(' || key === '√') {
    if (isOperandEnd(last)) return tokens;
    if (groupDepth(tokens) >= maxDepth) return 'Too many nested groups.';
    return key === '√' ? append(tokens, '√(') : [...tokens, '('];
  }
  if (key === ')') {
    return isOperandEnd(last) && groupDepth(tokens) > 0 ? [...tokens, ')'] : tokens;
  }
  if (key === '%') return isOperandEnd(last) ? append(tokens, '%') : tokens;
  if (!isOperator(key)) return tokens;
  if (isOperator(last)) return [...tokens.slice(0, -1), key];
  if (!isOperandEnd(last)) {
    if (key === '−' && last !== 'negate') return append(tokens, 'negate');
    return tokens.length ? tokens : ['0', key];
  }
  if (key !== '^' && endsInZeroDivisor(tokens)) return 'Cannot divide by zero.';
  return append(tokens, key);
}
