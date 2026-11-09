import { useMutation, useQuery } from "@tanstack/react-query";
import { isSdkError } from "@/api/errors";
import {
  acceptDeviceMutation,
  createTag,
  getDeviceLocations,
  getDeviceLocationsQueryKey,
  getDeviceOptions,
  pullTagFromDeviceMutation,
  pushTagToDevice,
  updateDeviceMutation,
  type Device,
} from "@/client";
import { useInvalidateByIds } from "@/hooks/useInvalidateQueries";
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
 * Accepts a pending device. The counts change with it, so the stats query is refreshed as well
 * as the lists — including the provisioning keys list, which carries each key's pending count.
 */
export function useAcceptDevice() {
  const invalidate = useInvalidateByIds(
    "getDevices",
    "getDevice",
    "getStatusDevices",
    "getStats",
    "provisioningKeyList",
    "listTenantRegistrations",
  );
  return useMutation({
    ...acceptDeviceMutation(),
    onSuccess: invalidate,
  });
}

/**
 * Renames a device, refreshing the device lists, the device itself and the status counts.
 */
export function useRenameDevice() {
  const invalidate = useInvalidateByIds(
    "getDevices",
    "getDevice",
    "getStatusDevices",
  );
  return useMutation({
    ...updateDeviceMutation(),
    onSuccess: invalidate,
  });
}

/**
 * Tags a device. The tag list is refreshed too, because the tag may be new.
 */
export function useAddDeviceTag() {
  const invalidate = useInvalidateByIds(
    "getDevices",
    "getDevice",
    "getStatusDevices",
    "getTags",
  );
  return useMutation({
    mutationFn: async (options: { path: { uid: string; name: string } }) => {
      try {
        await createTag({
          body: { name: options.path.name },
          throwOnError: true,
        });
      } catch (e) {
        if (!isSdkError(e) || e.status !== 409) throw e;
      }
      return pushTagToDevice({ ...options, throwOnError: true });
    },
    onSuccess: invalidate,
  });
}

/**
 * Removes a tag from a device. The tag survives on anything else carrying it.
 */
export function useRemoveDeviceTag() {
  const invalidate = useInvalidateByIds(
    "getDevices",
    "getDevice",
    "getStatusDevices",
  );
  return useMutation({
    ...pullTagFromDeviceMutation(),
    onSuccess: invalidate,
  });
}
