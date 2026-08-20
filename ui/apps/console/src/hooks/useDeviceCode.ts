import { useQuery, useMutation } from "@tanstack/react-query";
import {
  resolveDeviceLoginCodeOptions,
  acceptDevicePairingMutation,
} from "@/client";
import { useInvalidateByIds } from "./useInvalidateQueries";

export function useResolveDeviceCode(code: string) {
  const { data, isLoading, isError, error } = useQuery({
    ...resolveDeviceLoginCodeOptions({ path: { code } }),
    enabled: !!code,
    retry: false,
    staleTime: Infinity,
  });

  return { device: data ?? null, isLoading, isError, error };
}

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
