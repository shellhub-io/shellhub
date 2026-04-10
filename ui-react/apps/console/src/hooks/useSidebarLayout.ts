import {
  useState,
  useRef,
  useEffect,
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
  const [expanded, setExpanded] = useState(false);
  const [pinned, setPinned] = useState(false);
  const [drawerOpen, setDrawerOpen] = useState(false);

  const isDesktop = useSyncExternalStore(
    subscribeToMediaQuery,
    getIsDesktop,
    getIsDesktopServer,
  );

  const hoverTimer = useRef<ReturnType<typeof setTimeout>>(undefined);

  const isOpen = expanded || pinned;

  const openDrawer = () => setDrawerOpen(true);
  const closeDrawer = () => setDrawerOpen(false);
  const toggleDrawer = () => setDrawerOpen((prev) => !prev);

  // Clean up hover timer on unmount
  useEffect(() => () => clearTimeout(hoverTimer.current), []);

  const handleExpand = () => {
    clearTimeout(hoverTimer.current);
    hoverTimer.current = setTimeout(() => setExpanded(true), 75);
  };

  const handleCollapse = () => {
    clearTimeout(hoverTimer.current);
    hoverTimer.current = setTimeout(() => setExpanded(false), 150);
  };

  const handleToggle = () => { setPinned((prev) => !prev); };

  const handleDrawerKeyDown = (e: React.KeyboardEvent) => { if (e.key === "Escape") closeDrawer(); };

  return {
    expanded,
    pinned,
    isOpen,
    isDesktop,
    drawerOpen,
    handlers: {
      onMouseEnter: handleExpand,
      onMouseLeave: handleCollapse,
      onFocus: handleExpand,
      onBlur: handleCollapse,
      onToggle: handleToggle,
      openDrawer,
      closeDrawer,
      toggleDrawer,
      onDrawerKeyDown: handleDrawerKeyDown,
    },
  };
}
