// app/router — hash-based router with single / dashboard route.
//
// Spec detail route (#/spec/<name>) is reserved for dashboard-spec-detail-view.

import { useHashRoute } from '@shared/lib/hash-router';
import { DashboardPage } from '@pages/dashboard-page';

export function AppRouter(): JSX.Element {
  const route = useHashRoute();
  if (route.startsWith('spec/')) {
    // Reserved for dashboard-spec-detail-view; fall through to dashboard for now.
    return <DashboardPage />;
  }
  return <DashboardPage />;
}