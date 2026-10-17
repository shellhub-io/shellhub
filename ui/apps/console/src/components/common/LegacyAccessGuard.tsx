import { Navigate, Outlet } from "react-router-dom";
import { useNamespaceLegacy } from "@/hooks/useNamespaceLegacy";

/**
 * Redirects the legacy key-ACL, firewall and key-vault pages to Access Policies once the
 * namespace is new, where those features decide nothing and the pages would describe rules that
 * no longer apply. It matches the sidebar, which hides the same links, and holds the page while
 * the namespace loads rather than bouncing a legacy reader out of it.
 */
export default function LegacyAccessGuard() {
  const legacy = useNamespaceLegacy();

  if (legacy === false) {
    return <Navigate to="/access-policies" replace />;
  }

  return <Outlet />;
}
