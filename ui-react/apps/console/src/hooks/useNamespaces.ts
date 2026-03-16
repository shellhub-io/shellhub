import { useQuery } from "@tanstack/react-query";
import {
  getNamespacesOptions,
  getNamespaceOptions,
} from "../client/@tanstack/react-query.gen";
import type { Namespace as GeneratedNamespace, NamespaceMemberRole } from "../client";

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
