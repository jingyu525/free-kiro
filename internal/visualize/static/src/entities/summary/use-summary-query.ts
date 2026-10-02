// entities/summary/use-summary-query — react-query hook.
// fetchSummary wires through fetchJsonWithEtag inside shared/api/client;
// the cache stores the raw wire payload; normalize runs on select so
// the post-304 path produces the same normalized shape.

import { useQuery, type UseQueryResult } from '@tanstack/react-query';
import { fetchSummary } from './api';
import { normalizeProjectReport } from './normalize';
import type { ProjectReport } from './types';

export const SUMMARY_QUERY_KEY = ['summary'] as const;

export function useSummaryQuery(): UseQueryResult<ProjectReport, Error> {
  return useQuery({
    queryKey: SUMMARY_QUERY_KEY,
    queryFn: fetchSummary,
    select: normalizeProjectReport,
    staleTime: 5_000,
    refetchOnWindowFocus: false,
  });
}