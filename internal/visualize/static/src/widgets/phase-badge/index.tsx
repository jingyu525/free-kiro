// widgets/phase-badge — colour-coded spec phase chip with text+colour redundancy for a11y.

import type { SpecMeta } from '@entities/summary/types';

type Phase = SpecMeta['phase'];

const PHASE_STYLE: Record<Phase, { bg: string; color: string; label: string }> = {
  draft:        { bg: '#6b7280', color: '#fff', label: 'draft' },
  requirements: { bg: '#3b82f6', color: '#fff', label: 'requirements' },
  design:       { bg: '#8b5cf6', color: '#fff', label: 'design' },
  tasks:        { bg: '#eab308', color: '#000', label: 'tasks' },
  approved:     { bg: '#22c55e', color: '#fff', label: 'approved' },
  implementing: { bg: '#f97316', color: '#fff', label: 'implementing' },
  done:         { bg: '#15803d', color: '#fff', label: 'done' },
};

export function PhaseBadge({ phase }: { phase: Phase }): JSX.Element {
  const { bg, color, label } = PHASE_STYLE[phase] ?? PHASE_STYLE['draft'];
  return (
    <span
      className="phase-badge"
      style={{ backgroundColor: bg, color }}
      aria-label={`phase: ${label}`}
    >
      {label}
    </span>
  );
}
