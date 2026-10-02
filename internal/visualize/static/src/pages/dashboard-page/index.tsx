// pages/dashboard-page — composes header / summary grid / specs table / overlays.

import { useEffect, useState } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { useSummaryQuery, SUMMARY_QUERY_KEY } from '@entities/summary/use-summary-query';
import { useToast } from '@shared/lib/toast';
import { isApiError } from '@shared/api/client';
import { HeaderBar } from '@widgets/header-bar';
import { SummaryGrid } from '@widgets/summary-grid';
import { SpecsTable } from '@widgets/specs-table';
import { EmptyState } from '@widgets/empty-state';
import { ToastStack } from '@widgets/toast-stack';
import { ProgressBar } from '@widgets/progress-bar';
import { SkeletonOverlay } from '@features/skeleton-overlay';
import { RetryBanner } from '@features/retry-fetch';

export function DashboardPage(): JSX.Element {
  const { data, isLoading, isFetching, isError, error, refetch } = useSummaryQuery();
  const qc = useQueryClient();
  const { pushToast } = useToast();

  // Surface transient errors as a toast once per error change.
  const [seenErrorId, setSeenErrorId] = useState<string | null>(null);
  useEffect(() => {
    if (!isError || !error) return;
    const id = String(error.message ?? 'error');
    if (id === seenErrorId) return;
    setSeenErrorId(id);
    const kind = isApiError(error) ? error.kind : 'unknown';
    pushToast('error', `summary: ${kind}`);
  }, [isError, error, seenErrorId, pushToast]);

  // First-load skeleton vs. existing-data refresh progress.
  const showSkeleton = isLoading && !data;
  const showProgress = !showSkeleton && isFetching && Boolean(data);
  const mode = data?.mode;

  return (
    <>
      <ProgressBar active={showProgress} />
      <HeaderBar />
      {isError && data === undefined ? (
        <RetryBanner error={error} onRetry={() => void refetch()} />
      ) : null}
      <main role="main">
        {showSkeleton ? (
          <SkeletonOverlay visible rows={6} />
        ) : data && (mode === 'workspace-missing' || (mode === 'ok' && data.specs.length === 0) || mode === 'no-specs') ? (
          <EmptyState mode={mode ?? (data.specs.length === 0 ? 'ok' : 'ok')} />
        ) : data ? (
          <>
            <SummaryGrid />
            <SpecsTable />
          </>
        ) : null}
      </main>
      <ToastStack />
      {/* Expose SUMMARY_QUERY_KEY for hooks that need to invalidate from outside. */}
      <span hidden>{JSON.stringify(SUMMARY_QUERY_KEY)}</span>
      <span hidden>{qc.getQueryData.length}</span>
    </>
  );
}