import { useQuery } from "@tanstack/react-query";
import {
  provisioningKeyList,
  provisioningKeyListQueryKey,
  type ProvisioningKeyListData,
  type ProvisioningKey,
} from "../client";
import { paginatedQueryFn, type PaginatedResult } from "../api/pagination";
import { useNamespaceLegacy } from "./useNamespaceLegacy";

interface UseProvisioningKeysParams {
  page?: number;
  perPage?: number;
  sortBy?: string;
  orderBy?: "asc" | "desc";
  enabled?: boolean;
}

/**
 * A page of the namespace's provisioning keys, newest first. With `enabled` false the query stays
 * idle and the page reads as empty.
 */
export function useProvisioningKeys({
  page = 1,
  perPage = 10,
  sortBy = "created_at",
  orderBy = "desc",
  enabled = true,
}: UseProvisioningKeysParams = {}) {
  const options = {
    query: { page, per_page: perPage, sort_by: sortBy, order_by: orderBy },
  } satisfies { query: ProvisioningKeyListData["query"] };

  const result = useQuery<PaginatedResult<ProvisioningKey>>({
    queryKey: provisioningKeyListQueryKey(options),
    queryFn: paginatedQueryFn(provisioningKeyList, options),
    enabled,
  });

  return {
    provisioningKeys: result.data?.data ?? [],
    totalCount: result.data?.totalCount ?? 0,
    isLoading: result.isLoading,
    error: result.error,
  };
}

/**
 * The namespace's built-in legacy key, which a legacy namespace's tenant-only enrollments land
 * under, so its activity page is the pending queue. Null until the list has loaded, and idle
 * unless the active namespace is legacy, since only a legacy namespace needs it.
 */
export function useLegacyProvisioningKey() {
  const enabled = useNamespaceLegacy() === true;
  const { provisioningKeys } = useProvisioningKeys({ perPage: 100, enabled });

  return provisioningKeys.find((key) => key.type === "legacy") ?? null;
}
