import { Outlet, Navigate } from "react-router-dom";
import { useAdminLicense } from "@/hooks/useAdminLicense";
import Spinner from "@/components/common/Spinner";

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
          <Spinner />
          <span className="text-xs font-mono text-text-muted">
            Checking license...
          </span>
        </div>
      </div>
    );
  }

  const installedLicense = license && "grace_period" in license ? license : null;

  // Expired, missing, or error -> redirect to license page
  if (isError || !installedLicense || installedLicense.expired) {
    return <Navigate to="/admin/license" replace />;
  }

  return <Outlet />;
}
