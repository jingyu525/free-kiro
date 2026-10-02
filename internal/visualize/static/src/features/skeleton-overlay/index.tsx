// features/skeleton-overlay — first-load skeleton; defers to widgets/skeleton-row.

import { SkeletonRow } from '@shared/ui/SkeletonRow';

interface SkeletonOverlayProps {
  rows?: number;
  visible: boolean;
}

export function SkeletonOverlay({ rows = 6, visible }: SkeletonOverlayProps): JSX.Element | null {
  if (!visible) return null;
  return (
    <div className="skeleton-overlay" role="status" aria-label="Loading dashboard">
      {Array.from({ length: rows }).map((_, i) => (
        <SkeletonRow key={i} />
      ))}
    </div>
  );
}