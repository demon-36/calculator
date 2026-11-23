import { useEffect, useReducer } from 'react';
import { calculate } from '../services/api.js';
import { initialState, reducer } from '../calculator.js';

export function useCalculator() {
  const calculator = useReducer(reducer, initialState);
  const [state, dispatch] = calculator;
  const request = state.request;
  useEffect(() => {
    if (!request) return;
    const controller = new AbortController();
    const timeout = window.setTimeout(() => {
      controller.abort();
      dispatch({ type: 'failure', request, message: 'Request timed out. Press = to retry.' });
    }, 8000);
    calculate(request.body, controller.signal)
      .then(result => {
        if (!controller.signal.aborted) dispatch({ type: 'success', request, result });
      })
      .catch((error: Error) => {
        if (!controller.signal.aborted) dispatch({ type: 'failure', request, message: error.message });
      })
      .finally(() => window.clearTimeout(timeout));
    return () => {
      window.clearTimeout(timeout);
      controller.abort();
    };
  }, [request]);

  return calculator;
}
