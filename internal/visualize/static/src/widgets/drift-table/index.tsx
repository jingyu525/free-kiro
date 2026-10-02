// drift-table — renders drift signals as a table: key / baseline / current / delta.
// delta column uses data-trend for CSS styling (up=red, down=green, flat=gray).
import type { DriftSignal } from '@entities/summary/types';

export interface DriftTableProps {
  drift: ReadonlyArray<DriftSignal>;
}

function deltaLabel(delta: number): { text: string; trend: 'up' | 'down' | 'flat'; ariaLabel: string } {
  if (delta > 0) return { text: `+${delta}`, trend: 'up', ariaLabel: `increased by ${delta}` };
  if (delta < 0) return { text: `${delta}`, trend: 'down', ariaLabel: `decreased by ${Math.abs(delta)}` };
  return { text: '±0', trend: 'flat', ariaLabel: 'no change' };
}

export function DriftTable({ drift }: DriftTableProps): JSX.Element {
  if (drift.length === 0) {
    return <p className="empty">No drift signals.</p>;
  }

  return (
    <table className="drift-table">
      <caption>drift</caption>
      <thead>
        <tr>
          <th scope="col">key</th>
          <th scope="col">baseline</th>
          <th scope="col">current</th>
          <th scope="col">delta</th>
        </tr>
      </thead>
      <tbody>
        {drift.map((signal, i) => {
          const { text, trend, ariaLabel } = deltaLabel(signal.delta);
          return (
            <tr key={signal.key} aria-rowindex={i + 2}>
              <td>{signal.key}</td>
              <td>{signal.baseline}</td>
              <td>{signal.current}</td>
              <td data-trend={trend} aria-label={ariaLabel}>{text}</td>
            </tr>
          );
        })}
      </tbody>
    </table>
  );
}
