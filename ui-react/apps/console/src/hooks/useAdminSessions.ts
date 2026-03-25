import { useQuery } from "@tanstack/react-query";
import {
  getSessionsAdmin,
  type GetSessionsAdminData,
  type Session,
} from "../client";
import { getSessionsAdminQueryKey } from "../client/@tanstack/react-query.gen";
import { paginatedQueryFn, type PaginatedResult } from "../api/pagination";
import { useAuthStore } from "../stores/authStore";
import { isSdkError } from "../api/errors";

const SESSIONS_OPTIONS = { query: { page: 1, per_page: 5 } } satisfies { query: GetSessionsAdminData["query"] };

export function useAdminSessions() {
  const isAdmin = useAuthStore((s) => s.isAdmin);

  const result = useQuery<PaginatedResult<Session>>({
    queryKey: getSessionsAdminQueryKey(SESSIONS_OPTIONS),
    queryFn: paginatedQueryFn(getSessionsAdmin, SESSIONS_OPTIONS),
    enabled: isAdmin,
    staleTime: 5 * 60 * 1000, // 5 minutes
    retry: (count, err) => isSdkError(err) && err.status === 401 ? false : count < 1,
    refetchOnWindowFocus: false,
  });

  return {
    sessions: result.data?.data ?? [],
    totalCount: result.data?.totalCount ?? 0,
    isLoading: result.isLoading,
    error: result.error,
  };
}
