import { useQuery } from "@tanstack/react-query";
import { getNamespaceSupportOptions } from "@/client/@tanstack/react-query.gen";

export function useSupportIdentifier(
  tenantId: string | null | undefined,
  enabled: boolean,
) {
  const result = useQuery({
    ...getNamespaceSupportOptions({ path: { tenant: tenantId ?? "" } }),
    enabled: enabled && !!tenantId,
    // One retry covers transient network blips during the initial fetch
    // without dragging out the spinner when the operator simply hasn't set
    // SHELLHUB_CHATWOOT_IDENTITY_KEY (4xx is settled within ~1s either way).
    retry: 1,
  });

  return {
    identifier: result.data?.identifier ?? null,
    isLoading: result.isLoading,
    isError: result.isError,
  };
}
