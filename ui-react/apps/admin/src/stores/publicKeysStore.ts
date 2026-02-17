import { create } from "zustand";
import {
  getPublicKeys,
  createPublicKey as createPublicKeyApi,
  updatePublicKey as updatePublicKeyApi,
  deletePublicKey as deletePublicKeyApi,
} from "../api/publicKeys";
import { PublicKey, PublicKeyFilter } from "../types/publicKey";

interface PublicKeysState {
  publicKeys: PublicKey[];
  totalCount: number;
  loading: boolean;
  error: string | null;
  page: number;
  perPage: number;
  fetch: (page?: number, perPage?: number) => Promise<void>;
  create: (payload: {
    name: string;
    data: string;
    username: string;
    filter: PublicKeyFilter;
  }) => Promise<void>;
  update: (
    fingerprint: string,
    payload: { name: string; username: string; filter: PublicKeyFilter },
  ) => Promise<void>;
  remove: (fingerprint: string) => Promise<void>;
}

export const usePublicKeysStore = create<PublicKeysState>((set, get) => ({
  publicKeys: [],
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
      const { data, totalCount } = await getPublicKeys(p, pp);
      set({ publicKeys: data, totalCount, loading: false });
    } catch {
      set({ loading: false, error: "Failed to load public keys" });
    }
  },

  create: async (payload) => {
    await createPublicKeyApi(payload);
    await get().fetch();
  },

  update: async (fingerprint, payload) => {
    await updatePublicKeyApi(fingerprint, payload);
    await get().fetch();
  },

  remove: async (fingerprint) => {
    await deletePublicKeyApi(fingerprint);
    const { page, perPage, publicKeys } = get();
    const newPage = publicKeys.length === 1 && page > 1 ? page - 1 : page;
    await get().fetch(newPage, perPage);
  },
}));
