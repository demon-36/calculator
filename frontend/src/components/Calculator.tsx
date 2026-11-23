import { useEffect } from 'react';
import { HistoryMenu } from './HistoryMenu.js';
import { Display } from './Display.js';
import { Keypad } from './Keypad.js';
import { useCalculator } from '../hooks/useCalculator.js';
import { keys, shortcuts } from '../constants/keypad.js';

export function Calculator() {
  const [state, dispatch] = useCalculator();

  useEffect(() => {
    function handleKey(event: KeyboardEvent) {
      if (event.defaultPrevented || event.metaKey || event.ctrlKey || event.altKey || event.isComposing) return;
      const key = shortcuts[event.key] ?? event.key;
      if (keys.includes(key)) {
        event.preventDefault();
        dispatch(key);
      }
    }
    window.addEventListener('keydown', handleKey);
    return () => window.removeEventListener('keydown', handleKey);
  }, [dispatch]);

  const display = state.error || (state.request ? '…' : state.value);

  return (
    <section className="pad" aria-labelledby="calculator-title">
      <header className="instrument-header">
        <h1 id="calculator-title">Calculator</h1>
        <span className="solar-strip" aria-hidden="true" />
        <HistoryMenu history={state.history} />
      </header>
      <Display state={state} display={display} />
      <Keypad state={state} onKey={dispatch} />
    </section>
  );
}
