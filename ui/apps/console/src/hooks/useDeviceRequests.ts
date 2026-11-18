import { useMutation, useQuery } from "@tanstack/react-query";
import {
  acceptDeviceMutation,
  getDeviceLocations,
  getDeviceLocationsQueryKey,
  getDeviceOptions,
  pullTagFromDeviceMutation,
  pushTagToDevice,
  updateDeviceMutation,
  type Device,
} from "@/client";
import { ensureTag } from "@/hooks/useTagRequests";
import {
  useInvalidateByIds,
  DEVICE_ROW_QUERIES,
  DEVICE_STATUS_QUERIES,
} from "@/hooks/useInvalidateQueries";
import type { GeoIPSource } from "@/types/geoip";

/**
 * One device by UID. Idle until a UID is given. refetchInterval keeps it polling, and as a function
 * it gets the latest device and returns false to stop.
 */
export function useDevice(
  uid: string,
  {
    refetchInterval = false,
  }: {
    refetchInterval?: number | false | ((device?: Device) => number | false);
  } = {},
) {
  const result = useQuery({
    ...getDeviceOptions({ path: { uid } }),
    enabled: !!uid,
    refetchInterval:
      typeof refetchInterval === "function"
        ? (query) => refetchInterval(query.state.data)
        : refetchInterval,
  });

  return {
    device: result.data ?? null,
    isLoading: result.isLoading,
    error: result.error,
  };
}

const DEVICE_LOCATIONS_REFRESH_MS = 30_000;

function geoipSourceOf(response: Response): GeoIPSource | undefined {
  const source = response.headers.get("X-GeoIP-Source");

  return source === "dbip" || source === "maxmind" ? source : undefined;
}

/**
 * The namespace's accepted devices counted by position, refreshed every thirty seconds so the
 * online counts follow the fleet, and the GeoIP database the server placed them with, which the
 * map credits. locations is empty and source undefined until the first read succeeds. A failed
 * read keeps the last successful locations and source and sets error, so a caller that must not
 * act on stale positions checks error, not only locations. source is also undefined when the
 * server names no database. read is true while a successful read is cached: false while the first
 * read waits for the network or after it fails, when isLoading is false, and again after entering
 * another namespace drops the cache, until that namespace's first read succeeds.
 */
export function useDeviceLocations() {
  const result = useQuery({
    queryKey: getDeviceLocationsQueryKey(),
    queryFn: async ({ signal }) => {
      const { data, response } = await getDeviceLocations({
        signal,
        throwOnError: true,
      });

      return { locations: data, source: geoipSourceOf(response) };
    },
    refetchInterval: DEVICE_LOCATIONS_REFRESH_MS,
  });

  return {
    locations: result.data?.locations ?? [],
    source: result.data?.source,
    isLoading: result.isLoading,
    read: result.data !== undefined,
    error: result.error,
  };
}

/**
 * Accepts a pending device and refreshes every query a status change moves
 * (DEVICE_STATUS_QUERIES).
 */
export function useAcceptDevice() {
  const invalidate = useInvalidateByIds(...DEVICE_STATUS_QUERIES);
  return useMutation({
    ...acceptDeviceMutation(),
    onSuccess: invalidate,
  });
}

/**
 * Renames a device and refreshes every query that shows it (DEVICE_ROW_QUERIES).
 */
export function useRenameDevice() {
  const invalidate = useInvalidateByIds(...DEVICE_ROW_QUERIES);
  return useMutation({
    ...updateDeviceMutation(),
    onSuccess: invalidate,
  });
}

/**
 * Tags a device, creating the tag first when it does not exist yet.
 */
export function useAddDeviceTag() {
  const invalidate = useInvalidateByIds(...DEVICE_ROW_QUERIES, "getTags");
  return useMutation({
    mutationFn: async (options: { path: { uid: string; name: string } }) => {
      await ensureTag(options.path.name);
      return pushTagToDevice({ ...options, throwOnError: true });
    },
    onSuccess: invalidate,
  });
}

/**
 * Removes a tag from a device. The tag survives on anything else carrying it.
 */
export function useRemoveDeviceTag() {
  const invalidate = useInvalidateByIds(...DEVICE_ROW_QUERIES);
  return useMutation({
    ...pullTagFromDeviceMutation(),
    onSuccess: invalidate,
  });
}
