// stores/detail.ts — LRU cache of per-spec detail payloads.
// Stays empty in this MVP spec; consumed by the future
// dashboard-spec-detail-view spec for the detail dialog.
//
// Implementation note: we keep just the cache plumbing now so the
// module's contract is documented and the future spec can wire
// fetches without changing component APIs.

import { Signal } from '../lib/signal';

export interface DetailEntry {
  fetchedAt: number;
  data: unknown;
}

const MAX_ENTRIES = 50;
const TTL_MS = 60_000;

const cache = new Map<string, Signal<DetailEntry | null>>();

export function getDetailSignal(spec: string): Signal<DetailEntry | null> {
  let sig = cache.get(spec);
  if (!sig) {
    sig = new Signal<DetailEntry | null>(null);
    cache.set(spec, sig);
    evictIfNeeded();
  }
  return sig;
}

export function isFresh(entry: DetailEntry | null): boolean {
  if (!entry) return false;
  return Date.now() - entry.fetchedAt < TTL_MS;
}

function evictIfNeeded(): void {
  if (cache.size <= MAX_ENTRIES) return;
  // Evict oldest fetchedAt. Map iteration order is insertion order,
  // so we sort by fetchedAt from the values.
  const entries = [...cache.entries()];
  entries.sort((a, b) => (a[1].get()?.fetchedAt ?? 0) - (b[1].get()?.fetchedAt ?? 0));
  const toEvict = entries.slice(0, cache.size - MAX_ENTRIES);
  for (const [key] of toEvict) cache.delete(key);
}
