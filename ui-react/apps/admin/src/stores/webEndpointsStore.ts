import { create } from "zustand";
import {
  getWebEndpoints,
  createWebEndpoint as createWebEndpointApi,
  deleteWebEndpoint as deleteWebEndpointApi,
} from "../api/webEndpoints";
import { WebEndpoint, WebEndpointCreate } from "../types/webEndpoint";

interface WebEndpointsState {
  webEndpoints: WebEndpoint[];
  totalCount: number;
  loading: boolean;
  error: string | null;
  page: number;
  perPage: number;
  filter: string;
  fetch: (page?: number, perPage?: number, filter?: string) => Promise<void>;
  create: (payload: WebEndpointCreate) => Promise<void>;
  remove: (address: string) => Promise<void>;
}

export const useWebEndpointsStore = create<WebEndpointsState>((set, get) => ({
  webEndpoints: [],
  totalCount: 0,
  loading: false,
  error: null,
  page: 1,
  perPage: 10,
  filter: "",

  fetch: async (page?: number, perPage?: number, filter?: string) => {
    const p = page ?? get().page;
    const pp = perPage ?? get().perPage;
    const f = filter ?? get().filter;
    set({ loading: true, error: null, page: p, perPage: pp, filter: f });
    try {
      const encoded = f ? btoa(JSON.stringify([{ type: "property", params: { name: "address", operator: "contains", value: f } }])) : undefined;
      const { data, totalCount } = await getWebEndpoints(p, pp, encoded);
      set({ webEndpoints: data, totalCount, loading: false });
    } catch {
      set({ loading: false, error: "Failed to load web endpoints" });
    }
  },

  create: async (payload) => {
    await createWebEndpointApi(payload);
    await get().fetch();
  },

  remove: async (address) => {
    await deleteWebEndpointApi(address);
    const { page, perPage, webEndpoints } = get();
    const newPage = webEndpoints.length === 1 && page > 1 ? page - 1 : page;
    await get().fetch(newPage, perPage);
  },
}));
