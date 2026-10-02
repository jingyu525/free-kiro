// main.ts — dashboard bootstrap. Mounts <App> into #app and wires
// the global SSE listener that drives the summary store.
//
// Kept tiny on purpose: anything stateful lives in src/stores/ or
// src/lib/, anything visual lives in src/components/.

import './styles/index.css';

import { App } from './components/app';
import { summaryStore, refreshSummary } from './stores/summary';
import { connectSSE } from './lib/sse';
import { applyStoredTheme, initTheme } from './lib/theme';

// Theme must apply before App mounts so the first paint uses the
// correct colour scheme (no flash of dark-on-light / light-on-dark).
applyStoredTheme();

const root = document.getElementById('app');
if (!root) {
  throw new Error('missing <div id="app"> in index.html');
}
root.appendChild(App());

// Theme toggle button lives in the HeaderBar; it dispatches a custom
// event the theme store listens to. One-shot wiring at startup is
// enough because the toggle button is mounted synchronously above.
initTheme();

// Wire SSE → summary store. The SSE client lives for the page
// lifetime; teardown on navigation is unnecessary (single-page app).
connectSSE(() => refreshSummary());

// Kick off the initial fetch (SSE ping arrives within 2s but the
// first paint shouldn't wait).
refreshSummary();

// Log to the console so Playwright E2E can assert the dashboard
// reached steady state without poking internal state.
window.addEventListener('load', () => {
  console.log('free-kiro dashboard mounted');
  // Expose the summary store for E2E assertions. Same data the
  // UI subscribes to — keeps tests honest about what the user sees.
  (window as unknown as { __summaryStore: typeof summaryStore }).__summaryStore = summaryStore;
});
