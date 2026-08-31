import { Component, type ReactNode } from "react";
import { ExclamationTriangleIcon } from "@heroicons/react/24/outline";
import { Button } from "@shellhub/design-system/primitives";

interface Props {
  children: ReactNode;
}

interface State {
  error: Error | null;
}

/**
 * Catches a render error below it and shows a recoverable screen rather than a blank page. React
 * unmounts the whole tree on an uncaught error, so without this a fault in one panel takes the
 * console with it.
 */
export default class ErrorBoundary extends Component<Props, State> {
  state: State = { error: null };

  /**
   * Moves the boundary into its error state. React calls this during the render phase, so it must
   * stay pure — logging and reporting belong in componentDidCatch.
   */
  static getDerivedStateFromError(error: Error): State {
    return { error };
  }

  /**
   * Renders the children, or the fallback once an error has been caught.
   */
  render() {
    const { error } = this.state;
    if (!error) return this.props.children;

    return (
      <div className="flex items-center justify-center min-h-screen bg-background p-8">
        <div className="max-w-md w-full bg-surface border border-border rounded-lg p-8 text-center">
          <div className="w-12 h-12 mx-auto mb-4 rounded-full bg-accent-red/10 flex items-center justify-center">
            <ExclamationTriangleIcon className="w-6 h-6 text-accent-red" />
          </div>
          <h1 className="text-lg font-semibold text-text-primary mb-2">
            Something went wrong
          </h1>
          <p className="text-sm text-text-secondary mb-6">{error.message}</p>
          <Button onClick={() => window.location.reload()}>Reload page</Button>
        </div>
      </div>
    );
  }
}
