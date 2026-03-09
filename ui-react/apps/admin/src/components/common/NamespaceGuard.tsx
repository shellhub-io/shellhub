import { useEffect } from "react";
import { Outlet, useLocation } from "react-router-dom";
import {
  ExclamationTriangleIcon,
  ArrowPathIcon,
} from "@heroicons/react/24/outline";
import { useNamespacesStore } from "../../stores/namespacesStore";
import { useConnectivityStore } from "../../stores/connectivityStore";
import AmbientBackground from "./AmbientBackground";
import CreateNamespace from "./CreateNamespace";
import UserMenu from "../layout/UserMenu";

function MinimalHeader() {
  return (
    <header className="h-14 bg-surface border-b border-border px-5 flex items-center justify-between shrink-0">
      <img src="/logo.svg" alt="ShellHub" className="h-6" />
      <UserMenu />
    </header>
  );
}

function FetchErrorPage({
  error,
  onRetry,
}: {
  error: string;
  onRetry: () => void;
}) {
  return (
    <div
      className="relative min-h-screen flex flex-col items-center justify-center bg-background overflow-hidden"
      role="alert"
    >
      <AmbientBackground variant="error" />

      {/* Content */}
      <div className="flex flex-col items-center text-center px-6 animate-fade-in">
        <img
          src="/logo.svg"
          alt="ShellHub"
          className="h-8 mb-10 opacity-50"
        />

        <div className="animate-float mb-6">
          <div className="w-20 h-20 rounded-2xl bg-accent-red/10 border border-accent-red/20 flex items-center justify-center shadow-lg shadow-accent-red/5">
            <ExclamationTriangleIcon
              className="w-10 h-10 text-accent-red/60"
              strokeWidth={1.2}
            />
          </div>
        </div>

        <p className="text-2xs font-mono font-semibold uppercase tracking-wide text-accent-red/60 mb-2">
          Something went wrong
        </p>
        <h1 className="text-2xl font-bold text-text-primary mb-3">
          Could not load namespaces
        </h1>
        <p className="text-sm text-text-muted max-w-md leading-relaxed mb-8">
          {error}. This is likely temporary — check your connection or try
          again.
        </p>

        <button
          type="button"
          onClick={onRetry}
          className="inline-flex items-center gap-2.5 px-5 py-3 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold transition-all duration-200"
        >
          <ArrowPathIcon className="w-4 h-4" strokeWidth={2} />
          Try again
        </button>
      </div>
    </div>
  );
}

export default function NamespaceGuard() {
  const { namespaces, loaded, loading, error, fetch } = useNamespacesStore();
  const apiReachable = useConnectivityStore((s) => s.apiReachable);
  const { pathname } = useLocation();

  useEffect(() => {
    if (!loaded) fetch();
  }, [loaded, fetch]);

  // Retry namespace fetch when API comes back online
  useEffect(() => {
    if (apiReachable && !loaded) {
      fetch();
    }
  }, [apiReachable, loaded, fetch]);

  if (!loaded && !loading && error) {
    return <FetchErrorPage error={error} onRetry={fetch} />;
  }

  if (!loaded) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <div className="flex items-center gap-3">
          <span className="w-4 h-4 border-2 border-primary/30 border-t-primary rounded-full animate-spin" />
          <span className="text-xs font-mono text-text-muted">Loading…</span>
        </div>
      </div>
    );
  }

  if (namespaces.length === 0 && pathname !== "/profile") {
    return (
      <div className="min-h-screen bg-background flex flex-col">
        <MinimalHeader />
        <div className="flex-1 flex">
          <CreateNamespace />
        </div>
      </div>
    );
  }

  return <Outlet />;
}
