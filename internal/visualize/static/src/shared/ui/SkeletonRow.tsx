// shared/ui/SkeletonRow — animated gradient placeholder row during initial data fetch.

import { Skeleton } from '@shared/ui/Skeleton';

export function SkeletonRow(): JSX.Element {
  return (
    <div className="skeleton-row" aria-hidden="true">
      <Skeleton width="40%" height="1.2em" />
      <Skeleton width="12%" height="1.2em" />
      <Skeleton width="8%" height="1.2em" />
      <Skeleton width="10%" height="1.2em" />
    </div>
  );
}