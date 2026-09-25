import {
  useMutation,
  useQueryClient,
  type QueryClient,
} from "@tanstack/react-query";
import { useNavigate } from "react-router-dom";
import {
  editNamespaceMutation,
  setSshAccessModeMutation,
  getNamespaceToken,
  createNamespace as createNamespaceSdk,
  deleteNamespace as deleteNamespaceSdk,
  leaveNamespace as leaveNamespaceSdk,
} from "../client";
import { useAuthStore } from "../stores/authStore";
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

async function enterNamespace(
  queryClient: QueryClient,
  tenantId: string,
): Promise<{ refreshed: Promise<void> }> {
  const { data } = await getNamespaceToken({
    path: { tenant: tenantId },
    throwOnError: true,
  });
  useAuthStore.getState().setSession({
    token: data.token,
    tenant: tenantId,
    role: data.role,
  });
  queryClient.removeQueries({
    predicate: (query) =>
      !KEPT_ACROSS_NAMESPACES.has(queryOperationId(query.queryKey) ?? ""),
  });
  const refreshed = queryClient.invalidateQueries({
    predicate: (query) =>
      NAMESPACE_LIST_QUERIES.has(queryOperationId(query.queryKey) ?? ""),
  });
  return { refreshed };
}

/**
 * Makes a namespace the active one in place, without leaving the page: it re-issues the token and
 * drops what was cached for the namespace being left, since those keys do not carry the tenant.
 * The namespace list is refreshed rather than dropped, because dropping it would send
 * NamespaceGuard back to its loading screen and unmount the layout, open terminals included.
 * That refresh runs in the background, so the switch lands as soon as the token does. Rejects,
 * leaving the session as it was, when the token cannot be issued.
 */
export function useEnterNamespace() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (tenantId: string) => {
      await enterNamespace(queryClient, tenantId);
    },
  });
}

/**
 * Enters a namespace, as useEnterNamespace does, and lands on redirectTo once the namespace list
 * is fresh: a namespace just joined has to be in it before NamespaceGuard looks.
 */
export function useSwitchNamespace() {
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  return useMutation({
    mutationFn: async ({
      tenantId,
      redirectTo,
    }: {
      tenantId: string;
      redirectTo?: string;
    }) => {
      const { refreshed } = await enterNamespace(queryClient, tenantId);
      await refreshed;
      await navigate(redirectTo ?? "/dashboard");
    },
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
      const { refreshed } = await enterNamespace(queryClient, ns.tenant_id);
      await refreshed;
      const pendingCode = consumePendingDeviceCode();
      await navigate(
        pendingCode
          ? `/accept-device?code=${encodeURIComponent(pendingCode)}`
          : "/dashboard",
      );
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
