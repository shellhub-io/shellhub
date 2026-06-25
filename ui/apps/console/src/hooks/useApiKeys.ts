import { useQuery } from "@tanstack/react-query";
import {
  apiKeyList,
  type ApiKeyListData,
  type ApiKey,
} from "../client";
import { apiKeyListQueryKey } from "../client/@tanstack/react-query.gen";
import { paginatedQueryFn, type PaginatedResult } from "../api/pagination";

interface UseApiKeysParams {
  page?: number;
  perPage?: number;
  sortBy?: string;
  orderBy?: "asc" | "desc";
}

export function useApiKeys({
  page = 1,
  perPage = 10,
  sortBy = "created_at",
  orderBy = "desc",
}: UseApiKeysParams = {}) {
  const options = { query: { page, per_page: perPage, sort_by: sortBy, order_by: orderBy } } satisfies { query: ApiKeyListData["query"] };

  const result = useQuery<PaginatedResult<ApiKey>>({
    queryKey: apiKeyListQueryKey(options),
    queryFn: paginatedQueryFn(apiKeyList, options),
  });

  return {
    apiKeys: result.data?.data ?? [],
    totalCount: result.data?.totalCount ?? 0,
    isLoading: result.isLoading,
    error: result.error,
  };
}
