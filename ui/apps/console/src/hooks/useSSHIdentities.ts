import { useMemo } from "react";
import { useQuery } from "@tanstack/react-query";
import { listSshIdentitiesOptions, type SshIdentity } from "../client";

/**
 * The SSH identities. By default the caller's own; all includes every identity in the namespace,
 * which needs the permission to see them.
 */
export function useSSHIdentities(all = false) {
  const options = all ? { query: { all: true } } : {};
  const result = useQuery(listSshIdentitiesOptions(options));

  const identities = useMemo<SshIdentity[]>(
    () => result.data ?? [],
    [result.data],
  );

  return {
    identities,
    isLoading: result.isLoading,
    error: result.error,
  };
}
