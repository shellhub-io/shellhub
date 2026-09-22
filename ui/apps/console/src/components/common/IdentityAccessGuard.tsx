import { Navigate, Outlet } from "react-router-dom";
import { useAuthStore } from "@/stores/authStore";
import { useNamespace } from "@/hooks/useNamespaces";

/**
 * Redirects Access Policies to Public Keys when the namespace is in legacy access mode, where the
 * gateway reads no policy and the page would offer rules that decide nothing. It matches the
 * sidebar, which hides the link, and mirrors LegacyAccessGuard.
 *
 * It tests for legacy rather than against identity, so a namespace that has not resolved yet holds
 * the page instead of bouncing the reader out of one they are allowed to see.
 *
 * SSH Identities is deliberately not guarded: the approval routes live under it, and an approval
 * belongs to the namespace the SSH login targeted rather than the active one, so deciding by the
 * active namespace would strand a login whose namespace is not the one in the switcher.
 */
export default function IdentityAccessGuard() {
  const { tenant } = useAuthStore();
  const { namespace } = useNamespace(tenant ?? "");

  if (namespace?.settings?.ssh_access_mode === "legacy") {
    return <Navigate to="/sshkeys/public-keys" replace />;
  }

  return <Outlet />;
}
