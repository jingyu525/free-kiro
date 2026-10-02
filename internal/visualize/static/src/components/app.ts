// components/app.ts — top-level layout. Subscribes to all stores and
// re-renders the children on change. No virtual DOM; we just rebuild
// the relevant sub-tree.

import { HeaderBar } from './header-bar';
import { SummaryGrid } from './summary-grid';
import { SpecsTable } from './specs-table';
import { EmptyState } from './empty-state';
import { summaryStore } from '../stores/summary';
import { connectionStore } from '../stores/connection';
import { renderErrorBanner, renderToastStack } from './toast-stack';
import { renderLoadingSkeleton } from './skeleton-row';
import { renderProgressBar } from './progress-bar';

export function App(): HTMLElement {
  const root = document.createElement('main');
  root.id = 'fk-app';
  root.setAttribute('role', 'main');

  const headerSlot = document.createElement('div');
  const errorBannerSlot = document.createElement('div');
  const progressSlot = document.createElement('div');
  const summarySlot = document.createElement('section');
  summarySlot.setAttribute('aria-labelledby', 'specs-heading');
  const specsSlot = document.createElement('section');
  const toastSlot = document.createElement('div');
  toastSlot.id = 'fk-toasts';
  toastSlot.setAttribute('role', 'status');
  toastSlot.setAttribute('aria-live', 'polite');

  root.append(headerSlot, errorBannerSlot, progressSlot, summarySlot, specsSlot, toastSlot);

  // Renders the header bar once and keeps it in sync with the
  // summary + connection + theme stores.
  function rerenderHeader(): void {
    headerSlot.replaceChildren(
      HeaderBar({
        specCount: summaryStore.get().value?.specs.length ?? 0,
        lastRefreshAt: summaryStore.get().lastRefreshAt,
        connection: connectionStore.get(),
      }),
    );
  }

  function rerenderBody(): void {
    try {
      rerenderBodyInner();
    } catch (e) {
      console.error('[fk] rerenderBody failed:', e, 'value:', summaryStore.get().value);
    }
  }

  function rerenderBodyInner(): void {
    const { value, loadState } = summaryStore.get();

    // Error banner: only when in error state AND we have a previous
    // value to keep showing (otherwise the EmptyState handles it).
    if (loadState === 'error' && value) {
      errorBannerSlot.replaceChildren(renderErrorBanner(() => {
        import('../stores/summary').then((m) => m.refreshSummary());
      }));
    } else {
      errorBannerSlot.replaceChildren();
    }

    // Top progress bar: visible when refreshing after initial load.
    if (loadState === 'loading' && value) {
      progressSlot.replaceChildren(renderProgressBar());
    } else {
      progressSlot.replaceChildren();
    }

    if (!value) {
      // First load — show skeleton.
      specsSlot.replaceChildren();
      summarySlot.replaceChildren();
      const skel = document.createElement('div');
      skel.className = 'fk-skeleton-region';
      skel.setAttribute('aria-busy', 'true');
      skel.appendChild(renderLoadingSkeleton());
      specsSlot.appendChild(skel);
      return;
    }

    const mode = value.mode ?? 'ok';
    if (mode !== 'ok' || value.specs.length === 0) {
      summarySlot.replaceChildren();
      specsSlot.replaceChildren(EmptyState({ mode, specs: value.specs.length }));
      return;
    }

    summarySlot.replaceChildren(SummaryGrid({ specs: value.specs, active: value.active }));
    specsSlot.replaceChildren(SpecsTable({ specs: value.specs, active: value.active }));
  }

  function rerenderToasts(): void {
    toastSlot.replaceChildren();
    // ToastStack renders into toastSlot directly; placeholder for now.
    void renderToastStack;
  }

  summaryStore.subscribe(rerenderHeader);
  summaryStore.subscribe(rerenderBody);
  connectionStore.subscribe(rerenderHeader);
  // Periodic refresh of the relative-time label (every 1s) without
  // re-running the whole body — only the header needs the new
  // relative-time string.
  setInterval(rerenderHeader, 1_000);

  rerenderHeader();
  rerenderBody();
  rerenderToasts();

  return root;
}
