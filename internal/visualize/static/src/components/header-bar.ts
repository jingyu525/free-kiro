// components/header-bar.ts — title + meta + refresh + theme toggle
// + connection status indicator.

import { formatRelative, formatTimeHHMMSS } from '../lib/format';
import { themeStore } from '../stores/theme';
import { summaryStore, refreshSummary } from '../stores/summary';
import type { ConnectionStateValue } from '../stores/connection';

interface HeaderBarProps {
  specCount: number;
  lastRefreshAt: number;
  connection: ConnectionStateValue;
}

const REFRESH_DEBOUNCE_MS = 500;
let lastClickAt = 0;

export function HeaderBar(props: HeaderBarProps): HTMLElement {
  const header = document.createElement('header');
  header.className = 'fk-header';
  header.setAttribute('role', 'banner');

  // Title row.
  const titleRow = document.createElement('div');
  titleRow.className = 'fk-header__title-row';

  const title = document.createElement('h1');
  title.textContent = 'free-kiro dashboard';
  title.className = 'fk-header__title';
  titleRow.appendChild(title);

  titleRow.appendChild(renderThemeToggle());
  titleRow.appendChild(renderConnectionDot(props.connection));

  header.appendChild(titleRow);

  // Meta row: spec count + relative-time + HH:MM:SS + refresh button.
  const metaRow = document.createElement('div');
  metaRow.className = 'fk-header__meta-row';
  metaRow.id = 'fk-meta';

  const metaText = document.createElement('span');
  metaText.className = 'fk-header__meta-text';
  metaText.textContent = renderMetaText(props);
  metaRow.appendChild(metaText);

  metaRow.appendChild(renderRefreshButton());

  header.appendChild(metaRow);

  // Ctrl+R / Cmd+R interceptor — let the user hit the browser refresh
  // shortcut to trigger a data refresh without reloading the page.
  header.addEventListener('keydown', (e) => {
    if ((e.metaKey || e.ctrlKey) && e.key === 'r') {
      e.preventDefault();
      scheduleRefresh();
    }
  });

  return header;
}

function renderMetaText(props: HeaderBarProps): string {
  const ts = props.lastRefreshAt;
  const parts: string[] = [];
  parts.push(`${props.specCount} specs`);
  if (ts > 0) {
    parts.push(formatRelative(ts));
    parts.push(`generated ${formatTimeHHMMSS(new Date(ts))}`);
  } else {
    parts.push('loading…');
  }
  // Append "still refreshing…" if load has been pending > 3s.
  const startedAt = summaryStore.get().loadStartedAt;
  if (startedAt > 0 && summaryStore.get().loadState === 'loading' && Date.now() - startedAt > 3000) {
    parts.push('still refreshing…');
  }
  return parts.join(' · ');
}

function renderRefreshButton(): HTMLButtonElement {
  const btn = document.createElement('button');
  btn.type = 'button';
  btn.className = 'fk-header__refresh';
  btn.id = 'fk-refresh';
  btn.setAttribute('aria-label', 'Refresh dashboard');
  btn.textContent = '↻ Refresh';
  btn.addEventListener('click', scheduleRefresh);
  return btn;
}

function scheduleRefresh(): void {
  const now = Date.now();
  if (now - lastClickAt < REFRESH_DEBOUNCE_MS) return;
  lastClickAt = now;
  refreshSummary();
}

function renderThemeToggle(): HTMLButtonElement {
  const btn = document.createElement('button');
  btn.type = 'button';
  btn.className = 'fk-header__theme-toggle';
  btn.id = 'fk-theme-toggle';
  btn.setAttribute('aria-label', 'Toggle theme');
  // Symbol switches between sun/moon based on current theme.
  const update = () => {
    const cur = themeStore.get();
    const eff = cur === 'auto'
      ? (window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light')
      : cur;
    btn.textContent = eff === 'dark' ? '☀' : '☾';
  };
  themeStore.subscribe(update);
  update();
  btn.addEventListener('click', () => {
    window.dispatchEvent(new Event('theme:toggle'));
  });
  return btn;
}

function renderConnectionDot(conn: ConnectionStateValue): HTMLElement {
  const wrap = document.createElement('span');
  wrap.className = 'fk-header__conn';
  wrap.id = 'fk-conn';

  const dot = document.createElement('span');
  dot.className = 'fk-conn-dot';
  dot.dataset['state'] = conn.state;
  dot.setAttribute('role', 'status');
  dot.setAttribute('aria-label', `SSE ${conn.state}`);
  wrap.appendChild(dot);

  const label = document.createElement('span');
  label.className = 'fk-conn-label';
  label.textContent = connectionLabel(conn);
  wrap.appendChild(label);

  return wrap;
}

function connectionLabel(conn: ConnectionStateValue): string {
  switch (conn.state) {
    case 'online': return 'Live';
    case 'reconnecting': return `Reconnecting (attempt ${conn.attempt})…`;
    case 'offline': return 'Offline — polling fallback';
  }
}
