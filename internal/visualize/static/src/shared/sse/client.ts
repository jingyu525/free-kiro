// shared/sse/client — EventSource wrapper.
//
// Subscribes to /api/events and emits four normalized signals: 'open',
// 'refresh', 'heartbeat', 'error'. Reconnects on error with backoff and
// force-closes after 30s of silence to break stuck connections.
//
// Preserves dashboard-sse-bugfix event names: server emits 'ping' on
// connect and 'refresh' on file change; we map ping→heartbeat for
// liveness tracking.

import { SSE_URL } from '@shared/config';

export type SSEEventName = 'open' | 'refresh' | 'heartbeat' | 'error';

export interface SSEEvent {
  name: SSEEventName;
  ts: number;
}

export type SSEListener = (e: SSEEvent) => void;

const SILENCE_MS = 30_000;

export class SSEClient {
  private es: EventSource | null = null;
  private listeners = new Set<SSEListener>();
  private attempt = 0;
  private lastEventAt = 0;
  private silenceTimer: ReturnType<typeof setInterval> | null = null;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;

  connect(): void {
    if (this.es) return;
    this.es = new EventSource(SSE_URL, { withCredentials: false });
    this.es.addEventListener('open', () => {
      this.attempt = 0;
      this.lastEventAt = Date.now();
      this.emit('open');
      this.startSilenceWatch();
    });
    this.es.addEventListener('refresh', () => {
      this.lastEventAt = Date.now();
      this.emit('refresh');
    });
    // 'ping' is sent on connect by the server (dashboard-sse-bugfix).
    // We treat it as a heartbeat to drive liveness tracking.
    this.es.addEventListener('ping', () => {
      this.lastEventAt = Date.now();
      this.emit('heartbeat');
    });
    this.es.addEventListener('error', () => {
      this.attempt++;
      this.emit('error');
      // EventSource auto-reconnects, but we cap the perceived attempt
      // by closing after one cycle so we can re-establish a clean socket.
      this.scheduleReconnect();
    });
  }

  disconnect(): void {
    this.stopSilenceWatch();
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    if (this.es) {
      this.es.close();
      this.es = null;
    }
  }

  subscribe(fn: SSEListener): () => void {
    this.listeners.add(fn);
    return () => this.listeners.delete(fn);
  }

  getAttempt(): number {
    return this.attempt;
  }

  getLastEventAt(): number {
    return this.lastEventAt;
  }

  private emit(name: SSEEventName): void {
    const event: SSEEvent = { name, ts: Date.now() };
    for (const fn of this.listeners) {
      try {
        fn(event);
      } catch (e) {
        // Listeners must not throw; we swallow to keep SSE alive but log.
        console.error('[SSEClient] listener threw', e);
      }
    }
  }

  private startSilenceWatch(): void {
    this.stopSilenceWatch();
    this.silenceTimer = setInterval(() => {
      if (!this.lastEventAt) return;
      if (Date.now() - this.lastEventAt > SILENCE_MS) {
        // 30s without heartbeat/refresh → force reconnect.
        if (this.es) {
          this.es.close();
          this.es = null;
        }
        this.scheduleReconnect();
      }
    }, 5000);
  }

  private stopSilenceWatch(): void {
    if (this.silenceTimer) {
      clearInterval(this.silenceTimer);
      this.silenceTimer = null;
    }
  }

  private scheduleReconnect(): void {
    if (this.reconnectTimer) return;
    const delay = Math.min(1000 * Math.pow(2, this.attempt), 8000);
    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null;
      this.connect();
    }, delay);
  }
}