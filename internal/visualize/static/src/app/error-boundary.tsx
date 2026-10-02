// app/error-boundary — React class error boundary.
//
// Catches uncaught render errors so the dashboard shows a fallback UI
// instead of going blank. The dashboard-frontend-foundation MVP shipped
// with a known TypeError ("Cannot read properties of null") that crashes
// the whole tree; tracking that bug is out of scope for
// dashboard-frontend-react-vite-fsd (assigned to dashboard-frontend-components),
// so this boundary exists to keep the rest of the UI visible.

import { Component, type ReactNode } from 'react';

interface Props {
  children: ReactNode;
}

interface State {
  error: Error | null;
}

export class ErrorBoundary extends Component<Props, State> {
  override state: State = { error: null };

  static getDerivedStateFromError(error: Error): State {
    return { error };
  }

  override componentDidCatch(error: Error, info: { componentStack?: string }): void {
    console.error('[ErrorBoundary]', error, info.componentStack ?? '');
  }

  override render(): ReactNode {
    if (this.state.error) {
      return (
        <div className="error-fallback" role="alert" aria-live="assertive">
          <h2>Dashboard failed to render.</h2>
          <p>
              Reload the page. If the issue persists, check the console.
            </p>
          <pre>{String(this.state.error.message)}</pre>
        </div>
      );
    }
    return this.props.children;
  }
}