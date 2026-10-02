// widgets/header-bar — page banner with title, live meta row, refresh, theme toggle, and SSE conn dot.

import { useEffect, useState } from 'react';
import { useSummaryQuery } from '@entities/summary/use-summary-query';
import { useConnection } from '@entities/connection/provider';
import { formatRelative, formatTimeHHMMSS } from '@shared/lib/format';
import { RefreshButton } from '@features/refresh-dashboard';
import { ThemeToggle } from '@features/theme-toggle';

export function HeaderBar(): JSX.Element {
  const { data } = useSummaryQuery();
  const conn = useConnection();

  const specs = data?.specs ?? [];
  const lastRefreshAt = data?.last_refresh_at
    ? new Date(data.last_refresh_at).getTime()
    : Date.now();

  const [tick, setTick] = useState(0);

  useEffect(() => {
    const id = setInterval(() => setTick((n) => n + 1), 1000);
    return () => clearInterval(id);
  }, []);

  void tick; // suppress unused var lint warning

  const connLabel: Record<typeof conn.state, string> = {
    online: 'connected',
    reconnecting: 'reconnecting…',
    offline: 'disconnected',
  };

  return (
    <header role="banner" className="header-bar">
      <div className="header-bar__title-row">
        <h1>free-kiro dashboard</h1>
        <div className="header-bar__actions">
          <RefreshButton />
          <ThemeToggle />
          <span
            className="conn-dot"
            data-state={conn.state}
            title={connLabel[conn.state]}
            aria-label={`connection: ${connLabel[conn.state]}`}
          />
        </div>
      </div>
      <p className="header-bar__meta">
        {specs.length} specs &middot; {formatRelative(lastRefreshAt)} &middot; generated {formatTimeHHMMSS(new Date())}
      </p>
    </header>
  );
}
