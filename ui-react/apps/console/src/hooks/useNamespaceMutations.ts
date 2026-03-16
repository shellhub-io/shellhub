import { useMutation, useQueryClient } from "@tanstack/react-query";
import {
  editNamespaceMutation,
} from "../client/@tanstack/react-query.gen";
import {
  getNamespaceToken,
  createNamespace as createNamespaceSdk,
  deleteNamespace as deleteNamespaceSdk,
  leaveNamespace as leaveNamespaceSdk,
} from "../client";
import { useAuthStore } from "../stores/authStore";

function useInvalidateNamespaces() {
  const queryClient = useQueryClient();
  return () => queryClient.invalidateQueries({ predicate: (query) => {
    const key = query.queryKey[0];
    if (typeof key === "object" && key !== null && "_id" in key) {
      const id = (key as { _id: string })._id;
      return id === "getNamespaces" || id === "getNamespace";
    }
    return false;
  } });
}

export function useEditNamespace() {
  const invalidate = useInvalidateNamespaces();
  return useMutation({
    ...editNamespaceMutation(),
    onSuccess: invalidate,
  });
}

export function useSwitchNamespace() {
  return useMutation({
    mutationFn: async (tenantId: string) => {
      const { data } = await getNamespaceToken({
        path: { tenant: tenantId },
        throwOnError: true,
      });
      useAuthStore.getState().setSession({
        token: data.token,
        tenant: tenantId,
        role: data.role,
      });
      window.location.reload();
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
