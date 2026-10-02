// entities/summary/api — fetch wrappers for /api/summary.

import { fetchJson } from '@shared/api/client';
import type { ProjectReport } from './types';

export async function fetchSummary(): Promise<ProjectReport> {
  return fetchJson<ProjectReport>('/summary');
}