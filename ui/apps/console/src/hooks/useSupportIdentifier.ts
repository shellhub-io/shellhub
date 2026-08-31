import { useQuery } from "@tanstack/react-query";
import { getNamespaceSupportOptions } from "@/client";

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
