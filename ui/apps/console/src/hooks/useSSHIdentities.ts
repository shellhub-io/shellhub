import { useQuery } from "@tanstack/react-query";
import { listSshIdentitiesOptions, type SshIdentity } from "../client";

/**
 * The SSH identities the caller may see: their own, or every member's when they hold the
 * permission to manage them. The server decides which, so there is nothing to ask for here.
 */
export function useSSHIdentities() {
  const result = useQuery(listSshIdentitiesOptions());

  const identities: SshIdentity[] = result.data ?? [];

  return {
    identities,
    isLoading: result.isLoading,
    error: result.error,
  };
}
