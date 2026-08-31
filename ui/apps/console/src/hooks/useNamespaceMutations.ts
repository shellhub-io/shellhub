import { useMutation } from "@tanstack/react-query";
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
import { useInvalidateByIds } from "./useInvalidateQueries";

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

/**
 * Switches the active namespace, which re-issues the token and lands on redirectTo. Everything
 * cached belongs to the previous namespace, so this is a navigation rather than a refetch.
 */
export function useSwitchNamespace() {
  return useMutation({
    mutationFn: async ({
      tenantId,
      redirectTo,
    }: {
      tenantId: string;
      redirectTo?: string;
    }) => {
      const { data } = await getNamespaceToken({
        path: { tenant: tenantId },
        throwOnError: true,
      });
      window.location.href = redirectTo ?? "/dashboard";
      useAuthStore.getState().setSession({
        token: data.token,
        tenant: tenantId,
        role: data.role,
      });
    },
  });
}

/**
 * Creates a namespace and switches into it, so the user ends up inside what they just made.
 */
export function useCreateNamespace() {
  return useMutation({
    mutationFn: async (name: string) => {
      const { data: ns } = await createNamespaceSdk({
        body: { name },
        throwOnError: true,
      });
      const { data } = await getNamespaceToken({
        path: { tenant: ns.tenant_id },
        throwOnError: true,
      });
      const pendingCode = consumePendingDeviceCode();
      window.location.href = pendingCode
        ? `/accept-device?code=${encodeURIComponent(pendingCode)}`
        : "/dashboard";
      useAuthStore.getState().setSession({
        token: data.token,
        tenant: ns.tenant_id,
        role: data.role,
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
