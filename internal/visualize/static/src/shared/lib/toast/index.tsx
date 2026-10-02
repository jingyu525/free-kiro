// shared/lib/toast — global toast Context.
//
// Lives in shared so pages can call useToast without violating FSD's
// top-down import direction (pages → app is forbidden).

import {
  createContext,
  useCallback,
  useContext,
  useMemo,
  useReducer,
  type ReactNode,
} from 'react';

export type ToastKind = 'info' | 'error';

export interface ToastItem {
  id: string;
  kind: ToastKind;
  message: string;
  createdAt: number;
}

interface State {
  toasts: ReadonlyArray<ToastItem>;
}

type Action =
  | { type: 'push'; toast: ToastItem }
  | { type: 'dismiss'; id: string };

function reducer(state: State, action: Action): State {
  switch (action.type) {
    case 'push':
      return { toasts: [...state.toasts, action.toast] };
    case 'dismiss':
      return { toasts: state.toasts.filter((t) => t.id !== action.id) };
    default:
      return state;
  }
}

interface ToastContextApi {
  toasts: ReadonlyArray<ToastItem>;
  pushToast: (kind: ToastKind, message: string) => void;
  dismissToast: (id: string) => void;
}

const ToastContext = createContext<ToastContextApi | null>(null);

let counter = 0;
function nextId(): string {
  counter += 1;
  return `t${Date.now()}-${counter}`;
}

export function ToastProvider({ children }: { children: ReactNode }): JSX.Element {
  const [state, dispatch] = useReducer(reducer, { toasts: [] });

  const pushToast = useCallback((kind: ToastKind, message: string): void => {
    dispatch({
      type: 'push',
      toast: { id: nextId(), kind, message, createdAt: Date.now() },
    });
  }, []);

  const dismissToast = useCallback((id: string): void => {
    dispatch({ type: 'dismiss', id });
  }, []);

  const value = useMemo<ToastContextApi>(
    () => ({ toasts: state.toasts, pushToast, dismissToast }),
    [state.toasts, pushToast, dismissToast],
  );

  return <ToastContext.Provider value={value}>{children}</ToastContext.Provider>;
}

export function useToast(): ToastContextApi {
  const ctx = useContext(ToastContext);
  if (!ctx) throw new Error('useToast must be used inside <ToastProvider>');
  return ctx;
}