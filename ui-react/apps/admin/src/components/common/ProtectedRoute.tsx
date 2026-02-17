import { useEffect } from "react";
import { Outlet, Navigate } from "react-router-dom";
import { useAuthStore } from "../../stores/authStore";

export default function ProtectedRoute() {
  const { isLoggedIn, restoreSession } = useAuthStore();

  useEffect(() => {
    restoreSession();
  }, [restoreSession]);

  const token = localStorage.getItem("token");

  if (!isLoggedIn && !token) {
    return <Navigate to="/login" replace />;
  }

  return <Outlet />;
}
