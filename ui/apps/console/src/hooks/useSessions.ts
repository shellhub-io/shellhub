import { useQuery } from "@tanstack/react-query";
import {
  getSessions as getSessionsSdk,
  getSessionsQueryKey,
  type GetSessionsData,
  type Session,
} from "../client";
import { paginatedQueryFn, type PaginatedResult } from "../api/pagination";

interface UseSessionsParams {
  page?: number;
  perPage?: number;
}

/**
 * A page of the namespace's sessions, newest first.
 */
export function useSessions({ page = 1, perPage = 10 }: UseSessionsParams = {}) {
  const options = { query: { page, per_page: perPage } } satisfies { query: GetSessionsData["query"] };

  const result = useQuery<PaginatedResult<Session>>({
    queryKey: getSessionsQueryKey(options),
    queryFn: paginatedQueryFn(getSessionsSdk, options),
  });

  return {
    sessions: result.data?.data ?? [],
    totalCount: result.data?.totalCount ?? 0,
    isLoading: result.isLoading,
    error: result.error,
    refetch: result.refetch,
  };
}
