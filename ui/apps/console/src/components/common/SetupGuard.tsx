import { useEffect, useState } from "react";
import { Outlet, Navigate, useLocation } from "react-router-dom";
import { getInfo } from "@/client";
import { features } from "@/edition";
import { useAuthStore } from "@/stores/authStore";
import { isTokenExpired } from "@/utils/authToken";
import ScreenLoader from "./ScreenLoader";

/**
 * Holds the app until the instance has been set up, sending an unconfigured one to the setup
 * wizard. Skipped on cloud, where the instance is set up before anyone reaches it.
 */
export default function SetupGuard() {
  const { setupWizard } = features();
  const [loading, setLoading] = useState(setupWizard);
  const [setupDone, setSetupDone] = useState(true);
  const location = useLocation();
  const token = useAuthStore((s) => s.token);

  useEffect(() => {
    if (!setupWizard) return;

    const controller = new AbortController();
    getInfo({ throwOnError: true, signal: controller.signal })
      .then(({ data }) => {
        if (!controller.signal.aborted) setSetupDone(data.setup);
      })
      .catch(() => {
        if (!controller.signal.aborted) setSetupDone(true);
      })
      .finally(() => {
        if (!controller.signal.aborted) setLoading(false);
      });

    return () => controller.abort();
  }, [setupWizard, location.pathname]);

  const authed = token !== null && !isTokenExpired(token);

  if (loading && !authed) {
    return <ScreenLoader label="Loading…" />;
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
