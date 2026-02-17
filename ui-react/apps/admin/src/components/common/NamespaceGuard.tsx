import { useEffect } from "react";
import { Outlet, useNavigate } from "react-router-dom";
import { useNamespacesStore } from "../../stores/namespacesStore";
import { useAuthStore } from "../../stores/authStore";
import { useConnectivityStore } from "../../stores/connectivityStore";
import CreateNamespace from "./CreateNamespace";

function MinimalHeader() {
  const { user, logout } = useAuthStore();
  const navigate = useNavigate();

  const handleLogout = () => {
    logout();
    navigate("/login");
  };

  return (
    <header className="h-14 bg-surface border-b border-border px-5 flex items-center justify-between shrink-0">
      <img src="/v2/ui/logo.svg" alt="ShellHub" className="h-6" />
      <div className="flex items-center gap-1">
        {user && (
          <span className="text-xs font-mono text-text-secondary px-3 py-1.5">
            {user}
          </span>
        )}
        <div className="w-px h-5 bg-border mx-1" />
        <button
          onClick={handleLogout}
          className="text-xs font-medium text-text-muted hover:text-accent-red px-3 py-1.5 rounded-md hover:bg-accent-red/5 transition-all duration-150"
        >
          Logout
        </button>
      </div>
    </header>
  );
}

export default function NamespaceGuard() {
  const { namespaces, loaded, fetch } = useNamespacesStore();
  const apiReachable = useConnectivityStore((s) => s.apiReachable);

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

  if (namespaces.length === 0) {
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
