import { useLocation, useNavigate } from "react-router-dom";
import { apiErrorMessage } from "@/api/errors";
import { useEnterNamespace } from "@/hooks/useNamespaceMutations";
import { useNamespaces } from "@/hooks/useNamespaces";
import { useAuthStore } from "@/stores/authStore";
import { useTerminalStore, type TerminalSession } from "@/stores/terminalStore";
import {
  ACCOUNT_TAB_ID,
  ADMIN_TAB_ID,
  PREFERENCES_TAB_ID,
  adminTab,
  preferencesTab,
  namespaceTab,
  namespaceTabId,
  useWorkspaceTabsStore,
  type WorkspaceTab,
} from "@/stores/workspaceTabsStore";
import { ADMIN_UNAUTHORIZED_PATH, isAdminPath } from "@/utils/adminRoute";
import { isAccountPath } from "@/utils/accountRoute";
import { isPreferencesPath } from "@/utils/preferencesRoute";

interface OpenOptions {
  restoreSession?: string;
  landOn?: string;
}

/**
 * The context tabs and how to move between them. A namespace tab is entered in place (see
 * useEnterNamespace), so the open terminals outlive the switch. When a namespace cannot be
 * entered nothing moves: the terminals stay as they were, the tab records why for the strip to
 * show, and activate resolves false. restoreSession brings that terminal forward once the tab's
 * page is reached, rather than minimizing them all; landOn lands on that page instead of the one
 * the tab last showed. showTerminal brings a terminal forward inside
 * its own namespace, entering it first when another one is active, and reopening its tab if it
 * was closed. Closing the active tab moves to its neighbour first, and keeps the tab when the
 * neighbour cannot be entered.
 */
export function useWorkspaceTabs() {
  const tabs = useWorkspaceTabsStore((s) => s.tabs);
  const failures = useWorkspaceTabsStore((s) => s.failures);
  const { pathname } = useLocation();
  const navigate = useNavigate();
  const tenant = useAuthStore((s) => s.tenant);
  const enterNamespace = useEnterNamespace();
  const { namespaces } = useNamespaces();

  const activeId = isAccountPath(pathname)
    ? ACCOUNT_TAB_ID
    : isPreferencesPath(pathname)
      ? PREFERENCES_TAB_ID
      : isAdminPath(pathname) && pathname !== ADMIN_UNAUTHORIZED_PATH
        ? ADMIN_TAB_ID
        : tenant
          ? namespaceTabId(tenant)
          : null;

  const activate = async (
    tab: WorkspaceTab,
    { restoreSession, landOn }: OpenOptions = {},
  ): Promise<boolean> => {
    const tabsStore = useWorkspaceTabsStore.getState();
    const destination = landOn ?? tab.path;
    const land = () => {
      const terminals = useTerminalStore.getState();
      if (restoreSession) terminals.setRestoreAfterNavigation(restoreSession);
      else terminals.minimizeAll();
      if (destination === pathname && terminals.showPendingRestore()) return;
      void navigate(destination);
    };
    if (
      tab.kind === "namespace" &&
      tab.tenant !== useAuthStore.getState().tenant
    ) {
      try {
        await enterNamespace.mutateAsync({ tenantId: tab.tenant, land });
      } catch (error) {
        tabsStore.fail(tab.id, apiErrorMessage(error));
        return false;
      }
    } else {
      land();
    }
    tabsStore.clearFailure(tab.id);
    return true;
  };

  const open = (tab: WorkspaceTab, options: OpenOptions = {}) => {
    useWorkspaceTabsStore.getState().ensure(tab);
    const stored = useWorkspaceTabsStore
      .getState()
      .tabs.find((t) => t.id === tab.id);
    return stored ? activate(stored, options) : Promise.resolve(false);
  };

  const openNamespace = (
    namespaceTenant: string,
    name: string,
    options: OpenOptions = {},
  ) => open(namespaceTab(namespaceTenant, name), options);

  const openAdmin = () => open(adminTab());

  const openPreferences = () => open(preferencesTab());

  const showTerminal = (session: TerminalSession) => {
    const home = session.tenant;
    const name = namespaces.find((ns) => ns.tenant_id === home)?.name;
    if (!home || !name || activeId === namespaceTabId(home)) {
      useTerminalStore.getState().restore(session.id);
      return Promise.resolve(true);
    }
    return openNamespace(home, name, { restoreSession: session.id });
  };

  const close = (tab: WorkspaceTab) => {
    const index = tabs.findIndex((t) => t.id === tab.id);
    const remaining = tabs.filter((t) => t.id !== tab.id);
    if (remaining.length === 0) return;
    const remove = () => useWorkspaceTabsStore.getState().remove(tab.id);
    if (tab.id !== activeId) {
      remove();
      return;
    }
    void activate(remaining[Math.max(0, index - 1)]).then((moved) => {
      if (moved) remove();
    });
  };

  return {
    tabs,
    failures,
    activeId,
    activate,
    openNamespace,
    openAdmin,
    openPreferences,
    showTerminal,
    close,
  };
}
