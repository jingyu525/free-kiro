// stores/connection.ts — SSE connection state. Drives the conn-dot
// indicator in HeaderBar.

import { Signal } from '../lib/signal';

export type ConnectionState = 'online' | 'reconnecting' | 'offline';

export interface ConnectionStateValue {
  state: ConnectionState;
  attempt: number;
  lastEventAt: number;
}

export const connectionStore = new Signal<ConnectionStateValue>({
  state: 'offline',
  attempt: 0,
  lastEventAt: 0,
});
