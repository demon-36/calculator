import { keys, labels } from '../constants/keypad.js';
import { selectedOperator } from '../calculator.js';
import type { State } from '../types/calculator.js';

interface KeypadProps {
  state: State;
  onKey: (key: string) => void;
}

export function Keypad({ state, onKey }: KeypadProps) {
  return (
    <div className="keys" role="group" aria-label="Keypad">
      {keys.map(key => (
        <button
          type="button"
          key={key}
          data-key={key}
          aria-label={labels[key] ?? key}
          disabled={Boolean(state.request) && key !== 'AC'}
          data-selected={selectedOperator(state) === key}
          onClick={() => onKey(key)}
        >
          {key === '^' ? 'xʸ' : key}
        </button>
      ))}
    </div>
  );
}
