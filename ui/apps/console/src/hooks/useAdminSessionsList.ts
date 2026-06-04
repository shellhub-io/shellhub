import { useQuery } from "@tanstack/react-query";
import { getSessionsAdmin, type GetSessionsAdminData, type Session } from "../client";
import { getSessionsAdminQueryKey } from "../client/@tanstack/react-query.gen";
import { paginatedQueryFn, type PaginatedResult } from "../api/pagination";
import { useAuthStore } from "../stores/authStore";
import { isSdkError } from "../api/errors";

function toDisplayError(err: unknown): Error {
  if (isSdkError(err)) {
    if (err.status === 403) return new Error("You don't have permission to view sessions.");
    if (err.status >= 500) return new Error("Server error. Please try again later.");
    return new Error(`Failed to load sessions (${err.status}).`);
  }
  if (err instanceof Error) return err;
  return new Error("Failed to load sessions.");
}

export function useAdminSessionsList(page: number, perPage: number) {
  const isAdmin = useAuthStore((s) => s.isAdmin);
  const options = { query: { page, per_page: perPage } } satisfies { query: GetSessionsAdminData["query"] };

  const result = useQuery<PaginatedResult<Session>>({
    queryKey: getSessionsAdminQueryKey(options),
    queryFn: paginatedQueryFn(getSessionsAdmin, options),
    enabled: isAdmin,
    staleTime: 60 * 1000,
    retry: (count, err) => isSdkError(err) && err.status === 401 ? false : count < 1,
    refetchOnWindowFocus: false,
  });

  return {
    sessions: result.data?.data ?? [],
    totalCount: result.data?.totalCount ?? 0,
    isLoading: result.isLoading,
    error: result.error ? toDisplayError(result.error) : null,
  };
}
