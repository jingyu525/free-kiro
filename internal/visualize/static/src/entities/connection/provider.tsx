// entities/connection/provider — React Context exposing SSE status.
//
// Owns the SSEClient singleton and feeds its events into a reducer.
// Consumers (header-bar) read status via useConnection().

import {
  createContext,
  useContext,
  useEffect,
  useMemo,
  useReducer,
  type ReactNode,
} from 'react';
import { SSEClient, type SSEEvent } from '@shared/sse/client';
import {
  connectionReducer,
  initialConnection,
  type ConnectionStatus,
} from './state';

interface ConnectionContextValue {
  status: ConnectionStatus;
}

const ConnectionContext = createContext<ConnectionContextValue | null>(null);

export function ConnectionProvider({ children }: { children: ReactNode }): JSX.Element {
  const [status, dispatch] = useReducer(connectionReducer, initialConnection);

  const client = useMemo(() => new SSEClient(), []);

  useEffect(() => {
    const off = client.subscribe((e: SSEEvent) => {
      switch (e.name) {
        case 'open':
          dispatch({ type: 'open' });
          break;
        case 'refresh':
        case 'heartbeat':
          dispatch({ type: 'event' });
          break;
        case 'error':
          dispatch({ type: 'error' });
          break;
      }
    });
    client.connect();
    return () => {
      off();
      client.disconnect();
    };
  }, [client]);

  // 30s silence → mark offline.
  useEffect(() => {
    const id = setInterval(() => {
      if (!status.lastEventAt) return;
      if (Date.now() - status.lastEventAt > 30_000 && status.state !== 'offline') {
        dispatch({ type: 'offline' });
      }
    }, 5000);
    return () => clearInterval(id);
  }, [status.lastEventAt, status.state]);

  const value = useMemo<ConnectionContextValue>(() => ({ status }), [status]);
  return <ConnectionContext.Provider value={value}>{children}</ConnectionContext.Provider>;
}

export function useConnection(): ConnectionStatus {
  const ctx = useContext(ConnectionContext);
  if (!ctx) throw new Error('useConnection must be used inside <ConnectionProvider>');
  return ctx.status;
}