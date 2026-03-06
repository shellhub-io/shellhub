import { useEffect } from "react";
import { Outlet, useLocation } from "react-router-dom";
import { useNamespacesStore } from "../../stores/namespacesStore";
import { useConnectivityStore } from "../../stores/connectivityStore";
import CreateNamespace from "./CreateNamespace";
import UserMenu from "../layout/UserMenu";

function MinimalHeader() {
  return (
    <header className="h-14 bg-surface border-b border-border px-5 flex items-center justify-between shrink-0">
      <img src="/v2/ui/logo.svg" alt="ShellHub" className="h-6" />
      <UserMenu />
    </header>
  );
}

export default function NamespaceGuard() {
  const { namespaces, loaded, fetch } = useNamespacesStore();
  const apiReachable = useConnectivityStore((s) => s.apiReachable);
  const { pathname } = useLocation();

  useEffect(() => {
    if (!loaded) fetch();
  }, [loaded, fetch]);

  // Retry namespace fetch when API comes back online
  useEffect(() => {
    if (apiReachable && !loaded) {
      fetch();
    }
  }, [apiReachable, loaded, fetch]);

  if (!loaded) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <div className="flex items-center gap-3">
          <span className="w-4 h-4 border-2 border-primary/30 border-t-primary rounded-full animate-spin" />
          <span className="text-xs font-mono text-text-muted">Loading…</span>
        </div>
      </div>
    );
  }

  if (namespaces.length === 0 && pathname !== "/profile") {
    return (
      <div className="min-h-screen bg-background flex flex-col">
        <MinimalHeader />
        <div className="flex-1 flex">
          <CreateNamespace />
        </div>
      </div>
    );
  }

  return <Outlet />;
}
