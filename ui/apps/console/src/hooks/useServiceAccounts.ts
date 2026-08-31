import { useMemo } from "react";
import { useQuery } from "@tanstack/react-query";
import { listServiceAccountsOptions, type ServiceAccount } from "../client";

/**
 * The namespace's service accounts. Returns an empty array while loading, so a caller can map
 * over it without a guard.
 */
export function useServiceAccounts() {
  const result = useQuery(listServiceAccountsOptions());

  const serviceAccounts = useMemo<ServiceAccount[]>(
    () => result.data ?? [],
    [result.data],
  );

  return {
    serviceAccounts,
    isLoading: result.isLoading,
    error: result.error,
  };
}
