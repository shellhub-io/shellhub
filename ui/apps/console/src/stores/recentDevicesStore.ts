import { create } from "zustand";
import { persist, createJSONStorage } from "zustand/middleware";
import { useAuthStore } from "./authStore";

/** A device the user has connected to, surfaced in the palette's Recent section. */
export interface RecentDevice {
  uid: string;
  name: string;
  /** ISO timestamp of the most recent connection — fed to `formatRelative`. */
  connectedAt: string;
}

interface RecentDevicesState {
  /** MRU lists keyed by tenant. Array position is the MRU order (`record`
   *  prepends); `connectedAt` is display-only, not the sort key. */
  byTenant: Record<string, RecentDevice[]>;
  /** Record a connection to a device under the active tenant (no-op if none). */
  record: (uid: string, name: string) => void;
}

/* Keep a few more than the palette shows (`RECENT_LIMIT` in useCommandPalette),
 * so hiding currently-open devices still leaves a full Recent section. */
const STORE_CAP = 10;

/**
 * Persisted most-recently-used device list, partitioned by tenant so it never
 * leaks across namespaces. Recorded at the single connect choke point
 * (`terminalStore.open`), so every entry point — palette, Devices page, device
 * details, containers — populates it.
 */
export const useRecentDevicesStore = create<RecentDevicesState>()(
  persist(
    (set) => ({
      byTenant: {},

      record: (uid, name) => {
        const tenant = useAuthStore.getState().tenant;
        if (!tenant) return;
        const entry: RecentDevice = {
          uid,
          name,
          connectedAt: new Date().toISOString(),
        };
        set((state) => {
          const existing = state.byTenant[tenant] ?? [];
          // Drop any prior entry for this device, then prepend the fresh one.
          const next = [entry, ...existing.filter((d) => d.uid !== uid)].slice(
            0,
            STORE_CAP,
          );
          return { byTenant: { ...state.byTenant, [tenant]: next } };
        });
      },
    }),
    {
      name: "shellhub:recent-devices",
      version: 1,
      storage: createJSONStorage(() => localStorage),
      partialize: (state) => ({ byTenant: state.byTenant }),
    },
  ),
);
