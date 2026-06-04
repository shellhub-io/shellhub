import { useMemo } from "react";
import { useQuery } from "@tanstack/react-query";
import {
  getFirewallRules as getFirewallRulesSdk,
  type GetFirewallRulesData,
  type FirewallRulesResponse,
} from "../client";
import { getFirewallRulesQueryKey } from "../client/@tanstack/react-query.gen";
import { paginatedQueryFn, type PaginatedResult } from "../api/pagination";

interface UseFirewallRulesParams {
  page?: number;
  perPage?: number;
}

export function useFirewallRules({
  page = 1,
  perPage = 10,
}: UseFirewallRulesParams = {}) {
  const options = {
    query: { page, per_page: perPage } satisfies GetFirewallRulesData["query"],
  };

  const result = useQuery<PaginatedResult<FirewallRulesResponse>>({
    queryKey: getFirewallRulesQueryKey(options),
    queryFn: paginatedQueryFn(getFirewallRulesSdk, options),
  });

  const rules = useMemo(() => result.data?.data ?? [], [result.data]);

  return {
    rules,
    totalCount: result.data?.totalCount ?? 0,
    isLoading: result.isLoading,
    error: result.error,
  };
}
