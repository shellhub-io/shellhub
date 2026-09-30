import { useQuery } from "@tanstack/react-query";
import { getDeviceOptions, type Device } from "../client";

/**
 * One device by UID. Idle until a UID is given. refetchInterval keeps it polling, and as a function
 * it gets the latest device and returns false to stop.
 */
export function useDevice(
  uid: string,
  {
    refetchInterval = false,
  }: {
    refetchInterval?: number | false | ((device?: Device) => number | false);
  } = {},
) {
  const result = useQuery({
    ...getDeviceOptions({ path: { uid } }),
    enabled: !!uid,
    refetchInterval:
      typeof refetchInterval === "function"
        ? (query) => refetchInterval(query.state.data)
        : refetchInterval,
  });

  return {
    device: result.data ?? null,
    isLoading: result.isLoading,
    error: result.error,
    refetch: result.refetch,
  };
}
