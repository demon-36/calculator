export const keys = [
  'AC', '⌫', '(', ')',
  '^', '√', '%', '÷',
  '7', '8', '9', '×',
  '4', '5', '6', '−',
  '1', '2', '3', '+',
  '0', '.', '=',
];

export const labels: Record<string, string> = {
  AC: 'All clear', '⌫': 'Backspace', '÷': 'Divide', '×': 'Multiply',
  '−': 'Subtract', '+': 'Add', '.': 'Decimal point', '=': 'Equals',
  '(': 'Open parenthesis', ')': 'Close parenthesis', '^': 'Power', '√': 'Square root', '%': 'Percent',
};

export const shortcuts: Record<string, string> = {
  Escape: 'AC', Backspace: '⌫', Enter: '=', '/': '÷', '*': '×', '-': '−', r: '√', R: '√',
};
