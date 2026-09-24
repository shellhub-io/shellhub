import { useQuery } from "@tanstack/react-query";
import {
  provisioningKeyList,
  provisioningKeyListQueryKey,
  type ProvisioningKeyListData,
  type ProvisioningKey,
} from "../client";
import { paginatedQueryFn, type PaginatedResult } from "../api/pagination";

interface UseProvisioningKeysParams {
  page?: number;
  perPage?: number;
  sortBy?: string;
  orderBy?: "asc" | "desc";
}

/**
 * A page of the namespace's provisioning keys, newest first.
 */
export function useProvisioningKeys({
  page = 1,
  perPage = 10,
  sortBy = "created_at",
  orderBy = "desc",
}: UseProvisioningKeysParams = {}) {
  const options = {
    query: { page, per_page: perPage, sort_by: sortBy, order_by: orderBy },
  } satisfies { query: ProvisioningKeyListData["query"] };

  const result = useQuery<PaginatedResult<ProvisioningKey>>({
    queryKey: provisioningKeyListQueryKey(options),
    queryFn: paginatedQueryFn(provisioningKeyList, options),
  });

  return {
    provisioningKeys: result.data?.data ?? [],
    totalCount: result.data?.totalCount ?? 0,
    isLoading: result.isLoading,
    error: result.error,
  };
}
