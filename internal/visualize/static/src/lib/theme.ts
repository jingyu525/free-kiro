// theme.ts — dark/light/auto theme application. The toggle button
// in HeaderBar dispatches a 'theme:toggle' event; this module listens
// and updates <html data-theme> + localStorage.

import { themeStore } from '../stores/theme';

const STORAGE_KEY = 'fk-theme';

export function applyStoredTheme(): void {
  const stored = localStorage.getItem(STORAGE_KEY) as 'dark' | 'light' | 'auto' | null;
  const initial = stored ?? 'auto';
  themeStore.set(initial);
  apply(initial);
}

// initTheme wires the global 'theme:toggle' event. Called once from
// main.ts after App mounts (the toggle button lives inside App).
export function initTheme(): void {
  window.addEventListener('theme:toggle', () => {
    const cur = themeStore.get();
    // cycle: dark → light → auto → dark
    const next: 'dark' | 'light' | 'auto' =
      cur === 'dark' ? 'light' : cur === 'light' ? 'auto' : 'dark';
    themeStore.set(next);
    localStorage.setItem(STORAGE_KEY, next);
    apply(next);
  });

  // React to system theme changes while in 'auto' mode.
  const mql = window.matchMedia('(prefers-color-scheme: dark)');
  mql.addEventListener('change', () => {
    if (themeStore.get() === 'auto') apply('auto');
  });
}

function apply(value: 'dark' | 'light' | 'auto'): void {
  const html = document.documentElement;
  const effective = value === 'auto'
    ? (window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light')
    : value;
  html.dataset['theme'] = effective;
}
