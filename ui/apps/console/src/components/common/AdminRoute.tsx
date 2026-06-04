import { useEffect, useState } from "react";
import { Outlet, Navigate } from "react-router-dom";
import { useAuthStore } from "@/stores/authStore";
import Spinner from "@/components/common/Spinner";

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
        <Spinner size="lg" />
      </div>
    );
  }

  if (!isAdmin) {
    return <Navigate to="/admin/unauthorized" replace />;
  }

  return <Outlet />;
}
