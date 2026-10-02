// features/refresh-dashboard — manual refresh button + Ctrl/Cmd+R interception.

import { useEffect, useRef, useState } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { Button } from '@shared/ui/Button';
import { SUMMARY_QUERY_KEY } from '@entities/summary/use-summary-query';

const DEBOUNCE_MS = 500;

export function RefreshButton(): JSX.Element {
  const qc = useQueryClient();
  const [spinning, setSpinning] = useState(false);
  const lastClickRef = useRef(0);

  const doRefresh = async (): Promise<void> => {
    setSpinning(true);
    try {
      await qc.invalidateQueries({ queryKey: SUMMARY_QUERY_KEY });
      await qc.refetchQueries({ queryKey: SUMMARY_QUERY_KEY });
    } finally {
      setSpinning(false);
    }
  };

  const onClick = (): void => {
    const now = Date.now();
    if (now - lastClickRef.current < DEBOUNCE_MS) return;
    lastClickRef.current = now;
    void doRefresh();
  };

  useEffect(() => {
    const handler = (e: KeyboardEvent): void => {
      if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'r') {
        e.preventDefault();
        void doRefresh();
      }
    };
    window.addEventListener('keydown', handler);
    return () => window.removeEventListener('keydown', handler);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return (
    <Button variant="ghost" onClick={onClick} spinning={spinning} aria-label="Refresh dashboard">
      <span aria-hidden="true">↻</span> refresh
    </Button>
  );
}