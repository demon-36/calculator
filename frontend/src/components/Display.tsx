import type { CSSProperties } from 'react';
import type { State } from '../types/calculator.js';
import { displayedExpression } from '../calculator.js';

interface DisplayProps {
  state: State;
  display: string;
}

export function Display({ state, display }: DisplayProps) {
  return (
    <div className="screen" data-error={Boolean(state.error)}>
      <div className="expression">{displayedExpression(state)}</div>
      <output aria-label="Result" style={{ '--digits': display.length } as CSSProperties}>
        {display}
      </output>
    </div>
  );
}
