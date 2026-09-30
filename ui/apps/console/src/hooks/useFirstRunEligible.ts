import { useHasPermission } from "@/hooks/useHasPermission";
import { useNamespaces } from "@/hooks/useNamespaces";
import { useStats } from "@/hooks/useStats";

/**
 * Whether the current user is on their account's first run: a member who can accept devices,
 * whose only namespace has no accepted device. "loading" while the device counts that decide it
 * are still on their way; false when they could not be loaded. The dashboard gate and the
 * device list's way back both ask this.
 */
export function useFirstRunEligible(): boolean | "loading" {
  const canAccept = useHasPermission("device:accept");
  const { namespaces } = useNamespaces();
  const { stats, isLoading } = useStats();

  if (!canAccept || namespaces.length !== 1) return false;
  if (isLoading) return "loading";
  return !!stats && (stats.registered_devices ?? 0) === 0;
}
