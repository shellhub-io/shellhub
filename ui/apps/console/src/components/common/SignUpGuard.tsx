import { Outlet, Navigate } from "react-router-dom";
import { isCloud } from "@/env";

/**
 * Gates /sign-up, which is cloud's open self-registration. Everywhere else an account is created
 * only by invitation and completed on /accept-invite, so the route redirects rather than
 * offering a form that would be refused.
 */
export default function SignUpGuard() {
  if (!isCloud()) return <Navigate to="/login" replace />;
  return <Outlet />;
}
