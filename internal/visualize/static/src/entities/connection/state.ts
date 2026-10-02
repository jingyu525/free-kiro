// entities/connection/state — SSE connection state machine.
//
// States:
//   - 'online': last event < 30s ago
//   - 'reconnecting': last attempt failed, waiting for retry
//   - 'offline': 30s+ of silence; force-reconnect in progress

export type ConnectionState = 'online' | 'reconnecting' | 'offline';

export interface ConnectionStatus {
  state: ConnectionState;
  attempt: number;
  lastEventAt: number;
}

export const initialConnection: ConnectionStatus = {
  state: 'reconnecting',
  attempt: 0,
  lastEventAt: 0,
};

export type ConnectionAction =
  | { type: 'open' }
  | { type: 'event' } // any refresh/heartbeat
  | { type: 'error' }
  | { type: 'offline' }
  | { type: 'reconnected' };

export function connectionReducer(
  state: ConnectionStatus,
  action: ConnectionAction,
): ConnectionStatus {
  switch (action.type) {
    case 'open':
      return { ...state, state: 'online', attempt: 0 };
    case 'event':
      return { ...state, state: 'online', lastEventAt: Date.now() };
    case 'error':
      return {
        ...state,
        state: 'reconnecting',
        attempt: state.attempt + 1,
      };
    case 'offline':
      return { ...state, state: 'offline' };
    case 'reconnected':
      return { ...state, state: 'online', attempt: 0 };
    default:
      return state;
  }
}