// components/specs-table.ts — main spec table with phase chips and
// drift / tasks columns. spec names are deep-link anchors pointing
// at the (future) detail dialog; this MVP doesn't render the dialog.

import type { SpecReport } from '../types';
import { renderPhaseBadge } from './phase-badge';

interface SpecsTableProps {
  specs: SpecReport[];
  active: string;
}

export function SpecsTable(props: SpecsTableProps): HTMLElement {
  const section = document.createElement('div');
  section.className = 'fk-specs-table-region';

  const heading = document.createElement('h2');
  heading.id = 'specs-heading';
  heading.className = 'fk-specs-heading';
  heading.textContent = 'Specs';
  section.appendChild(heading);

  const table = document.createElement('table');
  table.className = 'fk-specs-table';
  table.setAttribute('aria-describedby', 'specs-heading');

  const caption = document.createElement('caption');
  caption.className = 'fk-sr-only';
  caption.textContent = `Project specs (${props.specs.length})`;
  table.appendChild(caption);

  const thead = document.createElement('thead');
  const headerRow = document.createElement('tr');
  for (const col of ['spec', 'phase', 'approved', 'drift', 'tasks']) {
    const th = document.createElement('th');
    th.setAttribute('scope', 'col');
    th.textContent = col;
    headerRow.appendChild(th);
  }
  thead.appendChild(headerRow);
  table.appendChild(thead);

  const tbody = document.createElement('tbody');
  for (const s of props.specs) {
    tbody.appendChild(renderRow(s));
  }
  table.appendChild(tbody);

  section.appendChild(table);
  return section;
}

function renderRow(s: SpecReport): HTMLTableRowElement {
  const tr = document.createElement('tr');
  if (s.active) tr.dataset['active'] = 'true';

  // Spec name as deep-link anchor. The MVP keeps the link pointing
  // at "#spec/<name>"; future detail-dialog spec wires the dialog.
  const specCell = document.createElement('td');
  const link = document.createElement('a');
  link.className = 'fk-spec-link';
  link.href = `#spec/${encodeURIComponent(s.meta.name)}`;
  link.textContent = s.meta.name;
  if (s.active) link.setAttribute('aria-current', 'true');
  specCell.appendChild(link);
  tr.appendChild(specCell);

  const phaseCell = document.createElement('td');
  phaseCell.appendChild(renderPhaseBadge(s.meta.phase));
  tr.appendChild(phaseCell);

  const approvedCell = document.createElement('td');
  approvedCell.textContent = s.meta.approved ? 'yes' : 'no';
  tr.appendChild(approvedCell);

  const driftCell = document.createElement('td');
  const drift = s.drift ?? [];
  if (drift.length > 0) {
    driftCell.className = 'fk-drift-cell';
    driftCell.textContent = `⚠ ${drift.length} keys`;
    driftCell.setAttribute('aria-label', `${drift.length} drift signals`);
  } else {
    driftCell.textContent = 'none';
  }
  tr.appendChild(driftCell);

  const tasksCell = document.createElement('td');
  if (s.tasks.total > 0) {
    tasksCell.textContent = `${s.tasks.done}/${s.tasks.total}`;
  } else {
    tasksCell.textContent = '—';
  }
  tr.appendChild(tasksCell);

  return tr;
}
