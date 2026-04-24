import {
  useState,
  useSyncExternalStore,
} from "react";

const lgQuery = "(min-width: 1024px)";

const lgMql =
  typeof window !== "undefined" ? window.matchMedia(lgQuery) : undefined;

function subscribeToMediaQuery(callback: () => void) {
  lgMql?.addEventListener("change", callback);
  return () => lgMql?.removeEventListener("change", callback);
}

function getIsDesktop() {
  return lgMql?.matches ?? true;
}

function getIsDesktopServer() {
  return true;
}

export function useSidebarLayout() {
  const [drawerOpen, setDrawerOpen] = useState(false);

  const isDesktop = useSyncExternalStore(
    subscribeToMediaQuery,
    getIsDesktop,
    getIsDesktopServer,
  );

  const isOpen = isDesktop;

  const openDrawer = () => setDrawerOpen(true);
  const closeDrawer = () => setDrawerOpen(false);
  const toggleDrawer = () => setDrawerOpen((prev) => !prev);

  const handleDrawerKeyDown = (e: React.KeyboardEvent) => { if (e.key === "Escape") closeDrawer(); };

  return {
    isOpen,
    isDesktop,
    drawerOpen,
    handlers: {
      openDrawer,
      closeDrawer,
      toggleDrawer,
      onDrawerKeyDown: handleDrawerKeyDown,
    },
  };
}
