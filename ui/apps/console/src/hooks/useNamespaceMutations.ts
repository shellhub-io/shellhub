import { useMutation } from "@tanstack/react-query";
import {
  getNamespaceToken,
  createNamespace as createNamespaceSdk,
} from "@/client/api";
import { useAuthStore } from "../stores/authStore";
import { consumePendingDeviceCode } from "@/utils/navigation";

/** Exchanges a tenant token and redirects, updating the auth store. */
export function useSwitchNamespace() {
  return useMutation({
    mutationFn: async ({
      tenantId,
      redirectTo,
    }: {
      tenantId: string;
      redirectTo?: string;
    }) => {
      const auth = await getNamespaceToken(tenantId);
      window.location.href = redirectTo ?? "/dashboard";
      useAuthStore.getState().setSession({
        token: auth.token,
        tenant: tenantId,
        role: auth.role,
      });
    },
  });
}

/** Creates a namespace, exchanges its token, and redirects to the dashboard. */
export function useCreateNamespace() {
  return useMutation({
    mutationFn: async (name: string) => {
      const { tenant_id: tenant } = await createNamespaceSdk({ name });
      const auth = await getNamespaceToken(tenant);
      const pendingCode = consumePendingDeviceCode();
      window.location.href = pendingCode
        ? `/accept-device?code=${encodeURIComponent(pendingCode)}`
        : "/dashboard";
      useAuthStore.getState().setSession({
        token: auth.token,
        tenant,
        role: auth.role,
      });
    },
  });
}
