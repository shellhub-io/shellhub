import { useQuery } from "@tanstack/react-query";
import { installKeyList, type InstallKeyListData, type InstallKey } from "../client";
import { installKeyListQueryKey } from "../client/@tanstack/react-query.gen";
import { paginatedQueryFn, type PaginatedResult } from "../api/pagination";

interface UseInstallKeysParams {
  page?: number;
  perPage?: number;
  sortBy?: string;
  orderBy?: "asc" | "desc";
}

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
