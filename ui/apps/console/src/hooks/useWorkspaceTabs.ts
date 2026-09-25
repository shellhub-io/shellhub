import { useLocation, useNavigate } from "react-router-dom";
import { apiErrorMessage } from "@/api/errors";
import { useEnterNamespace } from "@/hooks/useNamespaceMutations";
import { useAuthStore } from "@/stores/authStore";
import { useTerminalStore } from "@/stores/terminalStore";
import {
  ADMIN_TAB_ID,
  adminTab,
  namespaceTab,
  namespaceTabId,
  useWorkspaceTabsStore,
  type WorkspaceTab,
} from "@/stores/workspaceTabsStore";
import { isAdminPath } from "@/utils/adminRoute";

/**
 * The context tabs and how to move between them. A namespace tab is entered in place (see
 * useEnterNamespace), so the open terminals outlive the switch. When a namespace cannot be
 * entered nothing moves: the terminals stay as they were, the tab records why for the strip to
 * show, and activate resolves false. restoreSession brings that terminal forward once the tab's
 * page is reached, rather than minimizing them all.
 */
export function useWorkspaceTabs() {
  const tabs = useWorkspaceTabsStore((s) => s.tabs);
  const failures = useWorkspaceTabsStore((s) => s.failures);
  const { pathname } = useLocation();
  const navigate = useNavigate();
  const tenant = useAuthStore((s) => s.tenant);
  const enterNamespace = useEnterNamespace();

  const activeId = isAdminPath(pathname)
    ? ADMIN_TAB_ID
    : tenant
      ? namespaceTabId(tenant)
      : null;

  const activate = async (
    tab: WorkspaceTab,
    { restoreSession }: { restoreSession?: string } = {},
  ): Promise<boolean> => {
    const tabsStore = useWorkspaceTabsStore.getState();
    if (
      tab.kind === "namespace" &&
      tab.tenant !== useAuthStore.getState().tenant
    ) {
      try {
        await enterNamespace.mutateAsync(tab.tenant);
      } catch (error) {
        tabsStore.fail(tab.id, apiErrorMessage(error));
        return false;
      }
    }
    tabsStore.clearFailure(tab.id);

    const terminals = useTerminalStore.getState();
    if (restoreSession) terminals.setRestoreAfterNavigation(restoreSession);
    else terminals.minimizeAll();
    if (tab.path !== pathname || !terminals.restorePending()) {
      await navigate(tab.path);
    }
    return true;
  };

  const open = (
    tab: WorkspaceTab,
    options: { restoreSession?: string } = {},
  ) => {
    useWorkspaceTabsStore.getState().ensure(tab);
    const stored = useWorkspaceTabsStore
      .getState()
      .tabs.find((t) => t.id === tab.id);
    return stored ? activate(stored, options) : Promise.resolve(false);
  };

  const openNamespace = (
    namespaceTenant: string,
    name: string,
    options: { restoreSession?: string } = {},
  ) => open(namespaceTab(namespaceTenant, name), options);

  const openAdmin = () => open(adminTab());

  const close = (tab: WorkspaceTab) => {
    const index = tabs.findIndex((t) => t.id === tab.id);
    const remaining = tabs.filter((t) => t.id !== tab.id);
    if (remaining.length === 0) return;
    useWorkspaceTabsStore.getState().remove(tab.id);
    if (tab.id !== activeId) return;
    void activate(remaining[Math.max(0, index - 1)]);
  };

  return {
    tabs,
    failures,
    activeId,
    activate,
    openNamespace,
    openAdmin,
    close,
  };
}
