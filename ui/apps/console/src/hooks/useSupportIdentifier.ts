import { useQuery } from "@tanstack/react-query";
import { getNamespaceSupportOptions } from "@/client";

/**
 * The namespace's support identifier, used to attach a chat conversation to the account. Idle
 * without a tenant, since the identifier is per namespace.
 */
export function useSupportIdentifier(
  tenantId: string | null | undefined,
  enabled: boolean,
) {
  const result = useQuery({
    ...getNamespaceSupportOptions({ path: { tenant: tenantId ?? "" } }),
    enabled: enabled && !!tenantId,
    retry: 1,
  });

  return {
    identifier: result.data?.identifier ?? null,
    isLoading: result.isLoading,
    isError: result.isError,
  };
}
