// entities/summary/use-summary-query — react-query hook.

import { useQuery, type UseQueryResult } from '@tanstack/react-query';
import { fetchSummary } from './api';
import { normalizeProjectReport } from './normalize';
import type { ProjectReport } from './types';

export const SUMMARY_QUERY_KEY = ['summary'] as const;

export function useSummaryQuery(): UseQueryResult<ProjectReport, Error> {
  return useQuery({
    queryKey: SUMMARY_QUERY_KEY,
    queryFn: fetchSummary,
    // Run the raw wire payload through normalizeProjectReport so consumers
    // never see null drift/current/tasks fields. Single point of null-safety
    // (dashboard-frontend-components AC-2..AC-7).
    select: normalizeProjectReport,
    staleTime: 30_000,
    refetchOnWindowFocus: false,
  });
}