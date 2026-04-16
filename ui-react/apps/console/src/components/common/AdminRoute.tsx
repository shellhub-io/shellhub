import { useEffect, useState } from "react";
import { Outlet, Navigate } from "react-router-dom";
import { useAuthStore } from "@/stores/authStore";

export default function AdminRoute() {
  const fetchUser = useAuthStore((s) => s.fetchUser);
  const isAdmin = useAuthStore((s) => s.isAdmin);
  const [verified, setVerified] = useState(false);

  useEffect(() => {
    void fetchUser().finally(() => setVerified(true));
  }, [fetchUser]);

  if (!verified) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <div className="w-5 h-5 border-2 border-primary border-t-transparent rounded-full animate-spin" />
      </div>
    );
  }

  if (!isAdmin) {
    return <Navigate to="/admin/unauthorized" replace />;
  }

  return <Outlet />;
}
