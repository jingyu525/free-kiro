// components/progress-bar.ts — slim indeterminate bar shown in the
// header region during background refreshes. Keeps the previous
// data on screen (no skeleton) so the user sees consistent content.

export function renderProgressBar(): HTMLElement {
  const bar = document.createElement('div');
  bar.className = 'fk-progress-bar';
  bar.setAttribute('role', 'progressbar');
  bar.setAttribute('aria-label', 'Refreshing dashboard');
  const inner = document.createElement('div');
  inner.className = 'fk-progress-bar__inner';
  bar.appendChild(inner);
  return bar;
}
