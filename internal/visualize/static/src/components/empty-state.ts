// components/empty-state.ts — three-way tri-state CTA: workspace
// missing → init; no specs → new; ok-empty → same as no-specs.

import type { WorkspaceMode } from '../types';

interface EmptyStateProps {
  mode: WorkspaceMode;
  specs: number;
}

interface CTA {
  title: string;
  body: string;
  command?: string;
}

export function EmptyState(props: EmptyStateProps): HTMLElement {
  const cta = ctaForMode(props.mode);
  const section = document.createElement('section');
  section.className = 'fk-empty-state';
  section.setAttribute('role', 'region');
  section.setAttribute('aria-labelledby', 'fk-empty-title');

  const title = document.createElement('h2');
  title.id = 'fk-empty-title';
  title.textContent = cta.title;
  section.appendChild(title);

  const body = document.createElement('p');
  body.textContent = cta.body;
  section.appendChild(body);

  if (cta.command) {
    const pre = document.createElement('pre');
    pre.className = 'fk-empty-state__command';
    const code = document.createElement('code');
    code.textContent = cta.command;
    pre.appendChild(code);
    section.appendChild(pre);
  }

  return section;
}

function ctaForMode(mode: WorkspaceMode): CTA {
  switch (mode) {
    case 'workspace-missing':
      return {
        title: 'No .kiro/ workspace found',
        body: 'Bootstrap one to start tracking specs.',
        command: 'free-kiro init',
      };
    case 'no-specs':
      return {
        title: 'No specs in this workspace',
        body: 'Create your first spec to get started.',
        command: 'free-kiro spec new <name>',
      };
    case 'ok':
    default:
      return {
        title: 'No specs yet',
        body: 'Create one to populate the dashboard.',
        command: 'free-kiro spec new <name>',
      };
  }
}
