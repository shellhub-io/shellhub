import { create } from "zustand";
import {
  getFirewallRules,
  createFirewallRule as createApi,
  updateFirewallRule as updateApi,
  deleteFirewallRule as deleteApi,
} from "../api/firewallRules";
import { FirewallRule, FirewallFilter } from "../types/firewallRule";

interface FirewallRulesState {
  rules: FirewallRule[];
  totalCount: number;
  loading: boolean;
  error: string | null;
  page: number;
  perPage: number;
  fetch: (page?: number, perPage?: number) => Promise<void>;
  create: (payload: {
    priority: number;
    action: string;
    active: boolean;
    source_ip: string;
    username: string;
    filter: FirewallFilter;
  }) => Promise<void>;
  update: (
    id: string,
    payload: {
      priority: number;
      action: string;
      active: boolean;
      source_ip: string;
      username: string;
      filter: FirewallFilter;
    },
  ) => Promise<void>;
  remove: (id: string) => Promise<void>;
}

export const useFirewallRulesStore = create<FirewallRulesState>((set, get) => ({
  rules: [],
  totalCount: 0,
  loading: false,
  error: null,
  page: 1,
  perPage: 10,

  fetch: async (page?: number, perPage?: number) => {
    const p = page ?? get().page;
    const pp = perPage ?? get().perPage;
    set({ loading: true, error: null, page: p, perPage: pp });
    try {
      const { data, totalCount } = await getFirewallRules(p, pp);
      set({ rules: data, totalCount, loading: false });
    } catch {
      set({ loading: false, error: "Failed to load firewall rules" });
    }
  },

  create: async (payload) => {
    await createApi(payload);
    await get().fetch();
  },

  update: async (id, payload) => {
    await updateApi(id, payload);
    await get().fetch();
  },

  remove: async (id) => {
    await deleteApi(id);
    const { page, perPage, rules } = get();
    const newPage = rules.length === 1 && page > 1 ? page - 1 : page;
    await get().fetch(newPage, perPage);
  },
}));
