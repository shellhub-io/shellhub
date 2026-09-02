import { useMutation } from "@tanstack/react-query";
import { acceptDevicePairingMutation } from "@/client";
import { useInvalidateByIds } from "./useInvalidateQueries";

/**
 * Completes an enrolment from a code, adding the device to the namespace and refreshing the
 * device queries and counts.
 */
export function useAcceptDevicePairing() {
  const invalidate = useInvalidateByIds(
    "getDevices",
    "getDevice",
    "getStatusDevices",
    "getStats",
  );
  return useMutation({
    ...acceptDevicePairingMutation(),
    onSuccess: invalidate,
  });
}
