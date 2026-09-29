import { create } from "zustand";
import { persist } from "zustand/middleware";
import { moveById } from "@/utils/moveById";
import { PREFERENCES_PATH } from "@/utils/preferencesRoute";

/**
 * An open context in the tab strip: a namespace, the admin console, the user's own account, or
 * this browser's preferences. Each remembers the page
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
  | { id: "admin"; kind: "admin"; name: string; path: string }
  | { id: "account"; kind: "account"; name: string; path: string }
  | { id: "preferences"; kind: "preferences"; name: string; path: string };

/**
 * The id of the admin console tab; there is only ever one.
 */
export const ADMIN_TAB_ID = "admin";

/**
 * The id a namespace tab is stored under, so opening the same namespace twice finds its tab.
 */
export const namespaceTabId = (tenant: string) => `ns:${tenant}`;

/**
 * The id of the account tab; there is only ever one.
 */
export const ACCOUNT_TAB_ID = "account";

/**
 * The account's tab, landing on path when activated, as adminTab does for the admin console.
 */
export function accountTab(path = "/account"): WorkspaceTab {
  return { id: ACCOUNT_TAB_ID, kind: "account", name: "Account", path };
}

/**
 * The id of the preferences tab; there is only ever one.
 */
export const PREFERENCES_TAB_ID = "preferences";

/**
 * This browser's preferences tab, landing on path when activated, as accountTab does for the
 * account.
 */
export function preferencesTab(path = PREFERENCES_PATH): WorkspaceTab {
  return { id: PREFERENCES_TAB_ID, kind: "preferences", name: "Preferences", path };
}

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
 * follows the route and the session's tenant. Version 1 moved Appearance out of the account, so
 * an account tab stored on /account/appearance is brought back to the profile.
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
      version: 1,
      migrate: (persisted, version) => {
        const state = persisted as Pick<WorkspaceTabsState, "tabs">;
        if (version >= 1) return state;
        return {
          tabs: state.tabs.map((tab) =>
            tab.kind === "account" &&
            tab.path.startsWith("/account/appearance")
              ? { ...tab, path: "/account/profile" }
              : tab,
          ),
        };
      },
      partialize: (state) => ({ tabs: state.tabs }),
    },
  ),
);
