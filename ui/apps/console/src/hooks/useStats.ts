import { useQuery } from "@tanstack/react-query";
import { getStatusDevicesOptions } from "../client";

/**
 * The namespace's device counts, by status. Returns null while loading rather than zeroes, so a
 * dashboard does not flash "0 devices" at someone who has some.
 */
export function useStats() {
  const result = useQuery(getStatusDevicesOptions());

  return {
    stats: result.data ?? null,
    isLoading: result.isLoading,
    error: result.error,
    refetch: result.refetch,
  };
}
