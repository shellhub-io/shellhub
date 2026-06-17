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
} from "../client/@tanstack/react-query.gen";
import { createTag, pushTagToDevice } from "../client";
import { useInvalidateByIds } from "./useInvalidateQueries";

export function useAcceptDevice() {
  // "getStats" is included so the DeviceLimitBanner and admin dashboard stats card
  // reflect the updated accepted-device count immediately after a successful accept.
  // Invalidation fires on success only — a failed 402 does not change the count.
  // If the banner is not mounted the fresh count loads on next mount (observer-driven).
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

export function useRemoveDevice() {
  // "getStats" is included so the DeviceLimitBanner clears freed slots and the
  // admin dashboard stats card reflects the reduced count after a successful remove.
  // Invalidation fires on success only — a failed request does not change the count.
  // If the banner is not mounted the fresh count loads on next mount (observer-driven).
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

export function useSetDeviceCustomField() {
  const invalidate = useInvalidateByIds("getDevices", "getDevice");
  return useMutation({
    ...setDeviceCustomFieldMutation(),
    onSuccess: invalidate,
  });
}

export function useDeleteDeviceCustomField() {
  const invalidate = useInvalidateByIds("getDevices", "getDevice");
  return useMutation({
    ...deleteDeviceCustomFieldMutation(),
    onSuccess: invalidate,
  });
}

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
