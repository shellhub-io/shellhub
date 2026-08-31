import { Outlet, Navigate } from "react-router-dom";
import { useAdminLicense } from "@/hooks/useAdminLicense";
import PageLoader from "@/components/common/PageLoader";

/**
 * Blocks a route that an expired licence does not cover. Waits for the licence query rather than
 * assuming: refusing on a slow request would lock an admin out of the page that fixes it.
 */
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
