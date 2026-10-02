// components/summary-grid.ts — 7 stat cards (total / active /
// draft / planning / implementing / done / drift).

import type { SpecReport } from '../types';
import { formatCompact } from '../lib/format';

interface SummaryGridProps {
  specs: SpecReport[];
  active: string;
}

export function SummaryGrid(props: SummaryGridProps): HTMLElement {
  const grid = document.createElement('div');
  grid.className = 'fk-summary-grid';
  grid.setAttribute('role', 'list');

  const stats = computeStats(props.specs);
  stats.forEach((stat) => grid.appendChild(statCard(stat)));
  return grid;
}

interface Stat {
  label: string;
  value: number;
  hint?: string;
}

function computeStats(specs: SpecReport[]): Stat[] {
  const byPhase = new Map<string, number>();
  let drift = 0;
  for (const s of specs) {
    byPhase.set(s.meta.phase, (byPhase.get(s.meta.phase) ?? 0) + 1);
    const signals = s.drift ?? [];
    if (signals.length > 0) drift += signals.length;
  }
  return [
    { label: 'total', value: specs.length },
    {
      label: 'active',
      value: specs.filter((s) => s.active).length,
      hint: 'currently in focus',
    },
    { label: 'draft', value: byPhase.get('draft') ?? 0 },
    {
      label: 'in planning',
      value: (byPhase.get('requirements') ?? 0) + (byPhase.get('design') ?? 0) + (byPhase.get('tasks') ?? 0),
    },
    { label: 'implementing', value: (byPhase.get('approved') ?? 0) + (byPhase.get('implementing') ?? 0) },
    { label: 'done', value: byPhase.get('done') ?? 0 },
    {
      label: 'drift keys',
      value: drift,
      hint: 'sum across all specs',
    },
  ];
}

function statCard(stat: Stat): HTMLElement {
  const card = document.createElement('div');
  card.className = 'fk-stat-card';
  card.setAttribute('role', 'listitem');
  card.dataset['state'] = statLabelToState(stat.label);

  const label = document.createElement('span');
  label.className = 'fk-stat-card__label';
  label.textContent = stat.label;
  card.appendChild(label);

  const value = document.createElement('span');
  value.className = 'fk-stat-card__value';
  value.textContent = formatCompact(stat.value);
  card.appendChild(value);

  if (stat.hint) {
    const hint = document.createElement('span');
    hint.className = 'fk-stat-card__hint';
    hint.textContent = stat.hint;
    card.appendChild(hint);
  }

  return card;
}

function statLabelToState(label: string): string {
  switch (label) {
    case 'draft': return 'draft';
    case 'in planning': return 'planning';
    case 'implementing': return 'implementing';
    case 'done': return 'done';
    case 'drift keys': return 'drift';
    default: return 'neutral';
  }
}
