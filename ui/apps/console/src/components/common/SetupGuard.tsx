import { useEffect, useState } from "react";
import { Outlet, Navigate, useLocation } from "react-router-dom";
import { getInfo } from "@/client";
import { isCloud } from "@/env";
import { useAuthStore } from "@/stores/authStore";
import { Spinner } from "@shellhub/design-system/primitives";

export default function SetupGuard() {
  const isCloudEdition = isCloud();
  const [loading, setLoading] = useState(!isCloudEdition);
  const [setupDone, setSetupDone] = useState(true);
  const location = useLocation();
  const token = useAuthStore((s) => s.token);

  useEffect(() => {
    if (isCloudEdition) return;

    getInfo({ throwOnError: true })
      .then(({ data }) => setSetupDone(data.setup))
      .catch(() => setSetupDone(true))
      .finally(() => setLoading(false));
  }, [isCloudEdition, location.pathname]);

  const authed = !!token;

  if (loading && !authed) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <div className="flex items-center gap-3">
          <Spinner />
          <span className="text-xs font-mono text-text-muted">Loading...</span>
        </div>
      </div>
    );
  }

  const isSetupPage = location.pathname === "/setup";

  // Sending the user TO setup: an authenticated session means setup is already complete (you
  // can't get a token otherwise), so trust it and don't bounce a just-authenticated user back
  // to /setup while the getInfo refresh is still in flight.
  if (!setupDone && !authed && !isSetupPage) {
    return <Navigate to="/setup" replace />;
  }

  // Sending the user AWAY from /setup: key only on the real setup state, never the token.
  // loginWithToken sets the token synchronously (before the "Instance ready" screen renders),
  // so keying this on the token would fire mid-auto-login and force a trip through /login —
  // exactly the second login round-trip this feature removes.
  if (setupDone && isSetupPage) {
    return <Navigate to="/login" replace />;
  }

  return <Outlet />;
}
