// shared/sse/use-sse-subscription — bridge SSE events to react-query invalidation.
//
// Mounts in <App> (or equivalent). On every 'refresh' SSE event, invalidates
// the provided query keys so react-query refetches. Replaces the
// vanilla-era loadAll() in dashboard-frontend-foundation.
//
// Note: this hook lives in shared/sse (not in app) so it can be imported
// without violating FSD layers. To avoid a shared→entities reverse import,
// callers pass in the query keys they want invalidated.

import { useEffect } from 'react';
import { useQueryClient, type QueryKey } from '@tanstack/react-query';
import { SSEClient } from '@shared/sse/client';

export function useSSESubscription(keysToInvalidate: ReadonlyArray<QueryKey>): void {
  const qc = useQueryClient();
  useEffect(() => {
    const client = new SSEClient();
    const off = client.subscribe((e) => {
      if (e.name === 'refresh') {
        for (const key of keysToInvalidate) {
          void qc.invalidateQueries({ queryKey: key });
        }
      }
    });
    client.connect();
    return () => {
      off();
      client.disconnect();
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [qc]);
}