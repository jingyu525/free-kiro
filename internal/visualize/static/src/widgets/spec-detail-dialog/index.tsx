// SpecDetailDialog — modal dialog driven by hash route (spec/<name>).
// Renders 4 tabs: overview / drift / tasks / timeline with focus management.
import { useEffect, useRef, useState } from 'react';
import { useHashRoute, navigateToHash } from '@shared/lib/hash-router';
import { announce } from '@shared/lib/a11y/announce';
import { OverviewTab } from './tabs/overview';
import { DriftTab } from './tabs/drift';
import { TasksTab } from './tabs/tasks';
import { TimelineTab } from './tabs/timeline';

export function SpecDetailDialog(): JSX.Element | null {
  const route = useHashRoute();
  const isOpen = route.startsWith('spec/');
  const name = isOpen ? route.slice('spec/'.length) : '';

  const dialogRef = useRef<HTMLDialogElement>(null);
  const triggerRef = useRef<HTMLElement | null>(null);
  const [tab, setTab] = useState<'overview' | 'drift' | 'tasks' | 'timeline'>('overview');

  useEffect(() => {
    const dialog = dialogRef.current;
    if (!dialog) return;
    if (isOpen && !dialog.open) {
      triggerRef.current = document.activeElement as HTMLElement;
      dialog.showModal();
      const first = dialog.querySelector<HTMLElement>('button, [href], input, select, textarea');
      first?.focus();
      announce(`Spec ${name} details opened`);
    } else if (!isOpen && dialog.open) {
      dialog.close();
      triggerRef.current?.focus();
    }
  }, [isOpen, name]);

  const onClick = (e: React.MouseEvent<HTMLDialogElement>): void => {
    if (e.target === dialogRef.current) {
      navigateToHash('');
    }
  };

  const onClose = (): void => {
    navigateToHash('');
    setTab('overview');
  };

  if (!isOpen) return null;

  return (
    <dialog
      ref={dialogRef}
      className="spec-detail-dialog"
      aria-labelledby="dialog-title"
      onClose={onClose}
      onClick={onClick}
    >
      <header className="spec-detail-dialog__header">
        <h2 id="dialog-title">{name}</h2>
        <button
          type="button"
          className="btn btn-ghost"
          aria-label="Close"
          onClick={() => navigateToHash('')}
        >
          ×
        </button>
      </header>
      <nav className="spec-detail-dialog__tabs" role="tablist" aria-label="Spec detail tabs">
        <button
          role="tab"
          aria-selected={tab === 'overview'}
          aria-controls="panel-overview"
          id="tab-overview"
          onClick={() => setTab('overview')}
        >
          overview
        </button>
        <button
          role="tab"
          aria-selected={tab === 'drift'}
          aria-controls="panel-drift"
          id="tab-drift"
          onClick={() => setTab('drift')}
        >
          drift
        </button>
        <button
          role="tab"
          aria-selected={tab === 'tasks'}
          aria-controls="panel-tasks"
          id="tab-tasks"
          onClick={() => setTab('tasks')}
        >
          tasks
        </button>
        <button
          role="tab"
          aria-selected={tab === 'timeline'}
          aria-controls="panel-timeline"
          id="tab-timeline"
          onClick={() => setTab('timeline')}
        >
          timeline
        </button>
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
