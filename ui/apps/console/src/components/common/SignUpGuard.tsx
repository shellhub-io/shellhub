import { Outlet, Navigate } from "react-router-dom";
import { getConfig } from "@/env";

// /sign-up is Cloud's open self-registration. Everywhere else accounts are
// created only by invitation (completed on /accept-invite), so sign-up stays
// cloud-only.
export default function SignUpGuard() {
  if (!getConfig().cloud) return <Navigate to="/login" replace />;
  return <Outlet />;
}
