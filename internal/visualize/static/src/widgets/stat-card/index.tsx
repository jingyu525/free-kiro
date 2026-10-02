// widgets/stat-card — labelled metric tile with optional delta indicator.

import { Card } from '@shared/ui/Card';
import { formatDelta } from '@shared/lib/format';

interface StatCardProps {
  label: string;
  value: string | number;
  delta?: number;
}

export function StatCard({ label, value, delta }: StatCardProps): JSX.Element {
  return (
    <Card className="stat-card">
      <span className="stat-card__label">{label}</span>
      <span className="stat-card__value">{value}</span>
      {delta !== undefined && (
        <span className="stat-card__delta" aria-label={`delta: ${formatDelta(delta)}`}>
          {formatDelta(delta)}
        </span>
      )}
    </Card>
  );
}
