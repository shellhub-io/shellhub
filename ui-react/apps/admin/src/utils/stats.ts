import { Stats } from "@/types/stats";

/** True if the namespace has at least one device in any status. */
export function hasAnyDevices(stats: Stats | null): boolean {
  if (!stats) return false;
  return (
    stats.registered_devices > 0 ||
    stats.pending_devices > 0 ||
    stats.rejected_devices > 0
  );
}
