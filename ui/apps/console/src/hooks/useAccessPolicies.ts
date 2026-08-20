import { useMemo } from "react";
import { useQuery } from "@tanstack/react-query";
import { listAccessPoliciesOptions, type AccessPolicy } from "../client";

export function useAccessPolicies() {
  const result = useQuery(listAccessPoliciesOptions());

  const policies = useMemo<AccessPolicy[]>(
    () => result.data ?? [],
    [result.data],
  );

  return {
    policies,
    isLoading: result.isLoading,
    error: result.error,
  };
}
