import { useAdminLicense } from "@/hooks/useAdminLicense";
import { useAdminStats } from "@/hooks/useAdminStats";
import { deviceCapacity, type DeviceCapacity } from "@/utils/license";

/**
 * The installed licence and its accepted devices against its limit. capacity is null until both
 * the licence and the stats have answered, and stays null when either failed or no licence is
 * installed, so a caller checks isLoading and failed before reading a null as nothing to report.
 * failed names the request that failed, the licence's first when both did.
 */
export function useDeviceCapacity() {
  const license = useAdminLicense();
  const {
    stats,
    isLoading: statsLoading,
    isError: statsError,
  } = useAdminStats();

  const limit = license.installedLicense?.features.devices;
  const used = stats?.registered_devices;
  const failed: "license" | "stats" | null = license.isError
    ? "license"
    : statsError
      ? "stats"
      : null;
  const capacity: DeviceCapacity | null =
    typeof limit === "number" && typeof used === "number"
      ? deviceCapacity(limit, used)
      : null;

  return {
    license: license.installedLicense,
    capacity,
    isLoading: license.isLoading || statsLoading,
    isError: failed !== null,
    failed,
  };
}
