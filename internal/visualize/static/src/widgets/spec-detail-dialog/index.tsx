// SpecDetailDialog — modal dialog driven by hash route (spec/<name>?tab=…).
// Renders 4 tabs: overview / drift / tasks / timeline with focus management,
// URL sync, prev/next navigation, copy-link, and roving tabindex.

import { useEffect, useMemo, useRef } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import {
  useHashRoute,
  navigateToHash,
  type TabId,
} from '@shared/lib/hash-router';
import { announce } from '@shared/lib/a11y/announce';
import { SUMMARY_QUERY_KEY } from '@entities/summary/use-summary-query';
import { OverviewTab } from './tabs/overview';
import { DriftTab } from './tabs/drift';
import { TasksTab } from './tabs/tasks';
import { TimelineTab } from './tabs/timeline';

const TABS: ReadonlyArray<TabId> = ['overview', 'drift', 'tasks', 'timeline'];

export function SpecDetailDialog(): JSX.Element | null {
  const route = useHashRoute();
  const qc = useQueryClient();
  // Read cached summary without subscribing — avoids react-query's
  // useSyncExternalStore re-render path that, combined with our own
  // hash routing, can hit React 18 prod's "Maximum update depth" guard.
  const data = qc.getQueryData(SUMMARY_QUERY_KEY) as
    | { specs: Array<{ meta: { name: string } }> }
    | undefined;
  const isOpen = route.name !== '';
  const name = route.name;
  const tab = route.tab;

  const dialogRef = useRef<HTMLDialogElement>(null);
  const triggerRef = useRef<HTMLElement | null>(null);
  const tabListRef = useRef<HTMLDivElement>(null);
  // Track last seen (name, isOpen) so open/close effect fires only on
  // isOpen transitions, not on every render.
  const wasOpenRef = useRef<boolean>(false);
  // Track last announced tab so the tab-change effect does not re-announce
  // when the same tab is re-rendered for unrelated reasons.
  const lastTabRef = useRef<TabId | null>(null);

  // Open / close effect: keyed only on isOpen to prevent re-entry loops in
  // React 18 prod where StrictMode double-invocation can otherwise re-fire
  // the same setState chain.
  useEffect(() => {
    if (isOpen === wasOpenRef.current) return;
    wasOpenRef.current = isOpen;
    const dialog = dialogRef.current;
    if (!dialog) return;
    if (isOpen) {
      triggerRef.current = document.activeElement as HTMLElement;
      dialog.showModal();
      const focusEl =
        tabListRef.current?.querySelector<HTMLElement>('[role="tab"][aria-selected="true"]') ??
        tabListRef.current?.querySelector<HTMLElement>('[role="tab"]');
      focusEl?.focus();
      announce(`Spec ${name} details opened`);
      lastTabRef.current = tab;
    } else if (dialog.open) {
      dialog.close();
      triggerRef.current?.focus();
      announce('Spec details closed');
    }
    // name intentionally excluded — captured by `name` at call time.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isOpen]);

  // Tab change announce: separate effect, only fires when tab transitions
  // (and dialog is open). Independent of the open/close effect above.
  useEffect(() => {
    if (!isOpen) return;
    if (lastTabRef.current === tab) return;
    lastTabRef.current = tab;
    announce(`Tab changed to ${tab}`);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [tab, isOpen]);

  const onBackdropClick = (e: React.MouseEvent<HTMLDialogElement>): void => {
    if (e.target === dialogRef.current) {
      navigateToHash('');
    }
  };

  const onClose = (): void => {
    navigateToHash('');
  };

  const onTabKeyDown = (e: React.KeyboardEvent<HTMLDivElement>): void => {
    const target = e.target as HTMLElement;
    const tabEl = target.closest<HTMLButtonElement>('[role="tab"]');
    if (!tabEl) return;
    const currentIdx = TABS.indexOf((tabEl.dataset.tab as TabId) ?? 'overview');
    let nextIdx = currentIdx;
    if (e.key === 'ArrowRight') nextIdx = (currentIdx + 1) % TABS.length;
    else if (e.key === 'ArrowLeft') nextIdx = (currentIdx - 1 + TABS.length) % TABS.length;
    else if (e.key === 'Home') nextIdx = 0;
    else if (e.key === 'End') nextIdx = TABS.length - 1;
    else return;
    e.preventDefault();
    const nextTab = TABS[nextIdx]!;
    navigateToHash({ name, tab: nextTab });
    requestAnimationFrame(() => {
      const next = tabListRef.current?.querySelector<HTMLButtonElement>(
        `[role="tab"][data-tab="${nextTab}"]`,
      );
      next?.focus();
    });
  };

  const specs = data?.specs ?? [];
  const idx = useMemo(() => specs.findIndex((s) => s.meta.name === name), [specs, name]);
  const prevSpec = idx > 0 ? (specs[idx - 1] as { meta: { name: string } } | undefined)?.meta.name : null;
  const nextSpec = idx >= 0 && idx < specs.length - 1 ? (specs[idx + 1] as { meta: { name: string } } | undefined)?.meta.name : null;

  const goToSpec = (targetName: string): void => {
    navigateToHash({ name: targetName, tab });
  };

  const copyLink = async (): Promise<void> => {
    const url = window.location.href;
    try {
      await navigator.clipboard.writeText(url);
      announce('Link copied to clipboard');
    } catch {
      announce('Copy failed, link is in the URL bar', 'assertive');
    }
  };

  if (!isOpen) return null;

  return (
    <dialog
      ref={dialogRef}
      className="spec-detail-dialog"
      aria-labelledby={`dialog-title-${name}`}
      onClose={onClose}
      onClick={onBackdropClick}
    >
      <header className="spec-detail-dialog__header">
        <h2 id={`dialog-title-${name}`}>{name}</h2>
        <div role="toolbar" aria-label="Spec detail actions" className="dialog-toolbar">
          <button
            type="button"
            className="btn btn-ghost"
            aria-label="Previous spec"
            disabled={prevSpec === null}
            onClick={() => prevSpec && goToSpec(prevSpec)}
          >
            ←
          </button>
          <button
            type="button"
            className="btn btn-ghost"
            aria-label="Next spec"
            disabled={nextSpec === null}
            onClick={() => nextSpec && goToSpec(nextSpec)}
          >
            →
          </button>
          <button
            type="button"
            className="btn btn-ghost"
            aria-label="Copy deep link"
            onClick={() => void copyLink()}
          >
            🔗
          </button>
          <button
            type="button"
            className="btn btn-ghost"
            aria-label="Close"
            onClick={onClose}
          >
            ×
          </button>
        </div>
      </header>
      <nav
        ref={tabListRef}
        className="spec-detail-dialog__tabs"
        role="tablist"
        aria-label="Spec detail tabs"
        onKeyDown={onTabKeyDown}
      >
        {TABS.map((t) => (
          <button
            key={t}
            type="button"
            role="tab"
            id={`tab-${t}`}
            data-tab={t}
            aria-selected={tab === t}
            aria-controls={`panel-${t}`}
            tabIndex={tab === t ? 0 : -1}
            onClick={() => navigateToHash({ name, tab: t })}
          >
            {t}
          </button>
        ))}
      </nav>
      <section
        role="tabpanel"
        id={`panel-${tab}`}
        aria-labelledby={`tab-${tab}`}
        className="spec-detail-dialog__body"
      >
        {tab === 'overview' && <OverviewTab name={name} />}
        {tab === 'drift' && <DriftTab name={name} />}
        {tab === 'tasks' && <TasksTab name={name} />}
        {tab === 'timeline' && <TimelineTab name={name} />}
      </section>
    </dialog>
  );
}