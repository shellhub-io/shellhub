import { useMemo } from "react";
import { useQuery } from "@tanstack/react-query";
import {
  getFirewallRulesAdmin,
  type GetFirewallRulesAdminData,
  type FirewallRulesResponse,
} from "../client";
import {
  getFirewallRulesAdminQueryKey,
  getFirewallRuleAdminOptions,
} from "../client/@tanstack/react-query.gen";
import { paginatedQueryFn, type PaginatedResult } from "../api/pagination";
import { useAuthStore } from "../stores/authStore";
import { isSdkError } from "../api/errors";

import { normalizeFirewallRule } from "./useFirewallRules";
export type { FirewallRule, FirewallFilter } from "./useFirewallRules";

interface UseAdminFirewallRulesParams {
  page?: number;
  perPage?: number;
}

export function useAdminFirewallRules({
  page = 1,
  perPage = 10,
}: UseAdminFirewallRulesParams = {}) {
  const isAdmin = useAuthStore((s) => s.isAdmin);

  const options = {
    query: {
      page,
      per_page: perPage,
    } satisfies GetFirewallRulesAdminData["query"],
  };

  const result = useQuery<PaginatedResult<FirewallRulesResponse>>({
    queryKey: getFirewallRulesAdminQueryKey(options),
    queryFn: paginatedQueryFn(getFirewallRulesAdmin, options),
    enabled: isAdmin,
    staleTime: 5 * 60 * 1000,
    retry: (count, err) =>
      isSdkError(err) && err.status === 401 ? false : count < 1,
    refetchOnWindowFocus: false,
  });

  const rules = useMemo(
    () =>
      result.data?.data
        .filter(
          (r): r is FirewallRulesResponse & { id: string } =>
            r.id !== undefined,
        )
        .map(normalizeFirewallRule) ?? [],
    [result.data],
  );

  return {
    rules,
    totalCount: result.data?.totalCount ?? 0,
    isLoading: result.isLoading,
    error: result.error,
  };
}

export function useAdminFirewallRule(id: string) {
  const isAdmin = useAuthStore((s) => s.isAdmin);

  return useQuery({
    ...getFirewallRuleAdminOptions({ path: { id } }),
    enabled: isAdmin && !!id,
    staleTime: 5 * 60 * 1000,
    retry: (count, err) =>
      isSdkError(err) && err.status === 401 ? false : count < 1,
    refetchOnWindowFocus: false,
    select: (data) => {
      const rule = data;
      if (!rule.id) return undefined;
      return normalizeFirewallRule(
        rule as FirewallRulesResponse & { id: string },
      );
    },
  });
}
