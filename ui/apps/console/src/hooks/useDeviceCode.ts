import { useQuery, useMutation } from "@tanstack/react-query";
import {
  resolveDeviceLoginCodeOptions,
  acceptDevicePairingMutation,
} from "@/client";
import {
  useInvalidateByIds,
  DEVICE_STATUS_QUERIES,
} from "./useInvalidateQueries";

/**
 * Resolves an enrolment code to the device waiting behind it. Never retried and never stale: a
 * code is single-use, so a second attempt would fail and a cached answer would be wrong.
 */
export function useResolveDeviceCode(code: string) {
  const { data, isLoading, isFetching, isError, error, refetch } = useQuery({
    ...resolveDeviceLoginCodeOptions({ path: { code } }),
    enabled: !!code,
    retry: false,
    staleTime: Infinity,
  });

  return {
    device: data ?? null,
    isLoading,
    isFetching,
    isError,
    error,
    refetch,
  };
}

/**
 * Completes an enrolment from a code, adding the device to the namespace.
 */
export function useAcceptDevicePairing() {
  const invalidate = useInvalidateByIds(...DEVICE_STATUS_QUERIES);
  return useMutation({
    ...acceptDevicePairingMutation(),
    onSuccess: invalidate,
  });
}
