import { useMemo } from "react";
import { useQuery } from "@tanstack/react-query";
import {
  getFirewallRules as getFirewallRulesSdk,
  type GetFirewallRulesData,
  type FirewallRulesResponse,
} from "../client";
import { getFirewallRulesQueryKey } from "../client/@tanstack/react-query.gen";
import { paginatedQueryFn, type PaginatedResult } from "../api/pagination";

export interface FirewallFilter {
  hostname?: string;
  tags?: string[];
}

export interface FirewallRule {
  id: string;
  tenant_id: string;
  priority: number;
  action: "allow" | "deny";
  active: boolean;
  source_ip: string;
  username: string;
  filter: FirewallFilter;
}

export function normalizeFirewallRule(
  rule: FirewallRulesResponse & { id: string },
): FirewallRule {
  let filter: FirewallFilter;
  if (
    "tags" in rule.filter
    && Array.isArray(rule.filter.tags)
    && rule.filter.tags.length > 0
  ) {
    filter = {
      tags: rule.filter.tags.map((t) =>
        typeof t === "object" && t !== null && "name" in t ? t.name : String(t),
      ),
    };
  } else if ("hostname" in rule.filter) {
    filter = { hostname: rule.filter.hostname };
  } else {
    filter = { hostname: ".*" };
  }

  return {
    ...rule,
    filter,
  };
}

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
