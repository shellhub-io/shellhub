import { Navigate, Outlet } from "react-router-dom";
import { useAuthStore } from "@/stores/authStore";
import { useNamespace } from "@/hooks/useNamespaces";

// LegacyAccessGuard redirects the legacy key ACL, firewall, and key-vault pages
// to Access Policies when the namespace uses identity access mode, where those
// features are bypassed. It matches the sidebar, which hides the same links.
export default function LegacyAccessGuard() {
  const { tenant } = useAuthStore();
  const { namespace } = useNamespace(tenant ?? "");

  if (namespace?.settings?.ssh_access_mode === "identity") {
    return <Navigate to="/access-policies" replace />;
  }

  return <Outlet />;
}
