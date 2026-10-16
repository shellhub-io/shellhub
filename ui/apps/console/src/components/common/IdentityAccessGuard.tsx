import { Navigate, Outlet } from "react-router-dom";
import { useNamespaceLegacy } from "@/hooks/useNamespaceLegacy";

/**
 * Redirects the pages only a new namespace has, Access Policies and the Fleet tab, to Public Keys
 * while the namespace is legacy, where the gateway reads no policy and no provisioning key enrolls
 * anything. It matches the sidebar, which hides the links, and mirrors LegacyAccessGuard: it acts
 * only once the namespace has resolved as legacy.
 *
 * SSH Identities is deliberately not guarded: the approval routes live under it, and an approval
 * belongs to the namespace the SSH login targeted rather than the active one, so deciding by the
 * active namespace would strand a login whose namespace is not the one in the switcher.
 */
export default function IdentityAccessGuard() {
  const legacy = useNamespaceLegacy();

  if (legacy === true) {
    return <Navigate to="/sshkeys/public-keys" replace />;
  }

  return <Outlet />;
}
