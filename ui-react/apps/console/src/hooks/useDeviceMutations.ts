import { useMutation } from "@tanstack/react-query";
import { isSdkError } from "../api/errors";
import {
  acceptDeviceMutation,
  updateDeviceStatusMutation,
  deleteDeviceMutation,
  updateDeviceMutation,
  pullTagFromDeviceMutation,
} from "../client/@tanstack/react-query.gen";
import { createTag, pushTagToDevice } from "../client";
import { useInvalidateByIds } from "./useInvalidateQueries";

export function useAcceptDevice() {
  const invalidate = useInvalidateByIds("getDevices", "getDevice", "getStatusDevices");
  return useMutation({
    ...acceptDeviceMutation(),
    onSuccess: invalidate,
  });
}

export function useRejectDevice() {
  const invalidate = useInvalidateByIds("getDevices", "getDevice", "getStatusDevices");
  return useMutation({
    ...updateDeviceStatusMutation(),
    onSuccess: invalidate,
  });
}

export function useRemoveDevice() {
  const invalidate = useInvalidateByIds("getDevices", "getDevice", "getStatusDevices");
  return useMutation({
    ...deleteDeviceMutation(),
    onSuccess: invalidate,
  });
}

export function useRenameDevice() {
  const invalidate = useInvalidateByIds("getDevices", "getDevice", "getStatusDevices");
  return useMutation({
    ...updateDeviceMutation(),
    onSuccess: invalidate,
  });
}

export function useUpdateDeviceSSH() {
  const invalidate = useInvalidateByIds("getDevices", "getDevice", "getStatusDevices");
  return useMutation({
    ...updateDeviceMutation(),
    onSuccess: invalidate,
  });
}

export function useAddDeviceTag() {
  const invalidate = useInvalidateByIds("getDevices", "getDevice", "getStatusDevices", "getTags");
  return useMutation({
    mutationFn: async (options: { path: { uid: string; name: string } }) => {
      try {
        await createTag({ body: { name: options.path.name }, throwOnError: true });
      } catch (e) {
        if (!isSdkError(e) || e.status !== 409) throw e;
      }
      return pushTagToDevice({ ...options, throwOnError: true });
    },
    onSuccess: invalidate,
  });
}

export function useRemoveDeviceTag() {
  const invalidate = useInvalidateByIds("getDevices", "getDevice", "getStatusDevices");
  return useMutation({
    ...pullTagFromDeviceMutation(),
    onSuccess: invalidate,
  });
}
