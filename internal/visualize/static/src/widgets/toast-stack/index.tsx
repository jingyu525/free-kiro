// widgets/toast-stack — renders toasts from useToast context with 5s auto-dismiss.

import { useEffect } from 'react';
import { useToast } from '@shared/lib/toast';

export function ToastStack(): JSX.Element {
  const { toasts, dismissToast } = useToast();

  useEffect(() => {
    if (toasts.length === 0) return;
    const timers = toasts.map((t) =>
      setTimeout(() => dismissToast(t.id), 5000),
    );
    return () => timers.forEach(clearTimeout);
  }, [toasts, dismissToast]);

  if (toasts.length === 0) return <></>;

  return (
    <div className="toast-stack" aria-live="polite" aria-atomic="false">
      {toasts.map((t) => (
        <div
          key={t.id}
          className={`toast toast-${t.kind}`}
          role="alert"
          aria-live="polite"
        >
          {t.message}
        </div>
      ))}
    </div>
  );
}
