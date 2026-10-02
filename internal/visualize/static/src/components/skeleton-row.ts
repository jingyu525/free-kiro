// components/skeleton-row.ts — placeholder rows shown during the
// initial fetch. Pure DOM, no animations here (CSS handles the
// shimmer via background-position keyframes + a media query that
// disables it under prefers-reduced-motion).

export function renderLoadingSkeleton(): HTMLElement {
  const wrap = document.createElement('div');
  wrap.className = 'fk-skeleton-table';
  wrap.setAttribute('aria-hidden', 'true');

  for (let i = 0; i < 6; i++) {
    const row = document.createElement('div');
    row.className = 'fk-skeleton-row';
    // Three bars per row mimicking the 5 columns (name + 4 shorter).
    row.appendChild(makeBar('60%'));
    row.appendChild(makeBar('20%'));
    row.appendChild(makeBar('12%'));
    row.appendChild(makeBar('16%'));
    row.appendChild(makeBar('14%'));
    wrap.appendChild(row);
  }

  return wrap;
}

function makeBar(width: string): HTMLElement {
  const bar = document.createElement('span');
  bar.className = 'fk-skeleton-bar';
  bar.style.width = width;
  return bar;
}
