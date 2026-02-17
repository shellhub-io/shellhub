import { useEffect, useState } from "react";
import { Outlet, Navigate, useLocation } from "react-router-dom";
import { getInfo } from "../../api/system";
import { env } from "../../env";

export default function SetupGuard() {
  const [loading, setLoading] = useState(true);
  const [setupDone, setSetupDone] = useState(true);
  const location = useLocation();

  useEffect(() => {
    if (env.isCloud) {
      setLoading(false);
      return;
    }

    getInfo()
      .then((info) => setSetupDone(info.setup))
      .catch(() => setSetupDone(true))
      .finally(() => setLoading(false));
  }, []);

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <div className="flex items-center gap-3">
          <span className="w-4 h-4 border-2 border-primary/30 border-t-primary rounded-full animate-spin" />
          <span className="text-xs font-mono text-text-muted">Loading...</span>
        </div>
      </div>
    );
  }

  const isSetupPage = location.pathname === "/setup";

  if (!setupDone && !isSetupPage) {
    return <Navigate to="/setup" replace />;
  }

  if (setupDone && isSetupPage) {
    return <Navigate to="/login" replace />;
  }

  return <Outlet />;
}
