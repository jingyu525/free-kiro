// TimelineTab — fetches spec timeline events and renders chronologically.
import { useQuery } from '@tanstack/react-query';
import { fetchSpecTimeline } from '@entities/spec/api';
import type { SpecTimelineEvent } from '@entities/spec/types';
import { Spinner } from '@shared/ui/Spinner';
import { RetryBanner } from '@features/retry-fetch';
import { formatRelative } from '@shared/lib/format';

export function TimelineTab({ name }: { name: string }): JSX.Element {
  const { data, isLoading, isError, error, refetch } = useQuery<SpecTimelineEvent[]>({
    queryKey: ['spec', name, 'timeline'],
    queryFn: () => fetchSpecTimeline(name),
  });
  if (isLoading) return <Spinner />;
  if (isError) return <RetryBanner error={error} onRetry={() => void refetch()} />;
  const events = data ?? [];
  if (events.length === 0) return <p>No timeline events.</p>;
  return (
    <ol className="timeline">
      {events.map((e, i) => (
        <li key={`${e.ts}-${i}`} className="timeline-item" data-kind={e.kind}>
          <time dateTime={e.ts}>{formatRelative(new Date(e.ts).getTime())}</time>
          <strong>{e.kind}</strong>
          <span>{e.message}</span>
        </li>
      ))}
    </ol>
  );
}
