import { useMutation, useQueryClient } from "@tanstack/react-query";
import {
  acceptDeviceMutation,
  updateDeviceStatusMutation,
  deleteDeviceMutation,
  updateDeviceMutation,
  pullTagFromDeviceMutation,
} from "../client/@tanstack/react-query.gen";
import { createTag, pushTagToDevice } from "../client";

function useInvalidateDevices() {
  const queryClient = useQueryClient();
  return () => queryClient.invalidateQueries({ predicate: (query) => {
    const key = query.queryKey[0];
    if (typeof key === "object" && key !== null && "_id" in key) {
      const id = (key as { _id: string })._id;
      return id === "getDevices" || id === "getDevice" || id === "getStats";
    }
    return false;
  } });
}

export function useAcceptDevice() {
  const invalidate = useInvalidateDevices();
  return useMutation({
    ...acceptDeviceMutation(),
    onSuccess: invalidate,
  });
}

export function useRejectDevice() {
  const invalidate = useInvalidateDevices();
  return useMutation({
    ...updateDeviceStatusMutation(),
    onSuccess: invalidate,
  });
}

export function useRemoveDevice() {
  const invalidate = useInvalidateDevices();
  return useMutation({
    ...deleteDeviceMutation(),
    onSuccess: invalidate,
  });
}

export function useRenameDevice() {
  const invalidate = useInvalidateDevices();
  return useMutation({
    ...updateDeviceMutation(),
    onSuccess: invalidate,
  });
}

export function useAddDeviceTag() {
  const queryClient = useQueryClient();
  const invalidateDevices = useInvalidateDevices();
  return useMutation({
    mutationFn: async (options: { path: { uid: string; name: string } }) => {
      try {
        await createTag({ body: { name: options.path.name }, throwOnError: true });
      } catch (e) {
        const status = (e as { status?: number }).status;
        if (status !== 409) throw e;
      }
      return pushTagToDevice({ ...options, throwOnError: true });
    },
    onSuccess: async () => {
      await invalidateDevices();
      await queryClient.invalidateQueries({ predicate: (query) => {
        const key = query.queryKey[0];
        if (typeof key === "object" && key !== null && "_id" in key) {
          return (key as { _id: string })._id === "getTags";
        }
        return false;
      } });
    },
  });
}

export function useRemoveDeviceTag() {
  const invalidate = useInvalidateDevices();
  return useMutation({
    ...pullTagFromDeviceMutation(),
    onSuccess: invalidate,
  });
}
