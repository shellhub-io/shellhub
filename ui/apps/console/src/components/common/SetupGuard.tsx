import { useEffect, useState } from "react";
import { Outlet, Navigate, useLocation } from "react-router-dom";
import { getInfo } from "@/client";
import { isCloud } from "@/env";
import { useAuthStore } from "@/stores/authStore";
import { Spinner } from "@shellhub/design-system/primitives";

/**
 * Holds the app until the instance has been set up, sending an unconfigured one to the setup
 * wizard. Skipped on cloud, where the instance is set up before anyone reaches it.
 */
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

  if (!setupDone && !authed && !isSetupPage) {
    return <Navigate to="/setup" replace />;
  }

  if (setupDone && isSetupPage) {
    return <Navigate to="/login" replace />;
  }

  return <Outlet />;
}
