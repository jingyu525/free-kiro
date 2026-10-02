// api.ts — fetch wrapper that classifies failures into three kinds
// (network / http / parse) so the UI can pick the right banner colour
// without re-implementing instanceof checks at every callsite.

import { ApiError } from '../types';

const RETRY_DELAYS_MS = [1000, 2000, 4000, 8000] as const;

export async function fetchJson<T>(url: string, init?: RequestInit): Promise<T> {
  let res: Response;
  try {
    res = await fetch(url, {
      ...init,
      headers: { Accept: 'application/json', ...init?.headers },
    });
  } catch (e) {
    // TypeError covers DNS failure, connection refused, CORS preflight
    // rejection, etc. Anything that means "we never heard back".
    throw new ApiError('network', (e as Error).message || 'network error');
  }
  if (!res.ok) {
    throw new ApiError('http', `${res.status} ${res.statusText}`, res.status);
  }
  try {
    return (await res.json()) as T;
  } catch (e) {
    throw new ApiError('parse', (e as Error).message || 'invalid JSON');
  }
}

// fetchRetry retries on network + 5xx errors with exponential
// backoff (1s → 2s → 4s → 8s, then gives up). HTTP 4xx is not
// retried — caller asked for something the server explicitly
// rejected, retrying won't change the answer.
export async function fetchRetry<T>(
  url: string,
  init?: RequestInit,
  onAttempt?: (attempt: number) => void,
): Promise<T> {
  let lastErr: ApiError | null = null;
  for (let attempt = 0; attempt <= RETRY_DELAYS_MS.length; attempt++) {
    onAttempt?.(attempt);
    try {
      return await fetchJson<T>(url, init);
    } catch (e) {
      if (!(e instanceof ApiError)) throw e;
      lastErr = e;
      const retryable = e.kind === 'network' || (e.status !== undefined && e.status >= 500);
      if (!retryable || attempt === RETRY_DELAYS_MS.length) throw e;
      await sleep(RETRY_DELAYS_MS[attempt]!);
    }
  }
  // Unreachable: the loop either returns or throws.
  throw lastErr ?? new ApiError('network', 'retry loop exited unexpectedly');
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}
