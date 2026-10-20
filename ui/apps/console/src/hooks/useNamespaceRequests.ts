import { useEffect } from "react";
import {
  useMutation,
  useQuery,
  useQueryClient,
  type QueryClient,
} from "@tanstack/react-query";
import { useNavigate } from "react-router-dom";
import {
  createNamespace as createNamespaceSdk,
  deleteNamespace as deleteNamespaceSdk,
  editNamespaceMutation,
  getNamespaceOptions,
  getNamespacesOptions,
  getNamespaceTokenOptions,
  getStatusDevicesOptions,
  leaveNamespace as leaveNamespaceSdk,
  listNamespaceMembersOptions,
  type Namespace as GeneratedNamespace,
  type NamespaceMemberRole,
} from "@/client";
import { useAuthStore } from "@/stores/authStore";
import { useVaultStore } from "@/stores/vaultStore";
import { consumePendingDeviceCode } from "@/utils/navigation";
import { queryOperationId, useInvalidateByIds } from "@/hooks/useInvalidateQueries";

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
 * Fetches a fresh namespace token on every cold start, so the role in authStore comes from the
 * server. That covers a first login, which has no role yet, and a stale role restored from
 * localStorage. Call it once, at the authenticated boundary in NamespaceGuard. Idle while no
 * tenant is set.
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
  };
}

/**
 * The namespace's device counts, by status. Returns null while loading rather than zeroes, so a
 * dashboard does not flash "0 devices" at someone who has some.
 */
export function useStats() {
  const result = useQuery(getStatusDevicesOptions());

  return {
    stats: result.data ?? null,
    isLoading: result.isLoading,
    error: result.error,
  };
}

/**
 * Edits the namespace's settings, refreshing the namespace list and the namespace itself.
 */
export function useEditNamespace() {
  const invalidate = useInvalidateByIds("getNamespaces", "getNamespace");
  return useMutation({
    ...editNamespaceMutation(),
    onSuccess: invalidate,
  });
}

const NAMESPACE_LIST_QUERIES = new Set(["getNamespaces", "getNamespace"]);
const KEPT_ACROSS_NAMESPACES = new Set([
  ...NAMESPACE_LIST_QUERIES,
  "getNamespaceToken",
]);

const isNamespaceList = (queryKey: readonly unknown[]) =>
  NAMESPACE_LIST_QUERIES.has(queryOperationId(queryKey) ?? "");

async function enterNamespace(
  queryClient: QueryClient,
  tenantId: string,
  { land, freshList = false }: { land: () => void; freshList?: boolean },
) {
  const data = await queryClient.fetchQuery({
    ...getNamespaceTokenOptions({ path: { tenant: tenantId } }),
    staleTime: 0,
    retry: false,
  });
  if (freshList) {
    await queryClient.refetchQueries({
      predicate: (query) => isNamespaceList(query.queryKey),
    });
  }
  const leaving = useAuthStore.getState().tenant !== tenantId;
  if (leaving) useVaultStore.getState().lock();
  useAuthStore.getState().setSession({
    token: data.token,
    tenant: tenantId,
    role: data.role,
  });
  if (leaving) void useVaultStore.getState().refreshStatus();
  queryClient.removeQueries({
    predicate: (query) =>
      !KEPT_ACROSS_NAMESPACES.has(queryOperationId(query.queryKey) ?? ""),
  });
  land();
  if (!freshList) {
    void queryClient.invalidateQueries({
      predicate: (query) => isNamespaceList(query.queryKey),
    });
  }
}

/**
 * Makes a namespace the active one in place, without leaving the page. It re-issues the token,
 * locks the vault (each namespace has its own, and the key of the one being left must not
 * encrypt the next one's), swaps the session, reads the state of the next namespace's vault and
 * drops what was cached for the namespace being left, whose keys do not carry the tenant. land
 * runs in the same tick as the swap, so a caller that navigates there renders the new
 * namespace's page with its session at once, and the old page never refetches under the new
 * token. The token lands in the query cache too, where NamespaceGuard's role lookup reads it
 * instead of asking again. The namespace list is refreshed in the background rather than
 * dropped, because dropping it would send NamespaceGuard back to its loading screen and unmount
 * the layout, open terminals included. Re-entering the namespace already active leaves its vault
 * as it is. Rejects, leaving the session as it was, when the token cannot be issued.
 */
export function useEnterNamespace() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ tenantId, land }: { tenantId: string; land: () => void }) =>
      enterNamespace(queryClient, tenantId, { land }),
  });
}

/**
 * Enters a namespace, as useEnterNamespace does, and lands on redirectTo. The namespace list is
 * refreshed before rather than after: a namespace just joined has to be in it before
 * NamespaceGuard looks.
 */
export function useSwitchNamespace() {
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  return useMutation({
    mutationFn: ({
      tenantId,
      redirectTo,
    }: {
      tenantId: string;
      redirectTo?: string;
    }) =>
      enterNamespace(queryClient, tenantId, {
        freshList: true,
        land: () => void navigate(redirectTo ?? "/dashboard"),
      }),
  });
}

/**
 * Creates a namespace and enters it, so the user ends up inside what they just made.
 */
export function useCreateNamespace() {
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  return useMutation({
    mutationFn: async (name: string) => {
      const { data: ns } = await createNamespaceSdk({
        body: { name },
        throwOnError: true,
      });
      const pendingCode = consumePendingDeviceCode();
      await enterNamespace(queryClient, ns.tenant_id, {
        freshList: true,
        land: () =>
          void navigate(
            pendingCode
              ? `/accept-device?code=${encodeURIComponent(pendingCode)}`
              : "/dashboard",
          ),
      });
    },
  });
}

/**
 * Deletes a namespace along with everything in it. Irreversible.
 */
export function useDeleteNamespace() {
  return useMutation({
    mutationFn: async (tenantId: string) => {
      await deleteNamespaceSdk({
        path: { tenant: tenantId },
        throwOnError: true,
      });
      useAuthStore.getState().logout();
      window.location.replace("/login");
    },
  });
}

/**
 * Leaves a namespace. Unlike deleting, the namespace survives — this only removes the caller,
 * and an owner cannot be the one to go.
 */
export function useLeaveNamespace() {
  return useMutation({
    mutationFn: async (tenantId: string) => {
      await leaveNamespaceSdk({
        path: { tenant: tenantId },
        throwOnError: true,
      });
      useAuthStore.getState().logout();
      window.location.replace("/login");
    },
  });
}
