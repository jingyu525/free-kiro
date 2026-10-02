// widgets/specs-table — sortable specs table with 7 columns.

import { useSummaryQuery } from '@entities/summary/use-summary-query';
import { formatRelative } from '@shared/lib/format';
import { PhaseBadge } from '@widgets/phase-badge';

export function SpecsTable(): JSX.Element {
  const { data } = useSummaryQuery();
  const specs = data?.specs ?? [];

  return (
    <table className="specs-table">
      <caption>specs</caption>
      <thead>
        <tr>
          <th scope="col">name</th>
          <th scope="col">phase</th>
          <th scope="col">approved</th>
          <th scope="col">drift</th>
          <th scope="col">tasks</th>
          <th scope="col">waves</th>
          <th scope="col">updated</th>
        </tr>
      </thead>
      <tbody>
        {specs.map((s) => (
          <tr key={s.meta.name}>
            <td>
              <a href={`#/spec/${s.meta.name}`}>{s.meta.name}</a>
            </td>
            <td>
              <PhaseBadge phase={s.meta.phase} />
            </td>
            <td aria-label={s.meta.approved ? 'approved' : 'not approved'}>
              {s.meta.approved ? '✓' : '—'}
            </td>
            <td>{s.drift.length}</td>
            <td>
              {s.tasks.done}/{s.tasks.total}
            </td>
            <td>{s.tasks.waves}</td>
            <td>{formatRelative(new Date(s.meta.updated_at).getTime())}</td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}
