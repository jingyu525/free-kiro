// stores/summary.ts — primary dashboard data: the ProjectReport
// snapshot from /api/summary. Drives SummaryGrid + SpecsTable.

import { Signal } from '../lib/signal';
import { ApiError, type ProjectReport, type WorkspaceMode } from '../types';
import { fetchRetry } from '../lib/api';

export type LoadState = 'idle' | 'loading' | 'error' | 'success';

export interface SummaryState {
  value: ProjectReport | null;
  loadState: LoadState;
  lastError: ApiError | null;
  lastRefreshAt: number;
  loadStartedAt: number;
}

const initial: SummaryState = {
  value: null,
  loadState: 'idle',
  lastError: null,
  lastRefreshAt: 0,
  loadStartedAt: 0,
};

export const summaryStore = new Signal<SummaryState>(initial);

export function refreshSummary(): void {
  summaryStore.update((s) => ({
    ...s,
    loadState: 'loading',
    loadStartedAt: Date.now(),
  }));

  fetchRetry<ProjectReport>('/api/summary')
    .then((data) => {
      // Tolerate older servers that pre-date dashboard-backend-api-extensions.
      const mode: WorkspaceMode = data.mode ?? 'ok';
      summaryStore.set({
        value: { ...data, mode },
        loadState: 'success',
        lastError: null,
        lastRefreshAt: Date.now(),
        loadStartedAt: summaryStore.get().loadStartedAt,
      });
    })
    .catch((e: unknown) => {
      const err = e instanceof ApiError ? e : new ApiError('network', String(e));
      summaryStore.update((s) => ({
        ...s,
        loadState: 'error',
        lastError: err,
      }));
    });
}
