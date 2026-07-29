import { useMemo } from "react";
import { useQuery } from "@tanstack/react-query";
import { listAccessPoliciesOptions } from "../client/@tanstack/react-query.gen";
import type { AccessPolicy } from "../client";

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
