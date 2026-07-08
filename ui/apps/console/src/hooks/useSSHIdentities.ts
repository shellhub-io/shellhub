import { useMemo } from "react";
import { useQuery } from "@tanstack/react-query";
import { listSshIdentitiesOptions } from "../client/@tanstack/react-query.gen";
import type { SshIdentity } from "../client";

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
