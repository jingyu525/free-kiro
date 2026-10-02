// shared/lib/hash-router — read URL hash and subscribe to changes.

import { useSyncExternalStore } from 'react';

export type HashRoute = '' | `spec/${string}`;

function readHash(): HashRoute {
  const raw = window.location.hash.replace(/^#\/?/, '');
  if (raw.startsWith('spec/')) return raw as HashRoute;
  return '';
}

function subscribeHash(fn: () => void): () => void {
  window.addEventListener('hashchange', fn);
  return () => window.removeEventListener('hashchange', fn);
}

/** React hook returning the current hash route ('spec/<name>' or ''). */
export function useHashRoute(): HashRoute {
  return useSyncExternalStore(subscribeHash, readHash, () => '');
}

/** Programmatic navigation. */
export function navigateToHash(route: HashRoute): void {
  if (route === '') {
    window.location.hash = '';
  } else {
    window.location.hash = `#/${route}`;
  }
}