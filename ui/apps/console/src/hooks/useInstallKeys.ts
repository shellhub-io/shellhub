import { useQuery } from "@tanstack/react-query";
import {
  installKeyList,
  installKeyListQueryKey,
  type InstallKeyListData,
  type InstallKey,
} from "../client";
import { paginatedQueryFn, type PaginatedResult } from "../api/pagination";

interface UseInstallKeysParams {
  page?: number;
  perPage?: number;
  sortBy?: string;
  orderBy?: "asc" | "desc";
}

/**
 * A page of the namespace's install keys, newest first.
 */
export function useInstallKeys({
  page = 1,
  perPage = 10,
  sortBy = "created_at",
  orderBy = "desc",
}: UseInstallKeysParams = {}) {
  const options = {
    query: { page, per_page: perPage, sort_by: sortBy, order_by: orderBy },
  } satisfies { query: InstallKeyListData["query"] };

  const result = useQuery<PaginatedResult<InstallKey>>({
    queryKey: installKeyListQueryKey(options),
    queryFn: paginatedQueryFn(installKeyList, options),
  });

  return {
    installKeys: result.data?.data ?? [],
    totalCount: result.data?.totalCount ?? 0,
    isLoading: result.isLoading,
    error: result.error,
  };
}
