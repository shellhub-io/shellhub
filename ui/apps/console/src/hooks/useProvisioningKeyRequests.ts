import { useMutation, useQuery } from "@tanstack/react-query";
import {
  provisioningKeyList,
  provisioningKeyListQueryKey,
  provisioningKeyRevealOptions,
  provisioningKeyUpdateMutation,
} from "@/client";
import { paginatedQueryOptions } from "@/api/pagination";
import { useInvalidateByIds } from "@/hooks/useInvalidateQueries";

/**
 * The namespace's first hundred provisioning keys, newest first, for finding a key by its id.
 * Empty while loading and when the request fails.
 */
export function useProvisioningKeys() {
  const { data, isLoading } = useQuery(
    paginatedQueryOptions(provisioningKeyList, provisioningKeyListQueryKey, {
      query: { page: 1, per_page: 100, sort_by: "created_at", order_by: "desc" },
    }),
  );

  return { provisioningKeys: data?.data ?? [], isLoading };
}

/**
 * Updates a provisioning key's name or limits. The key itself is not re-issued.
 */
export function useUpdateProvisioningKey() {
  const invalidate = useInvalidateByIds("provisioningKeyList");
  return useMutation({
    ...provisioningKeyUpdateMutation(),
    onSuccess: invalidate,
  });
}

/**
 * A provisioning key's plaintext, fetched on demand. Nothing loads with the key list: the query
 * fires only once a key is named and the caller opts in with enabled, so the secret is decrypted
 * for an open install command or an explicit Show, and dropped from cache as soon as nothing
 * reads it.
 */
export function useRevealProvisioningKey(name: string | null, enabled = true) {
  const result = useQuery({
    ...provisioningKeyRevealOptions({ path: { key: name ?? "" } }),
    enabled: !!name && enabled,
    gcTime: 0,
  });

  return {
    key: result.data?.key ?? "",
    isLoading: result.isLoading,
    error: result.error,
  };
}
