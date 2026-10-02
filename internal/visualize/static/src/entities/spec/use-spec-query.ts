// entities/spec/use-spec-query — react-query hooks.
// fetchSpecTasks / fetchSpecDrift internally use fetchJsonWithEtag so
// the module-level ETag cache holds the *flattened* array shape that
// react-query consumers expect. No sentinel throw needed.

import { useQuery, type UseQueryResult } from '@tanstack/react-query';
import { fetchSpecDrift, fetchSpecTasks } from './api';
import type { DriftDetail, TaskProgress } from './types';

export function useSpecTasksQuery(
  name: string,
): UseQueryResult<TaskProgress[], Error> {
  return useQuery({
    queryKey: ['spec', name, 'tasks'],
    queryFn: () => fetchSpecTasks(name),
    staleTime: 5_000,
    enabled: Boolean(name),
  });
}

export function useSpecDriftQuery(
  name: string,
): UseQueryResult<DriftDetail[], Error> {
  return useQuery({
    queryKey: ['spec', name, 'drift'],
    queryFn: () => fetchSpecDrift(name),
    staleTime: 5_000,
    enabled: Boolean(name),
  });
}