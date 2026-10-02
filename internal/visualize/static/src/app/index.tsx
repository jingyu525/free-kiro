// app/index — React 18 root mount with all providers.

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
import { SpecDetailDialog } from '@widgets/spec-detail-dialog';

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
                <SpecDetailDialog />
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