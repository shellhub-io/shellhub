import { Outlet, Navigate } from "react-router-dom";
import { getConfig } from "../../env";

// Sign-up endpoints (/api/register, /api/user/resend_email, /api/user/validation_account)
// are registered exclusively in cloud. This guard restricts the sign-up route to cloud only.
export default function SignUpGuard() {
  if (!getConfig().cloud) return <Navigate to="/login" replace />;
  return <Outlet />;
}
