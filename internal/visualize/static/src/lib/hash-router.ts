// hash-router.ts — minimal URL hash router. Phase placeholder for the
// dashboard-spec-detail-view spec — exposes a subscribe API so future
// code can deep-link into a spec detail dialog without depending on
// any particular URL parsing implementation.

export type Route =
  | { name: 'home' }
  | { name: 'spec'; spec: string };

export function parseHash(hash: string): Route {
  const trimmed = hash.replace(/^#/, '').trim();
  if (trimmed === '' || trimmed === '/') return { name: 'home' };
  const m = trimmed.match(/^spec\/(.+)$/);
  if (m && m[1]) return { name: 'spec', spec: decodeURIComponent(m[1]) };
  return { name: 'home' };
}

export function routeHash(route: Route): string {
  if (route.name === 'spec') return `#spec/${encodeURIComponent(route.spec)}`;
  return '';
}

export function subscribeHash(onChange: (route: Route) => void): () => void {
  const handler = () => onChange(parseHash(window.location.hash));
  window.addEventListener('hashchange', handler);
  return () => window.removeEventListener('hashchange', handler);
}
