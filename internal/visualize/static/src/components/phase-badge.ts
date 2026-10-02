// components/phase-badge.ts — small coloured chip showing the spec
// phase. Colour comes from --color-state-* tokens; we don't put
// colour names in JS so the theme can swap them.

import type { Phase } from '../types';

export function renderPhaseBadge(phase: Phase): HTMLElement {
  const span = document.createElement('span');
  span.className = 'fk-phase-badge';
  span.dataset['phase'] = phase;
  span.textContent = phase;
  return span;
}
