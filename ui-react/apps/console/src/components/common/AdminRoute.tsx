import { Outlet, Navigate } from "react-router-dom";
import { useAuthStore } from "../../stores/authStore";

export default function AdminRoute() {
  const isAdmin = useAuthStore((s) => s.isAdmin);

  if (!isAdmin) {
    return <Navigate to="/dashboard" replace />;
  }

  return <Outlet />;
}
