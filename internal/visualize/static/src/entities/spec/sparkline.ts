// entities/spec/sparkline — derive a key's history from timeline events.
//
// The backend /api/spec/<name>/timeline endpoint returns
// Array<{ts, kind, message}>. We don't have a structured "key" field, so we
// parse messages loosely for "<key>=<number>" patterns and accumulate them
// in chronological order. If no matches are found we return empty samples
// (the widget renders a "—" placeholder per AC-12).

import { useQuery } from '@tanstack/react-query';
import { fetchSpecTimeline } from '@entities/spec/api';
import type { SpecTimelineEvent } from '@entities/spec/types';

export interface SparklineData {
  samples: number[];
  latest: number;
  delta: number; // latest - samples[0] (0 when fewer than 2 samples)
}

const PATTERN = /(?:^|[\s,;])([A-Za-z0-9_-]+)=(-?\d+(?:\.\d+)?)/g;
const MAX_SAMPLES = 60;

function parseTimeline(
  events: ReadonlyArray<SpecTimelineEvent>,
  key: string,
): SparklineData {
  const escapedKey = key.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
  const re = new RegExp(`(?:^|[\\s,;])${escapedKey}=(-?\\d+(?:\\.\\d+)?)`, 'g');
  const samples: number[] = [];
  for (const e of events) {
    let m: RegExpExecArray | null;
    re.lastIndex = 0;
    while ((m = re.exec(e.message)) !== null) {
      const v = Number(m[1]);
      if (Number.isFinite(v)) samples.push(v);
    }
    if (samples.length === 0) {
      // Fallback: try the generic KEY=VALUE form for any "key=" mention.
      PATTERN.lastIndex = 0;
      let g: RegExpExecArray | null;
      while ((g = PATTERN.exec(e.message)) !== null) {
        if (g[1] === key) {
          const v = Number(g[2]);
          if (Number.isFinite(v)) samples.push(v);
        }
      }
    }
  }
  const tail = samples.slice(-MAX_SAMPLES);
  const latest = tail.length > 0 ? tail[tail.length - 1]! : 0;
  const delta = tail.length >= 2 ? latest - tail[0]! : 0;
  return { samples: tail, latest, delta };
}

export function useSparklineData(name: string, key: string): SparklineData {
  const { data } = useQuery({
    queryKey: ['spec', name, 'timeline'],
    queryFn: () => fetchSpecTimeline(name),
    enabled: Boolean(name && key),
    staleTime: 30_000,
  });
  const events = Array.isArray(data) ? data : [];
  return parseTimeline(events, key);
}

export const __testing__ = { parseTimeline };