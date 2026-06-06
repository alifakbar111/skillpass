import { Component, type ErrorInfo, type ReactNode } from 'react';

interface Props {
  children: ReactNode;
  fallback?: ReactNode;
}

interface State {
  hasError: boolean;
  error: Error | null;
}

export class ErrorBoundary extends Component<Props, State> {
  state: State = { hasError: false, error: null };

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  override componentDidCatch(error: Error, info: ErrorInfo) {
    console.error('ErrorBoundary caught:', error, info);
  }

  reset = () => {
    this.setState({ hasError: false, error: null });
  };

  reload = () => {
    this.reset();
    window.location.reload();
  };

  override render() {
    if (this.state.hasError) {
      if (this.props.fallback) {
        return this.props.fallback;
      }
      return (
        <div className="flex items-center justify-center min-h-[50vh]">
          <div className="text-center space-y-3 max-w-sm">
            <h2 className="text-xl font-bold">Something went wrong</h2>
            <p className="text-sm opacity-70">{this.state.error?.message ?? 'An unexpected error occurred.'}</p>
            <div className="flex gap-2 justify-center">
              <button type="button" className="btn btn-primary btn-sm" onClick={this.reload}>
                Reload page
              </button>
              <button type="button" className="btn btn-ghost btn-sm" onClick={() => window.history.back()}>
                Go back
              </button>
            </div>
          </div>
        </div>
      );
    }
    return this.props.children;
  }
}
