import { Outlet, Navigate } from "react-router-dom";
import { useAuthStore } from "../../stores/authStore";

export default function ProtectedRoute() {
  const token = useAuthStore((s) => s.token);

  if (!token) {
    return <Navigate to="/login" replace />;
  }

  return <Outlet />;
}
