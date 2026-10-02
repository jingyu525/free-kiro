// shared/config — base URLs injected by Vite at build time.
// In dev, dashboard calls `free-kiro serve` on :7374 (or VITE_DEV_API_BASE override).
// In production, the dashboard is served from the same origin under /api.

const env = import.meta.env;

export const API_BASE: string =
  env.VITE_API_BASE ?? (env.DEV ? 'http://127.0.0.1:7374/api' : '/api');

export const ASSET_BASE: string = env.VITE_ASSET_BASE ?? '/assets/';

export const SSE_URL: string = `${API_BASE}/events`;