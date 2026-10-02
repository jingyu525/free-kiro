// entities/spec/api — fetch wrappers for /api/spec/<name>/*.
//
// Backend responses are wrapped objects (dashboard-backend-api-extensions):
//   /api/spec/<name>/tasks  → {spec, waves:[{index,done,total,tasks:[]}]}
//   /api/spec/<name>/drift  → {spec, signals:[]}
//   /api/spec/<name>/timeline → Array<{ts,kind,message}>
//
// We flatten these into the TaskProgress[] / DriftDetail[] shapes that
// widgets expect, doing null-safety on `deps: null` (backend omits deps for
// wave-1 root tasks).
//
// All wrappers go through fetchJsonWithEtag so the module-level ETag
// cache is keyed on the *flattened* result — react-query consumers see
// consistent array shapes across 200 / 304 responses.

import { fetchJsonWithEtag, fetchJson } from '@shared/api/client';
import type { DriftDetail, SpecOverview, TaskProgress, SpecTimelineEvent } from './types';

interface TasksResponse {
  spec: string;
  waves: Array<{
    index: number;
    done: number;
    total: number;
    tasks: Array<{
      id: number | string;
      title: string;
      done: boolean;
      deps: string[] | null;
      wave: number;
    }>;
  }>;
}

interface DriftResponse {
  spec: string;
  signals: DriftDetail[];
}

export async function fetchSpecTasks(name: string): Promise<TaskProgress[]> {
  // We fetch raw tasks and flatten in this function so the ETag cache
  // holds the *flattened* result that react-query consumers expect. If
  // we cached the raw `{spec, waves}` response here, 304 would return
  // that object — and downstream `.map(...)` on it would throw.
  const encoded = encodeURIComponent(name);
  const raw = await fetchJson<TasksResponse>(`/spec/${encoded}/tasks`, { allow404: true });
  if (!raw || !Array.isArray(raw.waves)) return [];
  return raw.waves.flatMap((w) =>
    (w.tasks ?? []).map((t) => ({
      task_id: String(t.id),
      title: t.title ?? '',
      deps: Array.isArray(t.deps) ? t.deps : [],
      done: Boolean(t.done),
      wave: typeof t.wave === 'number' ? t.wave : w.index,
    })),
  );
}

export async function fetchSpecDrift(name: string): Promise<DriftDetail[]> {
  const r = await fetchJsonWithEtag<DriftResponse>(`/spec/${encodeURIComponent(name)}/drift`, {
    allow404: true,
  });
  const body = r.data;
  return Array.isArray(body?.signals) ? body.signals : [];
}

export async function fetchSpecTimeline(name: string): Promise<SpecTimelineEvent[]> {
  const r = await fetchJsonWithEtag<Array<{ ts: string; kind: string; message: string }>>(
    `/spec/${encodeURIComponent(name)}/timeline`,
    { allow404: true },
  );
  const raw = r.data;
  return Array.isArray(raw)
    ? raw.map((e) => ({
        ts: e.ts,
        kind:
          e.kind === 'created' ||
          e.kind === 'approved' ||
          e.kind === 'started' ||
          e.kind === 'completed'
            ? e.kind
            : 'created',
        message: e.message,
      }))
    : [];
}

/** Aggregate meta + drift + current + tasks_summary in parallel. */
export async function fetchSpecOverview(name: string): Promise<SpecOverview> {
  const encoded = encodeURIComponent(name);
  const [statusRes, drift, tasksList] = await Promise.all([
    fetchJsonWithEtag<{
      meta?: SpecOverview['meta'];
      current?: Record<string, number>;
    }>(`/spec/${encoded}`, { allow404: true }),
    fetchSpecDrift(name),
    fetchSpecTasks(name),
  ]);
  const statusRaw = statusRes.data;
  const waves = tasksList.length === 0 ? 0 : Math.max(...tasksList.map((t) => t.wave));
  return {
    meta: statusRaw?.meta ?? {
      name,
      phase: 'unknown',
      approved: false,
      created_at: '',
      updated_at: '',
      workflow: '',
      spec_type: 'feature',
    },
    drift,
    current: statusRaw?.current ?? {},
    tasks_summary: {
      done: tasksList.filter((t) => t.done).length,
      total: tasksList.length,
      waves,
    },
  };
}