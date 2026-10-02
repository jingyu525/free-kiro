// sse.ts — EventSource wrapper that:
//   1. translates the browser's EventSource events into our store
//      states (connection.status + summary refetch trigger)
//   2. enforces a 30s heartbeat watchdog so a half-open connection
//      (TCP up but no events flowing) is force-closed and re-opened
//
// Reconnect strategy: we let the browser's native EventSource
// handle backoff (it retries every ~1-3s on close). We only force a
// close + reopen when lastEventAt is stale, because some proxies
// silently drop idle connections after 60s without sending RST.

import { connectionStore } from '../stores/connection';
import { summaryStore } from '../stores/summary';

const HEARTBEAT_STALE_MS = 30_000;
const HEARTBEAT_PING_TIMEOUT = setInterval;

export function connectSSE(onRefresh: () => void): () => void {
  const EventSourceCtor = window.EventSource;
  if (typeof EventSourceCtor !== 'function') {
    connectionStore.update((s) => ({ ...s, state: 'offline', attempt: s.attempt + 1 }));
    console.warn('EventSource unavailable; falling back to polling only');
    return () => {};
  }

  let es: EventSource | null = new EventSourceCtor('/api/events');
  let closed = false;
  let lastEventAt = Date.now();

  // Heartbeat watchdog — runs every 5s. If we haven't seen any event
  // for 30s (heartbeat interval is 15s server-side), the connection is
  // considered dead even if EventSource.onerror hasn't fired.
  const watchdog = setInterval(() => {
    const stale = Date.now() - lastEventAt;
    if (stale > HEARTBEAT_STALE_MS && connectionStore.get().state !== 'offline') {
      connectionStore.update((s) => ({ ...s, state: 'offline' }));
      es?.close();
      es = new EventSourceCtor('/api/events');
      connectionStore.update((s) => ({ ...s, state: 'reconnecting', attempt: s.attempt + 1 }));
      lastEventAt = Date.now();
    }
  }, 5_000);

  // Suppress unused-import warnings; HEARTBEAT_PING_TIMEOUT is
  // referenced for clarity even though Node doesn't ship setInterval
  // as a named binding in every TS lib version.
  void HEARTBEAT_PING_TIMEOUT;

  es.addEventListener('open', () => {
    connectionStore.update((s) => ({ ...s, state: 'online', attempt: 0 }));
    lastEventAt = Date.now();
  });

  es.addEventListener('ping', () => {
    lastEventAt = Date.now();
    connectionStore.update((s) => ({ ...s, lastEventAt: lastEventAt }));
  });

  es.addEventListener('heartbeat', () => {
    lastEventAt = Date.now();
    connectionStore.update((s) => ({ ...s, lastEventAt: lastEventAt }));
  });

  es.addEventListener('refresh', () => {
    lastEventAt = Date.now();
    connectionStore.update((s) => ({ ...s, lastEventAt: lastEventAt }));
    onRefresh();
  });

  es.addEventListener('error', () => {
    // EventSource auto-reconnects; we just record the attempt count
    // so the UI's "Reconnecting (attempt N)" tooltip stays accurate.
    if (closed) return;
    connectionStore.update((s) => ({ ...s, state: 'reconnecting', attempt: s.attempt + 1 }));
  });

  return () => {
    closed = true;
    clearInterval(watchdog);
    es?.close();
    es = null;
  };
}

// Surface the current loadState for Playwright assertions. Test-only.
export function _debugSummaryState(): unknown {
  return summaryStore.get();
}
