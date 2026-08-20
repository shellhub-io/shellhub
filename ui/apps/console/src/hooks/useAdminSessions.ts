import { useQuery } from "@tanstack/react-query";
import {
  getSessionsAdmin,
  getSessionsAdminQueryKey,
  type GetSessionsAdminData,
  type Session,
} from "@/client";
import { paginatedQueryFn, type PaginatedResult } from "@/api/pagination";
import { useAuthStore } from "@/stores/authStore";
import { isSdkError } from "@/api/errors";

export function useAdminSessions({ page = 1, perPage = 5 } = {}) {
  const isAdmin = useAuthStore((s) => s.isAdmin);
  const options = { query: { page, per_page: perPage } } satisfies {
    query: GetSessionsAdminData["query"];
  };

  const result = useQuery<PaginatedResult<Session>>({
    queryKey: getSessionsAdminQueryKey(options),
    queryFn: paginatedQueryFn(getSessionsAdmin, options),
    enabled: isAdmin,
    staleTime: 5 * 60 * 1000,
    retry: (count, err) =>
      isSdkError(err) && err.status === 401 ? false : count < 1,
    refetchOnWindowFocus: false,
  });

  return {
    sessions: result.data?.data ?? [],
    totalCount: result.data?.totalCount ?? 0,
    isLoading: result.isLoading,
    error: result.error,
  };
}
