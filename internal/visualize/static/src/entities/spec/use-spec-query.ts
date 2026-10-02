// entities/spec/use-spec-query — react-query hooks.

import { useQuery, type UseQueryResult } from '@tanstack/react-query';
import { fetchSpecDrift, fetchSpecTasks } from './api';
import type { DriftDetail, TaskProgress } from './types';

export function useSpecTasksQuery(
  name: string,
): UseQueryResult<TaskProgress[], Error> {
  return useQuery({
    queryKey: ['spec', name, 'tasks'],
    queryFn: () => fetchSpecTasks(name),
    staleTime: 30_000,
    enabled: Boolean(name),
  });
}

export function useSpecDriftQuery(
  name: string,
): UseQueryResult<DriftDetail[], Error> {
  return useQuery({
    queryKey: ['spec', name, 'drift'],
    queryFn: () => fetchSpecDrift(name),
    staleTime: 30_000,
    enabled: Boolean(name),
  });
}