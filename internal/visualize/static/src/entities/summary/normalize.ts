// entities/summary/normalize — null-safe projection for wire payloads.
//
// The Go backend (internal/visualize/report.go) serializes Go nil slices and
// nil maps as JSON null. The dashboard-frontend-foundation MVP shipped with
// a known TypeError ("Cannot read properties of null (reading 'length')")
// because widgets read `data.specs.length` / `spec.drift.length` / etc.
// without null-defensive coding.
//
// normalizeProjectReport() runs inside react-query's `select` option so every
// consumer sees a fully populated ProjectReport with non-null collections.
// Defensive guards stay here so widgets stay simple.

import type {
  ProjectReport,
  SpecMeta,
  SpecReport,
  SpecTaskProgress,
  SummaryMode,
} from './types';

const ZERO_TASKS: SpecTaskProgress = { done: 0, total: 0, waves: 0 };

const EMPTY_META: SpecMeta = {
  name: '',
  phase: 'draft',
  workflow: '',
  spec_type: 'feature',
  quick: false,
  approved: false,
  generator: '',
  prompt: '',
  created_at: '',
  updated_at: '',
};

function safeTasks(raw: Partial<SpecTaskProgress> | null | undefined): SpecTaskProgress {
  if (!raw) return ZERO_TASKS;
  return {
    done: typeof raw.done === 'number' ? raw.done : 0,
    total: typeof raw.total === 'number' ? raw.total : 0,
    waves: typeof raw.waves === 'number' ? raw.waves : 0,
  };
}

export function normalizeSpecReport(raw: SpecReport | null | undefined): SpecReport {
  return {
    meta: raw?.meta ?? EMPTY_META,
    current: raw?.current ?? {},
    drift: Array.isArray(raw?.drift) ? raw.drift : [],
    tasks: safeTasks(raw?.tasks),
    active: Boolean(raw?.active),
  };
}

export function normalizeProjectReport(
  raw: ProjectReport | null | undefined,
): ProjectReport {
  const specs = Array.isArray(raw?.specs) ? raw.specs : [];
  const mode: SummaryMode = raw?.mode ?? 'ok';
  return {
    generated_at: raw?.generated_at ?? '',
    specs: specs.map(normalizeSpecReport),
    active: raw?.active ?? '',
    mode,
    workspace_ref: raw?.workspace_ref,
    started_at: raw?.started_at,
    last_refresh_at: raw?.last_refresh_at,
  };
}