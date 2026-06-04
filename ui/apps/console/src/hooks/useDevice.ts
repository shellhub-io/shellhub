import { useQuery } from "@tanstack/react-query";
import { getDeviceOptions } from "../client/@tanstack/react-query.gen";

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
