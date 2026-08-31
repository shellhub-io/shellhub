import { useQuery, useMutation } from "@tanstack/react-query";
import {
  resolveDeviceLoginCodeOptions,
  acceptDevicePairingMutation,
} from "@/client";
import { useInvalidateByIds } from "./useInvalidateQueries";

/**
 * Resolves an enrolment code to the device waiting behind it. Never retried and never stale: a
 * code is single-use, so a second attempt would fail and a cached answer would be wrong.
 */
export function useResolveDeviceCode(code: string) {
  const { data, isLoading, isError, error } = useQuery({
    ...resolveDeviceLoginCodeOptions({ path: { code } }),
    enabled: !!code,
    retry: false,
    staleTime: Infinity,
  });

  return { device: data ?? null, isLoading, isError, error };
}

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
