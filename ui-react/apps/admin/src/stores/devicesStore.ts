import { create } from "zustand";
import {
  getDevices,
  getDevice,
  renameDevice,
  acceptDevice,
  rejectDevice,
  removeDevice,
  addDeviceTag,
  removeDeviceTag,
} from "../api/devices";
import { Device } from "../types/device";

interface DevicesState {
  devices: Device[];
  totalCount: number;
  loading: boolean;
  error: string | null;
  page: number;
  perPage: number;
  status: string;
  filterTags: string[];

  currentDevice: Device | null;
  deviceLoading: boolean;

  fetch: (
    page?: number,
    perPage?: number,
    status?: string,
    filterTags?: string[],
  ) => Promise<void>;
  fetchDevice: (uid: string) => Promise<void>;
  rename: (uid: string, name: string) => Promise<void>;
  accept: (uid: string) => Promise<void>;
  reject: (uid: string) => Promise<void>;
  remove: (uid: string) => Promise<void>;
  addTag: (uid: string, tag: string) => Promise<void>;
  removeTag: (uid: string, tag: string) => Promise<void>;
  setPage: (page: number) => void;
  setStatus: (status: string) => void;
  addFilterTag: (tag: string) => void;
  removeFilterTag: (tag: string) => void;
  clearFilterTags: () => void;
}

export const useDevicesStore = create<DevicesState>((set, get) => ({
  devices: [],
  totalCount: 0,
  loading: false,
  error: null,
  page: 1,
  perPage: 10,
  status: "accepted",
  filterTags: [],

  currentDevice: null,
  deviceLoading: false,

  fetch: async (
    page?: number,
    perPage?: number,
    status?: string,
    filterTags?: string[],
  ) => {
    const p = page ?? get().page;
    const pp = perPage ?? get().perPage;
    const s = status ?? get().status;
    const ft = filterTags ?? get().filterTags;
    set({ loading: true, error: null });
    try {
      const { data, totalCount } = await getDevices(p, pp, s, ft);
      set({
        devices: data,
        totalCount,
        loading: false,
        page: p,
        perPage: pp,
        status: s,
        filterTags: ft,
      });
    } catch {
      set({ loading: false, error: "Failed to load devices" });
    }
  },

  fetchDevice: async (uid: string) => {
    set({ deviceLoading: true, currentDevice: null });
    try {
      const device = await getDevice(uid);
      set({ currentDevice: device, deviceLoading: false });
    } catch {
      set({ deviceLoading: false });
    }
  },

  rename: async (uid: string, name: string) => {
    await renameDevice(uid, name);
    const device = get().currentDevice;
    if (device && device.uid === uid) {
      set({ currentDevice: { ...device, name } });
    }
  },

  accept: async (uid: string) => {
    await acceptDevice(uid);
    await get().fetch();
  },

  reject: async (uid: string) => {
    await rejectDevice(uid);
    await get().fetch();
  },

  remove: async (uid: string) => {
    await removeDevice(uid);
    await get().fetch();
  },

  addTag: async (uid: string, tag: string) => {
    await addDeviceTag(uid, tag);
    const device = get().currentDevice;
    if (device && device.uid === uid) {
      set({
        currentDevice: { ...device, tags: [...(device.tags || []), tag] },
      });
    }
  },

  removeTag: async (uid: string, tag: string) => {
    await removeDeviceTag(uid, tag);
    const device = get().currentDevice;
    if (device && device.uid === uid) {
      set({
        currentDevice: {
          ...device,
          tags: (device.tags || []).filter((t) => t !== tag),
        },
      });
    }
  },

  setPage: (page: number) => set({ page }),
  setStatus: (status: string) => set({ status, page: 1 }),

  addFilterTag: (tag: string) => {
    const current = get().filterTags;
    if (current.includes(tag)) return;
    const next = [...current, tag];
    set({ filterTags: next, page: 1 });
    get().fetch(1, undefined, undefined, next);
  },

  removeFilterTag: (tag: string) => {
    const next = get().filterTags.filter((t) => t !== tag);
    set({ filterTags: next, page: 1 });
    get().fetch(1, undefined, undefined, next);
  },

  clearFilterTags: () => {
    set({ filterTags: [], page: 1 });
    get().fetch(1, undefined, undefined, []);
  },
}));
