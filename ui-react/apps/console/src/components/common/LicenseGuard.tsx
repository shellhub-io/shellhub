import { Outlet, Navigate } from "react-router-dom";
import { useAdminLicense } from "@/hooks/useAdminLicense";
import PageLoader from "@/components/common/PageLoader";

export default function LicenseGuard() {
  const { data: license, isLoading, isError } = useAdminLicense();

  if (isLoading) {
    return <PageLoader label="Checking license..." showLabel padding="fill" />;
  }

  const installedLicense =
    license && "grace_period" in license ? license : null;

  // Expired, missing, or error -> redirect to license page
  if (isError || !installedLicense || installedLicense.expired) {
    return <Navigate to="/admin/license" replace />;
  }

  return <Outlet />;
}
