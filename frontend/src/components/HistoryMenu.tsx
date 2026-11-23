import { useEffect, useRef, useState } from 'react';
import type { Calculation } from '../types/calculator';

export function HistoryMenu({ history }: { history: Calculation[] }) {
  const [open, setOpen] = useState(false);
  const menu = useRef<HTMLDivElement>(null);
  const button = useRef<HTMLButtonElement>(null);

  useEffect(() => {
    if (!open) return;
    function dismissOutside(event: PointerEvent) {
      if (event.target instanceof Node && !menu.current?.contains(event.target)) setOpen(false);
    }
    function dismissOnEscape(event: KeyboardEvent) {
      if (event.key !== 'Escape') return;
      event.preventDefault();
      setOpen(false);
      button.current?.focus();
    }
    document.addEventListener('pointerdown', dismissOutside);
    document.addEventListener('keydown', dismissOnEscape);
    return () => {
      document.removeEventListener('pointerdown', dismissOutside);
      document.removeEventListener('keydown', dismissOnEscape);
    };
  }, [open]);

  return (
    <div className="history-menu" ref={menu}>
      <button
        type="button"
        className="history-toggle"
        ref={button}
        aria-label="History"
        aria-expanded={open}
        aria-controls="recent-history"
        onClick={() => setOpen(previous => !previous)}
        onKeyDown={event => {
          // Enter activates this button through the browser's native click.
          if (event.key === 'Enter') event.stopPropagation();
        }}
      >
        <span className="history-icon" aria-hidden="true" />
      </button>
      <aside id="recent-history" aria-labelledby="history-title" hidden={!open}>
        <h2 id="history-title">Recent history</h2>
        {history.length === 0 ? <p>No calculations yet.</p> : (
          <ol>
            {history.map((row, index) => (
              <li key={index}><span>{row.expression}</span><b>{row.value}</b></li>
            ))}
          </ol>
        )}
      </aside>
    </div>
  );
}
