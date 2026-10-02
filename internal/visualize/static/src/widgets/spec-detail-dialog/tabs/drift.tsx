// DriftTab — fetches spec drift and renders via DriftTable.
import { useSpecDriftQuery } from '@entities/spec/use-spec-query';
import { DriftTable } from '@widgets/drift-table';
import { Spinner } from '@shared/ui/Spinner';
import { RetryBanner } from '@features/retry-fetch';

export function DriftTab({ name }: { name: string }): JSX.Element {
  const { data, isLoading, isError, error, refetch } = useSpecDriftQuery(name);
  if (isLoading) return <Spinner />;
  if (isError) return <RetryBanner error={error} onRetry={() => void refetch()} />;
  return <DriftTable drift={data ?? []} specName={name} />;
}
