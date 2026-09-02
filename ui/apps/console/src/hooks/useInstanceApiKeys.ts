import { useQuery } from "@tanstack/react-query";
import {
  listInstanceApiKeys,
  listInstanceApiKeysQueryKey,
  type ListInstanceApiKeysData,
  type InstanceApiKey,
} from "@/client";
import { paginatedQueryFn, type PaginatedResult } from "@/api/pagination";

interface UseInstanceApiKeysParams {
  page?: number;
  perPage?: number;
  orderBy?: "asc" | "desc";
}

/**
 * A page of the instance's admin API keys, newest first.
 */
export function useInstanceApiKeys({
  page = 1,
  perPage = 10,
  orderBy = "desc",
}: UseInstanceApiKeysParams = {}) {
  const options = {
    query: { page, per_page: perPage, order_by: orderBy },
  } satisfies { query: ListInstanceApiKeysData["query"] };

  const result = useQuery<PaginatedResult<InstanceApiKey>>({
    queryKey: listInstanceApiKeysQueryKey(options),
    queryFn: paginatedQueryFn(listInstanceApiKeys, options),
  });

  return {
    apiKeys: result.data?.data ?? [],
    totalCount: result.data?.totalCount ?? 0,
    isLoading: result.isLoading,
    error: result.error,
  };
}
