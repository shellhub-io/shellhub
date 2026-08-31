import { useQuery } from "@tanstack/react-query";
import { getDeviceOptions } from "../client";

/**
 * One device by UID. Idle until a UID is given.
 */
export function useDevice(uid: string) {
  const result = useQuery({
    ...getDeviceOptions({ path: { uid } }),
    enabled: !!uid,
  });

  return {
    device: result.data ?? null,
    isLoading: result.isLoading,
    error: result.error,
    refetch: result.refetch,
  };
}
