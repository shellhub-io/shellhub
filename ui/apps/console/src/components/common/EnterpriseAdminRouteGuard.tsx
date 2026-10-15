import { Navigate, Outlet } from "react-router-dom";
import { isEnterprise } from "@/env";
import { ADMIN_INSTANCE_PATH } from "@/utils/adminRoute";

/**
 * Blocks an admin route only enterprise has. Any other edition is redirected to the instance
 * overview, replacing the history entry so Back does not return to it.
 */
export default function EnterpriseAdminRouteGuard() {
  if (!isEnterprise()) {
    return <Navigate to={ADMIN_INSTANCE_PATH} replace />;
  }

  return <Outlet />;
}
