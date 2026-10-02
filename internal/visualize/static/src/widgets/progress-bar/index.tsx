// widgets/progress-bar — slim top-bar progress indicator driven by active prop.

interface ProgressBarProps {
  active: boolean;
}

export function ProgressBar({ active }: ProgressBarProps): JSX.Element {
  return (
    <div
      className={`top-bar-progress${active ? ' active' : ''}`}
      role="progressbar"
      aria-label="Loading"
      aria-busy={active}
    />
  );
}
