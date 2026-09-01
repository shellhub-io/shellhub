import { useEffect } from "react";
import { useQuery } from "@tanstack/react-query";
import {
  type Namespace as GeneratedNamespace,
  type NamespaceMemberRole,
  getNamespacesOptions,
  getNamespaceOptions,
  getNamespaceTokenOptions,
  listNamespaceMembersOptions,
} from "../client";
import { useAuthStore } from "../stores/authStore";

/**
 * A namespace as the console uses it: the generated model plus the type the cloud API adds and
 * the spec does not describe.
 */
export type Namespace = GeneratedNamespace & { type?: string };

/**
 * A member of a namespace. status tells an accepted member from an invitation not yet taken up;
 * both appear in the same list, and only the first has an id that means anything.
 */
export interface NamespaceMember {
  id: string;
  role: NamespaceMemberRole;
  email: string;
  added_at?: string;
  status?: "accepted" | "pending";
  account_status?: "confirmed" | "not-confirmed";
  awaiting_approval?: boolean;
}

/**
 * Every namespace the user belongs to. Fetched as one page of a hundred — the switcher shows
 * them all, and nobody is in more.
 */
export function useNamespaces() {
  const result = useQuery({
    ...getNamespacesOptions({ query: { page: 1, per_page: 100 } }),
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
    useAuthStore
      .getState()
      .setSession({ token: data.token, tenant, role: data.role });
  }, [data, tenant]);
}

/**
 * One namespace by tenant id. Idle until an id is given.
 */
export function useNamespace(tenantId: string) {
  const result = useQuery({
    ...getNamespaceOptions({ path: { tenant: tenantId } }),
    enabled: !!tenantId,
  });

  return {
    namespace: result.data ?? null,
    isLoading: result.isLoading,
    error: result.error,
    refetch: result.refetch,
  };
}

/**
 * Lists a namespace's members with their full identity (name, username, email)
 * and a flattened account status. Backs the members table; the (cloud/enterprise)
 * pending invitations are fetched separately and merged in the component.
 * Member lists are small, so a single large page is fetched (no pagination UI).
 */
export function useNamespaceMembers(tenantId: string) {
  const result = useQuery({
    ...listNamespaceMembersOptions({
      path: { tenant: tenantId },
      query: { page: 1, per_page: 100 },
    }),
    enabled: !!tenantId,
  });

  return {
    members: result.data ?? [],
    isLoading: result.isLoading,
    error: result.error,
    refetch: result.refetch,
  };
}
