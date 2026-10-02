// entities/spec/api — fetch wrappers for /api/spec/<name>/*.

import { fetchJson } from '@shared/api/client';
import type { DriftDetail, TaskProgress } from './types';

export async function fetchSpecTasks(name: string): Promise<TaskProgress[]> {
  return fetchJson<TaskProgress[]>(`/spec/${encodeURIComponent(name)}/tasks`);
}

export async function fetchSpecDrift(name: string): Promise<DriftDetail[]> {
  return fetchJson<DriftDetail[]>(`/spec/${encodeURIComponent(name)}/drift`);
}

export async function fetchSpecTimeline(
  name: string,
): Promise<Array<{ ts: string; kind: string; message: string }>> {
  return fetchJson(`/spec/${encodeURIComponent(name)}/timeline`);
}