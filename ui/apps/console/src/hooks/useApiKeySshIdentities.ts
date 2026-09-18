import { useQuery } from "@tanstack/react-query";
import { listApiKeySshIdentitiesOptions, type SshIdentity } from "../client";

/**
 * The SSH credentials an API key owns. Asked for by key name, which is what its routes are
 * addressed by, and answered only to a caller who may manage identities.
 */
export function useApiKeySshIdentities(name: string, enabled = true) {
  const result = useQuery({
    ...listApiKeySshIdentitiesOptions({ path: { name } }),
    enabled: enabled && !!name,
  });

  const identities: SshIdentity[] = result.data ?? [];

  return {
    identities,
    isLoading: result.isLoading,
    error: result.error,
  };
}
