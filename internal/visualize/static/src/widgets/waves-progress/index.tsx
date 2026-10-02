// waves-progress — renders task completion progress grouped by wave.
import type { TaskProgress } from '@entities/spec/types';

export interface WavesProgressProps {
  tasks: ReadonlyArray<TaskProgress>;
}

export function WavesProgress({ tasks }: WavesProgressProps): JSX.Element {
  if (tasks.length === 0) {
    return <p>No tasks.</p>;
  }

  const groups = new Map<number, { done: number; total: number }>();
  for (const t of tasks) {
    const wave = Math.min(t.wave, 8);
    const g = groups.get(wave) ?? { done: 0, total: 0 };
    groups.set(wave, { done: g.done + (t.done ? 1 : 0), total: g.total + 1 });
  }

  const sorted = [...groups.entries()].sort(([a], [b]) => a - b);

  return (
    <div className="waves-progress">
      {sorted.map(([wave, { done, total }]) => (
        <div key={wave} className={`wave wave-${wave}`}>
          <h3>wave {wave}</h3>
          <progress max={total} value={done} aria-label={`Wave ${wave} progress`}>
            {done}/{total}
          </progress>
          <span>{done}/{total}</span>
        </div>
      ))}
    </div>
  );
}
