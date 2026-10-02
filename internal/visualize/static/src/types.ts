// types.ts — wire-format types for free-kiro dashboard API responses.
// Mirrors internal/visualize/report.go JSON tags so a rename on the
// Go side fails `tsc --noEmit` rather than silently corrupting the UI.

export type Phase =
  | 'draft'
  | 'requirements'
  | 'design'
  | 'tasks'
  | 'approved'
  | 'implementing'
  | 'done';

export type WorkspaceMode = 'workspace-missing' | 'no-specs' | 'ok';

export interface SpecMeta {
  name: string;
  phase: Phase;
  workflow: string;
  spec_type: 'feature' | 'bugfix';
  quick: boolean;
  approved: boolean;
  generator: string;
  prompt: string;
  created_at: string;
  updated_at: string;
}

export interface DriftSignal {
  key: string;
  baseline: number;
  current: number;
  delta: number;
}

export interface TaskProgress {
  done: number;
  total: number;
  waves: number;
}

export interface SpecReport {
  meta: SpecMeta;
  current: Record<string, number>;
  drift: DriftSignal[];
  tasks: TaskProgress;
  active: boolean;
}

export interface ProjectReport {
  generated_at: string;
  specs: SpecReport[];
  active: string;
  // New in dashboard-backend-api-extensions. Tolerated as optional
  // so older servers (pre-spec-5) don't break the dashboard — UI
  // treats missing as 'ok'.
  mode?: WorkspaceMode;
}

export interface HealthResponse {
  status: string;
  uptime_seconds: number;
  last_refresh_at: string;
  subscribers: number;
  version: string;
}

// Error classes for the api.ts layer. Mirrors the three failure modes
// of fetch (network, HTTP non-2xx, JSON parse) so UI can pick the
// right banner colour / toast wording.
export type ApiErrorKind = 'network' | 'http' | 'parse';

export class ApiError extends Error {
  readonly kind: ApiErrorKind;
  readonly status?: number;
  constructor(kind: ApiErrorKind, message: string, status?: number) {
    super(message);
    this.name = 'ApiError';
    this.kind = kind;
    this.status = status;
  }
}
