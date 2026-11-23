import { currentEntry, edit, format, isNumber, isOperator, repeatOperation, submissionError } from './editor.js';
import type { Calculation, CalculationBody, Reply, State, Submission } from './types/calculator.js';

export type { Calculation, CalculationBody, Reply, State, Submission };

export const initialState: State = {
  tokens: [], value: '0', finished: false, repeat: null, request: null, error: '', failedExpression: '', history: [],
};

export function displayedExpression(state: State): string {
  if (state.request) return state.request.body.expression;
  if (state.failedExpression) return state.failedExpression;
  if (state.finished) return state.history[0].expression + ' =';
  return format(state.tokens);
}

export function selectedOperator(state: State): string | undefined {
  const last = state.tokens.at(-1);
  return isOperator(last) ? last : undefined;
}

function submit(state: State): State {
  let tokens = state.tokens;
  if (state.finished || (tokens.length === 1 && isNumber(tokens[0]) && state.repeat)) {
    if (!state.repeat) return state;
    tokens = [state.value, ...state.repeat];
  }
  if (!tokens.length) return state;
  const error = submissionError(tokens);
  if (error) return { ...state, error };
  const expression = format(tokens);
  const request = { body: { expression }, repeat: repeatOperation(tokens) };
  return { ...state, request, error: '', failedExpression: '' };
}

function press(state: State, key: string): State {
  if (key === 'AC') return { ...initialState, history: state.history };
  if (state.request) return state;
  if (key === '=') return submit(state);
  if (state.finished && key === '⌫') return { ...initialState, history: state.history };

  let tokens = state.tokens;
  if (state.finished && (isOperator(key) || key === '%')) tokens = [state.value];
  const edited = edit(tokens, key);
  if (typeof edited === 'string') return { ...state, error: edited };
  if (edited === tokens) return state;
  const keepRepeat = /^[0-9.]$/.test(key) || key === '⌫';
  return {
    ...state, tokens: edited, value: currentEntry(edited), finished: false,
    repeat: keepRepeat ? state.repeat : null, error: '', failedExpression: '',
  };
}

export function reducer(state: State, action: string | Reply): State {
  if (typeof action === 'string') return press(state, action);
  if (state.request !== action.request) return state;
  if (action.type === 'failure') {
    return { ...state, error: action.message, failedExpression: action.request.body.expression, request: null };
  }
  const row = { expression: action.request.body.expression, value: String(action.result) };
  return {
    ...initialState, finished: true, value: row.value, repeat: action.request.repeat,
    history: [row, ...state.history].slice(0, 3),
  };
}
