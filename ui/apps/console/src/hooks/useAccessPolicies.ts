import { useMemo } from "react";
import { useQuery } from "@tanstack/react-query";
import { listAccessPoliciesOptions, type AccessPolicy } from "../client";

/**
 * The namespace's access policies. Returns an empty array rather than undefined while loading,
 * so a caller can map over it without a guard.
 */
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
