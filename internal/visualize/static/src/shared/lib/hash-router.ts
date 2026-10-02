// shared/lib/hash-router — read URL hash and subscribe to changes.
//
// Schema: `#/spec/<name>?tab=<tab>` where `<tab>` ∈
// { overview, drift, tasks, timeline } (default: overview).
//
// useHashRoute() returns { name, tab } so consumers don't have to re-parse.
// The returned object is cached per (name, tab) tuple to satisfy React 18's
// useSyncExternalStore contract that getSnapshot returns a stable reference.

import { useSyncExternalStore } from 'react';

export type HashRoute = '' | `spec/${string}`;
export type TabId = 'overview' | 'drift' | 'tasks' | 'timeline';
const VALID_TABS: ReadonlyArray<TabId> = ['overview', 'drift', 'tasks', 'timeline'];

export interface HashRouteState {
  name: string;
  tab: TabId;
}

const EMPTY: HashRouteState = { name: '', tab: 'overview' };
const CACHE = new Map<string, HashRouteState>();

function parseHash(rawHash: string): HashRouteState {
  const stripped = rawHash.replace(/^#\/?/, '');
  if (!stripped.startsWith('spec/')) return EMPTY;
  const [path, query] = stripped.split('?', 2) as [string, string | undefined];
  const name = decodeURIComponent(path.slice('spec/'.length));
  let tab: TabId = 'overview';
  if (query) {
    const params = new URLSearchParams(query);
    const t = params.get('tab');
    if (t && (VALID_TABS as ReadonlyArray<string>).includes(t)) {
      tab = t as TabId;
    }
  }
  const cacheKey = `${name}\u0000${tab}`;
  let cached = CACHE.get(cacheKey);
  if (!cached) {
    cached = { name, tab };
    CACHE.set(cacheKey, cached);
  }
  return cached;
}

function readHash(): HashRouteState {
  return parseHash(window.location.hash);
}

function subscribeHash(fn: () => void): () => void {
  window.addEventListener('hashchange', fn);
  return () => window.removeEventListener('hashchange', fn);
}

/** React hook returning the current { name, tab } parsed from URL hash. */
export function useHashRoute(): HashRouteState {
  return useSyncExternalStore(subscribeHash, readHash, () => EMPTY);
}

/**
 * Programmatic navigation. Accepts either a bare route (legacy `HashRoute`)
 * or a full state `{name, tab}`. Clears the hash when `name === ''`.
 */
export function navigateToHash(target: HashRoute | HashRouteState): void {
  if (typeof target === 'string') {
    window.location.hash = target === '' ? '' : `#/${target}`;
    return;
  }
  if (target.name === '') {
    window.location.hash = '';
    return;
  }
  const enc = encodeURIComponent(target.name);
  const tab = target.tab === 'overview' ? '' : `?tab=${target.tab}`;
  window.location.hash = `#/spec/${enc}${tab}`;
}