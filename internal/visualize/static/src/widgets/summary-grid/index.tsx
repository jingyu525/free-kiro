// widgets/summary-grid — 7-stat overview grid (total + per-phase counts).

import { useMemo } from 'react';
import { useSummaryQuery } from '@entities/summary/use-summary-query';
import type { SpecMeta } from '@entities/summary/types';
import { StatCard } from '@widgets/stat-card';

type Phase = SpecMeta['phase'];
type PhaseCount = Record<Phase, number>;

const PHASES: Phase[] = [
  'draft', 'requirements', 'design', 'tasks',
  'approved', 'implementing', 'done',
];

export function SummaryGrid(): JSX.Element {
  const { data } = useSummaryQuery();

  const counts = useMemo<PhaseCount>(() => {
    const init: PhaseCount = {
      draft: 0, requirements: 0, design: 0, tasks: 0,
      approved: 0, implementing: 0, done: 0,
    };
    return (data?.specs ?? []).reduce<PhaseCount>((acc, s) => {
      const p = s.meta.phase;
      if (p in acc) acc[p] += 1;
      return acc;
    }, init);
  }, [data?.specs]);

  const total = data?.specs.length ?? 0;

  return (
    <section className="summary-grid" aria-label="summary">
      <StatCard label="total specs" value={total} />
      {PHASES.map((p) => (
        <StatCard key={p} label={p} value={counts[p]} />
      ))}
    </section>
  );
}
