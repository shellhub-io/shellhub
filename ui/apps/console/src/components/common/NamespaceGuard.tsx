import { useEffect } from "react";
import { Outlet, useLocation } from "react-router-dom";
import { ArrowPathIcon } from "@heroicons/react/24/outline";
import { Button, Spinner } from "@shellhub/design-system/primitives";
import { useNamespaces, useInitRole } from "@/hooks/useNamespaces";
import { useConnectivityStore } from "@/stores/connectivityStore";
import { isAdminPath } from "@/utils/adminRoute";
import { isAccountPath } from "@/utils/accountRoute";
import { isPreferencesPath } from "@/utils/preferencesRoute";
import NamespaceTrail from "@/components/firstRun/NamespaceTrail";
import FramedShell from "@/components/layout/FramedShell";
import ScreenIntro from "@/components/layout/ScreenIntro";
import AuthActions from "@/components/auth/AuthActions";

function FetchErrorPage({
  error,
  onRetry,
}: {
  error: string;
  onRetry: () => void;
}) {
  return (
    <FramedShell>
      <div role="alert">
        <ScreenIntro
          eyebrow="Something went wrong"
          title="Could not load namespaces"
          lead={`${error}. This is likely temporary. Check your connection or try again.`}
        />
      </div>
      <AuthActions
        primary={
          <Button
            size="lg"
            fullWidth
            icon={<ArrowPathIcon className="w-4 h-4" strokeWidth={2} />}
            onClick={onRetry}
          >
            Try again
          </Button>
        }
      />
    </FramedShell>
  );
}

/**
 * Holds a route until the user has a namespace, and offers to create or join one when they do
 * not. Everything below assumes a tenant, so this is where that assumption is established. The
 * admin console shares the layout but is instance-wide, so /admin passes through ungated.
 */
export default function NamespaceGuard() {
  useInitRole();
  const { namespaces, isLoading, error, refetch } = useNamespaces();
  const apiReachable = useConnectivityStore((s) => s.apiReachable);
  const { pathname } = useLocation();

  useEffect(() => {
    if (apiReachable && error) {
      void refetch();
    }
  }, [apiReachable, error, refetch]);

  if (isAdminPath(pathname)) return <Outlet />;

  if (error && !isLoading) {
    const apiDown = !apiReachable;
    const message = apiDown
      ? "Unable to reach the API"
      : "Failed to load namespaces";
    return <FetchErrorPage error={message} onRetry={() => void refetch()} />;
  }

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <div className="flex items-center gap-3">
          <Spinner />
          <span className="text-xs font-mono text-text-muted">Loading…</span>
        </div>
      </div>
    );
  }

  if (
    namespaces.length === 0 &&
    !isAccountPath(pathname) &&
    !isPreferencesPath(pathname)
  ) {
    return <NamespaceTrail />;
  }

  return <Outlet />;
}
