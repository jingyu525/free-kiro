// entities/spec/types — spec-level API payload shapes.
//
// After dashboard-frontend-components normalize, all collections are guaranteed
// non-null at the consumer layer. The types below stay non-nullable so widgets
// stay terse, and normalize handles the wire nullability at the entry point.

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

export interface SpecOverview {
  meta: {
    name: string;
    phase: string;
    approved: boolean;
    created_at: string;
    updated_at: string;
    workflow: string;
    spec_type: string;
  };
  drift: DriftDetail[];
  current: Record<string, number>;
  tasks_summary: { done: number; total: number; waves: number };
}

export interface SpecTimelineEvent {
  ts: string;
  kind: 'created' | 'approved' | 'started' | 'completed';
  message: string;
}