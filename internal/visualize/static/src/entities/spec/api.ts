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

import { fetchJson } from '@shared/api/client';
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
  const r = await fetchJson<TasksResponse>(`/spec/${encodeURIComponent(name)}/tasks`, {
    allow404: true,
  });
  if (!r || !Array.isArray(r.waves)) return [];
  return r.waves.flatMap((w) =>
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
  const r = await fetchJson<DriftResponse>(`/spec/${encodeURIComponent(name)}/drift`, {
    allow404: true,
  });
  return Array.isArray(r?.signals) ? r.signals : [];
}

export async function fetchSpecTimeline(name: string): Promise<SpecTimelineEvent[]> {
  const raw = await fetchJson<Array<{ ts: string; kind: string; message: string }>>(
    `/spec/${encodeURIComponent(name)}/timeline`,
    { allow404: true },
  );
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
  const [statusRaw, drift, tasksList] = await Promise.all([
    fetchJson<{
      meta?: SpecOverview['meta'];
      current?: Record<string, number>;
    }>(`/spec/${encoded}`, { allow404: true }),
    fetchSpecDrift(name),
    fetchSpecTasks(name),
  ]);
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