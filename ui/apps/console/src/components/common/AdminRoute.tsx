import { Outlet, Navigate } from "react-router-dom";
import { useAuthStore } from "@/stores/authStore";
import { isEnterpriseOrCloud } from "@/env";
import { ADMIN_UNAUTHORIZED_PATH } from "@/utils/adminRoute";
import ErrorBoundary from "./ErrorBoundary";

/**
 * Gates an admin route. Community has no admin API, so it is sent to the home page. A non-admin
 * is redirected rather than shown an empty page. Both redirects replace the history entry so Back
 * does not return to a page they cannot see.
 */
export default function AdminRoute() {
  const isAdmin = useAuthStore((s) => s.isAdmin);

  if (!isEnterpriseOrCloud()) {
    return <Navigate to="/" replace />;
  }

  if (!isAdmin) {
    return <Navigate to={ADMIN_UNAUTHORIZED_PATH} replace />;
  }

  return (
    <ErrorBoundary>
      <Outlet />
    </ErrorBoundary>
  );
}
