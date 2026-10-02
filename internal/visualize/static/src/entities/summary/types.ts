// entities/summary/types — mirrors internal/visualize/report.go JSON shape.

export type SummaryMode = 'workspace-missing' | 'no-specs' | 'ok';

export interface ProjectReport {
  generated_at: string;
  specs: SpecReport[];
  active: string;
  mode?: SummaryMode;
  workspace_ref?: string;
  started_at?: string;
  last_refresh_at?: string;
}

export interface SpecMeta {
  name: string;
  phase:
    | 'draft'
    | 'requirements'
    | 'design'
    | 'tasks'
    | 'approved'
    | 'implementing'
    | 'done';
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

export interface SpecTaskProgress {
  done: number;
  total: number;
  waves: number;
}

export interface SpecReport {
  meta: SpecMeta;
  current: Record<string, number>;
  drift: DriftSignal[];
  tasks: SpecTaskProgress;
  active: boolean;
}