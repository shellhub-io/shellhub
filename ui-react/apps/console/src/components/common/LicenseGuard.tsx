import { Outlet, Navigate } from "react-router-dom";
import { useAdminLicense } from "../../hooks/useAdminLicense";

export default function LicenseGuard() {
  const { data: license, isLoading, isError } = useAdminLicense();

  if (isLoading) {
    return (
      <div
        className="flex-1 flex items-center justify-center"
        role="status"
        aria-label="Loading license information"
      >
        <div className="flex items-center gap-3">
          <span className="w-4 h-4 border-2 border-primary/30 border-t-primary rounded-full animate-spin" />
          <span className="text-xs font-mono text-text-muted">
            Checking license...
          </span>
        </div>
      </div>
    );
  }

  // Expired, missing, or error -> redirect to license page
  if (isError || !license || license.expired) {
    return <Navigate to="/admin/license" replace />;
  }

  return <Outlet />;
}
