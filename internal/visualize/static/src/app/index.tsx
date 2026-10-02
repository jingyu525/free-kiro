// app/index — React 18 root mount with all providers.
//
// SpecDetailDialog is intentionally NOT mounted at runtime in this build:
// react-query 5 + dialog's useSyncExternalStore getSnapshot triggers a
// "Maximum update depth exceeded" loop in React 18 prod minified. The
// dialog widget code lives at widgets/spec-detail-dialog/* for future fix.
// See dashboard-spec-detail-view spec for known-issue tracking.

import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { ConnectionProvider } from '@entities/connection/provider';
import { useSSESubscription } from '@shared/sse/use-sse-subscription';
import { QueryProvider } from '@app/providers/query-provider';
import { ThemeProvider } from '@shared/lib/theme';
import { ToastProvider } from '@shared/lib/toast';
import { AppRouter } from '@app/router';
import { SUMMARY_QUERY_KEY } from '@entities/summary/use-summary-query';
import { ErrorBoundary } from '@app/error-boundary';

import '@shared/styles/globals.css';

function SSEBridge(): null {
  useSSESubscription([SUMMARY_QUERY_KEY, ['specs']]);
  return null;
}

function Root(): JSX.Element {
  return (
    <StrictMode>
      <QueryProvider>
        <ThemeProvider>
          <ConnectionProvider>
            <ToastProvider>
              <a href="#main" className="skip-link">
                Skip to main content
              </a>
              <ErrorBoundary>
                <SSEBridge />
                <AppRouter />
                {/* SpecDetailDialog temporarily unmounted — see file header. */}
              </ErrorBoundary>
            </ToastProvider>
          </ConnectionProvider>
        </ThemeProvider>
      </QueryProvider>
    </StrictMode>
  );
}

const host = document.getElementById('app');
if (!host) throw new Error('Mount point #app not found in document');
createRoot(host).render(<Root />);