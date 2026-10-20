import { useMutation, useQuery } from "@tanstack/react-query";
import { isSdkError } from "@/api/errors";
import {
  acceptDeviceMutation,
  createTag,
  getDeviceOptions,
  pullTagFromDeviceMutation,
  pushTagToDevice,
  updateDeviceMutation,
  type Device,
} from "@/client";
import { useInvalidateByIds } from "@/hooks/useInvalidateQueries";

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
