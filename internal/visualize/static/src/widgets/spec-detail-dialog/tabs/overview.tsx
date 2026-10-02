// OverviewTab — fetches and displays spec overview metadata.
import { useQuery } from '@tanstack/react-query';
import { fetchSpecOverview } from '@entities/spec/api';
import type { SpecOverview } from '@entities/spec/types';
import { Spinner } from '@shared/ui/Spinner';
import { RetryBanner } from '@features/retry-fetch';

export function OverviewTab({ name }: { name: string }): JSX.Element {
  const { data, isLoading, isError, error, refetch } = useQuery<SpecOverview>({
    queryKey: ['spec', name, 'overview'],
    queryFn: () => fetchSpecOverview(name),
  });
  if (isLoading) return <Spinner />;
  if (isError || !data) return <RetryBanner error={error} onRetry={() => void refetch()} />;
  return (
    <dl className="overview-meta">
      <dt>phase</dt><dd>{data.meta.phase}</dd>
      <dt>approved</dt><dd>{data.meta.approved ? '✓' : '—'}</dd>
      <dt>workflow</dt><dd>{data.meta.workflow}</dd>
      <dt>spec_type</dt><dd>{data.meta.spec_type}</dd>
      <dt>created</dt><dd>{data.meta.created_at}</dd>
      <dt>updated</dt><dd>{data.meta.updated_at}</dd>
      <dt>tasks</dt><dd>{data.tasks_summary.done}/{data.tasks_summary.total} ({data.tasks_summary.waves} waves)</dd>
    </dl>
  );
}
