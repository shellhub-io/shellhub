import { useMutation } from "@tanstack/react-query";
import { clsoeSessionMutation, deleteSessionRecord } from "../client";
import { useInvalidateByIds } from "./useInvalidateQueries";

export function useCloseSession() {
  const invalidate = useInvalidateByIds("getSessions", "getSession", "getStatusDevices");
  return useMutation({
    ...clsoeSessionMutation(),
    onSuccess: invalidate,
  });
}

export function useDeleteSessionRecording() {
  const invalidate = useInvalidateByIds("getSessions", "getSession");
  return useMutation({
    mutationFn: async (uid: string) => {
      await deleteSessionRecord({ path: { uid, seat: 0 }, throwOnError: true });
    },
    onSuccess: invalidate,
  });
}
