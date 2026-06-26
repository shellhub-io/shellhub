import { Outlet, Navigate } from "react-router-dom";
import { useAdminLicense } from "@/hooks/useAdminLicense";
import PageLoader from "@/components/common/PageLoader";

export default function LicenseGuard() {
  const { isLoading, isError, isExpired } = useAdminLicense();

  if (isLoading) {
    return <PageLoader label="Checking license..." showLabel padding="fill" />;
  }

  if (isError || isExpired) {
    return <Navigate to="/admin/license" replace />;
  }

  return <Outlet />;
}
