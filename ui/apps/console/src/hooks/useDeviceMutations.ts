import { useMutation } from "@tanstack/react-query";
import { isSdkError } from "../api/errors";
import {
  acceptDeviceMutation,
  updateDeviceStatusMutation,
  deleteDeviceMutation,
  updateDeviceMutation,
  pullTagFromDeviceMutation,
  setDeviceCustomFieldMutation,
  deleteDeviceCustomFieldMutation,
  createTag,
  pushTagToDevice,
} from "../client";
import { useInvalidateByIds } from "./useInvalidateQueries";

/**
 * Accepts a pending device. The counts change with it, so the stats query is refreshed as well
 * as the lists.
 */
export function useAcceptDevice() {
  const invalidate = useInvalidateByIds(
    "getDevices",
    "getDevice",
    "getStatusDevices",
    "getStats",
  );
  return useMutation({
    ...acceptDeviceMutation(),
    onSuccess: invalidate,
  });
}

/**
 * Rejects a pending device. The device leaves the pending list but the accepted count is
 * unchanged, so stats are not refreshed.
 */
export function useRejectDevice() {
  const invalidate = useInvalidateByIds(
    "getDevices",
    "getDevice",
    "getStatusDevices",
  );
  return useMutation({
    ...updateDeviceStatusMutation(),
    onSuccess: invalidate,
  });
}

/**
 * Removes a device from the namespace, refreshing the lists and the counts.
 */
export function useRemoveDevice() {
  const invalidate = useInvalidateByIds(
    "getDevices",
    "getDevice",
    "getStatusDevices",
    "getStats",
  );
  return useMutation({
    ...deleteDeviceMutation(),
    onSuccess: invalidate,
  });
}

/**
 * Renames a device.
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
 * Sets a custom field on a device.
 */
export function useSetDeviceCustomField() {
  const invalidate = useInvalidateByIds("getDevices", "getDevice");
  return useMutation({
    ...setDeviceCustomFieldMutation(),
    onSuccess: invalidate,
  });
}

/**
 * Deletes a custom field from a device.
 */
export function useDeleteDeviceCustomField() {
  const invalidate = useInvalidateByIds("getDevices", "getDevice");
  return useMutation({
    ...deleteDeviceCustomFieldMutation(),
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
