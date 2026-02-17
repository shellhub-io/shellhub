import { create } from "zustand";
import {
  getApiKeys,
  generateApiKey,
  updateApiKey,
  deleteApiKey,
  type ApiKey,
} from "../api/team";

interface ApiKeysState {
  apiKeys: ApiKey[];
  totalCount: number;
  loading: boolean;
  page: number;
  perPage: number;
  fetch: (page?: number, perPage?: number) => Promise<void>;
  generate: (name: string, role: string, expiresAt: number) => Promise<string>;
  update: (currentName: string, name: string, role: string) => Promise<void>;
  remove: (name: string) => Promise<void>;
  setPage: (page: number) => void;
}

export const useApiKeysStore = create<ApiKeysState>((set, get) => ({
  apiKeys: [],
  totalCount: 0,
  loading: false,
  page: 1,
  perPage: 10,

  fetch: async (page?: number, perPage?: number) => {
    const p = page ?? get().page;
    const pp = perPage ?? get().perPage;
    set({ loading: true });
    try {
      const { data, totalCount } = await getApiKeys(p, pp);
      set({ apiKeys: data, totalCount, loading: false, page: p, perPage: pp });
    } catch {
      set({ loading: false });
    }
  },

  generate: async (name, role, expiresAt) => {
    const result = await generateApiKey({
      name,
      role,
      expires_at: expiresAt,
    });
    await get().fetch();
    return result.id;
  },

  update: async (currentName, name, role) => {
    await updateApiKey(currentName, { name, role });
    await get().fetch();
  },

  remove: async (name) => {
    await deleteApiKey(name);
    await get().fetch();
  },

  setPage: (page) => set({ page }),
}));
