// widgets/sparkline-cell — minimal SVG sparkline from SparklineData.

import { useSparklineData } from '@entities/spec/sparkline';

interface SparklineCellProps {
  name: string;
  key: string;
}

const W = 60;
const H = 16;
const PAD = 1;

function pathFor(samples: ReadonlyArray<number>): string {
  if (samples.length === 0) return '';
  const lo = Math.min(...samples);
  const hi = Math.max(...samples);
  const span = hi - lo || 1;
  const stepX = (W - 2 * PAD) / Math.max(samples.length - 1, 1);
  return samples
    .map((v, i) => {
      const x = PAD + i * stepX;
      const y = H - PAD - ((v - lo) / span) * (H - 2 * PAD);
      return `${i === 0 ? 'M' : 'L'}${x.toFixed(2)},${y.toFixed(2)}`;
    })
    .join(' ');
}

export function SparklineCell({ name, key }: SparklineCellProps): JSX.Element {
  const data = useSparklineData(name, key);
  if (data.samples.length === 0) {
    return (
      <span className="sparkline-empty" aria-label={`no history for ${key}`}>
        —
      </span>
    );
  }
  const trend =
    data.delta > 0 ? 'up' : data.delta < 0 ? 'down' : 'flat';
  const last = data.samples[data.samples.length - 1]!;
  const lastX = W - PAD;
  const lastY = ((): number => {
    const lo = Math.min(...data.samples);
    const hi = Math.max(...data.samples);
    const span = hi - lo || 1;
    return H - PAD - ((last - lo) / span) * (H - 2 * PAD);
  })();
  const ariaLabel = `history of ${key}: ${data.samples.length} samples, latest delta ${data.delta}`;
  return (
    <svg
      className={`sparkline sparkline-${trend}`}
      viewBox={`0 0 ${W} ${H}`}
      width={W}
      height={H}
      role="img"
      aria-label={ariaLabel}
    >
      <path d={pathFor(data.samples)} fill="none" stroke="currentColor" strokeWidth={1} />
      <circle cx={lastX} cy={lastY} r={2} fill="currentColor" />
    </svg>
  );
}