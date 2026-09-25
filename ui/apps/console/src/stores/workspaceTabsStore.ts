import { create } from "zustand";
import { persist } from "zustand/middleware";
import { moveById } from "@/utils/moveById";

/**
 * An open context in the tab strip: a namespace, or the admin console. Each remembers the page
 * it was last on, so coming back to it lands where the user left it.
 */
export type WorkspaceTab =
  | {
      id: string;
      kind: "namespace";
      tenant: string;
      name: string;
      path: string;
    }
  | { id: "admin"; kind: "admin"; name: string; path: string };

/**
 * The id of the admin console tab; there is only ever one.
 */
export const ADMIN_TAB_ID = "admin";

/**
 * The id a namespace tab is stored under, so opening the same namespace twice finds its tab.
 */
export const namespaceTabId = (tenant: string) => `ns:${tenant}`;

/**
 * The admin console's tab, landing on path when activated. ensure keeps an open tab's own path,
 * so path only decides where a tab not yet open lands.
 */
export function adminTab(path = "/admin/dashboard"): WorkspaceTab {
  return { id: ADMIN_TAB_ID, kind: "admin", name: "Admin Console", path };
}

/**
 * A namespace's tab, landing on path when activated, as adminTab does for the admin console.
 */
export function namespaceTab(
  tenant: string,
  name: string,
  path = "/dashboard",
): WorkspaceTab {
  return { id: namespaceTabId(tenant), kind: "namespace", tenant, name, path };
}

interface WorkspaceTabsState {
  tabs: WorkspaceTab[];
  failures: Record<string, string>;
  ensure: (tab: WorkspaceTab) => void;
  remember: (id: string, path: string) => void;
  remove: (id: string) => void;
  move: (id: string, to: number) => void;
  retain: (tenants: Set<string>) => void;
  fail: (id: string, reason: string) => void;
  clearFailure: (id: string) => void;
  clear: () => void;
}

/**
 * The contexts open as tabs. Only the tabs persist, so they survive a reload; why a tab last
 * failed to open is kept for the session alone, and the active tab is not kept at all, since it
 * follows the route and the session's tenant.
 */
export const useWorkspaceTabsStore = create<WorkspaceTabsState>()(
  persist(
    (set) => ({
      tabs: [],
      failures: {},
      ensure: (tab) =>
        set((state) => {
          const existing = state.tabs.find((t) => t.id === tab.id);
          if (!existing) return { tabs: [...state.tabs, tab] };
          if (existing.name === tab.name) return state;
          return {
            tabs: state.tabs.map((t) =>
              t.id === tab.id ? { ...t, name: tab.name } : t,
            ),
          };
        }),
      remember: (id, path) =>
        set((state) => ({
          tabs: state.tabs.map((t) => (t.id === id ? { ...t, path } : t)),
        })),
      remove: (id) =>
        set((state) => ({ tabs: state.tabs.filter((t) => t.id !== id) })),
      move: (id, to) =>
        set((state) => {
          const tabs = moveById(state.tabs, id, to);
          return tabs === state.tabs ? state : { tabs };
        }),
      retain: (tenants) =>
        set((state) => {
          const kept = state.tabs.filter(
            (t) => t.kind !== "namespace" || tenants.has(t.tenant),
          );
          return kept.length === state.tabs.length ? state : { tabs: kept };
        }),
      fail: (id, reason) =>
        set((state) => ({ failures: { ...state.failures, [id]: reason } })),
      clearFailure: (id) =>
        set((state) => {
          if (!(id in state.failures)) return state;
          const { [id]: _cleared, ...rest } = state.failures;
          return { failures: rest };
        }),
      clear: () => set({ tabs: [], failures: {} }),
    }),
    {
      name: "workspaceTabs",
      partialize: (state) => ({ tabs: state.tabs }),
    },
  ),
);
