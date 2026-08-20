import { useMemo } from "react";
import { useQuery } from "@tanstack/react-query";
import { listServiceAccountsOptions, type ServiceAccount } from "../client";

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
