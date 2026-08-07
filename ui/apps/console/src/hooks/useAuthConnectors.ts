import { useQuery } from "@tanstack/react-query";
import {
  listAuthConnectorsOptions,
  listAuthConnectorsQueryKey,
} from "@/client/@tanstack/react-query.gen";
import type { AuthConnector } from "@/client";

export { listAuthConnectorsQueryKey };

// The login page renders before anyone is authenticated, and an instance with no identity provider
// is the common case, so a failure here must leave the local form usable rather than block it.
export function useAuthConnectors(enabled = true) {
  const result = useQuery({
    ...listAuthConnectorsOptions(),
    enabled,
    staleTime: 5 * 60 * 1000,
    retry: false,
    refetchOnWindowFocus: false,
  });

  return {
    connectors: (result.data ?? []) as AuthConnector[],
    isLoading: result.isLoading,
  };
}
