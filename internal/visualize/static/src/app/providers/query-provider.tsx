// app/providers/query-provider — react-query wrapper with sane defaults.

import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useMemo, type ReactNode } from 'react';

interface QueryProviderProps {
  children: ReactNode;
}

export function QueryProvider({ children }: QueryProviderProps): JSX.Element {
  const client = useMemo(
    () =>
      new QueryClient({
        defaultOptions: {
          queries: {
            staleTime: 30_000,
            gcTime: 5 * 60_000,
            retry: 3,
            refetchOnWindowFocus: false,
          },
          mutations: { retry: 1 },
        },
      }),
    [],
  );
  return <QueryClientProvider client={client}>{children}</QueryClientProvider>;
}