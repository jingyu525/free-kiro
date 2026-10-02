// widgets/empty-state — contextual CTA shown when workspace is absent or has no specs.

import type { SummaryMode } from '@entities/summary/types';

interface EmptyStateProps {
  mode: SummaryMode | undefined;
}

export function EmptyState({ mode }: EmptyStateProps): JSX.Element | null {
  if (mode === 'workspace-missing') {
    return (
      <div className="empty-state" role="status">
        <p>No .kiro/ workspace found.</p>
        <button type="button" className="btn btn-primary" onClick={() => {
          window.location.hash = '#/init';
        }}>
          Run free-kiro init
        </button>
      </div>
    );
  }

  if (mode === 'no-specs' || mode === 'ok') {
    return (
      <div className="empty-state" role="status">
        <p>No specs in this workspace.</p>
        <button type="button" className="btn btn-primary" onClick={() => {
          window.location.hash = '#/spec/new';
        }}>
          Run free-kiro spec new &lt;name&gt;
        </button>
      </div>
    );
  }

  return null;
}
