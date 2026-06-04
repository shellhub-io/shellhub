import { useQuery } from "@tanstack/react-query";
import {
  getStatsOptions,
  getStatsQueryKey,
} from "../client/@tanstack/react-query.gen";
import { useAuthStore } from "../stores/authStore";
import { isSdkError } from "../api/errors";

export { getStatsQueryKey };

export function useAdminStats() {
  const isAdmin = useAuthStore((s) => s.isAdmin);

  const result = useQuery({
    ...getStatsOptions(),
    enabled: isAdmin,
    staleTime: 5 * 60 * 1000, // 5 minutes
    retry: (count, err) => isSdkError(err) && err.status === 401 ? false : count < 1,
    refetchOnWindowFocus: false,
  });

  return {
    stats: result.data,
    isLoading: result.isLoading,
    isError: result.isError,
    error: result.error,
  };
}
