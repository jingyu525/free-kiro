// entities/summary/use-summary-query — react-query hook.

import { useQuery, type UseQueryResult } from '@tanstack/react-query';
import { fetchSummary } from './api';
import type { ProjectReport } from './types';

export const SUMMARY_QUERY_KEY = ['summary'] as const;

export function useSummaryQuery(): UseQueryResult<ProjectReport, Error> {
  return useQuery({
    queryKey: SUMMARY_QUERY_KEY,
    queryFn: fetchSummary,
    staleTime: 30_000,
    refetchOnWindowFocus: false,
  });
}