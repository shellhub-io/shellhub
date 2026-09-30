import { useQuery } from "@tanstack/react-query";
import { provisioningKeyRevealOptions } from "../client";

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
