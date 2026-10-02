// shared/ui/Skeleton — generic skeleton placeholder block.

interface SkeletonProps {
  width?: string | number;
  height?: string | number;
  rounded?: boolean;
}

export function Skeleton({
  width = '100%',
  height = '1em',
  rounded = false,
}: SkeletonProps): JSX.Element {
  const style: React.CSSProperties = {
    width: typeof width === 'number' ? `${width}px` : width,
    height: typeof height === 'number' ? `${height}px` : height,
    borderRadius: rounded ? '999px' : undefined,
  };
  return <span className="skeleton-block" style={style} aria-hidden="true" />;
}