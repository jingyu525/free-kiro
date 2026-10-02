// shared/lib/format — pure formatters used by widgets.

export function formatRelative(ts: number, now: number = Date.now()): string {
  const delta = Math.max(0, Math.floor((now - ts) / 1000));
  if (delta < 5) return 'just now';
  if (delta < 60) return `${delta}s ago`;
  const mins = Math.floor(delta / 60);
  if (mins < 60) return `${mins}m ago`;
  const hours = Math.floor(mins / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.floor(hours / 24);
  return `${days}d ago`;
}

export function formatDelta(n: number): string {
  if (n === 0) return '±0';
  const sign = n > 0 ? '+' : '';
  return `${sign}${n}`;
}

export function formatTimeHHMMSS(d: Date | string): string {
  const date = typeof d === 'string' ? new Date(d) : d;
  const pad = (n: number): string => n.toString().padStart(2, '0');
  return `${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`;
}

export function formatNumber(n: number): string {
  if (n >= 1000) return `${(n / 1000).toFixed(1)}k`;
  return String(n);
}