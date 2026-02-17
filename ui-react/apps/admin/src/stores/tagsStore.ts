import { create } from "zustand";
import {
  getTags,
  createTag as createTagApi,
  updateTag as updateTagApi,
  deleteTag as deleteTagApi,
} from "../api/tags";
import { Tag } from "../types/tag";

interface TagsState {
  tags: Tag[];
  totalCount: number;
  loading: boolean;
  error: string | null;
  page: number;
  perPage: number;
  fetch: (page?: number, perPage?: number) => Promise<void>;
  create: (name: string) => Promise<void>;
  update: (currentName: string, newName: string) => Promise<void>;
  remove: (name: string) => Promise<void>;
}

export const useTagsStore = create<TagsState>((set, get) => ({
  tags: [],
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
      const { data, totalCount } = await getTags(p, pp);
      set({ tags: data, totalCount, loading: false });
    } catch {
      set({ loading: false, error: "Failed to load tags" });
    }
  },

  create: async (name) => {
    await createTagApi(name);
    await get().fetch();
  },

  update: async (currentName, newName) => {
    await updateTagApi(currentName, newName);
    await get().fetch();
  },

  remove: async (name) => {
    await deleteTagApi(name);
    const { page, perPage, tags } = get();
    const newPage = tags.length === 1 && page > 1 ? page - 1 : page;
    await get().fetch(newPage, perPage);
  },
}));
