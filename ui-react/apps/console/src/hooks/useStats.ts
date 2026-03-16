import { useQuery } from "@tanstack/react-query";
import { getStatusDevicesOptions } from "../client/@tanstack/react-query.gen";

export function useStats() {
  const result = useQuery(getStatusDevicesOptions());

  return {
    stats: result.data ?? null,
    isLoading: result.isLoading,
    error: result.error,
    refetch: result.refetch,
  };
}
