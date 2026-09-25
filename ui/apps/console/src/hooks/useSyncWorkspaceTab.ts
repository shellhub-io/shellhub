import { useEffect } from "react";
import { useAuthStore } from "@/stores/authStore";
import { useNamespace, useNamespaces } from "@/hooks/useNamespaces";
import {
  ADMIN_TAB_ID,
  adminTab,
  namespaceTab,
  namespaceTabId,
  useWorkspaceTabsStore,
} from "@/stores/workspaceTabsStore";
import { ADMIN_UNAUTHORIZED_PATH } from "@/utils/adminRoute";

/**
 * Keeps the context tabs in step with where the user is: the current context always has a tab,
 * even when reached by a link rather than the palette, that tab remembers the page it is on, and
 * a tab for a namespace the user no longer belongs to is dropped once the namespace list says so.
 * The admin console's refusal belongs to no context, so it neither opens a tab nor becomes the
 * page one returns to.
 */
export function useSyncWorkspaceTab(pathname: string, isAdminRoute: boolean) {
  const tenant = useAuthStore((s) => s.tenant);
  const { namespace } = useNamespace(tenant ?? "");
  const { namespaces, isLoading, error } = useNamespaces();

  useEffect(() => {
    if (pathname === ADMIN_UNAUTHORIZED_PATH) return;
    const store = useWorkspaceTabsStore.getState();
    if (isAdminRoute) {
      store.ensure(adminTab(pathname));
      store.remember(ADMIN_TAB_ID, pathname);
      return;
    }
    if (!tenant || !namespace) return;
    store.ensure(namespaceTab(tenant, namespace.name, pathname));
    store.remember(namespaceTabId(tenant), pathname);
  }, [pathname, isAdminRoute, tenant, namespace]);

  useEffect(() => {
    if (isLoading || error) return;
    useWorkspaceTabsStore
      .getState()
      .retain(new Set(namespaces.map((ns) => ns.tenant_id)));
  }, [namespaces, isLoading, error]);
}
