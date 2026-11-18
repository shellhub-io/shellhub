import { Outlet, Navigate } from "react-router-dom";
import { useAuthStore } from "@/stores/authStore";
import ErrorBoundary from "./ErrorBoundary";

/**
 * Gates a route on being signed in, sending anyone else to the sign-in screen.
 */
export default function ProtectedRoute() {
  const token = useAuthStore((s) => s.token);

  if (!token) {
    return <Navigate to="/login" replace />;
  }

  return (
    <ErrorBoundary>
      <Outlet />
    </ErrorBoundary>
  );
}
