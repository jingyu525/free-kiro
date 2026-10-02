// entities/spec/types — spec-level API payload shapes.

export interface DriftDetail {
  key: string;
  baseline: number;
  current: number;
  delta: number;
}

export interface TaskProgress {
  task_id: string;
  title: string;
  deps: string[];
  done: boolean;
  wave: number;
}

export interface SpecTimelineEvent {
  ts: string;
  kind: 'created' | 'approved' | 'started' | 'completed';
  message: string;
}