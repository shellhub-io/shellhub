import { Outlet, Navigate } from "react-router-dom";
import { useAuthStore } from "@/stores/authStore";
import ErrorBoundary from "./ErrorBoundary";

/**
 * Gates an admin route. A non-admin is redirected rather than shown an empty page, and the
 * redirect replaces the history entry so Back does not return to a page they cannot see.
 */
export default function AdminRoute() {
  const isAdmin = useAuthStore((s) => s.isAdmin);

  if (!isAdmin) {
    return <Navigate to="/admin/unauthorized" replace />;
  }

  return (
    <ErrorBoundary>
      <Outlet />
    </ErrorBoundary>
  );
}
