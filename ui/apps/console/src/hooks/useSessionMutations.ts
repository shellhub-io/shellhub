import { useMutation } from "@tanstack/react-query";
import { clsoeSessionMutation, deleteSessionRecord } from "../client";
import { useInvalidateByIds } from "./useInvalidateQueries";

/**
 * Closes an open session. The device's status changes with it, so the device counts are
 * refreshed alongside the session queries.
 */
export function useCloseSession() {
  const invalidate = useInvalidateByIds("getSessions", "getSession", "getStatusDevices");
  return useMutation({
    ...clsoeSessionMutation(),
    onSuccess: invalidate,
  });
}

/**
 * Deletes a session's recording, leaving the session itself. Seat zero is the only one the UI
 * records, so that is the one removed.
 */
export function useDeleteSessionRecording() {
  const invalidate = useInvalidateByIds("getSessions", "getSession");
  return useMutation({
    mutationFn: async (uid: string) => {
      await deleteSessionRecord({ path: { uid, seat: 0 }, throwOnError: true });
    },
    onSuccess: invalidate,
  });
}
