import { useMutation } from "@tanstack/react-query";
import {
  editNamespaceMutation,
  setDeviceAutoAcceptMutation,
} from "../client/@tanstack/react-query.gen";
import {
  getNamespaceToken,
  createNamespace as createNamespaceSdk,
  deleteNamespace as deleteNamespaceSdk,
  leaveNamespace as leaveNamespaceSdk,
} from "../client";
import { useAuthStore } from "../stores/authStore";
import { useInvalidateByIds } from "./useInvalidateQueries";

export function useEditNamespace() {
  const invalidate = useInvalidateByIds("getNamespaces", "getNamespace");
  return useMutation({
    ...editNamespaceMutation(),
    onSuccess: invalidate,
  });
}

export function useSetDeviceAutoAccept() {
  const invalidate = useInvalidateByIds("getNamespaces", "getNamespace");
  return useMutation({
    ...setDeviceAutoAcceptMutation(),
    onSuccess: invalidate,
  });
}

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
      useAuthStore.getState().setSession({
        token: data.token,
        tenant: tenantId,
        role: data.role,
      });
      if (redirectTo) window.location.href = redirectTo;
      else window.location.reload();
    },
  });
}

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
      useAuthStore.getState().setSession({
        token: data.token,
        tenant: ns.tenant_id,
        role: data.role,
      });
      window.location.reload();
    },
  });
}

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
