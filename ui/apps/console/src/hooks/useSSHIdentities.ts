import { useQuery } from "@tanstack/react-query";
import { listSshIdentitiesOptions, type SshIdentity } from "../client";

/**
 * The SSH identities the caller may see: their own, or every member's when they hold the
 * permission to manage them. The server decides which, so there is nothing to ask for here.
 * enabled holds the request back until the caller needs the list.
 */
export function useSSHIdentities({
  enabled = true,
}: { enabled?: boolean } = {}) {
  const result = useQuery({ ...listSshIdentitiesOptions(), enabled });

  const identities: SshIdentity[] = result.data ?? [];

  return {
    identities,
    isLoading: result.isLoading,
    error: result.error,
  };
}
