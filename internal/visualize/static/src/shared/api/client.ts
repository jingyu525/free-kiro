// shared/api/client — fetch wrapper with typed errors and exponential backoff.
//
// Three error classes (see AC-49, AC-50):
//   - ApiTypeError: network failure (fetch rejected)
//   - ApiHttpError: non-2xx response, carries status code
//   - ApiParseError: response body was not valid JSON
//
// AGENT_RULES §2: never swallow errors. fetchRetry surfaces every failure to
// the caller via the final rejected promise.

import { API_BASE } from '@shared/config';

export class ApiTypeError extends Error {
  readonly kind = 'network' as const;
  constructor(cause: unknown) {
    super('network error: fetch failed');
    this.name = 'ApiTypeError';
    this.cause = cause;
  }
}

export class ApiHttpError extends Error {
  readonly kind: 'http4xx' | 'http5xx';
  readonly status: number;
  constructor(status: number, body: string) {
    super(`http ${status}: ${body.slice(0, 200)}`);
    this.name = 'ApiHttpError';
    this.status = status;
    this.kind = status >= 500 ? 'http5xx' : 'http4xx';
  }
}

export class ApiParseError extends Error {
  readonly kind = 'parse' as const;
  constructor(cause: unknown) {
    super('parse error: invalid JSON');
    this.name = 'ApiParseError';
    this.cause = cause;
  }
}

export type ApiError = ApiTypeError | ApiHttpError | ApiParseError;

export function isApiError(e: unknown): e is ApiError {
  return (
    e instanceof ApiTypeError ||
    e instanceof ApiHttpError ||
    e instanceof ApiParseError
  );
}

export interface FetchJsonOptions extends Omit<RequestInit, 'signal'> {
  signal?: AbortSignal;
  /** When false, returns null on 404 instead of throwing. */
  allow404?: boolean;
}

/** Single-shot JSON fetch. Throws an ApiError on any failure. */
export async function fetchJson<T>(path: string, opts: FetchJsonOptions = {}): Promise<T> {
  const url = path.startsWith('http') ? path : `${API_BASE}${path}`;
  let response: Response;
  try {
    response = await fetch(url, {
      ...opts,
      headers: { Accept: 'application/json', ...(opts.headers ?? {}) },
    });
  } catch (cause) {
    throw new ApiTypeError(cause);
  }
  if (!response.ok) {
    if (response.status === 404 && opts.allow404) return null as T;
    const body = await response.text().catch(() => '');
    throw new ApiHttpError(response.status, body);
  }
  try {
    return (await response.json()) as T;
  } catch (cause) {
    throw new ApiParseError(cause);
  }
}

export interface FetchRetryOptions extends FetchJsonOptions {
  maxAttempts?: number;
  /** Base delay in ms; actual delay = baseMs * attempt² (cap 8000). */
  baseMs?: number;
}

/**
 * fetchJson with exponential backoff. Retries network/5xx errors up to
 * maxAttempts; throws the final error on 4xx or after exhaustion.
 */
export async function fetchRetry<T>(path: string, opts: FetchRetryOptions = {}): Promise<T> {
  const max = opts.maxAttempts ?? 4;
  const base = opts.baseMs ?? 1000;
  let lastErr: ApiError | null = null;
  for (let attempt = 1; attempt <= max; attempt++) {
    try {
      return await fetchJson<T>(path, opts);
    } catch (err) {
      if (!isApiError(err)) throw err;
      lastErr = err;
      if (err.kind === 'http4xx') throw err; // never retry 4xx
      if (attempt === max) break;
      const delay = Math.min(base * attempt * attempt, 8000);
      await new Promise((r) => setTimeout(r, delay));
    }
  }
  throw lastErr;
}