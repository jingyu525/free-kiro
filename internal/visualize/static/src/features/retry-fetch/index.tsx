// features/retry-fetch — red error banner with retry button.

import { Button } from '@shared/ui/Button';
import { isApiError } from '@shared/api/client';

interface RetryBannerProps {
  error: unknown;
  onRetry: () => void;
}

const LABEL: Record<string, string> = {
  network: 'Network error',
  http4xx: 'HTTP error',
  http5xx: 'Server error',
  parse: 'Invalid response',
};

function describeError(error: unknown): string {
  if (isApiError(error)) return LABEL[error.kind] ?? error.message;
  if (error instanceof Error) return error.message;
  return 'Unknown error';
}

export function RetryBanner({ error, onRetry }: RetryBannerProps): JSX.Element {
  return (
    <div className="error-banner" role="alert" aria-live="polite">
      <span className="error-banner-text">
        Couldn't refresh dashboard. <small>({describeError(error)})</small>
      </span>
      <Button variant="danger" onClick={onRetry}>
        retry
      </Button>
    </div>
  );
}