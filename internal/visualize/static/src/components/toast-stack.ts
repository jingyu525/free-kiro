// components/toast-stack.ts — transient notifications for API
// failures. MVP placeholder; the future dashboard-frontend-components
// spec wires this to the summary store's lastError.

import type { ApiError } from '../types';

export function renderErrorBanner(onRetry: () => void): HTMLElement {
  const banner = document.createElement('div');
  banner.className = 'fk-error-banner';
  banner.setAttribute('role', 'alert');

  const text = document.createElement('span');
  text.textContent = "Couldn't refresh dashboard.";
  banner.appendChild(text);

  const btn = document.createElement('button');
  btn.type = 'button';
  btn.className = 'fk-error-banner__retry';
  btn.textContent = 'Retry';
  btn.addEventListener('click', onRetry);
  banner.appendChild(btn);

  return banner;
}

// Toast placeholder: the MVP hooks the slot in App but does not
// dispatch any toasts. The store's lastError is surfaced via the
// banner above; future work can route through here for non-blocking
// notifications.
export function renderToastStack(_error: ApiError | null): HTMLElement {
  const wrap = document.createElement('div');
  wrap.className = 'fk-toast-stack';
  return wrap;
}
