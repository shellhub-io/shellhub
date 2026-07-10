import { useEffect } from "react";
import { useQuery } from "@tanstack/react-query";
import {
  getNamespacesOptions,
  getNamespaceOptions,
  getNamespaceTokenOptions,
  listNamespaceMembersOptions,
} from "../client/@tanstack/react-query.gen";
import type {
  Namespace as GeneratedNamespace,
  NamespaceMemberRole,
  MemberView,
} from "../client";
import { useAuthStore } from "../stores/authStore";

export type Namespace = GeneratedNamespace & { type?: string };

export interface NamespaceMember {
  id: string;
  role: NamespaceMemberRole;
  email: string;
  added_at?: string;
  status?: "accepted" | "pending";
  /** Underlying user account status. "not-confirmed" means the invitee hasn't
   *  finished setting up their account yet. Distinct from the invitation
   *  `status` above. */
  account_status?: "confirmed" | "not-confirmed";
  /** Enterprise only: true while a namespace admin provisioned the account but
   *  a system admin hasn't approved it. The account can't sign in until then. */
  awaiting_approval?: boolean;
}

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
    members: (result.data ?? []) as MemberView[],
    isLoading: result.isLoading,
    error: result.error,
    refetch: result.refetch,
  };
}
