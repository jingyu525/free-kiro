// TasksTab — fetches spec tasks and renders waves progress + task list.
import { useSpecTasksQuery } from '@entities/spec/use-spec-query';
import { TaskList } from '@widgets/task-list';
import { WavesProgress } from '@widgets/waves-progress';
import { Spinner } from '@shared/ui/Spinner';
import { RetryBanner } from '@features/retry-fetch';

export function TasksTab({ name }: { name: string }): JSX.Element {
  const { data, isLoading, isError, error, refetch } = useSpecTasksQuery(name);
  if (isLoading) return <Spinner />;
  if (isError) return <RetryBanner error={error} onRetry={() => void refetch()} />;
  const tasks = data ?? [];
  return (
    <>
      <WavesProgress tasks={tasks} />
      <TaskList tasks={tasks} />
    </>
  );
}
