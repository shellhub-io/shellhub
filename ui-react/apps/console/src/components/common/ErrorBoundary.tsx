import { Component, type ReactNode } from "react";

interface Props {
  children: ReactNode;
}

interface State {
  error: Error | null;
}

export default class ErrorBoundary extends Component<Props, State> {
  state: State = { error: null };

  static getDerivedStateFromError(error: Error): State {
    return { error };
  }

  render() {
    const { error } = this.state;
    if (!error) return this.props.children;

    return (
      <div className="flex items-center justify-center min-h-screen bg-background p-8">
        <div className="max-w-md w-full bg-surface border border-border rounded-lg p-8 text-center">
          <div className="w-12 h-12 mx-auto mb-4 rounded-full bg-accent-red/10 flex items-center justify-center">
            <svg
              className="w-6 h-6 text-accent-red"
              fill="none"
              viewBox="0 0 24 24"
              strokeWidth={1.5}
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z"
              />
            </svg>
          </div>
          <h1 className="text-lg font-semibold text-text-primary mb-2">
            Something went wrong
          </h1>
          <p className="text-sm text-text-secondary mb-6">{error.message}</p>
          <button
            onClick={() => window.location.reload()}
            className="px-4 py-2 bg-primary text-white text-sm font-medium rounded-lg hover:bg-primary-600 transition-colors"
          >
            Reload page
          </button>
        </div>
      </div>
    );
  }
}
