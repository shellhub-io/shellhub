import { useMutation } from "@tanstack/react-query";
import {
  getNamespaceToken,
  createNamespace as createNamespaceSdk,
  deleteNamespace as deleteNamespaceSdk,
  leaveNamespace as leaveNamespaceSdk,
} from "@/client/api";
import { useAuthStore } from "../stores/authStore";
import { consumePendingDeviceCode } from "@/utils/navigation";

/**
 *
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
 *
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
 *
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
 *
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
