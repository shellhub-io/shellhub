import { useEffect } from "react";
import { useQuery } from "@tanstack/react-query";
import {
  getNamespacesOptions,
  getNamespaceOptions,
  getNamespaceTokenOptions,
} from "../client/@tanstack/react-query.gen";
import type { Namespace as GeneratedNamespace, NamespaceMemberRole } from "../client";
import { useAuthStore } from "../stores/authStore";

export type Namespace = GeneratedNamespace & { type?: string };

export interface NamespaceMember {
  id: string;
  role: NamespaceMemberRole;
  email: string;
  added_at?: string;
  status?: "accepted" | "pending";
}

export function useNamespaces() {
  const result = useQuery({
    ...getNamespacesOptions({ query: { page: 1, per_page: 30 } }),
  });

  return {
    namespaces: (result.data ?? []) as Namespace[],
    isLoading: result.isLoading,
    error: result.error,
    refetch: result.refetch,
  };
}

/**
 * Fetches a fresh namespace token on every cold start so that
 * authStore.role is always populated from the server — covering first
 * login (where role is null) and stale-localStorage scenarios.
 *
 * Should be called once at the app's authenticated boundary
 * (NamespaceGuard).
 */
export function useInitRole() {
  const tenant = useAuthStore((s) => s.tenant);

  const { data } = useQuery({
    ...getNamespaceTokenOptions({ path: { tenant: tenant ?? "" } }),
    enabled: !!tenant,
  });

  useEffect(() => {
    if (!data || !tenant) return;
    useAuthStore.getState().setSession({ token: data.token, tenant, role: data.role });
  }, [data, tenant]);
}

export function useNamespace(tenantId: string) {
  const result = useQuery({
    ...getNamespaceOptions({ path: { tenant: tenantId } }),
    enabled: !!tenantId,
  });

  return {
    namespace: (result.data ?? null) as Namespace | null,
    isLoading: result.isLoading,
    error: result.error,
    refetch: result.refetch,
  };
}
