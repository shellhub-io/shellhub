import { useQuery } from "@tanstack/react-query";
import { apiKeyList, apiKeyListQueryKey } from "@/client";
import { paginatedQueryOptions } from "@/api/pagination";

/**
 * The namespace's first hundred API keys, newest first, for naming a key by its id. Empty while
 * loading and when the request fails.
 */
export function useApiKeys() {
  const { data } = useQuery(
    paginatedQueryOptions(apiKeyList, apiKeyListQueryKey, {
      query: { page: 1, per_page: 100, sort_by: "created_at", order_by: "desc" },
    }),
  );

  return data?.data ?? [];
}
