import { useMutation } from "@tanstack/react-query";
import { closeSessionMutation } from "@/client";
import { useInvalidateByIds } from "@/hooks/useInvalidateQueries";

/**
 * Closes an open session. The device's status changes with it, so the device counts are
 * refreshed alongside the session queries.
 */
export function useCloseSession() {
  const invalidate = useInvalidateByIds("getSessions", "getSession", "getStatusDevices");
  return useMutation({
    ...closeSessionMutation(),
    onSuccess: invalidate,
  });
}
