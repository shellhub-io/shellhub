import { useState, useCallback, useRef, useSyncExternalStore } from "react";
import { Outlet, useLocation } from "react-router-dom";
import AdminSidebar from "./AdminSidebar";
import AdminAppBar from "./AdminAppBar";

const lgQuery = "(min-width: 1024px)";

function subscribeToMediaQuery(callback: () => void) {
  const mql = window.matchMedia(lgQuery);
  mql.addEventListener("change", callback);
  return () => mql.removeEventListener("change", callback);
}

function getIsDesktop() {
  return window.matchMedia(lgQuery).matches;
}

function getIsDesktopServer() {
  return true;
}

export default function AdminLayout() {
  const [expanded, setExpanded] = useState(false);
  const [pinned, setPinned] = useState(false);
  const [drawerPathname, setDrawerPathname] = useState<string | null>(null);
  const { pathname } = useLocation();

  const isDesktop = useSyncExternalStore(
    subscribeToMediaQuery,
    getIsDesktop,
    getIsDesktopServer,
  );

  const isOpen = expanded || pinned;
  const drawerOpen = drawerPathname === pathname;
  const hoverTimer = useRef<ReturnType<typeof setTimeout>>(undefined);

  const handleMouseEnter = useCallback(() => {
    clearTimeout(hoverTimer.current);
    hoverTimer.current = setTimeout(() => setExpanded(true), 75);
  }, []);

  const handleMouseLeave = useCallback(() => {
    clearTimeout(hoverTimer.current);
    hoverTimer.current = setTimeout(() => setExpanded(false), 150);
  }, []);

  const handleToggle = useCallback(() => {
    setPinned((prev) => !prev);
  }, []);

  const openDrawer = useCallback(() => setDrawerPathname(pathname), [pathname]);
  const closeDrawer = useCallback(() => setDrawerPathname(null), []);

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === "Escape") closeDrawer();
    },
    [closeDrawer],
  );

  return (
    <div className="flex flex-col h-screen bg-background">
      <div className="flex flex-1 min-h-0 overflow-hidden">
        {isDesktop ? (
          <div
            onMouseEnter={handleMouseEnter}
            onMouseLeave={handleMouseLeave}
            onFocus={handleMouseEnter}
            onBlur={handleMouseLeave}
          >
            <AdminSidebar expanded={isOpen} pinned={pinned} onToggle={handleToggle} />
          </div>
        ) : (
          <div
            role="dialog"
            aria-modal={drawerOpen}
            aria-label="Navigation menu"
            className={`fixed inset-0 z-40 ${
              drawerOpen ? "" : "pointer-events-none"
            }`}
            onKeyDown={handleKeyDown}
            {...(!drawerOpen && { inert: "" })}
          >
            <div
              className={`absolute inset-0 bg-black/40 transition-opacity duration-200 ${
                drawerOpen ? "opacity-100" : "opacity-0"
              }`}
              onClick={closeDrawer}
              aria-hidden="true"
            />
            <div
              className={`fixed inset-y-0 left-0 z-50 w-[220px] transition-transform duration-200 ease-in-out ${
                drawerOpen ? "translate-x-0" : "-translate-x-full"
              }`}
            >
              <AdminSidebar
                expanded
                pinned={false}
                onToggle={closeDrawer}
                onClose={closeDrawer}
              />
            </div>
          </div>
        )}
        <div className="flex-1 flex flex-col min-w-0">
          <AdminAppBar onMenuToggle={isDesktop ? undefined : openDrawer} />
          <main className="flex-1 overflow-y-auto p-4 sm:p-8 relative">
            <div className="grid-bg scanline absolute inset-0 -z-10" />
            <div
              key={pathname}
              className="page-enter flex-1 flex flex-col"
            >
              <Outlet />
            </div>
          </main>
        </div>
      </div>
    </div>
  );
}
