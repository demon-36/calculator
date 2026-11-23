import type { CalculationBody } from '../types/calculator';

export async function calculate(body: CalculationBody, signal: AbortSignal): Promise<number> {
  const response = await fetch('/api/calculate', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
    signal,
  }).catch(() => {
    throw new Error('Cannot reach calculator. Press = to retry.');
  });
  const reply = await response.json().catch(() => null);
  if (!response.ok) {
    const message = response.status < 500 && typeof reply?.error?.message === 'string'
      ? reply.error.message : 'Service unavailable. Press = to retry.';
    throw new Error(message);
  }
  if (!Number.isFinite(reply?.result)) throw new Error('Invalid response. Press = to retry.');
  return reply.result;
}
