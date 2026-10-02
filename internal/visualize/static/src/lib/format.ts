// format.ts — small pure functions for human-readable formatting.
// Kept side-effect free so unit tests don't need a DOM.

export function formatRelative(ts: number, now: number = Date.now()): string {
  const seconds = Math.max(0, Math.floor((now - ts) / 1000));
  if (seconds < 5) return 'just now';
  if (seconds < 60) return `${seconds}s ago`;
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes}m ago`;
  const hours = Math.floor(minutes / 60);
  return `${hours}h ago`;
}

export function formatTimeHHMMSS(date: Date): string {
  const pad = (n: number) => n.toString().padStart(2, '0');
  return `${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`;
}

export function pluralize(n: number, singular: string, plural?: string): string {
  if (n === 1) return `${n} ${singular}`;
  return `${n} ${plural ?? singular + 's'}`;
}

// Compact integer with k/M suffix. Used in stat cards where
// "1234 specs" takes more space than the card has.
export function formatCompact(n: number): string {
  if (n < 1000) return n.toString();
  if (n < 1_000_000) return `${(n / 1000).toFixed(1).replace(/\.0$/, '')}k`;
  return `${(n / 1_000_000).toFixed(1).replace(/\.0$/, '')}M`;
}
