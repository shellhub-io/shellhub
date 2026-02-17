import { create } from "zustand";
import {
  getNamespaces,
  getNamespace,
  getNamespaceToken,
  createNamespace as createNamespaceApi,
  updateNamespace as updateNamespaceApi,
  deleteNamespace as deleteNamespaceApi,
  leaveNamespace as leaveNamespaceApi,
} from "../api/namespaces";
import { useAuthStore } from "./authStore";
import { useConnectivityStore } from "./connectivityStore";
import { Namespace } from "../types/namespace";

interface NamespacesState {
  namespaces: Namespace[];
  currentNamespace: Namespace | null;
  loading: boolean;
  loaded: boolean;
  error: string | null;
  fetch: () => Promise<void>;
  fetchCurrent: (tenantId: string) => Promise<void>;
  switchNamespace: (tenantId: string) => Promise<void>;
  createNamespace: (name: string) => Promise<void>;
  updateNamespace: (
    tenantId: string,
    data: {
      name?: string;
      settings?: { session_record?: boolean; connection_announcement?: string };
    },
  ) => Promise<void>;
  deleteNamespace: (tenantId: string) => Promise<void>;
  leaveNamespace: (tenantId: string) => Promise<void>;
}

export const useNamespacesStore = create<NamespacesState>((set) => ({
  namespaces: [],
  currentNamespace: null,
  loading: false,
  loaded: false,
  error: null,

  fetch: async () => {
    set({ loading: true, error: null });
    try {
      const namespaces = await getNamespaces();
      set({ namespaces, loading: false, loaded: true });
    } catch {
      const apiDown = !useConnectivityStore.getState().apiReachable;
      set({
        loading: false,
        loaded: !apiDown,
        error: apiDown
          ? "Unable to reach the API"
          : "Failed to load namespaces",
      });
    }
  },

  fetchCurrent: async (tenantId: string) => {
    try {
      const ns = await getNamespace(tenantId);
      set({ currentNamespace: ns });
    } catch {
      set({ currentNamespace: null });
    }
  },

  switchNamespace: async (tenantId: string) => {
    const { token, role } = await getNamespaceToken(tenantId);
    useAuthStore.getState().setSession({ token, tenant: tenantId, role });
    window.location.reload();
  },

  createNamespace: async (name: string) => {
    set({ loading: true, error: null });
    try {
      const ns = await createNamespaceApi(name);
      const { token, role } = await getNamespaceToken(ns.tenant_id);
      useAuthStore.getState().setSession({ token, tenant: ns.tenant_id, role });
      window.location.reload();
    } catch (err: unknown) {
      const message =
        err instanceof Error ? err.message : "Failed to create namespace";
      set({ loading: false, error: message });
      throw err;
    }
  },

  updateNamespace: async (tenantId: string, data) => {
    await updateNamespaceApi(tenantId, data);
    const ns = await getNamespace(tenantId);
    set({ currentNamespace: ns });
  },

  deleteNamespace: async (tenantId: string) => {
    await deleteNamespaceApi(tenantId);
    useAuthStore.getState().logout();
    window.location.replace("/v2/ui/login");
  },

  leaveNamespace: async (tenantId: string) => {
    await leaveNamespaceApi(tenantId);
    useAuthStore.getState().logout();
    window.location.replace("/v2/ui/login");
  },
}));
