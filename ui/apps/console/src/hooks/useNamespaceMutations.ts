import {
  useMutation,
  useQueryClient,
  type QueryClient,
} from "@tanstack/react-query";
import { useNavigate } from "react-router-dom";
import {
  editNamespaceMutation,
  setSshAccessModeMutation,
  createNamespace as createNamespaceSdk,
  deleteNamespace as deleteNamespaceSdk,
  leaveNamespace as leaveNamespaceSdk,
} from "../client";
import { getNamespaceTokenOptions } from "../client/@tanstack/react-query.gen";
import { useAuthStore } from "../stores/authStore";
import { useVaultStore } from "../stores/vaultStore";
import { consumePendingDeviceCode } from "@/utils/navigation";
import { queryOperationId, useInvalidateByIds } from "./useInvalidateQueries";

/**
 * Edits the namespace's settings.
 */
export function useEditNamespace() {
  const invalidate = useInvalidateByIds("getNamespaces", "getNamespace");
  return useMutation({
    ...editNamespaceMutation(),
    onSuccess: invalidate,
  });
}

/**
 * Sets how SSH access is granted in the namespace. It changes who can reach every device at
 * once, so the namespace queries are refreshed with it.
 */
export function useSetSshAccessMode() {
  const invalidate = useInvalidateByIds("getNamespaces", "getNamespace");
  return useMutation({
    ...setSshAccessModeMutation(),
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
