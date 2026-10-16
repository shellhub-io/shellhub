import { useHasPermission } from "@/hooks/useHasPermission";
import { useNamespaces } from "@/hooks/useNamespaces";
import { useNamespaceLegacy } from "@/hooks/useNamespaceLegacy";
import { useStats } from "@/hooks/useStats";

/**
 * Whether the current user is on their account's first run: a member who can accept devices,
 * whose only namespace has no accepted device. A new namespace with a pending device is past it:
 * the device came by a provisioning key or the deprecated tenant and waits to be accepted, which
 * the trail never offers. A legacy namespace's own first run goes through a pending device, so
 * there only accepted devices count. "loading" while the device counts that decide it are still
 * on their way; false when they could not be loaded. The dashboard gate and the device list's way
 * back both ask this.
 */
export function useFirstRunEligible(): boolean | "loading" {
  const canAccept = useHasPermission("device:accept");
  const { namespaces } = useNamespaces();
  const { stats, isLoading } = useStats();
  const legacy = useNamespaceLegacy() === true;

  if (!canAccept || namespaces.length !== 1) return false;
  if (isLoading) return "loading";
  if (!stats || (stats.registered_devices ?? 0) > 0) return false;
  return legacy || (stats.pending_devices ?? 0) === 0;
}
